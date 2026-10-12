<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.instances') }}</h2>
        <div class="sub">开机 / 关机 / 重启 / 销毁，以及 WebSSH / 文件管理 / VNC 控制台</div>
      </div>
      <div class="page-actions">
        <el-button @click="load"><el-icon><Refresh /></el-icon>刷新</el-button>
      </div>
    </div>

    <div class="inst-grid">
      <div v-for="row in list" :key="row.id" class="inst-card">
        <div class="head">
          <div class="row-flex" style="gap:8px">
            <span class="dot" :class="row.status"></span>
            <span class="name">{{ row.name }}</span>
          </div>
          <el-tag :type="stType(row.status)" effect="light">{{ stText(row.status) }}</el-tag>
        </div>
        <div class="chips">
          <span class="chip">{{ row.provider }}</span>
          <span class="chip plain">{{ row.cpu }}核 / {{ row.memory_mb / 1024 }}G / {{ row.disk_gb }}G</span>
        </div>
        <div class="kv"><span class="muted">公网 IP</span><b>{{ row.ip || '—' }}</b></div>
        <div class="kv"><span class="muted">镜像</span><b>{{ row.image }}</b></div>
        <div class="kv"><span class="muted">到期时间</span><b>{{ fmt(row.expired_at) }}</b></div>

        <div class="actions console-actions">
          <el-button size="small" type="primary" plain :disabled="row.status !== 'running'" @click="open(row, 'term')">
            <el-icon><Monitor /></el-icon>控制台
          </el-button>
          <el-button size="small" type="primary" plain :disabled="row.status !== 'running'" @click="open(row, 'files')">
            <el-icon><FolderOpened /></el-icon>文件
          </el-button>
          <el-button size="small" type="primary" plain :disabled="row.status !== 'running' || !row.vnc_port" @click="open(row, 'vnc')">
            <el-icon><View /></el-icon>VNC
          </el-button>
        </div>

        <div class="actions">
          <el-button size="small" :disabled="row.status !== 'stopped'" @click="act(row, 'start')">
            <el-icon><VideoPlay /></el-icon>开机
          </el-button>
          <el-button size="small" :disabled="row.status !== 'running'" @click="act(row, 'stop')">
            <el-icon><VideoPause /></el-icon>关机
          </el-button>
          <el-button size="small" :disabled="row.status !== 'running'" @click="act(row, 'restart')">
            <el-icon><RefreshRight /></el-icon>重启
          </el-button>
          <el-button size="small" type="danger" @click="destroy(row)">
            <el-icon><Delete /></el-icon>销毁
          </el-button>
        </div>
      </div>
      <div v-if="!list.length" class="empty" style="grid-column:1/-1">
        <div class="big">🖥️</div>还没有实例，去商城选购吧
      </div>
    </div>

    <el-drawer v-model="drawer" :size="drawerSize" :title="drawerTitle" :destroy-on-close="true" class="console-drawer">
      <ConsoleTerminal v-if="mode === 'term' && current" :instance-id="current.id" />
      <FileManager v-else-if="mode === 'files' && current" :instance-id="current.id" />
      <VncViewer v-else-if="mode === 'vnc' && current" :instance-id="current.id" />
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
import ConsoleTerminal from '../../components/ConsoleTerminal.vue'
import FileManager from '../../components/FileManager.vue'
import VncViewer from '../../components/VncViewer.vue'

const { t } = useI18n()
const list = ref([])
const fmt = (s) => (s ? new Date(s).toLocaleString('zh-CN') : '-')
const stType = (s) => ({ running: 'success', stopped: 'info', creating: 'warning', error: 'danger' }[s] || '')
const stText = (s) => ({ running: '运行中', stopped: '已关机', creating: '创建中', error: '异常' }[s] || s)

const drawer = ref(false)
const mode = ref('term')
const current = ref(null)
const drawerSize = computed(() => (mode.value === 'term' ? '82%' : mode.value === 'vnc' ? '82%' : '72%'))
const drawerTitle = computed(() => {
  const base = { term: 'WebSSH 控制台', files: 'SFTP 文件管理', vnc: 'VNC 远程桌面' }[mode.value] || '控制台'
  return current.value ? `${base} · ${current.value.name}` : base
})

function open(row, m) {
  current.value = row
  mode.value = m
  drawer.value = true
}

async function load() {
  list.value = await api.get('/user/instances')
}
async function act(row, a) {
  await api.post(`/user/instances/${row.id}/${a}`)
  ElMessage.success('操作成功')
  load()
}
async function destroy(row) {
  await ElMessageBox.confirm('销毁后数据不可恢复，确认销毁？', '危险操作', { type: 'warning' })
  await api.delete(`/user/instances/${row.id}`)
  ElMessage.success('已销毁')
  load()
}
onMounted(load)
</script>

<style scoped>
.inst-grid { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); }
.inst-card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  box-shadow: var(--shadow-1); padding: 18px; transition: transform .18s ease, box-shadow .18s ease;
}
.inst-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-2); }
.inst-card .head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.name { font-weight: 700; font-size: 15px; }
.dot { width: 9px; height: 9px; border-radius: 50%; display: inline-block; background: #cbd5e1; }
.dot.running { background: var(--success); box-shadow: 0 0 0 3px rgba(16, 185, 129, .18); }
.dot.stopped { background: #94a3b8; }
.dot.creating { background: var(--warning); }
.dot.error { background: var(--danger); }
.chips { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 12px; }
.kv { display: flex; align-items: center; justify-content: space-between; font-size: 13px; padding: 5px 0; }
.kv b { font-weight: 600; }
.actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 14px; padding-top: 14px; border-top: 1px dashed var(--border-strong); }
.console-actions { border-top: none; margin-top: 12px; padding-top: 0; }
</style>
