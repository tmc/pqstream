// pqs is the client for psqd which allows subscription to change events in a postgres database cluster.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/go-stomp/stomp"
	"github.com/google/gops/agent"
	_ "golang.org/x/net/trace"
	"google.golang.org/grpc"

	"github.com/golang/protobuf/jsonpb"
	_ "github.com/kardianos/minwinsvc" // import minwinsvc for windows service support
	"github.com/pkg/errors"
	"github.com/tmc/pqstream/ctxutil"
	"github.com/tmc/pqstream/pqs"
)

var (
	pqsdAddr      = flag.String("connect", ":7000", "pqsd address")
	tableRegexp   = flag.String("tables", ".*", "regexp of tables to match")
	debugAddr     = flag.String("debugaddr", ":7001", "listen debug addr")
	activeMqAddr  = flag.String("amqaddr", "", "ActiveMq server to send messages to")
	actvieMqQueue = flag.String("amqqueue", "", "ActiveMq queue to send messages to")
)

func main() {
	flag.Parse()
	if err := run(ctxutil.BackgroundWithSignals()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// starts the gops diagnostics agent
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		return err
	}

	conn, err := grpc.DialContext(ctx, *pqsdAddr, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "dial")
	}
	defer conn.Close()

	c, err := pqs.NewPQStreamClient(conn).Listen(ctx, &pqs.ListenRequest{
		TableRegexp: *tableRegexp,
	})
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		log.Println("context done.")
	}()
	go http.ListenAndServe(*debugAddr, nil)

	// TODO(tmc): add format flag to control output (probably offer text/template)
	m := &jsonpb.Marshaler{}
	for {
		ev, err := c.Recv()
		if err != nil {
			return err
		}
		if strings.Compare(*activeMqAddr, "") != 0 && strings.Compare(*actvieMqQueue, "") != 0 {
			message, err := m.MarshalToString(ev)
			if err != nil {
				return err
			}
			if err := sendToActiveMq(message); err != nil {
				return err
			}
		}
		if err := m.Marshal(os.Stdout, ev); err != nil {
			return err
		}
		fmt.Println()
	}
}

func sendToActiveMq(messge string) error {
	conn, err := stomp.Dial("tcp", *activeMqAddr)

	if err != nil {
		return err
	}
	defer conn.Disconnect()

	err = conn.Send(*actvieMqQueue, "text/plain", []byte(messge))
	if err != nil {
		return err
	}

	return nil
}
