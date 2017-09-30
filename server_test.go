package pqstream

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
)

var testConnectionString = "postgres://localhost?sslmode=disable"
var testConnectionStringTemplate = "postgres://localhost/%s?sslmode=disable"

var testDatabaseDDL = `create table notes (id serial, created_at timestamp, note text)`
var testInsert = `insert into notes values (default, default, 'here is a sample note')`
var testUpdate = `update notes set note = 'here is an updated note' where id=1`

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

func TestServer_HandleEvents(t *testing.T) {
	db := dbOrSkip(t)
	type testCase struct {
		name    string
		opts    []ServerOption
		fn      func(*testing.T, *Server)
		wantErr bool
	}
	tests := []testCase{
		{"basics", nil, nil, false},
		{"basic_insert", nil, func(t *testing.T, s *Server) {
			if _, err := s.db.Exec(testInsert); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"basic_insert_and_update", nil, func(t *testing.T, s *Server) {
			if _, err := s.db.Exec(testInsert); err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Second)
			if _, err := s.db.Exec(testUpdate); err != nil {
				t.Fatal(err)
			}
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseName := tt.name
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			cs, cleanup := testDBConn(t, db, caseName)
			defer cleanup()
			s, err := NewServer(cs, tt.opts...)
			s.listenerPingInterval = time.Second // move into a helper?
			if err != nil {
				t.Fatal(err)
			}
			s.InstallTriggers()
			defer s.RemoveTriggers()
			defer s.Close()
			go func(t *testing.T, tt testCase) {
				if err := s.HandleEvents(ctx); (err != nil) != tt.wantErr {
					t.Errorf("Server.HandleEvents(%v) error = %v, wantErr %v", ctx, err, tt.wantErr)
				}
			}(t, tt)
			if tt.fn != nil {
				tt.fn(t, s)
			}
			<-ctx.Done()

		})
	}
}
