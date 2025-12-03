package services

import (
	"context"
	"fmt"
	"time"

	"allbr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TryMongoDBLogin 尝试MongoDB登录
func TryMongoDBLogin(host string, port int, username, password string, timeout time.Duration) types.ScanResult {
	// 创建MongoDB连接选项
	clientOptions := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	// 连接MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "mongodb",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer client.Disconnect(context.Background())

	// 检查连接
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "mongodb",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}

	return types.ScanResult{
		Target:   host,
		Service:  "mongodb",
		Username: username,
		Password: password,
		Success:  true,
		Error:    "",
	}
}