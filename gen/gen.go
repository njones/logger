package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// gen.go generates the logger functions

var genFileIn = "logger_generated.go"
var genFileOut = "logger.go"

func minus(a, b int) int { return a - b }

func main() {
	lf, err := os.Create(genFileIn)
	if err != nil {
		log.Fatal(err)
	}
	defer lf.Close()

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, genFileOut, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments
	// and AST nodes.
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	internal := struct {
		Timestamp       time.Time
		LogLevels       map[string]uint
		LogLevelsShort  map[string]uint
		LogColors       map[string]string
		LogHasESCColors map[string]string
	}{
		Timestamp:       time.Now().UTC(),
		LogLevels:       make(map[string]uint),
		LogLevelsShort:  make(map[string]uint),
		LogColors:       make(map[string]string),
		LogHasESCColors: make(map[string]string),
	}

	rx := regexp.MustCompile("//\\s+`(.*)`")

	for kk, vv := range cmap {

		if val, ok := kk.(*ast.ValueSpec); ok {

			foundComment := rx.FindAllStringSubmatch(vv[0].List[0].Text, 1)
			if len(foundComment) > 0 && len(foundComment[0]) > 1 {
				commentTag := foundComment[0][1]
				keyvals := strings.Split(commentTag, ",")
				for _, kv := range keyvals {
					kvs := strings.Split(kv, ":")
					k := kvs[0]
					// v := strings.Trim(kvs[1], `"`)

					if k == "gen.color" {
						internal.LogColors[val.Names[0].Name] = val.Names[0].Name
					}
				}
			}
		}

		if val, ok := kk.(*ast.Field); ok {

			foundComment := rx.FindAllStringSubmatch(vv[0].List[0].Text, 1)
			if len(foundComment) > 0 && len(foundComment[0]) > 1 {
				commentTag := foundComment[0][1]
				keyvals := strings.Split(commentTag, ",")
				for _, kv := range keyvals {
					kvs := strings.Split(kv, ":")
					k := strings.TrimSpace(kvs[0])
					v := strings.Trim(strings.TrimSpace(kvs[1]), `"`)

					if k == "gen.show" {
						internal.LogHasESCColors[val.Names[0].Name] = v
					}
				}
			}
		}

		if val, ok := kk.(*ast.GenDecl); ok {

			if val.Tok == token.TYPE {
				if val.Specs[0].(*ast.TypeSpec).Name.Name == "level" {
					for k3, v3 := range val.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {

						var short string
						foundComment := rx.FindAllStringSubmatch(v3.Comment.List[0].Text, 1)
						if len(foundComment) > 0 && len(foundComment[0]) > 1 {
							commentTag := foundComment[0][1]
							keyvals := strings.Split(commentTag, ",")
							for _, kv := range keyvals {
								kvs := strings.Split(kv, ":")
								k := strings.TrimSpace(kvs[0])
								v := strings.Trim(strings.TrimSpace(kvs[1]), `"`)

								if k == "gen.short" {
									short = v
								}
							}
						}

						internal.LogLevels[v3.Names[0].Name] = 1 << uint(k3)
						internal.LogLevelsShort[short] = 1 << uint(k3)
					}
				}
			}
		}
	}

	err = packageTemplate.Execute(lf, internal)
	log.Println("Done:", err)
}

var funcMap = template.FuncMap{"minus": minus}
var packageTemplate = template.Must(template.New("").Funcs(funcMap).Parse(`// go generate
// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: {{ .Timestamp }} ~~

package logger

import (
	"fmt"
	"net/http"
)

// String is the string representation of the color
func (lc LogColor) String() string {
	switch lc {
	{{-  range $key, $value := .LogColors }}
	case {{ $value }}:
		return "{{ $key }}"
	{{- end}}
	}

	return "unknown"
}

// color2ESC returns the VT100 escape codes for a color
func color2ESC(color LogColor) string {
	return fmt.Sprintf("\x1b[%dm", int32(color))
}

// Level returns the log level used
func Level() (lvl level) {
{{-  range $key, $value := .LogLevels }}
	lvl.{{ $key }} = {{ $value }}
{{- end}}
	return lvl
}

// String is the string representation of the log level
func (ll LogLevel) String() string {
	switch ll {
	{{-  range $key, $value := .LogLevels }}
	case {{ $value }}:
		return "{{ $key }}"
	{{- end}}
	}

	return "unknown"
}

// Short is the short three letter abbreviation of the log level
func (ll LogLevel) Short() string {
	switch ll {
	{{-  range $key, $value := .LogLevelsShort }}
	case {{ $value }}:
		return "{{ $key }}"
	{{- end}}
	}

	return "unknown"
}

// Logger is the main interface that is presented as a logger
type Logger interface {
	Color(LogColor) Logger
	OnErr(error) Logger
	HTTPMiddleware(next http.Handler) http.Handler
	Suppress()
	UnSuppress()
{{ range $key, $value := .LogLevels }}
	{{ $key }}(...interface{})
	{{ $key }}f(string, ...interface{})
{{- end}}
}
{{ range $key, $value := .LogLevels }}
// {{ $key }} is the generated logger function to satisfy the interface
func (l *logger) {{ $key }}(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC({{ index $.LogHasESCColors $key }})
	}
	l.println({{ $value }}, iface...)
}

// {{ $key }}f is the generated logger function to satisfy the interface
func (l *logger) {{ $key }}f(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC({{ index $.LogHasESCColors $key }})
	}
	l.printf({{ $value }}, fmt, iface...)
}
{{- end}}

// HTTPMiddleware is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) HTTPMiddleware(next http.Handler) (r http.Handler) {
	return next
}

// Suppress is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Suppress()               {}

// UnSuppress is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) UnSuppress()             {}

// Color is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Color(x LogColor) Logger { return l }

// OnErr is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) OnErr(x error) Logger    { return l }
{{- range $key, $value := .LogLevels }}
{{- $klen := (minus 19 (len $key)) }}{{- $kfmt := printf "%%%ds" $klen }}
{{- $kflen := (minus 6 (len $key)) }}{{- $kffmt := printf "%%%ds" $kflen }}

// {{ $key }} is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) {{ $key }}(iface ...interface{}){{ printf $kfmt "" }}{ return }

// {{ $key }}f is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) {{ $key }}f(fmt string, iface ...interface{}){{ printf $kffmt "" }}{ return }
{{- end}}

`))
