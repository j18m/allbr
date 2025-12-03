package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"allbr/pkg/types"
	_ "github.com/go-sql-driver/mysql"
)

// TryMySQLLogin 尝试MySQL登录
func TryMySQLLogin(host string, port int, username, password string, timeout time.Duration) types.ScanResult {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%s",
		username, password, host, port, timeout)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "mysql",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer db.Close()

	db.SetConnMaxLifetime(timeout)

	// 尝试ping数据库
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return types.ScanResult{
			Target:   host,
			Service:  "mysql",
			Username: username,
			Password: password,
			Success:  false,
			Error:    err.Error(),
		}
	}

	return types.ScanResult{
		Target:   host,
		Service:  "mysql",
		Username: username,
		Password: password,
		Success:  true,
		Error:    "",
	}
}