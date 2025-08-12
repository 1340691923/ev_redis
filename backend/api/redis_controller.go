package api

import (
	"context"
	"ev-plugin/backend/dto"
	"ev-plugin/backend/response"
	"ev-plugin/backend/vo"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1340691923/eve-plugin-sdk-go/backend/logger"
	"github.com/1340691923/eve-plugin-sdk-go/ev_api"
	"github.com/1340691923/eve-plugin-sdk-go/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// Redis控制器
type RedisController struct {
	*BaseController
}

func NewRedisController(baseController *BaseController) *RedisController {
	return &RedisController{baseController}
}

// isConnectionRefusedError 检查错误是否为连接被拒绝的错误
func (this *RedisController) isConnectionRefusedError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 检查常见的连接被拒绝错误模式
	return strings.Contains(errStr, "connectex: No connection could be made because the target machine actively refused it") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial tcp") && strings.Contains(errStr, "refused")
}

// executeRedisCommandWithRetry 执行Redis命令并在连接被拒绝时重试
func (this *RedisController) executeRedisCommandWithRetry(ctx context.Context, api *ev_api.EvApiAdapter, database int, command string, args ...interface{}) (interface{}, error) {
	const maxRetries = 5
	const retryDelay = 3 * time.Second

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 构建完整的命令参数
		cmdArgs := make([]interface{}, 0, len(args)+1)
		cmdArgs = append(cmdArgs, command)
		cmdArgs = append(cmdArgs, args...)

		// 执行Redis命令
		result, err := api.RedisExecCommand(ctx, database, cmdArgs...)

		// 如果成功，直接返回结果
		if err == nil {
			if attempt > 0 {
				logger.DefaultLogger.Info("Redis命令重试成功",
					"command:", command,
					"attempt:", attempt+1,
					"total_attempts:", maxRetries)
			}
			return result, nil
		}

		// 检查是否为连接被拒绝的错误
		if !this.isConnectionRefusedError(err) {
			// 如果不是连接被拒绝的错误，直接返回，不重试
			return nil, err
		}

		lastErr = err

		// 如果是最后一次尝试，不需要等待
		if attempt == maxRetries-1 {
			logger.DefaultLogger.Error("Redis命令重试失败，已达到最大重试次数",
				"command:", command,
				"attempts:", maxRetries,
				"error:", err)
			break
		}

		// 记录重试信息
		logger.DefaultLogger.Warn("Redis连接被拒绝，准备重试",
			"command:", command,
			"attempt:", attempt+1,
			"max_attempts:", maxRetries,
			"retry_delay:", retryDelay,
			"error:", err)

		// 等待重试延迟
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
			// 继续下一次重试
		}
	}

	return nil, fmt.Errorf("Redis命令执行失败，已重试%d次: %w", maxRetries, lastErr)
}

// GetAllKeysAction 获取Redis所有key
func (this *RedisController) GetAllKeysAction(ctx *gin.Context) {
	req := new(dto.RedisKeysRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("查询Redis所有Keys", "conn_id:", req.EsConnect, "database:", req.Database)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 执行Redis KEYS * 命令获取所有key
	result, err := api.RedisExecCommand(ctx, req.Database, "KEYS", "*")
	if err != nil {
		logger.DefaultLogger.Error("查询Redis Keys失败", "error:", err)
		this.Error(ctx, err)
		return
	}

	// 将结果转换为字符串切片
	var keys []string
	if result != nil {
		// Redis KEYS命令返回的是字符串切片
		for _, key := range cast.ToSlice(result) {
			keyStr := cast.ToString(key)
			if keyStr != "" {
				keys = append(keys, keyStr)
			}

		}
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisKeysResponse{
		Keys: keys,
	})
}

// GetInfoOverviewAction 获取Redis信息总览
func (this *RedisController) GetInfoOverviewAction(ctx *gin.Context) {
	req := new(dto.RedisInfoRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("查询Redis信息总览", "conn_id:", req.EsConnect, "database:", req.Database)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 获取Redis INFO信息
	infoResult, err := api.RedisExecCommand(ctx, req.Database, "INFO")
	if err != nil {
		logger.DefaultLogger.Error("获取Redis INFO失败", "error:", err)
		this.Error(ctx, err)
		return
	}

	// 获取Redis INFO keyspace信息
	keyspaceResult, err := api.RedisExecCommand(ctx, req.Database, "INFO", "keyspace")
	if err != nil {
		logger.DefaultLogger.Error("获取Redis INFO keyspace失败", "error:", err)
		this.Error(ctx, err)
		return
	}

	// 解析INFO信息
	infoMap := make(map[string]string)
	if infoResult != nil {
		infoStr := cast.ToString(infoResult)
		if infoStr != "" {
			lines := strings.Split(infoStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "#") || line == "" {
					continue
				}
				kv := strings.SplitN(line, ":", 2)
				if len(kv) == 2 {
					infoMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}
	}

	// 解析keyspace信息
	var keyspace []map[string]interface{}
	if keyspaceResult != nil {
		keyspaceStr := cast.ToString(keyspaceResult)
		if keyspaceStr != "" {
			lines := strings.Split(keyspaceStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "db") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						db := parts[0]
						stats := strings.Split(parts[1], ",")
						m := map[string]interface{}{"db": db}
						for _, stat := range stats {
							kv := strings.SplitN(stat, "=", 2)
							if len(kv) == 2 {
								m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
							}
						}
						keyspace = append(keyspace, m)
					}
				}
			}
		}
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisInfoResponse{
		Info:     infoMap,
		Keyspace: keyspace,
	})
}

// GetDatabasesAction 获取Redis数据库列表
func (this *RedisController) GetDatabasesAction(ctx *gin.Context) {
	req := new(dto.RedisDatabasesRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("获取Redis数据库列表", "conn_id:", req.EsConnect)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 获取Redis INFO keyspace信息
	keyspaceResult, err := api.RedisExecCommand(ctx, 0, "INFO", "keyspace")
	if err != nil {
		logger.DefaultLogger.Error("获取Redis INFO keyspace失败", "error:", err)
		this.Error(ctx, err)
		return
	}

	var databases []vo.RedisDatabaseInfo

	// 解析keyspace信息
	if keyspaceResult != nil {
		keyspaceStr := cast.ToString(keyspaceResult)
		if keyspaceStr != "" {
			lines := strings.Split(keyspaceStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "db") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						// 解析数据库索引
						dbStr := parts[0]
						dbIndex := 0
						if dbNum := strings.TrimPrefix(dbStr, "db"); dbNum != dbStr {
							dbIndex = cast.ToInt(dbNum)
						}

						// 解析统计信息
						stats := strings.Split(parts[1], ",")
						dbInfo := vo.RedisDatabaseInfo{
							Database: dbIndex,
							Keys:     0,
							Expires:  0,
							AvgTTL:   0,
						}

						for _, stat := range stats {
							kv := strings.SplitN(stat, "=", 2)
							if len(kv) == 2 {
								key := strings.TrimSpace(kv[0])
								value := strings.TrimSpace(kv[1])

								switch key {
								case "keys":
									dbInfo.Keys = cast.ToInt64(value)
								case "expires":
									dbInfo.Expires = cast.ToInt64(value)
								case "avg_ttl":
									dbInfo.AvgTTL = cast.ToInt64(value)
								}
							}
						}

						databases = append(databases, dbInfo)
					}
				}
			}
		}
	}

	// 如果没有找到任何数据库信息，至少返回db0
	if len(databases) == 0 {
		databases = append(databases, vo.RedisDatabaseInfo{
			Database: 0,
			Keys:     0,
			Expires:  0,
			AvgTTL:   0,
		})
	}

	logger.DefaultLogger.Debug("获取数据库列表完成", "count:", len(databases))

	this.Success(ctx, response.SearchSuccess, vo.RedisDatabasesResponse{
		Databases: databases,
	})
}

// GetMemoryAnalysisAction 获取Redis内存分析 - 支持分页的SCAN扫描
func (this *RedisController) GetMemoryAnalysisAction(ctx *gin.Context) {
	req := new(dto.RedisMemoryAnalysisRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	// 设置默认值
	if req.Pattern == "" {
		req.Pattern = "*"
	}
	if req.Count <= 0 {
		req.Count = 50 // 默认返回50条
	}
	if req.Cursor == "" {
		req.Cursor = "0"
	}

	logger.DefaultLogger.Debug("开始Redis内存分析（返回所有匹配的Keys） - 将执行MEMORY USAGE命令",
		"conn_id:", req.EsConnect,
		"database:", req.Database,
		"pattern:", req.Pattern)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 获取数据库中的总Key数量（用于前端显示）
	totalCount := 0
	if dbSizeResult, err := this.executeRedisCommandWithRetry(ctx, api, req.Database, "DBSIZE"); err == nil && dbSizeResult != nil {
		totalCount = cast.ToInt(dbSizeResult)
	}

	// 扫描所有匹配的keys
	var allKeys []string
	cursor := "0"

	logger.DefaultLogger.Debug("开始扫描所有匹配的Keys")

	// 循环扫描所有匹配的keys
	for {
		scanResult, err := this.executeRedisCommandWithRetry(ctx, api, req.Database, "SCAN", cursor, "MATCH", req.Pattern, "COUNT", "1000")
		if err != nil {
			logger.DefaultLogger.Error("SCAN命令执行失败", "error:", err)
			this.Error(ctx, err)
			return
		}

		scanArray := cast.ToSlice(scanResult)
		if len(scanArray) != 2 {
			logger.DefaultLogger.Error("SCAN结果格式错误", "result:", scanResult)
			break
		}

		// 获取下一个cursor
		cursor = cast.ToString(scanArray[0])

		// 获取keys数组
		if keysData := scanArray[1]; keysData != nil {
			keysList := cast.ToSlice(keysData)
			for _, key := range keysList {
				keyStr := cast.ToString(key)
				if keyStr != "" {
					allKeys = append(allKeys, keyStr)
				}
			}
		}

		// cursor为"0"表示扫描完成
		if cursor == "0" {
			break
		}
	}

	logger.DefaultLogger.Debug("扫描完成",
		"总匹配Keys:", len(allKeys))

	// 分析每个key的内存使用情况 - 使用并发处理
	var keyMemoryInfos []vo.RedisKeyMemoryInfo
	var totalSize int64 = 0
	var mu sync.Mutex // 保护共享数据

	if len(allKeys) > 0 {
		// 创建带有并发限制的errgroup
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(2000) // 控制并发数为10，避免对Redis造成过大压力

		// 为每个key创建一个goroutine
		for _, key := range allKeys {
			key := key // 避免闭包问题
			g.Go(func() error {
				// 分析单个key的内存信息
				keyInfo, err := this.analyzeKeyMemory(gctx, api, req.Database, key)
				if err != nil {
					logger.DefaultLogger.Error("分析Key内存失败", "key:", key, "error:", err)
					// 不返回错误，继续处理其他key
					return nil
				}

				// 线程安全地添加结果
				mu.Lock()
				keyMemoryInfos = append(keyMemoryInfos, keyInfo)
				totalSize += keyInfo.SizeBytes
				mu.Unlock()

				return nil
			})
		}

		// 等待所有goroutine完成
		if err := g.Wait(); err != nil {
			logger.DefaultLogger.Error("并发分析内存失败", "error:", err)
			// 不直接返回错误，允许返回部分结果
		}
	}

	logger.DefaultLogger.Debug("内存分析完成 - 已执行MEMORY USAGE命令",
		"处理Keys:", len(keyMemoryInfos),
		"总大小:", totalSize)

	this.Success(ctx, response.SearchSuccess, vo.RedisMemoryAnalysisResponse{
		Keys:       keyMemoryInfos,
		TotalKeys:  len(keyMemoryInfos),
		TotalSize:  totalSize,
		NextCursor: "0", // 前端分页不需要cursor
		TotalCount: totalCount,
	})
}

// analyzeKeyMemory 分析单个key的内存使用情况
func (this *RedisController) analyzeKeyMemory(ctx context.Context, api *ev_api.EvApiAdapter, database int, key string) (vo.RedisKeyMemoryInfo, error) {
	logger.DefaultLogger.Debug("开始分析Key", "key:", key)

	// 获取key的类型 - 带重试机制
	keyType := ""
	var typeResult interface{}
	var err error

	// 使用重试机制获取key类型
	if typeResult, err = this.executeRedisCommandWithRetry(ctx, api, database, "TYPE", key); err == nil && typeResult != nil {
		keyType = cast.ToString(typeResult)
		if keyType == "" || keyType == "none" {
			keyType = "unknown"
		}
	} else {
		logger.DefaultLogger.Warn("TYPE命令执行失败", "key:", key, "error:", err)
		keyType = "unknown"
	}

	if keyType == "" || keyType == "none" {
		logger.DefaultLogger.Warn("TYPE命令最终失败", "key:", key, "error:", err, "result:", typeResult)
		keyType = "unknown"
	}

	// 获取key的内存使用大小 - 优先使用MEMORY USAGE
	var sizeBytes int64 = 0
	var useMemoryUsage = false

	// 尝试MEMORY USAGE命令
	logger.DefaultLogger.Debug("执行MEMORY USAGE命令", "key:", key)
	memoryResult, err := this.executeRedisCommandWithRetry(ctx, api, database, "MEMORY", "USAGE", key)

	if err != nil {
		logger.DefaultLogger.Debug("MEMORY USAGE命令执行失败", "key:", key, "error:", err)
	} else if memoryResult == nil {
		logger.DefaultLogger.Debug("MEMORY USAGE命令返回nil", "key:", key)
	} else {
		// 使用cast库进行类型转换
		if size, err := cast.ToInt64E(memoryResult); err == nil && size > 0 {
			sizeBytes = size
			useMemoryUsage = true
			logger.DefaultLogger.Debug("MEMORY USAGE结果转换成功", "key:", key, "size:", sizeBytes)
		} else {
			logger.DefaultLogger.Debug("MEMORY USAGE结果转换失败", "key:", key, "value:", memoryResult, "error:", err)
		}
	}

	// 如果MEMORY USAGE失败，使用备用估算
	if !useMemoryUsage {

		switch keyType {
		case "string":
			if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "STRLEN", key); err == nil && result != nil {
				if strlen, err := cast.ToInt64E(result); err == nil {
					sizeBytes = strlen + 50 // 字符串内容 + 元数据开销
				}
			}
		case "hash":
			if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "HLEN", key); err == nil && result != nil {
				if hlen, err := cast.ToInt64E(result); err == nil {
					sizeBytes = hlen * 100 // 估算每个字段平均100字节
				}
			}
		case "list":
			if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "LLEN", key); err == nil && result != nil {
				if llen, err := cast.ToInt64E(result); err == nil {
					sizeBytes = llen * 50 // 估算每个元素平均50字节
				}
			}
		case "set":
			if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "SCARD", key); err == nil && result != nil {
				if scard, err := cast.ToInt64E(result); err == nil {
					sizeBytes = scard * 50 // 估算每个元素平均50字节
				}
			}
		case "zset":
			if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "ZCARD", key); err == nil && result != nil {
				if zcard, err := cast.ToInt64E(result); err == nil {
					sizeBytes = zcard * 60 // 估算每个元素平均60字节（包含score）
				}
			}
		default:
			sizeBytes = 100 // 默认估算100字节
		}

		if sizeBytes == 0 {
			sizeBytes = 50 // 最小估算50字节
		}

		logger.DefaultLogger.Debug("估算大小", "key:", key, "type:", keyType, "estimatedSize:", sizeBytes)
	}

	// 不在列表中获取TTL，节省资源
	keyInfo := vo.RedisKeyMemoryInfo{
		Key:       key,
		SizeBytes: sizeBytes,
		Type:      keyType,
		TTL:       -1, // 列表中不显示TTL，只在详情中获取
	}

	return keyInfo, nil
}

// analyzeKeyMemoryFast 快速分析单个key的内存使用情况 - 直接使用估算，不使用MEMORY USAGE
func (this *RedisController) analyzeKeyMemoryFast(ctx context.Context, api *ev_api.EvApiAdapter, database int, key string) (vo.RedisKeyMemoryInfo, error) {
	// 获取key的类型
	typeResult, err := this.executeRedisCommandWithRetry(ctx, api, database, "TYPE", key)
	if err != nil {
		return vo.RedisKeyMemoryInfo{}, err
	}

	keyType := "unknown"
	if typeResult != nil {
		keyType = cast.ToString(typeResult)
		if keyType == "none" {
			keyType = "unknown"
		}
	}

	// 直接使用估算方法计算大小，不使用MEMORY USAGE命令
	var sizeBytes int64 = 50 // 默认最小50字节

	switch keyType {
	case "string":
		if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "STRLEN", key); err == nil && result != nil {
			if strlen, err := cast.ToInt64E(result); err == nil {
				sizeBytes = strlen + 50 // 字符串内容 + 元数据开销
			}
		}
	case "hash":
		if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "HLEN", key); err == nil && result != nil {
			if hlen, err := cast.ToInt64E(result); err == nil {
				sizeBytes = hlen * 100 // 估算每个字段平均100字节
			}
		}
	case "list":
		if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "LLEN", key); err == nil && result != nil {
			if llen, err := cast.ToInt64E(result); err == nil {
				sizeBytes = llen * 50 // 估算每个元素平均50字节
			}
		}
	case "set":
		if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "SCARD", key); err == nil && result != nil {
			if scard, err := cast.ToInt64E(result); err == nil {
				sizeBytes = scard * 50 // 估算每个元素平均50字节
			}
		}
	case "zset":
		if result, err := this.executeRedisCommandWithRetry(ctx, api, database, "ZCARD", key); err == nil && result != nil {
			if zcard, err := cast.ToInt64E(result); err == nil {
				sizeBytes = zcard * 60 // 估算每个元素平均60字节（包含score）
			}
		}
	default:
		sizeBytes = 100 // 默认估算100字节
	}

	// 确保最小值
	if sizeBytes < 50 {
		sizeBytes = 50
	}

	keyInfo := vo.RedisKeyMemoryInfo{
		Key:       key,
		SizeBytes: sizeBytes,
		Type:      keyType,
		TTL:       -1, // 列表中不显示TTL，只在详情中获取
	}

	return keyInfo, nil
}

// analyzeKeyMemoryOfficial 使用官方MEMORY USAGE API分析单个key的内存使用情况
func (this *RedisController) analyzeKeyMemoryOfficial(ctx context.Context, api *ev_api.EvApiAdapter, database int, key string) (vo.RedisKeyMemoryInfo, error) {
	// 获取key的类型
	typeResult, err := this.executeRedisCommandWithRetry(ctx, api, database, "TYPE", key)
	if err != nil {
		return vo.RedisKeyMemoryInfo{}, fmt.Errorf("获取Key类型失败: %w", err)
	}

	keyType := "unknown"
	if typeResult != nil {
		keyType = cast.ToString(typeResult)
		if keyType == "none" {
			return vo.RedisKeyMemoryInfo{}, fmt.Errorf("Key不存在")
		}
	}

	// 使用官方MEMORY USAGE命令获取精确内存使用
	memoryResult, err := this.executeRedisCommandWithRetry(ctx, api, database, "MEMORY", "USAGE", key)
	if err != nil {
		// 检查是否是版本不支持的错误
		if strings.Contains(err.Error(), "unknown command") ||
			strings.Contains(err.Error(), "ERR unknown command") ||
			strings.Contains(err.Error(), "MEMORY") {
			return vo.RedisKeyMemoryInfo{}, fmt.Errorf("Redis版本不支持MEMORY USAGE命令，请使用Redis 4.0+版本")
		}
		return vo.RedisKeyMemoryInfo{}, fmt.Errorf("执行MEMORY USAGE命令失败: %w", err)
	}

	if memoryResult == nil {
		return vo.RedisKeyMemoryInfo{}, fmt.Errorf("MEMORY USAGE命令返回空结果")
	}

	// 转换内存大小
	sizeBytes, err := cast.ToInt64E(memoryResult)
	if err != nil {
		return vo.RedisKeyMemoryInfo{}, fmt.Errorf("解析内存大小失败: %w", err)
	}

	if sizeBytes <= 0 {
		return vo.RedisKeyMemoryInfo{}, fmt.Errorf("无效的内存大小: %d", sizeBytes)
	}

	keyInfo := vo.RedisKeyMemoryInfo{
		Key:       key,
		SizeBytes: sizeBytes,
		Type:      keyType,
		TTL:       -1, // 列表中不显示TTL，只在详情中获取
	}

	return keyInfo, nil
}

// DeleteKeyAction 删除Redis Key
func (this *RedisController) DeleteKeyAction(ctx *gin.Context) {
	req := new(dto.RedisDeleteKeyRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("删除Redis Key", "conn_id:", req.EsConnect, "database:", req.Database, "key:", req.Key)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 执行删除命令
	result, err := api.RedisExecCommand(ctx, req.Database, "DEL", req.Key)
	if err != nil {
		logger.DefaultLogger.Error("删除Redis Key失败", "key:", req.Key, "error:", err)
		this.Error(ctx, err)
		return
	}

	success := false
	message := "删除失败"

	if result != nil {
		if deletedCount, err := cast.ToInt64E(result); err == nil && deletedCount > 0 {
			success = true
			message = "删除成功"
		}
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisOperationResponse{
		Success: success,
		Message: message,
	})
}

// GetKeyDetailAction 获取Redis Key详情
func (this *RedisController) GetKeyDetailAction(ctx *gin.Context) {
	req := new(dto.RedisKeyDetailRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("获取Redis Key详情", "conn_id:", req.EsConnect, "database:", req.Database, "key:", req.Key)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 获取key的类型
	typeResult, err := api.RedisExecCommand(ctx, req.Database, "TYPE", req.Key)
	if err != nil {
		logger.DefaultLogger.Error("获取Key类型失败", "key:", req.Key, "error:", err)
		this.Error(ctx, err)
		return
	}

	keyType := "unknown"
	if typeResult != nil {
		logger.DefaultLogger.Debug("TYPE命令结果", "key:", req.Key, "result:", typeResult, "type:", fmt.Sprintf("%T", typeResult))

		// 使用cast库进行类型转换
		if typeStr := cast.ToString(typeResult); typeStr != "" && typeStr != "none" {
			keyType = typeStr
			logger.DefaultLogger.Debug("TYPE转换成功", "key:", req.Key, "type:", keyType)
		} else {
			logger.DefaultLogger.Warn("TYPE转换失败或返回none", "key:", req.Key, "result:", typeResult)
		}
	}
	logger.DefaultLogger.Debug("最终Key类型", "key:", req.Key, "type:", keyType)

	// 获取key的内存使用大小
	memoryResult, err := api.RedisExecCommand(ctx, req.Database, "MEMORY", "USAGE", req.Key)
	var sizeBytes int64 = 0
	if err == nil && memoryResult != nil {
		if size, err := cast.ToInt64E(memoryResult); err == nil {
			sizeBytes = size
		}
	}

	// 获取key的TTL
	ttlResult, err := api.RedisExecCommand(ctx, req.Database, "TTL", req.Key)
	var ttl int64 = -1
	if err == nil && ttlResult != nil {
		if ttlValue, err := cast.ToInt64E(ttlResult); err == nil {
			ttl = ttlValue
		}
	}

	// 根据类型获取值
	var value interface{}
	switch keyType {
	case "string":
		value, _ = api.RedisExecCommand(ctx, req.Database, "GET", req.Key)
	case "hash":
		value, _ = api.RedisExecCommand(ctx, req.Database, "HGETALL", req.Key)
	case "list":
		value, _ = api.RedisExecCommand(ctx, req.Database, "LRANGE", req.Key, "0", "-1")
	case "set":
		value, _ = api.RedisExecCommand(ctx, req.Database, "SMEMBERS", req.Key)
	case "zset":
		value, _ = api.RedisExecCommand(ctx, req.Database, "ZRANGE", req.Key, "0", "-1", "WITHSCORES")
	default:
		value = "unsupported type"
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisKeyDetailResponse{
		Key:       req.Key,
		Type:      keyType,
		SizeBytes: sizeBytes,
		TTL:       ttl,
		Value:     value,
	})
}

// SetKeyAction 保存/更新Redis Key
func (this *RedisController) SetKeyAction(ctx *gin.Context) {
	req := new(dto.RedisSetKeyRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("保存Redis Key", "conn_id:", req.EsConnect, "database:", req.Database, "key:", req.Key, "type:", req.Type)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 根据类型保存数据
	switch req.Type {
	case "string":
		err = this.setStringKey(ctx, api, req)
	case "hash":
		err = this.setHashKey(ctx, api, req)
	case "list":
		err = this.setListKey(ctx, api, req)
	case "set":
		err = this.setSetKey(ctx, api, req)
	case "zset":
		err = this.setZSetKey(ctx, api, req)
	default:
		this.Error(ctx, fmt.Errorf("不支持的数据类型: %s", req.Type))
		return
	}

	if err != nil {
		logger.DefaultLogger.Error("保存Redis Key失败", "key:", req.Key, "type:", req.Type, "error:", err)
		this.Error(ctx, err)
		return
	}

	// 设置TTL
	if req.TTL > 0 {
		_, err = api.RedisExecCommand(ctx, req.Database, "EXPIRE", req.Key, strconv.FormatInt(req.TTL, 10))
		if err != nil {
			logger.DefaultLogger.Warn("设置TTL失败", "key:", req.Key, "ttl:", req.TTL, "error:", err)
		}
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisOperationResponse{
		Success: true,
		Message: "保存成功",
	})
}

// setStringKey 保存String类型的Key
func (this *RedisController) setStringKey(ctx *gin.Context, api *ev_api.EvApiAdapter, req *dto.RedisSetKeyRequest) error {
	value := cast.ToString(req.Value)
	_, err := api.RedisExecCommand(ctx, req.Database, "SET", req.Key, value)
	return err
}

// setHashKey 保存Hash类型的Key
func (this *RedisController) setHashKey(ctx *gin.Context, api *ev_api.EvApiAdapter, req *dto.RedisSetKeyRequest) error {
	// 先删除原有的hash
	_, _ = api.RedisExecCommand(ctx, req.Database, "DEL", req.Key)

	// 解析hash数据
	if hashData, ok := req.Value.(map[string]interface{}); ok {
		if len(hashData) == 0 {
			return nil // 空hash不需要创建
		}

		// 构建HMSET命令参数
		args := []interface{}{"HMSET", req.Key}
		for field, value := range hashData {
			args = append(args, field, cast.ToString(value))
		}

		_, err := api.RedisExecCommand(ctx, req.Database, args...)
		return err
	}

	return fmt.Errorf("invalid hash data format")
}

// setListKey 保存List类型的Key
func (this *RedisController) setListKey(ctx *gin.Context, api *ev_api.EvApiAdapter, req *dto.RedisSetKeyRequest) error {
	// 先删除原有的list
	_, _ = api.RedisExecCommand(ctx, req.Database, "DEL", req.Key)

	// 解析list数据
	if listData, ok := req.Value.([]interface{}); ok {
		if len(listData) == 0 {
			return nil // 空list不需要创建
		}

		// 使用RPUSH添加所有元素
		args := []interface{}{"RPUSH", req.Key}
		for _, item := range listData {
			args = append(args, cast.ToString(item))
		}

		_, err := api.RedisExecCommand(ctx, req.Database, args...)
		return err
	}

	return fmt.Errorf("invalid list data format")
}

// setSetKey 保存Set类型的Key
func (this *RedisController) setSetKey(ctx *gin.Context, api *ev_api.EvApiAdapter, req *dto.RedisSetKeyRequest) error {
	// 先删除原有的set
	_, _ = api.RedisExecCommand(ctx, req.Database, "DEL", req.Key)

	// 解析set数据
	if setData, ok := req.Value.([]interface{}); ok {
		if len(setData) == 0 {
			return nil // 空set不需要创建
		}

		// 使用SADD添加所有成员
		args := []interface{}{"SADD", req.Key}
		for _, member := range setData {
			args = append(args, cast.ToString(member))
		}

		_, err := api.RedisExecCommand(ctx, req.Database, args...)
		return err
	}

	return fmt.Errorf("invalid set data format")
}

// setZSetKey 保存ZSet类型的Key
func (this *RedisController) setZSetKey(ctx *gin.Context, api *ev_api.EvApiAdapter, req *dto.RedisSetKeyRequest) error {
	// 先删除原有的zset
	_, _ = api.RedisExecCommand(ctx, req.Database, "DEL", req.Key)

	// 解析zset数据
	if zsetData, ok := req.Value.([]interface{}); ok {
		if len(zsetData) == 0 {
			return nil // 空zset不需要创建
		}

		// 使用ZADD添加所有成员
		args := []interface{}{"ZADD", req.Key}
		for _, item := range zsetData {
			if itemMap, ok := item.(map[string]interface{}); ok {
				score := cast.ToFloat64(itemMap["score"])
				member := cast.ToString(itemMap["member"])
				args = append(args, score, member)
			}
		}

		if len(args) > 2 {
			_, err := api.RedisExecCommand(ctx, req.Database, args...)
			return err
		}
	}

	return fmt.Errorf("invalid zset data format")
}

// SearchKeysAction 搜索Redis Keys - 使用SCAN + strings.Contains方式
func (this *RedisController) SearchKeysAction(ctx *gin.Context) {
	req := new(dto.RedisSearchKeysRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	// 设置默认值
	if req.Count <= 0 {
		req.Count = 50 // 默认返回50条
	}
	if req.Cursor == "" {
		req.Cursor = "0"
	}
	if req.SearchText == "" {
		req.SearchText = "" // 空字符串表示搜索所有
	}

	logger.DefaultLogger.Debug("开始Redis Key搜索（后端搜索） - 不会执行MEMORY USAGE命令",
		"conn_id:", req.EsConnect,
		"database:", req.Database,
		"search_text:", req.SearchText,
		"count:", req.Count,
		"cursor:", req.Cursor,
		"case_sensitive:", req.CaseSensitive)

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 获取数据库中的总Key数量（用于前端显示）
	totalCount := 0
	if req.Cursor == "0" { // 只在第一次请求时获取总数
		if dbSizeResult, err := api.RedisExecCommand(ctx, req.Database, "DBSIZE"); err == nil && dbSizeResult != nil {
			totalCount = cast.ToInt(dbSizeResult)
		}
	}

	// 先扫描全部keys，然后使用strings.Contains进行过滤
	var allKeys []string
	var matchedKeys []string
	cursor := "0"

	// 准备搜索文本
	searchText := req.SearchText
	if !req.CaseSensitive {
		searchText = strings.ToLower(searchText)
	}

	logger.DefaultLogger.Debug("开始扫描全部Keys进行搜索")

	// 循环扫描所有keys
	for {
		scanResult, err := api.RedisExecCommand(ctx, req.Database, "SCAN", cursor, "COUNT", "1000")
		if err != nil {
			logger.DefaultLogger.Error("SCAN命令执行失败", "error:", err)
			this.Error(ctx, err)
			return
		}

		scanArray := cast.ToSlice(scanResult)
		if len(scanArray) != 2 {
			logger.DefaultLogger.Error("SCAN结果格式错误", "result:", scanResult)
			break
		}

		// 获取下一个cursor
		cursor = cast.ToString(scanArray[0])

		// 获取keys数组
		if keysData := scanArray[1]; keysData != nil {
			keysList := cast.ToSlice(keysData)
			for _, key := range keysList {
				keyStr := cast.ToString(key)
				if keyStr != "" {
					allKeys = append(allKeys, keyStr)
				}
			}
		}

		// cursor为"0"表示扫描完成
		if cursor == "0" {
			break
		}
	}

	logger.DefaultLogger.Debug("扫描完成，开始过滤",
		"总Keys数量:", len(allKeys),
		"搜索文本:", req.SearchText)

	// 使用strings.Contains进行过滤
	for _, key := range allKeys {
		keyToCheck := key
		if !req.CaseSensitive {
			keyToCheck = strings.ToLower(key)
		}

		// 如果搜索文本为空，返回所有keys；否则使用contains匹配
		if searchText == "" || strings.Contains(keyToCheck, searchText) {
			matchedKeys = append(matchedKeys, key)
		}
	}

	// 保存筛选后的总数量
	totalMatchedCount := len(matchedKeys)

	logger.DefaultLogger.Debug("过滤完成，准备返回所有匹配的Keys",
		"匹配的Keys数量:", totalMatchedCount)

	// Key搜索不需要分析内存等详细信息，只返回key名称列表
	var keyMemoryInfos []vo.RedisKeyMemoryInfo

	// 将匹配的key名称转换为简单的结构体列表
	for _, key := range matchedKeys {
		keyInfo := vo.RedisKeyMemoryInfo{
			Key:       key,
			SizeBytes: 0,  // 搜索时不获取大小信息
			Type:      "", // 搜索时不获取类型信息
			TTL:       -1, // 搜索时不获取TTL信息
		}
		keyMemoryInfos = append(keyMemoryInfos, keyInfo)
	}

	logger.DefaultLogger.Debug("Redis Key搜索完成 - 未执行任何MEMORY USAGE命令",
		"匹配Keys:", len(keyMemoryInfos))

	this.Success(ctx, response.SearchSuccess, vo.RedisSearchKeysResponse{
		Keys:         keyMemoryInfos,
		TotalKeys:    len(keyMemoryInfos), // 返回的key数量
		TotalSize:    0,                   // 搜索时不统计总大小
		NextCursor:   "0",                 // 前端分页不需要cursor
		TotalCount:   totalCount,          // 数据库中的总key数量
		SearchText:   req.SearchText,
		MatchedCount: totalMatchedCount, // 搜索匹配的总数量
	})
}

// BatchAddKeysAction 批量添加Keys - 100个协程并发添加1,000,000个key
func (this *RedisController) BatchAddKeysAction(ctx *gin.Context) {
	req := new(dto.RedisBatchAddKeysRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("开始批量添加Redis Keys", "conn_id:", req.EsConnect, "database:", req.Database)

	startTime := time.Now()

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	const totalKeys = 100000                            // 总共要添加的key数量
	const goroutineCount = 100                          // 协程数量
	const keysPerGoroutine = totalKeys / goroutineCount // 每个协程处理的key数量

	// 使用带有超时的context
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// 创建带有并发限制的errgroup
	g, gctx := errgroup.WithContext(timeoutCtx)

	// 统计成功添加的key数量
	var addedKeys int64
	var mu sync.Mutex

	// 启动100个协程
	for i := 0; i < goroutineCount; i++ {
		start := i * keysPerGoroutine
		end := start + keysPerGoroutine

		// 确保最后一个协程处理剩余的keys
		if i == goroutineCount-1 {
			end = totalKeys
		}

		g.Go(func() error {
			localAddedKeys := 0

			// 为每个协程处理一批keys
			for j := start; j < end; j++ {
				keyName := fmt.Sprintf("key%d", j+1)     // key1, key2, key3...
				keyValue := fmt.Sprintf("value_%d", j+1) // 简单的值

				// 执行SET命令
				_, err := api.RedisExecCommand(gctx, req.Database, "SET", keyName, keyValue)
				if err != nil {
					logger.DefaultLogger.Error("添加Key失败", "key:", keyName, "error:", err)
					// 如果是因为context取消，直接返回
					if gctx.Err() != nil {
						return gctx.Err()
					}
					continue // 继续处理下一个key
				}

				localAddedKeys++

				// 每1000个key检查一次context状态
				if (j-start+1)%1000 == 0 {
					if gctx.Err() != nil {
						break
					}
				}
			}

			// 线程安全地更新总计数
			mu.Lock()
			addedKeys += int64(localAddedKeys)
			mu.Unlock()

			return nil
		})
	}

	// 等待所有协程完成
	err = g.Wait()

	endTime := time.Now()
	timeTaken := endTime.Sub(startTime)

	finalAddedKeys := int(addedKeys)

	var message string
	if err != nil {
		if err == context.DeadlineExceeded {
			message = "批量添加Keys超时，部分添加成功"
		} else {
			message = "批量添加Keys过程中出现错误，部分添加成功"
		}
		logger.DefaultLogger.Error("批量添加Keys失败", "error:", err, "成功添加:", finalAddedKeys)
	} else {
		message = "批量添加Keys完成"
		logger.DefaultLogger.Debug("批量添加Keys成功", "总数:", totalKeys, "成功:", finalAddedKeys, "耗时:", timeTaken)
	}

	this.Success(ctx, response.SearchSuccess, vo.RedisBatchAddKeysResponse{
		Message:   message,
		TotalKeys: totalKeys,
		AddedKeys: finalAddedKeys,
		TimeTaken: fmt.Sprintf("%.2f秒", timeTaken.Seconds()),
	})
}

// BatchGetMemoryAnalysisAction 批量获取keys的内存分析
func (this *RedisController) BatchGetMemoryAnalysisAction(ctx *gin.Context) {
	req := new(dto.RedisBatchMemoryAnalysisRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		this.Error(ctx, err)
		return
	}

	logger.DefaultLogger.Debug("批量内存分析",
		"conn_id:", req.EsConnect,
		"database:", req.Database,
		"keys_count:", len(req.Keys))

	// 调用基座API
	api := ev_api.NewEvWrapApi(req.EsConnect, util.GetEvUserID(ctx))

	// 预先检查Redis版本是否支持MEMORY USAGE命令
	if len(req.Keys) > 0 {
		// 使用第一个key进行版本检查
		_, err := this.executeRedisCommandWithRetry(ctx, api, req.Database, "MEMORY", "USAGE", req.Keys[0])
		if err != nil {
			if strings.Contains(err.Error(), "unknown command") ||
				strings.Contains(err.Error(), "ERR unknown command") ||
				strings.Contains(err.Error(), "MEMORY") {
				logger.DefaultLogger.Error("Redis版本不支持MEMORY USAGE命令", "error:", err)
				this.Error(ctx, fmt.Errorf("Redis版本不支持MEMORY USAGE命令，请使用Redis 4.0+版本"))
				return
			}
		}
	}

	// 分析每个key的内存使用情况 - 使用并发处理
	var keyMemoryInfos []vo.RedisKeyMemoryInfo
	var totalSize int64 = 0
	var processedKeys int = 0
	var mu sync.Mutex // 保护共享数据

	if len(req.Keys) > 0 {
		// 创建带有并发限制的errgroup
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(2000) // 使用官方API时适当降低并发数，避免对Redis造成过大压力

		// 为每个key创建一个goroutine
		for _, key := range req.Keys {
			key := key // 避免闭包问题
			g.Go(func() error {
				// 使用官方MEMORY USAGE API分析内存信息
				keyInfo, err := this.analyzeKeyMemoryOfficial(gctx, api, req.Database, key)
				if err != nil {
					logger.DefaultLogger.Error("分析Key内存失败", "key:", key, "error:", err)
					// 不返回错误，继续处理其他key，但不计入结果
					return nil
				}

				// 线程安全地添加结果
				mu.Lock()
				keyMemoryInfos = append(keyMemoryInfos, keyInfo)
				totalSize += keyInfo.SizeBytes
				processedKeys++
				mu.Unlock()

				return nil
			})
		}

		// 等待所有goroutine完成
		if err := g.Wait(); err != nil {
			logger.DefaultLogger.Error("批量内存分析失败", "error:", err)
		}
	}

	logger.DefaultLogger.Debug("批量内存分析完成",
		"请求keys:", len(req.Keys),
		"成功处理:", processedKeys,
		"总大小:", totalSize)

	this.Success(ctx, response.SearchSuccess, vo.RedisBatchMemoryAnalysisResponse{
		KeyMemoryInfos: keyMemoryInfos,
		TotalSize:      totalSize,
		ProcessedKeys:  processedKeys,
		TotalKeys:      len(req.Keys),
	})
}
