package main

import (
	"flag"

	"github.com/sclem/garagedoor/esp8266server"
)

func main() {
	flag.Parse()
	esp8266server.RunServer()
}
