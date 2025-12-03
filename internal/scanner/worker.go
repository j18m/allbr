package scanner

import (
	"fmt"
	"strings"
	"sync"

	"allbr/internal/scanner/services"
	"allbr/internal/utils"
	"allbr/pkg/types"
)

// worker 工作线程
func worker(config *types.Config, jobs <-chan []string, results *[]types.ScanResult, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		if len(job) != 3 {
			continue
		}

		target := job[0]
		username := job[1]
		password := job[2]

		var result types.ScanResult

		// 根据服务类型调用相应的认证函数
		switch config.Service {
		case "ssh":
			result = services.TrySSHLogin(target, config.Port, username, password, config.Timeout)
		case "mysql":
			result = services.TryMySQLLogin(target, config.Port, username, password, config.Timeout)
		case "ftp":
			result = services.TryFTPLogin(target, config.Port, username, password, config.Timeout)
		case "rdp":
			result = services.TryRDPLogin(target, config.Port, username, password, config.Timeout)
		case "ldap":
			result = services.TryLDAPLogin(target, config.Port, username, password, config.Timeout)
		case "oracle":
			result = services.TryOracleLogin(target, config.Port, username, password, config.Timeout)
		case "mongodb":
			result = services.TryMongoDBLogin(target, config.Port, username, password, config.Timeout)
		case "redis":
			result = services.TryRedisLogin(target, config.Port, password, config.Timeout)
		default:
			result = types.ScanResult{
				Target:   target,
				Service:  config.Service,
				Username: username,
				Password: password,
				Success:  false,
				Error:    "不支持的服务器类型",
			}
		}

		mutex.Lock()
		*results = append(*results, result)
		if result.Success {
			fmt.Printf("[SUCCESS] %s %s:%d - %s/%s\n",
				strings.ToUpper(config.Service), target, config.Port, username, password)
			// 实时保存成功结果
			utils.SaveSuccessResult(resultFile, &resultFileLock, result)
		} else {
			fmt.Printf("[FAIL] %s %s:%d - %s/%s - %s\n",
				strings.ToUpper(config.Service), target, config.Port, username, password, result.Error)
		}
		mutex.Unlock()
	}
}