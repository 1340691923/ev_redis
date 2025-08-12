<template>
  <div class="key-detail-container" :class="{ 'dark-mode': isDarkMode }">
    <div class="key-header">
      <div class="key-info">
        <h3>{{ getKeyName() }}</h3>
        <div class="key-meta">
          <el-tag :type="getTypeTagType(getKeyType())" >
            {{ getKeyType() }}
          </el-tag>
          <span class="meta-item">大小: {{ formatSize(keyData?.detail?.sizeBytes || keyData?.sizeBytes || 0) }}</span>
          <span v-if="getTTL() !== -1" class="meta-item">TTL: {{ getTTL() }}s</span>
        </div>
      </div>
      <div class="key-actions">
        <el-button  @click="refreshKey" :loading="refreshing">刷新</el-button>
        <el-button type="primary"  @click="saveKey" :loading="saving">保存</el-button>
        <el-button type="danger"  @click="deleteKey">删除</el-button>
      </div>
    </div>

    <div class="key-content">
      <!-- String 类型 -->
      <div v-if="getKeyType() === 'string'">
        <div class="content-toolbar">
          <el-button-group >
            <el-button @click="formatJSON" :disabled="!isJSONString" :icon="MagicStick">
              格式化JSON
            </el-button>
            <el-button @click="compressJSON" :disabled="!isJSONString" :icon="Minus">
              压缩JSON
            </el-button>
            <el-button @click="copyValue" :icon="DocumentCopy">
              复制
            </el-button>
          </el-button-group>
        </div>
        <el-input
          type="textarea"
          v-model="stringValue"
          :rows="15"
          placeholder="字符串内容"
          class="string-editor"
        />
      </div>

      <!-- Hash 类型 - 改进后的界面 -->
      <div v-else-if="getKeyType() === 'hash'">
        <div class="content-toolbar">
          <el-button-group >

            <el-button type="primary" @click="addHashField" :icon="Plus">
              添加字段
            </el-button>
            <el-button @click="copyHashValue" :icon="DocumentCopy">
              复制全部
            </el-button>
          </el-button-group>
        </div>

        <!-- 只读表格展示 -->
        <el-table
          :data="hashValue"
          stripe
          style="width: 100%;"
          max-height="400"
          @row-click="handleHashRowClick"
          class="hash-readonly-table"
        >
          <el-table-column label="ID" width="60">
            <template #default="{ $index }">
              {{ $index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="field" label="Key" align="center" show-overflow-tooltip />
          <el-table-column prop="value" label="Value" align="center"  show-overflow-tooltip>
            <template #default="{ row }">
              <div class="hash-value-display">
                <span class="value-text" :title="row.value">{{ formatValueForDisplay(row.value) }}</span>
                <el-tag v-if="isValidJSON(row.value)"  type="success">JSON</el-tag>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="220">
            <template #default="{ row, $index }">

                <el-button
                  type="primary"
                  @click.stop="editHashRow(row, $index)"
                  title="编辑"
                  :icon="Edit"
                />
                <el-button
                  type="info"
                  @click.stop="copyHashFieldValue(row.value)"
                  title="复制值"
                  :icon="DocumentCopy"
                />
                <el-button
                  type="danger"
                  @click.stop="removeHashField($index)"
                  title="删除"
                  :icon="Delete"
                />

            </template>
          </el-table-column>
        </el-table>

        <!-- Hash 编辑抽屉 -->
        <el-drawer
          v-model="hashEditDrawerVisible"
          :title="hashEditMode === 'edit' ? '修改行' : '新增行'"
          direction="rtl"
          :size="isMobile?'100%':'80%'"
        >
          <div class="hash-edit-form">
            <el-form :model="currentHashItem" label-width="80px" label-position="top">
              <el-form-item label="Field">
                <el-input
                  v-model="currentHashItem.field"
                  placeholder="请输入字段名"
                  :disabled="hashEditMode === 'edit'"
                />
              </el-form-item>

              <el-form-item label="Value">
                <div class="value-editor-container">
                  <div class="value-toolbar">
                    <el-button-group >
                      <el-button
                        @click="formatCurrentJSON"
                        :disabled="!isValidJSON(currentHashItem.value)"
                        :icon="MagicStick"
                      >
                        格式化JSON
                      </el-button>
                      <el-button
                        @click="compressCurrentJSON"
                        :disabled="!isValidJSON(currentHashItem.value)"
                        :icon="Minus"
                      >
                        压缩JSON
                      </el-button>
                      <el-button
                        @click="copyCurrentValue"
                        :icon="DocumentCopy"
                      >
                        复制
                      </el-button>
                    </el-button-group>
                  </div>
                  <el-input
                    type="textarea"
                    v-model="currentHashItem.value"
                    :rows="12"
                    placeholder="请输入值"

                    class="value-textarea"
                  />
                  <div class="value-info">
                    <el-tag v-if="isValidJSON(currentHashItem.value)"  type="success">
                      <el-icon><Check /></el-icon> 有效的JSON
                    </el-tag>
                    <span class="char-count">{{ currentHashItem.value?.length || 0 }} 字符</span>
                  </div>
                </div>
              </el-form-item>
            </el-form>

            <div class="drawer-footer">
              <el-button @click="hashEditDrawerVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmHashEdit">确定</el-button>
            </div>
          </div>
        </el-drawer>
      </div>

      <!-- List 类型 -->
      <div v-else-if="getKeyType() === 'list'">
        <div class="content-toolbar">
          <el-button-group >

            <el-button type="primary" @click="addListItem" :icon="Plus">
              添加元素
            </el-button>
            <el-button @click="copyListValue" :icon="DocumentCopy">
              复制全部
            </el-button>
          </el-button-group>
        </div>

        <!-- 只读表格展示 -->
        <el-table
          :data="listValue"
          stripe
          style="width: 100%;"
          max-height="400"
          @row-click="handleListRowClick"
          class="list-readonly-table"
        >
          <el-table-column label="索引" width="80">
            <template #default="{ $index }">
              {{ $index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="value" label="值" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="list-value-display">
                <span class="value-text" :title="row.value">{{ formatValueForDisplay(row.value) }}</span>
                <el-tag v-if="isValidJSON(row.value)"  type="success">JSON</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="{ row, $index }">

                <el-button
                  type="primary"
                  @click.stop="editListRow(row, $index)"
                  title="编辑"
                  :icon="Edit"
                />
                <el-button
                  type="info"
                  @click.stop="copyListItemValue(row.value)"
                  title="复制值"
                  :icon="DocumentCopy"
                />
                <el-button
                  type="danger"
                  @click.stop="removeListItem($index)"
                  title="删除"
                  :icon="Delete"
                />

            </template>
          </el-table-column>
        </el-table>

        <!-- List 编辑抽屉 -->
        <el-drawer
          v-model="listEditDrawerVisible"
          :title="listEditMode === 'edit' ? '修改元素' : '新增元素'"
          direction="rtl"
          :size="isMobile?'100%':'80%'"
        >
          <div class="list-edit-form">
            <el-form :model="currentListItem" label-width="80px" label-position="top">
              <el-form-item label="索引" v-if="listEditMode === 'edit'">
                <el-input :value="currentListIndex + 1" disabled />
              </el-form-item>

              <el-form-item label="值">
                <div class="value-editor-container">
                  <div class="value-toolbar">
                    <el-button-group >
                      <el-button
                        @click="formatCurrentListJSON"
                        :disabled="!isValidJSON(currentListItem.value)"
                        :icon="MagicStick"
                      >
                        格式化JSON
                      </el-button>
                      <el-button
                        @click="compressCurrentListJSON"
                        :disabled="!isValidJSON(currentListItem.value)"
                        :icon="Minus"
                      >
                        压缩JSON
                      </el-button>
                      <el-button
                        @click="copyCurrentListValue"
                        :icon="DocumentCopy"
                      >
                        复制
                      </el-button>
                    </el-button-group>
                  </div>
                  <el-input
                    type="textarea"
                    v-model="currentListItem.value"
                    :rows="12"
                    placeholder="请输入值"
                    class="value-textarea"
                  />
                  <div class="value-info">
                    <el-tag v-if="isValidJSON(currentListItem.value)"  type="success">
                      <el-icon><Check /></el-icon> 有效的JSON
                    </el-tag>
                    <span class="char-count">{{ currentListItem.value?.length || 0 }} 字符</span>
                  </div>
                </div>
              </el-form-item>
            </el-form>

            <div class="drawer-footer">
              <el-button @click="listEditDrawerVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmListEdit">确定</el-button>
            </div>
          </div>
        </el-drawer>
      </div>

      <!-- Set 类型 -->
      <div v-else-if="getKeyType() === 'set'">
        <div class="content-toolbar">
          <el-button-group >
            <el-button type="primary" @click="addSetMember" :icon="Plus">
              添加成员
            </el-button>
            <el-button @click="copySetValue" :icon="DocumentCopy">
              复制全部
            </el-button>
          </el-button-group>
        </div>

        <!-- 只读表格展示 -->
        <el-table
          :data="setValue"
          stripe
          style="width: 100%;"
          max-height="400"
          @row-click="handleSetRowClick"
          class="set-readonly-table"
        >
          <el-table-column label="#" width="60">
            <template #default="{ $index }">
              {{ $index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="member" label="成员" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="set-value-display">
                <span class="value-text" :title="row.member">{{ formatValueForDisplay(row.member) }}</span>
                <el-tag v-if="isValidJSON(row.member)"  type="success">JSON</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="{ row, $index }">

                <el-button
                  type="primary"
                  @click.stop="editSetRow(row, $index)"
                  title="编辑"
                  :icon="Edit"
                />
                <el-button
                  type="info"
                  @click.stop="copySetMemberValue(row.member)"
                  title="复制值"
                  :icon="DocumentCopy"
                />
                <el-button
                  type="danger"
                  @click.stop="removeSetMember($index)"
                  title="删除"
                  :icon="Delete"
                />

            </template>
          </el-table-column>
        </el-table>

        <!-- Set 编辑抽屉 -->
        <el-drawer
          v-model="setEditDrawerVisible"
          :title="setEditMode === 'edit' ? '修改成员' : '新增成员'"
          direction="rtl"
          :size="isMobile?'100%':'80%'"
        >
          <div class="set-edit-form">
            <el-form :model="currentSetItem" label-width="80px" label-position="top">
              <el-form-item label="成员值">
                <div class="value-editor-container">
                  <div class="value-toolbar">
                    <el-button-group >
                      <el-button
                        @click="formatCurrentSetJSON"
                        :disabled="!isValidJSON(currentSetItem.member)"
                        :icon="MagicStick"
                      >
                        格式化JSON
                      </el-button>
                      <el-button
                        @click="compressCurrentSetJSON"
                        :disabled="!isValidJSON(currentSetItem.member)"
                        :icon="Minus"
                      >
                        压缩JSON
                      </el-button>
                      <el-button
                        @click="copyCurrentSetValue"
                        :icon="DocumentCopy"
                      >
                        复制
                      </el-button>
                    </el-button-group>
                  </div>
                  <el-input
                    type="textarea"
                    v-model="currentSetItem.member"
                    :rows="12"
                    placeholder="请输入成员值"
                    class="value-textarea"
                  />
                  <div class="value-info">
                    <el-tag v-if="isValidJSON(currentSetItem.member)"  type="success">
                      <el-icon><Check /></el-icon> 有效的JSON
                    </el-tag>
                    <span class="char-count">{{ currentSetItem.member?.length || 0 }} 字符</span>
                  </div>
                </div>
              </el-form-item>
            </el-form>

            <div class="drawer-footer">
              <el-button @click="setEditDrawerVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmSetEdit">确定</el-button>
            </div>
          </div>
        </el-drawer>
      </div>

      <!-- ZSet 类型 -->
      <div v-else-if="getKeyType() === 'zset'">
        <div class="content-toolbar">
          <el-button-group >
            <el-button type="primary" @click="addZSetMember" :icon="Plus">
              添加成员
            </el-button>
            <el-button @click="copyZSetValue" :icon="DocumentCopy">
              复制全部
            </el-button>
          </el-button-group>
        </div>

        <!-- 只读表格展示 -->
        <el-table
          :data="zsetValue"
          stripe
          style="width: 100%;"
          max-height="400"
          @row-click="handleZSetRowClick"
          class="zset-readonly-table"
        >
          <el-table-column label="#" width="60">
            <template #default="{ $index }">
              {{ $index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="score" label="分数" width="120" sortable>
            <template #default="{ row }">
              <el-tag  type="warning">{{ row.score }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="member" label="成员" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="zset-value-display">
                <span class="value-text" :title="row.member">{{ formatValueForDisplay(row.member) }}</span>
                <el-tag v-if="isValidJSON(row.member)"  type="success">JSON</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="{ row, $index }">

                <el-button
                  type="primary"
                  @click.stop="editZSetRow(row, $index)"
                  title="编辑"
                  :icon="Edit"
                />
                <el-button
                  type="info"
                  @click.stop="copyZSetMemberValue(row.member)"
                  title="复制值"
                  :icon="DocumentCopy"
                />
                <el-button
                  type="danger"
                  @click.stop="removeZSetMember($index)"
                  title="删除"
                  :icon="Delete"
                />

            </template>
          </el-table-column>
        </el-table>

        <!-- ZSet 编辑抽屉 -->
        <el-drawer
          v-model="zsetEditDrawerVisible"
          :title="zsetEditMode === 'edit' ? '修改成员' : '新增成员'"
          direction="rtl"
          :size="isMobile?'100%':'80%'"
        >
          <div class="zset-edit-form">
            <el-form :model="currentZSetItem" label-width="80px" label-position="top">
              <el-form-item label="分数">
                <el-input-number
                  v-model="currentZSetItem.score"
                  placeholder="请输入分数"
                  :precision="2"
                  style="width: 100%;"
                />
              </el-form-item>

              <el-form-item label="成员值">
                <div class="value-editor-container">
                  <div class="value-toolbar">
                    <el-button-group >
                      <el-button
                        @click="formatCurrentZSetJSON"
                        :disabled="!isValidJSON(currentZSetItem.member)"
                        :icon="MagicStick"
                      >
                        格式化JSON
                      </el-button>
                      <el-button
                        @click="compressCurrentZSetJSON"
                        :disabled="!isValidJSON(currentZSetItem.member)"
                        :icon="Minus"
                      >
                        压缩JSON
                      </el-button>
                      <el-button
                        @click="copyCurrentZSetValue"
                        :icon="DocumentCopy"
                      >
                        复制
                      </el-button>
                    </el-button-group>
                  </div>
                  <el-input
                    type="textarea"
                    v-model="currentZSetItem.member"
                    :rows="12"
                    placeholder="请输入成员值"
                    class="value-textarea"
                  />
                  <div class="value-info">
                    <el-tag v-if="isValidJSON(currentZSetItem.member)"  type="success">
                      <el-icon><Check /></el-icon> 有效的JSON
                    </el-tag>
                    <span class="char-count">{{ currentZSetItem.member?.length || 0 }} 字符</span>
                  </div>
                </div>
              </el-form-item>
            </el-form>

            <div class="drawer-footer">
              <el-button @click="zsetEditDrawerVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmZSetEdit">确定</el-button>
            </div>
          </div>
        </el-drawer>
      </div>

      <!-- 其他类型 -->
      <div v-else>
        <el-alert
          :title="`${getKeyType()} 类型`"
          description="此数据类型暂不支持编辑，仅支持查看"
          type="info"
          show-icon
        />
        <pre class="json-display">{{ JSON.stringify(keyData?.detail?.value, null, 2) }}</pre>
      </div>
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
            <el-button @click="formatDialogJSON" :icon="MagicStick">
              格式化
            </el-button>
            <el-button @click="compressDialogJSON" :icon="Minus">
              压缩
            </el-button>
            <el-button @click="copyDialogJSON" :icon="DocumentCopy">
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
import { ref, computed, onMounted, watch, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { sdk } from '@elasticview/plugin-sdk'
import { setRedisKey, deleteRedisKey, getRedisKeyDetail } from "@/api/redis";
import { Edit, DocumentCopy, Delete, MagicStick, Minus, Check, Plus } from '@element-plus/icons-vue';

const props = defineProps(['keyData', 'database'])
const emit = defineEmits(['refresh', 'save', 'delete'])

const saving = ref(false)
const refreshing = ref(false)
const forceRenderKey = ref(0)
const stringValue = ref('')
const hashValue = ref([])
const listValue = ref([])
const setValue = ref([])
const zsetValue = ref([])
const jsonDialogVisible = ref(false)
const dialogJsonValue = ref('')
const currentEditContext = ref(null) // 存储当前编辑的上下文

// Hash 编辑抽屉相关
const hashEditDrawerVisible = ref(false)
const hashEditMode = ref('add') // 'add' 或 'edit'
const currentHashItem = ref({ field: '', value: '' })
const currentHashIndex = ref(-1)

// List 编辑抽屉相关
const listEditDrawerVisible = ref(false)
const listEditMode = ref('add') // 'add' 或 'edit'
const currentListItem = ref({ value: '' })
const currentListIndex = ref(-1)

// Set 编辑抽屉相关
const setEditDrawerVisible = ref(false)
const setEditMode = ref('add') // 'add' 或 'edit'
const currentSetItem = ref({ member: '' })
const currentSetIndex = ref(-1)

// ZSet 编辑抽屉相关
const zsetEditDrawerVisible = ref(false)
const zsetEditMode = ref('add') // 'add' 或 'edit'
const currentZSetItem = ref({ member: '', score: 0 })
const currentZSetIndex = ref(-1)

// 计算属性
const isDarkMode = computed(() => {
  return sdk.isDarkTheme()
})

const isJSONString = computed(() => {
  return isValidJSON(stringValue.value)
})

// 安全获取方法
const getKeyName = () => {
  return props.keyData?.keyName || props.keyData?.key || 'Unknown Key'
}

const getKeyType = () => {
  return props.keyData?.keyType || props.keyData?.type || props.keyData?.detail?.type || 'unknown'
}

const getTTL = () => {
  return props.keyData?.detail?.ttl ?? props.keyData?.ttl ?? -1
}

// JSON 处理函数
const isValidJSON = (str) => {
  if (!str || typeof str !== 'string') return false
  try {
    let parsed = JSON.parse(str)
    return typeof parsed === 'object' && parsed !== null;
  } catch (e) {
    return false
  }
}

// 格式化显示值（用于表格展示）
const formatValueForDisplay = (value) => {
  if (!value) return ''
  if (value.length <= 100) return value
  return value.substring(0, 100) + '...'
}

// Hash 相关方法
const openHashEditDrawer = () => {
  hashEditMode.value = 'add'
  currentHashItem.value = { field: '', value: '' }
  currentHashIndex.value = -1
  hashEditDrawerVisible.value = true
}

const handleHashRowClick = (row) => {
  // 从 hashValue 中找到当前行的索引
  const index = hashValue.value.findIndex(item => item === row)
  if (index !== -1) {
    editHashRow(row, index)
  }
}

const editHashRow = (row, index) => {
  hashEditMode.value = 'edit'
  currentHashItem.value = { field: row.field, value: row.value }
  currentHashIndex.value = index
  hashEditDrawerVisible.value = true
}

const confirmHashEdit = () => {
  if (!currentHashItem.value.field.trim()) {
    ElMessage.error('字段名不能为空')
    return
  }

  if (hashEditMode.value === 'add') {
    // 检查字段名是否已存在
    const exists = hashValue.value.some(item => item.field === currentHashItem.value.field)
    if (exists) {
      ElMessage.error('字段名已存在')
      return
    }
    hashValue.value.push({ ...currentHashItem.value })
  } else {
    // 编辑模式
    hashValue.value[currentHashIndex.value] = { ...currentHashItem.value }
  }

  hashEditDrawerVisible.value = false
  ElMessage.success(hashEditMode.value === 'add' ? '添加成功' : '修改成功')
}

const formatCurrentJSON = () => {
  try {
    const parsed = JSON.parse(currentHashItem.value.value)
    currentHashItem.value.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressCurrentJSON = () => {
  try {
    const parsed = JSON.parse(currentHashItem.value.value)
    currentHashItem.value.value = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

// 兼容性更好的复制函数
const copyToClipboard = async (text) => {
  try {
    // 优先使用现代API
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
      return true
    }
    
    // fallback到传统方法（兼容移动端和老版本浏览器）
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'fixed'
    textArea.style.top = '-9999px'
    textArea.style.left = '-9999px'
    textArea.style.opacity = '0'
    document.body.appendChild(textArea)
    
    textArea.focus()
    textArea.select()
    textArea.setSelectionRange(0, text.length)
    
    const success = document.execCommand('copy')
    document.body.removeChild(textArea)
    
    if (!success) {
      throw new Error('execCommand failed')
    }
    
    return true
  } catch (error) {
    console.error('复制失败:', error)
    return false
  }
}

const copyCurrentValue = async () => {
  const success = await copyToClipboard(currentHashItem.value.value)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

const formatJSON = () => {
  try {
    const parsed = JSON.parse(stringValue.value)
    stringValue.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressJSON = () => {
  try {
    const parsed = JSON.parse(stringValue.value)
    stringValue.value = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const copyValue = async () => {
  const success = await copyToClipboard(stringValue.value)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// Hash 操作
const addHashField = () => {
  openHashEditDrawer()
}

const removeHashField = (index) => {
  hashValue.value.splice(index, 1)
}

const formatHashValue = (row, index) => {
  currentEditContext.value = { type: 'hash', index, field: 'value' }
  dialogJsonValue.value = row.value
  jsonDialogVisible.value = true
}

// Hash 整体操作
const canFormatHashJSON = computed(() => {
  return hashValue.value.some(item => isValidJSON(item.value))
})

const formatHashJSON = () => {
  let hasFormatted = false
  hashValue.value.forEach(item => {
    if (isValidJSON(item.value)) {
      try {
        const parsed = JSON.parse(item.value)
        item.value = JSON.stringify(parsed, null, 2)
        hasFormatted = true
      } catch (error) {
        // 跳过格式错误的项
      }
    }
  })
  if (hasFormatted) {
    ElMessage.success('JSON格式化成功')
  } else {
    ElMessage.warning('没有找到可格式化的JSON字段')
  }
}

const compressHashJSON = () => {
  let hasCompressed = false
  hashValue.value.forEach(item => {
    if (isValidJSON(item.value)) {
      try {
        const parsed = JSON.parse(item.value)
        item.value = JSON.stringify(parsed)
        hasCompressed = true
      } catch (error) {
        // 跳过格式错误的项
      }
    }
  })
  if (hasCompressed) {
    ElMessage.success('JSON压缩成功')
  } else {
    ElMessage.warning('没有找到可压缩的JSON字段')
  }
}

const copyHashValue = async () => {
  const hashObj = {}
  hashValue.value.forEach(item => {
    if (item.field && item.field.trim()) {
      hashObj[item.field] = item.value
    }
  })
  const copyText = JSON.stringify(hashObj, null, 2)
  const success = await copyToClipboard(copyText)
  if (success) {
    ElMessage.success('复制Hash全部数据成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

const copyHashFieldValue = async (value) => {
  const success = await copyToClipboard(value)
  if (success) {
    ElMessage.success('复制字段值成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// List 相关方法
const openListEditDrawer = () => {
  listEditMode.value = 'add'
  currentListItem.value = { value: '' }
  currentListIndex.value = -1
  listEditDrawerVisible.value = true
}

const handleListRowClick = (row) => {
  // 从 listValue 中找到当前行的索引
  const index = listValue.value.findIndex(item => item === row)
  if (index !== -1) {
    editListRow(row, index)
  }
}

const editListRow = (row, index) => {
  listEditMode.value = 'edit'
  currentListItem.value = { value: row.value }
  currentListIndex.value = index
  listEditDrawerVisible.value = true
}

const confirmListEdit = () => {
  if (listEditMode.value === 'add') {
    listValue.value.push({ ...currentListItem.value })
  } else {
    // 编辑模式
    listValue.value[currentListIndex.value] = { ...currentListItem.value }
  }

  listEditDrawerVisible.value = false
  ElMessage.success(listEditMode.value === 'add' ? '添加成功' : '修改成功')
}

const formatCurrentListJSON = () => {
  try {
    const parsed = JSON.parse(currentListItem.value.value)
    currentListItem.value.value = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressCurrentListJSON = () => {
  try {
    const parsed = JSON.parse(currentListItem.value.value)
    currentListItem.value.value = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const copyCurrentListValue = async () => {
  const success = await copyToClipboard(currentListItem.value.value)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// List 操作
const addListItem = () => {
  openListEditDrawer()
}

const isMobile = computed(() => {
  return sdk.IsMobile()
})


const removeListItem = (index) => {
  listValue.value.splice(index, 1)
}

const formatListValue = (row, index) => {
  currentEditContext.value = { type: 'list', index, field: 'value' }
  dialogJsonValue.value = row.value
  jsonDialogVisible.value = true
}

const copyListItemValue = async (value) => {
  const success = await copyToClipboard(value)
  if (success) {
    ElMessage.success('复制列表项成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

const copyListValue = async () => {
  const listArray = listValue.value.map(item => item.value)
  const copyText = JSON.stringify(listArray, null, 2)
  const success = await copyToClipboard(copyText)
  if (success) {
    ElMessage.success('复制List全部数据成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// Set 相关方法
const openSetEditDrawer = () => {
  setEditMode.value = 'add'
  currentSetItem.value = { member: '' }
  currentSetIndex.value = -1
  setEditDrawerVisible.value = true
}

const handleSetRowClick = (row) => {
  // 从 setValue 中找到当前行的索引
  const index = setValue.value.findIndex(item => item === row)
  if (index !== -1) {
    editSetRow(row, index)
  }
}

const editSetRow = (row, index) => {
  setEditMode.value = 'edit'
  currentSetItem.value = { member: row.member }
  currentSetIndex.value = index
  setEditDrawerVisible.value = true
}

const confirmSetEdit = () => {
  if (setEditMode.value === 'add') {
    // 检查成员是否已存在
    const exists = setValue.value.some(item => item.member === currentSetItem.value.member)
    if (exists) {
      ElMessage.error('成员已存在')
      return
    }
    setValue.value.push({ ...currentSetItem.value })
  } else {
    // 编辑模式
    setValue.value[currentSetIndex.value] = { ...currentSetItem.value }
  }

  setEditDrawerVisible.value = false
  ElMessage.success(setEditMode.value === 'add' ? '添加成功' : '修改成功')
}

const formatCurrentSetJSON = () => {
  try {
    const parsed = JSON.parse(currentSetItem.value.member)
    currentSetItem.value.member = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressCurrentSetJSON = () => {
  try {
    const parsed = JSON.parse(currentSetItem.value.member)
    currentSetItem.value.member = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const copyCurrentSetValue = async () => {
  const success = await copyToClipboard(currentSetItem.value.member)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// Set 操作
const addSetMember = () => {
  openSetEditDrawer()
}

const removeSetMember = (index) => {
  setValue.value.splice(index, 1)
}

const formatSetValue = (row, index) => {
  currentEditContext.value = { type: 'set', index, field: 'member' }
  dialogJsonValue.value = row.member
  jsonDialogVisible.value = true
}

const copySetMemberValue = async (value) => {
  const success = await copyToClipboard(value)
  if (success) {
    ElMessage.success('复制集合成员成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

const copySetValue = async () => {
  const setArray = setValue.value.map(item => item.member)
  const copyText = JSON.stringify(setArray, null, 2)
  const success = await copyToClipboard(copyText)
  if (success) {
    ElMessage.success('复制Set全部数据成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// ZSet 相关方法
const openZSetEditDrawer = () => {
  zsetEditMode.value = 'add'
  currentZSetItem.value = { member: '', score: 0 }
  currentZSetIndex.value = -1
  zsetEditDrawerVisible.value = true
}

const handleZSetRowClick = (row) => {
  // 从 zsetValue 中找到当前行的索引
  const index = zsetValue.value.findIndex(item => item === row)
  if (index !== -1) {
    editZSetRow(row, index)
  }
}

const editZSetRow = (row, index) => {
  zsetEditMode.value = 'edit'
  currentZSetItem.value = { member: row.member, score: row.score }
  currentZSetIndex.value = index
  zsetEditDrawerVisible.value = true
}

const confirmZSetEdit = () => {
  if (zsetEditMode.value === 'add') {
    // 检查成员是否已存在
    const exists = zsetValue.value.some(item => item.member === currentZSetItem.value.member)
    if (exists) {
      ElMessage.error('成员已存在')
      return
    }
    zsetValue.value.push({ ...currentZSetItem.value })
  } else {
    // 编辑模式
    zsetValue.value[currentZSetIndex.value] = { ...currentZSetItem.value }
  }

  zsetEditDrawerVisible.value = false
  ElMessage.success(zsetEditMode.value === 'add' ? '添加成功' : '修改成功')
}

const formatCurrentZSetJSON = () => {
  try {
    const parsed = JSON.parse(currentZSetItem.value.member)
    currentZSetItem.value.member = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressCurrentZSetJSON = () => {
  try {
    const parsed = JSON.parse(currentZSetItem.value.member)
    currentZSetItem.value.member = JSON.stringify(parsed)
    ElMessage.success('JSON压缩成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const copyCurrentZSetValue = async () => {
  const success = await copyToClipboard(currentZSetItem.value.member)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

// ZSet 操作
const addZSetMember = () => {
  openZSetEditDrawer()
}

const removeZSetMember = (index) => {
  zsetValue.value.splice(index, 1)
}

const formatZSetValue = (row, index) => {
  currentEditContext.value = { type: 'zset', index, field: 'member' }
  dialogJsonValue.value = row.member
  jsonDialogVisible.value = true
}

const copyZSetMemberValue = async (value) => {
  const success = await copyToClipboard(value)
  if (success) {
    ElMessage.success('复制有序集合成员成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

const copyZSetValue = async () => {
  const zsetArray = zsetValue.value.map(item => ({
    member: item.member,
    score: item.score
  }))
  const copyText = JSON.stringify(zsetArray, null, 2)
  const success = await copyToClipboard(copyText)
  if (success) {
    ElMessage.success('复制ZSet全部数据成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
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

const copyDialogJSON = async () => {
  const success = await copyToClipboard(dialogJsonValue.value)
  if (success) {
    ElMessage.success('复制成功')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
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

// 从 detail 数据加载 key 数据（用于刷新）
const loadKeyDataFromDetail = (detail) => {
  if (!detail) return

  const keyType = detail.type || 'unknown'
  const value = detail.value

  switch (keyType) {
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
      } else if (typeof value === 'object' && value !== null) {
        // 如果是对象格式
        Object.entries(value).forEach(([field, val]) => {
          hashValue.value.push({
            field: field,
            value: typeof val === 'string' ? val : JSON.stringify(val)
          })
        })
      }
      break
    case 'list':
      listValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          listValue.value.push({ value: typeof item === 'string' ? item : JSON.stringify(item) })
        })
      }
      break
    case 'set':
      setValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          setValue.value.push({ member: typeof item === 'string' ? item : JSON.stringify(item) })
        })
      }
      break
    case 'zset':
      zsetValue.value = []
      if (Array.isArray(value)) {
        // ZRANGE WITHSCORES返回的是[member1, score1, member2, score2, ...]格式
        for (let i = 0; i < value.length; i += 2) {
          zsetValue.value.push({
            member: typeof value[i] === 'string' ? value[i] : JSON.stringify(value[i]),
            score: parseFloat(value[i + 1]) || 0
          })
        }
      }
      break
  }
}

// 监听 props.keyData 变化
watch(() => props.keyData, (newData) => {
  if (newData?.detail) {
    loadKeyData()
  }
}, { deep: true, immediate: false })

onMounted(() => {
  loadKeyData()
})

const loadKeyData = () => {
  if (!props.keyData?.detail) return

  const keyType = getKeyType()
  const value = props.keyData.detail.value

  switch (keyType) {
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
      } else if (typeof value === 'object' && value !== null) {
        // 如果是对象格式
        Object.entries(value).forEach(([field, val]) => {
          hashValue.value.push({
            field: field,
            value: typeof val === 'string' ? val : JSON.stringify(val)
          })
        })
      }
      break
    case 'list':
      listValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          listValue.value.push({ value: typeof item === 'string' ? item : JSON.stringify(item) })
        })
      }
      break
    case 'set':
      setValue.value = []
      if (Array.isArray(value)) {
        value.forEach(item => {
          setValue.value.push({ member: typeof item === 'string' ? item : JSON.stringify(item) })
        })
      }
      break
    case 'zset':
      zsetValue.value = []
      if (Array.isArray(value)) {
        // ZRANGE WITHSCORES返回的是[member1, score1, member2, score2, ...]格式
        for (let i = 0; i < value.length; i += 2) {
          zsetValue.value.push({
            member: typeof value[i] === 'string' ? value[i] : JSON.stringify(value[i]),
            score: parseFloat(value[i + 1]) || 0
          })
        }
      }
      break
  }
}

const refreshKey = async () => {
  if (refreshing.value) return

  refreshing.value = true
  try {
    ElMessage.info('正在刷新数据...')

    const connId = sdk.GetSelectEsConnID()
    const res = await getRedisKeyDetail({
      es_connect: connId,
      database: props.database,
      key: getKeyName()
    })

    if (res.code === 0 && res.data) {
      // 使用新数据重新加载
      loadKeyDataFromDetail(res.data)

      // 强制重新渲染
      forceRenderKey.value++

      ElMessage.success('数据刷新成功')
      emit('refresh', getKeyName())
    } else {
      ElMessage.error(res.msg || '刷新失败')
    }
  } catch (error) {
    ElMessage.error('刷新失败: ' + error.message)
  } finally {
    refreshing.value = false
  }
}

const saveKey = async () => {
  saving.value = true
  try {
    const connId = sdk.GetSelectEsConnID()
    const keyType = getKeyType()
    let saveValue

    // 根据类型构建保存值
    switch (keyType) {
      case 'string':
        saveValue = stringValue.value
        break
      case 'hash':
        saveValue = {}
        hashValue.value.forEach(item => {
          if (item.field && item.field.trim()) {
            saveValue[item.field] = item.value
          }
        })
        break
      case 'list':
        saveValue = listValue.value.map(item => item.value).filter(v => v !== '')
        break
      case 'set':
        saveValue = setValue.value.map(item => item.member).filter(v => v !== '')
        break
      case 'zset':
        saveValue = zsetValue.value
          .filter(item => item.member && item.member.trim())
          .map(item => ({ member: item.member, score: item.score }))
        break
      default:
        ElMessage.error('不支持的数据类型')
        return
    }

    const res = await setRedisKey({
      es_connect: connId,
      database: props.database,
      key: getKeyName(),
      type: keyType,
      ttl: getTTL(),
      value: saveValue
    })

    if (res.code === 0) {
      ElMessage.success('保存成功')
      emit('save', getKeyName())
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } catch (error) {
    ElMessage.error('保存失败: ' + error.message)
  } finally {
    saving.value = false
  }
}

const deleteKey = async () => {
  try {
    await ElMessageBox.confirm('确定要删除这个Key吗？', '删除确认', { type: 'warning' })

    const connId = sdk.GetSelectEsConnID()
    const res = await deleteRedisKey({
      es_connect: connId,
      database: props.database,
      key: getKeyName()
    })

    if (res.code === 0) {
      ElMessage.success('删除成功')
      emit('delete', getKeyName())
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const getTypeTagType = (type) => {
  const typeMap = {
    string: 'primary',
    hash: 'success',
    list: 'warning',
    set: 'info',
    zset: 'danger'
  }
  return typeMap[type] || 'default'
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

<style scoped>
.key-detail-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
  transition: background-color 0.3s;
}

.key-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  transition: background-color 0.3s, border-color 0.3s;
}

.key-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: #303133;
  word-break: break-all;
  transition: color 0.3s;
}

.key-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #606266;
  flex-wrap: wrap;
}

.meta-item {
  background: #f0f2f5;
  padding: 2px 6px;
  border-radius: 3px;
  transition: background-color 0.3s;
}

.key-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.key-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.content-toolbar {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.string-editor {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

/* Hash 相关样式 */
.hash-readonly-table {
  cursor: pointer;
}

.hash-readonly-table :deep(.el-table__row) {
  cursor: pointer;
}

.hash-readonly-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa !important;
}

.hash-value-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.value-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ttl-text {
  color: #909399;
  font-size: 12px;
}

.hash-edit-form {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
}

.hash-edit-form :deep(.el-form) {
  width: 100%;
}

.hash-edit-form :deep(.el-form-item) {
  width: 100%;
}

.hash-edit-form :deep(.el-form-item__content) {
  width: 100% !important;
}

.value-editor-container {
  position: relative;
  width: 100%;
}

.value-toolbar {
  margin-bottom: 8px;
}

.value-textarea {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  margin-bottom: 8px;
  width: 100% !important;
}

.value-textarea :deep(.el-textarea__inner) {
  width: 100% !important;
  resize: vertical;
}

.value-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #909399;
}

.char-count {
  color: #c0c4cc;
}

.drawer-footer {
  padding: 16px 0;
  text-align: right;
  border-top: 1px solid #e4e7ed;
  margin-top: auto;
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
  min-width: 60px;
}

.hash-value-actions .el-button,
.list-value-actions .el-button,
.set-value-actions .el-button,
.zset-value-actions .el-button {
  padding: 4px 8px;
  margin: 0;
}

.json-display {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  overflow: auto;
  max-height: 500px;
  transition: background-color 0.3s;
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
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

/* 暗色模式样式 */
.dark-mode {
  background: #1a1a1a !important;
}

.dark-mode .key-header {
  background: #262626 !important;
  border-bottom-color: #4a4a4a !important;
}

.dark-mode .key-info h3 {
  color: #e4e7ed !important;
}

.dark-mode .key-meta {
  color: #909399 !important;
}

.dark-mode .meta-item {
  background: #3a3a3a !important;
  color: #e4e7ed !important;
}

.dark-mode .json-display {
  background: #2d2d2d !important;
  color: #e4e7ed !important;
}

.dark-mode .content-toolbar {
  border-bottom-color: #4a4a4a !important;
}

.dark-mode .json-dialog-toolbar {
  border-bottom-color: #4a4a4a !important;
}

.dark-mode .drawer-footer {
  border-top-color: #4a4a4a !important;
}

.dark-mode .hash-readonly-table :deep(.el-table__row:hover) {
  background-color: #3a3a3a !important;
}

/* List 相关样式 */
.list-readonly-table {
  cursor: pointer;
}

.list-readonly-table :deep(.el-table__row) {
  cursor: pointer;
}

.list-readonly-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa !important;
}

.list-value-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.list-edit-form {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
}

.list-edit-form :deep(.el-form) {
  width: 100%;
}

.list-edit-form :deep(.el-form-item) {
  width: 100%;
}

.list-edit-form :deep(.el-form-item__content) {
  width: 100% !important;
}

/* Set 相关样式 */
.set-readonly-table {
  cursor: pointer;
}

.set-readonly-table :deep(.el-table__row) {
  cursor: pointer;
}

.set-readonly-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa !important;
}

.set-value-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.set-edit-form {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
}

.set-edit-form :deep(.el-form) {
  width: 100%;
}

.set-edit-form :deep(.el-form-item) {
  width: 100%;
}

.set-edit-form :deep(.el-form-item__content) {
  width: 100% !important;
}

/* ZSet 相关样式 */
.zset-readonly-table {
  cursor: pointer;
}

.zset-readonly-table :deep(.el-table__row) {
  cursor: pointer;
}

.zset-readonly-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa !important;
}

.zset-value-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.zset-edit-form {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
}

.zset-edit-form :deep(.el-form) {
  width: 100%;
}

.zset-edit-form :deep(.el-form-item) {
  width: 100%;
}

.zset-edit-form :deep(.el-form-item__content) {
  width: 100% !important;
}

.dark-mode .list-readonly-table :deep(.el-table__row:hover),
.dark-mode .set-readonly-table :deep(.el-table__row:hover),
.dark-mode .zset-readonly-table :deep(.el-table__row:hover) {
  background-color: #3a3a3a !important;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .key-header {
    padding: 12px 16px;
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .key-actions {
    justify-content: flex-end;
  }

  .key-content {
    padding: 16px;
  }

  .key-meta {
    gap: 8px;
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

  .hash-edit-form,
  .list-edit-form,
  .set-edit-form,
  .zset-edit-form {
    padding: 16px;
  }

  .el-drawer {
    width: 90% !important;
  }
}
</style>
