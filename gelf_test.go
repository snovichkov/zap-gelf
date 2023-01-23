package gelf_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	gelf "github.com/snovichkov/zap-gelf"
)

func TestAddr(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Addr("127.0.0.1:80"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestHost(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Host("google.com"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestVersion(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Version("1.2"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestMessageKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.MessageKey("custom_message"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestLevelKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LevelKey("custom_level"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestTimeKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.TimeKey("custom_time"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestNameKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.NameKey("custom_name"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestCallerKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.CallerKey("custom_caller"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestFunctionKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.FunctionKey("custom_function"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestStacktraceKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.StacktraceKey("custom_stacktrace"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestSkipLineEnding(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.SkipLineEnding(true),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestLineEnding(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LineEnding("\r\n"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestEncodeDuration(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeDuration(zapcore.NanosDurationEncoder),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestEncodeCaller(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeCaller(zapcore.FullCallerEncoder),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestEncodeName(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeName(zapcore.FullNameEncoder),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestNewReflectedEncoder(t *testing.T) {
	var newEncoder = func(writer io.Writer) zapcore.ReflectedEncoder {
		return json.NewEncoder(writer)
	}
	var core, err = gelf.NewCore(
		gelf.NewReflectedEncoder(newEncoder),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestLevel(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Level(zap.ErrorLevel),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	assert.True(t, core.Enabled(zap.ErrorLevel))
	assert.False(t, core.Enabled(zap.WarnLevel))
}

func TestLevelString(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LevelString("error"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	assert.True(t, core.Enabled(zap.ErrorLevel))
	assert.False(t, core.Enabled(zap.WarnLevel))
}

func TestDynamicLevel(t *testing.T) {
	dynamicLevel := zap.NewAtomicLevel()
	dynamicLevel.SetLevel(zap.ErrorLevel)

	var core, err = gelf.NewCore(
		gelf.DynamicLevel(dynamicLevel),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	assert.True(t, core.Enabled(zap.ErrorLevel))
	assert.False(t, core.Enabled(zap.WarnLevel))

	dynamicLevel.SetLevel(zap.WarnLevel)

	assert.True(t, core.Enabled(zap.ErrorLevel))
	assert.True(t, core.Enabled(zap.WarnLevel))
}

func TestChunkSize(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.ChunkSize(2000),
	)
	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")

	core, err = gelf.NewCore(
		gelf.ChunkSize(gelf.MaxChunkSize + 1),
	)
	assert.Equal(t, gelf.ErrChunkTooLarge, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")

	core, err = gelf.NewCore(
		gelf.ChunkSize(gelf.MinChunkSize - 1),
	)
	assert.Equal(t, gelf.ErrChunkTooSmall, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")
}

func TestCompressionType(t *testing.T) {
	var (
		err              error
		core             zapcore.Core
		compressionTypes = []int{
			gelf.CompressionNone,
			gelf.CompressionGzip,
			gelf.CompressionZlib,
		}
	)

	for _, compressionType := range compressionTypes {
		core, err = gelf.NewCore(
			gelf.CompressionType(compressionType),
		)
		assert.Nil(t, err, "Unexpected error")
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}

	core, err = gelf.NewCore(
		gelf.CompressionType(13),
	)
	assert.Equal(t, gelf.ErrUnknownCompressionType, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")
}

func TestCompressionLevel(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.CompressionLevel(9),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}
