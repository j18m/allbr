package config

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"allbr/internal/utils"
	"allbr/pkg/types"

	"github.com/spf13/cobra"
)

// ParseFlagsForCommand 从cobra命令中解析参数
func ParseFlagsForCommand(cmd *cobra.Command) *types.Config {
	config := &types.Config{}

	// 从cobra命令中获取参数值
	targets, _ := cmd.Flags().GetString("targets")
	service, _ := cmd.Flags().GetString("service")
	port, _ := cmd.Flags().GetInt("port")
	usernames, _ := cmd.Flags().GetString("usernames")
	passwords, _ := cmd.Flags().GetString("passwords")
	strategy, _ := cmd.Flags().GetString("strategy")
	threads, _ := cmd.Flags().GetInt("threads")
	timeout, _ := cmd.Flags().GetInt("timeout")
	outputFile, _ := cmd.Flags().GetString("output")
	pingFirst, _ := cmd.Flags().GetBool("ping-first")
	portCheck, _ := cmd.Flags().GetBool("port-check")

	// 设置配置项
	config.Service = strings.ToLower(service)

	// 设置端口
	if port == 0 {
		config.Port = types.GetServicePort(config.Service)
	} else {
		config.Port = port
	}

	config.Threads = threads
	config.Timeout = time.Duration(timeout) * time.Second
	config.Strategy = strategy
	config.OutputFile = outputFile
	config.PingFirst = pingFirst
	config.PortCheck = portCheck

	// 解析目标
	config.Targets = ParseTargets(targets)

	// 设置默认用户名和密码
	defaultUsers, defaultPasswords := GetDefaultCredentials(config.Service)

	// 解析用户名
	config.Usernames = ParseInput(usernames, "usernames")
	if len(config.Usernames) == 0 {
		config.Usernames = defaultUsers
	}

	// 解析密码
	config.Passwords = ParseInput(passwords, "passwords")
	if len(config.Passwords) == 0 {
		config.Passwords = defaultPasswords
	}

	return config
}

// 旧版本ParseFlags函数，为了向后兼容保留
func ParseFlags() *types.Config {
	var targets, usernames, passwords, service string
	var port, threads int
	var timeout int
	var strategy, outputFile string
	var pingFirst, portCheck bool

	flag.StringVar(&targets, "t", "", "目标IP或IP段，多个用逗号分隔，或使用文件路径，支持CIDR(192.168.1.0/24)和范围(192.168.1.1-100)")
	flag.StringVar(&targets, "targets", "", "目标IP或IP段，多个用逗号分隔，或使用文件路径，支持CIDR(192.168.1.0/24)和范围(192.168.1.1-100)")
	flag.StringVar(&service, "s", "ssh", "服务类型: ssh, mysql, ftp, rdp, ldap, oracle, mongodb, redis")
	flag.StringVar(&service, "service", "ssh", "服务类型: ssh, mysql, ftp, rdp, ldap, oracle, mongodb, redis")
	flag.IntVar(&port, "p", 0, "目标端口(0=使用服务默认端口)")
	flag.IntVar(&port, "port", 0, "目标端口(0=使用服务默认端口)")
	flag.StringVar(&usernames, "u", "", "用户名，多个用逗号分隔，或使用文件路径")
	flag.StringVar(&usernames, "usernames", "", "用户名，多个用逗号分隔，或使用文件路径")
	flag.StringVar(&passwords, "w", "", "密码，多个用逗号分隔，或使用文件路径")
	flag.StringVar(&passwords, "P", "", "密码，多个用逗号分隔，或使用文件路径(兼容旧版本)")
	flag.StringVar(&passwords, "passwords", "", "密码，多个用逗号分隔，或使用文件路径")
	flag.StringVar(&strategy, "strategy", "user-first", "扫描策略: user-first(优先用户名) 或 pass-first(优先密码)")
	flag.IntVar(&threads, "n", 10, "并发线程数")
	flag.IntVar(&threads, "threads", 10, "并发线程数")
	flag.IntVar(&timeout, "o", 5, "连接超时时间(秒)")
	flag.IntVar(&timeout, "timeout", 5, "连接超时时间(秒)")
	flag.StringVar(&outputFile, "O", "", "输出文件")
	flag.StringVar(&outputFile, "o", "", "输出文件(兼容旧版本)")
	flag.StringVar(&outputFile, "output", "", "输出文件")
	flag.BoolVar(&pingFirst, "i", true, "先进行ping存活检测")
	flag.BoolVar(&pingFirst, "ping", true, "先进行ping存活检测(兼容旧版本)")
	flag.BoolVar(&pingFirst, "ping-first", true, "先进行ping存活检测")
	flag.BoolVar(&portCheck, "c", true, "先进行端口开放检测")
	flag.BoolVar(&portCheck, "port-check", true, "先进行端口开放检测")
	flag.BoolVar(&portCheck, "pc", true, "先进行端口开放检测(兼容旧版本)")

	// 检查是否有命令行参数
	if len(os.Args) == 1 {
		flag.Usage()
		fmt.Println("\n示例:")
		fmt.Println("  allbr -t 192.168.1.1 -u root,admin -P 123456,password")
		fmt.Println("  allbr -t targets.txt -u users.txt -P pass.txt -s ssh")
		fmt.Println("  allbr -t 192.168.1.0/24 -s mysql")
		os.Exit(0)
	}

	flag.Parse()

	// 创建一个临时的cobra命令对象来适配旧的参数解析
	// 这样可以复用ParseFlagsForCommand中的逻辑
	tempCmd := &cobra.Command{}
	tempCmd.Flags().String("targets", targets, "")
	tempCmd.Flags().String("service", service, "")
	tempCmd.Flags().Int("port", port, "")
	tempCmd.Flags().String("usernames", usernames, "")
	tempCmd.Flags().String("passwords", passwords, "")
	tempCmd.Flags().String("strategy", strategy, "")
	tempCmd.Flags().Int("threads", threads, "")
	tempCmd.Flags().Int("timeout", timeout, "")
	tempCmd.Flags().String("output", outputFile, "")
	tempCmd.Flags().Bool("ping-first", pingFirst, "")
	tempCmd.Flags().Bool("port-check", portCheck, "")

	// 复用ParseFlagsForCommand的逻辑来设置配置
	return ParseFlagsForCommand(tempCmd)
}

// GetDefaultCredentials 获取服务的默认凭据
func GetDefaultCredentials(service string) ([]string, []string) {
	switch service {
	case "ssh":
		return []string{"root", "admin", "test", "user", "ubuntu", "centos"},
			[]string{"123456", "password", "admin", "root", "12345678", "test", "1234"}
	case "mysql":
		return []string{"root", "admin", "test", "user"},
			[]string{"123456", "password", "admin", "root", "", "12345678"}
	case "ftp":
		return []string{"anonymous", "ftp", "admin", "root", "test"},
			[]string{"", "anonymous", "ftp", "admin", "root", "123456", "password"}
	case "rdp":
		return []string{"administrator", "admin", "user", "test"},
			[]string{"123456", "password", "admin", "12345678", "Password1", ""}
	case "ldap":
		return []string{"admin", "administrator", "root", "cn=admin"},
			[]string{"123456", "password", "admin", "12345678"}
	case "oracle":
		return []string{"sys", "system", "scott", "admin"},
			[]string{"123456", "password", "admin", "oracle", "change_on_install"}
	case "mongodb":
		return []string{"admin", "root", "user"},
			[]string{"123456", "password", "admin", ""}
	case "redis":
		return []string{""}, // Redis通常不需要用户名
			[]string{"", "123456", "redis", "password", "admin"}
	default:
		return []string{"root", "admin"}, []string{"123456", "password", "admin"}
	}
}

// ParseTargets 解析目标输入，支持文件、逗号分隔、CIDR和IP范围
func ParseTargets(input string) []string {
	if input == "" {
		return []string{}
	}

	// 如果是文件
	if _, err := os.Stat(input); err == nil {
		lines := ReadFile(input)
		return ExpandTargetList(lines)
	}

	// 如果是逗号分隔的字符串
	if strings.Contains(input, ",") {
		targets := SplitAndTrim(input, ",")
		return ExpandTargetList(targets)
	}

	// 单个值
	return ExpandTargetList([]string{strings.TrimSpace(input)})
}

// ExpandTargetList 扩展目标列表，处理CIDR和IP范围
func ExpandTargetList(targets []string) []string {
	var expandedTargets []string

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		// 检查是否是CIDR或IP范围
		if IsIPRange(target) {
			ips, err := utils.ExpandIPRange(target)
			if err != nil {
				log.Printf("警告: 无法解析IP范围 %s: %v", target, err)
				// 如果解析失败，保留原始目标
				expandedTargets = append(expandedTargets, target)
			} else {
				log.Printf("解析IP范围 %s -> %d个IP地址", target, len(ips))
				expandedTargets = append(expandedTargets, ips...)
			}
		} else {
			// 单个IP或主机名
			expandedTargets = append(expandedTargets, target)
		}
	}

	return expandedTargets
}

// IsIPRange 判断是否是IP范围（CIDR或IP范围表示法）
func IsIPRange(target string) bool {
	return strings.Contains(target, "/") || (strings.Contains(target, "-") && strings.Count(target, ".") == 3)
}

// ParseInput 解析输入，支持文件和逗号分隔的字符串
func ParseInput(input, inputType string) []string {
	if input == "" {
		return []string{}
	}

	// 如果是file://前缀的文件路径
	if strings.HasPrefix(input, "file://") {
		filePath := strings.TrimPrefix(input, "file://")
		// 处理Windows路径格式（如果前缀是/开头，去掉/）
		if len(filePath) > 1 && filePath[0] == '/' && (filePath[1] >= 'A' && filePath[1] <= 'Z' || filePath[1] >= 'a' && filePath[1] <= 'z') && filePath[2] == ':' {
			filePath = filePath[1:]
		}
		return ReadFile(filePath)
	}

	// 如果是普通文件路径（向后兼容）
	if _, err := os.Stat(input); err == nil {
		return ReadFile(input)
	}

	// 如果是逗号分隔的字符串
	if strings.Contains(input, ",") {
		return SplitAndTrim(input, ",")
	}

	// 单个值
	return []string{strings.TrimSpace(input)}
}

// ReadFile 读取文件内容
func ReadFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("无法打开文件 %s: %v", filename, err)
		return []string{}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取文件 %s 错误: %v", filename, err)
	}

	return lines
}

// SplitAndTrim 分割并去除空格
func SplitAndTrim(input, sep string) []string {
	parts := strings.Split(input, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ValidateConfig 验证配置
func ValidateConfig(config *types.Config) error {
	if len(config.Targets) == 0 {
		return fmt.Errorf("必须指定至少一个目标")
	}

	if len(config.Usernames) == 0 {
		return fmt.Errorf("必须指定至少一个用户名")
	}

	if len(config.Passwords) == 0 {
		return fmt.Errorf("必须指定至少一个密码")
	}

	if _, exists := types.ServicePorts[config.Service]; !exists {
		return fmt.Errorf("不支持的服务类型: %s", config.Service)
	}

	if config.Strategy != "user-first" && config.Strategy != "pass-first" {
		return fmt.Errorf("策略必须是 user-first 或 pass-first")
	}

	if config.Threads <= 0 {
		return fmt.Errorf("线程数必须大于0")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内")
	}

	return nil
}
