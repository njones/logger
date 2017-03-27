[![Build Status](https://travis-ci.org/njones/logger.svg?branch=master)](https://travis-ci.org/njones/logger) [![GoDoc](https://godoc.org/github.com/njones/logger?status.svg)](https://godoc.org/github.com/njones/logger)

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
)

var log = logger.New()

func main() {
	x := "xyz"
	user, emails := "someone", []string{"somewhere", "overtherainbow"}

	log.Debug("Starting the main function ...")

	// does something
	log.Warnf("Doing %s ...", x)

	// uses a custom color
	log.Color(logger.Blue).Info("This is info but displays in BLUE ...")
	log.Info("instead of info being GREEN.")

	// uses structured logging
	log.Error("The error occurred here.", logger.KV("user", user), logger.KV("email", emails))

	log.Trace("Finished with main.")
}
```

# License

Logger is available under the [MIT License](https://opensource.org/licenses/MIT).

Copyright (c) 2017 Nika Jones <copyright@nikajon.es> All Rights Reserved.