package main

import (
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"

	"github.com/mannemsolutions/PgQuartz/internal"
)

func main() {
	initLogger()
	if config, err := internal.NewConfig(); err != nil {
		log.Fatal(err)
	} else {
		h := jobs.NewHandler(config)
		h.Run()
	}
}
