/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/9 17:02:01
 * Remark：
 */

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/ioc"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
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
	// 初始化远程配置
	remoteConfig := ioc.InitLoadNacosConfig(configManager.Config)
	// 初始化数据库池
	dbPool := ioc.NewDbPool(remoteConfig.DatabaseConfig)
	configManager.RelyConfig.Db = config.Database{
		Careful: dbPool.CarefulDB,
	}

	// 自动迁移表
	system.NewUser().AutoMigrate(configManager.RelyConfig.Db.Careful)
	system.NewDept().AutoMigrate(configManager.RelyConfig.Db.Careful)

	// 创建部门
	err := ensureDefaultDept(configManager.RelyConfig.Db.Careful)
	if err != nil {
		fmt.Printf("创建部门失败: %v\n", err)
		os.Exit(1)
	}

	// 创建超级用户
	if err := createSuperUser(configManager.RelyConfig.Db.Careful); err != nil {
		fmt.Printf("创建超级用户失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("超级用户创建成功！")
}

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

	// 设置密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("密码加密失败: %v", err)
	}
	user.Password = string(hashedPassword)

	// 姓名
	fmt.Print("姓名: ")
	name, _ := reader.ReadString('\n')
	user.Name = strings.TrimSpace(name)

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

	// 头像
	fmt.Print("头像URL (可选): ")
	avatar, _ := reader.ReadString('\n')
	user.Avatar = strings.TrimSpace(avatar)

	// 获取可用部门列表
	var deptList []system.Dept
	if err := db.Where("status = ?", true).Find(&deptList).Error; err != nil {
		return fmt.Errorf("获取部门列表失败: %v", err)
	}
	fmt.Println("\n可用部门:")
	fmt.Println("0. 不分配部门")
	for i, dept := range deptList {
		fmt.Printf("%d. %s (%s)\n", i+1, dept.Name, dept.Code)
	}

	// 设置创建者和部门
	// 选择部门
	fmt.Printf("请选择部门 [0-%d] (0表示不分配部门): ", len(deptList))
	deptInput, _ := reader.ReadString('\n')
	deptInput = strings.TrimSpace(deptInput)

	if deptInput == "" || deptInput == "0" {
		user.DeptId = sql.NullString{
			String: "",
			Valid:  false,
		} // 不分配部门
		fmt.Println("用户将不分配到任何部门")
	} else {
		// 解析用户输入
		var deptIndex int
		if _, err := fmt.Sscanf(deptInput, "%d", &deptIndex); err != nil || deptIndex < 1 || deptIndex > len(deptList) {
			return fmt.Errorf("无效的部门选择")
		}
		selectedDept := &deptList[deptIndex-1]
		user.DeptId = sql.NullString{
			String: selectedDept.Id,
			Valid:  true,
		}
		fmt.Printf("用户将分配到部门: %s\n", selectedDept.Name)
	}

	// 验证数据
	if err := user.Validate(); err != nil {
		return err
	}

	user.Id = generateId()
	user.Timestamp = generateTimestamp()

	// 创建用户
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("数据库创建失败: %v", err)
	}

	fmt.Printf("\n超级用户 '%s' 创建成功！\n", user.Username)

	if user.DeptId.Valid {
		var dept system.Dept
		if err := db.Where("id = ?", user.DeptId.String).First(&dept).Error; err == nil {
			fmt.Printf("所属部门: %s\n", dept.Name)
		}
	} else {
		fmt.Println("未分配部门")
	}

	return nil
}

func ensureDefaultDept(db *gorm.DB) error {
	var count int64
	if err := db.Model(&system.Dept{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查部门表失败: %v", err)
	}
	// 如果没有部门，创建默认部门
	if count == 0 {
		defaultDept := &system.Dept{
			CoreModels: models.CoreModels{
				Id:        generateId(),
				Timestamp: generateTimestamp(),
			},
			Status: true,
			Name:   "用心集团有限公司",
			Code:   "CAREFUL-COMPANY",
			Owner:  "careful",
			Phone:  "13888888888",
			Email:  "careful@gmail.com",
			ParentID: sql.NullString{
				String: "",
				Valid:  false,
			},
			Level:      0,
			Path:       "",
			UserCount:  0,
			ChildCount: 0,
		}

		if err := db.Create(defaultDept).Error; err != nil {
			return fmt.Errorf("创建默认部门失败: %v", err)
		}

		fmt.Println("已创建默认部门: ", defaultDept.Name)
	}

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
