<template>
  <div class="memory-analysis-page" :class="{ 'mobile-layout': isMobile, 'dark-mode': isDarkMode }">
    <div class="page-header">
      <div class="header-left">
        <h2>
          <i class="el-icon-pie-chart"></i>
          内存分析
        </h2>
        <div class="stats-info">
          <span class="stat-item" v-if="memoryKeys.length > 0">Keys: {{ memoryKeys.length.toLocaleString() }}</span>
          <span class="stat-item" v-if="totalSize > 0">Size: {{ formatSize(totalSize) }}</span>
          <span class="stat-item" v-if="filteredKeys.length !== memoryKeys.length && memoryKeys.length > 0">
            Filtered: {{ filteredKeys.length.toLocaleString() }}
          </span>
        </div>
      </div>
    </div>

    <!-- 搜索和过滤区域 -->
    <el-card class="filter-card">
      <div class="filter-controls" :class="{ 'mobile-filters': isMobile }">
        <!-- 数据库选择 -->
        <div class="database-selector">
          <el-select
            v-model="selectedDatabase"
            placeholder="选择数据库"
            :size="isMobile ? 'small' : 'default'"
            style="width: 150px;"
            @change="onDatabaseChange"
            :loading="loadingDatabases"
          >
            <el-option
              v-for="db in databases"
              :key="db.database"
              :label="`DB${db.database} (${db.keys} keys)`"
              :value="db.database"
            />
          </el-select>
       
        </div>

        <el-input
          v-model="searchText"
          placeholder="搜索Key..."
          prefix-icon="el-icon-search"
          clearable
            :style="{ width: isMobile ? '200px' : '250px' }"
          :size="isMobile ? 'small' : 'default'"
            @keyup.enter="performSearch"
            @clear="performSearch"
          />
          <el-button
            type="primary"
            :size="isMobile ? 'small' : 'default'"
            @click="performSearch"
            :disabled="memoryKeys.length === 0"
          >
            搜索
          </el-button>

        <el-input
          v-model="minSizeKB"
          placeholder="最小大小(KB)"
          clearable
          type="number"
          :style="{ width: isMobile ? '100%' : '120px' }"
          :size="isMobile ? 'small' : 'default'"
        />



        <el-button
            type="primary"
            :size="isMobile ? 'small' : 'default'"
          :loading="analyzing"
          @click="startMemoryAnalysis"
          :disabled="analyzing"
        >
          {{ analyzing ? '分析中...' : (memoryKeys.length > 0 ? '重新分析' : '开始内存分析') }}
        </el-button>

        <el-button
          v-if="analyzing"
          type="danger"
          :size="isMobile ? 'small' : 'default'"
          @click="stopAnalysis"
        >
          停止分析
        </el-button>
      </div>

      <!-- 分析进度 -->
      <div v-if="analyzing || analysisProgress.total > 0" class="analysis-progress">
        <div class="progress-info">
          <span v-if="analysisProgress.total > 0">分析进度: {{ analysisProgress.processed }} / {{ analysisProgress.total }}</span>
          <span v-else>正在获取Keys列表...</span>
          <span class="progress-percent" v-if="analysisProgress.total > 0">
            ({{ Math.round((analysisProgress.processed / analysisProgress.total) * 100) }}%)
          </span>
        </div>
        <el-progress
          :percentage="analysisProgress.total > 0 ? Math.round((analysisProgress.processed / analysisProgress.total) * 100) : 100"
          :status="analyzing ? 'success' : 'success'"
          :stroke-width="8"
          :indeterminate="analysisProgress.total === 0"
        />
        <div class="progress-details">
          <span v-if="analysisProgress.totalBatches > 0">
            当前批次: {{ analysisProgress.currentBatch }} / {{ analysisProgress.totalBatches }}
          </span>
          <span v-if="analysisProgress.estimatedTime">
            预计剩余: {{ analysisProgress.estimatedTime }}
          </span>
          <span v-if="analysisProgress.total === 0">
            正在获取Keys列表...
          </span>
        </div>
      </div>
      
      <!-- 分析完成后的统计 -->
      <div v-if="!analyzing && memoryKeys.length > 0 && analysisProgress.startTime" class="analysis-result">
        <div class="result-info">
          <span>分析完成！共分析了 {{ memoryKeys.length }} 个Keys</span>
          <span class="result-time">耗时: {{ formatTime(Date.now() - analysisProgress.startTime) }}</span>
        </div>
      </div>
    </el-card>

    <!-- 内存分析结果表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <div class="table-controls">
          <span class="table-title">内存使用分析</span>
          <div class="sort-controls">
            <el-button-group>
              <el-button
                :type="sortBy === 'size' && sortOrder === 'desc' ? 'primary' : 'default'"
                @click="setSortBy('size', 'desc')"
              >
                <i class="el-icon-top"></i>
                按大小降序
              </el-button>
              <el-button
                :type="sortBy === 'size' && sortOrder === 'asc' ? 'primary' : 'default'"
                @click="setSortBy('size', 'asc')"
              >
                <i class="el-icon-bottom"></i>
                按大小升序
              </el-button>
            </el-button-group>
          </div>
        </div>
      </div>

      <div class="table-wrapper" v-loading="analyzing && memoryKeys.length === 0" element-loading-text="正在分析内存使用情况...">
        <!-- 空状态 -->
        <div v-if="memoryKeys.length === 0 && !analyzing" class="empty-state">
          <el-empty 
            description="点击开始分析按钮进行内存分析"
            :image-size="120"
          >
              <el-button
              type="primary" 
              @click="startMemoryAnalysis"
              :loading="analyzing"
            >
              开始内存分析
              </el-button>
          </el-empty>
        </div>

        <!-- 虚拟化结果表格 -->
        <el-auto-resizer v-else>
          <template #default="{ height, width }">
            <el-table-v2
              :key="tableKey"
              :columns="tableColumns"
              :data="filteredKeys"
              :width="width"
              :height="height"
              :sort-by="sortState"
              @column-sort="handleColumnSort"
              :row-height="isMobile ? 40 : 48"
              :header-height="isMobile ? 40 : 48"
              :estimated-row-height="isMobile ? 40 : 48"
            />
            </template>
        </el-auto-resizer>
      </div>

      <!-- 数据统计 -->
    
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, h } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

import { sdk } from '@elasticview/plugin-sdk'
import { getRedisKeys, getRedisBatchMemoryAnalysis, getRedisDatabases } from "@/api/redis";

const analyzing = ref(false)
const loadingDatabases = ref(false)
const searchText = ref('')
const appliedSearchText = ref('') // 实际应用的搜索文本
const minSizeKB = ref('')
const sortBy = ref('size')
const sortOrder = ref('desc')
const tableHeight = ref(600)

// 虚拟表格排序状态
const sortState = ref({
  key: 'sizeBytes',
  order: 'desc'
})

// 数据库相关
const databases = ref([])
const selectedDatabase = ref(0)

// Keys数据
const memoryKeys = ref([]) // 已分析内存的keys
const totalSize = ref(0)

// 分析进度
const analysisProgress = ref({
  processed: 0,
  total: 0,
  currentBatch: 0,
  totalBatches: 0,
  estimatedTime: '',
  startTime: null
})

// 表格重新渲染key
const tableKey = ref(0)

// 复制到剪贴板函数
const copyToClipboard = async (text) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制到剪贴板')
    } else {
      // 降级方案：使用传统的document.execCommand
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      const successful = document.execCommand('copy')
      document.body.removeChild(textArea)
      
      if (successful) {
        ElMessage.success('已复制到剪贴板')
      } else {
        ElMessage.error('复制失败')
      }
    }
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  }
}

// 检查是否为移动设备和暗色模式
const isMobile = computed(() => {
  return sdk.IsMobile()
})

const isDarkMode = computed(() => {
  return sdk.isDarkTheme()
})

// 虚拟表格列定义
const tableColumns = computed(() => {
  const columns = []

  // Key列
  columns.push({
    key: 'key',
    title: 'Key',
    dataKey: 'key',
    width: isMobile.value ? 200 : 300,
    sortable: true,
    cellRenderer: ({ rowData }) => {
      return h('div', {
        class: 'key-cell',
        title: rowData.key,
      }, [
        h('span', { class: 'key-text' }, rowData.key),
        h('span', {
          class: 'copy-btn',
          title: '复制Key',
          style: {
            border: '1px solid #409eff',
            borderRadius: '4px',
            padding: '2px 6px',
            cursor: 'pointer',
            color: '#409eff',
            fontSize: '12px',
            opacity: '0.7',
            transition: 'opacity 0.2s ease',
            marginLeft: '8px'
          },
          onClick: (e) => {
            e.stopPropagation()
            copyToClipboard(rowData.key)
          }
        }, '复制')
      ])
    }
  })

  // 大小列
  columns.push({
    key: 'size',
    title: '大小',
    dataKey: 'sizeBytes',
    width: 120,
    sortable: true,
    cellRenderer: ({ rowData }) => {
      return h('span', {
        style: { fontWeight: '500' }
      }, formatSize(rowData.sizeBytes || 0))
    }
  })



  return columns
})

// 过滤后的数据
const filteredKeys = computed(() => {
  let filtered = memoryKeys.value

  // 搜索过滤
  if (appliedSearchText.value && appliedSearchText.value.trim()) {
    const searchTerm = appliedSearchText.value.trim().toLowerCase()
    filtered = filtered.filter(item => {
      const key = item.key.toLowerCase()

      if (key.includes(searchTerm)) return true
      if (searchTerm.startsWith('size:')) {
        const sizeQuery = searchTerm.substring(5)
        if (sizeQuery.startsWith('>')) {
          const sizeTerm = sizeQuery.substring(1)
          const sizeBytes = parseSizeToBytes(sizeTerm)
          return sizeBytes > 0 && item.sizeBytes > sizeBytes
        } else if (sizeQuery.startsWith('<')) {
          const sizeTerm = sizeQuery.substring(1)
          const sizeBytes = parseSizeToBytes(sizeTerm)
          return sizeBytes > 0 && item.sizeBytes < sizeBytes
        }
      }
      if (searchTerm.includes('*')) {
        const regex = new RegExp(searchTerm.replace(/\*/g, '.*'), 'i')
        return regex.test(key)
      }
      return false
    })
  }

  // 最小大小过滤
  if (minSizeKB.value && !isNaN(minSizeKB.value)) {
    const minSizeBytes = parseInt(minSizeKB.value) * 1024
    filtered = filtered.filter(item => item.sizeBytes >= minSizeBytes)
  }

  // 排序
  if (sortBy.value && sortOrder.value) {
  filtered.sort((a, b) => {
      let valueA, valueB
      
      switch (sortBy.value) {
        case 'size':
          valueA = a.sizeBytes || 0
          valueB = b.sizeBytes || 0
          break
        case 'key':
          valueA = a.key || ''
          valueB = b.key || ''
          break
        default:
          return 0
      }
      
      if (sortOrder.value === 'asc') {
        return valueA > valueB ? 1 : valueA < valueB ? -1 : 0
      } else {
        return valueA < valueB ? 1 : valueA > valueB ? -1 : 0
      }
    })
  }

  return filtered
})



// 开始内存分析
const startMemoryAnalysis = async () => {
  if (analyzing.value) return

  analyzing.value = true
  memoryKeys.value = []
  totalSize.value = 0
  
  // 重置进度
  analysisProgress.value = {
    processed: 0,
    total: 0,
    currentBatch: 0,
    totalBatches: 0,
    estimatedTime: '',
    startTime: Date.now()
  }

  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      analyzing.value = false
      return
    }

    // 先获取keys并分批分析，以便显示真实进度
    // 第一步：获取所有keys
    const keysResult = await getRedisKeys({
      es_connect: connId,
      database: selectedDatabase.value
    })
    
    if (keysResult.code !== 0 || !keysResult.data || !keysResult.data.keys) {
      ElMessage.warning(keysResult.msg || '未找到任何Keys')
      analyzing.value = false
      return
    }
    
    const allKeys = keysResult.data.keys
    analysisProgress.value.total = allKeys.length
    
    if (allKeys.length === 0) {
      ElMessage.info('数据库中没有Keys')
      analyzing.value = false
      return
    }
     
    // 第二步：分批分析内存
    const batchSize = 2000 // 每批处理10000个keys
    const totalBatches = Math.ceil(allKeys.length / batchSize)
    analysisProgress.value.totalBatches = totalBatches
    
    for (let i = 0; i < allKeys.length && analyzing.value; i += batchSize) {
      const batch = allKeys.slice(i, i + batchSize)
      analysisProgress.value.currentBatch = Math.floor(i / batchSize) + 1
      
      try {
        const result = await getRedisBatchMemoryAnalysis({
      es_connect: connId,
        database: selectedDatabase.value,
          keys: batch
        })
        
        if (result.code === 0 && result.data && result.data.keyMemoryInfos) {
          memoryKeys.value.push(...result.data.keyMemoryInfos)
          totalSize.value += result.data.totalSize || 0
          
          analysisProgress.value.processed += result.data.processedKeys || batch.length
          
          // 计算预计剩余时间
          if (analysisProgress.value.processed > 0) {
            const elapsed = Date.now() - analysisProgress.value.startTime
            const rate = analysisProgress.value.processed / elapsed
            const remaining = (allKeys.length - analysisProgress.value.processed) / rate
            analysisProgress.value.estimatedTime = formatTime(remaining)
          }
          
          tableKey.value++
        }
      } catch (batchError) {
        console.error(`批次 ${analysisProgress.value.currentBatch} 分析失败:`, batchError)
        
        if (batchError.message && batchError.message.includes('Redis版本不支持')) {
          ElMessage.error('Redis版本不支持MEMORY USAGE命令，请使用Redis 4.0+版本')
          analyzing.value = false
          return
        }
      }
      
      // 短暂延迟，避免过度占用资源
      await new Promise(resolve => setTimeout(resolve, 50))
    }
    
    if (analyzing.value) {
      const totalTime = Date.now() - analysisProgress.value.startTime
      const timeStr = formatTime(totalTime)
      ElMessage.success(`内存分析完成！共分析了 ${memoryKeys.value.length} 个Keys，耗时 ${timeStr}`)
    }
  } catch (error) {
    console.error('内存分析失败:', error)
    
    if (error.message && error.message.includes('Redis版本不支持')) {
      ElMessage.error('Redis版本不支持MEMORY USAGE命令，请使用Redis 4.0+版本')
    } else {
      ElMessage.error('内存分析失败: ' + (error.message || '未知错误'))
    }
  } finally {
    analyzing.value = false
  }
}

// 停止分析
const stopAnalysis = () => {
  analyzing.value = false
  ElMessage.info('已停止内存分析')
}

// 重置分析进度
const resetAnalysisProgress = () => {
  analysisProgress.value = {
    processed: 0,
    total: 0,
    currentBatch: 0,
    totalBatches: 0,
    estimatedTime: '',
    startTime: null
  }
}

// 数据库变化处理
const onDatabaseChange = () => {
  memoryKeys.value = []
  totalSize.value = 0
  searchText.value = ''
  appliedSearchText.value = ''
  resetAnalysisProgress()
  tableKey.value++
}

// 执行搜索
const performSearch = () => {
  appliedSearchText.value = searchText.value
  tableKey.value++
}

// 设置排序
const setSortBy = (field, order) => {
  sortBy.value = field
  sortOrder.value = order
  
  // 同步虚拟表格排序状态
  sortState.value = {
    key: field,
    order: order
  }
  
  tableKey.value++
}

// 处理表格排序变化
const handleSortChange = ({ prop, order }) => {
  if (order === 'ascending') {
    setSortBy(prop, 'asc')
  } else if (order === 'descending') {
    setSortBy(prop, 'desc')
  }
}

// 处理虚拟表格列排序
const handleColumnSort = (sortBy) => {
  sortState.value = sortBy
  setSortBy(sortBy.key, sortBy.order)
}

// 抽屉相关函数已移除

// 工具函数
const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (ms) => {
  if (ms < 1000) return '< 1秒'
  if (ms < 60000) return Math.round(ms / 1000) + '秒'
  if (ms < 3600000) return Math.round(ms / 60000) + '分钟'
  return Math.round(ms / 3600000) + '小时'
}

const formatTTL = (ttl) => {
  if (ttl < 60) return ttl + '秒'
  if (ttl < 3600) return Math.round(ttl / 60) + '分钟'
  if (ttl < 86400) return Math.round(ttl / 3600) + '小时'
  return Math.round(ttl / 86400) + '天'
}

const parseSizeToBytes = (sizeStr) => {
  const match = sizeStr.match(/^(\d+(?:\.\d+)?)\s*(b|kb|mb|gb|tb)?$/i)
  if (!match) return 0
  
  const value = parseFloat(match[1])
  const unit = (match[2] || 'b').toLowerCase()
  
  const multipliers = {
    'b': 1,
    'kb': 1024,
    'mb': 1024 * 1024,
    'gb': 1024 * 1024 * 1024,
    'tb': 1024 * 1024 * 1024 * 1024
  }
  
  return value * (multipliers[unit] || 1)
}

const getSizeClass = (bytes) => {
  if (bytes > 1024 * 1024 * 10) return 'size-xl'
  if (bytes > 1024 * 1024) return 'size-lg'
  if (bytes > 1024 * 100) return 'size-md'
  return 'size-sm'
}



// 抽屉相关的computed属性已移除

// 加载数据库列表
const loadDatabases = async () => {
  loadingDatabases.value = true
  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      return
    }
    
    console.log('开始加载数据库列表...')
    
    // 创建默认的0~15数据库列表
    const defaultDatabases = []
    for (let i = 0; i <= 15; i++) {
      defaultDatabases.push({
        database: i,
        keys: 0,
        expires: 0,
        avgTTL: 0
      })
    }
    
    console.log('默认数据库列表:', defaultDatabases)
    
    const result = await getRedisDatabases({ es_connect: connId })
    
    console.log('后端返回结果:', result)
    
    if (result.code === 0) {
      const backendDatabases = result.data.databases || []
      console.log('后端数据库列表:', backendDatabases)
      
      // 更新默认数据库列表中存在的数据库信息
      backendDatabases.forEach(backendDb => {
        if (backendDb.database <= 15) {
          // 更新0~15范围内的数据库信息
          const defaultDb = defaultDatabases.find(db => db.database === backendDb.database)
          if (defaultDb) {
            defaultDb.keys = backendDb.keys
            defaultDb.expires = backendDb.expires
            defaultDb.avgTTL = backendDb.avgTTL
          }
        } else {
          // 添加大于15的数据库
          defaultDatabases.push(backendDb)
        }
      })
      
      // 按数据库索引排序
      defaultDatabases.sort((a, b) => a.database - b.database)
      
      console.log('最终数据库列表:', defaultDatabases)
      
      databases.value = defaultDatabases
      
      // 设置默认选中的数据库
      if (selectedDatabase.value === 0 || !databases.value.find(db => db.database === selectedDatabase.value)) {
        selectedDatabase.value = 0
      }
      
      console.log('设置后的数据库列表:', databases.value)
      console.log('选中的数据库:', selectedDatabase.value)
    } else {
      // 如果获取失败，至少显示默认的0~15数据库
      databases.value = defaultDatabases
      selectedDatabase.value = 0
      console.warn('获取数据库列表失败，使用默认数据库列表:', result.msg)
    }
  } catch (error) {
    console.error('获取数据库列表失败:', error)
    
    // 如果出现错误，至少显示默认的0~15数据库
    const defaultDatabases = []
    for (let i = 0; i <= 15; i++) {
      defaultDatabases.push({
        database: i,
        keys: 0,
        expires: 0,
        avgTTL: 0
      })
    }
    databases.value = defaultDatabases
    selectedDatabase.value = 0
    
    ElMessage.error('获取数据库列表失败: ' + (error.message || '未知错误'))
  } finally {
    loadingDatabases.value = false
  }
}

// 生命周期
onMounted(() => {
  loadDatabases()
})
</script>

<style scoped>
.memory-analysis-page {
  padding: 20px;
  transition: all 0.3s;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.header-left h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #303133;
}

.header-left h2 i {
  color: #409eff;
}

.stats-info {
  display: flex;
  gap: 15px;
  font-size: 14px;
  color: #666;
}

.stat-item {
  background: #f5f7fa;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 15px;
  flex-wrap: wrap;
}

.database-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.database-selector::before {
  content: "数据库:";
  font-size: 14px;
  color: #606266;
  white-space: nowrap;
}

.mobile-filters {
  flex-direction: column;
  gap: 10px;
}

.table-card {
  margin-bottom: 20px;
}

.table-header {
  margin-bottom: 16px;
}

.table-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.table-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.sort-controls {
  display: flex;
  gap: 8px;
}

.analysis-progress {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  margin-top: 16px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  color: #606266;
}

.progress-percent {
  font-weight: 600;
  color: #409eff;
}

.progress-details {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.analysis-result {
  background: #f0f9ff;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #b3d8ff;
  margin-top: 16px;
}

.result-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: #409eff;
}

.result-time {
  font-weight: 500;
  color: #67c23a;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

.table-wrapper {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e4e7ed;
  height: calc(100vh - 400px);
  min-height: 400px;
  max-height: 800px;
}

/* 虚拟表格样式 */
.key-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0 8px;
  line-height: 40px;
  width: 100%;
}

.key-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.copy-btn {
  flex-shrink: 0;
  opacity: 0.7;
  transition: opacity 0.2s ease;
  color: #409eff !important;
  cursor: pointer !important;
  font-size: 12px;
  user-select: none;
  padding: 2px 6px !important;
  margin-left: 8px;
  border: 1px solid #409eff !important;
  border-radius: 4px !important;
  background-color: transparent !important;
}

.copy-btn:hover {
  color: #79bbff !important;
  background-color: #f0f9ff !important;
  border-color: #79bbff !important;
  opacity: 1 !important;
}

.key-cell:hover .copy-btn {
  opacity: 1 !important;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.action-buttons .el-button {
  margin: 0;
  padding: 4px 8px;
}

/* 移动端虚拟表格优化 */
.mobile-layout .key-cell {
  line-height: 32px;
  font-size: 12px;
}

.mobile-layout .key-text {
  font-size: 12px;
}

.mobile-layout .copy-btn {
  padding: 1px 4px !important;
  font-size: 11px;
  opacity: 0.8 !important;
}

.mobile-layout .action-buttons .el-button {
  padding: 2px 6px;
  font-size: 12px;
}

.size-value {
  font-weight: 500;
}

.size-large {
  color: #f56c6c;
}

.size-medium {
  color: #e6a23c;
}

.size-small {
  color: #67c23a;
}

.ttl-never {
  color: #909399;
  font-size: 12px;
}

.ttl-value {
  color: #409eff;
  font-weight: 500;
}

.ttl-unknown {
  color: #c0c4cc;
}

.table-footer {
  margin-top: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-summary {
  font-size: 14px;
  color: #606266;
  display: flex;
  gap: 16px;
}

/* 暗色模式 */
.dark-mode {
  background: #1a1a1a;
  color: #e5e5e5;
}

.dark-mode .header-left h2 {
  color: #e5e5e5;
}

.dark-mode .stat-item {
  background: #2d2d2d;
  color: #c0c4cc;
}

.dark-mode .table-title {
  color: #e5e5e5;
}

.dark-mode .analysis-progress {
  background: #2d2d2d;
  border-color: #4a4a4a;
}

.dark-mode .progress-percent {
  color: #79bbff;
}

.dark-mode .progress-details {
  color: #909399;
}

.dark-mode .key-cell {
  color: #e5e5e5;
}

.dark-mode .copy-btn {
  color: #79bbff !important;
  border-color: #79bbff !important;
  background-color: transparent !important;
}

.dark-mode .copy-btn:hover {
  color: #a0cfff !important;
  background-color: #2d2d2d !important;
  border-color: #a0cfff !important;
}

.dark-mode .progress-info {
  color: #c0c4cc;
}

.dark-mode .table-wrapper {
  border-color: #4a4a4a;
}

.dark-mode .stats-summary {
  color: #c0c4cc;
}

.dark-mode .analysis-result {
  background: #2d2d2d;
  border-color: #4a4a4a;
}

.dark-mode .result-info {
  color: #79bbff;
}

.dark-mode .result-time {
  color: #95d475;
}

/* 抽屉相关样式已移除 */

/* 搜索容器样式 */
.search-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-container .el-input {
  flex: 1;
}

.search-help h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

.search-help ul {
  margin: 0;
  padding: 0 0 0 16px;
  list-style-type: disc;
}

.search-help li {
  margin: 6px 0;
  font-size: 13px;
  color: #606266;
}

.search-help code {
  background: #f5f7fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: #e6a23c;
}

/* 移动端适配 */
.mobile-layout {
  padding: 10px;
}

.mobile-layout .page-header {
  flex-direction: column;
  gap: 15px;
}

.mobile-layout .header-left h2 {
  font-size: 20px;
  text-align: center;
}

.mobile-layout .stats-info {
  justify-content: center;
  flex-wrap: wrap;
}

.mobile-layout .table-controls {
  flex-direction: column;
  gap: 12px;
}

.mobile-layout .sort-controls {
  flex-wrap: wrap;
}

.mobile-layout .stats-summary {
  flex-direction: column;
  gap: 8px;
}

.mobile-layout .search-container {
  width: 100%;
}

.mobile-layout .search-container .el-input {
  width: 100% !important;
}

/* 响应式表格 */
@media (max-width: 768px) {
  .memory-analysis-page {
    padding: 10px;
  }

  .table-wrapper {
    font-size: 12px;
  }

  .el-table .cell {
    padding: 0 5px;
  }
}

@media (max-width: 480px) {
  .memory-analysis-page {
    padding: 8px;
  }

  .header-left h2 {
    font-size: 18px;
  }

  .stats-info {
    font-size: 12px;
  }
}
</style>
