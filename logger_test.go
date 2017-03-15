package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

var strMarshalError = "Static Marshaling Test Error"

func withTime(t time.Time) OptFunc {
	return func(l *logger) {
		l.ts = &t
	}
}

func withStdErr(b *bytes.Buffer) OptFunc {
	return func(l *logger) {
		l.stderr = b
	}
}

func withStdOut(b *bytes.Buffer) OptFunc {
	return func(l *logger) {
		l.stdout = b
	}
}

func withFatal(l *logger) {
	l.fatal = func(i int) {}
}

func errorMarshal(i interface{}) ([]byte, error) {
	return nil, errors.New(strMarshalError)
}

func TestMultipleOutput(t *testing.T) {

	want := []string{
		"TEST \x1b[32m Info: This is a test \x1b[0m\n",
		"TEST \x1b[33m Warn: This is a test \x1b[0m\n",
		"TEST \x1b[31m Error: This is a test \x1b[0m\n",
		"TEST \x1b[36m Debug: This is a test \x1b[0m\n",
		"TEST \x1b[34m Trace: This is a test \x1b[0m\n",
		"TEST \x1b[31m Fatal: This is a test \x1b[0m\n",
		"TEST \x1b[35m Info: This is a test \x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{})
	l := New(WithOutput(have), WithTimeFormat("TEST"), withFatal)
	for i, lg := range []tt{l.Info, l.Warn, l.Error, l.Debug, l.Trace, l.Fatal} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Logf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(Magenta).Info("This is", "a test")

	x := len(want) - 1
	if want[x] != have.String() {
		t.Logf("\nwant: %q\n\nhave: %q\n", want[x], have.String())
	}
}

func TestMultipleShortOutput(t *testing.T) {

	want := []string{
		"TEST \x1b[32m [INF] This is a test \x1b[0m\n",
		"TEST \x1b[33m [WRN] This is a test \x1b[0m\n",
		"TEST \x1b[31m [ERR] This is a test \x1b[0m\n",
		"TEST \x1b[36m [DBG] This is a test \x1b[0m\n",
		"TEST \x1b[34m [TRC] This is a test \x1b[0m\n",
		"TEST \x1b[31m [FAT] This is a test \x1b[0m\n",
		"TEST \x1b[35m [INF] This is a test \x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{})
	l := New(WithOutput(have), WithShortPrefix, WithTimeFormat("TEST"), withFatal)
	for i, lg := range []tt{l.Info, l.Warn, l.Error, l.Debug, l.Trace, l.Fatal} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Logf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(Magenta).Info("This is", "a test")

	x := len(want) - 1
	if want[x] != have.String() {
		t.Logf("\nwant: %q\n\nhave: %q\n", want[x], have.String())
	}
}

func TestFilteredOutput(t *testing.T) {
	want := "TEST \x1b[32m Info: This is a simple test [1] \x1b[0m\nTEST \x1b[32m Info: This is a simple test [2] \x1b[0m\n"
	have := new(bytes.Buffer)

	fn := func(s string) bool {
		return !strings.Contains(s, "skip")
	}

	l := New(WithFilteredOutput(fn, have), WithTimeFormat("TEST"))

	l.Info("This is a simple test [1]")
	l.Info("This is a simple test that I skip")
	l.Info("This is a simple test [2]")

	if want != have.String() {
		t.Logf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestKVFieldsOutput(t *testing.T) {
	want := []string{
		"TEST \x1b[32m Info: Yes we KV log \x1b[0m basic=3 happy=people quarter=[pound flip]\n",
		"TEST \x1b[32m Info: Yes we KV log \x1b[0m {\"basic\":3,\"happy\":\"people\",\"quarter\":[\"pound\",\"flip\"]}\n",
	}
	have := new(bytes.Buffer)

	l := []Logger{
		New(WithOutput(have), WithTimeFormat("TEST")),
		New(WithOutput(have), WithTimeFormat("TEST"), WithKVMarshaler(json.Marshal)),
	}

	for i := range want {
		have.Reset()
		l[i].Field("happy", "people").Field("basic", 3).Field("quarter", []string{"pound", "flip"}).Info("Yes we KV log")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	for i := range want {
		have.Reset()
		l[i].Fields(KV("happy", "people"), KV("basic", 3), KV("quarter", []string{"pound", "flip"})).Info("Yes we KV log")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	for i := range want {
		have.Reset()
		l[i].Fields(KVMap(KeyValues{
			"happy":   "people",
			"basic":   3,
			"quarter": []string{"pound", "flip"},
		})...).Info("Yes we KV log")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}
}

func TestKVOutput(t *testing.T) {
	want := []string{
		"TEST \x1b[32m Info: Yes we KV log \x1b[0m basic=3 happy=people quarter=[pound flip]\n",
		"TEST \x1b[32m Info: Yes we KV log \x1b[0m {\"basic\":3,\"happy\":\"people\",\"quarter\":[\"pound\",\"flip\"]}\n",
		"TEST \x1b[32m Info: Yes we KV log \x1b[0m [ERR logger.go (marshal)]: map[string]interface {}{\"happy\":\"people\", \"basic\":3, \"quarter\":[]string{\"pound\", \"flip\"}}\n",
	}
	have := new(bytes.Buffer)

	errWant := "error marshaling: " + strMarshalError + "\n"
	errHave := new(bytes.Buffer)

	l := []Logger{
		New(WithOutput(have), WithTimeFormat("TEST")),
		New(WithOutput(have), WithTimeFormat("TEST"), WithKVMarshaler(json.Marshal)),
		New(WithOutput(have), WithTimeFormat("TEST"), WithKVMarshaler(errorMarshal), withStdErr(errHave)),
	}

	for i := range want {
		have.Reset()
		l[i].Info("Yes we KV log", KV("happy", "people"), KV("basic", 3), KV("quarter", []string{"pound", "flip"}))
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	if errWant != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant, errHave.String())
	}
}

func TestOnErrorOutput(t *testing.T) {
	want := "TEST \x1b[32m Info: This is a simple test \x1b[0m\n"
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeFormat("TEST"))

	var err error
	l.OnErr(err).Info("This is a simple test that I an error")
	err = errors.New("simple test")
	l.OnErr(err).Info("This is a", err)

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestMultipleFormattedOutput(t *testing.T) {

	want := []string{
		"TEST \x1b[32m Info: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[33m Warn: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[31m Error: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[36m Debug: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[34m Trace: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[31m Fatal: This is a test 4 0x01 \x1b[0m\n",
		"TEST \x1b[35m Info: This is a test 4 0x01 \x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(string, ...interface{})
	l := New(WithOutput(have), WithTimeFormat("TEST"), withFatal)
	for i, lg := range []tt{l.Infof, l.Warnf, l.Errorf, l.Debugf, l.Tracef, l.Fatalf} {
		have.Reset()
		lg("This is %s %d %#02x", "a test", 4, 1)
		if want[i] != have.String() {
			t.Logf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(Magenta).Infof("This is %s %d %#02x", "a test", 4, 1)

	x := len(want) - 1
	if want[x] != have.String() {
		t.Logf("\nwant: %q\n\nhave: %q\n", want[x], have.String())
	}
}

func TestUTCTime(t *testing.T) {
	want := "Mar-7-1971 17:03:01 \x1b[32m Info: This is a simple test \x1b[0m\n"
	have := new(bytes.Buffer)

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Error(err)
	}
	l := New(withTime(time.Date(1971, time.March, 7, 9, 3, 1, 0, loc)), WithOutput(have), WithTimeAsUTC)

	l.Info("This is a simple test")
	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestSuppressOutput(t *testing.T) {
	want := []string{
		"",
		"TEST \x1b[32m Info: This is a simple test \x1b[0m\n",
	}
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeFormat("TEST"))

	l.Suppress()
	l.Info("This is a simple test that is suppressed")

	if want[0] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[0], have.String())
	}

	l.UnSuppress()
	l.Info("This is a simple test")

	if want[1] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[1], have.String())
	}
}

func TestNetOutput(t *testing.T) {
	want := "TEST \x1b[32m Info: This is a simple test \x1b[0m\n"
	haveCh := make(chan string, 1)

	cont := make(chan struct{}, 1)
	go func() {
		l, err := net.Listen("tcp", ":5550")
		if err != nil {
			t.Log(err)
		}
		cont <- struct{}{} // make sure that the server is started before trying to connect
		for {
			conn, err := l.Accept()
			defer l.Close()

			go func() {
				conn.SetReadDeadline(<-time.After(5 * time.Second))
			}()

			var i int64
			have := new(bytes.Buffer)
			for i < int64(len(want)) {
				i, err = have.ReadFrom(conn)
				if err != nil {
					break
				}
			}

			haveCh <- have.String()
		}
	}()
	<-cont

	l := New(WithOutput(NetWriter("tcp", ":5550")), WithTimeFormat("TEST"))
	l.Info("This", "is a", "simple", "test")

	have := <-haveCh

	if want != have {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have)
	}

	errWant := []string{
		"error writting to log: dial abc: unknown network abc\n",
		"error writting to formatted log: dial abc: unknown network abc\n",
	}
	errHave := new(bytes.Buffer)

	l2 := New(WithOutput(NetWriter("abc", "123")), WithTimeFormat("TEST"), withStdErr(errHave))
	l2.Info("This", "is a", "simple", "test")

	if errWant[0] != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant[0], errHave.String())
	}

	errHave.Reset()
	l2.Infof("This %s %s %s", "is a", "simple", "test")

	if errWant[1] != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant[1], errHave.String())
	}
}

func TestHttpHandlerOutput(t *testing.T) {
	wantHeader, wantHeaderValue := "X-Session-Id", "Test"
	wantBody := `This is a simple test`

	want := "TEST \x1b[32m Info: This is a simple test \x1b[0m X-Session-Id=Test\n"
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithHTTPHeader(wantHeader), WithTimeFormat("TEST"))

	mux := http.NewServeMux()
	mux.Handle("/test/endpoint", l.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(wantHeader, wantHeaderValue)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantBody))
	})))

	req := httptest.NewRequest("GET", "/test/endpoint", nil)
	req.Header.Set(wantHeader, wantHeaderValue)

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	haveBodyBytes := res.Body.String()

	l.Info(wantBody)

	haveBody := string(haveBodyBytes)
	if wantBody != haveBody {
		t.Errorf("\nwant: %q\n\nhave: %q\n", wantBody, haveBody)
	}

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have)
	}

	// No WithHeader first
	l2 := New(WithOutput(have), WithTimeFormat("TEST"))

	mux2 := http.NewServeMux()
	mux2.Handle("/test/endpoint2", l2.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(wantHeader, wantHeaderValue)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantBody))
	})))

	req2 := httptest.NewRequest("GET", "/test/endpoint2", nil)
	req2.Header.Set(wantHeader, wantHeaderValue)

	res2 := httptest.NewRecorder()
	mux2.ServeHTTP(res2, req2)

	l2.Info(wantBody)
}

func TestDefaultStdoutOutput(t *testing.T) {
	want := "TEST \x1b[32m Info: This is a simple test \x1b[0m\n"
	have := new(bytes.Buffer)

	l := New(withStdOut(have), WithTimeFormat("TEST"))
	l.Info("This is a simple test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}

}

func TestConcurrentInfofWarnf(t *testing.T) {
	base := "This is a simple concurrent test (%s) #%02d"
	have := new(bytes.Buffer)

	finish := make(chan struct{}, 1)
	rand.Seed(time.Now().UnixNano())

	l := New(withStdOut(have), WithTimeFormat("TEST"))

	xx := 10
	nn := 4
	go func() {
		for ii := 0; ii < nn; ii++ {
			go func(n, n2 int, done chan struct{}) {
				x := n * n2
				max := x + n2
				for i := x; i < max; i++ {
					l.Infof(base, "info", i)
					<-time.After(time.Millisecond * time.Duration(rand.Intn(10)))
				}
				done <- struct{}{}
			}(ii, xx, finish)
		}
	}()

	go func() {
		for ii := 0; ii < nn; ii++ {
			go func(n, n2 int, done chan struct{}) {
				x := n * n2
				max := x + n2
				for i := x; i < max; i++ {
					l.Warnf(base, "warn", i)
					<-time.After(time.Millisecond * time.Duration(rand.Intn(10)))
				}
				done <- struct{}{}
			}(ii, xx, finish)
		}
	}()

	for ii := 0; ii < nn*2; ii++ {
		<-finish
	}
	close(finish)

	concurrentWant := "\n" + fmt.Sprintf(concurrentWantf, "\x1b[32m", "\x1b[33m", "\x1b[0m")
	haves := strings.Split(have.String(), "\n")
	sort.Strings(haves)
	concurrentHave := strings.Join(haves, "\n")

	if concurrentWant != concurrentHave {
		t.Errorf("\nwant: %q\n\nhave: %q\n", concurrentWant, concurrentHave)
	}
}

func BenchmarkNonFormat(b *testing.B) {
	have := new(bytes.Buffer)
	l := New(withStdOut(have))

	for i := 0; i < b.N; i++ {
		l.Info("Testing with a string of", i)
	}
}

func BenchmarkFormat(b *testing.B) {
	have := new(bytes.Buffer)
	l := New(withStdOut(have))

	for i := 0; i < b.N; i++ {
		l.Infof("Testing with a string of %02d", i)
	}
}

var concurrentWantf = `TEST %[1]s Info: This is a simple concurrent test (info) #00 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #01 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #02 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #03 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #04 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #05 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #06 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #07 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #08 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #09 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #10 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #11 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #12 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #13 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #14 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #15 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #16 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #17 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #18 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #19 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #20 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #21 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #22 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #23 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #24 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #25 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #26 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #27 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #28 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #29 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #30 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #31 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #32 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #33 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #34 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #35 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #36 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #37 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #38 %[3]s
TEST %[1]s Info: This is a simple concurrent test (info) #39 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #00 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #01 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #02 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #03 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #04 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #05 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #06 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #07 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #08 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #09 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #10 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #11 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #12 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #13 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #14 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #15 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #16 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #17 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #18 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #19 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #20 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #21 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #22 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #23 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #24 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #25 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #26 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #27 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #28 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #29 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #30 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #31 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #32 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #33 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #34 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #35 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #36 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #37 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #38 %[3]s
TEST %[2]s Warn: This is a simple concurrent test (warn) #39 %[3]s`
