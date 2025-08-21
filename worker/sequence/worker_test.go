package sequence

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xh3b4sd/logger"
)

// Test_Worker_Sequence_Daemon_error verifies that the *sequence.Worker executes
// all handlers in order, until an error occurs.
func Test_Worker_Sequence_Daemon_error(t *testing.T) {
	var buf syncBuffer

	var err error
	var sig chan int
	{
		err = errors.New("test error")
		sig = make(chan int)
	}

	var wor *Worker
	{
		wor = New(Config{
			Coo: time.Hour,
			Han: [][]handler.Ensure{
				{&errorHandler{sig, 3, nil}},
				{&errorHandler{sig, 4, nil}},
				{&errorHandler{sig, 5, err}},
				{&errorHandler{sig, 6, nil}},
				{&errorHandler{sig, 7, nil}},
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

	var act []int
	for x := range sig {
		act = append(act, x)
	}

	{
		time.Sleep(time.Millisecond)
	}

	{
		exp := []int{3, 4, 5}
		if dif := cmp.Diff(exp, act); dif != "" {
			t.Fatalf("-expected +actual:\n%s", dif)
		}
	}

	{
		exp := `"level":"error", "message":"worker execution failed", "stack":{"context":[{"key":"handler","value":"sequence"}],"description":"test error",`
		if !strings.Contains(buf.String(), exp) {
			t.Fatal("expected", true, "got", false)
		}
	}
}

// Test_Worker_Sequence_Daemon_cancel verifies that the *sequence.Worker
// does not log filtered errors.
func Test_Worker_Sequence_Daemon_filter(t *testing.T) {
	var buf syncBuffer

	var err error
	var sig chan int
	{
		err = errors.New("test error")
		sig = make(chan int)
	}

	var wor *Worker
	{
		wor = New(Config{
			Coo: time.Hour,
			Han: [][]handler.Ensure{
				{&errorHandler{sig, 3, nil}},
				{&errorHandler{sig, 4, nil}},
				{&errorHandler{sig, 5, err}},
				{&errorHandler{sig, 6, nil}},
				{&errorHandler{sig, 7, nil}},
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

	var act []int
	for x := range sig {
		act = append(act, x)
	}

	{
		time.Sleep(time.Millisecond)
	}

	{
		exp := []int{3, 4, 5}
		if dif := cmp.Diff(exp, act); dif != "" {
			t.Fatalf("-expected +actual:\n%s", dif)
		}
	}

	{
		exp := ""
		if dif := cmp.Diff(exp, buf.String()); dif != "" {
			t.Fatalf("-expected +actual:\n%s", dif)
		}
	}
}

// Test_Worker_Sequence_Daemon_order verifies that the *sequence.Worker executes
// all handlers in order.
func Test_Worker_Sequence_Daemon_order(t *testing.T) {
	var sig chan int
	{
		sig = make(chan int)
	}

	var wor *Worker
	{
		wor = New(Config{
			Coo: time.Hour,
			Han: [][]handler.Ensure{
				{&orderHandler{sig, 3, false}},
				{&orderHandler{sig, 4, false}},
				{&orderHandler{sig, 5, false}},
				{&orderHandler{sig, 6, false}},
				{&orderHandler{sig, 7, true}},
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

	{
		go wor.Daemon()
	}

	var act []int
	for x := range sig {
		act = append(act, x)
	}

	var exp []int
	{
		exp = []int{3, 4, 5, 6, 7}
	}

	if dif := cmp.Diff(exp, act); dif != "" {
		t.Fatalf("-expected +actual:\n%s", dif)
	}
}

// Test_Worker_Sequence_Daemon_metrics verifies that *sequence.Worker registers
// prometheus metrics properly.
func Test_Worker_Sequence_Daemon_metrics(t *testing.T) {
	var reg *prometheus.Registry
	{
		reg = prometheus.NewRegistry()
	}

	var sig chan int
	{
		sig = make(chan int)
	}

	var wor *Worker
	{
		wor = New(Config{
			Coo: time.Hour,
			Han: [][]handler.Ensure{
				{&orderHandler{sig, 3, false}},
				{&orderHandler{sig, 4, false}},
				{&orderHandler{sig, 5, false}},
				{&orderHandler{sig, 6, false}},
				{&orderHandler{sig, 7, true}},
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

	pat := `worker_handler_execution_duration_seconds_count\{env="testing",handler="sequence",otel_scope_name="workit\.testing\.splits\.org",otel_scope_schema_url="",otel_scope_version="[^"]*",success="true"\} 5`
	rgx := regexp.MustCompile(pat)

	if rgx.MatchString(res) {
		t.Fatal("expected", false, "got", true)
	}

	{
		go wor.Daemon()
	}

	var act []int
	for x := range sig {
		act = append(act, x)
	}

	var exp []int
	{
		exp = []int{3, 4, 5, 6, 7}
	}

	if dif := cmp.Diff(exp, act); dif != "" {
		t.Fatalf("-expected +actual:\n%s", dif)
	}

	{
		res = tesRes(url)
	}

	if !rgx.MatchString(res) {
		t.Fatal("expected", true, "got", false)
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

type errorHandler struct {
	sig chan int
	num int
	err error
}

func (h *errorHandler) Ensure() error {
	{
		h.sig <- h.num
	}

	if h.err != nil {
		close(h.sig)
		return h.err
	}

	return nil
}

//
//
//

type orderHandler struct {
	sig chan int
	num int
	clo bool
}

func (h *orderHandler) Ensure() error {
	{
		h.sig <- h.num
	}

	if h.clo {
		close(h.sig)
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
