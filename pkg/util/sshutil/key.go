package sshutil

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func GenerateSSHKeyPair(bits int) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicSSHKey := ssh.MarshalAuthorizedKey(publicKey)
	return string(privatePEM), string(publicSSHKey), nil
}

func CreateSSHKeyPair(folder, name string, numBits int) (string, string, error) {
	if strings.Contains(name, ".") {
		return "", "", fmt.Errorf("name should not contain '.'")
	}
	if strings.Contains(name, "/") {
		return "", "", fmt.Errorf("name should not contain '/'")
	}
	privkey, pubkey, err := GenerateSSHKeyPair(numBits)
	if err != nil {
		return "", "", err
	}
	slog.Info("ssh key pair generated", slog.Int("num_bits", numBits), slog.String("public_key", pubkey))
	fingerprint, err := GetSSHPublicKeyFingerprintMD5(pubkey)
	if err != nil {
		return "", "", err
	}
	slog.Info("public key fingerprint", slog.String("fingerprint", fingerprint))
	os.MkdirAll(folder, 0755)
	err = os.WriteFile(filepath.Join(
		folder, name,
	), []byte(privkey), 0600)
	if err != nil {
		return "", "", err
	}
	slog.Info("private key saved", slog.String("path", filepath.Join(folder, name)))
	err = os.WriteFile(filepath.Join(
		folder, fmt.Sprintf("%s.pub", name),
	), []byte(pubkey), 0644)
	if err != nil {
		return "", "", err
	}
	slog.Info("public key saved", slog.String("path", filepath.Join(folder, fmt.Sprintf("%s.pub", name))))
	return privkey, pubkey, nil
}

func LoadSSHKeyPair(folder, name string) (string, string, error) {
	if strings.Contains(name, ".") {
		return "", "", fmt.Errorf("name should not contain '.'")
	}
	if strings.Contains(name, "/") {
		return "", "", fmt.Errorf("name should not contain '/'")
	}
	pubkey, err := os.ReadFile(filepath.Join(folder, fmt.Sprintf("%s.pub", name)))
	if err != nil {
		return "", "", err
	}
	privkey, err := os.ReadFile(filepath.Join(folder, name))
	if err != nil {
		return "", "", err
	}
	return string(pubkey), string(privkey), nil
}

func LoadOrCreateSSHKeyPair(folder, name string) (string, string, error) {
	privkey, pubkey, err := LoadSSHKeyPair(folder, name)
	if err == nil {
		return privkey, pubkey, nil
	}
	return CreateSSHKeyPair(folder, name, 4096)
}

func GetSSHPublicKeyFingerprintMD5(publicKey string) (string, error) {
	sshPubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", err
	}
	hash := md5.Sum(sshPubKey.Marshal())
	var fingerprint strings.Builder
	for i, b := range hash {
		if i > 0 {
			fingerprint.WriteString(":")
		}
		fmt.Fprintf(&fingerprint, "%02x", b)
	}
	return fingerprint.String(), nil
}
