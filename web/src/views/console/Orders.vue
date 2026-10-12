<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.orders') }}</h2>
        <div class="sub">查看订单状态、余额支付与续费</div>
      </div>
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="list" stripe>
          <el-table-column prop="order_no" label="订单号" width="200" />
          <el-table-column prop="product_name" label="产品" min-width="140" />
          <el-table-column prop="spec" label="规格" width="200" class="hide-sm" />
          <el-table-column label="金额" width="110">
            <template #default="{ row }"><b>￥{{ yuan(row.amount) }}</b></template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="stType(row.status)" effect="light">{{ stText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="到期时间" width="180">
            <template #default="{ row }">{{ fmt(row.expired_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'unpaid'" size="small" type="primary" @click="pay(row)">余额支付</el-button>
              <el-button v-if="row.status === 'unpaid'" size="small" @click="cancel(row)">取消</el-button>
              <el-button v-if="row.status === 'active'" size="small" type="success" @click="renew(row)">续费</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
import { useUserStore } from '../../store/user'

const { t } = useI18n()
const userStore = useUserStore()
const list = ref([])
const yuan = (c) => ((c || 0) / 100).toFixed(2)
const fmt = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '-')
const stText = (s) =>
  ({ unpaid: '待支付', paid: '已支付', provisioning: '开通中', active: '已开通', failed: '开通失败', expired: '已到期', cancelled: '已取消' }[s] || s)
const stType = (s) =>
  ({ unpaid: 'warning', provisioning: 'info', active: 'success', failed: 'danger', cancelled: 'info' }[s] || '')

async function load() {
  list.value = await api.get('/user/orders')
}
async function pay(row) {
  await ElMessageBox.confirm(`确认使用余额支付 ￥${yuan(row.amount)} ？`, '支付确认', { type: 'warning' })
  await api.post(`/user/orders/${row.id}/pay`)
  ElMessage.success('支付成功，实例正在开通')
  await userStore.fetchProfile()
  load()
}
async function cancel(row) {
  await api.post(`/user/orders/${row.id}/cancel`)
  ElMessage.success('已取消')
  load()
}
async function renew(row) {
  await ElMessageBox.confirm('确认续费一个周期？将从余额扣费', '续费确认', { type: 'warning' })
  await api.post(`/user/orders/${row.id}/renew`)
  ElMessage.success('续费成功')
  await userStore.fetchProfile()
  load()
}
onMounted(load)
</script>
