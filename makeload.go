// Package makeload is a load generation library for testing. See the README
// for usage information.
package makeload

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Requester is an interface to modify how requests are made to the remote
// service. If supplied to the LoadGenerator struct, it will be used.
// Alternatively, LoadGenerator already supplies its own implementation.
type Requester interface {
	// Code to run when HTTP request delivery is made.
	Deliver(*url.URL) error
	// add status and possibly payload comparison functions to determine success
}

// Stats emcompasses the statistics for the requests already made.
type Stats struct {
	// Count of successes up to this point.
	successes uint64
	// Count of failures up to this point.
	failures uint64
}

func (s *Stats) Successes() uint64 {
	return s.successes
}

func (s *Stats) Failures() uint64 {
	return s.failures
}

// BatteryProperties are the list of properties common to each load generator.
type BatteryProperties struct {
	connTotal uint64

	// Concurrency is the number of concurrent request processors to make at a
	// time.
	concurrency uint64
	// SimultaneousConnections is the number of connections to maintain open at
	// any given time.
	simultaneousConnections uint64
	// URL is the url to connect to,
	url *url.URL
	// Ctx is a context which, if canceled, will stop load generation.
	ctx context.Context
	// Stats is the statistics generated by already processed load generation.
	stats Stats

	// Requester is the interface to HTTP requests.
	Requester

	client *http.Client
}

// NewBatteryProperties constructs a new BatteryProperties struct, which
// contains load generating parameters. Use this function to build it.
func NewBatteryProperties(ctx context.Context, concurrency, simultaneousConnections uint64, u *url.URL) *BatteryProperties {
	return &BatteryProperties{
		concurrency:             concurrency,
		simultaneousConnections: simultaneousConnections,
		url:                     u,
		ctx:                     ctx,
		client: &http.Client{
			Timeout: 0,
			Transport: &http.Transport{
				IdleConnTimeout:     0,
				MaxConnsPerHost:     0,
				MaxIdleConns:        0,
				MaxIdleConnsPerHost: 0,
			},
		},
	}
}

func (bp *BatteryProperties) Stats() *Stats {
	return &bp.stats
}

// Benchmarker is the main load generator struct for managing benchmarking.
// Call Spawn() on the result.
type Benchmarker struct {
	Properties *BatteryProperties
	Time       time.Duration
}

func NewBenchmarker(properties *BatteryProperties, duration time.Duration) *Benchmarker {
	return &Benchmarker{
		Properties: properties,
		Time:       duration,
	}
}

// Spawn launches the benchmarker. It will return an error if there was an
// error generating load. It does not error if there were issues making the
// individual HTTP requests. The Stats struct tracks the count of failures.
//
// To cancel load generation, cancel a passed context.
func (b *Benchmarker) Spawn() error {
	start := time.Now()
	return b.Properties.spawn(func(total uint64) bool { return start.Add(b.Time).Before(time.Now()) })
}

// LoadGenerator is the main load generator struct for managing load
// generation. One would construct this struct, filling out the proper
// parameters, and then would run Spawn() on the struct to apply load.
type LoadGenerator struct {
	Properties       *BatteryProperties
	TotalConnections uint64
}

func NewLoadGenerator(properties *BatteryProperties, total uint64) *LoadGenerator {
	return &LoadGenerator{
		Properties:       properties,
		TotalConnections: total,
	}
}

// Spawn launches the load generator. It will return an error if there was an
// error generating load. It does not error if there were issues making the
// individual HTTP requests. The Stats struct tracks the count of failures.
//
// To cancel load generation, cancel a passed context.
func (lg *LoadGenerator) Spawn() error {
	return lg.Properties.spawn(func(total uint64) bool { return lg.TotalConnections/lg.Properties.concurrency <= total })
}

func (p *BatteryProperties) spawn(cancelFunc func(uint64) bool) error {
	wg := &sync.WaitGroup{}
	wg.Add(int(p.concurrency))
	totals := make(chan uint64, p.concurrency)
	successes := make(chan uint64, p.concurrency)
	failures := make(chan uint64, p.concurrency)

	for i := uint64(0); i < p.concurrency; i++ {
		go makeRequests(wg, p, totals, successes, failures, cancelFunc)
	}

	for i := uint64(0); i < p.concurrency*3; i++ {
		select {
		case total := <-totals:
			p.connTotal += total
		case success := <-successes:
			p.stats.successes += success
		case failure := <-failures:
			p.stats.failures += failure
		}
	}

	wg.Wait()
	return p.ctx.Err()
}

// this function performs the actual request delivery. it is run in multiple goroutines.
func makeRequests(wg *sync.WaitGroup, properties *BatteryProperties, totals, successes, failures chan uint64, cancelFunc func(uint64) bool) {
	defer wg.Done()

	connCount := uint64(0)
	connTotal := uint64(0)
	successCount := uint64(0)
	failureCount := uint64(0)

	defer func() {
		totals <- connTotal
		failures <- failureCount
		successes <- successCount
	}()

	for {
		select {
		case <-properties.ctx.Done():
			return
		default:
		}

		if properties.simultaneousConnections <= connCount {
			continue
		}

		if cancelFunc(connTotal) {
			return
		}

		connCount += 1
		connTotal += 1

		err := properties.Deliver(properties.url)

		connCount -= 1

		if err != nil {
			failureCount += 1
		} else {
			successCount += 1
		}
	}

}

// Deliver satisfies the Requester interface and encompasses basic delivery of
// a HTTP GET request.
func (p *BatteryProperties) Deliver(u *url.URL) error {
	resp, err := p.client.Get(u.String())
	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Non-200 status code")
	}

	return nil
}
