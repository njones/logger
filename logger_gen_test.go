// GENERATED BY ./gen/main.go; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: 2020-02-22 04:38:59.3937254 +0000 UTC m=+0.747119401 ~~
package logger

import (
	"bytes"
	"testing"
)

func TestLogger(t *testing.T) {
	have := new(bytes.Buffer)
	log := New(WithOutput(have), WithTimeText("Jan-01-2000"))

	tests := []struct {
		name   string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs []interface{}
		fatal  int
		want   string
	}{

		{
			name:   "log.Print",
			method: log.Print,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 abcdefghi\n",
		}, {
			name:   "log.Info",
			method: log.Info,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[32mINFO: abcdefghi[0m\n",
		}, {
			name:   "log.Warn",
			method: log.Warn,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[33mWARN: abcdefghi[0m\n",
		}, {
			name:   "log.Debug",
			method: log.Debug,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[36mDEBUG: abcdefghi[0m\n",
		}, {
			name:   "log.Error",
			method: log.Error,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[35mERROR: abcdefghi[0m\n",
		}, {
			name:   "log.Trace",
			method: log.Trace,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[34mTRACE: abcdefghi[0m\n",
		}, {
			name:   "log.Fatal",
			method: log.Fatal,
			inputs: []interface{}{"abc", "def", "ghi"},
			fatal:  888,
			want: "Jan-01-2000 \x1b[31mFATAL: abcdefghi[0m\n",
		}, {
			name:   "log.Printf",
			method: log.Printf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 abc ghi abc def\n",
		}, {
			name:   "log.Infof",
			method: log.Infof,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[32mINFO: abc ghi abc def[0m\n",
		}, {
			name:   "log.Warnf",
			method: log.Warnf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[33mWARN: abc ghi abc def[0m\n",
		}, {
			name:   "log.Debugf",
			method: log.Debugf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[36mDEBUG: abc ghi abc def[0m\n",
		}, {
			name:   "log.Errorf",
			method: log.Errorf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[35mERROR: abc ghi abc def[0m\n",
		}, {
			name:   "log.Tracef",
			method: log.Tracef,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[34mTRACE: abc ghi abc def[0m\n",
		}, {
			name:   "log.Fatalf",
			method: log.Fatalf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			fatal:  888,
			want: "Jan-01-2000 \x1b[31mFATAL: abc ghi abc def[0m\n",
		}, {
			name:   "log.Println",
			method: log.Println,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 abc def ghi\n",
		}, {
			name:   "log.Infoln",
			method: log.Infoln,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[32mINFO: abc def ghi[0m\n",
		}, {
			name:   "log.Warnln",
			method: log.Warnln,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[33mWARN: abc def ghi[0m\n",
		}, {
			name:   "log.Debugln",
			method: log.Debugln,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[36mDEBUG: abc def ghi[0m\n",
		}, {
			name:   "log.Errorln",
			method: log.Errorln,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[35mERROR: abc def ghi[0m\n",
		}, {
			name:   "log.Traceln",
			method: log.Traceln,
			inputs: []interface{}{"abc", "def", "ghi"},

			want: "Jan-01-2000 \x1b[34mTRACE: abc def ghi[0m\n",
		}, {
			name:   "log.Fatalln",
			method: log.Fatalln,
			inputs: []interface{}{"abc", "def", "ghi"},
			fatal:  888,
			want: "Jan-01-2000 \x1b[31mFATAL: abc def ghi[0m\n",
		}}

	for _, test := range tests {
		have.Reset()
		t.Run(test.name, func(tt *testing.T) {
			var fatal int
			if test.name == "log.Fatal" {
				log.FatalInt(test.fatal)
				log.(*baseLogger).exit.Func = func(i int) { fatal = i }
			}
			switch fn := test.method.(type) {
			case func(...interface{}):
				fn(test.inputs...)
			case func(string, ...interface{}):
				fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: valid function signature\nhave: %T", fn)
			}
			if test.want != have.String() {
				tt.Errorf("\nhave: %q\n\nwant: %q\n", have.String(), test.want)
			}
			if test.name == "log.Fatal" {
				if fatal != test.fatal {
					tt.Errorf("\nhave: %d\n\nwant: %d\n", fatal, test.fatal)
				}
			}
		})
	}
}

func TestOnErrFalse(t *testing.T) {
	have := struct {
		output  *bytes.Buffer
		exitInt int
		rtn     Return
	}{
		output:  new(bytes.Buffer),
		exitInt: 100,
	}
	want := struct {
		hasErr  bool
		exitInt int
	}{
		hasErr:  false,
		exitInt: 100,
	}

	log := New(WithOutput(have.output), WithTimeText("Jan-01-2000")).OnErr(nil)

	tests := []struct {
		name   string
		prefix string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs []interface{}
		want   string
	}{

		{
			name:   "log.Print OnErr:False",
			method: log.Print,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Info OnErr:False",
			method: log.Info,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Warn OnErr:False",
			method: log.Warn,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Debug OnErr:False",
			method: log.Debug,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Error OnErr:False",
			method: log.Error,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Trace OnErr:False",
			method: log.Trace,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Fatal OnErr:False",
			method: log.Fatal,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Printf OnErr:False",
			method: log.Printf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Infof OnErr:False",
			method: log.Infof,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Warnf OnErr:False",
			method: log.Warnf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Debugf OnErr:False",
			method: log.Debugf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Errorf OnErr:False",
			method: log.Errorf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Tracef OnErr:False",
			method: log.Tracef,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Fatalf OnErr:False",
			method: log.Fatalf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Println OnErr:False",
			method: log.Println,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Infoln OnErr:False",
			method: log.Infoln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Warnln OnErr:False",
			method: log.Warnln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Debugln OnErr:False",
			method: log.Debugln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Errorln OnErr:False",
			method: log.Errorln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Traceln OnErr:False",
			method: log.Traceln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}, {
			name:   "log.Fatalln OnErr:False",
			method: log.Fatalln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "",
		}}

	for _, test := range tests {
		have.output.Reset()
		t.Run(test.name, func(tt *testing.T) {
			if test.name == "log.Fatal OnErr:False" {
				log.(*onErrLogger).b.exit.Func = func(i int) { have.exitInt = i }
			}
			switch fn := test.method.(type) {
			case func(...interface{}) Return:
				have.rtn = fn(test.inputs...)
			case func(string, ...interface{}) Return:
				have.rtn = fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: %T\nwant: <a valid function signature>", fn)
			}
			if have.output.String() != test.want {
				tt.Errorf("\nhave: %q\n\nwant: %q\n", have.output.String(), test.want)
			}
			if have.rtn.HasErr() != want.hasErr {
				tt.Errorf("\nhave: %t\nwant: %t\n", have.rtn.HasErr(), want.hasErr)
			}
			if have.rtn.HasErr() && have.rtn.Err() != bytes.ErrTooLarge {
				tt.Errorf("\nhave: %v\nwant: %v\n", have.rtn.Err(), bytes.ErrTooLarge)
			}
			if test.name == "log.Fatal OnErr:False" {
				if have.exitInt != want.exitInt {
					tt.Errorf("\nhave: %d\nwant: %d\n", have.exitInt, want.exitInt)
				}
			}
		})
	}
}
func TestOnErrTrue(t *testing.T) {
	have := struct {
		output  *bytes.Buffer
		exitInt int
		rtn     Return
	}{
		output:  new(bytes.Buffer),
		exitInt: 100,
	}
	want := struct {
		hasErr  bool
		exitInt int
	}{
		hasErr:  true,
		exitInt: 1,
	}

	log := New(WithOutput(have.output), WithTimeText("Jan-01-2000")).OnErr(bytes.ErrTooLarge)

	tests := []struct {
		name   string
		prefix string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs []interface{}
		want   string
	}{

		{
			name:   "log.Print OnErr:True",
			method: log.Print,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 abcdefghi\n",
		}, {
			name:   "log.Info OnErr:True",
			method: log.Info,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[32mINFO: abcdefghi[0m\n",
		}, {
			name:   "log.Warn OnErr:True",
			method: log.Warn,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[33mWARN: abcdefghi[0m\n",
		}, {
			name:   "log.Debug OnErr:True",
			method: log.Debug,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[36mDEBUG: abcdefghi[0m\n",
		}, {
			name:   "log.Error OnErr:True",
			method: log.Error,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[35mERROR: abcdefghi[0m\n",
		}, {
			name:   "log.Trace OnErr:True",
			method: log.Trace,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[34mTRACE: abcdefghi[0m\n",
		}, {
			name:   "log.Fatal OnErr:True",
			method: log.Fatal,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[31mFATAL: abcdefghi[0m\n",
		}, {
			name:   "log.Printf OnErr:True",
			method: log.Printf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 abc ghi abc def\n",
		}, {
			name:   "log.Infof OnErr:True",
			method: log.Infof,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[32mINFO: abc ghi abc def[0m\n",
		}, {
			name:   "log.Warnf OnErr:True",
			method: log.Warnf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[33mWARN: abc ghi abc def[0m\n",
		}, {
			name:   "log.Debugf OnErr:True",
			method: log.Debugf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[36mDEBUG: abc ghi abc def[0m\n",
		}, {
			name:   "log.Errorf OnErr:True",
			method: log.Errorf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[35mERROR: abc ghi abc def[0m\n",
		}, {
			name:   "log.Tracef OnErr:True",
			method: log.Tracef,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[34mTRACE: abc ghi abc def[0m\n",
		}, {
			name:   "log.Fatalf OnErr:True",
			method: log.Fatalf,
			format: "%s %[3]s %[1]s %[2]s",
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[31mFATAL: abc ghi abc def[0m\n",
		}, {
			name:   "log.Println OnErr:True",
			method: log.Println,
			inputs: []interface{}{"abc", "def", "ghi"},
			want:   "Jan-01-2000 abc def ghi\n",
		}, {
			name:   "log.Infoln OnErr:True",
			method: log.Infoln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[32mINFO: abc def ghi[0m\n",
		}, {
			name:   "log.Warnln OnErr:True",
			method: log.Warnln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[33mWARN: abc def ghi[0m\n",
		}, {
			name:   "log.Debugln OnErr:True",
			method: log.Debugln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[36mDEBUG: abc def ghi[0m\n",
		}, {
			name:   "log.Errorln OnErr:True",
			method: log.Errorln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[35mERROR: abc def ghi[0m\n",
		}, {
			name:   "log.Traceln OnErr:True",
			method: log.Traceln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[34mTRACE: abc def ghi[0m\n",
		}, {
			name:   "log.Fatalln OnErr:True",
			method: log.Fatalln,
			inputs: []interface{}{"abc", "def", "ghi"},
			want: "Jan-01-2000 \x1b[31mFATAL: abc def ghi[0m\n",
		}}

	for _, test := range tests {
		have.output.Reset()
		t.Run(test.name, func(tt *testing.T) {
			if test.name == "log.Fatal OnErr:True" {
				log.(*onErrLogger).b.exit.Func = func(i int) { have.exitInt = i }
			}
			switch fn := test.method.(type) {
			case func(...interface{}) Return:
				have.rtn = fn(test.inputs...)
			case func(string, ...interface{}) Return:
				have.rtn = fn(test.format, test.inputs...)
			default:
				tt.Errorf("\nhave: %T\nwant: <a valid function signature>", fn)
			}
			if have.output.String() != test.want {
				tt.Errorf("\nhave: %q\n\nwant: %q\n", have.output.String(), test.want)
			}
			if have.rtn.HasErr() != want.hasErr {
				tt.Errorf("\nhave: %t\nwant: %t\n", have.rtn.HasErr(), want.hasErr)
			}
			if have.rtn.HasErr() && have.rtn.Err() != bytes.ErrTooLarge {
				tt.Errorf("\nhave: %v\nwant: %v\n", have.rtn.Err(), bytes.ErrTooLarge)
			}
			if test.name == "log.Fatal OnErr:True" {
				if have.exitInt != want.exitInt {
					tt.Errorf("\nhave: %d\nwant: %d\n", have.exitInt, want.exitInt)
				}
			}
		})
	}
}