/**
 * Description：
 * FileName：models_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/9 15:54:49
 * Remark：
 */

package models

import (
	"database/sql"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/json_format"
	uuid7 "github.com/gofrs/uuid"
	uuid4 "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	// 生成一个随机的 UUID v4 (最常用)
	newUUID := uuid4.New()
	fmt.Println("UUIDv4:", newUUID, len(newUUID))

	// 生成 UUID v7
	u7, err := uuid7.NewV7()
	if err != nil {
		panic("failed to generate UUID v7: " + err.Error())
	}

	fmt.Println("UUID v7:", u7, len(u7))
	fmt.Println("String representation:", u7.String(), len(u7.String()))
}

func TestTimestamp(t *testing.T) {
	fmt.Println("time.Now() >>> ", time.Now())
	fmt.Println("time.Now().Unix() >>> ", time.Now().Unix())
	fmt.Println("time.Now().UnixNano() >>> ", time.Now().UnixNano())
	fmt.Println("time.Now().UnixMilli() >>> ", time.Now().UnixMilli())
	fmt.Println("time.Now().UnixMicro() >>> ", time.Now().UnixMicro())
}

func TestPassword(t *testing.T) {
	// 加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("careful@222"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("密码加密失败: %v", err)
	}
	fmt.Println(string(hashedPassword))

	// 解密
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("careful@222"))
	fmt.Println("err >>> ", err)
	if err != nil {
		fmt.Printf("密码解密失败: %v", err)
	}
}

func TestSqlNull(t *testing.T) {
	type User struct {
		ID        sql.NullInt64
		Name      sql.NullString
		IsActive  sql.NullBool
		Score     sql.NullFloat64
		BirthDate sql.NullTime
	}

	user := User{}

	// 设置整数
	user.ID = sql.NullInt64{Int64: 100, Valid: true}
	// 设置字符串
	user.Name = sql.NullString{String: "Alice", Valid: true}
	// 设置布尔值
	user.IsActive = sql.NullBool{Bool: true, Valid: true}
	// 设置浮点数
	user.Score = sql.NullFloat64{Float64: 95.5, Valid: true}
	// 设置时间
	user.BirthDate = sql.NullTime{Time: time.Now(), Valid: true}

	fmt.Println("user >>> ")
	json_format.PrintFormattedJSON(user, "")

	// 设置整数为 NULL
	user.ID = sql.NullInt64{Int64: 0, Valid: false}
	// 设置字符串为 NULL
	user.Name = sql.NullString{String: "", Valid: false}
	// 设置布尔值为 NULL
	user.IsActive = sql.NullBool{Bool: false, Valid: false}
	// 设置浮点数为 NULL
	user.Score = sql.NullFloat64{Float64: 0, Valid: false}
	// 设置时间为 NULL
	user.BirthDate = sql.NullTime{Time: time.Time{}, Valid: false}

	fmt.Println("user >>> ")
	json_format.PrintFormattedJSON(user, "")

	// 检查并读取整数值
	if user.ID.Valid {
		fmt.Println("User ID:", user.ID.Int64)
	} else {
		fmt.Println("User ID is NULL")
	}

	// 检查并读取字符串值
	if user.Name.Valid {
		fmt.Println("Name:", user.Name.String)
	} else {
		fmt.Println("Name is NULL")
	}

	// 检查并读取布尔值
	if user.IsActive.Valid {
		fmt.Println("Active:", user.IsActive.Bool)
	} else {
		fmt.Println("Active status is NULL")
	}

	// 使用示例
	user.Name = NewNullString("")     // 设置为 NULL
	user.Name = NewNullString("John") // 设置有效值
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
