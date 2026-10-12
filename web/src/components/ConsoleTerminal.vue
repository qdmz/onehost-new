<template>
  <div class="term-wrap">
    <div ref="el" class="term"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({ instanceId: { type: [Number, String], required: true } })
const el = ref(null)

let term = null
let fit = null
let ws = null
let ro = null
let timer = null

function wsURL() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token') || ''
  return `${proto}//${location.host}/api/v1/user/instances/${props.instanceId}/ssh?token=${encodeURIComponent(token)}&cols=120&rows=30`
}

function onResize() {
  if (ws && ws.readyState === WebSocket.OPEN && term) {
    ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
  }
}

onMounted(() => {
  term = new Terminal({
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
    fontSize: 13,
    cursorBlink: true,
    theme: { background: '#0b1020', foreground: '#e5e7eb', cursor: '#8b5cf6' }
  })
  fit = new FitAddon()
  term.loadAddon(fit)
  term.open(el.value)
  setTimeout(() => { try { fit.fit() } catch (e) { /* ignore */ } }, 40)

  term.write('\x1b[38;5;141m正在连接实例控制台…\x1b[0m\r\n')

  ws = new WebSocket(wsURL())
  ws.binaryType = 'arraybuffer'
  ws.onmessage = (ev) => {
    if (typeof ev.data === 'string') {
      try {
        const m = JSON.parse(ev.data)
        if (m.type === 'ready') {
          term.clear()
          onResize()
        } else if (m.type === 'error') {
          term.write('\r\n\x1b[31m' + m.msg + '\x1b[0m\r\n')
        } else if (m.type === 'exit') {
          term.write('\r\n\x1b[33m会话已结束 (exit ' + m.code + ')\x1b[0m\r\n')
        }
      } catch (e) { /* ignore */ }
    } else {
      term.write(new Uint8Array(ev.data))
    }
  }
  ws.onclose = () => term.write('\r\n\x1b[31m连接已断开\x1b[0m\r\n')
  ws.onerror = () => term.write('\r\n\x1b[31m连接异常\x1b[0m\r\n')

  term.onData((d) => {
    if (ws && ws.readyState === WebSocket.OPEN) ws.send(new TextEncoder().encode(d))
  })

  ro = new ResizeObserver(() => { try { fit.fit() } catch (e) { /* ignore */ } })
  ro.observe(el.value)
  timer = setInterval(onResize, 1500)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  if (ro) ro.disconnect()
  if (ws) { try { ws.close() } catch (e) { /* ignore */ } }
  if (term) { try { term.dispose() } catch (e) { /* ignore */ } }
})
</script>

<style scoped>
.term-wrap { background: #0b1020; border-radius: 12px; padding: 10px; height: 100%; box-sizing: border-box; }
.term { width: 100%; height: 100%; }
</style>
