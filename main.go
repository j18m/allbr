package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"allbr/internal/config"
	"allbr/internal/scanner"
	"allbr/internal/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "allbr",
	Short: "Allbr - 多功能暴力破解工具",
	Long:  "Allbr是一个功能强大的暴力破解工具，支持多种服务类型的密码破解和主机存活检测。",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var brCmd = &cobra.Command{
	Use:   "br",
	Short: "暴力破解模式",
	Long:  "暴力破解指定服务的密码，支持多种服务类型。",
	Run: func(cmd *cobra.Command, args []string) {
		// 解析命令行参数
		cfg := config.ParseFlagsForCommand(cmd)

		// 开始扫描
		results := scanner.StartScan(cfg)

		// 打印结果数量
		if len(results) > 0 {
			fmt.Printf("\n发现 %d 个可用凭据\n", len(results))
			log.Println("详细结果请查看结果文件")
		} else {
			fmt.Println("\n未发现可用凭据")
		}
	},
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "主机存活检测模式",
	Long:  "检测目标IP或IP段的主机存活状态。",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取目标参数
		targetsFlag, _ := cmd.Flags().GetString("targets")
		threadsFlag, _ := cmd.Flags().GetInt("threads")

		// 设置默认线程数
		if threadsFlag <= 0 {
			threadsFlag = 10
		}

		// 解析目标
		targets := config.ParseTargets(targetsFlag)
		if len(targets) == 0 {
			fmt.Println("请指定有效的目标IP或IP段")
			return
		}

		// 执行ping检测
		fmt.Printf("开始Ping存活检测，目标数量: %d\n", len(targets))
		aliveHosts := utils.FilterAliveHosts(targets, threadsFlag)

		// 输出结果
		fmt.Printf("\n存活检测完成\n")
		fmt.Printf("总目标数量: %d\n", len(targets))
		fmt.Printf("存活主机数量: %d\n", len(aliveHosts))

		// 保存结果
		if len(aliveHosts) > 0 {
			resultFile, fileType, err := utils.InitResultFile("")
			if err != nil {
				log.Printf("初始化结果文件失败: %v\n", err)
			} else {
				defer resultFile.Close()
				for _, host := range aliveHosts {
					if fileType == "csv" {
						// CSV格式写入
						writer := csv.NewWriter(resultFile)
						writer.Write([]string{host, "alive"})
						writer.Flush()
					} else {
						// TXT格式写入
						fmt.Fprintf(resultFile, "%s\n", host)
					}
				}
				log.Printf("存活主机列表已保存到: %s\n", utils.GetResultFilePath())
			}
		}
	},
}

func main() {
	// 添加共享的命令行参数
	addCommonFlags(brCmd)
	addCommonFlags(pingCmd)

	// 添加子命令
	rootCmd.AddCommand(brCmd)
	rootCmd.AddCommand(pingCmd)

	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

// addCommonFlags 添加共享的命令行参数
func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("targets", "t", "", "目标IP或IP段，多个用逗号分隔，或使用文件路径，支持CIDR(192.168.1.0/24)和范围(192.168.1.1-100)")
	cmd.Flags().IntP("threads", "n", 10, "并发线程数")

	// 为br命令添加额外的参数
	if cmd.Use == "br" {
		cmd.Flags().StringP("service", "s", "ssh", "服务类型: ssh, mysql, ftp, rdp, ldap, oracle, mongodb, redis")
		cmd.Flags().IntP("port", "p", 0, "目标端口(0=使用服务默认端口)")
		cmd.Flags().StringP("usernames", "u", "", "用户名，多个用逗号分隔，或使用文件路径")
		cmd.Flags().StringP("passwords", "w", "", "密码，多个用逗号分隔，或使用文件路径")
		cmd.Flags().String("strategy", "user-first", "扫描策略: user-first(优先用户名) 或 pass-first(优先密码)")
		cmd.Flags().IntP("timeout", "o", 5, "连接超时时间(秒)")
		cmd.Flags().StringP("output", "O", "", "输出文件")
		cmd.Flags().BoolP("ping-first", "i", true, "先进行ping存活检测")
		cmd.Flags().BoolP("port-check", "c", true, "先进行端口开放检测")
	}
}
