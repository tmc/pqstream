package pqstream

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/golang/protobuf/jsonpb"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tmc/pqstream/pqs"

	ptypes_struct "github.com/golang/protobuf/ptypes/struct"
)

const (
	minReconnectInterval = time.Second
	maxReconnectInterval = 10 * time.Second
	channel              = "pqstream_notify"

	fallbackIdColumnType = "integer" // TODO(tmc) parameterize
)

// subscription
type subscription struct {
	// while fn returns true the subscription will stay active
	fn func(*pqs.Event) bool
}

// Server implements PQStreamServer and manages both client connections and database event monitoring.
type Server struct {
	l  *pq.Listener
	db *sql.DB

	tableRe *regexp.Regexp

	//mu          sync.RWMutex // protects the following
	//subscribers map[subscriberFunc]time.Time
	subscribe chan *subscription

	redactions FieldRedactions
}

// statically assert that Server satisifes pqs.PQStreamServer
var _ pqs.PQStreamServer = (*Server)(nil)

// ServerOption allows customization of a new server.
type ServerOption func(*Server)

// WithTableRegexp controls which tables are managed.
func WithTableRegexp(re *regexp.Regexp) ServerOption {
	return func(s *Server) {
		s.tableRe = re
	}
}

// NewServer prepares a new pqstream server.
func NewServer(connectionString string, opts ...ServerOption) (*Server, error) {
	s := &Server{
		subscribe:  make(chan *subscription),
		redactions: make(FieldRedactions),
	}
	for _, o := range opts {
		o(s)
	}
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping")
	}
	s.l = pq.NewListener(connectionString, minReconnectInterval, maxReconnectInterval, func(ev pq.ListenerEventType, err error) {
		log.Printf("listener-event: %#v\n", ev)
		if err != nil {
			log.Println("listener-event-err:", err)
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
	rows, err := s.db.Query(fmt.Sprintf(sqlFetchRowById, e.Table, fallbackIdColumnType), e.Id)
	if err != nil {
		return errors.Wrap(err, "fallback query")
	}
	defer rows.Close()
	if rows.Next() {
		payload := ""
		if err := rows.Scan(&payload); err != nil {
			return errors.Wrap(err, "fallback scan")
		} else {
			e.Payload = &ptypes_struct.Struct{}
			if err := jsonpb.UnmarshalString(payload, e.Payload); err != nil {
				return errors.Wrap(err, "fallback unmarshal")
			}
		}
	}
	return nil
}

// HandleEvents processes events from the database and copies them to relevent clients.
func (s *Server) HandleEvents(ctx context.Context) error {
	subscribers := map[*subscription]bool{}
	events := s.l.NotificationChannel()
	eventPingInterval := time.Second * 9
	for {
		select {
		case <-ctx.Done():
			return nil
		case s := <-s.subscribe:
			log.Println("got subscriber")
			subscribers[s] = true
		case ev := <-events:
			log.Println("got event:", ev)

			e := &pqs.Event{}
			if err := jsonpb.UnmarshalString(ev.Extra, e); err != nil {
				return errors.Wrap(err, "jsonpb unmarshal")
			}

			if e.Payload != nil {
				s.redactFields(e)
			}

			if e.Payload == nil && e.Id != "" {
				if err := s.fallbackLookup(e); err != nil {
					log.Println("fallback lookup failed:", err)
				}

			}
			for s := range subscribers {
				if !s.fn(e) {
					delete(subscribers, s)
				}
			}
		case <-time.After(eventPingInterval):
			log.Println("pinging")
			if err := s.l.Ping(); err != nil {
				return errors.Wrap(err, "Ping")
			}
		}
	}
	return nil
}

// Listen handles a request to listen for database events and streams them to clients.
func (s *Server) Listen(r *pqs.ListenRequest, srv pqs.PQStream_ListenServer) error {
	ctx := srv.Context()
	log.Printf("got listen request: %#v\n", r)
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
		case <-ctx.Done():
			return nil
		case e := <-events:
			if err := srv.Send(e); err != nil {
				return err
			}
		}
	}
	return nil
}
