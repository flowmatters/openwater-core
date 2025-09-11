package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gonum.org/v1/hdf5"
)

func SetupLogger(quiet, verbose bool) {
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.ANSIC})
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.FloatingPointPrecision = 2
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		hdf5.DisplayErrors(true)
	} else if quiet {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		hdf5.DisplayErrors(false)
	}
}
