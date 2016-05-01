package esp8266server

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
)

const (
	moduleConnectionStr = "%s:9999"
)

var (
	port string
	// ModuleMap is the map of connections to ESP8266 modules
	ModuleMap = make(map[string]*Module, 0)
)

func init() {
	flag.StringVar(&port, "port", "1955", "Port to run server on")
	if !strings.Contains(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}
}

// RunServer reads the configs, inits connections, and runs the server
func RunServer() {
	modules, err := readConfigs()
	if err != nil {
		log.Fatalf("Unable to parse module configs. %v", err)
	}

	for _, m := range modules {
		m.connect()
		if err != nil {
			m.Active = false
			log.Printf("Could not open connection to module '%s' at '%s', skipping.", m.Name, m.Target)
		} else {
			m.Active = true
			log.Printf("Loading module '%s' at '%s'", m.Name, m.Target)
		}
		ModuleMap[m.Name] = m
		go checkHeartbeat(m)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// called on program exit
	go func(sigChan chan os.Signal) {
		for range sigChan {
			shutdown()
		}
	}(ch)

	r := getServer()
	r.Run(port)
}

//Shut down all active connections and exit
func shutdown() {
	for _, m := range ModuleMap {
		if m.conn != nil {
			m.conn.Close()
		}
	}
	os.Exit(0)
}

//Read config file and open connections at addrs
func readConfigs() ([]*Module, error) {
	data, err := ioutil.ReadFile("modules.json")
	if err != nil {
		return nil, err
	}
	var modules = make([]*Module, 0)
	if err := json.Unmarshal(data, &modules); err != nil {
		return nil, err
	}

	log.Printf("Loaded with %d modules", len(modules))
	return modules, nil
}

func getServer() *gin.Engine {
	//	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("", getModuleList)
	r.GET("/:name", getModule)
	r.GET("/:name/:command", performCommand)

	return r
}

//Get the list of garages available
func getModuleList(c *gin.Context) {
	modules := make([]string, 0)
	for key := range ModuleMap {
		modules = append(modules, key)
	}

	c.JSON(http.StatusOK, modules)
}
