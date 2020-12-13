package mysql

import (
	"context"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Brute is
func Brute(address string, usernames, passwords []string) (string, string, bool) {
	for _, username := range usernames {
		for _, password := range passwords {
			if Login(address, username, password) {
				return username, password, true
			}
		}
	}
	return "", "", false
}

// Login is
func Login(address string, username, password string) bool {
	connector, err := mysql.NewConnector(&mysql.Config{
		User:                    username,
		Passwd:                  password,
		Addr:                    address,
		DBName:                  "mysql",
		Collation:               "utf8mb4_general_ci",
		AllowCleartextPasswords: true,
		AllowNativePasswords:    true,
		AllowOldPasswords:       true,
	})
	if err != nil {
		// fmt.Println("1", err)
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := connector.Connect(ctx)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	return true
}
