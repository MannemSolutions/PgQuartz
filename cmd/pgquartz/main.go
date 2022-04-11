package main

import (
	"fmt"
	//"github.com/mannemsolutions/PgQuartz/pkg/jobrunner"
	"log"

	"github.com/mannemsolutions/PgQuartz/internal"
)

func main() {
	if config, err := internal.NewConfig(); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(config.String())
	}
}
