package main

import "github.com/flowmatters/openwater-core/util/logger"

func setupLogger() {
	logger.SetupLogger(quiet, verbose)
}
