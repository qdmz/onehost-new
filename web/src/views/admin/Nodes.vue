<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>节点管理</h2>
        <div class="sub">虚拟化后端接入（mock / lxd / incus）</div>
      </div>
      <div class="page-actions">
        <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>接入节点</el-button>
      </div>
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="list" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="type" label="类型" width="90">
            <template #default="{ row }"><span class="chip">{{ row.type }}</span></template>
          </el-table-column>
          <el-table-column prop="api_url" label="API 地址" min-width="200" />
          <el-table-column label="证书" width="100">
            <template #default="{ row }">
              <el-tag :type="row.has_cert ? 'success' : 'info'" effect="light">{{ row.has_cert ? '已配置' : '未配置' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="test(row)">测试</el-button>
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-dialog v-model="dlg" :title="form.id ? '编辑节点' : '接入节点'" width="640px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width:100%">
            <el-option label="mock（模拟，无外部依赖）" value="mock" />
            <el-option label="lxd" value="lxd" />
            <el-option label="incus" value="incus" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input v-model="form.api_url" placeholder="https://10.0.0.10:8443" />
        </el-form-item>
        <el-form-item label="IPv4 池(可选)"><el-input v-model="form.ip_pool" placeholder="10.0.0.100-10.0.0.200" /></el-form-item>
        <el-form-item label="客户端证书">
          <el-input v-model="form.cert" type="textarea" :rows="3" placeholder="-----BEGIN CERTIFICATE-----（PEM，可留空）" />
        </el-form-item>
        <el-form-item label="客户端私钥">
          <el-input v-model="form.key" type="textarea" :rows="3" placeholder="-----BEGIN PRIVATE KEY-----（PEM，可留空）" />
        </el-form-item>
        <el-form-item label="服务端证书">
          <el-input v-model="form.server_cert" type="textarea" :rows="3" placeholder="PEM（填写则启用严格校验，可留空）" />
        </el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const list = ref([])
const dlg = ref(false)
const form = reactive({
  id: 0, name: '', type: 'mock', api_url: '', ip_pool: '',
  cert: '', key: '', server_cert: '', status: 1
})

async function load() {
  list.value = await api.get('/admin/nodes')
}
function openCreate() {
  Object.assign(form, { id: 0, name: '', type: 'mock', api_url: '', ip_pool: '', cert: '', key: '', server_cert: '', status: 1 })
  dlg.value = true
}
function openEdit(row) {
  Object.assign(form, {
    id: row.id, name: row.name, type: row.type, api_url: row.api_url,
    ip_pool: row.ip_pool, cert: '', key: '', server_cert: '', status: row.status
  })
  dlg.value = true
}
async function submit() {
  const payload = {
    name: form.name, type: form.type, api_url: form.api_url, ip_pool: form.ip_pool,
    cert: form.cert, key: form.key, server_cert: form.server_cert, status: form.status
  }
  if (form.id) {
    await api.put(`/admin/nodes/${form.id}`, payload)
  } else {
    await api.post('/admin/nodes', payload)
  }
  ElMessage.success('已保存')
  dlg.value = false
  load()
}
async function test(row) {
  try {
    const d = await api.post(`/admin/nodes/${row.id}/test`)
    ElMessage.success(d.message || '连接正常')
  } catch (e) {}
}
async function remove(row) {
  await ElMessageBox.confirm(`确认删除节点「${row.name}」？`, '确认', { type: 'warning' })
  await api.delete(`/admin/nodes/${row.id}`)
  ElMessage.success('已删除')
  load()
}
onMounted(load)
</script>
