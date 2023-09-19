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

func createServer(t *testing.T) (net.Listener, *http.Server, *url.URL) {
	srv := &Server{}
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	httpSrv := &http.Server{
		Handler: srv,
	}

	go httpSrv.Serve(l)

	u, err := url.Parse("http://" + l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	return l, httpSrv, u
}

func TestLoadGenerator(t *testing.T) {
	defer goleak.VerifyNone(t)

	// don't annoy the tester, but allow the test to finish for large core counts.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.NumCPU()*20)*time.Second)
	defer cancel()

	l, httpSrv, u := createServer(t)
	defer httpSrv.Shutdown(context.Background())
	defer l.Close()

	lg := NewLoadGenerator(
		&BatteryProperties{
			Concurrency:             uint(runtime.NumCPU() / 2),    // give room for the server to work
			SimultaneousConnections: uint(runtime.NumCPU() * 1000), // a very conservative value for modern processors
			Ctx:                     ctx,
			URL:                     u,
		},
		uint(runtime.NumCPU()*100000), // roughly spoken, 100k conns * cpu count for the battery
	)

	if err := lg.Spawn(); err != nil {
		t.Fatal(err)
	}

	t.Log("total delivered: " + fmt.Sprintf("%d", lg.Properties.Stats.Successes+lg.Properties.Stats.Failures))
	t.Log("successes: " + fmt.Sprintf("%d", lg.Properties.Stats.Successes))
	t.Log("failures: " + fmt.Sprintf("%d", lg.Properties.Stats.Failures))
}

func TestBenchmarker(t *testing.T) {
	seconds := 30

	defer goleak.VerifyNone(t)

	// don't annoy the tester, but allow the test to finish for large core counts.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.NumCPU()*20)*time.Second)
	defer cancel()

	l, httpSrv, u := createServer(t)
	defer httpSrv.Shutdown(context.Background())
	defer l.Close()

	lg := NewBenchmarker(
		&BatteryProperties{
			Concurrency:             uint(runtime.NumCPU() / 2),    // give room for the server to work
			SimultaneousConnections: uint(runtime.NumCPU() * 1000), // a very conservative value for modern processors
			Ctx:                     ctx,
			URL:                     u,
		},
		time.Duration(seconds)*time.Second, // roughly spoken, 100k conns * cpu count for the battery
	)

	if err := lg.Spawn(); err != nil {
		t.Fatal(err)
	}

	t.Log("total delivered: " + fmt.Sprintf("%d", lg.Properties.Stats.Successes+lg.Properties.Stats.Failures))
	t.Log("successes: " + fmt.Sprintf("%d", lg.Properties.Stats.Successes))
	t.Log("failures: " + fmt.Sprintf("%d", lg.Properties.Stats.Failures))
	t.Log("requests/sec: " + fmt.Sprintf("%f", float64(lg.Properties.Stats.Successes)/float64(seconds)))
}
