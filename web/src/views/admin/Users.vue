<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>用户管理</h2>
        <div class="sub">账号状态与余额调整</div>
      </div>
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="list" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="email" label="邮箱" min-width="180" />
          <el-table-column prop="nickname" label="昵称" width="140" />
          <el-table-column prop="role" label="角色" width="90">
            <template #default="{ row }">
              <el-tag :type="row.role === 'admin' ? 'warning' : 'info'" effect="light">{{ row.role }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="余额" width="110">
            <template #default="{ row }"><b>￥{{ yuan(row.balance) }}</b></template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="light">{{ row.status === 1 ? '正常' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="注册时间" width="180">
            <template #default="{ row }">{{ fmt(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="toggle(row)">{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
              <el-button size="small" type="primary" @click="openBalance(row)">调整余额</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-dialog v-model="dlg" title="调整余额" width="420px">
      <p>用户：<b>{{ cur ? cur.email : '' }}</b></p>
      <p class="muted" style="font-size:13px">变动金额（元，可填负数）：</p>
      <el-input-number v-model="amount" :step="1" :precision="2" style="width:100%" />
      <el-input v-model="remark" placeholder="备注（可选）" style="margin-top: 12px" />
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" @click="doAdjust">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'

const list = ref([])
const dlg = ref(false)
const cur = ref(null)
const amount = ref(0)
const remark = ref('')
const yuan = (c) => ((c || 0) / 100).toFixed(2)
const fmt = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '-')

async function load() {
  list.value = await api.get('/admin/users')
}
async function toggle(row) {
  await api.put(`/admin/users/${row.id}/status`, { status: row.status === 1 ? 0 : 1 })
  ElMessage.success('已更新')
  load()
}
function openBalance(row) {
  cur.value = row
  amount.value = 0
  remark.value = ''
  dlg.value = true
}
async function doAdjust() {
  await api.post(`/admin/users/${cur.value.id}/balance`, {
    change: Math.round(amount.value * 100),
    remark: remark.value
  })
  ElMessage.success('余额已调整')
  dlg.value = false
  load()
}
onMounted(load)
</script>
