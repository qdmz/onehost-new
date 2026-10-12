<template>
  <div class="auth-page">
    <SiteHeader />

    <main class="auth-main">
      <div class="auth-box">
        <div class="auth-side">
          <h2>欢迎回来</h2>
          <p class="side-lead">登录后即可管理云主机、订单、钱包与代金券。</p>
          <ul class="points">
            <li><el-icon><Check /></el-icon> 商品化定价 · 在线下单 · 秒级开通</li>
            <li><el-icon><Check /></el-icon> 余额钱包 · 代金券 · 自动续费</li>
            <li><el-icon><Check /></el-icon> 开机 / 关机 / 重启 自助运维</li>
          </ul>
        </div>

        <div class="auth-card">
          <h3>登录账号</h3>
          <p class="muted">使用邮箱与密码继续</p>
          <el-form :model="form" @submit.prevent="onSubmit" size="large">
            <el-form-item>
              <el-input v-model="form.email" placeholder="邮箱" :prefix-icon="Message" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="form.password" type="password" show-password placeholder="密码"
                        :prefix-icon="Lock" @keyup.enter="onSubmit" />
            </el-form-item>
            <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="onSubmit">
              登 录
            </el-button>
          </el-form>
          <div class="auth-foot">
            还没有账号？<router-link to="/register">立即注册</router-link>
          </div>
        </div>
      </div>
    </main>

    <SiteFooter />
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Message, Lock, Check } from '@element-plus/icons-vue'
import { useUserStore } from '../store/user'
import SiteHeader from '../components/SiteHeader.vue'
import SiteFooter from '../components/SiteFooter.vue'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const form = reactive({ email: '', password: '' })

async function onSubmit() {
  if (!form.email || !form.password) return ElMessage.warning('请输入邮箱和密码')
  loading.value = true
  try {
    const user = await store.login(form.email, form.password)
    ElMessage.success('登录成功')
    router.replace(user.role === 'admin' ? '/admin' : '/console')
  } catch (e) {
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { min-height: 100vh; display: flex; flex-direction: column; background: var(--bg); }
.auth-main { flex: 1; display: flex; align-items: center; justify-content: center; padding: 48px 20px; }
.auth-box {
  width: 100%; max-width: 960px; display: grid; grid-template-columns: 1.1fr 1fr;
  background:
    radial-gradient(700px 400px at 0% 0%, rgba(99,102,241,.14), transparent 60%),
    radial-gradient(700px 500px at 100% 100%, rgba(139,92,246,.14), transparent 60%),
    #fff;
  border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden;
  box-shadow: 0 24px 60px rgba(15, 23, 42, .1);
}
.auth-side { padding: 52px 44px; display: flex; flex-direction: column; justify-content: center; }
.auth-side h2 { margin: 0 0 10px; font-size: 28px; }
.side-lead { color: var(--text-2); font-size: 15px; line-height: 1.8; margin: 0 0 26px; }
.points { list-style: none; padding: 0; margin: 0; display: grid; gap: 14px; }
.points li { display: flex; align-items: center; gap: 10px; color: #475569; font-size: 14.5px; }
.points .el-icon { color: var(--brand-1); }

.auth-card { padding: 52px 44px; background: #fff; border-left: 1px solid var(--border); }
.auth-card h3 { margin: 0 0 4px; font-size: 21px; }
.auth-card .muted { margin: 0 0 22px; font-size: 13px; }
.auth-foot { margin-top: 18px; text-align: center; font-size: 13.5px; color: #64748b; }
.auth-foot a { color: var(--brand-1); text-decoration: none; font-weight: 600; }

@media (max-width: 860px) {
  .auth-box { grid-template-columns: 1fr; max-width: 420px; }
  .auth-side { display: none; }
  .auth-card { padding: 36px 26px; border-left: none; }
}
</style>
