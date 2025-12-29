/**
 * Description：
 * FileName：server.go
 * Author：CJiaの用心
 * Create：2025/11/23 01:25:13
 * Remark：
 */

package ioc

import (
	"context"
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/internal/web/middleware"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	rely         config.RelyConfig
	locale       string
	staticPath   string
	Translator   ut.Translator
	routerEngine *gin.Engine
}

func NewServer(rely config.RelyConfig, locale string) *Server {
	srv := &Server{
		rely:   rely,
		locale: locale,
	}

	// 初始化静态资源路径
	staticDir := "./static"
	if absPath, err := filepath.Abs(staticDir); err == nil {
		if _, err := os.Stat(absPath); err == nil {
			srv.staticPath = absPath
		}
	}

	return srv
}

func (s *Server) InitTranslator() error {
	// 注册自定义tag名称函数
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("获取验证引擎失败")
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// 初始化翻译器
	enT := en.New()
	zhT := zh.New()
	uni := ut.New(enT, zhT, enT) // 默认英语，支持中文

	trans, ok := uni.GetTranslator(s.locale)
	if !ok {
		return fmt.Errorf("不支持的语种: %s", s.locale)
	}

	// 注册翻译器
	switch s.locale {
	case "zh":
		if err := zhtranslations.RegisterDefaultTranslations(v, trans); err != nil {
			return err
		}
	default: // 包括en和其他语言都默认使用英语
		if err := entranslations.RegisterDefaultTranslations(v, trans); err != nil {
			return err
		}
	}

	s.Translator = trans
	return nil
}

func (s *Server) StaticPath() string {
	// 如果路径无效则优雅降级
	if s.staticPath == "" {
		zap.L().Warn("静态资源路径无效，使用默认路径")
		if path, err := os.Getwd(); err == nil {
			return filepath.Join(path, "static")
		}
		return "./static"
	}
	return s.staticPath
}

func (s *Server) InitGinMiddlewares(rely config.RelyConfig) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.NewCorsMiddlewareBuilder().Build(), // 跨域支持
		middleware.NewLoginJWTMiddlewareBuilder(rely).
			IgnorePaths("/health").
			IgnorePaths("/api/v1/auth/login").
			IgnorePaths("/api/v1/auth/refresh-token").
			Build(), // 认证中间件
		// middleware.NewLogger(rely.Logger).Build(), // 请求日志
		// middleware.NewStorage(rely).Build(),       // 本地化日志
	}
}

func (s *Server) InitWebServer(middlewares []gin.HandlerFunc, debug bool) *gin.Engine {
	if debug {
		gin.SetMode(gin.DebugMode) // 开发模式
	} else {
		gin.SetMode(gin.ReleaseMode) // 默认生产模式
	}

	engine := gin.Default()
	engine.Use(middlewares...)

	// 注册静态资源路由
	if s.staticPath != "" {
		engine.Static("/static", s.staticPath)
		// zap.L().Info("静态资源已注册", zap.String("path", s.staticPath))
	}

	// 添加 favicon.ico 的路由处理
	engine.StaticFile("/favicon.svg", "./favicon.svg")

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "CarefulAdmin Services",
		})
	})

	s.routerEngine = engine
	return engine
}

// Run 优雅启动应用
func (s *Server) Run(host string, port int) error {
	if s.routerEngine == nil {
		return fmt.Errorf("路由引擎未初始化")
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: s.routerEngine,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// 定义颜色打印器
		greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
		yellowBold := color.New(color.FgYellow, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		magenta := color.New(color.FgMagenta).SprintFunc()

		// 输出go-admin的ASCII艺术字（可替换为你喜欢的样式）
		goAdminArt := `
   _______      ____    __
  / ____/ |    / __ \  / /
 / /    | |   / /_/ / / / 
/ /___  | |  / ____/ / /  
\____/  |_| /_/     /_/   
                          
`
		fmt.Println(magenta(goAdminArt))

		// 欢迎信息
		fmt.Printf("欢迎使用 %s %s\n", greenBold("go-admin"), yellowBold("v1.0.0"))
		fmt.Println("") // 空行分隔

		// 服务启动日志（原zap日志保留）
		zap.L().Info("服务启动中",
			zap.String("host", host),
			zap.Int("port", port))
		fmt.Println("")

		// 格式化输出服务地址（彩色区分）
		fullAddr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("%s\n", cyan("Server run at:"))
		fmt.Printf("  - Local:   http://localhost:%d\n", port)
		fmt.Printf("%s\n", cyan("Swagger run at:"))
		fmt.Printf("  - Local:   http%s%s/swagger/index.html\n", "://", fullAddr)
		fmt.Printf("%s\n", cyan("Health run at:"))
		fmt.Printf("  - Local:   http%s%s/health\n", "://", fullAddr)
		fmt.Println("")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("服务启动失败", zap.Error(err))
		}
	}()

	<-quit
	zap.L().Info("服务正在关闭...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("强制关闭服务: %w", err)
	}

	zap.L().Info("服务已停止")
	return nil
}
