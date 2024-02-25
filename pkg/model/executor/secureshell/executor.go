package secureshell

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	IP      string
	Port    int
	User    string
	KeyPath string
	client  *ssh.Client
}

func NewSSHExecutor(ip string, port int, user string, keyPath string) *SSHExecutor {
	return &SSHExecutor{
		IP:      ip,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
		client:  nil,
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

func (s *SSHExecutor) IsConnected() bool {
	return s.client != nil
}

func (s *SSHExecutor) Connect() error {
	if s.IsConnected() {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", s.IP, s.Port)
	key, err := os.ReadFile(s.KeyPath)
	if err != nil {
		slog.Error("error occured when reading key", slog.String("error", err.Error()))
		return err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		slog.Error("error occured when parsing key", slog.String("error", err.Error()))
		return err
	}
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	slog.Info("connecting", slog.String("addr", addr), slog.String("user", s.User), slog.String("key", s.KeyPath))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		slog.Error("error occured when dialing", slog.String("error", err.Error()))
		return err
	}
	slog.Info("connected", slog.String("addr", addr), slog.String("user", s.User), slog.String("key", s.KeyPath))
	s.client = client
	return nil
}

func (s *SSHExecutor) InstallDocker() error {
	err := s.UploadFile("assets/scripts/ubuntu_setup_docker.sh", "/tmp/ubuntu_setup_docker.sh")
	if err != nil {
		return err
	}
	_, _, err = s.RunCommand("bash -x /tmp/ubuntu_setup_docker.sh", true)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHExecutor) RunCommand(cmd string, debug bool) (string, string, error) {
	if debug {
		slog.Info("running command", slog.String("command", cmd))
	}
	session, err := s.client.NewSession()
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

	copyOutput := func(r *bufio.Reader, buffer *bytes.Buffer) {
		for {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				if debug {
					fmt.Println(line)
				}
				buffer.WriteString(line)
			}
			if err := scanner.Err(); err != nil {
				log.Printf("Error reading output: %v", err)
			}
		}
	}

	go copyOutput(bufio.NewReader(stdoutPipe), stdoutBuffer)
	go copyOutput(bufio.NewReader(stderrPipe), stderrBuffer)

	if err := session.Wait(); err != nil {
		return "", "", err
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}

func (s *SSHExecutor) UploadFile(localFilePath, remoteFilePath string) error {
	sftpClient, err := sftp.NewClient(s.client)
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
	sftpClient, err := sftp.NewClient(s.client)
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
	return s.client.Close()
}
