package sshutil

import (
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/crypto/ssh"
)

// SSHConnIdentifier uniquely identifies an SSH connection.
type SSHConnIdentifier struct {
	Host string
	Port int
	User string
}

// NewSSHConnIdentifier creates a new SSHConnIdentifier.
func NewSSHConnIdentifier(host string, port int, user string) SSHConnIdentifier {
	return SSHConnIdentifier{
		Host: host,
		Port: port,
		User: user,
	}
}

// String returns a string representation of SSHConnIdentifier.
func (id SSHConnIdentifier) String() string {
	return fmt.Sprintf("%s:%d@%s", id.User, id.Port, id.Host)
}

// SSHConnection represents an SSH connection.
type SSHConnection struct {
	Client *ssh.Client
	Host   string
}

// SSHConnectionPool manages a pool of SSH connections.
type SSHConnectionPool struct {
	connections map[string]*SSHConnection
	mutex       sync.Mutex
}

var (
	instance *SSHConnectionPool
	once     sync.Once
)

// GetSSHConnectionPool returns the singleton instance of SSHConnectionPool.
func GetSSHConnectionPool() *SSHConnectionPool {
	once.Do(func() {
		instance = &SSHConnectionPool{
			connections: make(map[string]*SSHConnection),
		}
	})
	return instance
}

// GetConnection retrieves or creates an SSH connection.
func (pool *SSHConnectionPool) GetConnection(id SSHConnIdentifier, config *ssh.ClientConfig) (*SSHConnection, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	idStr := id.String()
	conn, ok := pool.connections[idStr]

	// Check if the existing connection is active
	if ok && conn.Client != nil {
		_, _, err := conn.Client.SendRequest("keepalive@golang.org", true, nil)
		if err == nil {
			slog.Info("SSHConnectionPool: Reusing connection", "id", idStr)
			return conn, nil // The connection is still active
		}
		// The connection is not active, close the old connection
		conn.Client.Close()
	}

	slog.Info("SSHConnectionPool: Creating new connection", "id", idStr)
	// Establish a new connection
	newConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", id.Host, id.Port), config)
	if err != nil {
		return nil, err
	}

	sshConn := &SSHConnection{
		Client: newConn,
		Host:   id.Host,
	}

	pool.connections[idStr] = sshConn
	return sshConn, nil
}
