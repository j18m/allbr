package output

import (
	"fmt"
	"strings"

	"allbr/pkg/types"
)

// OutputResults 输出结果
func OutputResults(results []types.ScanResult, outputFile string) {
	var successCount int

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("扫描结果:")
	fmt.Println(strings.Repeat("=", 50))

	for _, result := range results {
		if result.Success {
			successCount++
			line := fmt.Sprintf("SUCCESS: %s %s - %s/%s\n",
				strings.ToUpper(result.Service), result.Target, result.Username, result.Password)
			fmt.Print(line)
		}
	}

	fmt.Printf("\n扫描完成! 成功: %d/%d\n", successCount, len(results))
}