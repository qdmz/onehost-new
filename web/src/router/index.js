import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/console' },
  { path: '/login', component: () => import('../views/Login.vue') },
  { path: '/register', component: () => import('../views/Register.vue') },
  {
    path: '/console',
    component: () => import('../layouts/ConsoleLayout.vue'),
    meta: { auth: true },
    children: [
      { path: '', redirect: '/console/dashboard' },
      { path: 'dashboard', component: () => import('../views/console/Dashboard.vue') },
      { path: 'products', component: () => import('../views/console/Products.vue') },
      { path: 'orders', component: () => import('../views/console/Orders.vue') },
      { path: 'instances', component: () => import('../views/console/Instances.vue') },
      { path: 'wallet', component: () => import('../views/console/Wallet.vue') },
      { path: 'vouchers', component: () => import('../views/console/Vouchers.vue') }
    ]
  },
  {
    path: '/admin',
    component: () => import('../layouts/AdminLayout.vue'),
    meta: { auth: true, admin: true },
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', component: () => import('../views/admin/Dashboard.vue') },
      { path: 'products', component: () => import('../views/admin/Products.vue') },
      { path: 'orders', component: () => import('../views/admin/Orders.vue') },
      { path: 'users', component: () => import('../views/admin/Users.vue') },
      { path: 'nodes', component: () => import('../views/admin/Nodes.vue') },
      { path: 'vouchers', component: () => import('../views/admin/Vouchers.vue') },
      { path: 'settings', component: () => import('../views/admin/Settings.vue') }
    ]
  }
]

const router = createRouter({ history: createWebHashHistory(), routes })

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.meta.auth && !token) return '/login'
  return true
})

export default router
