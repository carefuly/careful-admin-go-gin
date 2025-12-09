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
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/ioc"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/dept"
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

	// 初始化数据库池
	dbPool := ioc.NewDbPool(configManager.Config.Database)
	configManager.RelyConfig.Db = config.Database{
		Careful: dbPool.Careful,
	}

	// 自动迁移表
	system.NewUser().AutoMigrate(configManager.RelyConfig.Db.Careful)
	system.NewDept().AutoMigrate(configManager.RelyConfig.Db.Careful)
	system.NewMenu().AutoMigrate(configManager.RelyConfig.Db.Careful)

	// 创建菜单
	err := ensureDefaultMenu(configManager.RelyConfig.Db.Careful)
	if err != nil {
		fmt.Printf("创建菜单失败: %v\n", err)
		os.Exit(1)
	}

	// 创建部门
	err = ensureDefaultDept(configManager.RelyConfig.Db.Careful)
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

// ensureDefaultMenu 创建根菜单
func ensureDefaultMenu(db *gorm.DB) error {
	var count int64
	if err := db.Model(&system.Menu{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查菜单表失败: %v", err)
	}

	// 如果没有菜单，创建根菜单
	if count == 0 {
		// -------------------------- 1. 初始化根菜单 (ID = "root") --------------------------
		var rootMenu system.Menu
		rootMenuID := "root" // 根菜单的固定ID
		// 先查询ID为 "root" 的菜单是否存在
		if err := db.Where("id = ?", rootMenuID).First(&rootMenu).Error; err != nil {
			// 如果不存在，则创建根菜单
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rootMenu = system.Menu{
					CoreModels: models.CoreModels{
						Id:        rootMenuID, // 直接指定ID为 "root"
						Timestamp: generateTimestamp(),
						Remark:    "根菜单",
					},
					Status:        true,
					Name:          "Root",
					Path:          "root",
					Component:     "root",
					Title:         "root",
					Icon:          "",
					ShowBadge:     false,
					ShowTextBadge: "",
					IsHide:        false,
					IsHideTab:     false,
					Link:          "",
					IsIframe:      false,
					KeepAlive:     false,
					IsFirstLevel:  false,
					FixedTab:      false,
					ActivePath:    "",
					IsFullPage:    false,
					IsAuthButton:  false,
					AuthMark:      "",
					ParentID:      nil,
				}
				rootMenu.Id = "root"
				if err := db.Create(&rootMenu).Error; err != nil {
					// 创建根菜单失败
					return err
				}
			} else {
				// 查询过程中发生未知错误
				return err
			}
		}
		fmt.Println("已创建根菜单")
	}

	return nil
}

// ensureDefaultDept 创建根部门
func ensureDefaultDept(db *gorm.DB) error {
	var count int64
	if err := db.Model(&system.Dept{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查部门表失败: %v", err)
	}

	// 如果没有部门，创建根部门
	if count == 0 {
		// -------------------------- 1. 初始化根部门 (ID = "root") --------------------------
		var rootDept system.Dept
		rootDeptID := "root" // 根部门的固定ID
		// 先查询ID为 "root" 的部门是否存在
		if err := db.Where("id = ?", rootDeptID).First(&rootDept).Error; err != nil {
			// 如果不存在，则创建根部门
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rootDept = system.Dept{
					CoreModels: models.CoreModels{
						Id:        rootDeptID, // 直接指定ID为 "root"
						Timestamp: generateTimestamp(),
						Remark:    "根部门",
					},
					Status:      true,
					Name:        "根部门",  // 根部门的名称
					Code:        "ROOT", // 根部门的编码
					DeptType:    dept.TypeOther,
					Owner:       "",
					Phone:       "",
					Email:       "",
					Description: "",
					ParentID:    nil,
					Parent:      nil,
					Level:       0,   // 根节点层级为0
					Path:        "/", // 根部门的路径
				}
				rootDept.Id = "root"
				if err := db.Create(&rootDept).Error; err != nil {
					// 创建根部门失败
					return err
				}
			} else {
				// 查询过程中发生未知错误
				return err
			}
		}
		// -------------------------- 2. 初始化默认公司部门 --------------------------
		var defaultCompany system.Dept
		defaultCompanyCode := "CAREFUL-COMPANY" // 默认公司的编码
		// 先查询编码为 "CAREFUL-COMPANY" 的部门是否存在
		if err := db.Where("code = ?", defaultCompanyCode).First(&defaultCompany).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建默认公司部门，并将其父部门ID设为 "root"
				defaultCompany = system.Dept{
					CoreModels: models.CoreModels{
						Id:        generateId(),
						Timestamp: generateTimestamp(),
					},
					Status:      true,
					Name:        "用心集团有限公司", // 默认公司名称
					Code:        defaultCompanyCode,
					DeptType:    dept.TypeCompany,
					Owner:       "",
					Phone:       "",
					Email:       "",
					Description: "",
					ParentID:    &rootDeptID, // 父部门ID指向根部门的ID "root"
					Parent:      nil,
					Level:       1, // 层级为1，表示是根部门的子部门
					Path:        "/",
					// Path:        "/" + defaultCompany.Id + "/", // 路径由父部门路径和自身ID组成
				}
				// 路径由父部门路径和自身ID组成
				// defaultCompany.Path = "/" + defaultCompany.Id + "/"
				if err := db.Create(&defaultCompany).Error; err != nil {
					// 创建默认公司部门失败
					return err
				}
			} else {
				// 查询过程中发生未知错误
				return err
			}
		}
		fmt.Println("已创建根部门和默认部门")
	}

	return nil
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

	// 获取可用部门列表
	var deptList []system.Dept
	if err := db.Where("id != ?", "root").
		Where("status = ?", true).
		Find(&deptList).Error; err != nil {
		return fmt.Errorf("获取部门列表失败: %v", err)
	}
	fmt.Println("\n可用部门:")
	fmt.Println("0. 不分配部门")
	for i, d := range deptList {
		fmt.Printf("%d. %s (%s)\n", i+1, d.Name, d.Code)
	}

	// 设置创建者和部门
	// 选择部门
	fmt.Printf("请选择部门 [0-%d] (0表示不分配部门): ", len(deptList))
	deptInput, _ := reader.ReadString('\n')
	deptInput = strings.TrimSpace(deptInput)

	if deptInput == "" || deptInput == "0" {
		user.DeptID = nil // 不分配部门
		fmt.Println("用户将不分配到任何部门")
	} else {
		// 解析用户输入
		var deptIndex int
		if _, err := fmt.Sscanf(deptInput, "%d", &deptIndex); err != nil || deptIndex < 1 || deptIndex > len(deptList) {
			return fmt.Errorf("无效的部门选择")
		}
		selectedDept := &deptList[deptIndex-1]
		user.DeptID = &selectedDept.Id
		fmt.Printf("用户将分配到部门: %s\n", selectedDept.Name)
	}

	// 验证数据 (如果 Validate 方法已实现)
	if err := user.Validate(); err != nil {
		return err
	}

	user.Id = generateId()
	user.Timestamp = generateTimestamp()
	user.IsSuperuser = true

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
	if user.DeptID != nil {
		var d system.Dept
		if err := db.Where("id = ?", user.DeptID).First(&d).Error; err == nil {
			fmt.Printf("所属部门: %s\n", d.Name)
		}
	} else {
		fmt.Println("未分配部门")
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
