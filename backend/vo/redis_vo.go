package vo

// Redis Keys查询响应VO
type RedisKeysResponse struct {
	Keys []string `json:"keys"` // Redis所有key列表
}

// Redis信息总览响应VO
type RedisInfoResponse struct {
	Info     map[string]string        `json:"info"`     // Redis info信息
	Keyspace []map[string]interface{} `json:"keyspace"` // 各数据库的key统计信息
}

// Redis数据库信息
type RedisDatabaseInfo struct {
	Database int   `json:"database"` // 数据库索引
	Keys     int64 `json:"keys"`     // Key数量
	Expires  int64 `json:"expires"`  // 过期Key数量
	AvgTTL   int64 `json:"avgTtl"`   // 平均TTL
}

// Redis数据库列表响应VO
type RedisDatabasesResponse struct {
	Databases []RedisDatabaseInfo `json:"databases"` // 数据库列表
}

// Redis内存分析单个Key信息
type RedisKeyMemoryInfo struct {
	Key       string `json:"key"`       // Key名称
	SizeBytes int64  `json:"sizeBytes"` // 大小（字节）
	Type      string `json:"type"`      // 数据类型
	TTL       int64  `json:"ttl"`       // 过期时间（秒，-1表示永不过期）
}

// Redis内存分析响应VO
type RedisMemoryAnalysisResponse struct {
	Keys       []RedisKeyMemoryInfo `json:"keys"`       // Key内存信息列表
	TotalKeys  int                  `json:"totalKeys"`  // 当前返回的Key数量
	TotalSize  int64                `json:"totalSize"`  // 当前返回Keys的总大小（字节）
	NextCursor string               `json:"nextCursor"` // 下一页游标，为"0"表示已到末尾
	TotalCount int                  `json:"totalCount"` // 数据库中的总Key数量（估算）
}

// Redis Key详情响应VO
type RedisKeyDetailResponse struct {
	Key       string      `json:"key"`       // Key名称
	Type      string      `json:"type"`      // 数据类型
	SizeBytes int64       `json:"sizeBytes"` // 大小（字节）
	TTL       int64       `json:"ttl"`       // 过期时间
	Value     interface{} `json:"value"`     // Key的值（根据类型不同而不同）
}

// Redis操作响应VO
type RedisOperationResponse struct {
	Success bool   `json:"success"` // 操作是否成功
	Message string `json:"message"` // 操作结果消息
}

// Redis Key搜索响应VO (后端搜索)
type RedisSearchKeysResponse struct {
	Keys         []RedisKeyMemoryInfo `json:"keys"`         // 匹配的Key信息列表
	TotalKeys    int                  `json:"totalKeys"`    // 当前返回的Key数量
	TotalSize    int64                `json:"totalSize"`    // 当前返回Keys的总大小（字节）
	NextCursor   string               `json:"nextCursor"`   // 下一页游标，为"0"表示已到末尾
	TotalCount   int                  `json:"totalCount"`   // 数据库中的总Key数量（估算）
	SearchText   string               `json:"searchText"`   // 搜索文本
	MatchedCount int                  `json:"matchedCount"` // 匹配的总数量（估算）
}

// Redis 批量添加 Key 响应 VO
type RedisBatchAddKeysResponse struct {
	Message   string `json:"message"`   // 响应消息
	TotalKeys int    `json:"totalKeys"` // 计划添加的总 key 数量
	AddedKeys int    `json:"addedKeys"` // 成功添加的 key 数量
	TimeTaken string `json:"timeTaken"` // 耗时
}

// Redis 批量内存分析响应 VO
type RedisBatchMemoryAnalysisResponse struct {
	KeyMemoryInfos []RedisKeyMemoryInfo `json:"keyMemoryInfos"` // key内存信息列表
	TotalSize      int64                `json:"totalSize"`      // 总大小（字节）
	ProcessedKeys  int                  `json:"processedKeys"`  // 成功处理的key数量
	TotalKeys      int                  `json:"totalKeys"`      // 请求处理的key总数
}
