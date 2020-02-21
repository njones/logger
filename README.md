[![Build Status](https://travis-ci.org/njones/logger.svg?branch=master)](https://travis-ci.org/njones/logger) [![GoDoc](https://godoc.org/github.com/njones/logger?status.svg)](https://godoc.org/github.com/njones/logger) [![Go Report Card](https://goreportcard.com/badge/github.com/njones/logger)](https://goreportcard.com/report/github.com/njones/logger)

# Logger

Logger is a go library that provides a full featured logger. It has structured logging and color capabilities.

See the [GoDoc](https://godoc.org/github.com/njones/logger) for more information.

## Installation

    go get github.com/njones/logger

## Example of how to use the Logger

```go
package main

import (
	"github.com/njones/logger"
	"github.com/njones/logger/color"
)

var log = logger.New()

func main() {
	x := "xyz"
	user, emails := "someone", []string{"somewhere", "overtherainbow"}

	log.Debug("Starting the main function ...")

	// does something
	log.Warnf("Doing %s ...", x)

	// uses a custom color
	log.With(WithColor(color.Blue)).Info("This is info, but displays in BLUE ...")
	log.Info("instead of info being GREEN.")

	// uses structured logging
	log.Debug("The thing occurred here.", logger.KV("user", user), logger.KV("email", emails))
	logf := log.Field("user", user)
	logf.Debug("The thing occurred here.", logger.KV("email", emails)) // mix and match

	// can supress
	log.Supress(logger.Trace)
	log.Error("Finished with main.") // will show up
	log.Trace("Finished with main.") // won't show up

	// conditional log, will only log if there is an error
	err := doSomething()
	log.OnErr(err).Fatal("will log and exit on error: %v", logger.OnErr{}) // logger.OnErr{} is a placeholder for err 
	// log.Panic("will panic") will panic
}
```

# License

Logger is available under the [MIT License](https://opensource.org/licenses/MIT).

Copyright (c) 2017-2020 Nika Jones <copyright@nikajon.es> All Rights Reserved.