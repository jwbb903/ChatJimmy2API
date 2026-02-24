// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/config"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/client"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/handler"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/logger"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/metrics"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config/config.json", "配置文件路径")
	flag.Parse()

	// 确保配置目录存在
	dir := filepath.Dir(*configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("创建配置目录失败：%v\n", err)
		os.Exit(1)
	}

	// 初始化配置管理器
	cfgMgr, err := config.NewManager(*configPath)
	if err != nil {
		fmt.Printf("初始化配置管理器失败：%v\n", err)
		os.Exit(1)
	}
	defer cfgMgr.Close()

	// 保存初始配置（如果不存在）
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		if err := cfgMgr.Save(); err != nil {
			fmt.Printf("保存初始配置失败：%v\n", err)
		} else {
			fmt.Printf("已创建默认配置文件：%s\n", *configPath)
		}
	}

	cfg := cfgMgr.Get()

	// 初始化日志记录器
	logCfg := logger.DefaultConfig()
	logCfg.FilePath = cfg.LogFile
	log, err := logger.New(logCfg)
	if err != nil {
		fmt.Printf("初始化日志记录器失败：%v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("服务器启动", map[string]interface{}{
		"version": "1.0.0",
		"config":  *configPath,
	})

	// 初始化统计管理器
	statsMgr := metrics.NewManager(cfg.StatsFile, cfg.StatsFlushIntervalSec)
	defer statsMgr.Close()

	// 初始化上游客户端
	upstreamClient := client.NewChatJimmyClient(
		cfg.UpstreamBaseURL,
		cfg.UpstreamAPIKey,
		cfg.UpstreamTimeoutMs,
		cfg.UpstreamMaxRetries,
	)

	// 配置变更时更新客户端
	cfgMgr.OnChange(func(newCfg *config.Config) {
		upstreamClient.UpdateConfig(
			newCfg.UpstreamBaseURL,
			newCfg.UpstreamAPIKey,
			newCfg.UpstreamTimeoutMs,
			newCfg.UpstreamMaxRetries,
		)
		log.Info("配置已更新", map[string]interface{}{
			"upstream_base_url": newCfg.UpstreamBaseURL,
			"stream_mode":       newCfg.StreamMode,
		})
	})

	// 创建 API 处理器
	apiHandler := handler.NewAPIHandler(cfgMgr, statsMgr, log, upstreamClient)

	// 创建主 API 路由（OpenAI 兼容端点）
	gin.SetMode(gin.ReleaseMode)
	apiRouter := gin.New()
	apiRouter.Use(gin.Recovery())
	apiRouter.Use(func(c *gin.Context) {
		// 请求日志
		log.Info("请求", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		})
		c.Next()
	})
	apiHandler.RegisterRoutes(apiRouter)

	// 启动主 API 服务器
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		log.Info("启动 API 服务器", map[string]interface{}{
			"address": addr,
		})
		fmt.Printf("API 服务器监听于：%s\n", addr)
		if err := apiRouter.Run(addr); err != nil {
			log.Error("API 服务器失败", map[string]interface{}{"error": err.Error()})
			fmt.Printf("API 服务器失败：%v\n", err)
			os.Exit(1)
		}
	}()

	// 启动管理界面服务器（如果启用）
	if cfg.AdminEnabled {
		adminRouter := gin.New()
		adminRouter.Use(gin.Recovery())

		// 创建管理处理器
		adminHandler := handler.NewAdminHandler(cfgMgr, statsMgr, log)
		adminHandler.RegisterRoutes(adminRouter)

		go func() {
			addr := fmt.Sprintf("%s:%d", cfg.AdminHost, cfg.AdminPort)
			log.Info("启动管理界面", map[string]interface{}{
				"address": addr,
			})
			fmt.Printf("管理界面监听于：%s\n", addr)
			if err := adminRouter.Run(addr); err != nil {
				log.Error("管理界面失败", map[string]interface{}{"error": err.Error()})
				fmt.Printf("管理界面失败：%v\n", err)
			}
		}()

		// 在主 API 路由器上也注册管理界面路由
		adminHandler.RegisterWebRoutes(apiRouter)
	}

	// 等待退出
	fmt.Println("服务器已启动，按 Ctrl+C 退出")
	select {}
}
