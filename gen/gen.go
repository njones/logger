// Logger Method Gen is a tool to automate the creation of methods that satisfy the Logger
// interface. It looks for constants of type logLevel for which to derive the name of
// the methods that can be generated in the format:
//
// const LevelPrint logLevel = iota //`short:"INF" long:"Info" color:"green"`
//
// will produce:
//
// func (*baseLogger) Print(v ...interface{}) {}
// func (*baseLogger) Printf(format string, v ...interface{}) {}
// func (*baseLogger) Println(v ...interface{}) {}
//
// func (ll logLevel) ToString(kind levelType) string {
// 	switch kind {
// 	case LevelLongStr:
// 		switch ll {
// 		case LevelPrint:
// 			return "Info: "
// 		}
// 	case LevelShortStr:
// 		switch ll {
// 		case LevelPrint:
// 			return "INF "
// 		}
// 	case LevelShortBracketStr:
// 		switch ll {
// 		case LevelPrint:
// 			return "[INF] "
// 		}
// 	}
//
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type logLevelData struct {
	Raw      string
	Constant string
	Level    string
	Short    string
	Long     string
	Color    string

	AsPrint   bool
	AsPrintf  bool
	AsPrintln bool
}

type escStrData struct {
	Raw      string
	Name     string
	Value    int
	EscValue string
}

var blFields []string
var escs = map[string]map[string]escStrData{
	"colorType":  {},
	"formatType": {},
}
var loglevels []logLevelData
var leveltypes []string

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "../logger.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	gendFile, err := os.Create("../logger_generated.go")
	if err != nil {
		log.Fatal(err)
	}
	defer gendFile.Close()
	defer func() {
		cmd := exec.Command("gofmt", "-w", "-s", "../logger_generated.go")
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal("gofmt logger generated error: ", err)
		}
		log.Println("go Logger code generated and formatted.")
	}()

	gendTestFile, err := os.Create("../logger_generated_test.go")
	if err != nil {
		log.Fatal(err)
	}
	defer gendTestFile.Close()
	defer func() {
		cmd := exec.Command("gofmt", "-w", "-s", "../logger_generated_test.go")
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal("gofmt logger generated tests error: ", err)
		}
		log.Println("go Logger test code generated and formatted.")
	}()
	ast.Walk(VisitorFunc(FindTypes), node)

	outLog := gendFile //io.MultiWriter(os.Stdout, gendFile) //
	genTemplate.Execute(outLog, struct {
		Timestamp        time.Time
		Escs             map[string]map[string]escStrData
		Loglevels        []logLevelData
		LevelTypes       []string
		BaseLoggerFields []string
	}{
		Timestamp:        time.Now(),
		Escs:             escs,
		Loglevels:        loglevels,
		LevelTypes:       leveltypes,
		BaseLoggerFields: blFields,
	})

	outTestLog := gendTestFile // io.MultiWriter(os.Stdout, gendTestFile) //
	genTestTemplate.Execute(outTestLog, struct {
		Timestamp  time.Time
		Escs       map[string]map[string]escStrData
		Loglevels  []logLevelData
		LevelTypes []string
	}{
		Timestamp:  time.Now(),
		Escs:       escs,
		Loglevels:  loglevels,
		LevelTypes: leveltypes,
	})
}

// VisitorFunc a type
type VisitorFunc func(n ast.Node) ast.Visitor

// Visit does the node walking
func (f VisitorFunc) Visit(n ast.Node) ast.Visitor { return f(n) }

func findType_TypeSpec(node *ast.TypeSpec) ast.Visitor {
	var s *ast.StructType
	var ok bool
	if s, ok = node.Type.(*ast.StructType); !ok {
		return VisitorFunc(FindTypes)
	}
	if node.Name.Name != "baseLogger" {
		return VisitorFunc(FindTypes)
	}
	for _, f := range s.Fields.List {
		blFields = append(blFields, f.Names[0].Name)
	}
	return nil
}

func findType_GenDecl(n *ast.GenDecl) ast.Visitor {
	if n.Tok != token.CONST {
		return VisitorFunc(FindTypes)
	}

	var (
		currType    string
		currComment string
		currValue   string

		val int
		err error
	)

	for _, spec := range n.Specs {
		vspec := spec.(*ast.ValueSpec)
		for _, name := range vspec.Names {
			if ident, ok := vspec.Type.(*ast.Ident); ok {
				currType = ident.Name
			}

			if comment := vspec.Comment; comment != nil {
				for _, commGrp := range comment.List {
					currComment = commGrp.Text // assume only one comment, otherwise we're overwritting here
				}
			}

			if values := vspec.Values; values != nil {
				for _, expression := range values {
					currValue = exprToStr(expression)
				}
			}

			raw := fmt.Sprintf("%s (%s:%q) = %s\n", name.Name, currType, currComment, currValue)

			switch currType {
			case "levelType":
				leveltypes = append(leveltypes, name.Name)
			case "logLevel":
				cMap := commTagToMap(currComment)
				level := strings.TrimPrefix(name.Name, "Level")
				if _, ok := cMap["long"]; !ok {
					cMap["long"] = level
				}

				if cMap["short"] == "-" {
					cMap["short"] = ""
				}

				if cMap["long"] == "-" {
					cMap["long"] = ""
				}
				p, f, ln := true, true, true
				if fnValues, ok := cMap["fn"]; ok {
					p, f, ln = false, false, false
					for _, fnVal := range strings.Split(fnValues, ",") {
						switch fnVal {
						case "-":
							continue // keep them all false same as "empty"
						case "p":
							p = true
						case "f":
							f = true
						case "ln":
							ln = true
						}
					}
				}
				lld := logLevelData{
					Raw:       raw,
					Constant:  name.Name,
					Level:     level,
					Short:     cMap["short"],
					Long:      cMap["long"],
					Color:     cMap["color"],
					AsPrint:   p,
					AsPrintf:  f,
					AsPrintln: ln,
				}
				loglevels = append(loglevels, lld)
			case "formatType", "colorType":
				if currValue != "" {
					val, err = strconv.Atoi(currValue)
					if err != nil {
						// this assumes a "iota + x" or "iota << 1" is what is breaking it
						strVal := strings.Split(currValue, " ")
						val, err = strconv.Atoi(strVal[len(strVal)-1])
						if err != nil {
							log.Printf("second err parsing: %v\n", currValue)
							continue
						}
					}
				} else {
					val++
				}

				if name.Name == "_" {
					continue
				}

				var valEscStr string
				if val != 255 {
					valEscStr = fmt.Sprintf(`\x1b[%dm`, val)
				}
				esd := escStrData{
					Raw:      "",
					Name:     name.Name,
					Value:    val,
					EscValue: valEscStr,
				}
				escs[currType][name.Name] = esd
			}

			currComment = ""
			currValue = ""

		}
	}
	return VisitorFunc(FindTypes)

}

// FindTypes loops through the nodes
func FindTypes(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		return VisitorFunc(FindTypes)
	case *ast.File:
		return VisitorFunc(FindTypes)
	case *ast.TypeSpec:
		return findType_TypeSpec(n)
	case *ast.GenDecl:
		return findType_GenDecl(n)
	}
	return nil
}

func commTagToMap(s string) map[string]string {
	m := make(map[string]string)

	// the following code assumes the tags will not
	// have a space inside the quoted values
	s = strings.TrimPrefix(s, "//")
	s = strings.Trim(s, "`")
	ss := strings.Split(s, " ")
	for _, v := range ss {
		kv := strings.Split(v, ":")
		if len(kv) != 2 {
			log.Printf("%q", s)
			continue
		}
		m[kv[0]] = strings.Trim(kv[1], `"`)
	}
	return m
}

func exprToStr(a ast.Expr) (out string) {
	switch vv := a.(type) {
	case *ast.Ident:
		if vv.Name == "iota" {
			return "0"
		}
		return vv.Name
	case *ast.UnaryExpr:
		x := exprToStr(vv.X)
		return vv.Op.String() + x
	case *ast.BasicLit:
		return vv.Value
	case *ast.BinaryExpr:
		x := exprToStr(vv.X)
		y := exprToStr(vv.Y)
		return x + " " + vv.Op.String() + " " + y
	}
	return fmt.Sprintf("%T - %#[1]v", a)
}

func levelTypeStr(a string, b logLevelData) string {
	if len(b.Long) == 0 {
		return ""
	}
	switch a {
	case "LevelLongStr":
		return b.Long + ": "
	case "LevelShortStr":
		return b.Short + " "
	case "LevelShortBracketStr":
		return "[" + b.Short + "] "
	}
	return "-unknown-"
}

func colorStr(p, s string) string {
	return p + "Color" + strings.Title(s)
}

var funcMap = template.FuncMap{"lt_str": levelTypeStr, "color": colorStr}
var genTemplate = template.Must(template.New("").Funcs(funcMap).Parse(`	// go generate gen
// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: {{ .Timestamp }} ~~

package logger

import (
	"sync"
	"net/http"
)

// Logger the main interface
type Logger interface {
{{- range $idx, $value := .Loglevels  }}
{{- if $value.AsPrint }}
	{{$value.Level}}(...interface{}) Return
{{- end -}}
{{- if $value.AsPrintf }}
	{{$value.Level}}f(string, ...interface{}) Return
{{- end -}}
{{- if $value.AsPrintln }}
	{{$value.Level}}ln(...interface{}) Return
{{- end -}}
{{end}}
	Color(colorType) Logger
	FatalInt(int) Logger
	Field(string, interface{}) Logger
	Fields(map[string]interface{}) Logger
	HTTPMiddleware(http.Handler) http.Handler
	NoColor() Logger
	OnErr(error) Logger
	Suppress(logLevel) Logger
	With(...optFunc) Logger
}

var levelTypeToStringMap = map[levelType]map[logLevel]string{
	{{- range $idx1, $value1 := .LevelTypes }}
	{{$value1}}: map[logLevel]string{
	{{- range $idx2, $value2 := $.Loglevels }}
		{{$value2.Constant}}: "{{lt_str $value1 $value2}}",
	{{- end }}
	},
	{{- end }}
}

func (ll logLevel) ToString(kind levelType) string {
	if knd, ok := levelTypeToStringMap[kind]; ok {
		if str, ok := knd[ll]; ok {
			return str
		}
	}

	return "unknown"
}

// copyBaseLogger returns a deep copy of the baseLogger
func copyBaseLogger(l *baseLogger) *baseLogger {
	return &baseLogger{
		{{- range $idx, $value := .BaseLoggerFields }}
		{{$value}}:         l.{{$value}},
		{{- end }}
	}
} 

func (l *baseLogger) levelStr(ll logLevel) string {
	return ll.ToString(l.logLevel)
}

{{ range $idx, $value := .Loglevels }}
{{ if $value.AsPrint }}
func (l *baseLogger) {{$value.Level}}(v ...interface{}) Return {
	if l.skip&{{$value.Constant}} != 0 {
		return l.rtn()
	}
	ctx := context{
		is:          asPrint,
		tsStrCh:     tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:      [3]ESCStringer{{color "{" $value.Color}}, l.color, SeqReset},
		level:       {{$value.Constant}},
		levelStr:    l.levelStr({{$value.Constant}}),
		values:      v,
		{{- if (eq $value.Long "Panic") }}
		panicCh:     make(chan string, 1),
		{{ end -}}
		wg:          &sync.WaitGroup{},
	}

	if l.color == ColorUnk {
		ctx.colors = [3]ESCStringer{ColorUnk, ColorUnk, ColorUnk}
	}

	for k, val := range l.kv {
		if ctx.kvMap == nil {
			ctx.kvMap = make(map[string]interface{})
		}
		ctx.kvMap[k] = val
	}
	for i, val := range v {
		if f, ok := val.(KVStruct); ok {
			if ctx.kvMap == nil {
				ctx.kvMap = make(map[string]interface{})
			}
			ctx.kvMap[f.key] = f.value
			v[i] = ""
		}
	}

	ctx.wg.Add(1)
	l.to <- ctx
	ctx.wg.Wait()
	{{- if (eq $value.Long "Panic") }}
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
	{{ end -}}
	{{- if (eq $value.Long "Fatal") }}
	l.fatal(l.fatali)
	{{ end }}
	return l.rtn()
}
{{ end }}
{{ if $value.AsPrintf }}
func (l *baseLogger) {{$value.Level}}f(format string, v ...interface{}) Return {
	if l.skip&{{$value.Constant}} != 0 {
		return l.rtn()
	}
	ctx := context{
		is:          asPrintf,
		tsStrCh:     tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr:   format,
		colors:      [3]ESCStringer{{color "{" $value.Color}}, l.color, SeqReset},
		level:       {{$value.Constant}},
		levelStr:    l.levelStr({{$value.Constant}}),
		values:      make([]interface{}, 0, len(v) * 2),
		{{- if (eq $value.Long "Panic") }}
		panicCh:     make(chan string, 1),
		{{ end -}}
		wg:          &sync.WaitGroup{},
	}

	if l.color == ColorUnk {
		ctx.colors = [3]ESCStringer{ColorUnk, ColorUnk, ColorUnk}
	}

	for k, val := range l.kv {
		if ctx.kvMap == nil {
			ctx.kvMap = make(map[string]interface{})
		}
		ctx.kvMap[k] = val
	}
	for i, val := range v {
		if f, ok := val.(KVStruct); ok {
			if ctx.kvMap == nil {
				ctx.kvMap = make(map[string]interface{})
			}
			ctx.kvMap[f.key] = f.value
			continue
		}
		ctx.values = append(ctx.values, v[i])
	}

	ctx.wg.Add(1)
	l.to <- ctx
	ctx.wg.Wait()
	{{- if (eq $value.Long "Panic") }}
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
	{{ end -}}
	{{- if (eq $value.Long "Fatal") }}
	l.fatal(l.fatali)
	{{ end }}
	return l.rtn()
}
{{ end }}
{{ if $value.AsPrintln }}
func (l *baseLogger) {{$value.Level}}ln(v ...interface{}) Return {
	if l.skip&{{$value.Constant}} != 0 {
		return l.rtn()
	}
	ctx := context{
		is:          asPrintln,
		{{- if (eq $value.Constant "LevelHTTP") }}
		tsStrCh:     tsChanText(""), // sending back empty text only for HTTPln, because we don't want to display it
		{{- else}}
		tsStrCh:     tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		{{- end }}
		colors:      [3]ESCStringer{{color "{" $value.Color}}, l.color, SeqReset},
		level:       {{$value.Constant}},
		levelStr:    l.levelStr({{$value.Constant}}),
		values:      make([]interface{}, 0, len(v) * 2),
		{{- if (eq $value.Long "Panic") }}
		panicCh:     make(chan string, 1),
		{{ end -}}
		wg:          &sync.WaitGroup{},
	}

	if l.color == ColorUnk {
		ctx.colors = [3]ESCStringer{ColorUnk, ColorUnk, ColorUnk}
	}

	for k, val := range l.kv {
		if ctx.kvMap == nil {
			ctx.kvMap = make(map[string]interface{})
		}
		ctx.kvMap[k] = val
	}
	var sp = " "
	for i, val := range v {
		if f, ok := val.(KVStruct); ok {
			if ctx.kvMap == nil {
				ctx.kvMap = make(map[string]interface{})
			}
			ctx.kvMap[f.key] = f.value
			continue
		}
		if i == 0 {
			ctx.values = append(ctx.values, v[i])
		} else {
			ctx.values = append(ctx.values, []interface{}{sp, v[i]}...)
		}
	}

	ctx.wg.Add(1)
	l.to <- ctx
	ctx.wg.Wait()
	{{- if (eq $value.Long "Panic") }}
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
	{{ end -}}
	{{- if (eq $value.Long "Fatal") }}
	l.fatal(l.fatali)
	{{ end }}
	return l.rtn()
}
{{- end -}}
{{ end}}

{{- range $idx, $value := .Loglevels }}
{{- if $value.AsPrint }}
func (l nilLogger) {{$value.Level}}(v ...interface{}) Return                 { return Return{} }
{{- end -}}
{{- if $value.AsPrintf }}
func (l nilLogger) {{$value.Level}}f(format string, v ...interface{}) Return { return Return{} }
{{- end -}}
{{- if $value.AsPrintln }}
func (l nilLogger) {{$value.Level}}ln(v ...interface{}) Return               { return Return{} }
{{- end -}}
{{- end}}
func (l nilLogger) Color(colorType) Logger               { return l }
func (l nilLogger) Field(string, interface{}) Logger     { return l }
func (l nilLogger) Fields(map[string]interface{}) Logger { return l }
func (l nilLogger) NoColor() Logger                      { return l }
func (l nilLogger) OnErr(error) Logger                   { return l }
func (l nilLogger) Suppress(logLevel) Logger             { return l }
func (l nilLogger) With(...optFunc) Logger               { return l }
func (l nilLogger) FatalInt(int) Logger                  { return l }

func (l nilLogger) HTTPMiddleware(h http.Handler) http.Handler { return h }

{{ range $key, $value1 := .Escs}}
func (t {{$key}}) ESCStr() string {
	switch t {
	{{- range $idx, $value2 := $value1}}
	case {{$value2.Name}}:
		return "{{$value2.EscValue}}"
	{{- end}}
	}
	return "\x1b[0munknown"
}
{{end}}
`))

var genTestTemplate = template.Must(template.New("").Funcs(funcMap).Parse(`	// go generate
	// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
	// GENERATED ON: {{ .Timestamp }}

	package logger

	import (
		"testing"
		"bytes"
	)

	{{ range $idx, $value := .Loglevels }}
	{{- if and $value.AsPrint (ne $value.Long "Panic")}}
	{{- $colorType := (color "" $value.Color) -}}
	func Test{{$value.Level}}(t *testing.T) {
		want := "TEST {{ (index (index $.Escs "colorType") $colorType).EscValue }}{{$value.Long}}: This is an automated test\x1b[0m\n"

		have := new(bytes.Buffer)
		l := New(WithOutput(have), WithTimeText("TEST"){{- if (eq $value.Long "Fatal") }}, withFatal{{ end -}})
		l.{{$value.Level}}("This ", "is ", "an ", "automated ", "test")

		if want != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
		}
	}

	func Test{{$value.Level}}f(t *testing.T) {
		want := "TEST {{ (index (index $.Escs "colorType") $colorType).EscValue }}{{$value.Long}}: This is an automated test\x1b[0m\n"

		have := new(bytes.Buffer)
		l := New(WithOutput(have), WithTimeText("TEST"){{- if (eq $value.Long "Fatal") }}, withFatal{{ end -}})
		l.{{$value.Level}}f("This is an %s test", "automated")

		if want != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
		}
	}

	func Test{{$value.Level}}ln(t *testing.T) {
		want := "TEST {{ (index (index $.Escs "colorType") $colorType).EscValue }}{{$value.Long}}: This is an automated test\x1b[0m\n"

		have := new(bytes.Buffer)
		l := New(WithOutput(have), WithTimeText("TEST"){{- if (eq $value.Long "Fatal") }}, withFatal{{ end -}})
		l.{{$value.Level}}ln("This", "is", "an", "automated", "test")

		if want != have.String() {
			t.Errorf("\nwant: %q\n\nhave: %q\n", want, have.String())
		}
	}

	func TestNilLogger{{$value.Level}}(t *testing.T) {
		l := new(nilLogger)
		l.{{$value.Level}}("does", "not", "work")
	}
	
	func TestNilLogger{{$value.Level}}f(t *testing.T) {
		l := new(nilLogger)
		l.{{$value.Level}}f("%s %s %s", "does", "not", "work")
	}

	func TestNilLogger{{$value.Level}}ln(t *testing.T) {
		l := new(nilLogger)
		l.{{$value.Level}}ln("does", "not", "work")
	}
	{{ end }}
	{{ end }}

	func TestNilLoggerColor(t *testing.T) {
		l := new(nilLogger)
		l.Color(ColorBlack)
	}

	func TestNilLoggerField(t *testing.T) {
		l := new(nilLogger)
		l.Field("nil", struct{ Logging string }{ "made up" })
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
`))
