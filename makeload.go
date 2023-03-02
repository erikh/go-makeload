package makeload

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Requester interface {
	// add status and possibly payload comparison functions to determine success
	Deliver(*url.URL) error
}

type Stats struct {
	// add delivery time metrics as well as mean concurrency and other useful
	// stats for testing
	mutex     sync.Mutex
	Successes uint
	Failures  uint
}

type LoadGenerator struct {
	connMutex sync.RWMutex
	connCount uint
	connTotal uint

	Concurrency             uint
	SimultaneousConnections uint
	TotalConnections        uint
	URL                     *url.URL
	Ctx                     context.Context
	Stats                   Stats

	Requester
}

func (lg *LoadGenerator) Spawn() error {
	wg := &sync.WaitGroup{}
	wg.Add(int(lg.Concurrency))

	for i := uint(0); i < lg.Concurrency; i++ {
		go lg.MakeRequests(wg)
	}

	wg.Wait()
	return lg.Ctx.Err()
}

func (lg *LoadGenerator) MakeRequests(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-lg.Ctx.Done():
			return
		default:
		}

		lg.connMutex.RLock()
		if lg.SimultaneousConnections <= lg.connCount {
			time.Sleep(10 * time.Millisecond) // make this tweakable
			continue
		}

		if lg.TotalConnections <= lg.connTotal {
			lg.connMutex.RUnlock()
			return
		}
		lg.connMutex.RUnlock()

		lg.connMutex.Lock()
		lg.connCount++
		lg.connTotal++
		lg.connMutex.Unlock()

		err := lg.Deliver(lg.URL)

		lg.connMutex.Lock()
		lg.connCount--
		lg.connMutex.Unlock()

		lg.Stats.mutex.Lock()
		if err != nil {
			lg.Stats.Failures++
		} else {
			lg.Stats.Successes++
		}
		lg.Stats.mutex.Unlock()
	}
}

func (lg *LoadGenerator) Deliver(u *url.URL) error {
	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Non-200 status code")
	}

	return nil
}
