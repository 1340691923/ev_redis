package router

import (
	"ev-plugin/backend/api"
	"ev-plugin/backend/response"

	"github.com/1340691923/eve-plugin-sdk-go/backend/web_engine"
)

type WebServer struct {
	engine          *web_engine.WebEngine
	redisController *api.RedisController
}

// 依赖注入
func NewWebServer(app *web_engine.WebEngine) *WebServer {
	baseController := api.NewBaseController(response.NewResponse())
	redisController := api.NewRedisController(baseController)
	return &WebServer{
		engine:          app,
		redisController: redisController,
	}
}

func NewRouter(engine *web_engine.WebEngine) {

	//后端api
	webSvr := NewWebServer(engine)
	group := webSvr.engine.Group("redis管理", "/api")

	group.POST(false, "获取redis的key列表", "/RedisKeys", webSvr.redisController.GetAllKeysAction)
	group.POST(false, "获取redis信息概览", "/RedisInfoOverview", webSvr.redisController.GetInfoOverviewAction)
	group.POST(false, "获取redis数据库列表", "/RedisDatabases", webSvr.redisController.GetDatabasesAction)
	group.POST(false, "获取redis内存分析", "/RedisMemoryAnalysis", webSvr.redisController.GetMemoryAnalysisAction)
	group.POST(false, "搜索redis key", "/RedisSearchKeys", webSvr.redisController.SearchKeysAction)
	group.POST(false, "获取redis key详情", "/RedisKeyDetail", webSvr.redisController.GetKeyDetailAction)
	group.POST(true, "删除redis key", "/RedisDeleteKey", webSvr.redisController.DeleteKeyAction)
	group.POST(true, "设置redis key", "/RedisSetKey", webSvr.redisController.SetKeyAction)
	group.POST(false, "批量获取keys内存分析", "/RedisBatchMemoryAnalysis", webSvr.redisController.BatchGetMemoryAnalysisAction)

}
