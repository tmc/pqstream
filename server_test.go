package pqstream

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmc/pqstream/pqs"
	"google.golang.org/grpc"
)

var testConnectionString = "postgres://localhost?sslmode=disable"
var testConnectionStringTemplate = "postgres://localhost/%s?sslmode=disable"

var (
	testDatabaseDDL    = `create table notes (id serial, created_at timestamp, note text)`
	testInsert         = `insert into notes values (default, default, 'here is a sample note')`
	testInsertTemplate = `insert into notes values (default, default, '%s')`
	testUpdate         = `update notes set note = 'here is an updated note' where id=1`
	testUpdateTemplate = `update notes set note = 'i%s' where id=1`
)

func init() {
	if s := os.Getenv("PQSTREAM_TEST_DB_URL"); s != "" {
		testConnectionString = s
	}
	if s := os.Getenv("PQSTREAM_TEST_DB_TMPL_URL"); s != "" {
		testConnectionStringTemplate = s
	}
}

func TestWithTableRegexp(t *testing.T) {
	re := regexp.MustCompile(".*")
	tests := []struct {
		name string
		want *regexp.Regexp
	}{
		{"basic", re},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewServer(testConnectionString, WithTableRegexp(re))
			if err != nil {
				t.Fatal(err)
			}
			if got := s.tableRe; got != tt.want {
				t.Errorf("WithTableRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		connectionString string
		opts             []ServerOption
	}
	tests := []struct {
		name    string
		args    args
		check   func(t *testing.T, s *Server)
		wantErr bool
	}{
		{"bad", args{
			connectionString: "this is an invalid connection string",
		}, nil, true},
		{"empty", args{
			connectionString: "",
		}, nil, true},
		{"good", args{
			connectionString: testConnectionString,
		}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(tt.args.connectionString, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func dbOrSkip(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", testConnectionString)
	if err != nil {
		t.Skip(err)
	}
	if err := db.Ping(); err != nil {
		t.Skip(errors.Wrap(err, "ping"))
	}
	return db
}

func testDBConn(t *testing.T, db *sql.DB, testcase string) (connectionString string, cleanup func()) {
	s := fmt.Sprintf("test_pqstream_%s", testcase)
	db.Exec(fmt.Sprintf("drop database %s", s))
	_, err := db.Exec(fmt.Sprintf("create database %s", s))
	if err != nil {
		t.Fatal(err)
	}
	dsn := fmt.Sprintf(testConnectionStringTemplate, s)
	newDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skip(err)
	}
	if err := db.Ping(); err != nil {
		t.Skip(errors.Wrap(err, "ping"))
	}
	defer newDB.Close()
	_, err = newDB.Exec(testDatabaseDDL)
	if err != nil {
		t.Fatal(err)
	}
	return dsn, func() {
		_, err := db.Exec(fmt.Sprintf("drop database %s", s))
		if err != nil {
			t.Fatal(err)
		}
	}
}
func mkString(len int, c byte) string {
	buf := make([]byte, len)
	for i := range buf {
		buf[i] = c
	}
	return string(buf)
}

type logWriter struct {
	*testing.T
}

func (l logWriter) Write(b []byte) (int, error) {
	l.Log(string(b))
	return len(b), nil
}

func loggerFromT(t *testing.T) *logrus.Logger {
	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	}
	logger.Formatter.(*logrus.TextFormatter).ForceColors = true
	logger.Out = logWriter{t}
	return logger
}

func TestServer_HandleEvents(t *testing.T) {
	db := dbOrSkip(t)
	type testCase struct {
		name    string
		fn      func(*testing.T, *Server)
		wantErr bool
	}
	tests := []testCase{
		{"basics", nil, false},
		{"basic_insert", func(t *testing.T, s *Server) {
			if _, err := s.db.Exec(testInsert); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"basic_insert_and_update", func(t *testing.T, s *Server) {
			if _, err := s.db.Exec(testInsert); err != nil {
				t.Fatal(err)
			}
			time.Sleep(10 * time.Millisecond)
			if _, err := s.db.Exec(testUpdate); err != nil {
				t.Fatal(err)
			}
		}, false},
	}

	mkTestCase := func(n int, alsoUpdate bool) testCase {
		caseName := fmt.Sprintf("test_%vb_insert", n)
		if alsoUpdate {
			caseName += "_and_update"
		}
		return testCase{caseName, func(t *testing.T, s *Server) {
			insert := fmt.Sprintf(testInsertTemplate, mkString(n, '.'))
			s.logger.Debugln("inserting", n)
			if _, err := s.db.Exec(insert); err != nil {
				t.Fatal(err)
			}
			if alsoUpdate {
				time.Sleep(10 % time.Millisecond)
				update := fmt.Sprintf(testUpdateTemplate, mkString(n, '-'))
				if _, err := s.db.Exec(update); err != nil {
					t.Fatal(err)
				}
			}
		}, false}
	}

	// TODO(tmc): encode the expected properties of the payloads in test
	// cross the 8k boundary for inserts
	for i := 7870; i <= 7900; i = i + 10 {
		tests = append(tests, mkTestCase(i, false))
	}
	// cross the 8k boundary for updates (and drop previous payloads)
	for i := 3890; i <= 4000; i = i + 10 {
		tests = append(tests, mkTestCase(i, true))
	}
	// cross the 8k boundary for updates (and drop payloads)
	for i := 7870; i <= 7900; i = i + 10 {
		tests = append(tests, mkTestCase(i, true))
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			cs, cleanup := testDBConn(t, db, tt.name)
			defer cleanup()
			s, err := NewServer(cs, WithLogger(loggerFromT(t)))
			s.listenerPingInterval = time.Second // move into a helper?
			if err != nil {
				t.Fatal(err)
			}
			s.InstallTriggers()
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			go func(t *testing.T, tt testCase) {
				if err := s.HandleEvents(ctx); (err != nil) != tt.wantErr {
					t.Errorf("Server.HandleEvents(%v) error = %v, wantErr %v", ctx, err, tt.wantErr)
				}
			}(t, tt)
			if tt.fn != nil {
				tt.fn(t, s)
			}
			if err := s.RemoveTriggers(); err != nil {
				t.Error(err)
			}
			<-ctx.Done()

		})
	}
}

const integrationExec = "TEST_INTEGRATION_EXEC_ENV"

func TestServer_Listen(t *testing.T) {
	db := dbOrSkip(t)
	type args struct {
		NInserts           int
		IntegrationExecEnv string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"basics", args{NInserts: 2}, true},
		{"python_integration", args{
			NInserts:           10,
			IntegrationExecEnv: "TEST_PYTHON_INTEGRATION_EXEC",
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.args.NInserts+1)*time.Second)
			defer cancel()

			lis, err := net.Listen("tcp", "")
			if err != nil {
				t.Fatal(err)
			}
			defer lis.Close()
			port := lis.Addr().(*net.TCPAddr).Port
			srv := grpc.NewServer()

			cs, cleanup := testDBConn(t, db, tt.name)
			defer cleanup()
			l := logrus.New()
			l.Level = logrus.DebugLevel
			sctx, scancel := context.WithCancel(ctx)
			s, err := NewServer(cs, WithLogger(l), WithContext(sctx))

			if err != nil {
				t.Fatal(err)
			}
			s.InstallTriggers()
			defer s.RemoveTriggers()
			defer s.Close()

			defer srv.Stop()
			pqs.RegisterPQStreamServer(srv, s)
			go srv.Serve(lis)

			// if the case supplies an integration client to run, start it.
			if tt.args.IntegrationExecEnv != "" {
				integrationExec := os.Getenv(tt.args.IntegrationExecEnv)
				t.Log("intergration exec", integrationExec)
				cmd := exec.Command("sh", []string{"-c", integrationExec}...)
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%v", port))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				go func() {
					if err := cmd.Run(); err != nil {
						t.Fatal(err)
					}
				}()
			}

			go func() {
				defer scancel()
				go s.HandleEvents(sctx)
				// generate some traffic
				for i := 0; i < tt.args.NInserts; i++ {
					log.Println("inserting", i)
					time.Sleep(time.Second)
					if _, err := s.db.Exec(testInsert); err != nil {
						s.logger.Error(err)
					}
				}
			}()

			conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure())
			if err != nil {
				t.Fatal(err)
			}
			c := pqs.NewPQStreamClient(conn)

			client, err := c.Listen(ctx, &pqs.ListenRequest{})
			if err != nil {
				t.Fatal(err)
			}
			for {
				ev, err := client.Recv()
				if err != nil {
					if err == io.EOF {
						return
					}
					t.Fatal(err)
				}
				t.Log("got event", ev)
			}
		})
	}
}
