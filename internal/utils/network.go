package utils

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
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

// PortScanResult 端口探测结果结构体
type PortScanResult struct {
	Host   string
	Port   int
	Status string
	Title  string
}

// ScanPorts 执行端口扫描
func ScanPorts(targets []string, ports []int, threads int, getTitle bool, probeOrder string) []PortScanResult {
	var results []PortScanResult
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	// 创建任务通道
	jobs := make(chan struct {
		host string
		port int
	}, len(targets)*len(ports))

	// 启动工作线程
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go portScanWorker(jobs, &results, &resultMutex, getTitle, &wg)
	}

	// 发送任务，根据探测顺序决定发送方式
	if probeOrder == "ip" {
		// 优先探测同一个IP的多个端口
		for _, host := range targets {
			for _, port := range ports {
				jobs <- struct {
					host string
					port int
				}{host, port}
			}
		}
	} else {
		// 优先探测多个IP的同一个端口（默认）
		for _, port := range ports {
			for _, host := range targets {
				jobs <- struct {
					host string
					port int
				}{host, port}
			}
		}
	}

	close(jobs)
	wg.Wait()

	return results
}

// portScanWorker 端口扫描工作线程
func portScanWorker(jobs <-chan struct {
	host string
	port int
}, results *[]PortScanResult, resultMutex *sync.Mutex, getTitle bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		status := "closed"
		title := ""

		// 检查端口是否开放
		if IsPortOpen(job.host, job.port, 2*time.Second) {
			status = "open"

			// 如果需要获取标题且是HTTP服务，尝试获取标题
			if getTitle && (job.port == 80 || job.port == 443) {
				scheme := "http"
				if job.port == 443 {
					scheme = "https"
				}
				title = GetHTTPTitle(fmt.Sprintf("%s://%s:%d", scheme, job.host, job.port))
			} else if getTitle {
				// 尝试HTTP和HTTPS协议
				title = GetHTTPTitle(fmt.Sprintf("http://%s:%d", job.host, job.port))
				if title == "" {
					title = GetHTTPTitle(fmt.Sprintf("https://%s:%d", job.host, job.port))
				}
			}

			// 输出结果到控制台
			if title != "" {
				fmt.Printf("[PORT] 开放: %s:%d - %s (%s)\n", job.host, job.port, status, title)
			} else {
				fmt.Printf("[PORT] 开放: %s:%d - %s\n", job.host, job.port, status)
			}

			// 添加结果到列表
			resultMutex.Lock()
			*results = append(*results, PortScanResult{
				Host:   job.host,
				Port:   job.port,
				Status: status,
				Title:  title,
			})
			resultMutex.Unlock()
		}
	}
}

// GetHTTPTitle 获取HTTP服务的标题
func GetHTTPTitle(url string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// 读取响应内容的前10KB来查找标题
	reader := bufio.NewReader(resp.Body)
	body, _ := io.ReadAll(io.LimitReader(reader, 10240))
	
	// 尝试使用UTF-8或GBK编码解码内容
	bodyStr := tryDecodeUTF8(body)

	// 使用正则表达式查找标题
	titleRegex := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	match := titleRegex.FindStringSubmatch(bodyStr)
	if len(match) > 1 {
		// 清理标题
		title := strings.TrimSpace(match[1])
		title = strings.ReplaceAll(title, "\n", "")
		title = strings.ReplaceAll(title, "\r", "")
		title = strings.ReplaceAll(title, "\t", "")
		return title
	}

	return ""
}

// tryDecodeUTF8 尝试使用UTF-8或GBK编码解码内容
func tryDecodeUTF8(body []byte) string {
	// 首先尝试使用UTF-8
	bodyStr := string(body)
	
	// 如果包含乱码，尝试使用GBK
	if containsGarbledText(bodyStr) {
		decoder := simplifiedchinese.GBK.NewDecoder()
		if gbkBody, err := decoder.Bytes(body); err == nil {
			bodyStr = string(gbkBody)
		}
	}
	
	return bodyStr
}

// containsGarbledText 检查字符串是否包含乱码
func containsGarbledText(s string) bool {
	// 检查是否包含典型的乱码字符
	garbledPatterns := []string{"鐢ㄦ埛", "鎿嶄綔", "鏁版嵁", "缃戠粶", "瑕佹眰"}
	for _, pattern := range garbledPatterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
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
