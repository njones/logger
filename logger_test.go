package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	logg "log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/njones/logger/color"
)

func BenchmarkPrintln(b *testing.B) {
	b.ReportAllocs()
	l := New(WithOutput(ioutil.Discard))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("Testing with a string of", i)
	}
}

func BenchmarkPrintf10KV(b *testing.B) {
	b.ReportAllocs()
	l := New(WithOutput(ioutil.Discard))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// The lucky line below has more parameters than directives because the KV's get taken out.
		l.Printf("Testing with a string of %02d\n", i, KV("hello0", "world"), KV("hello1", "world"), KV("hello2", "world"), KV("hello3", "world"), KV("hello4", "world"), KV("hello5", "world"), KV("hello6", "world"), KV("hello7", "world"), KV("hello8", "world"), KV("hello9", "world"))
	}
}

func BenchmarkFieldsKV10(b *testing.B) {
	b.ReportAllocs()
	l := New(WithOutput(ioutil.Discard))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// The lucky line below has more parameters than directives because the KV's get taken out.
		l.Fields(map[string]interface{}{
			"hello1": "world",
			"hello2": "world",
			"hello3": "world",
			"hello4": "world",
			"hello5": "world",
			"hello6": "world",
			"hello7": "world",
			"hello8": "world",
			"hello9": "world",
			"hello0": "world",
		}).Printf("Testing with a string of %02d\n", i)
	}
}

func BenchmarkPreFieldsKV10(b *testing.B) {
	b.ReportAllocs()
	l := New(WithOutput(ioutil.Discard)).Fields(map[string]interface{}{
		"hello1": "world",
		"hello2": "world",
		"hello3": "world",
		"hello4": "world",
		"hello5": "world",
		"hello6": "world",
		"hello7": "world",
		"hello8": "world",
		"hello9": "world",
		"hello0": "world",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// The lucky line below has more parameters than directives because the KV's get taken out.
		l.Printf("Testing with a string of %02d\n", i)
	}
}

func BenchmarkLogPrintln(b *testing.B) {
	b.ReportAllocs()
	l := NewLog(nil, WithOutput(ioutil.Discard))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("Testing with a string of", i)
	}
}

func TestColor(t *testing.T) {
	have := new(bytes.Buffer)
	log := New(WithOutput(have), WithTimeText("Jan-01-2000"))

	const (
		usePrint = iota
		useDebug
	)

	tests := []struct {
		name   string
		log    interface{}
		prefix string
		inputs []interface{}
		kind   int
		want   string
	}{
		{
			name:   "log.Print (Blue)",
			log:    log.With(WithColor(color.Blue)),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 \x1b[34mabcdefghi\x1b[0m\n",
		},
		{
			name:   "log.Print OnErr (Cyan)",
			log:    log.With(WithColor(color.Cyan)).OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 \x1b[36mabcdefghi\x1b[0m\n",
		},
		{
			name:   "log.Print (NoColor)",
			log:    log.With(WithColor(color.NoColor)),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 abcdefghi\n",
		},
		{
			name:   "log.Debug (Red)",
			kind:   useDebug,
			log:    log.With(WithColor(color.Red)),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 \x1b[31mDEBUG: abcdefghi\x1b[0m\n",
		},
		{
			name:   "log.Debug (NoColor)",
			kind:   useDebug,
			log:    log.With(WithColor(color.NoColor)),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 DEBUG: abcdefghi\n",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			switch l := test.log.(type) {
			case Logger:
				switch test.kind {
				case usePrint:
					l.Print(test.inputs...)
				case useDebug:
					l.Debug(test.inputs...)
				}
			case OnErrLogger:
				switch test.kind {
				case usePrint:
					l.Print(test.inputs...)
				case useDebug:
					l.Debug(test.inputs...)
				}
			}

			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("Jan-01-2000"), withTime(time.Date(1975, 1, 1, 8, 0, 20, 2020, time.Local)))

	var do sync.WaitGroup
	do.Add(1)
	go func() {
		defer do.Done()
		for i := 0; i < 500; i++ {
			do.Add(1)
			go func(i int) {
				defer do.Done()
				time.Sleep(randomDuration())
				l.Printf("[1] This is a test (%03d)", i)
			}(i)
		}
	}()

	do.Add(1)
	go func() {
		defer do.Done()
		for i := 0; i < 500; i++ {
			do.Add(1)
			go func(i int) {
				defer do.Done()
				time.Sleep(randomDuration())
				l.Warnf("[2] This is a test (%03d)", i)
			}(i)
		}
	}()
	do.Wait()

	hav := strings.Split(have.String(), "\n")
	sort.Strings(hav[:len(hav)-1]) // this doesn't sort the last newline

	for i := 0; i < 500; i++ {
		want := fmt.Sprintf("Jan-01-2000 [33mWARN: [2] This is a test (%03d)[0m", i)
		if hav[i] != want {
			t.Fatalf("\nhave: %q\nwant: %q", hav[i], want)
		}
	}
	for i := 0; i < 500; i++ {
		want := fmt.Sprintf("Jan-01-2000 [1] This is a test (%03d)", i)
		if hav[i+500] != want {
			t.Fatalf("\nhave: %q\nwant: %q", hav[i+500], want)
		}
	}
}

func TestDropCR(t *testing.T) {

	w := &dropCRWriter{w: new(bytes.Buffer)}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic",
			input: "this is a test",
			want:  "this is a test",
		},
		{
			name:  "newline in middle",
			input: "this is\na test",
			want:  "this is\na test",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(tt *testing.T) {
			w.w.(*bytes.Buffer).Reset()
			fmt.Fprintln(w, test.input) // auto-add a newline

			{
				have := w.w.(*bytes.Buffer)
				if have.String() != test.want {
					t.Fatalf("\nhave: %q\nwant: %q", have.String(), test.want)
				}
			}

			w.w.(*bytes.Buffer).Reset()

			fmt.Fprint(w, test.input) // doesn't auto-add a newline

			{
				have := w.w.(*bytes.Buffer)
				if have.String() != test.want {
					t.Fatalf("\nhave: %q\nwant: %q", have.String(), test.want)
				}
			}
		})
	}
}

type customFloatFmt float64

func (cf customFloatFmt) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%.4f", float64(cf))
}

func TestFields(t *testing.T) {

	have := new(bytes.Buffer)
	log := New(WithOutput(have), WithTimeText("Jan-01-2000"))
	logField := New(
		WithOutput(have),
		WithTimeText("Jan-01-2000"),
		withTime(time.Date(1975, 1, 1, 0, 0, 0, 0, time.Local)),
	).Field("the", "quick").Field("brown", "fox")

	logFields := New(
		WithOutput(have),
		WithTimeText("Jan-01-2000"),
		withTime(time.Date(1975, 1, 1, 0, 0, 0, 0, time.Local)),
	).Fields(map[string]interface{}{"the": "quick", "brown": "fox"})

	tests := []struct {
		name   string
		method func(...interface{})
		input  []interface{}
		want   string
	}{
		{
			name:   "log KV different value types",
			method: log.Println,
			input:  []interface{}{"This is a", "test for field", KV("the", 0), KV("quick", true), KV("brown", 3.0), KV("fox", customFloatFmt(4.0))},
			want:   "Jan-01-2000 This is a test for field brown=3, fox=4.0000, quick=true, the=0\n",
		},
		{
			name:   "log.Println Field",
			method: logField.Println,
			input:  []interface{}{"This is a", "test for field"},
			want:   "Jan-01-2000 This is a test for field brown=fox, the=quick\n",
		},
		{
			name:   "log.Println Field (KeyVal struct)",
			method: logField.Println,
			input:  []interface{}{"This is a test for field", KeyVal{"jumped", "over"}},
			want:   "Jan-01-2000 This is a test for field brown=fox, jumped=over, the=quick\n",
		},
		{
			name:   "log.Println Field (KeyVal struct) multi insert",
			method: logField.Println,
			input:  []interface{}{KeyVal{"jumped", "over"}, "This is a test for field", KeyVal{"lazy", "dog"}}, //the extra `the` was dropped
			want:   "Jan-01-2000 This is a test for field brown=fox, jumped=over, lazy=dog, the=quick\n",
		},
		{
			name:   "log.Println Field (func KV)",
			method: logField.Println,
			input:  []interface{}{"This is a test for field", KV("jumped", "over")},
			want:   "Jan-01-2000 This is a test for field brown=fox, jumped=over, the=quick\n",
		},
		{
			name:   "log.Println Field (map[K]V)",
			method: logField.Println,
			input:  []interface{}{map[K]V{"jumped": "over", "lazy": "dog"}, "This is a test for field"},
			want:   "Jan-01-2000 This is a test for field brown=fox, jumped=over, lazy=dog, the=quick\n",
		},
		{
			name:   "log.Println Field (KVMap)",
			method: logField.Println,
			input:  []interface{}{KVMap(map[K]V{"jumped": "over", "lazy": "dog"}), "This is a test for field"},
			want:   "Jan-01-2000 This is a test for field brown=fox, jumped=over, lazy=dog, the=quick\n",
		},
		{
			name:   "log.Println Fields",
			method: logFields.Println,
			input:  []interface{}{"This", "is a test", "for fields"},
			want:   "Jan-01-2000 This is a test for fields brown=fox, the=quick\n",
		},
		{
			name:   "log.Println Fields (KeyVal)",
			method: logFields.Println,
			input:  []interface{}{"This is a test for fields", KeyVal{"jumped", "over"}},
			want:   "Jan-01-2000 This is a test for fields brown=fox, jumped=over, the=quick\n",
		},
		{
			name:   "log.Println Fields (KeyVal) multi kind insert",
			method: logFields.Println,
			input:  []interface{}{KV("jumped", "over"), "This is a test", KeyVal{"lazy", "dog"}, "for fields"}, //the extra `the` was dropped
			want:   "Jan-01-2000 This is a test for fields brown=fox, jumped=over, lazy=dog, the=quick\n",
		},
		{
			name:   "log.Println Field JSON marshaler",
			method: logField.With(WithKVMarshaler(json.Marshal)).Println,
			input:  []interface{}{"This is a", "test for field"},
			want:   `Jan-01-2000 This is a test for field {"brown":"fox","the":"quick"}` + "\n",
		},
		{
			name:   "log.Println Field With Error in Marshaler",
			method: logField.With(WithKVMarshaler(errorMarshal)).Println,
			input:  []interface{}{"This is a", "test for field"},
			want:   "",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			test.method(test.input...)
			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestFilename(t *testing.T) {
	have := new(bytes.Buffer)
	log := New(WithOutput(have), withTime(time.Date(1970, 01, 01, 20, 20, 00, int(2020*time.Microsecond), time.UTC)))

	tests := []struct {
		name  string
		flags int
		log   interface{}
		input string
		depth int
		want  *regexp.Regexp
	}{
		{
			name:  "log.Output Short Filename (Timeless)",
			flags: Lshortfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Time{})), // simulate an empty time
			input: "abcdefghi",
			want:  regexp.MustCompile(`\d+-\w+-\d+ line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output Short Filename",
			flags: Lshortfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))),
			input: "abcdefghi",
			want:  regexp.MustCompile(`70-JAN-01 line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output long Filename",
			flags: Llongfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))),
			input: "abcdefghi",
			want:  regexp.MustCompile(`70-JAN-01 (/\w+)+/line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output long Filename",
			flags: Llongfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))),
			input: "abcdefghi",
			depth: 50,
			want:  regexp.MustCompile(`70-JAN-01 \?\?\?:0 abcdefghi\n`),
		},

		{
			name:  "log.Output Short Filename (Timeless) OnErr",
			flags: Lshortfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Time{})).
				OnErr(bytes.ErrTooLarge), // simulate an empty time
			input: "abcdefghi",
			want:  regexp.MustCompile(`\d+-\w+-\d+ line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output Short Filename OnErr",
			flags: Lshortfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))).
				OnErr(bytes.ErrTooLarge),
			input: "abcdefghi",
			want:  regexp.MustCompile(`70-JAN-01 line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output long Filename OnErr",
			flags: Llongfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))).
				OnErr(bytes.ErrTooLarge),
			input: "abcdefghi",
			want:  regexp.MustCompile(`70-JAN-01 (/\w+)+/line.go:\d+ abcdefghi\n`),
		},
		{
			name:  "log.Output long Filename OnErr",
			flags: Llongfile,
			log: log.With(
				WithTimeFormat("06-Jan-02", strings.ToUpper),
				withTime(time.Date(1970, 1, 1, 0, 20, 0, 0, time.UTC))).
				OnErr(bytes.ErrTooLarge),
			input: "abcdefghi",
			depth: 50,
			want:  regexp.MustCompile(`70-JAN-01 \?\?\?:0 abcdefghi\n`),
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			var flags int
			switch l := test.log.(type) {
			case Logger:
				l.SetFlags(test.flags)
				l.Output(test.depth, test.input)
				flags = l.Flags()
			case OnErrLogger:
				l.SetFlags(test.flags)
				l.Output(test.depth, test.input)
				flags = l.Flags()
			}

			if flags != test.flags {
				tt.Errorf("\nhave: %q\nwant: %q\n", flags, test.flags)
			}
			if test.want.FindString(have.String()) == "" {
				tt.Errorf("\n[[ regexp ]]\nhave: %q\nwant: %q\n", have.String(), test.want.String())
			}
			// tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want.String())

			log.(*baseLogger).ts.stamp = defaultTS // reset before the next time
		})
	}
}

func TestFilterWriter(t *testing.T) {
	have := new(bytes.Buffer)
	haveFilter := new(bytes.Buffer)

	fn := func(s string) bool {
		return strings.Contains(s, "whoot")
	}

	tests := []struct {
		name   string
		prefix string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs [][]interface{}
		want   []string
	}{
		{
			name:   "log.Print (filter - string)",
			method: New(WithTimeText(""), WithOutput(have, FilterWriter(haveFilter, StringFuncFilter(fn)))).Print,
			inputs: [][]interface{}{
				{"abc", "def", "ghi"},
				{"whoot, there it is."},
				{"jkl", "mno", "pqr"},
			},
			want: []string{
				"abcdefghi\nwhoot, there it is.\njklmnopqr\n",
				"abcdefghi\njklmnopqr\n",
			},
		},
		{
			name:   "log.Print (filter - regex)",
			method: New(WithTimeText("TEST"), WithOutput(have, FilterWriter(haveFilter, RegexFilter(`..hoot,\st?here`)))).Print,
			inputs: [][]interface{}{
				{"abc", "def", "ghi"},
				{"whoot, there it is"},
				{"whoot, here you are"},
				{"jkl", "mno", "pqr"},
			},
			want: []string{
				"TEST abcdefghi\nTEST whoot, there it is\nTEST whoot, here you are\nTEST jklmnopqr\n",
				"TEST abcdefghi\nTEST jklmnopqr\n",
			},
		}, {
			name:   "log.Print (filter - regex)",
			method: New(WithTimeText("TEST"), WithOutput(have, FilterWriter(haveFilter, NotFilter(RegexFilter(`..hoot,\st?here`))))).Print,
			inputs: [][]interface{}{
				{"abc", "def", "ghi"},
				{"whoot, there it is"},
				{"whoot, here you are"},
				{"jkl", "mno", "pqr"},
			},
			want: []string{
				"TEST abcdefghi\nTEST whoot, there it is\nTEST whoot, here you are\nTEST jklmnopqr\n",
				"TEST whoot, there it is\nTEST whoot, here you are\n",
			},
		},
	}

	for _, test := range tests {
		have.Reset()
		haveFilter.Reset()
		t.Run(test.name, func(tt *testing.T) {
			switch fn := test.method.(type) {
			case func(...interface{}):
				for _, input := range test.inputs {
					fn(input...)
				}
			case func(string, ...interface{}):
				for _, input := range test.inputs {
					fn(test.format, input...)
				}
			default:
				tt.Errorf("\nhave: valid function signature\nhave: %T", fn)
			}

			if have.String() != test.want[0] {
				tt.Errorf("\n(unfiltered)\nhave: %q\nwant: %q\n", have.String(), test.want[0])
			}
			if haveFilter.String() != test.want[1] {
				tt.Errorf("\n(filtered)\nhave: %q\n\nwant: %q\n", haveFilter.String(), test.want[1])
			}
		})
	}
}

func TestHttpHandler(t *testing.T) {
	type data struct {
		name    string
		pattern string
		headers []string
		values  []string
		inputs  []interface{}
		handler func(data) http.HandlerFunc
		logger  func(data, Logger)
		body    string
		want    string
	}

	have := new(bytes.Buffer)

	tests := []data{
		{
			name:    "basic",
			pattern: "/test/endpoint",
			headers: []string{"X-Session-Id"},
			values:  []string{"Test"},
			inputs:  []interface{}{`This is a simple test`},
			handler: func(d data) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					for i, k := range d.headers {
						w.Header().Set(k, d.values[i])
					}
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, d.inputs...)
				}
			},
			body: "This is a simple test",
			want: "192.0.2.1:1234 - - [20/Feb/2020:13:17:56 +0000] \"GET /test/endpoint HTTP/1.1\" 404 21 Testing 123/1.0 X-Session-Id=Test\n",
		},
		{
			name:    "basic with extra log",
			pattern: "/test/endpoint",
			headers: []string{"X-Session-Id"},
			values:  []string{"Test"},
			inputs:  []interface{}{`This is a simple test`},
			handler: func(d data) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					for i, k := range d.headers {
						w.Header().Set(k, d.values[i])
					}
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, d.inputs...)
				}
			},
			logger: func(d data, log Logger) { log.Info(d.inputs...) },
			body:   "This is a simple test",
			want:   "192.0.2.1:1234 - - [20/Feb/2020:13:17:56 +0000] \"GET /test/endpoint HTTP/1.1\" 404 21 Testing 123/1.0 X-Session-Id=Test\n2020-Feb-20 \x1b[32mINFO: This is a simple test\x1b[0m\n",
		},
		{
			name:    "no status code",
			pattern: "/test/endpoint",
			headers: []string{"X-Session-Id"},
			values:  []string{"Test"},
			inputs:  []interface{}{`This is a simple test`},
			handler: func(d data) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					for i, k := range d.headers {
						w.Header().Set(k, d.values[i])
					}
					fmt.Fprint(w, d.inputs...)
				}
			},
			logger: func(d data, log Logger) { log.Info(d.inputs...) },
			body:   "This is a simple test",
			want:   "192.0.2.1:1234 - - [20/Feb/2020:13:17:56 +0000] \"GET /test/endpoint HTTP/1.1\" 200 21 Testing 123/1.0 X-Session-Id=Test\n2020-Feb-20 \x1b[32mINFO: This is a simple test\x1b[0m\n",
		},
	}

	l := New(WithOutput(have), withTime(time.Date(2020, 02, 20, 13, 17, 56, int(4000*time.Microsecond), time.UTC)))

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			have.Reset()

			log := l.With(WithHTTPHeader(test.headers...))

			res := httptest.NewRecorder()
			req := httptest.NewRequest("GET", test.pattern, nil)
			req.Header.Set("User-Agent", "Testing 123/1.0")
			for i, k := range test.headers {
				req.Header.Set(k, test.values[i])
			}

			mux := http.NewServeMux()
			mux.Handle(test.pattern, log.HTTPMiddleware(test.handler(test)))
			mux.ServeHTTP(res, req)

			body := res.Body

			if test.logger != nil {
				test.logger(test, log)
			}

			if body.String() != test.body {
				tt.Fatalf("\nhave: %q\nwant: %q\n", body.String(), test.body)
			}

			if have.String() != test.want {
				tt.Fatalf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestLogLogger(t *testing.T) {
	var log *logg.Logger
	var have = new(bytes.Buffer)

	{
		// THIS IS TESTING WITH A Nil MAP
		want := "Jan-01-2000 This is a test\n"
		log = NewLog(nil, WithOutput(have), WithTimeText("Jan-01-2000"))
		log.Println("This is a test")
		if have.String() != want {
			t.Errorf("\n(nil rxMap)\nhave: %q\nwant: %q\n", have.String(), want)
		}
		have.Reset()
	}

	rxm := map[*regexp.Regexp]func(Logger, map[string]string){
		nil: func(_ Logger, _ map[string]string) {},
		regexp.MustCompile(`\[info\] (?P<data>.*)`): func(l Logger, m map[string]string) {
			l.Info(m["data"])
		},
	}

	log = NewLog(rxm, WithOutput(have), WithTimeText("Jan-01-2000"))

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "None",
			line: "This is a test",
			want: "Jan-01-2000 This is a test\n",
		},
		{
			name: "Info",
			line: "[info] This is a test",
			want: "Jan-01-2000 \x1b[32mINFO: This is a test\x1b[0m\n",
		},
		{
			name: "None With Newline",
			line: "This is a test\n",
			want: "Jan-01-2000 This is a test\n",
		},
	}

	// TESTING WITH A NON-Nil MAP
	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			log.Print(test.line)

			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestNetWriter(t *testing.T) {
	var port = ":2000"
	var setup = make(chan net.Listener)
	var teardown = make(chan struct{}, 1)
	var have = make(chan string, 1)

	go serveTCP(port, have, setup, teardown, t)
	listener := <-setup

	log := New(WithOutput(NetWriter("tcp", port)), WithTimeText("Jan-01-2000"))

	tests := []struct {
		name   string
		method interface{}
		format string
		inputs []interface{}
		want   string
	}{
		{
			name:   "log.Print",
			method: log.Print,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 abcdefghi\n",
		},
		{
			name:   "log.Printf",
			method: log.Printf,
			format: "%[3]s%[2]s%[1]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 ghidefabc\n",
		},
		{
			name:   "log.Debug",
			method: log.Debug,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 DEBUG: abcdefghi\n",
		},
	}

	var n int
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			switch fn := test.method.(type) {
			case func(...interface{}):
				fn(test.inputs...)
			case func(string, ...interface{}):
				fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: %T\nwant: <a valid function signature>", fn)
			}

			Have := <-have
			if Have != test.want {
				tt.Fatalf("\n[[ network ]]\nhave: %q\nwant: %q\n", Have, test.want)
			}
			n++
		})
	}

	if n != len(tests) {
		t.Fatalf("\nhave: %d\nwant: %d\n", n, len(tests))
	}

	close(teardown)
	listener.Close()
}

func TestNetWriterMulti(t *testing.T) {
	var port = ":2000"
	var setup = make(chan net.Listener)
	var teardown = make(chan struct{}, 1)

	type want struct {
		net, log string
	}

	var have = struct {
		log *bytes.Buffer
		net chan string
	}{
		log: new(bytes.Buffer),
		net: make(chan string, 1),
	}

	go serveTCP(port, have.net, setup, teardown, t)
	listener := <-setup

	log := New(WithOutput(NetWriter("tcp", ":2000"), have.log), WithTimeText("Jan-01-2000"))

	tests := []struct {
		name   string
		method interface{}
		format string
		inputs []interface{}
		want   want
	}{
		{
			name:   "log.Print",
			method: log.Print,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: want{
				net: "Jan-01-2000 abcdefghi\n",
				log: "Jan-01-2000 abcdefghi\n",
			},
		},
		{
			name:   "log.Printf",
			method: log.Printf,
			format: "%[3]s%[2]s%[1]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: want{
				net: "Jan-01-2000 ghidefabc\n",
				log: "Jan-01-2000 ghidefabc\n",
			},
		},
		{
			name:   "log.Debug",
			method: log.Debug,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: want{
				log: "Jan-01-2000 \x1b[36mDEBUG: abcdefghi\x1b[0m\n",
				net: "Jan-01-2000 DEBUG: abcdefghi\n",
			},
		},
	}

	var n int
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			have.log.Reset()

			switch fn := test.method.(type) {
			case func(...interface{}):
				fn(test.inputs...)
			case func(string, ...interface{}):
				fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: %T\nwant: <a valid function signature>", fn)
			}

			if have.log.String() != test.want.log {
				tt.Fatalf("\n[[ buffer ]]\nhave: %q\nwant: %q\n", have.log.String(), test.want.log)
			}

			haveNet := <-have.net
			if haveNet != test.want.net {
				tt.Fatalf("\n[[ network ]]\nhave: %q\nwant: %q\n", haveNet, test.want.net)
			}
			n++
		})
	}

	if n != len(tests) {
		t.Fatalf("\nhave: %d\nwant: %d\n", n, len(tests))
	}

	close(teardown)
	listener.Close()
}

func TestNetWriterTimeout(t *testing.T) {
	var port = ":2000"
	var setup = make(chan net.Listener)
	var teardown = make(chan struct{}, 1)

	go serveTCP(port, nil, setup, teardown, t)
	listener := <-setup

	nw := NetWriter("tcp", port, tod(1*time.Nanosecond)).(*netWriter)
	log := New(WithOutput(nw), WithTimeText("Jan-01-2000"))
	log.Println("test")

	err := nw.Err()
	if _, ok := err.(*net.OpError); !ok {
		t.Fatalf("\nhave: %v\nwant: %s\n", err, fmt.Sprintf("dial tcp %s: connect: connection refused", port))
	}

	nw.Close()
	close(teardown)
	listener.Close()
}

func TestNetWriterClose(t *testing.T) {
	var port = ":2000"
	var setup = make(chan net.Listener)
	var teardown = make(chan struct{}, 1)

	go serveTCP(port, nil, setup, teardown, t)
	listener := <-setup

	nw := NetWriter("tcp", port).(*netWriter)
	log := New(WithOutput(nw), WithTimeText("Jan-01-2000"))
	log.Println("test")

	nw.Close()
	if err := nw.Err(); err != nil {
		t.Fatalf("\nhave: %v\nwhat: <nil>\n", err)
	}

	close(teardown)
	listener.Close()
}

func TestOnErrValueExchange(t *testing.T) {
	have := new(bytes.Buffer)

	log := New(WithOutput(have), WithTimeText("Jan-01-2000")).OnErr(bytes.ErrTooLarge)

	tests := []struct {
		name   string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs []interface{}
		want   string
	}{
		{
			name:   "log.Printf",
			method: log.Printf,
			format: "%s%s%s: %v",
			inputs: []interface{}{"abc", "def", "ghi", ErrSub{}},
			want:   "Jan-01-2000 abcdefghi: bytes.Buffer: too large\n",
		},
		{
			name:   "log.Infof",
			method: log.Infof,
			format: "%s %s%[4]s: %[3]v",
			inputs: []interface{}{"abc", "def", ErrSub{}, "ghi"},
			want:   "Jan-01-2000 \x1b[32mINFO: abc defghi: bytes.Buffer: too large\x1b[0m\n",
		},
		{
			name:   "log.Debugf mutli error",
			method: log.Debugf,
			format: "%s%s%s: %v: %v",
			inputs: []interface{}{"abc", "def", "ghi", ErrSub{}, ErrSub{}},
			want:   "Jan-01-2000 \x1b[36mDEBUG: abcdefghi: bytes.Buffer: too large: bytes.Buffer: too large\x1b[0m\n",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			test.method.(func(string, ...interface{}) Return)(test.format, test.inputs...)
			if have.String() != test.want {
				tt.Errorf("\nhave: %q\n\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestOutputEmpty(t *testing.T) {
	log := New(WithTimeText("Jan-01-2000"), WithOutput()) // use stdout

	have := fmt.Sprintf("%p", log.Writer())
	want := fmt.Sprintf("%p", os.Stdout)

	if have != want {
		t.Fatalf("\nhave: %s\nwant: %s\n", have, want)
	}

	// check the we provide a multi-writer
	b1, b2 := new(bytes.Buffer), new(bytes.Buffer)
	log2 := New(WithOutput(b1, b2)).OnErr(bytes.ErrTooLarge)

	w := log2.Writer()
	have2 := fmt.Sprintf("%p", w) // this should be a multiwriter now
	want2 := fmt.Sprintf("%p", os.Stdout)

	if have2 == want2 { // they should not equal each other, it's a problem if we pickup the stdout
		t.Fatalf("\nhave: %s\nwant: %s\n", have, want)
	}

	w.Write([]byte("This is a test"))

	if b1.String() != b2.String() {
		t.Fatalf("\nhave: %s\nwant: %s\n", b1.String(), b2.String())
	}
}

func TestPanic(t *testing.T) {
	have := new(bytes.Buffer)

	log := New(WithOutput(have), WithTimeText("Jan-01-2000"))
	logOnErrYes := New(WithOutput(have), WithTimeText("Jan-01-2000")).OnErr(bytes.ErrTooLarge)
	logOnErrNil := New(WithOutput(have), WithTimeText("Jan-01-2000")).OnErr(nil)

	type want struct {
		log        string
		rtn, panic bool
	}
	tests := []struct {
		name   string
		method interface{}
		format string
		inputs []interface{}
		want   want
	}{
		{
			name:   "log.Panic",
			method: log.Panic,
			inputs: []interface{}{"This i", "s a test"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", panic: true},
		},
		{
			name:   "log.Panicf",
			method: log.Panicf,
			format: "%[2]s%[1]s",
			inputs: []interface{}{"s a test", "This i"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", panic: true},
		},
		{
			name:   "log.Panicln",
			method: log.Panicln,
			inputs: []interface{}{"This is", "a test"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", panic: true},
		},
		{
			name:   "log.Panic (onErr)",
			method: logOnErrYes.Panic,
			inputs: []interface{}{"This i", "s a test"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", rtn: true, panic: true},
		},
		{
			name:   "log.Panicf (onErr)",
			method: logOnErrYes.Panicf,
			format: "%[2]s%[1]s",
			inputs: []interface{}{"s a test", "This i"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", rtn: true, panic: true},
		},
		{
			name:   "log.Panicln (onErr)",
			method: logOnErrYes.Panicln,
			inputs: []interface{}{"This is", "a test"},
			want:   want{log: "Jan-01-2000 PANIC: This is a test\n", rtn: true, panic: true},
		},
		{
			name:   "log.Panic (onErr=nil)",
			method: logOnErrNil.Panic,
			inputs: []interface{}{"This i", "s a test"},
			want:   want{log: "", rtn: false, panic: false},
		},
		{
			name:   "log.Panicf (onErr=nil)",
			method: logOnErrNil.Panicf,
			format: "%s[2]%s[1]",
			inputs: []interface{}{"s a test", "This i"},
			want:   want{log: "", rtn: false, panic: false},
		},
		{
			name:   "log.Panicln (onErr=nil)",
			method: logOnErrNil.Panicln,
			inputs: []interface{}{"This is", "a test"},
			want:   want{log: "", rtn: false, panic: false},
		},
	}

	var recoverNum int
	for _, test := range tests {
		var hasRecovered bool
		have.Reset()

		t.Run(test.name, func(tt *testing.T) {
			defer func() {
				if test.want.panic && !hasRecovered {
					tt.Errorf("\n[[ recovery ]] no panic detected\n")
				}
			}()
			defer func() {
				recoverNum++
				if haver := recover(); haver != nil {
					if haver != test.want.log {
						tt.Errorf("\n[[ recovery ]]\nhave %q\n\nwant: %q\n", haver, test.want.log)
					}
					hasRecovered = true
				}
			}()

			switch fn := test.method.(type) {
			case func(...interface{}):
				fn(test.inputs...)
			case func(string, ...interface{}):
				fn(test.format, test.inputs...)
			case func(...interface{}) Return:
				fn(test.inputs...)

			case func(string, ...interface{}) Return:
				fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: %T\nwant: <a valid function signature>", fn)
			}
			if have.String() != test.want.log {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want.log)
			}
		})
	}

	if recoverNum != len(tests) {
		t.Fatalf("\nhave: %d\nwant: %v\n", recoverNum, len(tests))
	}
}

func TestPrefix(t *testing.T) {
	have := new(bytes.Buffer)
	log := New(WithTimeText("Jan-01-2000")) // using log.SetOutput later...

	tests := []struct {
		name   string
		prefix string
		log    interface{}
		inputs []interface{}
		want   string
	}{
		{
			name:   "log",
			log:    log,
			prefix: "[Test Prefix]",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 [Test Prefix] abcdefghi\n",
		},
		{
			name:   "log.OnErr",
			log:    log.OnErr(bytes.ErrTooLarge),
			prefix: "[Test Prefix]",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 [Test Prefix] abcdefghi\n",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			var prefix string
			var writer string
			switch l := test.log.(type) {
			case Logger:
				l.SetPrefix(test.prefix)
				l.SetOutput(have)
				l.Print(test.inputs...)
				prefix = l.Prefix()
				writer = fmt.Sprintf("%p", l.Writer())
			case OnErrLogger:
				l.SetPrefix(test.prefix)
				l.SetOutput(have)
				l.Print(test.inputs...)
				prefix = l.Prefix()
				writer = fmt.Sprintf("%p", l.Writer())
			}

			if prefix != test.prefix {
				tt.Errorf("\nhave: %q\nwant: %q\n", prefix, test.prefix)
			}
			if fmt.Sprintf("%p", have) != writer {
				tt.Errorf("\nhave: %q\nwant: %q\n", fmt.Sprintf("%p", have), writer)
			}
			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestSupress(t *testing.T) {
	have := new(bytes.Buffer)
	log := New(WithOutput(have), WithTimeText("Jan-01-2000"))

	tests := []struct {
		name     string
		suppress logLevel
		want     string
	}{
		{
			name:     "Suppress Info",
			suppress: Info,
			want:     "Jan-01-2000 \x1b[33mWARN: Pack my box with five dozen liquor jugs\x1b[0m\nJan-01-2000 \x1b[36mDEBUG: Cozy lummox gives smart squid who asks for job pen\x1b[0m\n",
		},
		{
			name:     "Suppress Debug and Warn",
			suppress: Debug | Warn,
			want:     "Jan-01-2000 \x1b[32mINFO: The quick brown fox jumped over the lazy dog\x1b[0m\n",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {

			log.Suppress(test.suppress)
			log.Info("The quick brown fox jumped over the lazy dog")
			log.Warn("Pack my box with five dozen liquor jugs")
			log.Debug("Cozy lummox gives smart squid who asks for job pen")

			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}
		})
	}
}

func TestTime(t *testing.T) {
	PDT, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	have := new(bytes.Buffer)
	log := New(WithOutput(have), withTime(time.Date(1970, 01, 01, 20, 20, 00, int(2020*time.Microsecond), time.UTC)))

	tests := []struct {
		name   string
		flags  int
		log    interface{}
		inputs []interface{}
		want   string
	}{
		{
			name:   "log.Print Standard",
			log:    log,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970-Jan-01 abcdefghi\n",
		},
		{
			name:   "log.Print LstdFlags",
			flags:  LstdFlags,
			log:    log,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970/01/01 20:20:00 abcdefghi\n",
		},
		{
			name:   "log.Print (LstdFlags | Lmicroseconds)",
			flags:  LstdFlags | Lmicroseconds,
			log:    log,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970/01/01 20:20:00.002020 abcdefghi\n",
		},
		{
			name:   "log.Print (time format)",
			log:    log.With(WithTimeFormat("Jan 02 2006 (03:04pm)")),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20pm) abcdefghi\n",
		},
		{
			name: "log.Print (time format am works)",
			log: log.With(
				WithTimeFormat("Jan 02 2006 (03:04pm)"),
				withTime(time.Date(1970, 1, 1, 8, 20, 00, int(2020*time.Microsecond), time.UTC))),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20am) abcdefghi\n",
		},
		{
			name:   "log.Print (time format override flags)",
			flags:  LstdFlags,
			log:    log.With(WithTimeFormat("Jan 02 2006 (03:04pm)")),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20pm) abcdefghi\n",
		},
		{
			name:  "log.Print UTC Flag",
			flags: LstdFlags | LUTC,
			log: log.With(
				WithTimeFormat("Jan 02 2006 (03:04pm)"),
				withTime(time.Date(1970, 1, 1, 0, 20, 00, int(2020*time.Microsecond), PDT))),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20am) abcdefghi\n",
		},
		{
			name:  "log.Print Empty Format",
			flags: LstdFlags,
			log: log.With(
				WithTimeFormat(""),
				withTime(time.Date(1970, 1, 1, 0, 20, 00, int(2020*time.Microsecond), time.UTC))),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "abcdefghi\n",
		},
		{
			name:   "log.Print OnErr Standard",
			log:    log.OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970-Jan-01 abcdefghi\n",
		},
		{
			name:   "log.Print OnErr (LstdFlags)",
			flags:  LstdFlags,
			log:    log.OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970/01/01 20:20:00 abcdefghi\n",
		},
		{
			name:   "log.Print OnErr (LstdFlags | Lmicroseconds)",
			flags:  LstdFlags | Lmicroseconds,
			log:    log.OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "1970/01/01 20:20:00.002020 abcdefghi\n",
		},
		{
			name:   "log.Print OnErr (time format)",
			log:    log.With(WithTimeFormat("Jan 01 2006 (03:04pm)")).OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20pm) abcdefghi\n",
		},
		{
			name: "log.Print OnErr (time format am works) OnErr",
			log: log.With(
				WithTimeFormat("Jan 02 2006 (03:04pm)"),
				withTime(time.Date(1970, 1, 1, 8, 20, 00, int(2020*time.Microsecond), time.UTC))).
				OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20am) abcdefghi\n",
		},
		{
			name:   "log.Print OnErr (time format override flags)",
			flags:  LstdFlags,
			log:    log.With(WithTimeFormat("Jan 02 2006 (03:04pm)")).OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20pm) abcdefghi\n",
		},
		{
			name:  "log.Print OnErr UTC Flag",
			flags: LstdFlags | LUTC,
			log: log.With(
				WithTimeFormat("Jan 02 2006 (03:04pm)"),
				withTime(time.Date(1970, 1, 1, 0, 20, 00, int(2020*time.Microsecond), PDT))).
				OnErr(bytes.ErrTooLarge),
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan 01 1970 (08:20am) abcdefghi\n",
		},
	}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			var flags int
			switch l := test.log.(type) {
			case Logger:
				l.SetFlags(test.flags)
				l.Print(test.inputs...)
				flags = l.Flags()
			case OnErrLogger:
				l.SetFlags(test.flags)
				l.Print(test.inputs...)
				flags = l.Flags()
			}

			if flags != test.flags {
				tt.Errorf("\nhave: %q\nwant: %q\n", flags, test.flags)
			}
			if have.String() != test.want {
				tt.Errorf("\nhave: %q\nwant: %q\n", have.String(), test.want)
			}

			log.(*baseLogger).ts.stamp = defaultTS // reset before the next time
		})
	}
}

func withTime(t time.Time) optFunc {
	return func(b *baseLogger) {
		b.ts.now = t
	}
}

func randomDuration() time.Duration {
	return time.Duration(rand.Intn(10)) * time.Millisecond
}

func errorMarshal(v interface{}) ([]byte, error) {
	return nil, bytes.ErrTooLarge
}

func serveTCP(port string, have chan string, setup chan net.Listener, teardown chan struct{}, t *testing.T) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatal(err)
	}
	setup <- listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-teardown:
				// nothing
			default:
				t.Error(err)
			}
			return
		}

		go func(c net.Conn) {
			scn := bufio.NewScanner(c)
			for scn.Scan() {
				have <- scn.Text() + "\n"
			}
			c.Close()
		}(conn)
	}
}
