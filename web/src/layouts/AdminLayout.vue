<template>
  <div class="shell" :class="{ collapsed }">
    <!-- 桌面侧栏 -->
    <aside v-if="!isMobile" class="sidebar">
      <div class="brand" @click="$router.push('/admin')">
        <div class="logo">O</div>
        <span v-show="!collapsed" class="brand-text">OneHost<small>管理后台</small></span>
      </div>
      <el-menu :default-active="active" router :collapse="collapsed" :collapse-transition="false" class="side-menu">
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <template #title>{{ m.title }}</template>
        </el-menu-item>
      </el-menu>
      <div class="side-foot">
        <el-button text class="collapse-btn" @click="collapsed = !collapsed">
          <el-icon><component :is="collapsed ? 'Expand' : 'Fold'" /></el-icon>
          <span v-show="!collapsed" style="margin-left:8px">收起菜单</span>
        </el-button>
      </div>
    </aside>

    <!-- 移动端抽屉 -->
    <el-drawer v-model="drawer" direction="ltr" :with-header="false" size="248px" class="side-drawer">
      <div class="drawer-inner">
        <div class="brand" @click="$router.push('/admin'); drawer = false">
          <div class="logo">O</div>
          <span class="brand-text">OneHost<small>管理后台</small></span>
        </div>
        <el-menu :default-active="active" router class="side-menu" @select="drawer = false">
          <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
            <el-icon><component :is="m.icon" /></el-icon>
            <template #title>{{ m.title }}</template>
          </el-menu-item>
        </el-menu>
      </div>
    </el-drawer>

    <!-- 主体 -->
    <div class="main-col">
      <header class="topbar">
        <el-button v-if="isMobile" text class="hamburger" @click="drawer = true">
          <el-icon size="20"><Menu /></el-icon>
        </el-button>
        <div class="titles">
          <h1>{{ currentTitle }}</h1>
          <span class="crumb">管理后台</span>
        </div>
        <div class="spacer"></div>
        <el-button text class="lang" @click="toggleLang">{{ locale === 'zh' ? 'EN' : '中' }}</el-button>
        <el-button text @click="$router.push('/console')">
          <el-icon><Back /></el-icon><span class="hide-sm" style="margin-left:4px">{{ t('common.console') }}</span>
        </el-button>
        <el-dropdown @command="onCommand">
          <span class="user">
            <span class="avatar">A</span>
            <span class="uname hide-sm">管理员</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">
                <el-icon><SwitchButton /></el-icon>{{ t('common.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <main class="content"><router-view /></main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../store/user'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const collapsed = ref(false)
const drawer = ref(false)
const width = ref(window.innerWidth)
const onResize = () => (width.value = window.innerWidth)
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))
const isMobile = computed(() => width.value < 992)

const menus = [
  { path: '/admin/dashboard', icon: 'DataLine', title: '平台总览' },
  { path: '/admin/products', icon: 'Goods', title: '产品管理' },
  { path: '/admin/orders', icon: 'Tickets', title: '订单管理' },
  { path: '/admin/users', icon: 'User', title: '用户管理' },
  { path: '/admin/nodes', icon: 'Cpu', title: '节点管理' },
  { path: '/admin/vouchers', icon: 'Discount', title: '代金券' },
  { path: '/admin/settings', icon: 'Setting', title: '系统设置' }
]

const active = computed(() => route.path)
const currentTitle = computed(() => {
  const hit = menus.find((m) => m.path === route.path)
  return hit ? hit.title : '管理后台'
})

function toggleLang() {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
  localStorage.setItem('lang', locale.value)
}
function onCommand(cmd) {
  if (cmd === 'logout') {
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  }
}
</script>

<style scoped>
.shell { display: flex; min-height: 100vh; background: var(--bg); }
.sidebar {
  width: var(--sidebar-w); flex: none; display: flex; flex-direction: column;
  background: var(--side-bg); color: #fff; position: sticky; top: 0; height: 100vh; transition: width .2s ease;
}
.shell.collapsed .sidebar { width: var(--sidebar-w-collapsed); }
.brand { height: var(--topbar-h); display: flex; align-items: center; gap: 12px; padding: 0 20px; cursor: pointer; flex: none; }
.shell.collapsed .brand { padding: 0; justify-content: center; }
.logo {
  width: 34px; height: 34px; border-radius: 10px; flex: none; display: flex; align-items: center;
  justify-content: center; font-weight: 800; color: #fff; background: var(--brand-grad);
  box-shadow: 0 8px 18px rgba(99, 102, 241, .4);
}
.brand-text { font-size: 17px; font-weight: 700; white-space: nowrap; }
.brand-text small { font-size: 11px; font-weight: 400; opacity: .55; margin-left: 6px; }
.side-menu { flex: 1; border-right: none; background: transparent; padding-top: 6px; overflow-y: auto; }
.side-menu :deep(.el-menu-item) { color: var(--side-text); height: 46px; line-height: 46px; margin: 4px 10px; border-radius: 10px; }
.side-menu :deep(.el-menu-item:hover) { background: rgba(255, 255, 255, .07); color: #fff; }
.side-menu :deep(.el-menu-item.is-active) { background: var(--brand-grad); color: #fff; box-shadow: 0 10px 20px rgba(99, 102, 241, .38); }
.side-foot { padding: 12px; border-top: 1px solid rgba(255, 255, 255, .07); }
.collapse-btn { color: var(--side-text); width: 100%; justify-content: flex-start; }
.collapse-btn:hover { color: #fff; }
.shell.collapsed .collapse-btn { justify-content: center; }
.drawer-inner { height: 100%; background: var(--side-bg); }
.side-drawer :deep(.el-drawer__body) { padding: 0; }
.main-col { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.topbar {
  height: var(--topbar-h); flex: none; display: flex; align-items: center; gap: 10px; padding: 0 20px;
  position: sticky; top: 0; z-index: 20; background: rgba(255, 255, 255, .82);
  backdrop-filter: saturate(180%) blur(12px); border-bottom: 1px solid var(--border);
}
.titles { display: flex; flex-direction: column; line-height: 1.15; }
.titles h1 { font-size: 17px; margin: 0; font-weight: 700; }
.crumb { font-size: 12px; color: var(--muted); }
.hamburger { padding: 6px; }
.lang { font-weight: 700; color: var(--text-2); }
.user { display: inline-flex; align-items: center; gap: 8px; cursor: pointer; color: var(--text); }
.avatar {
  width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  font-weight: 700; color: #fff; background: var(--brand-grad); font-size: 13px;
}
.uname { font-size: 14px; }
.content { flex: 1; min-width: 0; }
</style>
