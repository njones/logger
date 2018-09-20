package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var strMarshalError = "Static Marshaling Test Error"

func withStdErr(b io.Writer) optFunc {
	return func(l *baseLogger) {
		l.stderr = b
	}
}

// WithTime overwrites the current timestamp with a static time stamp
func withTime(ts time.Time) optFunc {
	return func(l *baseLogger) {
		l.ts = &ts
	}
}

func withStdOut(b *bytes.Buffer) optFunc {
	return func(l *baseLogger) {
		l.stdout = b
	}
}

func withFatal(l *baseLogger) {
	l.fatal = func(i int) {}
}

func errorMarshal(i interface{}) ([]byte, error) {
	return nil, errors.New(strMarshalError)
}

func TestWithPrefix(t *testing.T) {

	want := []string{
		"TEST \x1b[32mInfo: [prefix included] This is a test\x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{}) Return
	l := New(WithOutput(have), WithTimeText("TEST"), WithPrefix("[prefix included]"))
	for i, lg := range []tt{l.Println} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}
}

func TestMultipleOutput(t *testing.T) {

	want := []string{
		"TEST \x1b[32mInfo: This is a test\x1b[0m\n",
		"TEST \x1b[32mInfo: This is a test\x1b[0m\n",
		"TEST \x1b[33mWarn: This is a test\x1b[0m\n",
		"TEST \x1b[35mError: This is a test\x1b[0m\n",
		"TEST \x1b[36mDebug: This is a test\x1b[0m\n",
		"TEST \x1b[34mTrace: This is a test\x1b[0m\n",
		"TEST \x1b[31mFatal: This is a test\x1b[0m\n",
		"TEST \x1b[35mInfo: This is a test\x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{}) Return
	l := New(WithOutput(have), WithTimeText("TEST"), withFatal)
	for i, lg := range []tt{l.Println, l.Infoln, l.Warnln, l.Errorln, l.Debugln, l.Traceln, l.Fatalln} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(ColorMagenta).Println("This is", "a test")

	if want[len(want)-1] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[0], have.String())
	}
}

func TestMultipleShortOutput(t *testing.T) {
	want := []string{
		"TEST \x1b[32m[INF] This is a test\x1b[0m\n",
		"TEST \x1b[33m[WRN] This is a test\x1b[0m\n",
		"TEST \x1b[35m[ERR] This is a test\x1b[0m\n",
		"TEST \x1b[36m[DBG] This is a test\x1b[0m\n",
		"TEST \x1b[34m[TRC] This is a test\x1b[0m\n",
		"TEST \x1b[31m[FAT] This is a test\x1b[0m\n",
		"TEST \x1b[31m[FAT] This isa test\x1b[0m\n",
		"TEST \x1b[35m[INF] This is a test\x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{}) Return
	l := New(WithOutput(have), WithLevelPrefix(LevelShortBracketStr), WithTimeText("TEST"), withFatal)
	for i, lg := range []tt{l.Infoln, l.Warnln, l.Errorln, l.Debugln, l.Traceln, l.Fatalln, l.Fatal} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(ColorMagenta).Infoln("This is", "a test")

	x := len(want) - 1
	if want[x] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[x], have.String())
	}
}

func TestFatalInt(t *testing.T) {
	ch := make(chan int, 1)
	withFatalInt := func(l *baseLogger) {
		l.fatal = func(i int) { ch <- i }
	}

	want := 4

	l := New(WithOutput(ioutil.Discard), withFatalInt)
	l.FatalInt(want).Fatal("Doesn't Matter...")

	have := <-ch
	if want != have {
		t.Errorf("\nwant: %d have: %d\n", want, have)
	}
}

func TestWithOutput(t *testing.T) {
	want1 := []string{
		"TEST \x1b[32m[INF] This is a test\x1b[0m\n",
		"TEST \x1b[33m[WRN] This is a test\x1b[0m\n",
		"TEST \x1b[35m[ERR] This is a test\x1b[0m\n",
		"TEST \x1b[36m[DBG] This is a test\x1b[0m\n",
		"TEST \x1b[34m[TRC] This is a test\x1b[0m\n",
		"TEST \x1b[31m[FAT] This is a test\x1b[0m\n",

		"TEST \x1b[35m[INF] This is a test\x1b[0m\n",
	}

	want2 := []string{
		"TEST \x1b[32mINF This is a test\x1b[0m\n",
		"TEST \x1b[33mWRN This is a test\x1b[0m\n",
		"TEST \x1b[35mERR This is a test\x1b[0m\n",
		"TEST \x1b[36mDBG This is a test\x1b[0m\n",
		"TEST \x1b[34mTRC This is a test\x1b[0m\n",
		"TEST \x1b[31mFAT This is a test\x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{}) Return
	l := New(WithOutput(have), WithLevelPrefix(LevelShortBracketStr), WithTimeText("TEST"), withFatal)
	l2 := l.With(WithLevelPrefix(LevelShortStr))
	for i, lg := range []tt{l.Infoln, l.Warnln, l.Errorln, l.Debugln, l.Traceln, l.Fatalln} {
		have.Reset()
		lg("This is", "a test")
		if want1[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want1[i], have.String())
		}
	}
	for i, lg := range []tt{l2.Infoln, l2.Warnln, l2.Errorln, l2.Debugln, l2.Traceln, l2.Fatalln} {
		have.Reset()
		lg("This is", "a test")
		if want2[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want2[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(ColorMagenta).Infoln("This is", "a test")

	x := len(want1) - 1
	if want1[x] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want1[x], have.String())
	}
}

func TestPanicPrint(t *testing.T) {
	want := []string{
		"TEST \x1b[31m[PAN] This isa test\x1b[0m",
	}

	var r interface{}
	defer func() {
		if _, ok := r.(*string); ok {
			t.Errorf("The panic was never fired")
		}
	}()
	defer func() {
		if r = recover(); r != nil {
			x := len(want) - 1
			if want[x] != r {
				t.Errorf("\nwant: %q\n\nhave: %q\n", want[x], r)
			}
		}
	}()

	type tt func(...interface{}) Return
	l := New(WithOutput(ioutil.Discard), WithLevelPrefix(LevelShortBracketStr), WithTimeText("TEST"))
	for _, lg := range []tt{l.Panic} {
		lg("This is", "a test")
	}
}

func TestPanicPrintf(t *testing.T) {
	want := []string{
		"TEST \x1b[31m[PAN] This is a test\x1b[0m",
	}

	var r interface{}
	defer func() {
		if _, ok := r.(*string); ok {
			t.Errorf("The panic was never fired")
		}
	}()
	defer func() {
		if r = recover(); r != nil {
			x := len(want) - 1
			if want[x] != r {
				t.Errorf("\nwant: %q\n\nhave: %q\n", want[x], r)
			}
		}
	}()

	type tt func(string, ...interface{}) Return
	l := New(WithOutput(ioutil.Discard), WithLevelPrefix(LevelShortBracketStr), WithTimeText("TEST"))
	for _, lg := range []tt{l.Panicf} {
		lg("%s %s", "This is", "a test")
	}
}

func TestPanicPrintln(t *testing.T) {
	want := []string{
		"TEST \x1b[31m[PAN] This is a test\x1b[0m",
	}

	var r interface{}
	defer func() {
		if _, ok := r.(*string); ok {
			t.Errorf("The panic was never fired")
		}
	}()
	defer func() {
		if r = recover(); r != nil {
			x := len(want) - 1
			if want[x] != r {
				t.Errorf("\nwant: %q\n\nhave: %q\n", want[x], r)
			}
		}
	}()

	type tt func(...interface{}) Return
	l := New(WithOutput(ioutil.Discard), WithLevelPrefix(LevelShortBracketStr), WithTimeText("TEST"))
	for _, lg := range []tt{l.Panicln} {
		lg("This is", "a test")
	}
}

func TestMultipleNoColorOutput(t *testing.T) {

	want := []string{
		"TEST Info: This isa test\n",
		"TEST Info: This is a test\n",
		"TEST Warn: This is a test\n",
		"TEST Error: This is a test\n",
		"TEST Debug: This is a test\n",
		"TEST Trace: This is a test\n",
		"TEST Fatal: This is a test\n",
	}

	have := new(bytes.Buffer)

	type tt func(...interface{}) Return
	l := New(WithOutput(have), WithTimeText("TEST"), withFatal)
	for i, lg := range []tt{
		l.NoColor().Print,
		l.NoColor().Infoln,
		l.NoColor().Warnln,
		l.NoColor().Errorln,
		l.NoColor().Debugln,
		l.NoColor().Traceln,
		l.NoColor().Fatalln,
	} {
		have.Reset()
		lg("This is", "a test")
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}
}

func TestMultipleFormattedOutput(t *testing.T) {

	want := []string{
		"TEST \x1b[32mInfo: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[32mInfo: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[33mWarn: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[35mError: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[36mDebug: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[34mTrace: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[31mFatal: This is a test 4 0x01\x1b[0m\n",
		"TEST \x1b[35mInfo: This is a test 4 0x01\x1b[0m\n",
	}

	have := new(bytes.Buffer)

	type tt func(string, ...interface{}) Return
	l := New(WithOutput(have), WithTimeText("TEST"), withFatal)
	for i, lg := range []tt{l.Printf, l.Infof, l.Warnf, l.Errorf, l.Debugf, l.Tracef, l.Fatalf} {
		have.Reset()
		lg("This is %s %d %#02x", "a test", 4, 1)
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	// check if can use color (but don't evaluate before other tests)
	have.Reset()
	l.Color(ColorMagenta).Infof("This is %s %d %#02x", "a test", 4, 1)

	x := len(want) - 1
	if want[x] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[x], have.String())
	}
}

func TestFilteredStringOutput(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test [1]\x1b[0m\nTEST \x1b[32mInfo: This is a simple test [2]\x1b[0m\n"
	have := new(bytes.Buffer)

	fn := func(s string) bool {
		return strings.Contains(s, "skip")
	}

	l := New(WithFilterOutput(have, StringFuncFilter(fn)), WithTimeText("TEST"))

	l.Info("This is a simple test [1]")
	l.Info("This is a simple test that I skip")
	l.Info("This is a simple test [2]")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNegativeFilteredStringOutput(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test that I skip NOT!\x1b[0m\n"
	have := new(bytes.Buffer)

	fn := func(s string) bool {
		return strings.Contains(s, "skip")
	}

	l := New(WithFilterOutput(have, NotFilter(StringFuncFilter(fn))), WithTimeText("TEST"))

	l.Info("This is a simple test [1]")
	l.Info("This is a simple test that I skip NOT!")
	l.Info("This is a simple test [2]")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestFilteredStringOutput2(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test that I keep\x1b[0m\n"
	have := new(bytes.Buffer)

	// This shows that we're looking at the data passed in and not including the time or color stuff
	fn := func(s string) bool {
		if strings.HasSuffix(s, "[1]") {
			return true
		}
		if strings.HasPrefix(s, "[3]") {
			return true
		}
		return false
	}

	l := New(WithFilterOutput(have, StringFuncFilter(fn)), WithTimeText("TEST"))

	l.Info("This is a simple test [1]")
	l.Info("This is a simple test that I keep")
	l.Infof("%s [%d]", "This is a simple test that I skip", 1)
	l.Infof("[%d] %s", KV("skip", '⇨'), 3, "This is a simple test that I skip", KV("skip", '←'))
	l.Info("[3] This is a simple test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestFilteredOutput3(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test that I keep\x1b[0m skip=8680\n"
	have := new(bytes.Buffer)

	l := New(WithFilterOutput(have, RegexFilter(`(\[\d+\]|I\s*sk.p)`)), WithTimeText("TEST"))

	l.Info("This is a simple test [1]")
	l.Info("This is a simple test that I keep", KV("skip", '⇨'))
	l.Infof("%s [%d]", "This is a simple test that I skip", 1)
	l.Infof("[%d] %s", KV("skip", '⇨'), 3, "This is a simple test that I skip", KV("skip", '←'))
	l.Info("[3] This is a simple test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestKVFieldsOutput(t *testing.T) {
	want := []string{
		"TEST \x1b[32mInfo: Yes we KV log\x1b[0m basic=3 happy=people quarter=[pound flip]\n",
		"TEST \x1b[32mInfo: Yes we KV log\x1b[0m {\"basic\":3,\"happy\":\"people\",\"quarter\":[\"pound\",\"flip\"]}\n",
	}
	have := new(bytes.Buffer)

	l := []Logger{
		New(WithOutput(have), WithTimeText("TEST")),
		New(WithOutput(have), WithTimeText("TEST"), WithKVMarshaler(json.Marshal)),
	}

	for i := range want {
		have.Reset()
		l[i].Field("happy", "people").
			Field("basic", 3).
			Field("quarter", []string{"pound", "flip"}).
			Info("Yes we KV log")

		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	for i := range want {
		have.Reset()
		l[i].Fields(map[string]interface{}{"happy": "people", "basic": 3}).
			Fields(map[string]interface{}{"quarter": []string{"pound", "flip"}}).
			Info("Yes we KV log")

		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}
}

func TestInlineKVOutput(t *testing.T) {
	want := []string{
		"TEST \x1b[32mInfo: Yes we KV log\x1b[0m basic=3 happy=people quarter=[pound flip]\n",
		"TEST \x1b[32mInfo: Yes we KV log\x1b[0m {\"basic\":3,\"happy\":\"people\",\"quarter\":[\"pound\",\"flip\"]}\n",
		"TEST \x1b[32mInfo: Yes we KV log\x1b[0m [ERR logger.go (marshal)]: map[string]interface {}{\"happy\":\"people\", \"basic\":3, \"quarter\":[]string{\"pound\", \"flip\"}}\n",
	}
	have := new(bytes.Buffer)

	// what will be output to stderr...
	errWant := "error marshaling: " + strMarshalError + "\n"
	errHave := new(bytes.Buffer)

	l := []Logger{
		New(WithOutput(have), WithTimeText("TEST")),
		New(WithOutput(have), WithTimeText("TEST"), WithKVMarshaler(json.Marshal)),
		New(WithOutput(have), WithTimeText("TEST"), WithKVMarshaler(errorMarshal), withStdErr(errHave)),
	}

	for i := range want {
		have.Reset()
		l[i].Print("Yes we KV log", KV("happy", "people"), KV("basic", 3), KV("quarter", []string{"pound", "flip"}))

		// HARDCODED //
		// if this is an error marshaling then do a different type of compare, because the map can be randomized
		if i == 2 { // 2 is the third element of the array change this if things change
			sep := "interface {}"
			haveSplit := strings.Split(have.String(), sep)
			wantSplit := strings.Split(want[i], sep)

			if wantSplit[0] != haveSplit[0] {
				t.Errorf("\nwant[a]: %q\n\nhave[a]: %q\n", want[i], have.String())
			}

			cutset := "{,}\n"
			haveSplitMap := strings.Split(strings.Trim(haveSplit[1], cutset), " ")
			wantSplitMap := strings.Split(strings.Trim(wantSplit[1], cutset), " ")

			sort.Strings(haveSplitMap)
			sort.Strings(wantSplitMap)

			for j := range wantSplitMap {
				if strings.Trim(wantSplitMap[j], cutset) != strings.Trim(haveSplitMap[j], cutset) {
					t.Errorf("\nwant[%d]: %q\n\nhave[%d]: %q\n", j, want[i], j, have.String())
					t.Errorf("\nwant[%d]: %q\n\nhave[%d]: %q\n", j, wantSplitMap[j], j, haveSplitMap[j])
				}
			}
			continue
		}
		if want[i] != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want[i], have.String())
		}
	}

	if errWant != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant, errHave.String())
	}
}

func TestOnErrorOutput(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test\x1b[0m\n"
	wantErrCheck1 := false
	wantErrCheck2 := true
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeText("TEST"))

	var err error
	haveErrCheck1 := l.OnErr(err).Info("This is a simple test that is NOT an error")
	err = errors.New("simple test")
	haveErrCheck2 := l.OnErr(err).Infoln("This is a", err)

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}

	if wantErrCheck1 != haveErrCheck1.HasErr {
		t.Errorf("\nwant: %t\n\nhave: %t\n", wantErrCheck1, haveErrCheck1.HasErr)
	}

	if wantErrCheck2 != haveErrCheck2.HasErr {
		t.Errorf("\nwant: %t\n\nhave: %t\n", wantErrCheck2, haveErrCheck2.HasErr)
	}
}

func TestUTCTime(t *testing.T) {
	want := "Mar-6-1974 16:03:01 \x1b[32mInfo: This is a simple test\x1b[0m\n"
	have := new(bytes.Buffer)

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Error(err)
	}
	l := New(withTime(time.Date(1974, time.March, 6, 9, 3, 1, 0, loc)), WithOutput(have), WithTimeAsUTC())

	l.Info("This is a simple test")
	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}

	want2 := "Mar-27-1976 17:03:01 \x1b[32mInfo: This is a simple test\x1b[0m\n"
	have2 := new(bytes.Buffer)

	loc2, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Error(err)
	}
	l2 := New(withTime(time.Date(1976, time.March, 27, 9, 3, 1, 0, loc2)), WithOutput(have2), WithTimeAsUTC())

	l2.Info("This is a simple test")
	if want2 != have2.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want2, have2.String())
	}
}

func TestTimeNow(t *testing.T) {
	want := fmt.Sprintf("%s \x1b[32mInfo: This is a simple test\x1b[0m\n", time.Now().Format(time.RFC3339))
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeFormat(time.RFC3339))

	l.Info("This is a simple test")
	if want != have.String() {
		// turn back a sec and see if we crossed a boundary
		seconds := want[17:19]
		i, err := strconv.Atoi(seconds)
		if err != nil {
			t.Errorf("error parsing seconds: %v", err)
			return
		}
		i--
		want2 := fmt.Sprintf("%s%d%s", want[:17], i, want[19:])
		if want2 != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
			t.Errorf("\nwant: %q\n\nhave: %q\n", want2, have.String())
		}
	}
}

func TestSuppressOutput(t *testing.T) {
	want := []string{
		"",
		"TEST \x1b[32mInfo: This is a simple test\x1b[0m\n",
		"TEST \x1b[32mInfo: This is a simple test that is not suppressed\x1b[0m\n",
		"TEST \x1b[33mWarn: This is a simple test that won't be suppressed\x1b[0m\nTEST \x1b[32mInfo: This is a simple test\x1b[0m\n",
	}
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeText("TEST"))

	l.Suppress(LevelInfo).Infoln("This is a simple test that is suppressed")

	if want[0] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[0], have.String())
	}

	l.Info("This is a simple test")

	if want[1] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[1], have.String())
	}

	have.Reset()
	l.Suppress(LevelInfo).Println("This is a simple test that is not suppressed")

	if want[2] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[2], have.String())
	}

	have.Reset()
	l2 := l.Suppress(LevelInfo | LevelDebug)
	l2.Infoln("This is a simple test that is suppressed")
	l2.Warnln("This is a simple test that won't be suppressed")
	l2.Debugln("This is a simple test that is suppressed")
	l.Infoln("This is a simple test")

	if want[3] != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want[3], have.String())
	}
}

func TestNetOutput(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test\x1b[0m\n"
	errHave := new(bytes.Buffer)

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
				<-time.After(5 * time.Second)
				conn.Close()
			}()

			var i int64
			have := new(bytes.Buffer)
			mx := 100
			m := 0
			for i < int64(len(want)) {
				i, err = io.Copy(have, conn)
				if i == 0 {
					var bb = make([]byte, 1)
					conn.Read(bb)
					fmt.Print(".")
					if m >= mx {
						t.Log("breaking... nil")
						break
					}
					m++
				}
				if err != nil {
					t.Log("breaking...", err)
					break
				}
			}
			haveCh <- have.String()
		}
	}()
	<-cont

	errWant1 := ""

	l := New(WithOutput(NetWriter("tcp", ":5550")), WithTimeText("TEST"), withStdErr(errHave))
	l.Infoln("This", "is a", "simple", "test")

	have := <-haveCh

	if want != have {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have)
	}

	if errWant1 != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant1, errHave.String())
	}

	errWant2 := []string{
		"error writing to log: dial abc: unknown network abc\n",
		"error writing to log: dial abc: unknown network abc\n",
		"error writing to log: dial abc: unknown network abc\n",
	}

	errHave.Reset()
	l2 := New(WithOutput(NetWriter("abc", "123")), WithTimeText("TEST"), withStdErr(errHave))
	l2.Infoln("This", "is a", "simple", "test")

	if errWant2[0] != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant2[0], errHave.String())
	}

	errHave.Reset()
	l2.Infof("This %s %s %s", "is a", "simple", "test")

	if errWant2[1] != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant2[1], errHave.String())
	}

	errHave.Reset()
	l2.Print("This is a simple test")

	if errWant2[2] != errHave.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", errWant2[2], errHave.String())
	}
}

func TestHttpHandlerOutput(t *testing.T) {
	wantHeader, wantHeaderValue := "X-Session-Id", "Test"
	wantBody := `This is a simple test`

	want := "192.0.2.1:1234 - - [TEST] \"GET /test/endpoint HTTP/1.1\" 200 21  \x1b[0m X-Session-Id=Test\nTEST \x1b[32mInfo: This is a simple test\x1b[0m\n"
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithHTTPHeader(wantHeader), WithTimeText("TEST"))

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
	l2 := New(WithOutput(have), WithTimeText("TEST"))

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

type writeCountTest struct {
	w io.Writer
	t *testing.T
}

func (w writeCountTest) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if n != len(p) {
		w.t.Errorf("want: %d have: %d (strip write count)", len(p), n)
	}
	return
}

func TestWithNoColorOutput(t *testing.T) {
	want := "TEST Info: This is a simple test\n"
	have := new(bytes.Buffer)

	sw := writeCountTest{w: StripWriter(have), t: t}
	l := New(WithOutput(sw), WithTimeText("TEST"))
	l.Info("This is a simple test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestPrintOutput(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is a simple test\x1b[0m\n"
	have := new(bytes.Buffer)

	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Println("This is a simple test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestConcurrentInfofWarnf(t *testing.T) {
	base := "This is a simple concurrent test (%s) #%02d"
	have := new(bytes.Buffer)

	finish := make(chan struct{}, 1)
	rand.Seed(time.Now().UnixNano())

	l := New(WithOutput(have), WithTimeText("TEST"))

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

func BenchmarkLinet(b *testing.B) {
	b.ReportAllocs()

	have := new(bytes.Buffer)
	l := New(WithOutput(have))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("Testing with a string of", i)
	}
}

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()

	have := new(bytes.Buffer)
	l := New(WithOutput(have))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// The lucky line below has more parameters than directives because the KV's get taken out.
		l.Printf("Testing with a string of %02d\n", i, KV("hello0", "world"), KV("hello1", "world"), KV("hello2", "world"), KV("hello3", "world"), KV("hello4", "world"), KV("hello5", "world"), KV("hello6", "world"), KV("hello7", "world"), KV("hello8", "world"), KV("hello9", "world"))
	}
}

func BenchmarkFormat2(b *testing.B) {
	b.ReportAllocs()

	have := new(bytes.Buffer)
	l := New(WithOutput(have))

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

func BenchmarkFormat3(b *testing.B) {
	b.ReportAllocs()

	have := new(bytes.Buffer)
	l := New(WithOutput(have))
	l2 := l.Fields(map[string]interface{}{
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
		l2.Printf("Testing with a string of %02d\n", i)
	}
}

func BenchmarkNone(b *testing.B) {
	b.ReportAllocs()

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Print("Testing with a string of ", i)
	}
}

var concurrentWantf = `TEST %[1]sInfo: This is a simple concurrent test (info) #00%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #01%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #02%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #03%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #04%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #05%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #06%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #07%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #08%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #09%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #10%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #11%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #12%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #13%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #14%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #15%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #16%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #17%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #18%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #19%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #20%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #21%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #22%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #23%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #24%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #25%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #26%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #27%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #28%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #29%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #30%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #31%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #32%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #33%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #34%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #35%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #36%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #37%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #38%[3]s
TEST %[1]sInfo: This is a simple concurrent test (info) #39%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #00%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #01%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #02%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #03%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #04%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #05%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #06%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #07%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #08%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #09%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #10%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #11%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #12%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #13%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #14%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #15%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #16%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #17%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #18%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #19%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #20%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #21%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #22%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #23%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #24%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #25%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #26%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #27%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #28%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #29%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #30%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #31%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #32%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #33%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #34%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #35%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #36%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #37%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #38%[3]s
TEST %[2]sWarn: This is a simple concurrent test (warn) #39%[3]s`
