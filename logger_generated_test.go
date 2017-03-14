package logger

import (
	"errors"
	"net/http"
	"testing"
)

func TestNilLogger(t *testing.T) {
	nl := new(nilLogger)

	nl.Suppress()
	nl.UnSuppress()

	nl.Color(Black).OnErr(errors.New("Test")).Info("Test")
	nl.HTTPMiddleware(http.NotFoundHandler())

	nl.Debug("Test")
	nl.Debugf("fmt", "Test")
	nl.Error("Test")
	nl.Errorf("fmt", "Test")
	nl.Fatal("Test")
	nl.Fatalf("fmt", "Test")
	nl.Info("Test")
	nl.Infof("fmt", "Test")
	nl.Trace("Test")
	nl.Tracef("fmt", "Test")
	nl.Warn("Test")
	nl.Warnf("fmt", "Test")
}

func TestLogLevels(t *testing.T) {
	i := 1
	if Level().Info != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Info, i)
	}

	i += i
	if Level().Warn != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Warn, i)
	}

	i += i
	if Level().Error != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Error, i)
	}

	i += i
	if Level().Debug != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Debug, i)
	}

	i += i
	if Level().Trace != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Trace, i)
	}

	i += i
	if Level().Fatal != LogLevel(i) {
		t.Logf("\nwant: %d\n\nhave: %d\n", Level().Fatal, i)
	}

	ukn := "unknown"
	if ukn != LogLevel(3).String() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogLevel(3).String())
	}

	if ukn != LogLevel(3).Short() {
		t.Errorf("\nwant: %s\n\nhave: %s\n", ukn, LogLevel(3).Short())
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
