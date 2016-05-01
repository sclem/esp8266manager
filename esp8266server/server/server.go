package main

import (
	"flag"

	"github.com/sclem/esp8266manager/esp8266server"
)

func main() {
	flag.Parse()
	esp8266server.RunServer()
}
