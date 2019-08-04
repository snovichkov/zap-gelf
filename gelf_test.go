package gelf_test

import (
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

func TestLevel(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Level(zap.DebugLevel),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
}

func TestLevelString(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LevelString("debug"),
	)

	assert.Nil(t, err, "Unexpected error")
	assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
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
