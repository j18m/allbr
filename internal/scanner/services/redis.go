package services

import (
	"context"
	"net"
	"strconv"
	"time"

	"allbr/pkg/types"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// TryRedisLogin 尝试Redis登录
func TryRedisLogin(host string, port int, password string, timeout time.Duration) types.ScanResult {
	// Redis通常不需要用户名
	client := redis.NewClient(&redis.Options{
		Addr:        net.JoinHostPort(host, strconv.Itoa(port)),
		Password:    password,
		DB:          0,
		DialTimeout: timeout,
	})
	defer client.Close()

	// 测试连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "redis",
			Username: "", // Redis通常不需要用户名
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}

	return types.ScanResult{
		Target:   host,
		Service:  "redis",
		Username: "",
		Password: password,
		Success:  true,
		Error:    "",
	}
}