<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>平台总览</h2>
        <div class="sub">运营核心指标一览</div>
      </div>
    </div>

    <div class="stat-grid">
      <div class="stat-card">
        <div class="ico"><el-icon><User /></el-icon></div>
        <div class="meta"><div class="label">用户数</div><div class="value">{{ stats.users }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico cyan"><el-icon><Tickets /></el-icon></div>
        <div class="meta"><div class="label">订单数</div><div class="value">{{ stats.orders }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico green"><el-icon><Monitor /></el-icon></div>
        <div class="meta"><div class="label">实例数</div><div class="value">{{ stats.instances }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico amber"><el-icon><Money /></el-icon></div>
        <div class="meta"><div class="label">累计收入</div><div class="value">￥{{ yuan(stats.revenue) }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico rose"><el-icon><Cpu /></el-icon></div>
        <div class="meta"><div class="label">节点数量</div><div class="value">{{ stats.nodes }}</div></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'

const stats = ref({ users: 0, orders: 0, instances: 0, nodes: 0, revenue: 0 })
const yuan = (c) => ((c || 0) / 100).toFixed(2)

onMounted(async () => {
  stats.value = await api.get('/admin/stats')
})
</script>
