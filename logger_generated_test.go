package logger

import (
	"errors"
	"net/http"
	"testing"
)

// this just runs the nil logger and doesn't test for anything
// because nothing is supposed to happen with a nil logger
func TestNilLogger(t *testing.T) {
	nl := new(nilLogger)

	nl.Suppress()
	nl.UnSuppress()

	nl.Color(Black).OnErr(errors.New("Test")).Info("Test")
	nl.NoColor().Info("Test")
	nl.HTTPMiddleware(http.NotFoundHandler())

	nl.Debug("Test")
	nl.Debugf("fmt %s", "Test")
	nl.Debugln("Test")
	nl.Error("Test")
	nl.Errorf("fmt %s", "Test")
	nl.Errorln("Test")
	nl.Fatal("Test")
	nl.Fatalf("fmt %s", "Test")
	nl.Fatalln("Test")
	nl.Print("Test")
	nl.Printf("fmt %s", "Test")
	nl.Println("Test")
	nl.Panic("Test")
	nl.Panicf("fmt %s", "Test")
	nl.Panicln("Test")
	nl.Info("Test")
	nl.Infof("fmt %s", "Test")
	nl.Infoln("Test")
	nl.Trace("Test")
	nl.Tracef("fmt %s", "Test")
	nl.Traceln("Test")
	nl.Warn("Test")
	nl.Warnf("fmt %s", "Test")
	nl.Warnln("Test")

	nl.Field("key", "value")
	nl.Fields(KV("key", "value"), KV("key", "value"))
	nl.Fields(KVMap(KeyValues{"key": "value"})...)
}

func TestLogLevels(t *testing.T) {
	i := 1
	if Level().Info != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Info, i)
	}

	i += i
	if Level().Warn != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Warn, i)
	}

	i += i
	if Level().Error != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Error, i)
	}

	i += i
	if Level().Debug != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Debug, i)
	}

	i += i
	if Level().Trace != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Trace, i)
	}

	i += i
	if Level().Fatal != LogLevel(i) {
		t.Errorf("\nwant: %d\n\nhave: %d\n", Level().Fatal, i)
	}

	ukn := "unknown"
	if ukn != LogLevel(3).String() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogLevel(3).String())
	}

	if ukn != LogLevel(3).Short() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogLevel(3).Short())
	}

	if ukn != LogLevel(3).StringWithColon() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogLevel(3).StringWithColon())
	}
}

func TestLogLevelString(t *testing.T) {
	want := []string{"Debug", "Error", "Fatal", "Info", "Trace", "Warn", "Panic"}
	wantWithColon := []string{"Debug:", "Error:", "Fatal:", "Info:", "Trace:", "Warn:", "Panic:"}
	have := []LogLevel{Level().Debug, Level().Error, Level().Fatal, Level().Info, Level().Trace, Level().Warn, Level().Panic}

	if len(want) != len(have) {
		t.Error("Test want and test have don't have equal lengths, check the test.")
	}

	for i := range want {
		if want[i] != have[i].String() {
			t.Errorf("\nwant: %s\n\nhave: %s\n", want[i], have[i].String())
		}
	}

	for i := range wantWithColon {
		if wantWithColon[i] != have[i].StringWithColon() {
			t.Errorf("\nwant: %s\n\nhave: %s\n", wantWithColon[i], have[i].StringWithColon())
		}
	}
}

func TestLogColors(t *testing.T) {
	want := []string{"Black", "Blue", "Cyan", "Green", "Magenta", "Red", "White", "Yellow"}
	have := []LogColor{Black, Blue, Cyan, Green, Magenta, Red, White, Yellow}

	if len(want) != len(have) {
		t.Error("Test want and test have don't have equal lengths, check the test.")
	}

	for i := range want {
		if want[i] != have[i].String() {
			t.Errorf("\nwant: %s\n\nhave: %s\n", want[i], have[i].String())
		}
	}

	ukn := "unknown"
	if ukn != LogColor(200).String() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogColor(200).String())
	}
}

func TestLogESCColors(t *testing.T) {
	want := []string{"\x1b[30m", "\x1b[34m", "\x1b[36m", "\x1b[32m", "\x1b[35m", "\x1b[31m", "\x1b[37m", "\x1b[33m"}
	have := []LogColor{Black, Blue, Cyan, Green, Magenta, Red, White, Yellow}

	if len(want) != len(have) {
		t.Error("Test want and test have don't have equal lengths, check the test.")
	}

	for i := range want {
		if want[i] != have[i].ToESCColor() {
			t.Errorf("\nwant: %s\n\nhave: %s\n", want[i], have[i].ToESCColor())
		}
	}

	ukn := "0"
	if ukn != LogColor(200).ToESCColor() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogColor(200).ToESCColor())
	}
}
