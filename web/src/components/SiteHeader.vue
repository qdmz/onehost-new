<template>
  <header class="site-header">
    <div class="sh-inner">
      <a class="brand" @click="goTop">
        <span class="logo">☁</span>
        <span class="nm">{{ siteName }}<small>云主机</small></span>
      </a>

      <nav class="nav hide-sm">
        <a v-for="n in navs" :key="n.sel" @click="jump(n.sel)">{{ n.label }}</a>
      </nav>

      <div class="spacer"></div>

      <div class="actions">
        <template v-if="!isLogin">
          <el-button text @click="$router.push('/login')">登录</el-button>
          <el-button type="primary" round @click="$router.push('/register')">免费注册</el-button>
        </template>
        <template v-else>
          <el-button type="primary" round @click="$router.push(user.isAdmin ? '/admin' : '/console')">
            进入控制台
          </el-button>
        </template>
        <el-button text class="burger hide-md" @click="open = true">
          <el-icon size="20"><Menu /></el-icon>
        </el-button>
      </div>
    </div>

    <el-drawer v-model="open" direction="ttb" size="auto" :with-header="false" class="sh-drawer">
      <div class="drawer-body">
        <a v-for="n in navs" :key="n.sel" class="dl" @click="jump(n.sel); open = false">{{ n.label }}</a>
        <div class="drawer-cta">
          <template v-if="!isLogin">
            <el-button style="width:100%" @click="$router.push('/login'); open = false">登录</el-button>
            <el-button type="primary" style="width:100%; margin:10px 0 0" @click="$router.push('/register'); open = false">免费注册</el-button>
          </template>
          <el-button v-else type="primary" style="width:100%" @click="$router.push(user.isAdmin ? '/admin' : '/console'); open = false">进入控制台</el-button>
        </div>
      </div>
    </el-drawer>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import api from '../api'

const route = useRoute()
const router = useRouter()
const user = useUserStore()
const open = ref(false)
const siteName = ref('OneHost')
const isLogin = computed(() => !!user.token)

const navs = [
  { sel: '#features', label: '产品特性' },
  { sel: '#preview', label: '界面预览' },
  { sel: '#pricing', label: '套餐价格' },
  { sel: '#faq', label: '常见问题' }
]

function goTop() {
  if (route.path !== '/') router.push('/')
  else window.scrollTo({ top: 0, behavior: 'smooth' })
}
function jump(sel) {
  if (route.path !== '/') {
    router.push('/').then(() => nextTick(scrollTo(sel)))
  } else {
    scrollTo(sel)
  }
}
function scrollTo(sel) {
  const el = document.querySelector(sel)
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(async () => {
  try {
    const d = await api.get('/settings/public')
    if (d && d.site_name) siteName.value = d.site_name
  } catch (e) {}
  if (user.token && !user.user) { try { await user.fetchProfile() } catch (e) {} }
})
</script>

<style scoped>
.site-header {
  position: sticky; top: 0; z-index: 50;
  background: rgba(255, 255, 255, .82);
  backdrop-filter: saturate(180%) blur(14px);
  border-bottom: 1px solid var(--border);
}
.sh-inner {
  max-width: 1200px; margin: 0 auto; height: 68px;
  display: flex; align-items: center; gap: 22px; padding: 0 22px;
}
.brand { display: flex; align-items: center; gap: 12px; cursor: pointer; flex: none; }
.logo {
  width: 38px; height: 38px; border-radius: 11px; display: grid; place-items: center;
  font-size: 20px; color: #fff; background: var(--brand-grad);
  box-shadow: 0 8px 18px rgba(99, 102, 241, .38);
}
.nm { font-size: 18px; font-weight: 800; color: var(--text); letter-spacing: .3px; }
.nm small { font-size: 11px; font-weight: 400; color: var(--muted); margin-left: 6px; }

.nav { display: flex; align-items: center; gap: 6px; }
.nav a {
  color: var(--text-2); font-size: 14.5px; font-weight: 500; padding: 8px 12px;
  border-radius: 9px; cursor: pointer; transition: .15s;
}
.nav a:hover { color: var(--brand-1); background: var(--brand-grad-soft); }

.actions { display: flex; align-items: center; gap: 6px; }

.drawer-body { padding: 12px 6px; }
.drawer-body .dl {
  display: block; padding: 14px 12px; font-size: 16px; color: var(--text);
  border-bottom: 1px solid var(--border); cursor: pointer;
}
.drawer-cta { padding: 18px 12px 6px; }
.hide-md { display: none; }
@media (max-width: 860px) {
  .hide-md { display: inline-flex; }
  .sh-inner { gap: 12px; padding: 0 16px; }
}
</style>
