// pqsd is an agent that connects to a postgresql cluster and manages stream emission.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	_ "net/http/pprof"

	_ "golang.org/x/net/trace"
	"google.golang.org/grpc"

	"github.com/pkg/errors"
	"github.com/tmc/pqstream"
	"github.com/tmc/pqstream/ctxutil"
	"github.com/tmc/pqstream/pqs"
)

var (
	postgresCluster = flag.String("connect", "", "postgresql cluster address")
	tableRegexp     = flag.String("tables", ".*", "regexp of tables to manage")
	remove          = flag.Bool("remove", false, "if true, remove triggers and exit")
	grpcAddr        = flag.String("addr", ":7000", "listen addr")
	debugAddr       = flag.String("debugaddr", ":7001", "listen debug addr")
	redactions      = flag.String("redactions", "", "details of fields to redact in JSON format i.e '{\"public\":{\"users\":[\"password\",\"ssn\"]}}'")
)

const (
	gracefulStopMaxWait = 10 * time.Second
)

func main() {
	flag.Parse()
	if err := run(ctxutil.BackgroundWithSignals()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		return err
	}

	tableRe, err := regexp.Compile(*tableRegexp)
	if err != nil {
		return err
	}

	opts := []pqstream.ServerOption{
		pqstream.WithTableRegexp(tableRe),
	}

	if (len(*redactions)) > 0 {
		rfields := make(pqstream.FieldRedactions)
		if err := json.NewDecoder(strings.NewReader(*redactions)).Decode(&rfields); err != nil {
			return errors.Wrap(err, "decoding redactions")
		}

		if len(rfields) > 0 {
			opts = append(opts, pqstream.WithFieldRedactions(rfields))
		}
	}

	server, err := pqstream.NewServer(*postgresCluster, opts...)
	if err != nil {
		return err
	}

	err = errors.Wrap(server.RemoveTriggers(), "RemoveTriggers")
	if err != nil || *remove {
		return err
	}

	if err := server.InstallTriggers(); err != nil {
		return errors.Wrap(err, "InstallTriggers")
	}

	go func() {
		if err := server.HandleEvents(ctx); err != nil {
			// TODO(tmc): try to be more graceful
			log.Fatalln(err)
		}
	}()

	s := grpc.NewServer()
	pqs.RegisterPQStreamServer(s, server)
	go func() {
		<-ctx.Done()
		s.GracefulStop()
		<-time.After(gracefulStopMaxWait)
		s.Stop()
	}()
	log.Println("listening on", *grpcAddr, "and", *debugAddr)
	err = s.Serve(lis)
	if strings.Contains(err.Error(), "use of closed network connection") {
		// return nil for expected error from the accept loop
		return nil
	}
	return err
}
