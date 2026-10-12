<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.vouchers') }}</h2>
        <div class="sub">输入代金券码，兑换金额将直接充值到账户余额</div>
      </div>
    </div>

    <div class="voucher-wrap">
      <div class="card voucher-card">
        <div class="badge"><el-icon size="30"><Discount /></el-icon></div>
        <h3>兑换代金券</h3>
        <p class="muted">券码通常形如 OHXXXXXXXXXXXXXXXX</p>
        <el-input v-model="code" placeholder="请输入代金券码" size="large" class="mt"
                  @keyup.enter="redeem">
          <template #prefix><el-icon><Ticket /></el-icon></template>
        </el-input>
        <el-button type="primary" size="large" style="width:100%;margin-top:14px"
                   :loading="loading" @click="redeem">立即兑换</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import api from '../../api'
import { useUserStore } from '../../store/user'

const { t } = useI18n()
const userStore = useUserStore()
const code = ref('')
const loading = ref(false)

async function redeem() {
  if (!code.value) {
    ElMessage.warning('请输入券码')
    return
  }
  loading.value = true
  try {
    const d = await api.post('/user/voucher/redeem', { code: code.value })
    ElMessage.success(`兑换成功，+￥${(d.amount / 100).toFixed(2)}`)
    code.value = ''
    await userStore.fetchProfile()
  } catch (e) {
    /* 拦截器已提示 */
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.voucher-wrap { display: flex; justify-content: center; }
.voucher-card { width: 100%; max-width: 460px; text-align: center; padding: 30px 24px; }
.voucher-card h3 { margin: 12px 0 2px; }
.badge {
  width: 64px; height: 64px; margin: 0 auto; border-radius: 18px;
  display: flex; align-items: center; justify-content: center; color: #fff;
  background: var(--brand-grad); box-shadow: 0 12px 26px rgba(99, 102, 241, .35);
}
</style>
