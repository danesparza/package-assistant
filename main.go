/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/danesparza/package-assistant/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"strings"
	"time"
)

func main() {
	//	Set log info:
	log.Logger = log.With().Caller().Logger()

	//	Set log level (default to info)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	switch strings.ToLower(os.Getenv("LOGGER_LEVEL")) {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		break
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		break
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		break
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		break
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
		break
	}

	//	Set the error stack marshaller
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	//	Set log time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	cmd.Execute()
}
