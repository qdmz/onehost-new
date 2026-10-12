<template>
  <div class="fm">
    <div class="fm-bar">
      <el-button size="small" :disabled="path === '.'" @click="up"><el-icon><Top /></el-icon>上级</el-button>
      <el-button size="small" @click="load"><el-icon><Refresh /></el-icon>刷新</el-button>
      <el-button size="small" @click="mkdir"><el-icon><FolderAdd /></el-icon>新建目录</el-button>
      <el-upload :show-file-list="false" :http-request="upload" class="up">
        <el-button size="small" type="primary"><el-icon><Upload /></el-icon>上传文件</el-button>
      </el-upload>
      <span class="crumb"><el-icon><FolderOpened /></el-icon>{{ path }}</span>
    </div>
    <el-table :data="entries" v-loading="loading" size="small" height="100%" @row-dblclick="enter">
      <el-table-column label="名称" min-width="260">
        <template #default="{ row }">
          <span class="fname" :class="{ dir: row.is_dir }" @click="row.is_dir ? enter(row) : null">
            <el-icon><component :is="row.is_dir ? 'Folder' : 'Document'" /></el-icon>
            {{ row.name }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="110">
        <template #default="{ row }">{{ row.is_dir ? '—' : fmtSize(row.size) }}</template>
      </el-table-column>
      <el-table-column label="权限" width="130" prop="mode" />
      <el-table-column label="修改时间" width="180">
        <template #default="{ row }">{{ fmtTime(row.mod_time) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" align="right">
        <template #default="{ row }">
          <el-button v-if="!row.is_dir" link type="primary" @click="download(row)">下载</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const props = defineProps({ instanceId: { type: [Number, String], required: true } })
const path = ref('.')
const entries = ref([])
const loading = ref(false)

const fmtSize = (n) => {
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB'
  return (n / 1024 / 1024).toFixed(2) + ' MB'
}
const fmtTime = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '')

function join(base, name) {
  return base === '.' ? './' + name : base.replace(/\/+$/, '') + '/' + name
}

async function load() {
  loading.value = true
  try {
    const d = await api.get(`/user/instances/${props.instanceId}/files`, { params: { path: path.value } })
    entries.value = d.entries || []
  } finally {
    loading.value = false
  }
}

function enter(row) {
  if (!row.is_dir) return
  path.value = join(path.value, row.name)
  load()
}

function up() {
  if (path.value === '.') return
  const parts = path.value.replace(/^\.\//, '').split('/').filter(Boolean)
  parts.pop()
  path.value = parts.length ? './' + parts.join('/') : '.'
  load()
}

async function mkdir() {
  const { value } = await ElMessageBox.prompt('新目录名称', '新建目录', { inputPattern: /\S+/, inputErrorMessage: '名称不能为空' })
  await api.post(`/user/instances/${props.instanceId}/files/mkdir`, null, { params: { path: join(path.value, value) } })
  ElMessage.success('已创建')
  load()
}

async function upload(opt) {
  const fd = new FormData()
  fd.append('file', opt.file)
  await api.post(`/user/instances/${props.instanceId}/files/upload`, fd, {
    params: { dir: path.value },
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  ElMessage.success('上传成功')
  load()
}

function download(row) {
  const token = localStorage.getItem('token') || ''
  const p = join(path.value, row.name)
  const url = `/api/v1/user/instances/${props.instanceId}/files/download?path=${encodeURIComponent(p)}&token=${encodeURIComponent(token)}`
  window.open(url, '_blank')
}

async function remove(row) {
  await ElMessageBox.confirm(`确认删除「${row.name}」？`, '危险操作', { type: 'warning' })
  await api.delete(`/user/instances/${props.instanceId}/files`, { params: { path: join(path.value, row.name) } })
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<style scoped>
.fm { height: 100%; display: flex; flex-direction: column; gap: 12px; }
.fm-bar { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.fm-bar .up { display: inline-block; }
.crumb { margin-left: auto; font-size: 13px; color: var(--text-2); display: inline-flex; align-items: center; gap: 6px; background: var(--surface); border: 1px solid var(--border); border-radius: 999px; padding: 4px 12px; }
.fname { display: inline-flex; align-items: center; gap: 6px; cursor: default; }
.fname.dir { cursor: pointer; color: var(--brand-1); font-weight: 600; }
</style>
