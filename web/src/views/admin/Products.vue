<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>产品管理</h2>
        <div class="sub">定价、规格与库存</div>
      </div>
      <div class="page-actions">
        <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>新建产品</el-button>
      </div>
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="list" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column label="价格" width="120">
            <template #default="{ row }"><b>￥{{ yuan(row.price) }}</b>/{{ row.period_days }}天</template>
          </el-table-column>
          <el-table-column label="规格" width="180">
            <template #default="{ row }">{{ row.cpu }}核/{{ row.memory_mb / 1024 }}G/{{ row.disk_gb }}G</template>
          </el-table-column>
          <el-table-column prop="image" label="镜像" width="150" />
          <el-table-column prop="stock" label="库存" width="70" />
          <el-table-column prop="max_per_user" label="限购" width="70" />
          <el-table-column prop="node_id" label="节点" width="70" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-dialog v-model="dlg" :title="form.id ? '编辑产品' : '新建产品'" width="640px" class="resp-dialog">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-row :gutter="10">
          <el-col :xs="24" :sm="12">
            <el-form-item label="价格(元)"><el-input-number v-model="form.priceYuan" :min="0" :step="1" :precision="2" /></el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="周期(天)"><el-input-number v-model="form.period_days" :min="1" /></el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :xs="24" :sm="8"><el-form-item label="CPU核"><el-input-number v-model="form.cpu" :min="1" /></el-form-item></el-col>
          <el-col :xs="24" :sm="8"><el-form-item label="内存MB"><el-input-number v-model="form.memory_mb" :min="128" :step="128" /></el-form-item></el-col>
          <el-col :xs="24" :sm="8"><el-form-item label="硬盘GB"><el-input-number v-model="form.disk_gb" :min="1" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :xs="24" :sm="8"><el-form-item label="流量GB"><el-input-number v-model="form.traffic_gb" :min="0" /></el-form-item></el-col>
          <el-col :xs="24" :sm="8"><el-form-item label="库存"><el-input-number v-model="form.stock" :min="-1" /></el-form-item></el-col>
          <el-col :xs="24" :sm="8"><el-form-item label="限购"><el-input-number v-model="form.max_per_user" :min="0" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="镜像别名"><el-input v-model="form.image" placeholder="images:debian/12" /></el-form-item>
        <el-row :gutter="10">
          <el-col :xs="24" :sm="12">
            <el-form-item label="节点">
              <el-select v-model="form.node_id" placeholder="选择节点" style="width:100%">
                <el-option v-for="n in nodes" :key="n.id" :label="`${n.name} (${n.type})`" :value="n.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="上架"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
          </el-col>
        </el-row>
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
const nodes = ref([])
const dlg = ref(false)
const yuan = (c) => (c / 100).toFixed(2)

const form = reactive({
  id: 0, name: '', description: '', priceYuan: 9.9, period_days: 30,
  cpu: 1, memory_mb: 1024, disk_gb: 10, traffic_gb: 500,
  image: 'images:debian/12', stock: 100, max_per_user: 3, node_id: null, status: 1
})

async function load() {
  list.value = await api.get('/admin/products')
  nodes.value = await api.get('/admin/nodes')
}
function reset() {
  Object.assign(form, {
    id: 0, name: '', description: '', priceYuan: 9.9, period_days: 30,
    cpu: 1, memory_mb: 1024, disk_gb: 10, traffic_gb: 500,
    image: 'images:debian/12', stock: 100, max_per_user: 3,
    node_id: nodes.value[0] ? nodes.value[0].id : null, status: 1
  })
}
function openCreate() {
  reset()
  dlg.value = true
}
function openEdit(row) {
  Object.assign(form, { ...row, priceYuan: row.price / 100 })
  dlg.value = true
}
async function submit() {
  const payload = {
    name: form.name, description: form.description,
    price: Math.round(form.priceYuan * 100), period_days: form.period_days,
    cpu: form.cpu, memory_mb: form.memory_mb, disk_gb: form.disk_gb,
    traffic_gb: form.traffic_gb, image: form.image, stock: form.stock,
    max_per_user: form.max_per_user, node_id: form.node_id, status: form.status
  }
  if (form.id) {
    await api.put(`/admin/products/${form.id}`, payload)
  } else {
    await api.post('/admin/products', payload)
  }
  ElMessage.success('已保存')
  dlg.value = false
  load()
}
async function remove(row) {
  await ElMessageBox.confirm(`确认删除产品「${row.name}」？`, '确认', { type: 'warning' })
  await api.delete(`/admin/products/${row.id}`)
  ElMessage.success('已删除')
  load()
}
onMounted(load)
</script>
