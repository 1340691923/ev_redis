<template>
  <div class="key-creator-container" :class="{ 'dark-mode': isDarkMode }">
    <div class="creator-header">
      <h3>创建新Key</h3>
      <div class="creator-actions">
        <el-button @click="cancelCreate">取消</el-button>
        <el-button type="primary" @click="createKey" :loading="creating">创建</el-button>
      </div>
    </div>

    <div class="creator-content">
      <el-form :model="keyForm" label-width="100px" ref="keyFormRef" class="key-form">
        <el-form-item label="Key名称" required>
          <el-input v-model="keyForm.key" placeholder="请输入Key名称" />
        </el-form-item>

        <el-form-item label="数据类型" required>
          <el-select v-model="keyForm.type" placeholder="选择数据类型" style="width: 200px;" @change="onTypeChange">
            <el-option label="String (字符串)" value="string" />
            <el-option label="Hash (哈希)" value="hash" />
            <el-option label="List (列表)" value="list" />
            <el-option label="Set (集合)" value="set" />
            <el-option label="ZSet (有序集合)" value="zset" />
          </el-select>
        </el-form-item>

        <el-form-item label="过期时间">
          <el-input v-model.number="keyForm.ttl" placeholder="-1表示永不过期" type="number" style="width: 200px;" />
        </el-form-item>

        <!-- String 类型 -->
        <el-form-item v-if="keyForm.type === 'string'" label="内容" required>
          <div class="content-editor-container">
            <div class="content-toolbar">
              <el-button-group >
                <el-button @click="formatStringJSON" :disabled="!isStringJSONValid">
                  <i class="el-icon-magic-stick"></i>
                  格式化JSON
                </el-button>
                <el-button @click="compressStringJSON" :disabled="!isStringJSONValid">
                  <i class="el-icon-minus"></i>
                  压缩JSON
                </el-button>
              </el-button-group>
            </div>
          <el-input
            type="textarea"
              v-model="stringValue"
            :rows="10"
              placeholder="请输入字符串内容"
              class="content-editor"
            />
          </div>
        </el-form-item>

        <!-- Hash 类型 -->
        <el-form-item v-if="keyForm.type === 'hash'" label="Hash字段" required>
          <div class="hash-editor-container">
            <div class="content-toolbar">
              <el-button type="primary"  @click="addHashField">
                <i class="el-icon-plus"></i>
                添加字段
              </el-button>
            </div>
            <el-table :data="hashValue" stripe style="width: 100%;" max-height="300">
              <el-table-column type="index" label="#" width="60" />
              <el-table-column label="字段名" width="150">
                <template #default="{ row, $index }">
                  <el-input
                    v-model="row.field"
                    placeholder="字段名"

                  />
                </template>
              </el-table-column>
              <el-table-column label="值">
                <template #default="{ row, $index }">
                  <div class="hash-value-editor">
                    <el-input
                      v-model="row.value"
                      type="textarea"
                      :rows="2"
                      placeholder="值"

                    />
                    <div class="hash-value-actions">
                      <el-button

                        type="text"
                        @click="formatHashValue(row, $index)"
                        :disabled="!isValidJSON(row.value)"
                      >
                        JSON
                      </el-button>
                    </div>
                  </div>
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
        </el-form-item>

        <!-- List 类型 -->
        <el-form-item v-if="keyForm.type === 'list'" label="List元素" required>
          <div class="list-editor-container">
            <div class="content-toolbar">
              <el-button type="primary"  @click="addListItem">
                <i class="el-icon-plus"></i>
                添加元素
              </el-button>
            </div>
            <el-table :data="listValue" stripe style="width: 100%;" max-height="300">
              <el-table-column type="index" label="索引" width="80" />
              <el-table-column label="值">
                <template #default="{ row, $index }">
                  <div class="list-value-editor">
                    <el-input
                      v-model="row.value"
                      type="textarea"
                      :rows="2"
                      placeholder="值"

                    />
                    <div class="list-value-actions">
                      <el-button

                        type="text"
                        @click="formatListValue(row, $index)"
                        :disabled="!isValidJSON(row.value)"
                      >
                        JSON
                      </el-button>
                    </div>
                  </div>
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
        </el-form-item>

        <!-- Set 类型 -->
        <el-form-item v-if="keyForm.type === 'set'" label="Set成员" required>
          <div class="set-editor-container">
            <div class="content-toolbar">
              <el-button type="primary"  @click="addSetMember">
                <i class="el-icon-plus"></i>
                添加成员
              </el-button>
            </div>
            <el-table :data="setValue" stripe style="width: 100%;" max-height="300">
              <el-table-column type="index" label="#" width="60" />
              <el-table-column label="成员">
                <template #default="{ row, $index }">
                  <div class="set-value-editor">
                    <el-input
                      v-model="row.member"
                      type="textarea"
                      :rows="2"
                      placeholder="成员值"

                    />
                    <div class="set-value-actions">
                      <el-button

                        type="text"
                        @click="formatSetValue(row, $index)"
                        :disabled="!isValidJSON(row.member)"
                      >
                        JSON
                      </el-button>
                    </div>
                  </div>
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
        </el-form-item>

        <!-- ZSet 类型 -->
        <el-form-item v-if="keyForm.type === 'zset'" label="ZSet成员" required>
          <div class="zset-editor-container">
            <div class="content-toolbar">
              <el-button type="primary"  @click="addZSetMember">
                <i class="el-icon-plus"></i>
                添加成员
              </el-button>
            </div>
            <el-table :data="zsetValue" stripe style="width: 100%;" max-height="300">
              <el-table-column type="index" label="#" width="60" />
              <el-table-column label="分数" width="120">
                <template #default="{ row, $index }">
                  <el-input-number
                    v-model="row.score"
                    placeholder="分数"

                    :precision="2"
                    style="width: 100%;"
                  />
                </template>
              </el-table-column>
              <el-table-column label="成员">
                <template #default="{ row, $index }">
                  <div class="zset-value-editor">
                    <el-input
                      v-model="row.member"
                      type="textarea"
                      :rows="2"
                      placeholder="成员值"

                    />
                    <div class="zset-value-actions">
                      <el-button

                        type="text"
                        @click="formatZSetValue(row, $index)"
                        :disabled="!isValidJSON(row.member)"
                      >
                        JSON
                      </el-button>
                    </div>
                  </div>
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
        </el-form-item>
      </el-form>
    </div>

    <!-- JSON 格式化对话框 -->
    <el-dialog
      title="JSON 格式化"
      :visible.sync="jsonDialogVisible"
      width="80%"

    >
      <div class="json-dialog-content">
        <div class="json-dialog-toolbar">
          <el-button-group >
            <el-button @click="formatDialogJSON">
              <i class="el-icon-magic-stick"></i>
              格式化
            </el-button>
            <el-button @click="compressDialogJSON">
              <i class="el-icon-minus"></i>
              压缩
            </el-button>
            <el-button @click="copyDialogJSON">
              <i class="el-icon-document-copy"></i>
              复制
            </el-button>
          </el-button-group>
        </div>
        <el-input
          type="textarea"
          v-model="dialogJsonValue"
          :rows="20"
          placeholder="JSON 内容"
          class="json-dialog-editor"
        />
      </div>
      <div slot="footer" class="dialog-footer">
        <el-button @click="jsonDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmJsonEdit">确定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { sdk } from '@elasticview/plugin-sdk'
import { setRedisKey } from "@/api/redis";

const props = defineProps(['database'])
const emit = defineEmits(['save', 'cancel'])

const creating = ref(false)
const keyFormRef = ref()
const stringValue = ref('')
const hashValue = ref([])
const listValue = ref([])
const setValue = ref([])
const zsetValue = ref([])
const jsonDialogVisible = ref(false)
const dialogJsonValue = ref('')
const currentEditContext = ref(null)

// 计算属性
const isDarkMode = computed(() => {
  return sdk.isDarkTheme()
})

const isStringJSONValid = computed(() => {
  return isValidJSON(stringValue.value)
})

const keyForm = ref({
  key: '',
  type: 'string',
  ttl: -1
})

// JSON 处理函数
const isValidJSON = (str) => {
  if (!str || typeof str !== 'string') return false
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}

// String JSON 操作
const formatStringJSON = () => {
  try {
    const parsed = JSON.parse(stringValue.value)
    stringValue.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressStringJSON = () => {
  try {
    const parsed = JSON.parse(stringValue.value)
    stringValue.value = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

// Hash 操作
const addHashField = () => {
  hashValue.value.push({ field: '', value: '' })
}

const removeHashField = (index) => {
  hashValue.value.splice(index, 1)
}

const formatHashValue = (row, index) => {
  currentEditContext.value = { type: 'hash', index, field: 'value' }
  dialogJsonValue.value = row.value
  jsonDialogVisible.value = true
}

// List 操作
const addListItem = () => {
  listValue.value.push({ value: '' })
}

const removeListItem = (index) => {
  listValue.value.splice(index, 1)
}

const formatListValue = (row, index) => {
  currentEditContext.value = { type: 'list', index, field: 'value' }
  dialogJsonValue.value = row.value
  jsonDialogVisible.value = true
}

// Set 操作
const addSetMember = () => {
  setValue.value.push({ member: '' })
}

const removeSetMember = (index) => {
  setValue.value.splice(index, 1)
}

const formatSetValue = (row, index) => {
  currentEditContext.value = { type: 'set', index, field: 'member' }
  dialogJsonValue.value = row.member
  jsonDialogVisible.value = true
}

// ZSet 操作
const addZSetMember = () => {
  zsetValue.value.push({ member: '', score: 0 })
}

const removeZSetMember = (index) => {
  zsetValue.value.splice(index, 1)
}

const formatZSetValue = (row, index) => {
  currentEditContext.value = { type: 'zset', index, field: 'member' }
  dialogJsonValue.value = row.member
  jsonDialogVisible.value = true
}

// JSON 对话框操作
const formatDialogJSON = () => {
  try {
    const parsed = JSON.parse(dialogJsonValue.value)
    dialogJsonValue.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressDialogJSON = () => {
  try {
    const parsed = JSON.parse(dialogJsonValue.value)
    dialogJsonValue.value = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const copyDialogJSON = () => {
  navigator.clipboard.writeText(dialogJsonValue.value).then(() => {
    ElMessage.success('复制成功')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const confirmJsonEdit = () => {
  if (!currentEditContext.value) return

  const { type, index, field } = currentEditContext.value

  switch (type) {
    case 'hash':
      hashValue.value[index][field] = dialogJsonValue.value
      break
    case 'list':
      listValue.value[index][field] = dialogJsonValue.value
      break
    case 'set':
      setValue.value[index][field] = dialogJsonValue.value
      break
    case 'zset':
      zsetValue.value[index][field] = dialogJsonValue.value
      break
  }

  jsonDialogVisible.value = false
  currentEditContext.value = null
}

// 类型改变处理
const onTypeChange = () => {
  // 清空所有数据
  stringValue.value = ''
  hashValue.value = []
  listValue.value = []
  setValue.value = []
  zsetValue.value = []

  // 为新类型添加默认项
  switch (keyForm.value.type) {
    case 'hash':
      addHashField()
      break
    case 'list':
      addListItem()
      break
    case 'set':
      addSetMember()
      break
    case 'zset':
      addZSetMember()
      break
  }
}

// 创建 Key
const createKey = async () => {
  if (!keyForm.value.key) {
    ElMessage.error('请输入Key名称')
    return
  }

  // 验证内容
  let hasContent = false
  let saveValue

  switch (keyForm.value.type) {
    case 'string':
      hasContent = stringValue.value.trim() !== ''
      saveValue = stringValue.value
      break
    case 'hash':
      const validHashFields = hashValue.value.filter(item => item.field && item.field.trim())
      hasContent = validHashFields.length > 0
      saveValue = {}
      validHashFields.forEach(item => {
        saveValue[item.field] = item.value
      })
      break
    case 'list':
      const validListItems = listValue.value.filter(item => item.value && item.value.trim())
      hasContent = validListItems.length > 0
      saveValue = validListItems.map(item => item.value)
      break
    case 'set':
      const validSetMembers = setValue.value.filter(item => item.member && item.member.trim())
      hasContent = validSetMembers.length > 0
      saveValue = validSetMembers.map(item => item.member)
      break
    case 'zset':
      const validZSetMembers = zsetValue.value.filter(item => item.member && item.member.trim())
      hasContent = validZSetMembers.length > 0
      saveValue = validZSetMembers.map(item => ({ member: item.member, score: item.score }))
      break
  }

  if (!hasContent) {
    ElMessage.error('请添加Key内容')
    return
  }

  creating.value = true
  try {
    const connId = sdk.GetSelectEsConnID()
    const res = await setRedisKey({
      es_connect: connId,
      database: props.database,
      key: keyForm.value.key,
      type: keyForm.value.type,
      ttl: keyForm.value.ttl,
      value: saveValue
    })

    if (res.code === 0) {
      ElMessage.success('Key创建成功')
      emit('save', keyForm.value)
    } else {
      ElMessage.error(res.msg || 'Key创建失败')
    }
  } catch (error) {
    ElMessage.error('Key创建失败: ' + error.message)
  } finally {
    creating.value = false
  }
}

const cancelCreate = () => {
  emit('cancel')
}

// 初始化
onMounted(() => {
  // 默认添加一个空项
  addHashField()
  addListItem()
  addSetMember()
  addZSetMember()
})
</script>

<style scoped>
.key-creator-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
  transition: background-color 0.3s;
}

.creator-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background-color 0.3s, border-color 0.3s;
}

.creator-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
  transition: color 0.3s;
}

.creator-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.creator-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.key-form {
  max-width: 800px;
}

.content-editor,
.json-dialog-editor {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.content-editor-container,
.hash-editor-container,
.list-editor-container,
.set-editor-container,
.zset-editor-container {
  width: 100%;
}

.content-toolbar {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

.hash-value-editor,
.list-value-editor,
.set-value-editor,
.zset-value-editor {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.hash-value-actions,
.list-value-actions,
.set-value-actions,
.zset-value-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.json-dialog-content {
  height: 500px;
  display: flex;
  flex-direction: column;
}

.json-dialog-toolbar {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

.json-dialog-editor {
  flex: 1;
}

/* 暗色模式样式 */
.dark-mode {
  background: #1a1a1a !important;
}

.dark-mode .creator-header {
  background: #262626 !important;
  border-bottom-color: #4a4a4a !important;
}

.dark-mode .creator-header h3 {
  color: #e4e7ed !important;
}

.dark-mode .content-toolbar,
.dark-mode .json-dialog-toolbar {
  border-bottom-color: #4a4a4a !important;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .creator-header {
    padding: 12px 16px;
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .creator-actions {
    justify-content: flex-end;
  }

  .creator-content {
    padding: 16px;
  }

  .key-form {
    max-width: none;
  }

  .key-form :deep(.el-form-item__label) {
    width: auto !important;
    margin-bottom: 8px;
  }

  .key-form :deep(.el-form-item__content) {
    margin-left: 0 !important;
  }

  .key-form :deep(.el-select),
  .key-form :deep(.el-input) {
    width: 100% !important;
  }

  .hash-value-editor,
  .list-value-editor,
  .set-value-editor,
  .zset-value-editor {
    flex-direction: column;
  }

  .hash-value-actions,
  .list-value-actions,
  .set-value-actions,
  .zset-value-actions {
    flex-direction: row;
    justify-content: flex-end;
  }
}
</style>
