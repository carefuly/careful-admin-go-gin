/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/11/24 16:18:31
 * Remark：
 */

package main

import (
	"bufio"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/ioc"
	uuid7 "github.com/gofrs/uuid"
	uuid4 "github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== CarefulAdmin 超级用户初始化工具 ===")

	// 初始化日志
	loggerManager := ioc.InitLogger()

	// 初始化配置管理器
	configManager := ioc.InitConfig("./application.yaml")
	configManager.RelyConfig.Logger = loggerManager.GetLogger()
	// 启动配置文件监听
	if err := configManager.StartWatching(); err != nil {
		zap.S().Fatal("启动配置文件监听失败", err)
	}
	defer configManager.StopWatching()

	// 初始化数据库池
	dbPool := ioc.NewDbPool(configManager.Config.Database)
	configManager.RelyConfig.Db = config.Database{
		Careful: dbPool.Careful,
	}

	// 自动迁移表
	system.NewUser().AutoMigrate(configManager.RelyConfig.Db.Careful)

	// 创建超级用户
	if err := createSuperUser(configManager.RelyConfig.Db.Careful); err != nil {
		fmt.Printf("创建超级用户失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("超级用户创建成功！")
}

// createSuperUser 创建超级用户
func createSuperUser(db *gorm.DB) error {
	reader := bufio.NewReader(os.Stdin)
	user := system.NewUser()

	fmt.Println("\n请填写超级用户信息:")

	// 用户名
	fmt.Print("用户名: ")
	username, _ := reader.ReadString('\n')
	user.Username = strings.TrimSpace(username)

	// 检查用户名是否已存在
	var existingUser system.User
	if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return fmt.Errorf("用户名 '%s' 已存在", user.Username)
	}

	// 密码
	fmt.Print("密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// 确认密码
	fmt.Print("确认密码: ")
	confirmPassword, _ := reader.ReadString('\n')
	confirmPassword = strings.TrimSpace(confirmPassword)

	if password != confirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 设置密码 (加密)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}
	user.Password = string(hashedPassword)

	// 性别
	fmt.Print("性别 (1-男, 2-女, 3-保密) [默认:1]: ")
	genderInput, _ := reader.ReadString('\n')
	genderInput = strings.TrimSpace(genderInput)
	if genderInput == "" {
		genderInput = "1"
	}

	switch genderInput {
	case "1":
		user.Gender = 1 // Male
	case "2":
		user.Gender = 2 // Female
	case "3":
		user.Gender = 3 // Secret
	default:
		return fmt.Errorf("无效的性别选项")
	}

	// 邮箱
	fmt.Print("邮箱: ")
	email, _ := reader.ReadString('\n')
	user.Email = strings.TrimSpace(email)

	// 手机
	fmt.Print("手机: ")
	mobile, _ := reader.ReadString('\n')
	user.Mobile = strings.TrimSpace(mobile)

	// 姓名
	fmt.Print("姓名: ")
	name, _ := reader.ReadString('\n')
	user.Name = strings.TrimSpace(name)

	// 头像 (可选)
	fmt.Print("头像URL (可选): ")
	avatar, _ := reader.ReadString('\n')
	user.Avatar = strings.TrimSpace(avatar)

	// 生日 (可选)
	fmt.Print("生日 (格式: YYYY-MM-DD, 可选): ")
	birthdayStr, _ := reader.ReadString('\n')
	birthdayStr = strings.TrimSpace(birthdayStr)
	if birthdayStr != "" {
		birthday, err := time.Parse("2006-01-02", birthdayStr)
		if err != nil {
			return fmt.Errorf("生日格式不正确，请使用 YYYY-MM-DD 格式")
		}
		user.Birthday = &birthday
	}

	// 所在城市 (可选)
	fmt.Print("所在城市 (可选): ")
	city, _ := reader.ReadString('\n')
	user.City = strings.TrimSpace(city)

	// 详细地址 (可选)
	fmt.Print("详细地址 (可选): ")
	address, _ := reader.ReadString('\n')
	user.Address = strings.TrimSpace(address)

	// 个人简介 (可选)
	fmt.Print("个人简介 (可选): ")
	bio, _ := reader.ReadString('\n')
	user.Bio = strings.TrimSpace(bio)

	// 验证数据 (如果 Validate 方法已实现)
	if err := user.Validate(); err != nil {
		return err
	}

	user.Id = generateId()
	user.Timestamp = generateTimestamp()

	// 创建用户
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("数据库创建失败: %v", err)
	}

	// 打印创建成功的信息
	fmt.Printf("\n超级用户 '%s' 创建成功！\n", user.Username)
	fmt.Println("用户信息汇总:")
	fmt.Printf("ID: %s\n", user.Id)
	fmt.Printf("用户名: %s\n", user.Username)
	fmt.Printf("姓名: %s\n", user.Name)
	fmt.Printf("性别: %v\n", user.Gender)
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("手机: %s\n", user.Mobile)
	if user.Birthday != nil {
		fmt.Printf("生日: %s\n", user.Birthday.Format("2006-01-02"))
	} else {
		fmt.Println("生日: 未设置")
	}
	fmt.Printf("所在城市: %s\n", user.City)
	fmt.Printf("详细地址: %s\n", user.Address)
	fmt.Printf("个人简介: %s\n", user.Bio)

	return nil
}

func generateId() string {
	var id string
	u7, err := uuid7.NewV7()
	if err != nil {
		id = strings.ToUpper(uuid4.New().String())
	} else {
		id = strings.ToUpper(u7.String())
	}
	return id
}

func generateTimestamp() int64 {
	return time.Now().UnixMicro()
}
