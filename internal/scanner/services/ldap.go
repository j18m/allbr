package services

import (
	"net"
	"strconv"
	"time"

	"allbr/pkg/types"
	"github.com/go-ldap/ldap/v3"
)

// TryLDAPLogin 尝试LDAP登录
func TryLDAPLogin(host string, port int, username, password string, timeout time.Duration) types.ScanResult {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ldap",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer conn.Close()

	// 使用go-ldap库进行认证
	l, err := ldap.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ldap",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer l.Close()

	// 尝试绑定（认证）
	err = l.Bind(username, password)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "ldap",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}

	return types.ScanResult{
		Target:   host,
		Service:  "ldap",
		Username: username,
		Password: password,
		Success:  true,
		Error:    "",
	}
}