<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.wallet') }}</h2>
        <div class="sub">充值、查看余额流水</div>
      </div>
    </div>

    <div class="grid-2">
      <div class="card">
        <div class="balance-hero">
          <div class="muted" style="font-size:13px">账户余额</div>
          <div class="big-amount">￥{{ yuan(balance) }}</div>
        </div>
        <el-divider />
        <div class="form-row">
          <span class="label">充值金额</span>
          <el-input-number v-model="amount" :min="1" :max="100000" />
          <span class="muted">元</span>
        </div>
        <div class="form-row">
          <span class="label">支付方式</span>
          <el-radio-group v-model="method">
            <el-radio-button label="alipay">支付宝</el-radio-button>
            <el-radio-button label="wxpay">微信支付</el-radio-button>
          </el-radio-group>
        </div>
        <el-button type="primary" size="large" style="width:100%" @click="recharge">
          <el-icon><Money /></el-icon>生成充值订单
        </el-button>
        <el-alert v-if="mockMode" class="mt" type="warning" :closable="false"
                  title="未接入真实支付网关，可在右侧充值订单点「模拟支付」完成入账" />
      </div>

      <div>
        <div class="card">
          <h3 class="card-title">充值订单</h3>
          <div class="table-wrap">
            <el-table :data="recharges" size="small" max-height="280">
              <el-table-column prop="order_no" label="订单号" width="180" />
              <el-table-column label="金额" width="90">
                <template #default="{ row }">￥{{ yuan(row.amount) }}</template>
              </el-table-column>
              <el-table-column prop="method" label="方式" width="80" />
              <el-table-column label="状态" width="90">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'paid' ? 'success' : 'warning'" effect="light">
                    {{ row.status === 'paid' ? '已支付' : '待支付' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="110">
                <template #default="{ row }">
                  <el-button v-if="row.status !== 'paid'" size="small" type="primary" @click="payOrder(row)">模拟支付</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <div class="card">
          <h3 class="card-title">余额流水</h3>
          <div class="table-wrap">
            <el-table :data="logs" size="small" max-height="320">
              <el-table-column label="时间" width="170">
                <template #default="{ row }">{{ fmt(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="变动" width="110">
                <template #default="{ row }">
                  <span :style="{ color: row.change >= 0 ? 'var(--success)' : 'var(--danger)', fontWeight: 600 }">
                    {{ row.change >= 0 ? '+' : '' }}{{ yuan(row.change) }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="余额" width="100">
                <template #default="{ row }">￥{{ yuan(row.balance_after) }}</template>
              </el-table-column>
              <el-table-column prop="type" label="类型" width="90" />
              <el-table-column prop="remark" label="备注" min-width="140" />
            </el-table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import api from '../../api'
import { useUserStore } from '../../store/user'

const { t } = useI18n()
const userStore = useUserStore()
const amount = ref(10)
const method = ref('alipay')
const mockMode = ref(false)
const recharges = ref([])
const logs = ref([])

const balance = computed(() => (userStore.user ? userStore.user.balance : 0))
const yuan = (c) => ((c || 0) / 100).toFixed(2)
const fmt = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '-')

async function load() {
  recharges.value = await api.get('/user/recharge/orders')
  logs.value = await api.get('/user/balance/logs')
}
async function recharge() {
  const d = await api.post('/user/recharge', { amount: amount.value * 100, method: method.value })
  mockMode.value = d.mock
  if (d.pay_url) {
    window.open(d.pay_url, '_blank')
    ElMessage.success('已跳转到支付页面')
  } else {
    ElMessage.success('已生成充值订单，请完成支付')
  }
  load()
}
async function payOrder(row) {
  await api.post(`/user/recharge/${row.id}/mockpay`)
  ElMessage.success('模拟支付成功，余额已到账')
  await userStore.fetchProfile()
  load()
}
onMounted(async () => {
  try {
    await userStore.fetchProfile()
  } catch (e) {}
  load()
})
</script>

<style scoped>
.balance-hero {
  border-radius: var(--radius-sm); padding: 20px; text-align: center;
  background: var(--brand-grad-soft);
}
.big-amount { font-size: 38px; font-weight: 800; color: var(--brand-1); margin-top: 6px; }
.form-row { display: flex; align-items: center; gap: 12px; margin-bottom: 14px; }
.label { width: 72px; color: var(--text-2); font-size: 13px; }
</style>
