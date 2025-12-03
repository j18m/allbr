package main

import (
	"allbr/internal/config"
	"allbr/internal/scanner"
)

func main() {
	// 解析命令行参数
	config := config.ParseFlags()

	// 开始扫描
	scanner.StartScan(config)
}