// Package logging just provides a flag for setting the log.level
package logging

import (
	"errors"
	"flag"

	log "github.com/Sirupsen/logrus"
)

var (
	LogLevel  levelFlag
	LogFormat formatFlag
)

type (
	levelFlag  string
	formatFlag string
)

// String implements flag.Value.
func (f levelFlag) String() string {
	return "info"
}

// Set implements flag.Value.
func (f levelFlag) Set(level string) error {
	l, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(l)
	return nil
}

func (f formatFlag) String() string {
	return "text"
}

func (f formatFlag) Set(format string) error {
	switch format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		return errors.New("Unable to parse log format")
	}
	return nil
}

func init() {
	flag.Var(&LogLevel, "log.level", "Set log level")
	flag.Var(&LogFormat, "log.format", "Set log formatter")
}
