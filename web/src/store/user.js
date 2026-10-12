import { defineStore } from 'pinia'
import api from '../api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null
  }),
  getters: {
    isLogin: (s) => !!s.token,
    isAdmin: (s) => !!s.user && s.user.role === 'admin'
  },
  actions: {
    setToken(t) {
      this.token = t
      localStorage.setItem('token', t)
    },
    async fetchProfile() {
      this.user = await api.get('/user/profile')
      return this.user
    },
    async login(email, password) {
      const d = await api.post('/auth/login', { email, password })
      this.setToken(d.token)
      this.user = d.user
      return d.user
    },
    async register(payload) {
      const d = await api.post('/auth/register', payload)
      this.setToken(d.token)
      this.user = d.user
      return d.user
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    }
  }
})
