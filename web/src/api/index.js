import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const api = axios.create({ baseURL: '/api/v1', timeout: 60000 })

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = 'Bearer ' + token
  return cfg
})

api.interceptors.response.use(
  (res) => {
    const d = res.data
    if (d && d.code === 0) return d.data
    ElMessage.error((d && d.msg) || '请求失败')
    return Promise.reject(new Error((d && d.msg) || 'error'))
  },
  (err) => {
    const status = err.response && err.response.status
    if (status === 401) {
      localStorage.removeItem('token')
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
    } else {
      const msg = err.response && err.response.data && err.response.data.msg
      ElMessage.error(msg || '网络错误')
    }
    return Promise.reject(err)
  }
)

export default api
