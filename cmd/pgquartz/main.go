package main

import (
	"github.com/mannemsolutions/PgQuartz/internal"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
)

func main() {
	initLogger()
	if config, err := internal.NewConfig(); err != nil {
		log.Fatal(err)
	} else {
		enableDebug(config.Debug)
		config.Initialize()
		h := jobs.NewHandler(config)
		h.VerifyConfig()
		h.RunSteps()
		h.RunChecks()
	}
}
