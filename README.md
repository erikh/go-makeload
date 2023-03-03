## makeload: a small library to make HTTP load

I haven't written code in some time and needed to warm up. So I wrote this in preparation for a project that needs it, and eventually as I improve this library, this message will disappear and more information about what it actually does will appear in its place.

This library is a HTTP load generator, similar in function to `wrk` or `ab`. It has a programmable interface, and is intended for use in integration testing of a HTTP service. It has a very small statistics collector, an interface for delivering requests, and concurrency and connection controls. It is currently very focused on lean functionality and accuracy. The library has reliably tested to deliver the exact amount of requests you deliver it, saturating the exact number of cores fitting the concurrency mark.

### Usage

`makeload` is a library, which you can use by incorporating the library with `go get` and then using the code to drive it. Here is an example from the test suite:

```go 
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

  "github.com/erikh/go-makeload"
)

type Server struct{}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

func TestBasic(t *testing.T) {
	srv := &Server{}
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	go http.Serve(l, srv)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // don't annoy the tester.
	defer cancel()

	u, err := url.Parse("http://" + l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	lg := &makeload.LoadGenerator{
		Concurrency:             uint(runtime.NumCPU() / 2),          // give room for the server to work
		SimultaneousConnections: uint(runtime.NumCPU() * 1000),       // a very conservative value for modern processors
		TotalConnections:        uint(runtime.NumCPU() * 1000 * 100), // roughly spoken, 100k conns * cpu count for the battery
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
```

Upon running this test with `go test -v`, you would see output like:

```
=== RUN   TestBasic
    main_test.go:49: total delivered: 800000
    main_test.go:50: successes: 798799
    main_test.go:51: failures: 1201
--- PASS: TestBasic (38.34s)
PASS
```

### The future

As mentioned, more is to come with regards to this library's functionality. Here are some things that will probably show up eventually:

- [ ] Statistics for mean delivery time
- [ ] Programmable functionality for determining errors / valid responses (right now just non-200's are errors)
- [ ] Programmable request delivery
- [ ] Some more self-testing

As mentioned, I'm shipping this to be a part of another product. If you file bugs for it, I will attempt to service requests, but if they conflict with the other project's goals, I strongly suggest you fork this library instead of push harder for your changes, which is MIT licensed for a reason.

May peace be with you.

### Author

Erik Hollensbe <erik+github@hollensbe.org>
