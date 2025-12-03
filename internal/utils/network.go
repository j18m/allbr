package utils

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FilterAliveHosts 过滤存活主机
func FilterAliveHosts(targets []string, threads int) []string {
	fmt.Println("开始Ping存活检测...")

	var aliveHosts []string
	var wg sync.WaitGroup

	// 创建任务通道
	jobs := make(chan string, len(targets))
	results := make(chan string, len(targets))

	// 启动工作线程
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go pingWorker(jobs, results, &wg)
	}

	// 发送任务
	go func() {
		for _, target := range targets {
			jobs <- target
		}
		close(jobs)
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(results)
	}()

	// 处理结果
	for host := range results {
		if host != "" {
			aliveHosts = append(aliveHosts, host)
			fmt.Printf("[PING] 存活: %s\n", host)
		}
	}

	fmt.Printf("Ping检测完成，存活主机: %d/%d\n", len(aliveHosts), len(targets))
	return aliveHosts
}

// PingWorker Ping工作线程
func pingWorker(jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for target := range jobs {
		if IsHostAlive(target) {
			results <- target
		} else {
			results <- "" // 发送空字符串表示不存活
		}
	}
}

// IsHostAlive 检测主机是否存活
func IsHostAlive(host string) bool {
	// 尝试ICMP Ping
	if PingHost(host) {
		return true
	}

	// 如果ICMP失败，尝试TCP连接（只扫描最常用的2个端口，并设置更短的超时时间）
	ports := []int{80, 443} // 只扫描HTTP和HTTPS端口
	for _, port := range ports {
		if IsPortOpen(host, port, 500*time.Millisecond) { // 超时时间缩短到500毫秒
			return true
		}
	}

	return false
}

// PingHost ICMP Ping检测
func PingHost(host string) bool {
	// 使用系统ping命令（跨平台兼容）
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", host)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", host)
	}

	err := cmd.Run()
	return err == nil
}

// IsPortOpen 检测端口是否开放
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// FilterPortOpenHosts 过滤开放指定端口的主机
func FilterPortOpenHosts(targets []string, port, threads int) []string {
	fmt.Printf("开始检测端口 %d 开放情况...\n", port)

	var openHosts []string
	var wg sync.WaitGroup

	// 创建任务通道
	jobs := make(chan string, len(targets))
	results := make(chan string, len(targets))

	// 启动工作线程
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go portCheckWorker(jobs, results, port, &wg)
	}

	// 发送任务
	go func() {
		for _, target := range targets {
			jobs <- target
		}
		close(jobs)
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(results)
	}()

	// 处理结果
	for host := range results {
		if host != "" {
			openHosts = append(openHosts, host)
			fmt.Printf("[PORT] 开放: %s:%d\n", host, port)
		}
	}

	fmt.Printf("端口检测完成，开放端口 %d 的主机: %d/%d\n", port, len(openHosts), len(targets))
	return openHosts
}

// PortCheckWorker 端口检测工作线程
func portCheckWorker(jobs <-chan string, results chan<- string, port int, wg *sync.WaitGroup) {
	defer wg.Done()

	for target := range jobs {
		if IsPortOpen(target, port, 3*time.Second) {
			results <- target
		} else {
			results <- "" // 发送空字符串表示端口不开放
		}
	}
}

// ExpandIPRange 扩展：支持IP段解析
func ExpandIPRange(ipRange string) ([]string, error) {
	// CIDR表示法
	if strings.Contains(ipRange, "/") {
		ip, ipnet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, err
		}

		var ips []string

		// 将IP转换为4字节表示
		ip = ip.To4()
		if ip == nil {
			return nil, fmt.Errorf("不是有效的IPv4地址")
		}

		// 计算网络掩码
		mask := ipnet.Mask
		ones, bits := mask.Size()
		if bits != 32 {
			return nil, fmt.Errorf("不是有效的IPv4掩码")
		}

		// 计算网络中的IP数量
		numIPs := 1 << uint(32-ones)

		// 对于小网络（/31, /32），特殊处理
		if ones >= 31 {
			ips = append(ips, ipnet.IP.String())
			if ones == 31 {
				// /31 网络有两个可用地址
				ip := make(net.IP, len(ipnet.IP))
				copy(ip, ipnet.IP)
				ip[3]++
				ips = append(ips, ip.String())
			}
			return ips, nil
		}

		// 对于常规网络，排除网络地址和广播地址
		for i := 1; i < numIPs-1; i++ {
			ip := make(net.IP, len(ipnet.IP))
			copy(ip, ipnet.IP)

			// 递增IP地址
			for j := len(ip) - 1; j >= 0; j-- {
				ip[j] += byte(i >> uint((3-j)*8) & 0xFF)
				if ip[j] != 0 {
					break
				}
			}

			ips = append(ips, ip.String())
		}

		return ips, nil
	}

	// 范围表示法 192.168.1.1-100
	if strings.Contains(ipRange, "-") {
		parts := strings.Split(ipRange, ".")
		if len(parts) != 4 {
			return nil, fmt.Errorf("无效的IP范围格式")
		}

		rangeParts := strings.Split(parts[3], "-")
		if len(rangeParts) != 2 {
			return nil, fmt.Errorf("无效的IP范围格式")
		}

		start, err1 := strconv.Atoi(rangeParts[0])
		end, err2 := strconv.Atoi(rangeParts[1])
		if err1 != nil || err2 != nil || start > end || start < 0 || end > 255 {
			return nil, fmt.Errorf("无效的IP范围")
		}

		var ips []string
		for i := start; i <= end; i++ {
			ips = append(ips, fmt.Sprintf("%s.%s.%s.%d", parts[0], parts[1], parts[2], i))
		}
		return ips, nil
	}

	// 单个IP
	return []string{ipRange}, nil
}
