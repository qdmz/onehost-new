import { createI18n } from 'vue-i18n'

const messages = {
  zh: {
    menu: {
      dashboard: '仪表盘',
      products: '云主机商城',
      orders: '我的订单',
      instances: '我的实例',
      wallet: '钱包',
      vouchers: '代金券',
      users: '用户管理',
      nodes: '节点管理',
      settings: '系统设置'
    },
    common: {
      login: '登录',
      register: '注册',
      logout: '退出登录',
      admin: '管理后台',
      console: '用户中心',
      site: 'OneHost 云主机',
      email: '邮箱',
      password: '密码',
      nickname: '昵称',
      save: '保存',
      cancel: '取消',
      confirm: '确定',
      create: '新建',
      edit: '编辑',
      delete: '删除',
      actions: '操作',
      yuan: '元',
      balance: '余额',
      status: '状态',
      submit: '提交'
    }
  },
  en: {
    menu: {
      dashboard: 'Dashboard',
      products: 'Store',
      orders: 'My Orders',
      instances: 'My Instances',
      wallet: 'Wallet',
      vouchers: 'Vouchers',
      users: 'Users',
      nodes: 'Nodes',
      settings: 'Settings'
    },
    common: {
      login: 'Sign in',
      register: 'Sign up',
      logout: 'Sign out',
      admin: 'Admin',
      console: 'Console',
      site: 'OneHost Cloud',
      email: 'Email',
      password: 'Password',
      nickname: 'Nickname',
      save: 'Save',
      cancel: 'Cancel',
      confirm: 'OK',
      create: 'Create',
      edit: 'Edit',
      delete: 'Delete',
      actions: 'Actions',
      yuan: 'CNY',
      balance: 'Balance',
      status: 'Status',
      submit: 'Submit'
    }
  }
}

export default createI18n({
  legacy: false,
  locale: localStorage.getItem('lang') || 'zh',
  fallbackLocale: 'zh',
  messages
})
