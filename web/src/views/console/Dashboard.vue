<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.dashboard') }}</h2>
        <div class="sub">欢迎回来，{{ nickname }}</div>
      </div>
      <div class="page-actions">
        <el-button type="primary" @click="$router.push('/console/products')">
          <el-icon><Plus /></el-icon>选购云主机
        </el-button>
      </div>
    </div>

    <div class="stat-grid">
      <div class="stat-card">
        <div class="ico"><el-icon><Wallet /></el-icon></div>
        <div class="meta"><div class="label">账户余额</div><div class="value">￥{{ yuan(balance) }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico cyan"><el-icon><Tickets /></el-icon></div>
        <div class="meta"><div class="label">我的订单</div><div class="value">{{ orders.length }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico green"><el-icon><Monitor /></el-icon></div>
        <div class="meta"><div class="label">我的实例</div><div class="value">{{ instances.length }}</div></div>
      </div>
      <div class="stat-card">
        <div class="ico amber"><el-icon><VideoPlay /></el-icon></div>
        <div class="meta"><div class="label">运行中实例</div><div class="value">{{ running }}</div></div>
      </div>
    </div>

    <div class="card mt">
      <h3 class="card-title">快速操作</h3>
      <div class="muted" style="font-size:13px;margin-bottom:12px">常用入口，一步直达</div>
      <el-space wrap :size="12">
        <el-button type="primary" @click="$router.push('/console/products')">
          <el-icon><Goods /></el-icon>选购云主机
        </el-button>
        <el-button @click="$router.push('/console/wallet')">
          <el-icon><Money /></el-icon>账户充值
        </el-button>
        <el-button @click="$router.push('/console/vouchers')">
          <el-icon><Discount /></el-icon>兑换代金券
        </el-button>
        <el-button @click="$router.push('/console/instances')">
          <el-icon><Monitor /></el-icon>管理实例
        </el-button>
      </el-space>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../../api'
import { useUserStore } from '../../store/user'

const { t } = useI18n()
const userStore = useUserStore()
const orders = ref([])
const instances = ref([])
const yuan = (c) => ((c || 0) / 100).toFixed(2)
const balance = computed(() => (userStore.user ? userStore.user.balance : 0))
const nickname = computed(() => (userStore.user ? userStore.user.nickname || userStore.user.email : ''))
const running = computed(() => instances.value.filter((i) => i.status === 'running').length)

onMounted(async () => {
  try {
    await userStore.fetchProfile()
  } catch (e) {}
  orders.value = await api.get('/user/orders')
  instances.value = await api.get('/user/instances')
})
</script>
