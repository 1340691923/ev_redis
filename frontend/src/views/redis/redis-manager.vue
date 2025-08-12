<template>
  <div class="redis-manager-page" :class="{ 'mobile-layout': isMobile, 'dark-mode': isDarkMode }">
    <!-- 桌面端布局 -->
    <div v-if="!isMobile" class="desktop-layout">
      <el-container class="manager-container">
        <el-aside class="keys-sidebar" :width="sidebarWidth + 'px'">
          <div class="search-container">
            <div class="search-row">
              <el-select v-model="selectedDatabase" placeholder="选择数据库"  style="width: 120px;" @change="onDatabaseChange" :loading="loadingDatabases">
                <el-option v-for="db in databases" :key="db.database" :label="`DB${db.database} (${db.keys})`" :value="db.database" />
              </el-select>
                             <el-input
                 v-model="searchPattern"
                 placeholder="Key"
                 prefix-icon="el-icon-search"
                 style="flex: 1;"
                 @keyup.enter="handleSearch"
                 clearable
               >
           
               </el-input>
               <el-button 
                 type="primary" 
                 @click="handleSearch"
                 :loading="loadingKeys"
                 style="margin-left: 8px;"
               >
                 <i class="el-icon-search"></i>
                 搜索
               </el-button>
               <el-button type="primary"  @click="openNewKeyTab"><i class="el-icon-plus"></i>新增</el-button>
            </div>
          </div>
          <!-- 分页组件 -->
          <div v-if="allKeysList.length > 0" class="pagination-container">
            <div class="pagination-info">
              <span>共 {{ allKeysList.length }} 条记录</span>
              <span>第 {{ currentPage }} / {{ totalPages }} 页</span>
            </div>
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[20, 50, 100]"
              :total="allKeysList.length"
              layout="sizes, prev, pager, next"
              @current-change="handlePageChange"
              @size-change="handlePageSizeChange"
              small
            />
          </div>
          <div class="keys-container" v-loading="loadingKeys">
            
            <div class="keys-list">
              <!-- 虚拟滚动列表 -->
              <div
                v-if="filteredKeys.length > 0"
                class="virtual-container"
                ref="virtualContainer"
                @scroll="handleVirtualScroll"
              >
                <div
                  class="virtual-content"
                  :style="{ height: totalHeight + 'px' }"
                >
                  <div
                    v-for="(item, index) in visibleKeys"
                    :key="item.key"
                    class="key-item virtual-key-item"
                    :style="{
                      transform: `translateY(${item.top}px)`,
                      position: 'absolute',
                      width: '100%'
                    }"
                    @click="handleKeyClick(item)"
                  >
                    <div class="key-info">
                      <i :class="getKeyIcon(item)" class="key-icon"></i>
                      <span class="key-name">{{ item.key }}</span>
                      <el-button 
                        type="text" 
                        size="small"
                        class="copy-btn"
                        title="复制Key"
                        @click.stop="handleCopyClick(item.key)"
                      >
                        复制
                      </el-button>
                    </div>
                    <div class="key-actions">
                      <el-button type="text"  @click.stop="deleteKey(item)">
                        <i class="el-icon-delete"></i>
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 空状态 -->
              <div v-if="keysList.length === 0 && !loadingKeys" class="empty-keys">
                <i class="el-icon-document empty-icon"></i>
                <p>暂无Key数据</p>
              </div>
            </div>
          </div>
          <div class="resize-handle" @mousedown="startResize"></div>
        </el-aside>
        <el-main class="content-main">
          <el-tabs v-model="activeTab" type="card" closable @tab-remove="removeTab" @tab-click="onTabClick" class="content-tabs">
            <el-tab-pane v-for="tab in tabs" :key="tab.name" :name="tab.name" :label="tab.label" :closable="tab.closable">
              <template #label>
                <span
                  class="tab-label"
                  @contextmenu.prevent="showTabContextMenu($event, tab)"
                >
                  <i :class="getTabIcon(tab)"></i>{{ tab.label }}
                </span>
              </template>
              <KeyDetail v-if="tab.type === 'key'" :key-data="tab.keyData" :database="selectedDatabase" @refresh="refreshKey" @save="saveKey" @delete="deleteKeyFromTab" />
              <KeyCreator v-else-if="tab.type === 'new'" :database="selectedDatabase" @save="createKey" @cancel="removeTab(tab.name)" />
              <div v-else-if="tab.type === 'welcome'" class="welcome-content">
                <div class="welcome-center">
                  <i class="el-icon-folder-opened welcome-icon"></i>
                  <h2>Key 管理器</h2>
                  <p>点击左侧的Key查看详情，或创建新的Key</p>
                  <div class="welcome-actions">
                    <el-button type="primary" @click="openNewKeyTab"><i class="el-icon-plus"></i>创建新Key</el-button>
                    <el-button @click="refreshAllData"><i class="el-icon-refresh"></i>刷新数据</el-button>
                  </div>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
          <div v-if="tabs.length === 0" class="empty-state">
            <div class="empty-center">
              <i class="el-icon-document-add empty-icon"></i>
              <h3>暂无打开的标签页</h3>
              <p>点击左侧的Key或创建新Key开始管理</p>
              <el-button type="primary" @click="openWelcomeTab"><i class="el-icon-view"></i>打开欢迎页</el-button>
            </div>
          </div>
        </el-main>
      </el-container>

      <!-- Tab右键菜单 -->
      <div
        v-if="contextMenuVisible"
        class="tab-context-menu"
        :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
        @click.stop
      >
        <div class="context-menu-item" @click="closeOtherTabs">
          <i class="el-icon-close"></i>
          <span>关闭其他</span>
        </div>
        <div class="context-menu-item" @click="closeRightTabs">
          <i class="el-icon-arrow-right"></i>
          <span>关闭右边</span>
        </div>
        <div class="context-menu-item" @click="closeAllTabs">
          <i class="el-icon-delete"></i>
          <span>关闭所有</span>
        </div>
      </div>
    </div>
    <!-- 移动端布局 -->
    <div v-else class="mobile-layout">
      <div class="mobile-container">
        <div class="mobile-search-section">
          <div class="mobile-search-header">
            <h3>Key 管理器</h3>
            <el-button type="primary"  @click="openNewKeyTab" class="mobile-new-btn">
              <i class="el-icon-plus"></i>
              新增
            </el-button>
          </div>
          <div class="mobile-search-controls">
            <el-select v-model="selectedDatabase" placeholder="选择数据库"  style="width: 100px;" @change="onDatabaseChange" :loading="loadingDatabases">
              <el-option v-for="db in databases" :key="db.database" :label="`DB${db.database}`" :value="db.database" />
            </el-select>
                         <el-input
               v-model="searchPattern"
               placeholder="搜索Key..."
               prefix-icon="el-icon-search"
               style="flex: 1; margin-left: 8px;"
               @keyup.enter="handleSearch"
               clearable
             >
               <template #suffix>
                 <el-icon v-if="loadingKeys" class="is-loading"><Loading /></el-icon>
               </template>
             </el-input>
             <el-button 
               type="primary" 
               @click="handleSearch"
               :loading="loadingKeys"
               style="margin-left: 8px;"
               size="small"
             >
               <i class="el-icon-search"></i>
             </el-button>
          </div>
          <div class="mobile-key-stats">
            <span>{{ displayKeysCount }}</span>
          </div>
        </div>
        <div class="mobile-content-section">
          <div v-if="!activeTab || (!getTabByName(activeTab) || (getTabByName(activeTab)?.type !== 'key' && getTabByName(activeTab)?.type !== 'new'))" class="mobile-keys-list">
            <!-- 移动端分页 -->
            <div v-if="allKeysList.length > 0" class="mobile-pagination">
              <div class="mobile-pagination-info">
                共 {{ allKeysList.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页
              </div>
              <el-pagination
                v-model:current-page="currentPage"
                v-model:page-size="pageSize"
                :page-sizes="[20, 50]"
                :total="allKeysList.length"
                layout="sizes, prev, pager, next"
                @current-change="handlePageChange"
                @size-change="handlePageSizeChange"
                small
              />
            </div>
            <div class="keys-list" v-loading="loadingKeys">
              <div v-for="key in filteredKeys" :key="key.key" class="mobile-key-item" @click="handleKeyClick(key)">
                <div class="key-info">
                  <i :class="getKeyIcon(key)" class="key-icon"></i>
                  <span class="key-name">{{ key.key }}</span>
                  <el-button 
                    type="text" 
                    size="small"
                    class="copy-btn"
                    title="复制Key"
                    @click.stop="handleCopyClick(key.key)"
                  >
                    复制
                  </el-button>
                </div>
                <div class="key-actions">
                  <el-button type="text"  @click.stop="deleteKey(key)" style="color: #f56c6c;"><i class="el-icon-delete"></i></el-button>
                </div>
              </div>
              <div v-if="keysList.length === 0 && !loadingKeys" class="empty-keys">
                <i class="el-icon-document empty-icon"></i>
                <p>暂无Key数据</p>
                <el-button type="primary" @click="openNewKeyTab" style="margin-top: 16px;">
                  <i class="el-icon-plus"></i>
                  创建第一个Key
                </el-button>
              </div>
            </div>
          </div>
          <div v-else-if="getTabByName(activeTab)?.type === 'key'" class="mobile-key-detail">
            <div class="mobile-detail-header">
              <el-button type="text" @click="backToKeysList" class="back-btn"><i class="el-icon-arrow-left"></i>返回</el-button>
              <span class="detail-title">{{ getTabByName(activeTab)?.label }}</span>
              <el-button type="text" @click="refreshCurrentTab" class="refresh-btn"><i class="el-icon-refresh"></i></el-button>
            </div>
            <div class="mobile-detail-content">
              <KeyDetail :key-data="getTabByName(activeTab)?.keyData" :database="selectedDatabase" @refresh="refreshKey" @save="saveKey" @delete="deleteKeyFromTab" />
            </div>
          </div>
          <div v-else-if="getTabByName(activeTab)?.type === 'new'" class="mobile-key-creator">
            <div class="mobile-detail-header">
              <el-button type="text" @click="backToKeysList" class="back-btn"><i class="el-icon-arrow-left"></i>返回</el-button>
              <span class="detail-title">新增Key</span>
            </div>
            <div class="mobile-detail-content">
              <KeyCreator :database="selectedDatabase" @save="createKey" @cancel="backToKeysList" />
            </div>
          </div>
        </div>
      </div>

      <!-- 移动端浮动新增按钮 -->
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Loading } from '@element-plus/icons-vue';
import { sdk } from '@elasticview/plugin-sdk'
import { getRedisDatabases, deleteRedisKey, getRedisKeyDetail, searchRedisKeys } from "@/api/redis";
import KeyDetail from "./components/KeyDetail.vue";
import KeyCreator from "./components/KeyCreator.vue";

// 响应式数据
const sidebarWidth = ref(400)  // PC端默认增加宽度，便于查看key名称
const selectedDatabase = ref(0)
const searchPattern = ref('')
const loadingDatabases = ref(false)
const loadingKeys = ref(false)
const totalKeysCount = ref(0)
const matchedKeysCount = ref(0)

// 前端分页相关
const allKeysList = ref([]) // 存储所有keys
const currentPage = ref(1)
const pageSize = ref(10)  // 改为10个key为一页，更适合查看
const totalPages = ref(0)
const databases = ref([])
const keysList = ref([])
const filteredKeys = ref([])
const activeTab = ref('')
const tabs = ref([])

// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuTab = ref(null)

// 虚拟滚动相关
const virtualContainer = ref(null)
const itemHeight = 48 // 每个key项的高度（已更新以匹配新样式）
const visibleCount = ref(20) // 可见项数量
const scrollTop = ref(0)
const visibleKeys = ref([])
const totalHeight = ref(0)

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

let tabIndex = 0

// 计算属性
const isMobile = computed(() => sdk.IsMobile())
const isDarkMode = computed(() => sdk.isDarkTheme())
const displayKeysCount = computed(() => {
  // 如果有搜索条件，显示匹配的总数量
  if (searchPattern.value && searchPattern.value.trim() && matchedKeysCount.value > 0) {
    return `${matchedKeysCount.value} 个匹配的Key`
  }
  // 没有搜索条件时显示当前加载的数量
  return `${keysList.value.length} 个Key`
})

// 虚拟滚动计算属性
const updateVisibleKeys = () => {
  if (!filteredKeys.value.length) {
    visibleKeys.value = []
    totalHeight.value = 0
    return
  }

  totalHeight.value = filteredKeys.value.length * itemHeight

  const containerHeight = virtualContainer.value?.clientHeight || 400
  const bufferedCount = Math.ceil(containerHeight / itemHeight) + 5 // 增加缓冲区

  const startIndex = Math.max(0, Math.floor(scrollTop.value / itemHeight) - 2)
  const endIndex = Math.min(filteredKeys.value.length, startIndex + bufferedCount)

  visibleKeys.value = filteredKeys.value.slice(startIndex, endIndex).map((item, index) => ({
    ...item,
    top: (startIndex + index) * itemHeight
  }))
}

// 虚拟滚动方法
const handleVirtualScroll = (e) => {
  scrollTop.value = e.target.scrollTop
  updateVisibleKeys()
}

// 监听filteredKeys变化
watch(filteredKeys, () => {
  nextTick(() => updateVisibleKeys())
}, { immediate: true })

// 生命周期
onMounted(async () => {
  await loadDatabases()
  await loadKeys()
  openWelcomeTab()
})

// 数据库相关方法
const loadDatabases = async () => {
  loadingDatabases.value = true
  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      return
    }
    
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
    
    const res = await getRedisDatabases({ es_connect: connId })
    
    if (res.code === 0) {
      const backendDatabases = res.data.databases || []
      
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
      
      databases.value = defaultDatabases
      
      // 设置默认选中的数据库
      if (selectedDatabase.value === 0 || !databases.value.find(db => db.database === selectedDatabase.value)) {
        selectedDatabase.value = 0
      }
    } else {
      // 如果获取失败，至少显示默认的0~15数据库
      databases.value = defaultDatabases
      selectedDatabase.value = 0
      console.warn('获取数据库列表失败，使用默认数据库列表:', res.msg)
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
    
    ElMessage.error('获取数据库列表失败: ' + error.message)
  } finally {
    loadingDatabases.value = false
  }
}

// Key列表相关方法
const loadKeys = async () => {
  loadingKeys.value = true
  currentPage.value = 1 // 重置页码
  
  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      return
    }

    // Key管理器始终使用轻量级搜索，不执行MEMORY USAGE命令
    const searchText = searchPattern.value && searchPattern.value.trim() ? searchPattern.value.trim() : ''
    
    // 统一使用搜索接口，空搜索文本会返回所有keys
    const res = await searchRedisKeys({
      es_connect: connId,
      database: selectedDatabase.value,
      search_text: searchText, // 空字符串表示获取所有keys
      count: 999999, // 请求所有匹配的keys
      cursor: '0',
      case_sensitive: false // 默认不区分大小写
    })

    if (res.code === 0) {
      const allKeys = res.data.keys || []
      allKeysList.value = allKeys // 存储所有keys
      totalKeysCount.value = res.data.totalCount || 0

      // 更新匹配的总数量
      if (searchText) {
        matchedKeysCount.value = res.data.matchedCount || allKeys.length
        // 搜索结果提示
        ElMessage.success(`搜索完成，找到 ${allKeys.length} 个匹配的Key`)
      } else {
        matchedKeysCount.value = 0 // 没有搜索时重置
      }

      // 计算总页数
      totalPages.value = Math.ceil(allKeys.length / pageSize.value)
      
      // 更新当前页的keys
      updateCurrentPageKeys()
    } else {
      ElMessage.error(res.msg || '获取Keys失败')
    }
  } catch (error) {
    console.error('获取Keys失败:', error)
    ElMessage.error('获取Keys失败: ' + error.message)
  } finally {
    loadingKeys.value = false
  }
}

// 更新当前页显示的keys
const updateCurrentPageKeys = () => {
  const startIndex = (currentPage.value - 1) * pageSize.value
  const endIndex = startIndex + pageSize.value
  keysList.value = allKeysList.value.slice(startIndex, endIndex)
  filteredKeys.value = keysList.value
  
  // 更新虚拟滚动
  setTimeout(() => updateVisibleKeys(), 0)
}

// 页码变化
const handlePageChange = (page) => {
  currentPage.value = page
  updateCurrentPageKeys()
}

// 每页数量变化
const handlePageSizeChange = (size) => {
  pageSize.value = size
  totalPages.value = Math.ceil(allKeysList.value.length / pageSize.value)
  currentPage.value = 1 // 重置到第一页
  updateCurrentPageKeys()
}

const onDatabaseChange = () => {
  // 清空搜索状态
  searchPattern.value = ''
  matchedKeysCount.value = 0
  currentPage.value = 1
  // 重新加载keys
  loadKeys()
}

const handleSearch = () => {
  loadKeys()
}

const clearSearch = () => {
  searchPattern.value = ''
  matchedKeysCount.value = 0
  currentPage.value = 1
  loadKeys()
}

// Key操作相关方法
const handleKeyClick = async (key) => await openKeyTab(key)

const openKeyTab = async (keyData) => {
  const keyName = keyData?.keyName || keyData?.key
  if (!keyName) {
    ElMessage.error('Key数据不正确')
    return
  }
  const tabName = `key_${keyName}_${Date.now()}`
  const existingTab = tabs.value.find(tab => tab.type === 'key' && tab.keyData && (tab.keyData.keyName === keyName || tab.keyData.key === keyName))
  if (existingTab) {
    activeTab.value = existingTab.name
    return
  }
  try {
    const connId = sdk.GetSelectEsConnID()
    const res = await getRedisKeyDetail({
      es_connect: connId,
      database: selectedDatabase.value,
      key: keyName
    })
    if (res.code === 0) {
      const newTab = {
        name: tabName,
        label: keyName && keyName.length > 20 ? keyName.substring(0, 20) + '...' : keyName,
        type: 'key',
        closable: true,
        keyData: { ...keyData, keyName: keyName, key: keyName, detail: res.data }
      }
      tabs.value.push(newTab)
      activeTab.value = tabName
    } else {
      ElMessage.error('获取Key详情失败: ' + res.msg)
    }
  } catch (error) {
    console.error('获取Key详情失败:', error)
    ElMessage.error('获取Key详情失败: ' + error.message)
  }
}

const deleteKey = async (keyData) => {
  const keyName = keyData?.keyName || keyData?.key
  if (!keyName) {
    ElMessage.error('Key数据不正确')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除Key "${keyName}" 吗？此操作不可恢复！`, '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const connId = sdk.GetSelectEsConnID()
    const res = await deleteRedisKey({
      es_connect: connId,
      database: selectedDatabase.value,
      key: keyName
    })
    if (res.code === 0 && res.data.success) {
      ElMessage.success('删除成功')
      loadKeys()
      const relatedTab = tabs.value.find(tab => tab.type === 'key' && tab.keyData && (tab.keyData.keyName === keyName || tab.keyData.key === keyName))
      if (relatedTab) {
        removeTab(relatedTab.name)
      }
    } else {
      ElMessage.error(res.data?.message || res.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除Key失败:', error)
      ElMessage.error('删除Key失败: ' + error.message)
    }
  }
}

// 标签页相关方法
const openNewKeyTab = () => {
  console.log('openNewKeyTab 被调用') // 调试信息
  const tabName = `new_key_${++tabIndex}`
  const newTab = { name: tabName, label: '新增Key', type: 'new', closable: true }
  tabs.value.push(newTab)
  activeTab.value = tabName
  console.log('新标签页已创建:', newTab) // 调试信息
}

const openWelcomeTab = () => {
  const welcomeTab = tabs.value.find(tab => tab.type === 'welcome')
  if (!welcomeTab) {
    const newTab = { name: 'welcome', label: '欢迎', type: 'welcome', closable: false }
    tabs.value.push(newTab)
  }
  activeTab.value = 'welcome'
}

const removeTab = (targetName) => {
  const targetIndex = tabs.value.findIndex(tab => tab.name === targetName)
  if (targetIndex === -1) return
  tabs.value.splice(targetIndex, 1)
  if (activeTab.value === targetName) {
    if (tabs.value.length > 0) {
      const nextIndex = targetIndex >= tabs.value.length ? tabs.value.length - 1 : targetIndex
      activeTab.value = tabs.value[nextIndex].name
    } else {
      activeTab.value = ''
    }
  }
}

const onTabClick = (tab) => { activeTab.value = tab.name }

// 右键菜单相关方法
const showTabContextMenu = (event, tab) => {
  contextMenuTab.value = tab
  contextMenuPosition.value = {
    x: event.clientX,
    y: event.clientY
  }
  contextMenuVisible.value = true

  // 点击其他地方隐藏菜单
  const hideMenu = () => {
    contextMenuVisible.value = false
    document.removeEventListener('click', hideMenu)
  }
  setTimeout(() => {
    document.addEventListener('click', hideMenu)
  }, 100)
}

const closeOtherTabs = () => {
  if (!contextMenuTab.value) return

  const keepTab = contextMenuTab.value
  const closableTabs = tabs.value.filter(tab =>
    tab.name !== keepTab.name && tab.closable !== false
  )

  // 删除其他可关闭的标签页
  closableTabs.forEach(tab => {
    const index = tabs.value.findIndex(t => t.name === tab.name)
    if (index !== -1) {
      tabs.value.splice(index, 1)
    }
  })

  // 设置当前标签页为活动标签页
  activeTab.value = keepTab.name
  contextMenuVisible.value = false

  ElMessage.success(`已关闭其他 ${closableTabs.length} 个标签页`)
}

const closeRightTabs = () => {
  if (!contextMenuTab.value) return

  const currentTabIndex = tabs.value.findIndex(tab => tab.name === contextMenuTab.value.name)
  if (currentTabIndex === -1) return

  const rightTabs = tabs.value.slice(currentTabIndex + 1).filter(tab => tab.closable !== false)

  // 删除右边的可关闭标签页
  rightTabs.forEach(tab => {
    const index = tabs.value.findIndex(t => t.name === tab.name)
    if (index !== -1) {
      tabs.value.splice(index, 1)
    }
  })

  contextMenuVisible.value = false

  if (rightTabs.length > 0) {
    ElMessage.success(`已关闭右边 ${rightTabs.length} 个标签页`)
  } else {
    ElMessage.info('右边没有可关闭的标签页')
  }
}

const closeAllTabs = async () => {
  try {
    await ElMessageBox.confirm('确定要关闭所有标签页吗？', '关闭确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const closableTabs = tabs.value.filter(tab => tab.closable !== false)
    const closableCount = closableTabs.length

    // 删除所有可关闭的标签页
    tabs.value = tabs.value.filter(tab => tab.closable === false)

    // 重置活动标签页
    if (tabs.value.length > 0) {
      activeTab.value = tabs.value[0].name
    } else {
      activeTab.value = ''
    }

    contextMenuVisible.value = false

    ElMessage.success(`已关闭所有 ${closableCount} 个标签页`)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('关闭标签页失败:', error)
    }
  }
}

// 工具方法
const getKeyIcon = (key) => {
  const typeIcons = {
    'string': 'el-icon-document',
    'hash': 'el-icon-collection',
    'list': 'el-icon-menu',
    'set': 'el-icon-postcard',
    'zset': 'el-icon-rank'
  }
  return typeIcons[key.type] || 'el-icon-document'  // 搜索时type为空，统一显示文档图标
}

const getTabIcon = (tab) => {
  if (tab.type === 'welcome') return 'el-icon-house'
  if (tab.type === 'new') return 'el-icon-plus'
  if (tab.type === 'key') {
    const typeIcons = {
      'string': 'el-icon-document',
      'hash': 'el-icon-collection',
      'list': 'el-icon-menu',
      'set': 'el-icon-postcard',
      'zset': 'el-icon-rank'
    }
    return typeIcons[tab.keyData?.keyType] || 'el-icon-document'
  }
  return 'el-icon-document'
}

const getTabByName = (name) => tabs.value.find(tab => tab.name === name)

// 移动端专用方法
const backToKeysList = () => { activeTab.value = '' }
const refreshCurrentTab = () => {
  const currentTab = getTabByName(activeTab.value)
  if (currentTab && currentTab.type === 'key') {
    refreshKey(currentTab.keyData.key)
  }
}

const refreshKey = async (keyName) => {
  console.log('refreshKey called in redis-manager, keyName:', keyName)
  const tab = tabs.value.find(tab => tab.type === 'key' && tab.keyData && (tab.keyData.keyName === keyName || tab.keyData.key === keyName))
  if (tab) {
    try {
      const connId = sdk.GetSelectEsConnID()
      const res = await getRedisKeyDetail({
        es_connect: connId,
        database: selectedDatabase.value,
        key: keyName
      })
      if (res.code === 0) {
        // 更新现有标签页的数据
        tab.keyData = { ...tab.keyData, keyName: keyName, key: keyName, detail: res.data }
        // 强制重新渲染组件
        const currentTabName = tab.name
        activeTab.value = ''
        await nextTick()
        activeTab.value = currentTabName
        ElMessage.success('Key数据已刷新')
      } else {
        ElMessage.error('刷新Key详情失败: ' + res.msg)
      }
    } catch (error) {
      console.error('刷新Key详情失败:', error)
      ElMessage.error('刷新Key详情失败: ' + error.message)
    }
  }
}

const saveKey = () => loadKeys()
const createKey = () => {
  loadKeys()
  removeTab(activeTab.value)
}
const deleteKeyFromTab = () => {
  loadKeys()
  removeTab(activeTab.value)
}
const refreshAllData = () => {
  loadDatabases()
  loadKeys()
}

// 拖拽调整侧边栏宽度
const startResize = (e) => {
  const startX = e.clientX
  const startWidth = sidebarWidth.value
  const onMouseMove = (e) => {
    const deltaX = e.clientX - startX
    const newWidth = startWidth + deltaX
    if (newWidth >= 200 && newWidth <= 600) {
      sidebarWidth.value = newWidth
    }
  }
  const onMouseUp = () => {
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

// 复制按钮点击处理方法
const handleCopyClick = (key) => {
  console.log('复制按钮被点击，key:', key)
  copyToClipboard(key)
}
</script>

<style scoped>
.redis-manager-page {
  height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.desktop-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.manager-container { height: 100%; }

.keys-sidebar {
  background: white;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  position: relative;
}

.search-container {
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-row .el-button {
  flex-shrink: 0;
}

.keys-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.keys-header {
  padding: 14px 16px;
  border-bottom: 2px solid #e1f3ff; /* 更明显的底边框 */
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #f8fbff 0%, #e8f4fd 100%); /* 渐变背景 */
  border-radius: 6px 6px 0 0; /* 顶部圆角 */
  margin: 8px 8px 0 8px; /* 与外边框对齐 */
  box-shadow: 0 1px 3px rgba(64, 158, 255, 0.1); /* 淡蓝色阴影 */
}

.keys-count { font-size: 12px; color: #909399; }

.keys-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.search-status {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
}

.keys-list {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 200px; /* 设置合适的最小高度 */
  background: #fafcff; /* 淡蓝色背景，使列表区域更明显 */
  border: 1px solid #e1f3ff; /* 淡蓝色边框 */
  border-radius: 6px;
  margin: 8px;
}

.virtual-container {
  flex: 1;
  overflow-y: auto;
  position: relative;
  background: white; /* 确保内容区域是白色背景 */
  border-radius: 4px;
  margin: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05); /* 添加轻微阴影 */
}

.virtual-content {
  position: relative;
  width: 100%;
}

.virtual-key-item {
  padding: 14px 16px; /* 增加内边距 */
  border-bottom: 1px solid #e8f4fd; /* 更明显的分割线 */
  background: white;
  box-sizing: border-box;
  height: 48px; /* 增加高度 */
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent; /* 左边框用于悬停效果 */
}

.virtual-key-item:hover {
  background-color: #f0f9ff; /* 更明显的悬停背景 */
  border-left-color: #409eff; /* 悬停时显示蓝色左边框 */
  transform: translateX(2px); /* 轻微的位移效果 */
}

.virtual-key-item .key-info {
  display: flex;
  align-items: center;
  flex: 1;
  overflow: hidden;
}

.virtual-key-item .key-actions {
  opacity: 0.6;
  transition: opacity 0.2s;
}

.virtual-key-item:hover .key-actions {
  opacity: 1;
}

.key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background-color 0.2s;
  min-height: 44px;
}

.key-item:hover { background-color: #f5f7fa; }

.key-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
}

.key-icon {
  color: #409eff; /* 更明显的蓝色图标 */
  font-size: 16px; /* 稍大的图标 */
  flex-shrink: 0;
  margin-right: 2px;
}

.key-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  color: #303133; /* 深色文本，更明显 */
  font-weight: 500; /* 稍粗的字体 */
  letter-spacing: 0.5px; /* 字母间距 */
  flex: 1;
  margin-right: 8px;
}

.copy-btn {
  flex-shrink: 0;
  opacity: 0;
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
}

.key-info:hover .copy-btn {
  opacity: 1 !important;
}

.key-actions {
  opacity: 0;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.key-item:hover .key-actions { opacity: 1; }

.empty-keys {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #409eff; /* 蓝色主题色 */
  text-align: center;
  background: linear-gradient(135deg, #f8fbff 0%, #e8f4fd 100%); /* 与header一致的渐变 */
  border-radius: 8px;
  margin: 20px;
  border: 2px dashed #c6e2ff; /* 虚线边框 */
  min-height: 300px; /* 确保有足够高度 */
}

.empty-icon {
  font-size: 64px; /* 更大的图标 */
  margin-bottom: 16px;
  opacity: 0.8;
  color: #409eff; /* 蓝色图标 */
}

.pagination-container {
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.pagination-info {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 16px;
}

.resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  background: transparent;
}

.resize-handle:hover { background: #409eff; }

.content-main {
  padding: 0;
  background: white;
  height: 100%;
  overflow: hidden;
}

.content-tabs { height: 100%; }

.content-tabs :deep(.el-tabs__content) {
  height: calc(100% - 40px);
  overflow: auto;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.welcome-content {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-center { text-align: center; color: #606266; }

.welcome-icon {
  font-size: 64px;
  color: #409eff;
  margin-bottom: 16px;
}

.welcome-center h2 {
  margin: 0 0 12px 0;
  color: #303133;
}

.welcome-center p {
  margin: 0 0 24px 0;
  color: #909399;
}

.welcome-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-center { text-align: center; color: #606266; }

.empty-center .empty-icon {
  font-size: 64px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.empty-center h3 {
  margin: 0 0 12px 0;
  color: #303133;
}

.empty-center p {
  margin: 0 0 20px 0;
  color: #909399;
}

/* 移动端布局 */
.mobile-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.mobile-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.mobile-search-section {
  background: white;
  border-bottom: 1px solid #e4e7ed;
  padding: 12px 16px;
  flex-shrink: 0;
}

.mobile-search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.mobile-search-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.mobile-new-btn {
  background: #409eff;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.mobile-search-controls {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.mobile-search-controls .el-button {
  flex-shrink: 0;
}

.mobile-key-stats {
  font-size: 12px;
  color: #909399;
  text-align: center;
}

.total-count {
  color: #606266;
  font-weight: 500;
}

.mobile-content-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.mobile-keys-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.mobile-key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background-color 0.2s;
  min-height: 48px;
  -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
}

.mobile-key-item:hover { background-color: #f5f7fa; }
.mobile-key-item .key-actions { opacity: 1; }

.mobile-key-item .copy-btn {
  font-size: 11px;
  opacity: 0.8;
  padding: 1px 4px !important;
}

.mobile-key-item:hover .copy-btn {
  opacity: 1 !important;
}

.mobile-pagination {
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.mobile-pagination-info {
  font-size: 12px;
  color: #909399;
  text-align: center;
}

.mobile-add-key-btn {
  background: #409eff !important;
  border-color: #409eff !important;
  font-weight: 500;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.3);
}

/* 移动端浮动新增按钮 */
.mobile-fab {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 1000;
}

.fab-button {
  width: 56px;
  height: 56px;
  background: #409eff !important;
  border-color: #409eff !important;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
  font-size: 24px;
  transition: all 0.3s ease;
}

.fab-button:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.6);
}

.mobile-key-detail,
.mobile-key-creator {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.mobile-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: white;
  flex-shrink: 0;
}

.back-btn,
.refresh-btn {
  color: #409eff;
  font-size: 16px;
  padding: 8px;
}

.detail-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  flex: 1;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-detail-content {
  flex: 1;
  overflow: auto;
  background: white;
}

/* 暗色模式 */
.dark-mode { background: #1a1a1a; }

.dark-mode .keys-sidebar,
.dark-mode .mobile-search-section,
.dark-mode .mobile-detail-header,
.dark-mode .mobile-detail-content,
.dark-mode .content-main {
  background: #2d2d2d !important;
  color: #e5e5e5;
}

.dark-mode .search-container,
.dark-mode .keys-header,
.dark-mode .pagination-container,
.dark-mode .mobile-pagination {
  border-color: #4a4a4a !important;
  background: #262626 !important;
}

.dark-mode .key-item,
.dark-mode .mobile-key-item,
.dark-mode .virtual-key-item {
  border-color: #4a4a4a !important;
  color: #e5e5e5;
  background: #2d2d2d !important;
}

.dark-mode .key-item:hover,
.dark-mode .mobile-key-item:hover,
.dark-mode .virtual-key-item:hover {
  background: #3a3a3a !important;
}

.dark-mode .keys-count,
.dark-mode .mobile-key-stats {
  color: #909399 !important;
}

.dark-mode .welcome-center h2,
.dark-mode .empty-center h3,
.dark-mode .mobile-search-header h3,
.dark-mode .detail-title {
  color: #e5e5e5 !important;
}

.dark-mode .fab-button {
  background: #409eff !important;
  border-color: #409eff !important;
}

.dark-mode .mobile-add-key-btn {
  background: #409eff !important;
  border-color: #409eff !important;
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

/* 响应式优化 */
@media (max-width: 768px) {
  .key-item,
  .mobile-key-item { padding: 10px 12px; }

  .mobile-search-section { padding: 10px 12px; }
  .mobile-detail-header { padding: 10px 12px; }
}

@media (max-width: 480px) {
  .mobile-search-header h3 { font-size: 15px; }
  .mobile-search-section { padding: 8px 10px; }

  .mobile-key-item {
    padding: 8px 10px;
    min-height: 44px;
  }
}

/* Tab右键菜单样式 */
.tab-context-menu {
  position: fixed;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 3000;
  min-width: 120px;
  padding: 4px 0;
  font-size: 14px;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  color: #606266;
}

.context-menu-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.context-menu-item i {
  font-size: 14px;
  width: 16px;
  text-align: center;
}

.tab-label {
  user-select: none;
}

/* 暗色模式下的右键菜单 */
.dark-mode .tab-context-menu {
  background: #2d2d2d !important;
  border-color: #4a4a4a !important;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

.dark-mode .context-menu-item {
  color: #e5e5e5 !important;
}

.dark-mode .context-menu-item:hover {
  background-color: #3a3a3a !important;
  color: #409eff !important;
}

/* 搜索加载动画 */
.el-icon.is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
