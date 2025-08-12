import {request} from '@elasticview/plugin-sdk'

// 获取Redis所有keys
export function getRedisKeys(data: any) {
  return request({
    url: `/api/RedisKeys`,
    method: 'post',
    data
  })
}

// 获取Redis信息总览
export function getRedisInfoOverview(data: any) {
  return request({
    url: '/api/RedisInfoOverview',
    method: 'post',
    data
  })
}

// 获取Redis数据库列表
export function getRedisDatabases(data: any) {
  return request({
    url: '/api/RedisDatabases',
    method: 'post',
    data
  })
}

// 获取Redis内存分析
export function getRedisMemoryAnalysis(data: any) {
  return request({
    url: '/api/RedisMemoryAnalysis',
    method: 'post',
    data
  })
}

// 删除Redis Key
export function deleteRedisKey(data: any) {
  return request({
    url: '/api/RedisDeleteKey',
    method: 'post',
    data
  })
}

// 获取Redis Key详情
export function getRedisKeyDetail(data: any) {
  return request({
    url: '/api/RedisKeyDetail',
    method: 'post',
    data
  })
}

// 保存/更新Redis Key
export function setRedisKey(data: any) {
  return request({
    url: '/api/RedisSetKey',
    method: 'post',
    data
  })
}

// 搜索Redis Keys (后端搜索，使用SCAN + strings.Contains)
export function searchRedisKeys(data: any) {
  return request({
    url: '/api/RedisSearchKeys',
    method: 'post',
    data
  })
}

// 批量添加Redis key
export function batchAddRedisKeys(data: any) {
  return request({
    url: '/api/RedisBatchAddKeys',
    method: 'post',
    data
  })
}

// 批量获取Redis Keys的内存分析
export function getRedisBatchMemoryAnalysis(data: any) {
  return request({
    url: '/api/RedisBatchMemoryAnalysis',
    method: 'post',
    data
  })
}
