package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"allbr/pkg/types"
)

// GetResultFilePath 获取结果文件路径
func GetResultFilePath() string {
	exePath, err := os.Executable()
	if err != nil {
		// 如果无法获取可执行文件路径，使用当前目录
		return "multibrute.txt"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "multibrute.txt")
}

// InitResultFile 初始化结果文件
func InitResultFile() (*os.File, error) {
	filePath := GetResultFilePath()

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// 文件不存在，创建新文件
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}

		// 写入文件头
		header := fmt.Sprintf("# 多服务暴力破解扫描结果\n")
		header += fmt.Sprintf("# 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		header += fmt.Sprintf("# 格式: 服务类型 目标IP:端口 用户名 密码\n\n")
		file.WriteString(header)
		return file, nil
	} else {
		// 文件存在，以追加模式打开
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		// 添加分隔符和新的扫描会话标记
		separator := fmt.Sprintf("\n%s 新的扫描会话开始 %s\n",
			strings.Repeat("=", 20), time.Now().Format("2006-01-02 15:04:05"))
		file.WriteString(separator)
		return file, nil
	}
}

// SaveSuccessResult 实时保存成功结果到文件
func SaveSuccessResult(resultFile *os.File, resultFileLock *sync.Mutex, result types.ScanResult) {
	resultFileLock.Lock()
	defer resultFileLock.Unlock()

	line := fmt.Sprintf("%s %s:%d %s %s\n",
		strings.ToUpper(result.Service), result.Target, types.GetServicePort(result.Service), result.Username, result.Password)

	if _, err := resultFile.WriteString(line); err != nil {
		log.Printf("写入结果文件失败: %v", err)
	}

	// 立即刷新到磁盘
	resultFile.Sync()
}
