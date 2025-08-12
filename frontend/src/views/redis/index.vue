<template>
  <div class="redis-page" :class="{ 'mobile-layout': isMobile }">
    <div class="page-header">
      <h2>Redis 信息总览</h2>
      
      <!-- 刷新控制区域 -->
      <div class="refresh-controls" :class="{ 'mobile-controls': isMobile }">
    
       
         
        <div class="auto-refresh-section">
          <el-tooltip 
            content="每2秒刷新一次"
            placement="bottom"
            :disabled="!autoRefresh"
          >
            <el-switch
              v-model="autoRefresh"
              :size="isMobile ? 'small' : 'default'"
              active-text="自动刷新"
              :active-color="'#409eff'"
              @change="handleAutoRefreshChange"
            />
          </el-tooltip>
        </div>
        
        <div class="last-refresh-time" v-if="lastRefreshTime">
          <span class="time-label">最后刷新:</span>
          <span class="time-value">{{ lastRefreshTime }}</span>
        </div>
      </div>
    </div>

    <!-- 服务器、内存、状态信息卡片 -->
    <el-row :gutter="isMobile ? 10 : 20" class="info-cards">
      <el-col :span="isMobile ? 24 : 8" :class="{ 'mobile-col': isMobile }">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <i class="el-icon-monitor"></i>
              <span>服务器信息</span>
            </div>
          </template>
          <div class="info-item">
            <span class="label">Redis版本:</span>
            <span class="value">{{ info.redis_version || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">操作系统:</span>
            <span class="value">{{ info.os || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">进程ID:</span>
            <span class="value">{{ info.process_id || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">运行模式:</span>
            <span class="value">{{ info.redis_mode || '-' }}</span>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="isMobile ? 24 : 8" :class="{ 'mobile-col': isMobile }">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <i class="el-icon-pie-chart"></i>
              <span>内存信息</span>
            </div>
          </template>
          <div class="info-item">
            <span class="label">已用内存:</span>
            <span class="value">{{ info.used_memory_human || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">内存峰值:</span>
            <span class="value">{{ info.used_memory_peak_human || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Lua内存:</span>
            <span class="value">{{ info.used_memory_lua_human || info.used_memory_lua || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">内存占用率:</span>
            <span class="value">{{ info.used_memory_rss_human || '-' }}</span>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="isMobile ? 24 : 8" :class="{ 'mobile-col': isMobile }">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <i class="el-icon-data-line"></i>
              <span>状态信息</span>
            </div>
          </template>
          <div class="info-item">
            <span class="label">客户端连接数:</span>
            <span class="value">{{ info.connected_clients || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">历史连接数:</span>
            <span class="value">{{ info.total_connections_received || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">历史命令数:</span>
            <span class="value">{{ info.total_commands_processed || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">运行时间:</span>
            <span class="value">{{ formatUptime(info.uptime_in_seconds) }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 键值统计表格 -->
    <el-card class="table-card" v-if="keyspace.length > 0">
      <template #header>
        <div class="card-header">
          <i class="el-icon-key"></i>
          <span>键值统计</span>
        </div>
      </template>
      <div class="table-wrapper">
        <el-table :data="keyspace" stripe :size="isMobile ? 'small' : 'default'">
          <el-table-column prop="db" label="数据库" :width="isMobile ? 80 : undefined"/>
          <el-table-column prop="keys" label="Keys" :width="isMobile ? 80 : undefined"/>
          <el-table-column prop="expires" label="过期" :width="isMobile ? 80 : undefined"/>
          <el-table-column prop="avg_ttl" label="TTL" :width="isMobile ? 100 : undefined"/>
        </el-table>
      </div>
    </el-card>
 
    <!-- Redis信息全集表格 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <i class="el-icon-document"></i>
          <span>Redis信息全集</span>
        </div>
      </template>
      
      <!-- 搜索框 -->
      <div class="search-container" :class="{ 'mobile-search': isMobile }">
        <el-input
          v-model="searchText"
          placeholder="搜索Key或Value..."
          prefix-icon="el-icon-search"
          clearable
          :style="{ width: isMobile ? '100%' : '300px' }"
          :size="isMobile ? 'small' : 'default'"
        />
        <span class="search-info" v-if="!isMobile">共 {{ filteredInfoData.length }} 条记录</span>
        <div class="mobile-search-info" v-if="isMobile">
          共 {{ filteredInfoData.length }} 条记录
        </div>
    </div>

      <div class="table-wrapper">
        <el-table 
          :data="filteredInfoData" 
          stripe 
          :max-height="isMobile ? 300 : 400"
          :size="isMobile ? 'small' : 'default'"
        >
          <el-table-column 
            prop="key" 
            label="Key" 
            :width="isMobile ? 120 : undefined"
            :show-overflow-tooltip="true"
          />
          <el-table-column 
            prop="value" 
            label="Value" 
            :show-overflow-tooltip="true"
          />
        </el-table>
    </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { sdk } from '@elasticview/plugin-sdk'
import { getRedisInfoOverview, batchAddRedisKeys } from "@/api/redis";

const info = ref({})
const keyspace = ref([])
const searchText = ref('')
const autoRefresh = ref(false)
const lastRefreshTime = ref('')
const batchAddLoading = ref(false)
let refreshTimer = null

// 检查是否为移动设备
const isMobile = computed(() => {
  return sdk.IsMobile()
})

// 计算属性：将info对象转换为表格数据
const infoTableData = computed(() => {
  return Object.entries(info.value).map(([key, value]) => ({ key, value }))
})

// 过滤信息数据
const filteredInfoData = computed(() => {
  if (!searchText.value) return infoTableData.value
  const lowerCaseSearchText = searchText.value.toLowerCase()
  return infoTableData.value.filter((item) => {
    return item.key.toLowerCase().includes(lowerCaseSearchText) || 
           item.value.toLowerCase().includes(lowerCaseSearchText)
  })
})

// 格式化运行时间
const formatUptime = (seconds) => {
  if (!seconds) return '-'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (isMobile.value) {
    // 移动端显示简化版本
    if (days > 0) return `${days}天${hours}时`
    if (hours > 0) return `${hours}时${minutes}分`
    return `${minutes}分钟`
  }
  
  return `${days}天 ${hours}小时 ${minutes}分钟`
}

// 格式化时间
const formatTime = (date) => {
  const now = new Date(date)
  const hours = now.getHours().toString().padStart(2, '0')
  const minutes = now.getMinutes().toString().padStart(2, '0')
  const seconds = now.getSeconds().toString().padStart(2, '0')
  return `${hours}:${minutes}:${seconds}`
}

// 获取Redis信息总览
const fetchRedisInfo = async (showMessage = true) => {
  try {
    const connId = sdk.GetSelectEsConnID()
    console.log('获取Redis信息总览，连接ID:', connId)
    
    const res = await getRedisInfoOverview({ 
      es_connect: connId,
      database: 0 
    })
    
    console.log('Redis信息总览响应:', res)
    
    if (res.code === 0) {
      info.value = res.data.info || {}
      keyspace.value = res.data.keyspace || []
      // 更新最后刷新时间
      lastRefreshTime.value = formatTime(new Date())
      if (showMessage) {
        ElMessage.success('Redis信息获取成功')
      }
    } else {
      ElMessage.error(res.msg || '获取Redis信息失败')
    }
  } catch (error) {
    console.error('获取Redis信息失败:', error)
    ElMessage.error('网络请求失败: ' + error.message)
  }
}

// 处理自动刷新开关变化
const handleAutoRefreshChange = (value) => {
  if (value) {
    // 开启自动刷新
    ElMessage.success('已开启自动刷新，每2秒刷新一次')
    startAutoRefresh()
  } else {
    // 关闭自动刷新
    ElMessage.info('已关闭自动刷新')
    stopAutoRefresh()
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(() => {
    fetchRedisInfo(false) // 自动刷新时不显示成功消息
  }, 2000) // 每2秒刷新一次
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 处理批量添加Keys
const handleBatchAddKeys = async () => {
  // 确认对话框
  const confirmResult = await ElMessageBox.confirm(
    '将使用100个协程并发添加1,000,000个Key (key1-key1000000)，此操作可能需要较长时间，是否继续？',
    '批量添加确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).catch(() => false)
  
  if (!confirmResult) {
    return
  }
  
  batchAddLoading.value = true
  try {
    const connId = sdk.GetSelectEsConnID()
    console.log('批量添加Keys，连接ID:', connId)
    
    const res = await batchAddRedisKeys({ 
      es_connect: connId,
      database: 0 
    })
    
    console.log('批量添加Keys响应:', res)
    
    if (res.code === 0) {
      const data = res.data
      ElMessage.success(`${data.message}！成功添加 ${data.addedKeys} / ${data.totalKeys} 个Key，耗时: ${data.timeTaken}`)
      // 刷新页面信息
      fetchRedisInfo(false)
    } else {
      ElMessage.error(res.msg || '批量添加Keys失败')
    }
  } catch (error) {
    console.error('批量添加Keys失败:', error)
    ElMessage.error('网络请求失败: ' + error.message)
  } finally {
    batchAddLoading.value = false
  }
}

onMounted(() => {
  console.log('Redis信息总览页面已挂载')
  console.log('是否为移动设备:', isMobile.value)
  fetchRedisInfo()
})

onUnmounted(() => {
  // 组件卸载时清理定时器
  stopAutoRefresh()
})
</script>

<style scoped>
.redis-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 500;
}

.refresh-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.auto-refresh-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.last-refresh-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.last-refresh-time .time-label {
  color: #909399;
}

.last-refresh-time .time-value {
  color: #409eff;
  font-weight: 500;
}

.info-cards {
  margin-bottom: 20px;
}

.info-card {
  height: 200px;
}

.card-header {
  display: flex;
  align-items: center;
  font-weight: 500;
}

.card-header i {
  margin-right: 8px;
  font-size: 16px;
  color: #409eff;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.info-item .label {
  font-size: 14px;
}

.info-item .value {
  font-weight: 500;
  font-size: 14px;
}

.table-card {
  margin-top: 20px;
}

.table-card .card-header {
  font-size: 16px;
}

.search-container {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
}

.search-container .el-input {
  margin-right: 10px;
}

.search-container .search-info {
  font-size: 14px;
  color: #909399;
}

.table-wrapper {
  overflow-x: auto;
}

/* 移动端适配 */
.mobile-layout {
  padding: 10px;
}

.mobile-layout .page-header {
  flex-direction: column;
  align-items: stretch;
  gap: 15px;
}

.mobile-layout .page-header h2 {
  font-size: 20px;
  text-align: center;
}

.mobile-controls {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.mobile-controls .el-button {
  width: 100%;
  margin-bottom: 5px;
}

.mobile-controls .auto-refresh-section {
  flex-shrink: 0;
}

.mobile-controls .last-refresh-time {
  flex-shrink: 0;
  font-size: 11px;
}

.mobile-col {
  margin-bottom: 15px;
}

.mobile-layout .info-card {
  height: auto;
  min-height: 160px;
}

.mobile-layout .info-item {
  padding: 6px 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.mobile-layout .info-item .label {
  font-size: 12px;
  color: #909399;
}

.mobile-layout .info-item .value {
  font-size: 14px;
  font-weight: 600;
}

.mobile-layout .card-header {
  font-size: 14px;
}

.mobile-layout .card-header i {
  font-size: 14px;
}

.mobile-search {
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

.mobile-search .el-input {
  flex: 1;
  margin-right: 0;
}

.mobile-search-info {
  font-size: 11px;
  color: #909399;
  white-space: nowrap;
  flex-shrink: 0;
}

.mobile-layout .table-card {
  margin-top: 15px;
}

/* 移动端表格优化 */
@media (max-width: 768px) {
  .redis-page {
    padding: 10px;
  }
  
  .info-cards .el-col {
    margin-bottom: 10px;
  }
  
  .table-wrapper {
    font-size: 12px;
  }
  
  .el-table .cell {
    padding: 0 5px;
  }
}

/* 超小屏幕适配 */
@media (max-width: 480px) {
  .redis-page {
    padding: 8px;
  }
  
  .page-header h2 {
    font-size: 18px;
  }
  
  .info-item .label {
    font-size: 11px;
  }
  
  .info-item .value {
    font-size: 13px;
  }
  
  .last-refresh-time {
    font-size: 11px;
  }
}

.dark-theme .info-item .value {
  color: #ffffff;
}

.dark-theme .info-item {
  border-bottom-color: #4d4d4d;
}
</style>