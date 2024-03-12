package secureshell

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/sshutil"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	IP         string
	Port       int
	User       string
	KeyPath    string
	connection *sshutil.SSHConnection
}

func NewSSHExecutor(ip string, port int, user string, keyPath string) *SSHExecutor {
	return &SSHExecutor{
		IP:      ip,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
	}
}

func (s *SSHExecutor) String() string {
	return fmt.Sprintf(
		"SSHExecutor{IP: %s, Port: %d, User: %s, KeyPath: %s}",
		s.IP,
		s.Port,
		s.User,
		s.KeyPath,
	)
}

func (s *SSHExecutor) GetConfig() (*ssh.ClientConfig, error) {
	key, err := os.ReadFile(s.KeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

// Connect now uses the SSHConnectionPool to get the connection
func (s *SSHExecutor) Connect() error {
	// Using the connection pool to get the connection
	pool := sshutil.GetSSHConnectionPool()

	// Creating SSHConnIdentifier
	identifier := sshutil.SSHConnIdentifier{
		Host: s.IP,
		Port: s.Port,
		User: s.User,
	}

	// Getting the config
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	// Getting the connection from the pool
	sshConnection, err := pool.GetConnection(identifier, config)
	if err != nil {
		return err
	}

	// Storing the ssh.Client from the pool
	s.connection = sshConnection
	return nil
}

func (s *SSHExecutor) RunExecutable(path string) error {
	log.Info("running executable", "path", path)
	targetExecutablePath := filepath.Join(
		"/tmp",
		"dode",
		uuid.New().String(),
	)
	s.RunCommand("mkdir -p " + filepath.Dir(targetExecutablePath))
	log.Info("uploading executable", "src", path, "dst", targetExecutablePath)
	err := s.UploadFile(path, targetExecutablePath)
	if err != nil {
		return err
	}
	log.Info("changing permissions", "path", targetExecutablePath)
	_, _, err = s.RunCommand(fmt.Sprintf("chmod +x %s", targetExecutablePath))
	if err != nil {
		return err
	}
	_, _, err = s.RunCommand(targetExecutablePath)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHExecutor) RunCommand(cmd string) (string, string, error) {
	log.Info("running command", "cmd", cmd)
	session, err := s.connection.Client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return "", "", err
	}

	if err := session.Start(cmd); err != nil {
		return "", "", err
	}

	// new a stdoutBuffer to store output
	stdoutBuffer := bytes.NewBuffer(nil)
	stderrBuffer := bytes.NewBuffer(nil)

	wg := &sync.WaitGroup{}

	copyOutput := func(filename string, r *bufio.Reader, buffer *bytes.Buffer) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			log.Info("output", filename, text)
			buffer.WriteString(text)
		}
		if err := scanner.Err(); err != nil {
			log.Error("Error reading output: %v", err)
		}
	}

	wg.Add(2)
	go copyOutput("stdout", bufio.NewReader(stdoutPipe), stdoutBuffer)
	go copyOutput("stderr", bufio.NewReader(stderrPipe), stderrBuffer)

	if err := session.Wait(); err != nil {
		return "", "", err
	}

	wg.Wait()

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}

func (s *SSHExecutor) UploadFile(localFilePath, remoteFilePath string) error {
	sftpClient, err := sftp.NewClient(s.connection.Client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	return err
}

func (s *SSHExecutor) DownloadFile(remoteFilePath, localFilePath string) error {
	sftpClient, err := sftp.NewClient(s.connection.Client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	return err
}

func (s *SSHExecutor) Close() error {
	return s.connection.Client.Close()
}
