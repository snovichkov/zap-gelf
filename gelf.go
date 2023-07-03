package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MinChunkSize minimal chunk size in bytes.
	MinChunkSize = 512

	// MaxChunkSize maximal chunk size in bytes.
	// See http://docs.graylog.org/en/2.4/pages/gelf.html.
	MaxChunkSize = 8192

	// MaxChunkCount maximal chunk per message count.
	// See http://docs.graylog.org/en/2.4/pages/gelf.html.
	MaxChunkCount = 128

	// DefaultChunkSize is default WAN chunk size.
	DefaultChunkSize = 1420

	// CompressionNone don't use compression.
	CompressionNone = 0

	// CompressionGzip use gzip compression.
	CompressionGzip = 1

	// CompressionZlib use zlib compression.
	CompressionZlib = 2
)

type (
	// Option interface.
	Option interface {
		apply(conf *optionConf) error
	}

	// coreConf core.
	optionConf struct {
		addr             string
		host             string
		version          string
		enabler          zap.AtomicLevel
		encoder          zapcore.EncoderConfig
		chunkSize        int
		writeSyncers     []zapcore.WriteSyncer
		compressionType  int
		compressionLevel int
	}

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(conf *optionConf) error

	// implement io.Writer
	writer struct {
		conn             net.Conn
		chunkSize        int
		chunkDataSize    int
		compressionType  int
		compressionLevel int
		writeCloserPool  *sync.Pool
	}

	writeCloserResetter interface {
		io.WriteCloser
		Reset(w io.Writer)
	}

	// implement io.WriteCloser.
	writeCloser struct {
		err    error
		buffer io.Writer
	}

	// implement zapcore.Core.
	wrappedCore struct {
		core zapcore.Core
	}
)

var (
	// ErrChunkTooSmall triggered when chunk size to small.
	ErrChunkTooSmall = errors.New("chunk size too small")

	// ErrChunkTooLarge triggered when chunk size too large.
	ErrChunkTooLarge = errors.New("chunk size too large")

	// ErrUnknownCompressionType triggered when passed invalid compression type.
	ErrUnknownCompressionType = errors.New("unknown compression type")

	// chunkedMagicBytes chunked message magic bytes.
	// See http://docs.graylog.org/en/2.4/pages/gelf.html.
	chunkedMagicBytes = []byte{0x1e, 0x0f}

	// Ensure *writer implements zapcore.WriteSyncer.
	_ zapcore.WriteSyncer = (*writer)(nil)
)

// NewCore zap core constructor.
func NewCore(options ...Option) (_ zapcore.Core, err error) {
	var conf = optionConf{
		addr: "127.0.0.1:12201",
		host: "localhost",
		encoder: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			NameKey:        "_logger",
			LevelKey:       "level",
			CallerKey:      "_caller",
			MessageKey:     "short_message",
			StacktraceKey:  "full_message",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeName:     zapcore.FullNameEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeLevel:    levelEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
		version:          "1.1",
		enabler:          zap.NewAtomicLevel(),
		chunkSize:        DefaultChunkSize,
		writeSyncers:     make([]zapcore.WriteSyncer, 0, 8),
		compressionType:  CompressionGzip,
		compressionLevel: gzip.BestCompression,
	}

	for _, option := range options {
		if err = option.apply(&conf); err != nil {
			return nil, err
		}
	}

	var w = &writer{
		chunkSize:        conf.chunkSize,
		chunkDataSize:    conf.chunkSize - 12, // chunk size - chunk header size
		compressionType:  conf.compressionType,
		compressionLevel: conf.compressionLevel,
	}

	w.writeCloserPool = &sync.Pool{
		New: w.newWriteCloser,
	}

	if w.conn, err = net.Dial("udp", conf.addr); err != nil {
		return nil, err
	}

	var ws zapcore.WriteSyncer = w
	if len(conf.writeSyncers) > 0 {
		var writers = append([]zapcore.WriteSyncer{w}, conf.writeSyncers...)
		ws = zapcore.NewMultiWriteSyncer(writers...)
	}

	var core = zapcore.NewCore(
		zapcore.NewJSONEncoder(conf.encoder),
		ws,
		conf.enabler,
	)

	return &wrappedCore{
		core: core.With([]zapcore.Field{
			zap.String("host", conf.host),
			zap.String("version", conf.version),
		}),
	}, nil
}

// Addr set GELF address.
func Addr(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.addr = value
		return nil
	})
}

// Host set GELF host.
func Host(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.host = value
		return nil
	})
}

// Version set GELF version.
func Version(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.version = value
		return nil
	})
}

// MessageKey set zapcore.EncoderConfig MessageKey property.
func MessageKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.MessageKey = escapeKey(value)
		return nil
	})
}

// LevelKey set zapcore.EncoderConfig LevelKey property.
func LevelKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.LevelKey = escapeKey(value)
		return nil
	})
}

// TimeKey set zapcore.EncoderConfig TimeKey property.
func TimeKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.TimeKey = escapeKey(value)
		return nil
	})
}

// LevelAtomic set atomic logging level which can be changed dynamically.
func LevelAtomic(level zap.AtomicLevel) Option {
	return optionFunc(func(conf *optionConf) (err error) {
		conf.enabler = level
		return nil
	})
}

// NameKey set zapcore.EncoderConfig NameKey property.
func NameKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.NameKey = escapeKey(value)
		return nil
	})
}

// CallerKey set zapcore.EncoderConfig CallerKey property.
func CallerKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.CallerKey = escapeKey(value)
		return nil
	})
}

// FunctionKey set zapcore.EncoderConfig FunctionKey property.
func FunctionKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.FunctionKey = escapeKey(value)
		return nil
	})
}

// StacktraceKey set zapcore.EncoderConfig StacktraceKey property.
func StacktraceKey(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.StacktraceKey = escapeKey(value)
		return nil
	})
}

// SkipLineEnding set zapcore.EncoderConfig SkipLineEnding property.
func SkipLineEnding(value bool) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.SkipLineEnding = value
		return nil
	})
}

// LineEnding set zapcore.EncoderConfig LineEnding property.
func LineEnding(value string) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.LineEnding = value
		return nil
	})
}

// EncodeDuration set zapcore.EncoderConfig EncodeDuration property.
func EncodeDuration(value zapcore.DurationEncoder) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.EncodeDuration = value
		return nil
	})
}

// EncodeCaller set zapcore.EncoderConfig EncodeCaller property.
func EncodeCaller(value zapcore.CallerEncoder) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.EncodeCaller = value
		return nil
	})
}

// EncodeName set zapcore.EncoderConfig EncodeName property.
func EncodeName(value zapcore.NameEncoder) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.EncodeName = value
		return nil
	})
}

// WriteSyncers sets additional zapcore.WriteSyncers on the core.
func WriteSyncers(value ...zapcore.WriteSyncer) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.writeSyncers = append(conf.writeSyncers, value...)
		return nil
	})
}

// NewReflectedEncoder set zapcore.EncoderConfig NewReflectedEncoder property.
func NewReflectedEncoder(value func(io.Writer) zapcore.ReflectedEncoder) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.encoder.NewReflectedEncoder = value
		return nil
	})
}

// Level set logging level.
func Level(value zapcore.Level) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.enabler.SetLevel(value)
		return nil
	})
}

// LevelString set logging level.
func LevelString(value string) Option {
	return optionFunc(func(conf *optionConf) (err error) {
		err = conf.enabler.UnmarshalText([]byte(value))
		return err
	})
}

// ChunkSize set GELF chunk size.
func ChunkSize(value int) Option {
	return optionFunc(func(conf *optionConf) error {
		if value < MinChunkSize {
			return ErrChunkTooSmall
		}

		if value > MaxChunkSize {
			return ErrChunkTooLarge
		}

		conf.chunkSize = value

		return nil
	})
}

// CompressionType set GELF compression type.
func CompressionType(value int) Option {
	return optionFunc(func(conf *optionConf) error {
		switch value {
		case CompressionNone, CompressionGzip, CompressionZlib:
		default:
			return ErrUnknownCompressionType
		}

		conf.compressionType = value

		return nil
	})
}

// CompressionLevel set GELF compression level.
func CompressionLevel(value int) Option {
	return optionFunc(func(conf *optionConf) error {
		conf.compressionLevel = value
		return nil
	})
}

// Write implements io.Writer.
func (w *writer) Write(buf []byte) (n int, err error) {
	var (
		cw   writeCloserResetter
		cBuf bytes.Buffer
	)

	cw = w.writeCloserPool.Get().(writeCloserResetter)

	cw.Reset(&cBuf)

	if n, err = cw.Write(buf); err != nil {
		return n, err
	}

	if cw.Close() == nil {
		w.writeCloserPool.Put(cw)
	}
	cw = nil

	var cBytes = cBuf.Bytes()
	if count := w.chunkCount(cBytes); count > 1 {
		return w.writeChunked(count, cBytes)
	}

	if n, err = w.conn.Write(cBytes); err != nil {
		return n, err
	}

	if n != len(cBytes) {
		return n, fmt.Errorf("writed %d bytes but should %d bytes", n, len(cBytes))
	}

	return n, nil
}

// Sync is a no-op, but required to implement the zapcore.WriteSyncer interface.
func (w *writer) Sync() error {
	return nil
}

func (w *writer) newWriteCloser() (cw interface{}) {
	var err error
	switch w.compressionType {
	case CompressionNone:
		cw = &writeCloser{nil, nil}
	case CompressionGzip:
		cw, err = gzip.NewWriterLevel(nil, w.compressionLevel)
	case CompressionZlib:
		cw, err = zlib.NewWriterLevel(nil, w.compressionLevel)
	}
	if err != nil {
		cw = &writeCloser{err, nil}
	}
	return cw
}

// Close implementation of io.WriteCloser.
func (*writeCloser) Close() error {
	return nil
}

func (wc *writeCloser) Reset(buffer io.Writer) {
	wc.buffer = buffer
}

func (wc *writeCloser) Write(p []byte) (n int, err error) {
	if wc.err != nil {
		return 0, err
	}
	return wc.buffer.Write(p)
}

// Enabled implementation of zapcore.Core.
func (w *wrappedCore) Enabled(l zapcore.Level) bool {
	return w.core.Enabled(l)
}

// With implementation of zapcore.Core.
func (w *wrappedCore) With(fields []zapcore.Field) zapcore.Core {
	return &wrappedCore{core: w.core.With(w.escape(fields))}
}

// Check implementation of zapcore.Core.
func (w *wrappedCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if w.Enabled(e.Level) {
		return ce.AddCore(e, w)
	}

	return ce
}

// Write implementation of zapcore.Core.
func (w *wrappedCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	return w.core.Write(e, w.escape(fields))
}

// Sync implementation of zapcore.Core.
func (w *wrappedCore) Sync() error {
	return w.core.Sync()
}

// apply implements Option.
func (f optionFunc) apply(conf *optionConf) error {
	return f(conf)
}

// escape prefixed additional gelf fields.
func (w *wrappedCore) escape(fields []zapcore.Field) []zapcore.Field {
	if len(fields) == 0 {
		return fields
	}

	var escaped = make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		field.Key = escapeKey(field.Key)
		escaped = append(escaped, field)
	}

	return escaped
}

// escapeKey append prefix to additional field keys.
func escapeKey(value string) string {
	switch value {
	case "id":
		return "__id"
	case "version", "host", "short_message", "full_message", "timestamp", "level":
		return value
	}

	if len(value) == 0 {
		return "_"
	}

	if value[0] == '_' {
		return value
	}

	return "_" + value
}

// levelEncoder maps the zap log levels to the gelf levels.
// See http://docs.graylog.org/en/2.4/pages/gelf.html.
func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendInt(7)
	case zapcore.InfoLevel:
		enc.AppendInt(6)
	case zapcore.WarnLevel:
		enc.AppendInt(4)
	case zapcore.ErrorLevel:
		enc.AppendInt(3)
	case zapcore.DPanicLevel:
		enc.AppendInt(0)
	case zapcore.PanicLevel:
		enc.AppendInt(0)
	case zapcore.FatalLevel:
		enc.AppendInt(0)
	}
}

// chunkCount calculate the number of GELF chunks.
func (w *writer) chunkCount(b []byte) int {
	lenB := len(b)
	if lenB <= w.chunkSize {
		return 1
	}

	return len(b)/w.chunkDataSize + 1
}

// writeChunked send message by chunks.
func (w *writer) writeChunked(count int, cBytes []byte) (n int, err error) {
	if count > MaxChunkCount {
		return 0, fmt.Errorf("need %d chunks but shold be later or equal to %d", count, MaxChunkCount)
	}

	var (
		cBuf = bytes.NewBuffer(
			make([]byte, 0, w.chunkSize),
		)
		nChunks   = uint8(count)
		messageID = make([]byte, 8)
	)

	if n, err = io.ReadFull(rand.Reader, messageID); err != nil || n != 8 {
		return 0, fmt.Errorf("rand.Reader: %d/%s", n, err)
	}

	var (
		off       int
		chunkLen  int
		bytesLeft = len(cBytes)
	)

	for i := uint8(0); i < nChunks; i++ {
		off = int(i) * w.chunkDataSize
		chunkLen = w.chunkDataSize
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}

		cBuf.Reset()
		cBuf.Write(chunkedMagicBytes)
		cBuf.Write(messageID)
		cBuf.WriteByte(i)
		cBuf.WriteByte(nChunks)
		cBuf.Write(cBytes[off : off+chunkLen])

		if n, err = w.conn.Write(cBuf.Bytes()); err != nil {
			return len(cBytes) - bytesLeft + n, err
		}

		if n != len(cBuf.Bytes()) {
			n = len(cBytes) - bytesLeft + n
			return n, fmt.Errorf("writed %d bytes but should %d bytes", n, len(cBytes))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return len(cBytes) - bytesLeft, fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}

	return len(cBytes), nil
}
