# Zap GELF [![GitHub license][license-img]][license] [![Go Report Card][report-img]][report] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoDoc][doc-img]][doc]

Zap GELF added availability to zap logger send your logs to Graylog server over UDP. All zap fields will be sent as 
additional fields on Graylog. 

## Installation

```bash
go get -u github.com/snovichkov/zap-gelf
```

## Features

* Use fast zap JSON serializer
* Support chunking over UPD
* Support gzip/zlib compression
    
## Quick Start

```go
package main 

import (
	"os"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/snovichkov/zap-gelf"
)

func main() {
	var (
        err  error
        host string
        core zapcore.Core
    )
    
    if host, err = os.Hostname(); err != nil {
        panic(err)
    }
    
    if core, err = gelf.NewCore(
        gelf.Addr("127.0.0.1:12001"),
        gelf.Host(host),
    ); err != nil {
        panic(err)
    }
    
    var logger = zap.New(
        core,
        zap.AddCaller(),
        zap.AddStacktrace(zap.LevelEnablerFunc(func(l zapcore.Level) bool {
            return core.Enabled(l)
        })),
    )
    defer logger.Sync()
    
    logger.
        With(
            zap.String("with", "field"),
        ).
        Error(
            "An error was accrued",
            zap.String("an_filed", "Hello word!"),
            zap.String("id", "an_id"),
        )
    
    logger.
        Sugar().
        With(
            "an_filed", "Hello word!",
            "id", "an_id",
        ).
        Error("An error was accrued")
}
```

[ci-img]: https://travis-ci.org/snovichkov/zap-gelf.svg?branch=master
[ci]: https://travis-ci.org/snovichkov/zap-gelf
[doc-img]: https://godoc.org/github.com/snovichkov/zap-gelf?status.svg
[doc]: https://godoc.org/github.com/snovichkov/zap-gelf
[cov-img]: https://codecov.io/gh/snovichkov/zap-gelf/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/snovichkov/zap-gelf
[report-img]: https://goreportcard.com/badge/github.com/snovichkov/zap-gelf
[report]: https://goreportcard.com/report/github.com/snovichkov/zap-gelf
[license-img]: https://img.shields.io/github/license/snovichkov/zap-gelf.svg
[license]: https://github.com/snovichkov/zap-gelf/blob/master/LICENSE
