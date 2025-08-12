package dto

// Redis操作请求DTO
type RedisKeysRequest struct {
	EsConnect int `json:"es_connect"` // 数据源连接ID
	Database  int `json:"database"`   // Redis数据库索引，默认为0
}

// Redis Key查询响应DTO
type RedisKeysResponse struct {
	Keys []string `json:"keys"` // Redis所有key列表
}

// Redis信息总览请求DTO
type RedisInfoRequest struct {
	EsConnect int `json:"es_connect"` // 数据源连接ID
	Database  int `json:"database"`   // Redis数据库索引，默认为0
}

// Redis数据库列表请求DTO
type RedisDatabasesRequest struct {
	EsConnect int `json:"es_connect"` // 数据源连接ID
}

// Redis内存分析请求DTO
type RedisMemoryAnalysisRequest struct {
	EsConnect int    `json:"es_connect"` // 数据源连接ID
	Database  int    `json:"database"`   // Redis数据库索引，默认为0
	Pattern   string `json:"pattern"`    // Key匹配模式，默认为*
	Count     int    `json:"count"`      // 单次返回数量，默认为50
	Cursor    string `json:"cursor"`     // 游标，用于分页
}

// Redis Key删除请求DTO
type RedisDeleteKeyRequest struct {
	EsConnect int    `json:"es_connect"` // 数据源连接ID
	Database  int    `json:"database"`   // Redis数据库索引，默认为0
	Key       string `json:"key"`        // 要删除的Key
}

// Redis Key详情请求DTO
type RedisKeyDetailRequest struct {
	EsConnect int    `json:"es_connect"` // 数据源连接ID
	Database  int    `json:"database"`   // Redis数据库索引，默认为0
	Key       string `json:"key"`        // 要查询的Key
}

// Redis Key保存请求DTO
type RedisSetKeyRequest struct {
	EsConnect int         `json:"es_connect"` // 数据源连接ID
	Database  int         `json:"database"`   // Redis数据库索引，默认为0
	Key       string      `json:"key"`        // Key名称
	Type      string      `json:"type"`       // 数据类型 (string, hash, list, set, zset)
	TTL       int64       `json:"ttl"`        // 过期时间（秒），-1表示永不过期
	Value     interface{} `json:"value"`      // 值，根据类型不同而不同
}

// Redis Key搜索请求DTO (后端搜索)
type RedisSearchKeysRequest struct {
	EsConnect     int    `json:"es_connect"`     // 数据源连接ID
	Database      int    `json:"database"`       // Redis数据库索引，默认为0
	SearchText    string `json:"search_text"`    // 搜索文本，使用 strings.Contains 匹配
	Count         int    `json:"count"`          // 单次返回数量，默认为50
	Cursor        string `json:"cursor"`         // 游标，用于分页
	CaseSensitive bool   `json:"case_sensitive"` // 是否区分大小写，默认false
}

// Redis 批量添加 Key 请求 DTO
type RedisBatchAddKeysRequest struct {
	EsConnect int `json:"es_connect"` // 数据源连接ID
	Database  int `json:"database"`   // Redis数据库索引，默认为0
}

// Redis 批量内存分析请求 DTO - 接收key数组，返回每个key的内存消耗
type RedisBatchMemoryAnalysisRequest struct {
	EsConnect int      `json:"es_connect"` // 数据源连接ID
	Database  int      `json:"database"`   // Redis数据库索引，默认为0
	Keys      []string `json:"keys"`       // 要分析的key数组
}
