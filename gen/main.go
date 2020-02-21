package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "../level.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	gendFile, err := os.Create("../logger_gen.go")
	if err != nil {
		log.Fatal(err)
	}
	defer gendFile.Close()
	defer func() {
		cmd := exec.Command("gofmt", "-w", "-s", "../logger_gen.go")
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal("gofmt logger generated error: ", err)
		}
		log.Println("go Logger code generated and formatted.")
	}()

	gendTestFile, err := os.Create("../logger_gen_test.go")
	if err != nil {
		log.Fatal(err)
	}
	defer gendTestFile.Close()
	defer func() {
		cmd := exec.Command("gofmt", "-w", "-s", "../logger_gen_test.go")
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal("gofmt logger generated test error: ", err)
		}
		log.Println("go Logger test code generated and formatted.")
	}()

	ast.Walk(VisitorFunc(FindTypes), node)

	node2, err := parser.ParseFile(fset, "../logger.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	ast.Walk(VisitorFunc(FindTypes), node2)

	timestampMap := make(map[byte]string)

	if len(timestampMap) > 20 {
		panic("the map is arbitrarily too long, becuase we're counting back from 31 in ASCII")
	}

	for i, v := range tsMap {
		timestampMap[byte(0x1F-i)] = v
	}

	sort.Sort(strSorter(tsMap))

	genTemplate.Execute(gendFile, struct {
		LevelDisplay    []string
		DuplicateFields []string
		Timestamp       time.Time
		TimestampMap    map[byte]string
		Levels          []데이터
	}{
		LevelDisplay:    []string{"default", "box", "short", "short.box"},
		DuplicateFields: fields,
		Timestamp:       time.Now(),
		TimestampMap:    timestampMap,
		Levels:          levelsData,
	})

	type onErrData struct {
		HasErr, ExitInt, Name, Err string
	}

	genTestTemplate.Execute(gendTestFile, struct {
		LevelDisplay    []string
		DuplicateFields []string
		Timestamp       time.Time
		TimestampMap    map[byte]string
		Levels          []데이터
		Text1           []string
		Format1         string
		OnErr           map[string]onErrData
	}{
		LevelDisplay:    []string{"default", "box", "short", "short.box"},
		DuplicateFields: fields,
		Timestamp:       time.Now(),
		TimestampMap:    timestampMap,
		Levels:          levelsData,
		Text1:           []string{"abc", "def", "ghi"},
		Format1:         "%s %[3]s %[1]s %[2]s",
		OnErr: map[string]onErrData{
			"TestOnErrTrue": {
				HasErr:  "true",
				ExitInt: "1",
				Name:    "OnErr:True",
				Err:     "bytes.ErrTooLarge",
			},
			"TestOnErrFalse": {
				HasErr:  "false",
				ExitInt: "100",
				Name:    "OnErr:False",
				Err:     "nil",
			},
		},
	})
}

type strSorter []string

func (s strSorter) Len() int {
	return len(s)
}

func (s strSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s strSorter) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

// Mon Jan 2 15:04:05 -0700 MST 2006
var tsMap = []string{"Monday", "January", "02", "15", "04", "05", "-0700", "MST", "2006", "Mon", "Jan", "01", "06", "03", "pm", "05.000000"}

// VisitorFunc a type
type VisitorFunc func(n ast.Node) ast.Visitor

// Visit does the node walking
func (f VisitorFunc) Visit(n ast.Node) ast.Visitor { return f(n) }

// FindTypes loops through the nodes
func FindTypes(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		return VisitorFunc(FindTypes)
	case *ast.File:
		return VisitorFunc(FindTypes)
	case *ast.TypeSpec:
		return duplicate(n)
	case *ast.GenDecl:
		return generate(n)
	}
	return nil
}

type 데이터 struct {
	// for creating the maps
	Levels map[string]string
	Color  string

	// for creating the functions
	FuncName  string
	LevelName string // in the case of Print (the level is different)

	Writeize string
	Levelize string
	Colorize string
	Timeize  string

	HasOnErr bool

	AsPrint   bool
	AsPrintf  bool
	AsPrintln bool

	AsPrintTest   bool
	AsPrintfTest  bool
	AsPrintlnTest bool
}

var levelsData []데이터
var fields []string

func duplicate(node *ast.TypeSpec) ast.Visitor {
	var s *ast.StructType
	var ok bool
	if s, ok = node.Type.(*ast.StructType); !ok {
		return VisitorFunc(FindTypes)
	}
	if node.Name.Name != "baseLogger" {
		return VisitorFunc(FindTypes)
	}
	for _, f := range s.Fields.List {
		fields = append(fields, f.Names[0].Name)
	}
	return nil
}

func generate(n *ast.GenDecl) ast.Visitor {
	if n.Tok != token.CONST {
		return VisitorFunc(FindTypes)
	}

	var 𝜎Kind, 𝜎Comment string

	for _, spec := range n.Specs {
		vspec := spec.(*ast.ValueSpec)
		for _, name := range vspec.Names {

			if ident, ok := vspec.Type.(*ast.Ident); ok {
				𝜎Kind = ident.Name
			}

			if comment := vspec.Comment; comment != nil {
				for _, commGrp := range comment.List {
					𝜎Comment = commGrp.Text // assume only one comment, otherwise we're overwritting here
				}
			}

			if 𝜎Kind == "logLevel" {
				𓁹𓁹(name.String(), 𝜎Comment) // looks at the name and comment tag to populate the templateData ("see" what I did there...)
			}
		}
	}
	return VisitorFunc(FindTypes)
}

var colors = []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

func 𓁹𓁹(name, comment string) {
	// check to see if we are going to be parsing this comment
	if !strings.HasPrefix(comment, "//`") {
		return // do nothing...
	}

	var ᄀ 데이터

	ᄀ.HasOnErr = true
	ᄀ.AsPrint, ᄀ.AsPrintf, ᄀ.AsPrintln = true, true, true
	ᄀ.AsPrintTest, ᄀ.AsPrintfTest, ᄀ.AsPrintlnTest = true, true, true
	ᄀ.FuncName = strings.Title(name)
	ᄀ.LevelName = ᄀ.FuncName
	ᄀ.Levelize = fmt.Sprintf(", %s.levelize(b.display)", name)
	ᄀ.Colorize = fmt.Sprintf(", %s.colorize()", name)

	ᄀ.Levels = map[string]string{
		"default": strings.ToUpper(name) + ":",
		"box":     "[" + strings.ToUpper(name) + "]",
	}
	// TODO(njones): make this much more robust ... alot of assumptions are here
	//	like what you ask?
	//		For starters that there is a valid `` open and close tag
	//		Second, that there are no spaces inside of the quote tags
	comment = strings.TrimPrefix(comment, "//`")
	fields := strings.Fields(comment)
	for _, field := range fields {
		kv := strings.Split(field, ":")
		if len(kv) != 2 {
			continue
		}
		val := strings.Title(strings.Trim(kv[1], "`\""))
		if val == "-" {
			if kv[0] == "long" {
				delete(ᄀ.Levels, "default")
				delete(ᄀ.Levels, "box")
			}
			continue
		}
		switch kv[0] {
		case "long":
			ᄀ.LevelName = val
		case "short":
			ᄀ.Levels["short"] = strings.ToUpper(val) + ":"
			ᄀ.Levels["short.box"] = "[" + strings.ToUpper(val) + "]"
		case "color":
			for i, v := range colors {
				if v == strings.ToLower(val) {
					ᄀ.Color = fmt.Sprintf(`\x1b[%dm`, i+30)
				}
			}
		case "fn":
			ᄀ.AsPrint, ᄀ.AsPrintf, ᄀ.AsPrintln = false, false, false
			types := strings.Split(strings.ToLower(val), ",")
			for _, t := range types {
				switch t {
				case "base":
					ᄀ.AsPrint = true
				case "f":
					ᄀ.AsPrintf = true
				case "ln":
					ᄀ.AsPrintln = true
				}
			}
		}
	}
	if ᄀ.FuncName == "Panic" {
		ᄀ.AsPrintTest, ᄀ.AsPrintfTest, ᄀ.AsPrintlnTest = false, false, false
		ᄀ.Writeize = ", writeize{b.exit.buf}"
	}

	// a request from the "Keep Print Plain campaign"...
	if ᄀ.FuncName == "Print" {
		ᄀ.Levels = map[string]string{}
		ᄀ.Levelize = ""
		ᄀ.Colorize = ""
		ᄀ.Color = ""
	}

	if ᄀ.FuncName == "HTTP" {
		ᄀ.HasOnErr = false
		ᄀ.AsPrintTest, ᄀ.AsPrintfTest, ᄀ.AsPrintlnTest = false, false, false
		ᄀ.Levelize = ""
		ᄀ.Colorize = ""
		ᄀ.Timeize = ", timeize(nil)"
		ᄀ.Color = ""
	}

	levelsData = append(levelsData, ᄀ)
}

// genTemplate is the template for the logger... note that every end of line should end
// with a semicolon, because spacing can get borked easily, and we'll just
// use `gofmt` later to make things nice and pretty
var genTemplate = template.Must(template.New("").Parse(`{{ define "preHook" }}
{{- if (eq .FuncName "Panic") -}}
	b.exit.buf = new(bytes.Buffer);
{{- end -}}
{{ end }}
{{ define "postHook" }}
{{- if (eq .FuncName "Panic") -}}
	panic(b.exit.buf.String());
{{- end -}}
{{- if (eq .FuncName "Fatal") -}}
	b.exit.Func(b.exit.Int);
{{- end -}}
{{ end }}
// GENERATED BY ./gen/main.go; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: {{ .Timestamp }} ~~
package logger

import (
	"bytes"
)

// Maps

var tsFmtMap = map[string]byte{
	{{ range $key, $val := .TimestampMap -}}
	"{{ $val }}": {{ printf "0x%x" $key }},
	{{ end -}}
}

var tsRuneMap = map[rune]string{
	{{ range $key, $val := .TimestampMap -}}
	{{ printf "0x%x" $key }}: "{{ $val }}",
	{{ end -}}
}

var llMap = map[int]map[logLevel]string{
	{{- $out := . -}}
	{{ range $idx, $val := .LevelDisplay }}
	{{ printf "%d" $idx }}: map[logLevel]string{
		{{- range $index, $value := $out.Levels -}}
		{{ with (index $value.Levels $val) }}{{ $value.FuncName }}: "{{ . }}",{{ end }}
		{{ end -}}
	},
	{{- end }}
}

func (ll logLevel) flag() int { return int(ll) }

func (ll logLevel) levelize(display int) levelize {
	return levelize(llMap[display][ll])
}

func (ll logLevel) colorize() colorize {
	m := map[logLevel]string{
		{{- range $index, $value := $out.Levels -}}
		{{- if (and (gt $index 0) (ne $value.Color "" )) -}}
		{{- $value.FuncName -}}: "{{ $value.Color }}",
		{{- end }}
		{{ end -}}
	}
	return colorize(m[ll])
}

// Helper Functions

func duplicate(b *baseLogger) *baseLogger {
	bb := new(baseLogger);
	{{- range $idx, $val := .DuplicateFields }}
	bb.{{ $val }} = b.{{ $val }};
	{{- end -}}
	return bb;
}

// Standard and Extended Logger Functions

{{ range $idx, $value := .Levels }}
{{ if $value.AsPrint }}
func (b *baseLogger) {{ $value.FuncName }}(v ...interface{}) {
	if !hasFlag(b.supress, {{ $value.LevelName }}.flag()) {
		{{- template "preHook" . -}}
		b.print(bPrint, v{{ $value.Writeize }}{{ $value.Levelize }}{{ $value.Colorize }}{{ $value.Timeize }});
		{{- template "postHook" . -}}
	}
}
{{ end }}
{{ if $value.AsPrintf }}
func (b *baseLogger) {{ $value.FuncName }}f(f string, v ...interface{}) {
	if !hasFlag(b.supress, {{ $value.LevelName }}.flag()) {
		{{- template "preHook" . -}}
		b.print(bPrintf, v, formatize(f){{ $value.Writeize }}{{ $value.Levelize }}{{ $value.Colorize }}{{ $value.Timeize }});
		{{- template "postHook" . -}}
	}
}
{{ end }}
{{ if $value.AsPrintln }}
func (b *baseLogger) {{ $value.FuncName }}ln(v ...interface{}) {
	if !hasFlag(b.supress, {{ $value.LevelName }}.flag()) {
		{{- template "preHook" . -}}
		b.print(bPrintln, v{{ $value.Writeize }}{{ $value.Levelize }}{{ $value.Colorize }}{{ $value.Timeize }});
		{{- template "postHook" . -}}
	}
}
{{ end }}
{{ end }}

// Application Logger Function

{{ range $idx, $value := .Levels }}
{{ if and $value.AsPrint $value.HasOnErr }}
func (e *onErrLogger) {{$value.FuncName}}(v ...interface{}) (rtn Return) {
	rtn.err = e.err
	if e.popOnErr(v) {
		e.b.{{$value.FuncName}}(v...)
	}
	return rtn
}
{{ end }}
{{ if and $value.AsPrintf $value.HasOnErr }}
func (e *onErrLogger) {{$value.FuncName}}f(f string, v ...interface{}) (rtn Return) {
	rtn.err = e.err
	if e.popOnErr(v) {
		e.b.{{$value.FuncName}}f(f, v...)
	}
	return rtn
}
{{ end }}
{{ if and $value.AsPrintln $value.HasOnErr }}
func (e *onErrLogger) {{$value.FuncName}}ln(v ...interface{}) (rtn Return) {
	rtn.err = e.err
	if e.popOnErr(v) {
		e.b.{{$value.FuncName}}ln(v...)
	}
	return rtn
}
{{ end }}
{{ end }}
`))

var genTestTemplate = template.Must(template.New("").Funcs(template.FuncMap{
	"fmtPrint": func(ss []string) string {
		var ii = make([]interface{}, len(ss))
		for i, v := range ss {
			ii[i] = v
		}
		return fmt.Sprint(ii...)
	},
	"fmtPrintf": func(s string, ss []string) string {
		var ii = make([]interface{}, len(ss))
		for i, v := range ss {
			ii[i] = v
		}
		return fmt.Sprintf(s, ii...)
	},
	"fmtPrintln": func(ss []string) string {
		var ii = make([]interface{}, len(ss))
		for i, v := range ss {
			ii[i] = v
		}
		o := fmt.Sprintln(ii...)
		return o[:len(o)-1]
	},
	"stop": func(s string) string {
		if s == "" {
			return ""
		}
		return "\x1b[0m"
	},
	"iface": func(ss []string) string {
		var ii = make([]interface{}, len(ss))
		for i, v := range ss {
			ii[i] = v
		}
		return fmt.Sprintf("%#v", ii)
	},
}).Parse(`// GENERATED BY ./gen/main.go; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: {{ .Timestamp }} ~~
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
		{{ $text1 := .Text1}}
		{{ $format1 := .Format1}}
		{{- range $idx, $val := .Levels -}}
		{{- if and $val.AsPrint $val.AsPrintTest -}}
		{
			name:   "log.{{ $val.FuncName }}",
			method: log.{{ $val.FuncName }},
			inputs: {{ iface $text1 }},
			{{ if eq $val.FuncName "Fatal" -}}fatal: 888,{{- end }}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrint $text1 }}{{ stop $val.Color }}\n",
		},
		{{- end -}}
		{{ end -}}
		{{- range $idx, $val := .Levels }}
		{{- if and $val.AsPrintf $val.AsPrintfTest -}}
		{
			name:   "log.{{ $val.FuncName }}f",
			method: log.{{ $val.FuncName }}f,
			format: "{{ $format1 }}",
			inputs: {{ iface $text1 }},
			{{ if eq $val.FuncName "Fatal" -}}fatal: 888,{{- end }}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrintf $format1 $text1 }}{{ stop $val.Color }}\n",
		},
		{{- end -}}
		{{ end -}}
		{{- range $idx, $val := .Levels }}
		{{- if and $val.AsPrintln $val.AsPrintlnTest -}}
		{
			name:   "log.{{ $val.FuncName }}ln",
			method: log.{{ $val.FuncName }}ln,
			inputs: {{ iface $text1 }},
			{{ if eq $val.FuncName "Fatal" -}}fatal: 888,{{- end }}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrintln $text1 }}{{ stop $val.Color }}\n",
		},
		{{- end -}}
		{{ end -}}
	}

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

{{ $o := .}}
{{ range $k, $v := .OnErr -}}
func {{ $k }}(t *testing.T) {
	have := struct {
		output *bytes.Buffer
		exitInt int
		rtn Return
	}{
		output: new(bytes.Buffer),
		exitInt: 100,
	}
	want := struct { 
		hasErr bool 
		exitInt int
	}{ 
		hasErr:  {{ $v.HasErr }},
		exitInt: {{ $v.ExitInt }},
	}
	
	log := New(WithOutput(have.output), WithTimeText("Jan-01-2000")).OnErr({{ $v.Err }})

	tests := []struct {
		name   string
		prefix string
		format string
		method interface{} // either func(...interface{}) or func(f, ...interface{})
		inputs []interface{}
		want   string
	}{
		{{ $text1 := $o.Text1}}
		{{ $format1 := $o.Format1}}
		{{- range $idx, $val := $o.Levels -}}
		{{- if and $val.AsPrint $val.AsPrintTest -}}
		{
			name:   "log.{{ $val.FuncName }} {{ $v.Name }}",
			method: log.{{ $val.FuncName }},
			inputs: {{ iface $text1 }},
			{{ if eq $v.HasErr "true" -}}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrint $text1 }}{{ stop $val.Color }}\n",
			{{- else -}}
			want:   "",
			{{ end }}
		},
		{{- end -}}
		{{ end -}}
		{{- range $idx, $val := $o.Levels -}}
		{{- if and $val.AsPrintf $val.AsPrintfTest -}}
		{
			name:   "log.{{ $val.FuncName }}f {{ $v.Name }}",
			method: log.{{ $val.FuncName }}f,
			format: "{{ $format1 }}",
			inputs: {{ iface $text1 }},
			{{ if eq $v.HasErr "true" -}}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrintf $format1 $text1 }}{{ stop $val.Color }}\n",
			{{- else -}}
			want:   "",
			{{ end }}
		},
		{{- end -}}
		{{ end -}}
		{{- range $idx, $val := $o.Levels -}}
		{{- if and $val.AsPrintln $val.AsPrintlnTest -}}
		{
			name:   "log.{{ $val.FuncName }}ln {{ $v.Name }}",
			method: log.{{ $val.FuncName }}ln,
			inputs: {{ iface $text1 }},
			{{ if eq $v.HasErr "true" -}}
			want:   "Jan-01-2000 {{ $val.Color }}{{ with (index $val.Levels "default") }}{{ printf "%s " . }}{{ end }}{{ fmtPrintln $text1 }}{{ stop $val.Color }}\n",
			{{- else -}}
			want:   "",
			{{ end }}
		},
		{{- end -}}
		{{ end -}}
	}

	for _, test := range tests {
		have.output.Reset()
		t.Run(test.name, func(tt *testing.T) {
			if test.name == "log.Fatal {{ $v.Name }}" {
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
			if test.name == "log.Fatal {{ $v.Name }}" {
				if have.exitInt != want.exitInt {
					tt.Errorf("\nhave: %d\nwant: %d\n", have.exitInt, want.exitInt)
				}
			}
		})
	}
}
{{ end }}`))
