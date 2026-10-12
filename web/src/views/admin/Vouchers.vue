<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>代金券管理</h2>
        <div class="sub">批量生成与作废</div>
      </div>
    </div>

    <div class="card mb">
      <h3 class="card-title">批量生成</h3>
      <el-form :inline="true" class="gen-form">
        <el-form-item label="面额(元)">
          <el-input-number v-model="gen.amount" :min="1" :precision="2" />
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="gen.count" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="有效期至">
          <el-date-picker v-model="gen.expired_at" type="date" value-format="YYYY-MM-DD" placeholder="可留空" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="generate"><el-icon><Plus /></el-icon>生成</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="lastBatch" type="success" :closable="false"
                :title="`最近批次：${lastBatch}（共 ${lastCount} 张）`" />
    </div>

    <div class="card" style="padding:0">
      <div class="table-wrap">
        <el-table :data="list" stripe>
          <el-table-column prop="code" label="券码" width="240" />
          <el-table-column label="面额" width="100">
            <template #default="{ row }"><b>￥{{ yuan(row.amount) }}</b></template>
          </el-table-column>
          <el-table-column prop="batch" label="批次" width="140" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="stType(row.status)" effect="light">{{ stText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="used_by" label="使用者ID" width="100" />
          <el-table-column label="有效期至" width="140">
            <template #default="{ row }">{{ row.expired_at ? row.expired_at.slice(0, 10) : '不限' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="110" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 0" size="small" type="danger" @click="voidIt(row)">作废</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const list = ref([])
const gen = reactive({ amount: 10, count: 10, expired_at: '' })
const lastBatch = ref('')
const lastCount = ref(0)
const yuan = (c) => ((c || 0) / 100).toFixed(2)
const stText = (s) => ({ 0: '未使用', 1: '已使用', 2: '已作废' }[s] || s)
const stType = (s) => ({ 0: 'success', 1: 'info', 2: 'danger' }[s] || '')

async function load() {
  list.value = await api.get('/admin/vouchers')
}
async function generate() {
  const d = await api.post('/admin/vouchers/generate', {
    amount: Math.round(gen.amount * 100),
    count: gen.count,
    expired_at: gen.expired_at || ''
  })
  lastBatch.value = d.batch
  lastCount.value = d.count
  ElMessage.success(`已生成 ${d.count} 张代金券`)
  load()
}
async function voidIt(row) {
  await ElMessageBox.confirm(`确认作废代金券 ${row.code} ？`, '确认', { type: 'warning' })
  await api.post(`/admin/vouchers/${row.id}/void`)
  ElMessage.success('已作废')
  load()
}
onMounted(load)
</script>

<style scoped>
.gen-form { row-gap: 6px; }
</style>
