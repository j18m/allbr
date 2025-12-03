package types

import (
	"time"
)

// Config 配置结构体
type Config struct {
	Targets    []string
	Port       int
	Usernames  []string
	Passwords  []string
	Service    string // 服务类型: ssh, mysql, ftp, rdp, ldap, oracle, mongodb, redis
	Strategy   string // "user-first" 或 "pass-first"
	Threads    int
	Timeout    time.Duration
	OutputFile string
	PingFirst  bool // 是否先进行ping检测
	PortCheck  bool // 是否进行端口开放检测
}

// ScanResult 扫描结果
type ScanResult struct {
	Target   string
	Service  string
	Username string
	Password string
	Success  bool
	Error    string
}

// ServicePorts 服务默认端口映射
var ServicePorts = map[string]int{
	"ssh":     22,
	"mysql":   3306,
	"ftp":     21,
	"rdp":     3389,
	"ldap":    389,
	"oracle":  1521,
	"mongodb": 27017,
	"redis":   6379,
}

// GetServicePort 获取服务默认端口
func GetServicePort(service string) int {
	if port, exists := ServicePorts[service]; exists {
		return port
	}
	return 22 // 默认SSH端口
}
