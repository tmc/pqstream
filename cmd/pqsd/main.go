// pqsd is an agent that connects to a postgresql cluster and manages stream emission.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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
	remove          = flag.Bool("remove", false, "if true, remove triggers and exit")
	grpcAddr        = flag.String("addr", ":7000", "listen addr")
	debugAddr       = flag.String("debugaddr", ":7001", "listen debug addr")
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

	server, err := pqstream.NewServer(*postgresCluster)
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
