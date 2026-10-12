<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>订单管理</h2>
        <div class="sub">全平台订单与开通状态</div>
      </div>
      <div class="page-actions">
        <el-select v-model="status" placeholder="全部状态" clearable style="width: 160px" @change="load">
          <el-option label="待支付" value="unpaid" />
          <el-option label="开通中" value="provisioning" />
          <el-option label="已开通" value="active" />
          <el-option label="开通失败" value="failed" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
      </div>
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="rows" stripe>
          <el-table-column label="订单号" width="190">
            <template #default="{ row }">{{ row.order.order_no }}</template>
          </el-table-column>
          <el-table-column label="用户" min-width="180">
            <template #default="{ row }">{{ row.user_email }}</template>
          </el-table-column>
          <el-table-column label="产品" min-width="140">
            <template #default="{ row }">{{ row.order.product_name }}</template>
          </el-table-column>
          <el-table-column label="金额" width="100">
            <template #default="{ row }"><b>￥{{ yuan(row.order.amount) }}</b></template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="stType(row.order.status)" effect="light">{{ stText(row.order.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="备注" min-width="120">
            <template #default="{ row }">{{ row.order.remark || '-' }}</template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">{{ fmt(row.order.created_at) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'

const rows = ref([])
const status = ref('')
const yuan = (c) => (c / 100).toFixed(2)
const fmt = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '-')
const stText = (s) =>
  ({ unpaid: '待支付', paid: '已支付', provisioning: '开通中', active: '已开通', failed: '开通失败', expired: '已到期', cancelled: '已取消' }[s] || s)
const stType = (s) =>
  ({ unpaid: 'warning', provisioning: 'info', active: 'success', failed: 'danger', cancelled: 'info' }[s] || '')

async function load() {
  const q = status.value ? `?status=${status.value}` : ''
  rows.value = await api.get('/admin/orders' + q)
}
onMounted(load)
</script>
