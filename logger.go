package logger

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
	"time"

	"github.com/njones/logger/kv"
)

var defaultTS = convertStamp("2006-Jan-02")

var mPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

var lPool = sync.Pool{
	New: func() interface{} {
		return &line{}
	},
}

type baseLogger struct {
	flags    int
	depth    int
	display  int
	suppress int

	ts struct {
		now   time.Time
		fns   []func(string) string // for ultimate formatting functions
		stamp string
		text  []byte
	}

	color  []byte
	prefix struct {
		user []byte
	}

	kv struct {
		set     map[string]interface{}
		marshal func(v interface{}) ([]byte, error)
	}

	http struct {
		headers  KVMap
		formatFn HTTPLogFormatFunc
	}

	out struct {
		raw []io.Writer

		w     io.Writer
		cw    io.Writer
		close []io.Closer
	}

	exit struct {
		Int  int
		Func func(int)
		buf  *bytes.Buffer
	}

	sync struct {
		fw *sync.WaitGroup
		ln *sync.Mutex

		fwCnt int
	}
}

func New(opts ...optFunc) Logger {
	b := &baseLogger{}

	b.ts.stamp = defaultTS
	b.kv.marshal = kv.Marshal
	b.kv.set = make(map[string]interface{})
	b.exit.Int = 1
	b.exit.Func = os.Exit
	b.sync.ln = new(sync.Mutex)

	b.writers([]io.Writer{os.Stdout})

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *baseLogger) Flags() int {
	return b.flags
}

func (b *baseLogger) Output(calldepth int, s string) error {
	b.depth = calldepth
	return b.print(bPrint, []interface{}{s})
}

func (b *baseLogger) Prefix() string { return string(b.prefix.user) }

func (b *baseLogger) SetFlags(f int) {
	b.flags = f

	if f == 0 || b.ts.stamp != defaultTS {
		return
	}

	b.ts.stamp = "" // = 2009/01/23 01:23:23
	if (b.flags & Ldate) != 0 {
		b.ts.stamp += convertStamp("2006/01/02")
	}
	if (b.flags & Ltime) != 0 {
		if len(b.ts.stamp) > 0 {
			b.ts.stamp += " "
		}
		b.ts.stamp += convertStamp("15:04:05")
		if (b.flags & Lmicroseconds) != 0 {
			b.ts.stamp = b.ts.stamp[:len(b.ts.stamp)-1] + convertStamp("05.000000") // overwrite previous seconds
		}
	}
}

func (b *baseLogger) SetOutput(w io.Writer) { b.writers([]io.Writer{w}) }

func (b *baseLogger) SetPrefix(s string) { b.prefix.user = []byte(s) }

func (b *baseLogger) Writer() io.Writer { return b.out.w }

func (b *baseLogger) FatalInt(i int) Logger { b.exit.Int = i; return b }

func (b *baseLogger) Field(key string, value interface{}) Logger {
	b.kv.set[key] = value
	return b
}

func (b *baseLogger) Fields(kvs map[string]interface{}) Logger {
	for key, value := range kvs {
		b.kv.set[key] = value
	}
	return b
}

func (b *baseLogger) OnErr(err error) OnErrLogger { return &onErrLogger{b: b, err: err} }

func (b *baseLogger) Suppress(i logLevel) Logger {
	b.suppress = int(i)
	return b
}

func (b *baseLogger) With(opts ...optFunc) Logger {
	bb := duplicate(b)
	for _, opt := range opts {
		opt(bb)
	}
	return bb
}

func (*baseLogger) scan(fn func(string)) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		scan := bufio.NewScanner(pr)
		for scan.Scan() {
			fn(scan.Text())
		}
	}()
	return pw
}

func (b *baseLogger) time() []byte {
	if b.ts.text != nil {
		return b.ts.text
	}

	if b.ts.stamp == "" {
		return nil
	}

	var now = b.ts.now
	if now.IsZero() {
		now = time.Now()
	}

	if (b.flags & LUTC) != 0 {
		now = b.ts.now.UTC()
	}

	var (
		r    rune
		n, i int
		ts   = make([]byte, 0, 20)
	)

	for i, r = range b.ts.stamp {
		if tsFormat, ok := tsRuneMap[r]; ok {
			if i > n {
				ts = append(ts, []byte(b.ts.stamp[n:i])...)
			}
			ts = now.AppendFormat(ts, tsFormat)
			n = i + 1
			continue
		}
	}
	if i >= n {
		ts = append(ts, []byte(b.ts.stamp[n:len(b.ts.stamp)])...)
	}

	if len(b.ts.fns) > 0 {
		s := string(ts)
		for _, fn := range b.ts.fns {
			s = fn(s)
		}
		ts = []byte(s)
	}

	return ts
}

func (b *baseLogger) filter(v []interface{}) (_ []interface{}, kv string, err error) {
	var m = mPool.Get().(map[string]interface{})

	for k, v := range b.kv.set {
		m[k] = v
	}

	fv := v[:0]
	for _, i := range v {
		if p, ok := i.(KeyVal); ok {
			m[string(p.Key)] = p.Value
			continue
		}
		if p, ok := i.(KVMap); ok {
			for key, value := range p {
				m[string(key)] = value
			}
			continue
		}
		if p, ok := i.(map[K]V); ok {
			for key, value := range p {
				m[string(key)] = value
			}
			continue
		}
		fv = append(fv, i)
	}

	if m != nil && len(m) > 0 {
		_kv, err := b.kv.marshal(m)
		if err != nil {
			return nil, "", err
		}
		kv = string(_kv)
		for k := range m {
			delete(m, k)
		}
	}

	mPool.Put(m)

	return fv, kv, nil
}

func (b *baseLogger) writers(ws []io.Writer) {
	cws := make([]io.Writer, 0, len(ws))
	fws := make([]*filterWriter, 0, len(ws))

	ow := ws[:0]
	for _, w := range ws {
		if _, ok := w.(NoColorWriter); !ok {
			cws = append(cws, w)
		}
		if fw, ok := w.(*filterWriter); ok {
			fws = append(fws, fw)
			continue
		}
		ow = append(ow, w)
	}

	b.sync.fw = new(sync.WaitGroup)
	b.sync.fwCnt = len(fws)
	for _, fw := range fws {
		w := b.scan(func(text string) {
			for _, filter := range fw.filters {
				if !filter.Check(text) {
					fw.Write(append([]byte(text), '\n'))
				}
				b.sync.fw.Done()
			}
		})

		ow = append(ow, w)
	}

	b.out.raw = ws

	if len(ws) == 1 {
		b.out.w = ws[0]
		if _, ok := ws[0].(NoColorWriter); !ok {
			b.out.cw = ws[0]
		}
		return
	}

	b.out.w = io.MultiWriter(ow...)
	b.out.cw = io.MultiWriter(cws...)
}

func (b *baseLogger) print(prnt printKind, v []interface{}, settings ...setize) (err error) {
	var ln = lPool.Get().(*line)
	defer func() {

		ln.flags = 0
		ln.depth = 0
		ln.time = nil
		ln.prefixLevel = nil
		ln.format = ""
		ln.kv = ""
		ln.out.dw.w = nil
		ln.err = nil

		lPool.Put(ln)
	}()

	ln.do = prnt
	ln.flags = b.flags
	ln.depth = b.depth
	ln.time = b.time()
	ln.color = b.color
	ln.prefixUser = b.prefix.user
	if ln.out.dw == nil {
		ln.out.dw = &dropCRWriter{}
	}

	ln.out.w = b.out.w
	ln.out.cw = b.out.cw

	for _, s := range settings {
		s.set(ln)
	}

	if ln.v, ln.kv, err = b.filter(v); err != nil {
		return err
	}

	b.sync.ln.Lock()
	defer b.sync.ln.Unlock()

	b.sync.fw.Add(b.sync.fwCnt) // wait for filters... otherwise a race condition
	defer b.sync.fw.Wait()

	return ln.write()
}
