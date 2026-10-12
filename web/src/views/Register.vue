<template>
  <div class="auth-wrap">
    <div class="auth-brand">
      <div class="logo">☁</div>
      <div class="brand-name">{{ siteName }}</div>
      <div class="brand-slogan">注册即享云主机自助购买与运维体验</div>
      <ul class="brand-points">
        <li><el-icon><Check /></el-icon> 30 秒完成注册，立即开通</li>
        <li><el-icon><Check /></el-icon> 钱包 / 代金券 灵活付费</li>
        <li><el-icon><Check /></el-icon> 开机 / 关机 / 重启 / 续费 一站管理</li>
      </ul>
    </div>

    <div class="auth-card">
      <h3>创建账号</h3>
      <p class="muted">只需邮箱与密码即可开始</p>
      <el-form :model="form" @submit.prevent="onSubmit" size="large">
        <el-form-item>
          <el-input v-model="form.email" placeholder="邮箱" :prefix-icon="Message" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.nickname" placeholder="昵称（可选）" :prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" show-password placeholder="密码" :prefix-icon="Lock" />
        </el-form-item>
        <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="onSubmit">注 册</el-button>
      </el-form>
      <div class="auth-foot">
        已有账号？<router-link to="/login">返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Message, Lock, Check, User } from '@element-plus/icons-vue'
import { useUserStore } from '../store/user'
import api from '../api'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const siteName = ref('OneHost')
const form = reactive({ email: '', nickname: '', password: '' })

onMounted(async () => {
  try {
    const d = await api.get('/settings/public')
    if (d && d.site_name) siteName.value = d.site_name
  } catch (e) {}
})

async function onSubmit() {
  if (!form.email || !form.password) return ElMessage.warning('请输入邮箱和密码')
  loading.value = true
  try {
    await store.register({ email: form.email, nickname: form.nickname, password: form.password })
    ElMessage.success('注册成功')
    router.replace('/console')
  } catch (e) {
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  min-height: 100vh; display: grid; grid-template-columns: 1.1fr 1fr;
  background: radial-gradient(1200px 600px at 10% 10%, #eef1ff, #f4f6fb 60%);
}
.auth-brand {
  padding: 8vh 6vw; display: flex; flex-direction: column; justify-content: center;
  color: #1e293b;
}
.auth-brand .logo {
  width: 64px; height: 64px; border-radius: 18px; display: grid; place-items: center;
  font-size: 32px; color: #fff; background: var(--brand-grad);
  box-shadow: 0 12px 30px rgba(99, 102, 241, .4);
}
.brand-name { font-size: 34px; font-weight: 800; margin: 22px 0 8px; letter-spacing: .5px; }
.brand-slogan { color: #64748b; font-size: 15px; }
.brand-points { list-style: none; padding: 0; margin: 32px 0 0; display: grid; gap: 14px; }
.brand-points li { display: flex; align-items: center; gap: 10px; color: #475569; font-size: 14.5px; }
.brand-points .el-icon { color: var(--brand-1); }

.auth-card {
  align-self: center; justify-self: center; width: min(400px, 88vw);
  background: #fff; border-radius: 20px; padding: 36px 32px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, .12);
}
.auth-card h3 { margin: 0 0 4px; font-size: 22px; }
.auth-card .muted { margin: 0 0 22px; font-size: 13px; }
.auth-foot { margin-top: 18px; text-align: center; font-size: 13.5px; color: #64748b; }
.auth-foot a { color: var(--brand-1); text-decoration: none; font-weight: 600; }

@media (max-width: 900px) {
  .auth-wrap { grid-template-columns: 1fr; }
  .auth-brand { display: none; }
}
</style>
