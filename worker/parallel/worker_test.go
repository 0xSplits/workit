package parallel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/registry"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xh3b4sd/logger"
)

// Test_Worker_Parallel_Daemon_error verifies that the *parallel.Worker logs any
// error that occurs.
func Test_Worker_Parallel_Daemon_error(t *testing.T) {
	var buf syncBuffer

	var err error
	{
		err = errors.New("test error")
	}

	var wor *Worker
	{
		wor = New(Config{
			Han: []handler.Cooler{
				&testHandler{coo: time.Hour},
				&testHandler{coo: time.Hour, err: err},
			},
			Log: logger.New(logger.Config{
				Filter: logger.NewLevelFilter("error"),
				Writer: &buf,
			}),
			Reg: registry.New(registry.Config{
				Env: "testing",
				Log: logger.Fake(),
				Met: recorder.NewMeter(recorder.MeterConfig{
					Env: "testing",
					Sco: "workit",
					Ver: "v0.1.0",
				}),
			}),
		})
	}

	{
		go wor.Daemon()
	}

	{
		<-wor.rdy
	}

	{
		time.Sleep(time.Millisecond)
	}

	{
		exp := `"level":"error", "message":"worker execution failed", "stack":{"context":[{"key":"handler","value":"parallel"}],"description":"test error",`
		if !strings.Contains(buf.String(), exp) {
			t.Fatal("expected", true, "got", false)
		}
	}
}

// Test_Worker_Parallel_Daemon_cancel verifies that the *parallel.Worker
// does not log filtered errors.
func Test_Worker_Parallel_Daemon_filter(t *testing.T) {
	var buf syncBuffer

	var err error
	{
		err = errors.New("test error")
	}

	var wor *Worker
	{
		wor = New(Config{
			Han: []handler.Cooler{
				&testHandler{coo: time.Hour},
				&testHandler{coo: time.Hour, err: err},
			},
			Log: logger.New(logger.Config{
				Filter: logger.NewLevelFilter("error"),
				Writer: &buf,
			}),
			Reg: registry.New(registry.Config{
				Env: "testing",
				Fil: isErr,
				Log: logger.Fake(),
				Met: recorder.NewMeter(recorder.MeterConfig{
					Env: "testing",
					Sco: "workit",
					Ver: "v0.1.0",
				}),
			}),
		})
	}

	{
		go wor.Daemon()
	}

	{
		<-wor.rdy
	}

	{
		time.Sleep(time.Millisecond)
	}

	{
		exp := ""
		if !strings.Contains(buf.String(), exp) {
			t.Fatal("expected", true, "got", false)
		}
	}
}

// Test_Worker_Parallel_Daemon_metrics verifies that *parallel.Worker registers
// prometheus metrics properly.
func Test_Worker_Parallel_Daemon_metrics(t *testing.T) {
	var reg *prometheus.Registry
	{
		reg = prometheus.NewRegistry()
	}

	var wor *Worker
	{
		wor = New(Config{
			Han: []handler.Cooler{
				&testHandler{coo: time.Hour},
				&testHandler{coo: time.Hour},
			},
			Log: logger.Fake(),
			Reg: registry.New(registry.Config{
				Env: "testing",
				Log: logger.Fake(),
				Met: recorder.NewMeter(recorder.MeterConfig{
					Env: "testing",
					Reg: reg,
					Sco: "workit",
					Ver: "v0.1.0",
				}),
			}),
		})
	}

	var ser *httptest.Server
	var url string
	{
		ser, url = tesSer(reg)
	}

	{
		defer ser.Close()
	}

	var res string
	{
		res = tesRes(url)
	}

	pat := `worker_handler_execution_duration_seconds_count\{env="testing",handler="parallel",otel_scope_name="workit\.testing\.splits\.org",otel_scope_schema_url="",otel_scope_version="[^"]*",success="true"\} 2`
	rgx := regexp.MustCompile(pat)

	if rgx.MatchString(res) {
		t.Fatal("expected", false, "got", true)
	}

	{
		go wor.Daemon()
	}

	{
		<-wor.rdy
	}

	{
		res = tesRes(url)
	}

	if !rgx.MatchString(res) {
		t.Fatal("expected", true, "got", false)
	}
}

// Test_Worker_Parallel_Daemon_pipeline verifies the isolation of worker handler
// failure domains so that the failure of one worker handler does not affect the
// execution of another. For this test to be of any use, it must be executed
// using the -race flag, which we do in CI according to
// .github/workflows/go-build.yaml.
func Test_Worker_Parallel_Daemon_pipeline(t *testing.T) {
	// For this test we use signal channels that control the fake worker handlers
	// that we schedule during testing. Every test handler receives its own input
	// channel in order to simulate work. The behaviour of each test handler
	// henceforth is contigent on the respective channel buffer. E.g. handler A is
	// configured with a minimally buffered signal channel, causing handler A to
	// block indefinitely. All handlers will use the same output channel, which we
	// use to verify the actually executed engine schedule.

	var in1 chan string
	var in2 chan string
	var out chan string
	{
		in1 = make(chan string, 1)
		in2 = make(chan string, 3)
		out = make(chan string, 6)
	}

	// We create a worker instance in order to test the engine's scheduling. Note
	// that handler A defines a Cooler() of 60 minutes, which goes way beyond the
	// exxecution time of this task. In other words, after the first execution,
	// handler A blocks forever, even if it were to receive more execution tickets
	// from its signal channel.

	var wor *Worker
	{
		wor = New(Config{
			Han: []handler.Cooler{
				&testHandler{inp: in1, out: out, coo: time.Hour},
				&testHandler{inp: in2, out: out, coo: 0},
			},
			Log: logger.Fake(),
			Reg: registry.New(registry.Config{
				Env: "testing",
				Log: logger.Fake(),
				Met: recorder.NewMeter(recorder.MeterConfig{
					Env: "testing",
					Sco: "workit",
					Ver: "v0.1.0",
				}),
			}),
		})
	}

	// Start the worker engine by calling Worker.Daemon() and wait for that
	// goroutine to register. Once the unbuffered ready channel is closed,
	// <-wor.rdy unblocks and the test execution starts.

	{
		go wor.Daemon()
	}

	{
		<-wor.rdy
	}

	// We fill up the signal channels to equal amounts in order to verify that
	// handler A is operating on a different schedule than handler B.

	go func() {
		in1 <- "one"
		in1 <- "one"
		in1 <- "one"
	}()

	go func() {
		in2 <- "two"
		in2 <- "two"
		in2 <- "two"
	}()

	// Define the expected result and collect the actual output signals generated
	// by the fake handlers. Once the two slices exp and act are of the same
	// length, we stop the test by closing the output channel. That will then
	// break the for loop below, allowing us to verify the test result.

	var exp []string
	{
		exp = []string{
			"one",

			"two",
			"two",
			"two",
		}
	}

	var act []string
	for x := range out {
		{
			act = append(act, x)
		}

		if len(act) == len(exp) {
			close(out)
		}
	}

	// Finally we either timeout or verify the result received over the output
	// channel. If all goes well, <-out should not block since the output channel
	// should have been closed once we collected the expected amount of output
	// signals. If the output channel were never to close, then the test would
	// fail in a timeout. This pattern is important for concurency tests, because
	// not guarding against silence failure may hide relevant scheduling issues.

	// Note that we are not verifying the exact order of received output signals,
	// because the nature of concurrent execution is quite probabilistic. In other
	// words, it is impossible to predict the order of code path executions within
	// goroutines. What we are verifying here is that handler A does not block the
	// execution of handler B, while handler A itself gets stuck.

	select {
	case <-out:
		{
			slices.Sort(act)
			slices.Sort(exp)
		}

		if !slices.Equal(act, exp) {
			t.Fatalf("expected %#v got %#v", exp, act)
		}
	case <-time.After(time.Second):
		t.Fatal("test timeout")
	}
}

//
//
//

func tesRes(url string) string {
	var err error

	var res *http.Response
	{
		res, err = http.Get(url)
		if err != nil {
			panic(err)
		}
	}

	{
		defer res.Body.Close() // nolint:errcheck
	}

	var byt []byte
	{
		byt, err = io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
	}

	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprint("expected", http.StatusOK, "got", res.Status))
	}

	return string(byt)
}

func tesSer(reg *prometheus.Registry) (*httptest.Server, string) {
	var rtr *mux.Router
	{
		rtr = mux.NewRouter()
	}

	{
		rtr.NewRoute().Methods("GET").Path("/metrics").Handler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	}

	var ser *httptest.Server
	{
		ser = httptest.NewServer(rtr)
	}

	return ser, ser.URL + "/metrics"
}

//
//
//

type testHandler struct {
	coo time.Duration
	err error
	inp chan string
	out chan string
}

// Cooler simply returns the underlying cooldown duration, defining how long
// this handler ought to sleep before being scheduled again.
func (h *testHandler) Cooler() time.Duration {
	return h.coo
}

// Ensure simply takes a signal out of the underlying input channel and puts it
// back into the underlying output channel.
func (h *testHandler) Ensure() error {
	if h.err != nil {
		return h.err
	}

	if h.inp == nil || h.out == nil {
		return nil
	}

	var s string
	{
		s = <-h.inp
	}

	{
		h.out <- s
	}

	return nil
}

//
//
//

// syncBuffer is needed to synchronize the concurrent io.Writer operations
// related to the logger interface that we are testing against.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
