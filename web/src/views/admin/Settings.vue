<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>系统设置</h2>
        <div class="sub">站点信息与支付网关</div>
      </div>
    </div>

    <div class="card" style="max-width: 680px">
      <el-form :model="form" label-width="140px" label-position="right">
        <el-form-item label="站点名称"><el-input v-model="form.site_name" /></el-form-item>
        <el-form-item label="开放注册">
          <el-switch v-model="form.register_enabled" />
        </el-form-item>
        <el-divider>易支付配置</el-divider>
        <el-form-item label="启用易支付">
          <el-switch v-model="form.yipay_enabled" />
        </el-form-item>
        <el-form-item label="网关地址">
          <el-input v-model="form.yipay_gateway" placeholder="https://pay.example.com" />
        </el-form-item>
        <el-form-item label="商户号 PID"><el-input v-model="form.yipay_pid" /></el-form-item>
        <el-form-item label="商户密钥 KEY"><el-input v-model="form.yipay_key" type="password" show-password /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save"><el-icon><Check /></el-icon>保存设置</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'

const form = reactive({
  site_name: '',
  register_enabled: true,
  yipay_enabled: false,
  yipay_gateway: '',
  yipay_pid: '',
  yipay_key: ''
})

onMounted(async () => {
  const d = await api.get('/admin/settings')
  form.site_name = d.site_name || ''
  form.register_enabled = d.register_enabled !== 'false'
  form.yipay_enabled = d.yipay_enabled === 'true'
  form.yipay_gateway = d.yipay_gateway || ''
  form.yipay_pid = d.yipay_pid || ''
  form.yipay_key = d.yipay_key || ''
})

async function save() {
  await api.put('/admin/settings', {
    site_name: form.site_name,
    register_enabled: form.register_enabled ? 'true' : 'false',
    yipay_enabled: form.yipay_enabled ? 'true' : 'false',
    yipay_gateway: form.yipay_gateway,
    yipay_pid: form.yipay_pid,
    yipay_key: form.yipay_key
  })
  ElMessage.success('设置已保存')
}
</script>
