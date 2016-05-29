package esp8266server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Module defines a ESP8266 Module
type Module struct {
	conn     net.Conn
	Name     string             `json:"name"`
	Target   string             `json:"target"`
	Commands map[string]Command `json:"commands"`
	Active   bool               `json:"active"`
	*sync.RWMutex
}

// Command is a command that either sends a value with an optional delay
// Or defines a list of sub commands
type Command struct {
	Value       uint8     `json:"value"`
	Delay       uint64    `json:"delay"`
	SubCommands []Command `json:"commands"`
}

// Connects to a module
func (m *Module) connect() (err error) {
	m.conn, err = net.DialTimeout("tcp", fmt.Sprintf(moduleConnectionStr, m.Target), time.Second*30)
	return
}

func (m *Module) isClosed() bool {
	if m.conn == nil {
		return true
	}
	buf := []byte{0x00}
	m.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	_, err := m.conn.Read(buf)
	return err == io.EOF
}

// check heartbeat of module.
func checkHeartbeat(m *Module) {
	var hasLogged bool
	for {
		//Connection is closed
		if m.isClosed() {
			m.Active = false
			if err := m.connect(); err != nil {
				if !hasLogged {
					hasLogged = true
					log.Printf("No connection to module %s at %s", m.Name, m.Target)
				}
			} else {
				log.Printf("Successfully reconnected to module %s at %s", m.Name, m.Target)
				m.Active = true
				hasLogged = false
			}
		}
	}
}

//SendMessage Sends a message to the module on an open connection
func (m *Module) SendMessage(msg uint8) error {
	if m.conn == nil {
		return errors.New("Connection is unavailable")
	}
	m.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	log.Printf("Sending '%d' to '%s'", msg, m.conn.RemoteAddr())
	_, err := m.conn.Write([]byte(fmt.Sprintf("%d", msg)))
	return err
}

//Rest handler for getting a module's commands and status
func getModule(c *gin.Context) {
	name := c.Param("name")
	m, exists := ModuleMap[name]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "unknown module",
		})
		return
	}

	c.JSON(http.StatusOK, m)
}

// Performs a command and its subroutines
func (m *Module) doCommand(c Command) error {
	if c.SubCommands == nil || len(c.SubCommands) == 0 {
		time.Sleep(time.Duration(c.Delay) * time.Millisecond)
		return m.SendMessage(c.Value)
	}
	for _, sub := range c.SubCommands {
		if err := m.doCommand(sub); err != nil {
			return err
		}
	}
	return nil
}

//Rest handler for performing a command by name
func performCommand(c *gin.Context) {
	name := c.Param("name")
	command := c.Param("command")
	m, exists := ModuleMap[name]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "unknown module",
		})
		return
	}

	cmd, exists := m.Commands[command]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "unknown command",
		})
		return
	}

	go func() {
		m.Lock()
		if err := m.doCommand(cmd); err != nil {
			log.Printf("Error sending command: %+v", err)
		}
		m.Unlock()
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Sent command",
	})
	return
}
