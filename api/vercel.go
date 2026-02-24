package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/taalas/chatjimmy2api/pkg/config"
	"github.com/taalas/chatjimmy2api/pkg/client"
	"github.com/taalas/chatjimmy2api/pkg/handler"
	"github.com/taalas/chatjimmy2api/pkg/logger"
	"github.com/taalas/chatjimmy2api/pkg/metrics"
)

var router *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	router.Use(gin.Recovery())

	// 获取配置 - Vercel 环境下使用 /var/task 目录
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 检查是否在 Vercel 环境
		if os.Getenv("VERCEL") == "1" {
			configPath = "/var/task/conf/config.json"
		} else {
			configPath = "config/config.json"
		}
	}
	cfgMgr, err := config.NewManager(configPath)
	if err != nil {
		// 如果配置加载失败，使用默认配置
		cfgMgr = config.NewDefaultManager()
	}
	cfg := cfgMgr.Get()

	// 初始化日志 - Vercel 环境使用 /tmp 目录
	logCfg := logger.DefaultConfig()
	logCfg.FilePath = "/tmp/logs/server.log"
	log, err := logger.New(logCfg)
	if err != nil {
		// 日志初始化失败，使用默认日志（输出到 stdout）
		logCfg.FilePath = ""
		log, _ = logger.New(logCfg)
	}

	// 初始化统计 - Vercel 环境使用 /tmp 目录
	metricsMgr := metrics.NewManager("/tmp/data/stats.json", cfg.StatsFlushIntervalSec)

	// 初始化上游客户端
	upstreamClient := client.NewChatJimmyClient(
		cfg.UpstreamBaseURL,
		cfg.UpstreamAPIKey,
		cfg.UpstreamTimeoutMs,
		cfg.UpstreamMaxRetries,
	)

	// 注册 API 路由
	apiHandler := handler.NewAPIHandler(cfgMgr, metricsMgr, log, upstreamClient)
	apiHandler.RegisterRoutes(router)

	// 注册管理界面路由（Vercel Serverless 模式）
	adminHandler := handler.NewAdminHandler(cfgMgr, metricsMgr, log)
	adminHandler.RegisterWebRoutes(router)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
