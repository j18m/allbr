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

	// 总是创建新文件，避免结果重复
	fileMode := os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	// 打开文件
	file, err := os.OpenFile(filePath, fileMode, 0644)
	if err != nil {
		return nil, "", err
	}

	// 写入文件头（由于总是创建新文件，不需要检查文件是否存在）
	if fileType == "csv" {
		// CSV格式写入标题行
		writer := csv.NewWriter(file)
		// 写入端口扫描专用的标题行
		writer.Write([]string{"Host", "Port", "Status", "Title"})
		writer.Flush()
	} else {
		// TXT格式写入文件头
		header := fmt.Sprintf("# 扫描结果\n")
		header += fmt.Sprintf("# 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		file.WriteString(header)
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
