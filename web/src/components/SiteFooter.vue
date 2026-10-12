<template>
  <footer class="site-footer">
    <div class="sf-inner">
      <div class="sf-brand">
        <div class="sf-logo">
          <span class="logo">☁</span>
          <span class="nm">{{ siteName }}</span>
        </div>
        <p class="desc">轻量、稳定、易用的云主机售卖与管理平台。产品定价、在线下单、自动开通、自助运维，一站式完成。</p>
        <div class="badges">
          <span class="chip"><el-icon><Lightning /></el-icon> 秒级开通</span>
          <span class="chip"><el-icon><Lock /></el-icon> 数据安全</span>
        </div>
      </div>

      <div class="sf-cols">
        <div class="col">
          <h4>产品</h4>
          <a @click="jump('#features')">产品特性</a>
          <a @click="jump('#pricing')">套餐价格</a>
          <a @click="go('/console/products')">云主机商城</a>
        </div>
        <div class="col">
          <h4>用户</h4>
          <a @click="go('/login')">登录</a>
          <a @click="go('/register')">注册</a>
          <a @click="go('/console/wallet')">钱包充值</a>
        </div>
        <div class="col">
          <h4>支持</h4>
          <a @click="jump('#faq')">常见问题</a>
          <a>服务协议</a>
          <a>隐私政策</a>
        </div>
      </div>
    </div>

    <div class="sf-bottom">
      <span>© {{ year }} {{ siteName }}. All rights reserved.</span>
      <span class="made">Go · Vue 3 · Element Plus</span>
    </div>
  </footer>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'

const route = useRoute()
const router = useRouter()
const siteName = ref('OneHost')
const year = new Date().getFullYear()

function jump(sel) {
  if (route.path !== '/') router.push('/').then(() => nextTick(() => {
    const el = document.querySelector(sel); if (el) el.scrollIntoView({ behavior: 'smooth' })
  }))
  else { const el = document.querySelector(sel); if (el) el.scrollIntoView({ behavior: 'smooth' }) }
}
function go(path) { router.push(path) }

onMounted(async () => {
  try { const d = await api.get('/settings/public'); if (d && d.site_name) siteName.value = d.site_name } catch (e) {}
})
</script>

<style scoped>
.site-footer { background: #141826; color: #9aa4bb; margin-top: 0; }
.sf-inner {
  max-width: 1200px; margin: 0 auto; padding: 52px 22px 32px;
  display: grid; grid-template-columns: 1.4fr 1.6fr; gap: 40px;
}
.sf-logo { display: flex; align-items: center; gap: 12px; }
.sf-logo .logo {
  width: 38px; height: 38px; border-radius: 11px; display: grid; place-items: center;
  font-size: 20px; color: #fff; background: var(--brand-grad);
}
.sf-logo .nm { font-size: 19px; font-weight: 800; color: #fff; }
.desc { margin: 16px 0; font-size: 13.5px; line-height: 1.8; max-width: 420px; }
.badges { display: flex; gap: 10px; flex-wrap: wrap; }
.badges .chip { background: rgba(255,255,255,.06); color: #c3cadb; }
.badges .chip .el-icon { color: var(--brand-2); }

.sf-cols { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.col h4 { color: #fff; font-size: 14px; margin: 0 0 14px; font-weight: 700; }
.col a { display: block; color: #9aa4bb; font-size: 13.5px; padding: 6px 0; cursor: pointer; transition: .15s; }
.col a:hover { color: #fff; }

.sf-bottom {
  border-top: 1px solid rgba(255,255,255,.08); max-width: 1200px; margin: 0 auto;
  padding: 18px 22px; display: flex; justify-content: space-between; gap: 10px; flex-wrap: wrap;
  font-size: 12.5px; color: #7c85a0;
}
@media (max-width: 768px) {
  .sf-inner { grid-template-columns: 1fr; gap: 28px; padding: 40px 18px 24px; }
  .sf-cols { grid-template-columns: repeat(2, 1fr); }
}
</style>
