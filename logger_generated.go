// go generate gen
// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: 2018-01-20 15:24:23.014698883 +0000 UTC m=+0.011261566 ~~

package logger

import (
	"net/http"
	"sync"
)

// Logger the main interface
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})
	Trace(...interface{})
	Tracef(string, ...interface{})
	Traceln(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
	HTTPln(...interface{})
	Color(colorType) Logger
	Field(string, interface{}) Logger
	Fields(map[string]interface{}) Logger
	HTTPMiddleware(http.Handler) http.Handler
	NoColor() Logger
	OnErr(error) Logger
	Suppress(logLevel) Logger
	With(...optFunc) Logger
}

func (ll logLevel) ToString(kind levelType) string {
	switch kind {
	case LevelLongStr:
		switch ll {
		case LevelPrint:
			return "Info: "
		case LevelInfo:
			return "Info: "
		case LevelWarn:
			return "Warn: "
		case LevelDebug:
			return "Debug: "
		case LevelError:
			return "Error: "
		case LevelTrace:
			return "Trace: "
		case LevelFatal:
			return "Fatal: "
		case LevelPanic:
			return "Panic: "
		case LevelHTTP:
			return ""
		}
	case LevelShortStr:
		switch ll {
		case LevelPrint:
			return "INF "
		case LevelInfo:
			return "INF "
		case LevelWarn:
			return "WRN "
		case LevelDebug:
			return "DBG "
		case LevelError:
			return "ERR "
		case LevelTrace:
			return "TRC "
		case LevelFatal:
			return "FAT "
		case LevelPanic:
			return "PAN "
		case LevelHTTP:
			return ""
		}
	case LevelShortBracketStr:
		switch ll {
		case LevelPrint:
			return "[INF] "
		case LevelInfo:
			return "[INF] "
		case LevelWarn:
			return "[WRN] "
		case LevelDebug:
			return "[DBG] "
		case LevelError:
			return "[ERR] "
		case LevelTrace:
			return "[TRC] "
		case LevelFatal:
			return "[FAT] "
		case LevelPanic:
			return "[PAN] "
		case LevelHTTP:
			return ""
		}
	}
	return "unknown"
}

// copyBaseLogger returns a deep copy of the baseLogger
func copyBaseLogger(l *baseLogger) *baseLogger {
	return &baseLogger{
		o:             l.o,
		stdout:        l.stdout,
		stderr:        l.stderr,
		to:            l.to,
		ts:            l.ts,
		tsIsUTC:       l.tsIsUTC,
		tsText:        l.tsText,
		tsFormat:      l.tsFormat,
		logLevel:      l.logLevel,
		skip:          l.skip,
		hasFilter:     l.hasFilter,
		kv:            l.kv,
		color:         l.color,
		httpk:         l.httpk,
		httpLogFormat: l.httpLogFormat,
		marshal:       l.marshal,
		fatal:         l.fatal,
	}
}

func (l *baseLogger) levelStr(ll logLevel) string {
	return ll.ToString(l.logLevel)
}

func (l *baseLogger) Print(v ...interface{}) {
	if l.skip&LevelPrint != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:    LevelPrint,
		levelStr: l.levelStr(LevelPrint),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Printf(format string, v ...interface{}) {
	if l.skip&LevelPrint != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:     LevelPrint,
		levelStr:  l.levelStr(LevelPrint),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Println(v ...interface{}) {
	if l.skip&LevelPrint != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:    LevelPrint,
		levelStr: l.levelStr(LevelPrint),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Info(v ...interface{}) {
	if l.skip&LevelInfo != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:    LevelInfo,
		levelStr: l.levelStr(LevelInfo),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Infof(format string, v ...interface{}) {
	if l.skip&LevelInfo != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:     LevelInfo,
		levelStr:  l.levelStr(LevelInfo),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Infoln(v ...interface{}) {
	if l.skip&LevelInfo != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:    LevelInfo,
		levelStr: l.levelStr(LevelInfo),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Warn(v ...interface{}) {
	if l.skip&LevelWarn != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorYellow, l.color, SeqReset},
		level:    LevelWarn,
		levelStr: l.levelStr(LevelWarn),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Warnf(format string, v ...interface{}) {
	if l.skip&LevelWarn != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorYellow, l.color, SeqReset},
		level:     LevelWarn,
		levelStr:  l.levelStr(LevelWarn),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Warnln(v ...interface{}) {
	if l.skip&LevelWarn != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorYellow, l.color, SeqReset},
		level:    LevelWarn,
		levelStr: l.levelStr(LevelWarn),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Debug(v ...interface{}) {
	if l.skip&LevelDebug != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorCyan, l.color, SeqReset},
		level:    LevelDebug,
		levelStr: l.levelStr(LevelDebug),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Debugf(format string, v ...interface{}) {
	if l.skip&LevelDebug != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorCyan, l.color, SeqReset},
		level:     LevelDebug,
		levelStr:  l.levelStr(LevelDebug),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Debugln(v ...interface{}) {
	if l.skip&LevelDebug != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorCyan, l.color, SeqReset},
		level:    LevelDebug,
		levelStr: l.levelStr(LevelDebug),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Error(v ...interface{}) {
	if l.skip&LevelError != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorMagenta, l.color, SeqReset},
		level:    LevelError,
		levelStr: l.levelStr(LevelError),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Errorf(format string, v ...interface{}) {
	if l.skip&LevelError != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorMagenta, l.color, SeqReset},
		level:     LevelError,
		levelStr:  l.levelStr(LevelError),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Errorln(v ...interface{}) {
	if l.skip&LevelError != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorMagenta, l.color, SeqReset},
		level:    LevelError,
		levelStr: l.levelStr(LevelError),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Trace(v ...interface{}) {
	if l.skip&LevelTrace != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorBlue, l.color, SeqReset},
		level:    LevelTrace,
		levelStr: l.levelStr(LevelTrace),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Tracef(format string, v ...interface{}) {
	if l.skip&LevelTrace != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorBlue, l.color, SeqReset},
		level:     LevelTrace,
		levelStr:  l.levelStr(LevelTrace),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Traceln(v ...interface{}) {
	if l.skip&LevelTrace != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorBlue, l.color, SeqReset},
		level:    LevelTrace,
		levelStr: l.levelStr(LevelTrace),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}

func (l *baseLogger) Fatal(v ...interface{}) {
	if l.skip&LevelFatal != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:    LevelFatal,
		levelStr: l.levelStr(LevelFatal),
		values:   v, wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	l.fatal(1)
}

func (l *baseLogger) Fatalf(format string, v ...interface{}) {
	if l.skip&LevelFatal != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:     LevelFatal,
		levelStr:  l.levelStr(LevelFatal),
		values:    make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	l.fatal(1)
}

func (l *baseLogger) Fatalln(v ...interface{}) {
	if l.skip&LevelFatal != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:    LevelFatal,
		levelStr: l.levelStr(LevelFatal),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	l.fatal(1)
}

func (l *baseLogger) Panic(v ...interface{}) {
	if l.skip&LevelPanic != 0 {
		return
	}
	ctx := context{
		is:       asPrint,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:    LevelPanic,
		levelStr: l.levelStr(LevelPanic),
		values:   v,
		panicCh:  make(chan string, 1),
		wg:       &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
}

func (l *baseLogger) Panicf(format string, v ...interface{}) {
	if l.skip&LevelPanic != 0 {
		return
	}
	ctx := context{
		is:        asPrintf,
		tsStrCh:   tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		formatStr: format,
		colors:    [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:     LevelPanic,
		levelStr:  l.levelStr(LevelPanic),
		values:    make([]interface{}, 0, len(v)*2),
		panicCh:   make(chan string, 1),
		wg:        &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
}

func (l *baseLogger) Panicln(v ...interface{}) {
	if l.skip&LevelPanic != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChan(l.tsText, l.tsFormat, l.ts, l.tsIsUTC),
		colors:   [3]ESCStringer{ColorRed, l.color, SeqReset},
		level:    LevelPanic,
		levelStr: l.levelStr(LevelPanic),
		values:   make([]interface{}, 0, len(v)*2),
		panicCh:  make(chan string, 1),
		wg:       &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
	// firing the panic from here, so it's not swallowed by the go routine
	panic(<-ctx.panicCh)
}

func (l *baseLogger) HTTPln(v ...interface{}) {
	if l.skip&LevelHTTP != 0 {
		return
	}
	ctx := context{
		is:       asPrintln,
		tsStrCh:  tsChanText(""), // sending back empty text only for HTTPln, because we don't want to display it
		colors:   [3]ESCStringer{ColorGreen, l.color, SeqReset},
		level:    LevelHTTP,
		levelStr: l.levelStr(LevelHTTP),
		values:   make([]interface{}, 0, len(v)*2), wg: &sync.WaitGroup{},
	}

	if l.color == UnkColor {
		ctx.colors = [3]ESCStringer{UnkColor, UnkColor, UnkColor}
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
}
func (l nilLogger) Print(v ...interface{})                 {}
func (l nilLogger) Printf(format string, v ...interface{}) {}
func (l nilLogger) Println(v ...interface{})               {}
func (l nilLogger) Info(v ...interface{})                  {}
func (l nilLogger) Infof(format string, v ...interface{})  {}
func (l nilLogger) Infoln(v ...interface{})                {}
func (l nilLogger) Warn(v ...interface{})                  {}
func (l nilLogger) Warnf(format string, v ...interface{})  {}
func (l nilLogger) Warnln(v ...interface{})                {}
func (l nilLogger) Debug(v ...interface{})                 {}
func (l nilLogger) Debugf(format string, v ...interface{}) {}
func (l nilLogger) Debugln(v ...interface{})               {}
func (l nilLogger) Error(v ...interface{})                 {}
func (l nilLogger) Errorf(format string, v ...interface{}) {}
func (l nilLogger) Errorln(v ...interface{})               {}
func (l nilLogger) Trace(v ...interface{})                 {}
func (l nilLogger) Tracef(format string, v ...interface{}) {}
func (l nilLogger) Traceln(v ...interface{})               {}
func (l nilLogger) Fatal(v ...interface{})                 {}
func (l nilLogger) Fatalf(format string, v ...interface{}) {}
func (l nilLogger) Fatalln(v ...interface{})               {}
func (l nilLogger) Panic(v ...interface{})                 {}
func (l nilLogger) Panicf(format string, v ...interface{}) {}
func (l nilLogger) Panicln(v ...interface{})               {}
func (l nilLogger) HTTPln(v ...interface{})                {}
func (l nilLogger) Color(colorType) Logger                 { return l }
func (l nilLogger) Field(string, interface{}) Logger       { return l }
func (l nilLogger) Fields(map[string]interface{}) Logger   { return l }
func (l nilLogger) NoColor() Logger                        { return l }
func (l nilLogger) OnErr(error) Logger                     { return l }
func (l nilLogger) Suppress(logLevel) Logger               { return l }
func (l nilLogger) With(...optFunc) Logger                 { return l }

func (l nilLogger) HTTPMiddleware(h http.Handler) http.Handler { return h }

func (t colorType) ESCStr() string {
	switch t {
	case ColorBlack:
		return "\x1b[30m"
	case ColorBlue:
		return "\x1b[34m"
	case ColorCyan:
		return "\x1b[36m"
	case ColorGreen:
		return "\x1b[32m"
	case ColorMagenta:
		return "\x1b[35m"
	case ColorRed:
		return "\x1b[31m"
	case ColorWhite:
		return "\x1b[37m"
	case ColorYellow:
		return "\x1b[33m"
	case UnkColor:
		return ""
	}
	return "\x1b[0munknown"
}

func (t formatType) ESCStr() string {
	switch t {
	case SeqBlink:
		return "\x1b[5m"
	case SeqBright:
		return "\x1b[1m"
	case SeqDim:
		return "\x1b[2m"
	case SeqHidden:
		return "\x1b[8m"
	case SeqReset:
		return "\x1b[0m"
	case SeqReverse:
		return "\x1b[7m"
	case SeqUnderscore:
		return "\x1b[4m"
	case UnkSeq:
		return ""
	}
	return "\x1b[0munknown"
}
