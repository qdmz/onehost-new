<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ t('menu.products') }}</h2>
        <div class="sub">按需选择配置，支持在线下单、即时开通</div>
      </div>
    </div>

    <div class="prod-grid">
      <div v-for="p in products" :key="p.id" class="prod-card">
        <div class="top">
          <div class="pname">{{ p.name }}</div>
          <el-tag v-if="p.stock === 0" type="danger" size="small">缺货</el-tag>
          <el-tag v-else-if="p.stock < 0" type="info" size="small">不限量</el-tag>
          <el-tag v-else type="success" size="small">有货</el-tag>
        </div>
        <div class="price">￥{{ yuan(p.price) }}<span class="per">/{{ p.period_days }}天</span></div>
        <div class="chips">
          <span class="chip">{{ p.cpu }} 核 CPU</span>
          <span class="chip">{{ p.memory_mb / 1024 }} GB 内存</span>
          <span class="chip">{{ p.disk_gb }} GB 硬盘</span>
          <span class="chip plain">{{ p.traffic_gb }} GB 流量</span>
        </div>
        <div class="desc">{{ p.description }}</div>
        <div class="foot">
          <span class="muted" style="font-size:12px">库存：{{ p.stock < 0 ? '不限' : p.stock }}</span>
          <el-button type="primary" :disabled="p.stock === 0" @click="buy(p)">立即购买</el-button>
        </div>
      </div>
      <div v-if="!products.length" class="empty" style="grid-column:1/-1">
        <div class="big">☁️</div>暂无可用产品
      </div>
    </div>

    <el-dialog v-model="dlg" title="确认下单" width="420px">
      <div v-if="cur">
        <div class="dname">{{ cur.name }}</div>
        <div class="row-flex" style="justify-content:space-between;margin:10px 0">
          <span class="muted">单价</span><b>￥{{ yuan(cur.price) }} / {{ cur.period_days }} 天</b>
        </div>
        <div class="row-flex" style="justify-content:space-between;margin:10px 0">
          <span class="muted">数量</span>
          <el-input-number v-model="qty" :min="1" :max="cur.stock < 0 ? 99 : cur.stock" />
        </div>
        <el-divider style="margin:10px 0" />
        <div class="row-flex" style="justify-content:space-between">
          <span>应付</span><span class="total">￥{{ yuan(cur.price * qty) }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" @click="doBuy">提交订单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import api from '../../api'

const { t } = useI18n()
const router = useRouter()
const products = ref([])
const dlg = ref(false)
const cur = ref(null)
const qty = ref(1)
const yuan = (c) => ((c || 0) / 100).toFixed(2)

async function load() {
  products.value = await api.get('/products')
}
function buy(p) {
  cur.value = p
  qty.value = 1
  dlg.value = true
}
async function doBuy() {
  await api.post('/user/orders', { product_id: cur.value.id, quantity: qty.value })
  ElMessage.success('下单成功，请前往订单页支付')
  dlg.value = false
  router.push('/console/orders')
}
onMounted(load)
</script>

<style scoped>
.prod-grid { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); }
.prod-card {
  display: flex; flex-direction: column;
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  box-shadow: var(--shadow-1); padding: 20px; transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease;
}
.prod-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-3); border-color: #dcdcfd; }
.prod-card .top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.pname { font-size: 17px; font-weight: 700; }
.price { color: var(--danger); font-size: 28px; font-weight: 800; margin: 10px 0 12px; }
.per { font-size: 13px; color: var(--muted); font-weight: 400; }
.chips { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 12px; }
.desc { color: var(--muted); font-size: 13px; min-height: 38px; line-height: 1.6; }
.foot { display: flex; align-items: center; justify-content: space-between; margin-top: auto; padding-top: 14px; }
.dname { font-weight: 700; font-size: 16px; }
.total { color: var(--danger); font-weight: 800; font-size: 20px; }
</style>
