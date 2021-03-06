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
	"sync"
	"syscall"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

const (
	moduleConnectionStr = "%s:9999"
)

var (
	port       string
	modulePath string
	webroot    string
	debug      *bool
	// ModuleMap is the map of connections to ESP8266 modules
	ModuleMap = make(map[string]*Module, 0)
)

func init() {
	debug = flag.Bool("debug", false, "Debug logging flag")
	flag.StringVar(&port, "port", "1955", "Port to run server on")
	flag.StringVar(&modulePath, "modulePath", "../setup/modules.json", "Path to module directory")
	flag.StringVar(&webroot, "webroot", "../esp8266web/www", "Path to web root")
}

// RunServer reads the configs, inits connections, and runs the server
func RunServer() {
	modules, err := readConfigs()
	if err != nil {
		log.Fatalf("Unable to parse module configs. %v", err)
	}

	for _, m := range modules {
		go func(m *Module) {
			m.RWMutex = &sync.RWMutex{}
			if err := m.connect(); err != nil {
				m.Active = false
				log.Printf("Could not open connection to module '%s' at '%s', skipping.", m.Name, m.Target)
			} else {
				m.Active = true
				log.Printf("Found module '%s' at '%s'", m.Name, m.Target)
			}
			ModuleMap[m.Name] = m
			go checkHeartbeat(m)
		}(m)
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

	if !strings.Contains(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}

	log.Printf("Starting esp8266 server on port %s", port)
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
	data, err := ioutil.ReadFile(modulePath)
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
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(static.Serve("/", static.LocalFile(webroot, true)))

	m := r.Group("/modules")
	{
		m.GET("", getModuleList)
		m.GET("/:name", getModule)
		m.GET("/:name/:command", performCommand)
	}

	return r
}

//Get the list of modules available
func getModuleList(c *gin.Context) {
	modules := make([]string, 0)
	for key := range ModuleMap {
		modules = append(modules, key)
	}

	c.JSON(http.StatusOK, modules)
}
