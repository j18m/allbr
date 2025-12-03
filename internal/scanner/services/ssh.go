package services

import (
	"net"
	"strconv"
	"time"

	"allbr/pkg/types"
	"golang.org/x/crypto/ssh"
)

// TrySSHLogin 尝试SSH登录
func TrySSHLogin(host string, port int, username, password string, timeout time.Duration) types.ScanResult {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ssh",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer client.Close()

	return types.ScanResult{
		Target:   host,
		Service:  "ssh",
		Username: username,
		Password: password,
		Success:  true,
		Error:    "",
	}
}