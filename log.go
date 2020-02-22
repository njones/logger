package logger

import (
	"bufio"
	"bytes"
	"io"
	stdlog "log"
	"regexp"
	"sync"
)

type rxMap map[*regexp.Regexp]func(Logger, map[string]string)

type syncWriter struct {
	w  io.Writer
	sync *sync.WaitGroup
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	sw.sync.Add(bytes.Count(p, []byte("\n")))
	defer sw.sync.Wait()

	n, err = sw.w.Write(p)
	return
}

func byteCounter(sw *syncWriter) (func([]byte, bool) (int, []byte, error), func()) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		return
	}, func() { sw.sync.Done() }
}

func NewLog(rxm rxMap, opts ...optFunc) *stdlog.Logger {
	logger := New(opts...)

	pr, pw := io.Pipe()
	sw := &syncWriter{w: pw, sync: new(sync.WaitGroup)}
	split, waitForPrintToFinishBeforeTheNextScan := byteCounter(sw)

	go func() {
		scan := bufio.NewScanner(pr)
		scan.Split(split)

		for scan.Scan() {
			text := scan.Text()
			for rx, fn := range rxm {
				if rx == nil {
					continue
				}
				if m, ok := match(rx, text); ok {
					fn(logger, m)
					goto Wait
				}
			}
			logger.Println(text)
		Wait:
			waitForPrintToFinishBeforeTheNextScan()
		}
	}()

	return stdlog.New(sw, "", 0)
}

func match(rx *regexp.Regexp, txt string) (map[string]string, bool) {
	m := make(map[string]string)
	matches := rx.FindStringSubmatch(txt)
	ok := matches != nil
	if ok {
		for i, name := range rx.SubexpNames() {
			if i != 0 && name != "" {
				m[name] = matches[i]
			}
		}
	}
	return m, ok
}
