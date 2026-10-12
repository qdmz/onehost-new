<template>
  <div class="vnc-wrap">
    <div class="vnc-bar" :class="state">{{ statusText }}</div>
    <div class="vnc-stage">
      <canvas ref="cv" class="vnc-canvas"></canvas>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import RFB from '@novnc/novnc/lib/rfb.js'

const props = defineProps({ instanceId: { type: [Number, String], required: true } })
const cv = ref(null)
const state = ref('connecting')
const statusText = ref('正在连接 VNC…')
let rfb = null

function wsURL() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token') || ''
  return `${proto}//${location.host}/api/v1/user/instances/${props.instanceId}/vnc?token=${encodeURIComponent(token)}`
}

onMounted(() => {
  try {
    rfb = new RFB(cv.value, wsURL(), { credentials: {} })
    rfb.scaleViewport = true
    rfb.resizeSession = true
    rfb.addEventListener('connect', () => { state.value = 'ok'; statusText.value = '已连接 VNC' })
    rfb.addEventListener('disconnect', (e) => {
      state.value = 'off'
      statusText.value = 'VNC 连接已断开' + (e.detail && e.detail.clean === false ? '（异常）' : '')
    })
    rfb.addEventListener('securityfailure', () => { state.value = 'err'; statusText.value = 'VNC 安全校验失败' })
  } catch (e) {
    state.value = 'err'
    statusText.value = 'VNC 初始化失败: ' + (e && e.message)
  }
})

onBeforeUnmount(() => {
  if (rfb) { try { rfb.disconnect() } catch (e) { /* ignore */ } }
})
</script>

<style scoped>
.vnc-wrap { height: 100%; display: flex; flex-direction: column; background: #0b1020; border-radius: 12px; overflow: hidden; }
.vnc-bar { padding: 8px 14px; font-size: 13px; color: #cbd5e1; background: #111827; border-bottom: 1px solid #1f2937; }
.vnc-bar.ok { color: var(--success); }
.vnc-bar.err, .vnc-bar.off { color: var(--danger); }
.vnc-stage { flex: 1; min-height: 0; display: flex; align-items: center; justify-content: center; }
.vnc-canvas { width: 100%; height: 100%; display: block; }
</style>
