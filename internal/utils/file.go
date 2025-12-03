package utils

import (
	"encoding/csv"
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

// InitResultFile 初始化结果文件，默认txt格式
func InitResultFile(outputPath string) (*os.File, string, error) {
	var filePath string
	var fileType string

	// 如果没有指定输出路径，使用默认路径
	if outputPath == "" {
		filePath = GetResultFilePath()
		fileType = "txt"
	} else {
		filePath = outputPath
		// 根据文件扩展名判断类型
		if strings.HasSuffix(strings.ToLower(filePath), ".csv") {
			fileType = "csv"
		} else {
			fileType = "txt"
		}
	}

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	fileMode := os.O_WRONLY | os.O_APPEND
	if os.IsNotExist(err) {
		fileMode |= os.O_CREATE
	}

	// 打开文件
	file, err := os.OpenFile(filePath, fileMode, 0644)
	if err != nil {
		return nil, "", err
	}

	// 如果是新文件，写入文件头
	if os.IsNotExist(err) {
		if fileType == "csv" {
			// CSV格式写入标题行
			writer := csv.NewWriter(file)
			writer.Write([]string{"服务类型", "目标IP", "端口", "用户名", "密码", "时间"})
			writer.Flush()
		} else {
			// TXT格式写入文件头
			header := fmt.Sprintf("# 多服务暴力破解扫描结果\n")
			header += fmt.Sprintf("# 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			header += fmt.Sprintf("# 格式: 服务类型 目标IP:端口 用户名 密码\n\n")
			file.WriteString(header)
		}
	} else {
		// 文件存在，添加分隔符（仅TXT格式）
		if fileType == "txt" {
			separator := fmt.Sprintf("\n%s 新的扫描会话开始 %s\n",
				strings.Repeat("=", 20), time.Now().Format("2006-01-02 15:04:05"))
			file.WriteString(separator)
		}
	}

	return file, fileType, nil
}

// SaveSuccessResult 实时保存成功结果到文件，支持txt和csv格式
func SaveSuccessResult(resultFile *os.File, resultFileLock *sync.Mutex, fileType string, result types.ScanResult) {
	resultFileLock.Lock()
	defer resultFileLock.Unlock()

	if fileType == "csv" {
		// CSV格式保存
		writer := csv.NewWriter(resultFile)
		writer.Write([]string{
			strings.ToUpper(result.Service),
			result.Target,
			fmt.Sprintf("%d", types.GetServicePort(result.Service)),
			result.Username,
			result.Password,
			time.Now().Format("2006-01-02 15:04:05"),
		})
		writer.Flush()
		if writer.Error() != nil {
			log.Printf("写入CSV结果失败: %v", writer.Error())
		}
	} else {
		// TXT格式保存
		line := fmt.Sprintf("%s %s:%d %s %s\n",
			strings.ToUpper(result.Service), result.Target, types.GetServicePort(result.Service), result.Username, result.Password)

		if _, err := resultFile.WriteString(line); err != nil {
			log.Printf("写入结果文件失败: %v", err)
		}
	}

	// 立即刷新到磁盘
	resultFile.Sync()
}
