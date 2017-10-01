package pqstream

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tmc/pqstream/pqs"

	ptypes_struct "github.com/golang/protobuf/ptypes/struct"
)

const (
	minReconnectInterval = time.Second
	maxReconnectInterval = 10 * time.Second
	defaultPingInterval  = 9 * time.Second
	channel              = "pqstream_notify"

	fallbackIDColumnType = "integer" // TODO(tmc) parameterize
)

// subscription
type subscription struct {
	// while fn returns true the subscription will stay active
	fn func(*pqs.Event) bool
}

// Server implements PQStreamServer and manages both client connections and database event monitoring.
type Server struct {
	logger logrus.FieldLogger
	l      *pq.Listener
	db     *sql.DB
	ctx    context.Context

	tableRe *regexp.Regexp

	listenerPingInterval time.Duration
	subscribe            chan *subscription
	redactions           FieldRedactions
}

// statically assert that Server satisfies pqs.PQStreamServer
var _ pqs.PQStreamServer = (*Server)(nil)

// ServerOption allows customization of a new server.
type ServerOption func(*Server)

// WithTableRegexp controls which tables are managed.
func WithTableRegexp(re *regexp.Regexp) ServerOption {
	return func(s *Server) {
		s.tableRe = re
	}
}

// WithLogger allows attaching a custom logger.
func WithLogger(l logrus.FieldLogger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

// WithContext allows supplying a custom context.
func WithContext(ctx context.Context) ServerOption {
	return func(s *Server) {
		s.ctx = ctx
	}
}

// NewServer prepares a new pqstream server.
func NewServer(connectionString string, opts ...ServerOption) (*Server, error) {
	s := &Server{
		subscribe:  make(chan *subscription),
		redactions: make(FieldRedactions),

		ctx:                  context.Background(),
		listenerPingInterval: defaultPingInterval,
	}
	for _, o := range opts {
		o(s)
	}
	if s.logger == nil {
		s.logger = logrus.StandardLogger()
	}
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping")
	}
	s.l = pq.NewListener(connectionString, minReconnectInterval, maxReconnectInterval, func(ev pq.ListenerEventType, err error) {
		s.logger.WithField("listener-event", ev).Debugln("got listener event")
		if err != nil {
			s.logger.WithField("listener-event", ev).WithError(err).Errorln("got listener event error")
		}
	})
	if err := s.l.Listen(channel); err != nil {
		return nil, errors.Wrap(err, "listen")
	}
	if err := s.l.Listen(channel + "-ctl"); err != nil {
		return nil, errors.Wrap(err, "listen")
	}
	s.db = db
	return s, nil
}

// Close stops the pqstream server.
func (s *Server) Close() error {
	errL := s.l.Close()
	errDB := s.db.Close()
	if errL != nil {
		return errors.Wrap(errL, "listener")
	}
	if errDB != nil {
		return errors.Wrap(errDB, "DB")
	}
	return nil
}

// InstallTriggers sets up triggers to start observing changes for the set of tables in the database.
func (s *Server) InstallTriggers() error {
	_, err := s.db.Exec(sqlTriggerFunction)
	if err != nil {
		return err
	}
	// TODO(tmc): watch for new tables
	tableNames, err := s.tableNames()
	if err != nil {
		return err
	}
	for _, t := range tableNames {
		if err := s.installTrigger(t); err != nil {
			return errors.Wrap(err, fmt.Sprintf("installTrigger table %s", t))
		}
	}
	if len(tableNames) == 0 {
		return errors.New("no tables found")
	}
	return nil
}

func (s *Server) tableNames() ([]string, error) {
	rows, err := s.db.Query(sqlQueryTables)
	if err != nil {
		return nil, err
	}
	var tableNames []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintln("tableNames scan, after", len(tableNames)))
		}
		if s.tableRe != nil && !s.tableRe.MatchString(t) {
			continue
		}
		tableNames = append(tableNames, t)
	}
	return tableNames, nil
}

func (s *Server) installTrigger(table string) error {
	q := fmt.Sprintf(sqlInstallTrigger, table)
	_, err := s.db.Exec(q)
	return err
}

// RemoveTriggers removes triggers from the database.
func (s *Server) RemoveTriggers() error {
	tableNames, err := s.tableNames()
	if err != nil {
		return err
	}
	for _, t := range tableNames {
		if err := s.removeTrigger(t); err != nil {
			return errors.Wrap(err, fmt.Sprintf("removeTrigger table:%s", t))
		}
	}
	return nil
}

func (s *Server) removeTrigger(table string) error {
	q := fmt.Sprintf(sqlRemoveTrigger, table)
	_, err := s.db.Exec(q)
	return err
}

// fallbackLookup will be invoked if we have apparently exceeded the 8000 byte notify limit.
func (s *Server) fallbackLookup(e *pqs.Event) error {
	rows, err := s.db.Query(fmt.Sprintf(sqlFetchRowByID, e.Table, fallbackIDColumnType), e.Id)
	if err != nil {
		return errors.Wrap(err, "fallback query")
	}
	defer rows.Close()
	if rows.Next() {
		payload := ""
		if err := rows.Scan(&payload); err != nil {
			return errors.Wrap(err, "fallback scan")
		}
		e.Payload = &ptypes_struct.Struct{}
		if err := jsonpb.UnmarshalString(payload, e.Payload); err != nil {
			return errors.Wrap(err, "fallback unmarshal")
		}
	}
	return nil
}

func (s *Server) handleEvent(subscribers map[*subscription]bool, ev *pq.Notification) error {
	if ev == nil {
		return errors.New("got nil event")
	}

	re := &pqs.RawEvent{}
	if err := jsonpb.UnmarshalString(ev.Extra, re); err != nil {
		return errors.Wrap(err, "jsonpb unmarshal")
	}

	// perform field redactions
	s.redactFields(re)

	e := &pqs.Event{
		Schema:  re.Schema,
		Table:   re.Table,
		Op:      re.Op,
		Id:      re.Id,
		Payload: re.Payload,
	}

	if re.Op == pqs.Operation_UPDATE {
		if patch, err := generatePatch(re.Payload, re.Previous); err != nil {
			s.logger.WithField("event", e).WithError(err).Infoln("issue generating json patch")
		} else {
			e.Changes = patch
		}
	}

	if e.Payload == nil && e.Id != "" {
		if err := s.fallbackLookup(e); err != nil {
			s.logger.WithField("event", e).WithError(err).Errorln("fallback lookup failed")
		}

	}
	for s := range subscribers {
		if !s.fn(e) {
			delete(subscribers, s)
		}
	}
	return nil
}

// HandleEvents processes events from the database and copies them to relevant clients.
func (s *Server) HandleEvents(ctx context.Context) error {
	subscribers := map[*subscription]bool{}
	events := s.l.NotificationChannel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case sub := <-s.subscribe:
			s.logger.Debugln("got subscriber")
			subscribers[sub] = true
		case ev := <-events:
			// TODO(tmc): separate case handling into method
			s.logger.WithField("event", ev).Debugln("got event")
			if err := s.handleEvent(subscribers, ev); err != nil {
				return err
			}
		case <-time.After(s.listenerPingInterval):
			s.logger.WithField("interval", s.listenerPingInterval).Debugln("pinging")
			if err := s.l.Ping(); err != nil {
				return errors.Wrap(err, "Ping")
			}
		}
	}
}

// Listen handles a request to listen for database events and streams them to clients.
func (s *Server) Listen(r *pqs.ListenRequest, srv pqs.PQStream_ListenServer) error {
	ctx := srv.Context()
	s.logger.WithField("listen-request", r).Infoln("got listen request")
	tableRe, err := regexp.Compile(r.TableRegexp)
	if err != nil {
		return err
	}
	events := make(chan *pqs.Event) // TODO(tmc): will likely buffer after benchmarking
	s.subscribe <- &subscription{fn: func(e *pqs.Event) bool {
		if !tableRe.MatchString(e.Table) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case events <- e:
			return true
		}
	}}
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		case e := <-events:
			if err := srv.Send(e); err != nil {
				return err
			}
		}
	}
}
