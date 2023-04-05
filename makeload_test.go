package makeload

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type Server struct{}

func (s *Server) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestBasic(t *testing.T) {
	defer goleak.VerifyNone(t)

	srv := &Server{}
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := &http.Server{
		Handler: srv,
	}

	go httpSrv.Serve(l)
	defer httpSrv.Shutdown(context.Background())
	defer l.Close()

	// don't annoy the tester, but allow the test to finish for large core counts.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.NumCPU()*20)*time.Second)
	defer cancel()

	u, err := url.Parse("http://" + l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	lg := &LoadGenerator{
		Concurrency:             uint(runtime.NumCPU() / 2),      // give room for the server to work
		SimultaneousConnections: uint(runtime.NumCPU() * 1000),   // a very conservative value for modern processors
		TotalConnections:        uint(runtime.NumCPU() * 100000), // roughly spoken, 100k conns * cpu count for the battery
		Ctx:                     ctx,
		URL:                     u,
	}

	if err := lg.Spawn(); err != nil {
		t.Fatal(err)
	}

	t.Log("total delivered: " + fmt.Sprintf("%d", lg.Stats.Successes+lg.Stats.Failures))
	t.Log("successes: " + fmt.Sprintf("%d", lg.Stats.Successes))
	t.Log("failures: " + fmt.Sprintf("%d", lg.Stats.Failures))
}
