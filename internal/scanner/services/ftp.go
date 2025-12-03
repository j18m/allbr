package services

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"time"

	"allbr/pkg/types"
)

// TryFTPLogin 尝试FTP登录
func TryFTPLogin(host string, port int, username, password string, timeout time.Duration) types.ScanResult {
	// 使用简单的TCP连接尝试FTP登录
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ftp",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer conn.Close()

	// 读取欢迎消息
	reader := bufio.NewReader(conn)
	_, err = reader.ReadString('\n')
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ftp",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}

	// 发送USER命令
	conn.Write([]byte("USER " + username + "\r\n"))
	response, _ := reader.ReadString('\n')

	if !strings.HasPrefix(response, "331") {
		return types.ScanResult{
			Target:   host,
			Service:  "ftp",
			Username: username,
			Password: password,
			Success:  false,
			Error:    "USER命令失败: " + response,
		}
	}

	// 发送PASS命令
	conn.Write([]byte("PASS " + password + "\r\n"))
	response, _ = reader.ReadString('\n')

	if strings.HasPrefix(response, "230") {
		return types.ScanResult{
			Target:   host,
			Service:  "ftp",
			Username: username,
			Password: password,
			Success:  true,
			Error:    "",
		}
	}

	return types.ScanResult{
		Target:   host,
		Service:  "ftp",
		Username: username,
		Password: password,
		Success:  false,
		Error:    "登录失败: " + response,
	}
}