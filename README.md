## makeload: a small library to make HTTP load

This library is a HTTP load generator, similar in function to `wrk` or `ab`. It
has a programmable interface, and is intended for use in integration testing of
a HTTP service. It has a very small statistics collector, an interface for
delivering requests, and concurrency and connection controls. It is currently
very focused on lean functionality and accuracy. The library has reliably
tested to deliver the exact amount of requests you deliver it, saturating the
exact number of cores fitting the concurrency mark.

It has benchmarking and load generation functions. See the library
documentation for more.

### Usage

`makeload` is a library, which you can use by incorporating the library with
`go get` and then using the code to drive it. Here is an example from the test
suite:

```go
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
```

Upon running this test with `go test -v`, you would see output like:

```
=== RUN   TestLoadGenerator
    makeload_test.go:49: total delivered: 800000
    makeload_test.go:50: successes: 798799
    makeload_test.go:51: failures: 1201
--- PASS: TestLoadGenerator (38.34s)
PASS
```

### The future

As mentioned, more is to come with regards to this library's functionality.
Here are some things that will probably show up eventually:

-   [ ] Statistics for mean delivery time
-   [ ] Programmable functionality for determining errors / valid responses
        (right now just non-200's are errors)
-   [ ] Programmable request delivery
-   [ ] Some more self-testing

As mentioned, I'm shipping this to be a part of another product. If you file
bugs for it, I will attempt to service requests, but if they conflict with the
other project's goals, I strongly suggest you fork this library instead of push
harder for your changes, which is MIT licensed for a reason.

May peace be with you.

### Author

Erik Hollensbe <erik+github@hollensbe.org>
