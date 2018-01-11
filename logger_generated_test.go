// go generate
// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
// GENERATED ON: 2018-01-11 21:25:03.866872736 +0000 UTC m=+0.208257819

package logger

import (
	"bytes"
	"testing"
)

func TestPrint(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Print("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestPrintf(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Printf("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestPrintln(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Println("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerPrint(t *testing.T) {
	l := new(nilLogger)
	l.Print("does", "not", "work")
}

func TestNilLoggerPrintf(t *testing.T) {
	l := new(nilLogger)
	l.Printf("%s %s %s", "does", "not", "work")
}

func TestNilLoggerPrintln(t *testing.T) {
	l := new(nilLogger)
	l.Println("does", "not", "work")
}

func TestInfo(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Info("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestInfof(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Infof("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestInfoln(t *testing.T) {
	want := "TEST \x1b[32mInfo: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Infoln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerInfo(t *testing.T) {
	l := new(nilLogger)
	l.Info("does", "not", "work")
}

func TestNilLoggerInfof(t *testing.T) {
	l := new(nilLogger)
	l.Infof("%s %s %s", "does", "not", "work")
}

func TestNilLoggerInfoln(t *testing.T) {
	l := new(nilLogger)
	l.Infoln("does", "not", "work")
}

func TestWarn(t *testing.T) {
	want := "TEST \x1b[33mWarn: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Warn("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestWarnf(t *testing.T) {
	want := "TEST \x1b[33mWarn: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Warnf("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestWarnln(t *testing.T) {
	want := "TEST \x1b[33mWarn: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Warnln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerWarn(t *testing.T) {
	l := new(nilLogger)
	l.Warn("does", "not", "work")
}

func TestNilLoggerWarnf(t *testing.T) {
	l := new(nilLogger)
	l.Warnf("%s %s %s", "does", "not", "work")
}

func TestNilLoggerWarnln(t *testing.T) {
	l := new(nilLogger)
	l.Warnln("does", "not", "work")
}

func TestDebug(t *testing.T) {
	want := "TEST \x1b[36mDebug: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Debug("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestDebugf(t *testing.T) {
	want := "TEST \x1b[36mDebug: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Debugf("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestDebugln(t *testing.T) {
	want := "TEST \x1b[36mDebug: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Debugln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerDebug(t *testing.T) {
	l := new(nilLogger)
	l.Debug("does", "not", "work")
}

func TestNilLoggerDebugf(t *testing.T) {
	l := new(nilLogger)
	l.Debugf("%s %s %s", "does", "not", "work")
}

func TestNilLoggerDebugln(t *testing.T) {
	l := new(nilLogger)
	l.Debugln("does", "not", "work")
}

func TestError(t *testing.T) {
	want := "TEST \x1b[35mError: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Error("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestErrorf(t *testing.T) {
	want := "TEST \x1b[35mError: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Errorf("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestErrorln(t *testing.T) {
	want := "TEST \x1b[35mError: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Errorln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerError(t *testing.T) {
	l := new(nilLogger)
	l.Error("does", "not", "work")
}

func TestNilLoggerErrorf(t *testing.T) {
	l := new(nilLogger)
	l.Errorf("%s %s %s", "does", "not", "work")
}

func TestNilLoggerErrorln(t *testing.T) {
	l := new(nilLogger)
	l.Errorln("does", "not", "work")
}

func TestTrace(t *testing.T) {
	want := "TEST \x1b[34mTrace: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Trace("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestTracef(t *testing.T) {
	want := "TEST \x1b[34mTrace: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Tracef("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestTraceln(t *testing.T) {
	want := "TEST \x1b[34mTrace: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Traceln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerTrace(t *testing.T) {
	l := new(nilLogger)
	l.Trace("does", "not", "work")
}

func TestNilLoggerTracef(t *testing.T) {
	l := new(nilLogger)
	l.Tracef("%s %s %s", "does", "not", "work")
}

func TestNilLoggerTraceln(t *testing.T) {
	l := new(nilLogger)
	l.Traceln("does", "not", "work")
}

func TestFatal(t *testing.T) {
	want := "TEST \x1b[31mFatal: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Fatal("This ", "is ", "an ", "automated ", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestFatalf(t *testing.T) {
	want := "TEST \x1b[31mFatal: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Fatalf("This is an %s test", "automated")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestFatalln(t *testing.T) {
	want := "TEST \x1b[31mFatal: This is an automated test\x1b[0m\n"

	have := new(bytes.Buffer)
	l := New(WithOutput(have), WithTimeText("TEST"))
	l.Fatalln("This", "is", "an", "automated", "test")

	if want != have.String() {
		t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
	}
}

func TestNilLoggerFatal(t *testing.T) {
	l := new(nilLogger)
	l.Fatal("does", "not", "work")
}

func TestNilLoggerFatalf(t *testing.T) {
	l := new(nilLogger)
	l.Fatalf("%s %s %s", "does", "not", "work")
}

func TestNilLoggerFatalln(t *testing.T) {
	l := new(nilLogger)
	l.Fatalln("does", "not", "work")
}

func TestNilLoggerColor(t *testing.T) {
	l := new(nilLogger)
	l.Color(ColorBlack)
}

func TestNilLoggerField(t *testing.T) {
	l := new(nilLogger)
	l.Field("nil", struct{ Logging string }{"made up"})
}

func TestNilLoggerFields(t *testing.T) {
	l := new(nilLogger)
	l.Fields(map[string]interface{}{"nothing": "to see here"})
}

func TestNilLoggerHTTPMiddleware(t *testing.T) {
	l := new(nilLogger)
	l.HTTPMiddleware(nil)
}

func TestNilLoggerNoColor(t *testing.T) {
	l := new(nilLogger)
	l.NoColor()
}

func TestNilLoggerOnErr(t *testing.T) {
	l := new(nilLogger)
	l.OnErr(nil)
}

func TestNilLoggerSuppress(t *testing.T) {
	l := new(nilLogger)
	l.Suppress(0)
}

func TestNilLoggerWith(t *testing.T) {
	l := new(nilLogger)
	l.With(WithTimeAsUTC())
}
