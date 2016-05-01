package esp8266server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Module defines a ESP8266 Module
type Module struct {
	conn     net.Conn
	Name     string            `json:"name"`
	Target   string            `json:"target"`
	Commands map[string]uint32 `json:"commands"`
	Active   bool              `json:"active"`
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
func (m *Module) SendMessage(msg uint32) error {
	if m.conn == nil {
		return errors.New("Connection is unavailable")
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, msg)
	m.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	_, err := m.conn.Write(buf)
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

	if err := m.SendMessage(cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Unable to send command",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully sent command",
	})
	return
}
