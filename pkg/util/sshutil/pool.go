package sshutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
)

// SSHConnIdentifier uniquely identifies an SSH connection by its host, port, and user.
type SSHConnIdentifier struct {
	Host string
	Port int
	User string
}

// NewSSHConnIdentifier creates and returns a new SSHConnIdentifier.
func NewSSHConnIdentifier(host string, port int, user string) SSHConnIdentifier {
	return SSHConnIdentifier{
		Host: host,
		Port: port,
		User: user,
	}
}

// String returns a string representation of SSHConnIdentifier in the format user:port@host.
func (id SSHConnIdentifier) String() string {
	return fmt.Sprintf("%s@%s:%d", id.User, id.Host, id.Port)
}

// SSHConnection represents an SSH connection with a client and its host.
type SSHConnection struct {
	Client *ssh.Client
	Host   string
}

// SSHConnectionPool manages a pool of SSH connections using a concurrent-safe map.
type SSHConnectionPool struct {
	connections sync.Map
}

var (
	// instance holds the singleton instance of SSHConnectionPool.
	instance *SSHConnectionPool
	// once ensures that the singleton instance is initialized only once.
	once sync.Once
)

// GetSSHConnectionPool initializes (if not already) and returns the singleton instance of SSHConnectionPool.
func GetSSHConnectionPool() *SSHConnectionPool {
	once.Do(func() {
		instance = &SSHConnectionPool{
			connections: sync.Map{},
		}
	})
	return instance
}

// getConnectionID generates the identifier for a connection and retrieves the connection from the pool, if it exists.
func (pool *SSHConnectionPool) getConnectionID(id SSHConnIdentifier) (string, *SSHConnection, bool) {
	idStr := id.String()
	conn, ok := pool.connections.Load(idStr)
	if ok {
		return idStr, conn.(*SSHConnection), true
	}
	return idStr, nil, false
}

// GetConnection retrieves or creates an SSH connection from the pool.
func (pool *SSHConnectionPool) GetConnection(id SSHConnIdentifier, config *ssh.ClientConfig) (*SSHConnection, error) {
	maxRetries := 128
	retryDelay := time.Second

	idStr, conn, ok := pool.getConnectionID(id)
	if ok && conn.Client != nil {
		_, _, err := conn.Client.SendRequest("keepalive@golang.org", true, nil)
		if err == nil {
			log.Debug("reusing active connection", "id", idStr)
			return conn, nil
		}
		log.Debug("closing inactive connection", "id", idStr)
		conn.Client.Close()
	}

	for i := 0; i < maxRetries; i++ {
		newConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", id.Host, id.Port), config)
		if err != nil {
			log.Error("failed to establish connection", "id", idStr, "error", err)
			time.Sleep(retryDelay)
			continue
		}
		log.Info("connection established", "id", idStr)
		sshConn := &SSHConnection{Client: newConn, Host: id.Host}
		pool.connections.Store(idStr, sshConn)
		return sshConn, nil
	}

	log.Error("failed to establish connection", "id", idStr, "num_retries", maxRetries)
	return nil, fmt.Errorf("failed to establish connection after %d retries", maxRetries)
}
