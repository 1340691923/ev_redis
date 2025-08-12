import pluginJson from '../../../plugin.json'

export const routes = [
  {
    path: `/${pluginJson.plugin_alias}`,//根目录必须为插件名
    component: ()=>import("@/layouts/Layout.vue"),
    children: [
      {
        path: 'redis-info',
        component: ()=>import('@/views/redis/index.vue'),
        name: 'redis-info',
        meta: {
          title: 'Redis信息总览',
          icon: 'el-icon-monitor'
        }
      },
      {
        path: 'memory-analysis',
        component: ()=>import('@/views/redis/memory-analysis.vue'),
        name: 'memory-analysis',
        meta: {
          title: '内存分析',
          icon: 'el-icon-pie-chart'
        }
      },

      {
        path: 'redis-manager',
        component: ()=>import('@/views/redis/redis-manager.vue'),
        name: 'redis-manager',
        meta: {
          title: 'Key管理器',
          icon: 'el-icon-s-data'
        }
      },
    ]
  },
]

export default routes
