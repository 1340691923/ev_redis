<template>
  <div class="key-editor-page" :class="{ 'mobile-layout': isMobile }">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>
          <i class="el-icon-edit"></i>
          {{ isEdit ? '编辑Key' : '新增Key' }}
        </h2>
        <div class="key-info" v-if="isEdit">
          <span class="info-item">类型: {{ keyType }}</span>
          <span class="info-item">大小: {{ formatSize(keySize) }}</span>
          <span class="info-item">TTL: {{ keyTTL === -1 ? '永不过期' : keyTTL + '秒' }}</span>
        </div>
      </div>

      <div class="header-right">
        <el-button
          type="default"
          :size="isMobile ? 'small' : 'default'"
          @click="goBack"
        >
          返回
        </el-button>
        <el-button
          type="primary"
          :size="isMobile ? 'small' : 'default'"
          :loading="saving"
          @click="saveKey"
        >
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </div>
    </div>

    <!-- Key基本信息 -->
    <el-card class="basic-info-card">
      <div slot="header">
        <span>基本信息</span>
      </div>

      <el-form :model="keyForm" label-width="100px" ref="keyForm" :rules="keyRules">
        <el-row :gutter="20">
          <el-col :span="isMobile ? 24 : 12">
            <el-form-item label="数据库" prop="database">
              <el-select
                v-model="keyForm.database"
                placeholder="选择数据库"
                :disabled="isEdit"
                style="width: 100%;"
              >
                <el-option
                  v-for="db in databases"
                  :key="db.database"
                  :label="`DB${db.database} (${db.keys} keys)`"
                  :value="db.database"
                />
              </el-select>
            </el-form-item>
          </el-col>

          <el-col :span="isMobile ? 24 : 12">
            <el-form-item label="Key名称" prop="key">
              <el-input
                v-model="keyForm.key"
                placeholder="请输入Key名称"
                :disabled="isEdit"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="isMobile ? 24 : 12">
            <el-form-item label="数据类型" prop="type">
              <el-select
                v-model="keyForm.type"
                placeholder="选择数据类型"
                :disabled="false"
                style="width: 100%;"
                @change="onTypeChange"
              >
                <el-option label="String (字符串)" value="string" />
                <el-option label="Hash (哈希)" value="hash" />
                <el-option label="List (列表)" value="list" />
                <el-option label="Set (集合)" value="set" />
                <el-option label="ZSet (有序集合)" value="zset" />
              </el-select>
              <div v-if="isEdit" style="font-size: 12px; color: #909399; margin-top: 4px;">
                编辑模式 - 当前类型: {{ keyForm.type || '未知' }}
              </div>
            </el-form-item>
          </el-col>

          <el-col :span="isMobile ? 24 : 12">
            <el-form-item label="过期时间(秒)">
              <el-input
                v-model.number="keyForm.ttl"
                placeholder="-1表示永不过期"
                type="number"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 数据内容编辑 -->
    <el-card class="content-card">
      <div slot="header">
        <span>数据内容</span>
        <el-button-group style="float: right; margin-top: -5px;">
          <el-tooltip content="格式化JSON" v-if="keyForm.type === 'string'">
            <el-button  @click="formatJSON">
              <i class="el-icon-magic-stick"></i>
            </el-button>
          </el-tooltip>
          <el-tooltip content="清空内容">
            <el-button  @click="clearContent">
              <i class="el-icon-delete"></i>
            </el-button>
          </el-tooltip>
        </el-button-group>
      </div>

      <!-- String类型 -->
      <div v-if="keyForm.type === 'string'" class="string-content">
        <el-input
          type="textarea"
          v-model="stringValue"
          placeholder="请输入字符串内容"
          :rows="isMobile ? 10 : 15"
          resize="vertical"
        />
      </div>

      <!-- Hash类型 -->
      <div v-if="keyForm.type === 'hash'" class="hash-content">
        <div class="operation-bar">
          <el-button type="primary"  @click="addHashField">
            <i class="el-icon-plus"></i>
            添加字段
          </el-button>
        </div>

        <el-table :data="hashValue" stripe style="width: 100%;">
          <el-table-column type="index" label="#" width="60" />
          <el-table-column label="字段名" :width="isMobile ? '120' : '200'">
            <template #default="{ row, $index }">
              <el-input
                v-model="row.field"
                placeholder="字段名"

              />
            </template>
          </el-table-column>
          <el-table-column label="值">
            <template #default="{ row, $index }">
              <el-input
                v-model="row.value"
                placeholder="值"

              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row, $index }">
              <el-button
                type="danger"

                @click="removeHashField($index)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- List类型 -->
      <div v-if="keyForm.type === 'list'" class="list-content">
        <div class="operation-bar">
          <el-button type="primary"  @click="addListItem">
            <i class="el-icon-plus"></i>
            添加元素
          </el-button>
        </div>

        <el-table :data="listValue" stripe style="width: 100%;">
          <el-table-column type="index" label="索引" width="80" />
          <el-table-column label="值">
            <template #default="{ row, $index }">
              <el-input
                v-model="row.value"
                placeholder="列表元素值"

              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row, $index }">
              <el-button
                type="danger"

                @click="removeListItem($index)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Set类型 -->
      <div v-if="keyForm.type === 'set'" class="set-content">
        <div class="operation-bar">
          <el-button type="primary"  @click="addSetMember">
            <i class="el-icon-plus"></i>
            添加成员
          </el-button>
        </div>

        <el-table :data="setValue" stripe style="width: 100%;">
          <el-table-column type="index" label="#" width="60" />
          <el-table-column label="成员值">
            <template #default="{ row, $index }">
              <el-input
                v-model="row.member"
                placeholder="集合成员值"

              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row, $index }">
              <el-button
                type="danger"

                @click="removeSetMember($index)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- ZSet类型 -->
      <div v-if="keyForm.type === 'zset'" class="zset-content">
        <div class="operation-bar">
          <el-button type="primary"  @click="addZSetMember">
            <i class="el-icon-plus"></i>
            添加成员
          </el-button>
        </div>

        <el-table :data="zsetValue" stripe style="width: 100%;">
          <el-table-column type="index" label="#" width="60" />
          <el-table-column label="分值" width="120">
            <template #default="{ row, $index }">
              <el-input
                v-model.number="row.score"
                placeholder="分值"
                type="number"

              />
            </template>
          </el-table-column>
          <el-table-column label="成员值">
            <template #default="{ row, $index }">
              <el-input
                v-model="row.member"
                placeholder="成员值"

              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row, $index }">
              <el-button
                type="danger"

                @click="removeZSetMember($index)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 默认提示 -->
      <div v-if="!keyForm.type"
           style="text-align: center; padding: 40px; color: #909399;">
        <i class="el-icon-info" style="font-size: 48px; margin-bottom: 16px;"></i>
        <p>请选择数据类型</p>
        <p style="font-size: 12px;">当前类型: {{ keyForm.type || '未设置' }}</p>
      </div>

      <!-- 调试信息 -->
      <div v-if="keyForm.type" style="font-size: 12px; color: #909399; text-align: center; margin: 10px;">
        调试: 当前类型 = {{ keyForm.type }}
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { sdk } from '@elasticview/plugin-sdk'
import { getRedisDatabases, getRedisKeyDetail, setRedisKey } from "@/api/redis";

const route = useRoute()
const router = useRouter()

// 页面状态
const isEdit = ref(false)
const saving = ref(false)
const databases = ref([])

// Key信息
const keyType = ref('')
const keySize = ref(0)
const keyTTL = ref(-1)

// 表单数据
const keyForm = ref({
  database: 0,
  key: '',
  type: 'string',
  ttl: -1
})

// 表单验证规则
const keyRules = ref({
  key: [
    { required: true, message: '请输入Key名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择数据类型', trigger: 'change' }
  ]
})

// 不同类型的数据
const stringValue = ref('')
const hashValue = ref([])
const listValue = ref([])
const setValue = ref([])
const zsetValue = ref([])

// 检查是否为移动设备
const isMobile = computed(() => {
  return sdk.IsMobile()
})

// 初始化页面
onMounted(async () => {
  console.log('页面初始化开始')
  console.log('路由参数:', route.query)
  console.log('初始keyForm:', keyForm.value)

  // 获取数据库列表
  await loadDatabases()

  // 检查路由参数确定模式
  const mode = route.query.mode
  const database = route.query.database
  const key = route.query.key

  console.log('解析的参数:', { mode, database, key })

  if (mode === 'edit' && key) {
    console.log('进入编辑模式')
    isEdit.value = true
    keyForm.value.key = key
    keyForm.value.database = database ? parseInt(database) : 0
    console.log('编辑模式keyForm:', keyForm.value)
    await loadKeyDetail()
  } else {
    console.log('进入新增模式')
    isEdit.value = false
    if (database !== undefined) {
      keyForm.value.database = parseInt(database)
    }
    // 确保新增模式有默认类型
    keyForm.value.type = 'string'
    console.log('新增模式keyForm:', keyForm.value)
    // 新增模式，初始化一些默认数据
    initializeDefaultData()
  }

  console.log('页面初始化完成，最终keyForm:', keyForm.value)
})

// 初始化默认数据
const initializeDefaultData = () => {
  // 确保类型已设置
  if (!keyForm.value.type) {
    keyForm.value.type = 'string'
  }

  console.log('初始化默认数据，当前类型:', keyForm.value.type)

  // 为每种类型准备默认数据
  addHashField()
  addListItem()
  addSetMember()
  addZSetMember()

  // 为当前类型设置默认内容
  if (keyForm.value.type === 'string') {
    stringValue.value = ''
  }
}

// 获取数据库列表
const loadDatabases = async () => {
  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      return
    }

    const res = await getRedisDatabases({
      es_connect: connId
    })

    if (res.code === 0) {
      databases.value = res.data.databases || []
      if (databases.value.length > 0 && !isEdit.value) {
        keyForm.value.database = databases.value[0].database
      }
    }
  } catch (error) {
    console.error('获取数据库列表失败:', error)
  }
}

// 加载Key详情
const loadKeyDetail = async () => {
  try {
    const connId = sdk.GetSelectEsConnID()
    console.log('加载Key详情，参数:', {
      es_connect: connId,
      database: keyForm.value.database,
      key: keyForm.value.key
    })

    const res = await getRedisKeyDetail({
      es_connect: connId,
      database: keyForm.value.database,
      key: keyForm.value.key
    })

    console.log('Key详情API响应:', res)

    if (res.code === 0) {
      const detail = res.data
      console.log('Key详情数据:', detail)

      keyType.value = detail.type
      keySize.value = detail.sizeBytes
      keyTTL.value = detail.ttl
      keyForm.value.type = detail.type
      keyForm.value.ttl = detail.ttl

      console.log('设置后的keyForm.type:', keyForm.value.type)

      // 根据类型加载数据
      await loadKeyData(detail.value, detail.type)
    } else {
      console.error('Key详情API返回错误:', res)
      ElMessage.error(res.msg || '获取Key详情失败')
    }
  } catch (error) {
    console.error('获取Key详情失败:', error)
    ElMessage.error('获取Key详情失败: ' + error.message)
  }
}

// 根据类型加载数据
const loadKeyData = async (value, type) => {
  switch (type) {
    case 'string':
      stringValue.value = typeof value === 'string' ? value : JSON.stringify(value)
      break
    case 'hash':
      hashValue.value = []
      if (Array.isArray(value)) {
        // HGETALL返回的是[field1, value1, field2, value2, ...]格式
        for (let i = 0; i < value.length; i += 2) {
          hashValue.value.push({
            field: value[i],
            value: value[i + 1]
          })
        }
      }
      break
    case 'list':
      listValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          listValue.value.push({ value: item })
        })
      }
      break
    case 'set':
      setValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          setValue.value.push({ member: item })
        })
      }
      break
    case 'zset':
      zsetValue.value = []
      if (Array.isArray(value)) {
        // ZRANGE WITHSCORES返回的是[member1, score1, member2, score2, ...]格式
        for (let i = 0; i < value.length; i += 2) {
          zsetValue.value.push({
            member: value[i],
            score: parseFloat(value[i + 1]) || 0
          })
        }
      }
      break
  }
}

// 类型改变处理
const onTypeChange = () => {
  console.log('类型改变为:', keyForm.value.type)
  // 清空之前类型的数据
  clearContent()
}

// 清空内容
const clearContent = () => {
  stringValue.value = ''
  hashValue.value = []
  listValue.value = []
  setValue.value = []
  zsetValue.value = []

  // 为新类型添加默认项
  if (keyForm.value.type === 'hash') {
    addHashField()
  } else if (keyForm.value.type === 'list') {
    addListItem()
  } else if (keyForm.value.type === 'set') {
    addSetMember()
  } else if (keyForm.value.type === 'zset') {
    addZSetMember()
  }
}

// 格式化JSON
const formatJSON = () => {
  try {
    const parsed = JSON.parse(stringValue.value)
    stringValue.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// Hash操作
const addHashField = () => {
  hashValue.value.push({ field: '', value: '' })
}

const removeHashField = (index) => {
  hashValue.value.splice(index, 1)
}

// List操作
const addListItem = () => {
  listValue.value.push({ value: '' })
}

const removeListItem = (index) => {
  listValue.value.splice(index, 1)
}

// Set操作
const addSetMember = () => {
  setValue.value.push({ member: '' })
}

const removeSetMember = (index) => {
  setValue.value.splice(index, 1)
}

// ZSet操作
const addZSetMember = () => {
  zsetValue.value.push({ member: '', score: 0 })
}

const removeZSetMember = (index) => {
  zsetValue.value.splice(index, 1)
}

// 保存Key
const saveKey = async () => {
  // 表单验证
  if (!keyForm.value.key || !keyForm.value.type) {
    ElMessage.error('请填写完整的Key信息')
    return
  }

  saving.value = true

  try {
    const connId = sdk.GetSelectEsConnID()
    if (!connId) {
      ElMessage.error('请先选择Redis连接')
      return
    }

    // 构建保存数据
    const saveData = {
      es_connect: connId,
      database: keyForm.value.database,
      key: keyForm.value.key,
      type: keyForm.value.type,
      ttl: keyForm.value.ttl,
      value: getValueByType()
    }

    console.log('保存数据:', saveData)

    const res = await setRedisKey(saveData)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      goBack()
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败: ' + error.message)
  } finally {
    saving.value = false
  }
}

// 根据类型获取值
const getValueByType = () => {
  switch (keyForm.value.type) {
    case 'string':
      return stringValue.value
    case 'hash':
      const hashData = {}
      hashValue.value.forEach(item => {
        if (item.field && item.value !== undefined) {
          hashData[item.field] = item.value
        }
      })
      return hashData
    case 'list':
      return listValue.value
        .filter(item => item.value !== '')
        .map(item => item.value)
    case 'set':
      return setValue.value
        .filter(item => item.member !== '')
        .map(item => item.member)
    case 'zset':
      const zsetData = []
      zsetValue.value.forEach(item => {
        if (item.member !== '') {
          zsetData.push({
            member: item.member,
            score: item.score || 0
          })
        }
      })
      return zsetData
    default:
      return ''
  }
}

// 格式化文件大小
const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// 返回
const goBack = () => {
  router.back()
}
</script>

<style scoped>
.key-editor-page {
  padding: 20px;
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
}

.header-left h2 i {
  color: #409eff;
}

.key-info {
  display: flex;
  gap: 15px;
  font-size: 14px;
  color: #666;
}

.info-item {
  background: #f5f7fa;
  padding: 4px 8px;
  border-radius: 4px;
}

.basic-info-card {
  margin-bottom: 20px;
}

.content-card {
  margin-bottom: 20px;
}

.operation-bar {
  margin-bottom: 15px;
}

.string-content {
  width: 100%;
}

.string-content .el-textarea {
  font-family: 'Courier New', monospace;
}

.hash-content,
.list-content,
.set-content,
.zset-content {
  max-height: 500px;
  overflow-y: auto;
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

.mobile-layout .key-info {
  justify-content: center;
  flex-wrap: wrap;
}

/* 表格样式优化 */
.el-table {
  border-radius: 6px;
}

.el-table :deep(.el-input) {
  border: none;
  background: transparent;
}

.el-table :deep(.el-input__inner) {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .key-editor-page {
    padding: 10px;
  }

  .header-left h2 {
    font-size: 18px;
  }

  .key-info {
    font-size: 12px;
  }
}
</style>
