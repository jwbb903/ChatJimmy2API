package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/config"
	pkgclient "github.com/jwbb903/ChatJimmy2API/api/_internal/client"
	pkgHandler "github.com/jwbb903/ChatJimmy2API/api/_internal/handler"
	pkglogger "github.com/jwbb903/ChatJimmy2API/api/_internal/logger"
	pkgmetrics "github.com/jwbb903/ChatJimmy2API/api/_internal/metrics"
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
	logCfg := pkglogger.DefaultConfig()
	logCfg.FilePath = "/tmp/logs/server.log"
	log, err := pkglogger.New(logCfg)
	if err != nil {
		// 日志初始化失败，使用默认日志（输出到 stdout）
		logCfg.FilePath = ""
		log, _ = pkglogger.New(logCfg)
	}

	// 初始化统计 - Vercel 环境使用 /tmp 目录
	metricsMgr := pkgmetrics.NewManager("/tmp/data/stats.json", cfg.StatsFlushIntervalSec)

	// 初始化上游客户端
	upstreamClient := pkgclient.NewChatJimmyClient(
		cfg.UpstreamBaseURL,
		cfg.UpstreamAPIKey,
		cfg.UpstreamTimeoutMs,
		cfg.UpstreamMaxRetries,
	)

	// 注册 API 路由
	apiHandler := pkgHandler.NewAPIHandler(cfgMgr, metricsMgr, log, upstreamClient)
	apiHandler.RegisterRoutes(router)

	// 注册管理界面路由（Vercel Serverless 模式）
	adminHandler := pkgHandler.NewAdminHandler(cfgMgr, metricsMgr, log)
	adminHandler.RegisterWebRoutes(router)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
