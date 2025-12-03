package scanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"allbr/internal/output"
	"allbr/internal/utils"
	"allbr/pkg/types"
)

// 全局变量，用于实时保存结果
var (
	resultFile     *os.File
	resultFileLock sync.Mutex
	ctx            = context.Background()
)

// StartScan 开始扫描
func StartScan(config *types.Config) []types.ScanResult {
	// 验证配置
	// 基本验证逻辑

	// 初始化结果文件
	var err error
	resultFile, err = utils.InitResultFile()
	if err != nil {
		log.Fatal("初始化结果文件失败:", err)
	}
	defer resultFile.Close()

	fmt.Printf("开始%s暴力破解扫描...\n", strings.ToUpper(config.Service))
	fmt.Printf("目标数量: %d\n", len(config.Targets))
	fmt.Printf("用户名数量: %d\n", len(config.Usernames))
	fmt.Printf("密码数量: %d\n", len(config.Passwords))
	fmt.Printf("服务类型: %s\n", config.Service)
	fmt.Printf("目标端口: %d\n", config.Port)
	fmt.Printf("扫描策略: %s\n", config.Strategy)
	fmt.Printf("线程数: %d\n", config.Threads)
	fmt.Printf("超时时间: %v\n", config.Timeout)
	fmt.Printf("Ping检测: %v\n", config.PingFirst)
	fmt.Printf("端口检测: %v\n", config.PortCheck)
	fmt.Printf("结果文件: %s\n", utils.GetResultFilePath())
	fmt.Println(strings.Repeat("=", 50))

	// 如果启用ping检测，先过滤存活主机
	if config.PingFirst {
		config.Targets = utils.FilterAliveHosts(config.Targets, config.Threads)
		if len(config.Targets) == 0 {
			fmt.Println("没有检测到存活主机，扫描终止")
			return []types.ScanResult{}
		}
		fmt.Printf("存活主机数量: %d\n", len(config.Targets))
	}

	// 如果启用端口检测，过滤开放指定端口的主机
	if config.PortCheck {
		config.Targets = utils.FilterPortOpenHosts(config.Targets, config.Port, config.Threads)
		if len(config.Targets) == 0 {
			fmt.Printf("没有检测到开放端口 %d 的主机，扫描终止\n", config.Port)
			return []types.ScanResult{}
		}
		fmt.Printf("开放端口 %d 的主机数量: %d\n", config.Port, len(config.Targets))
	}

	// 执行扫描
	results := executeScan(config)

	// 输出结果
	output.OutputResults(results, config.OutputFile)

	return results
}

// executeScan 执行扫描
func executeScan(config *types.Config) []types.ScanResult {
	var results []types.ScanResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 创建工作通道
	jobs := make(chan []string, config.Threads*10)

	// 启动工作线程
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go worker(config, jobs, &results, &mutex, &wg)
	}

	// 生成任务
	go func() {
		generateJobs(config, jobs)
		close(jobs)
	}()

	// 等待所有工作完成
	wg.Wait()

	return results
}

// generateJobs 生成扫描任务
func generateJobs(config *types.Config, jobs chan<- []string) {
	if config.Strategy == "user-first" {
		// 优先用户名策略：对每个用户名，尝试所有密码
		for _, target := range config.Targets {
			for _, username := range config.Usernames {
				for _, password := range config.Passwords {
					jobs <- []string{target, username, password}
				}
			}
		}
	} else {
		// 优先密码策略：对每个密码，尝试所有用户名
		for _, target := range config.Targets {
			for _, password := range config.Passwords {
				for _, username := range config.Usernames {
					jobs <- []string{target, username, password}
				}
			}
		}
	}
}
