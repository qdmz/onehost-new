<template>
  <div class="home">
    <SiteHeader />

    <!-- Hero -->
    <section class="hero">
      <div class="hero-inner">
        <div class="hero-copy">
          <div class="tag"><el-icon><Lightning /></el-icon> 一键下单 · 秒级开通</div>
          <h1>把云主机，做成<br /><span class="grad">人人买得起、开得快</span>的商品</h1>
          <p class="lead">
            产品定价、在线支付、自动开通、自助运维、到期续费 —— 完整的云主机售卖闭环，
            开箱即用，支持对接 LXD / Incus 虚拟化后端。
          </p>
          <div class="hero-cta">
            <el-button type="primary" size="large" round @click="start">免费开始使用</el-button>
            <el-button size="large" round @click="goConsole">
              <el-icon><Monitor /></el-icon>&nbsp;进入控制台
            </el-button>
          </div>
          <div class="hero-meta">
            <span><el-icon><Select /></el-icon> 无需信用卡</span>
            <span><el-icon><Select /></el-icon> 余额 / 代金券付费</span>
            <span><el-icon><Select /></el-icon> 7×24 稳定运行</span>
          </div>
        </div>

        <div class="hero-visual">
          <div class="mock">
            <div class="mock-top">
              <span class="dot red"></span><span class="dot amber"></span><span class="dot green"></span>
              <span class="mock-title">控制台 · 我的实例</span>
            </div>
            <div class="mock-body">
              <div class="mock-stat">
                <div><small>运行中实例</small><b>3</b></div>
                <div><small>账户余额</small><b>￥ 128.00</b></div>
              </div>
              <div v-for="i in 3" :key="i" class="mock-row">
                <span class="mi-dot"></span>
                <div class="mi">
                  <b>云主机 #{{ i }}</b>
                  <small>{{ i === 1 ? '1核1G · Debian 12' : i === 2 ? '2核2G · Ubuntu 22' : '1核2G · Debian 12' }}</small>
                </div>
                <span class="mi-ip">10.0.0.1{{ i }}</span>
                <span class="mi-tag">运行中</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 数据带 -->
    <section class="band">
      <div class="band-inner">
        <div v-for="s in stats" :key="s.label" class="band-item">
          <div class="bi-val">{{ s.val }}</div>
          <div class="bi-label">{{ s.label }}</div>
        </div>
      </div>
    </section>

    <!-- 特性 -->
    <section id="features" class="section">
      <div class="sec-head">
        <h2>为“卖云主机”而生的完整能力</h2>
        <p>从商品上架到实例运维，覆盖售卖闭环的每一个环节</p>
      </div>
      <div class="feat-grid">
        <div v-for="f in features" :key="f.title" class="feat-card">
          <div class="fi" :class="f.tone"><el-icon><component :is="f.icon" /></el-icon></div>
          <h3>{{ f.title }}</h3>
          <p>{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <!-- 使用流程 -->
    <section id="how" class="section alt">
      <div class="sec-head">
        <h2>四步开走，几分钟拿到云主机</h2>
        <p>从注册到可用实例，全程线上自助完成</p>
      </div>
      <div class="steps">
        <div v-for="(s, i) in steps" :key="s.t" class="step">
          <div class="step-no">{{ i + 1 }}</div>
          <h3>{{ s.t }}</h3>
          <p>{{ s.d }}</p>
        </div>
      </div>
    </section>

    <!-- 界面预览 -->
    <section id="preview" class="section">
      <div class="sec-head">
        <h2>真实界面预览</h2>
        <p>以下截图来自系统实际运行（用户控制台 · 云主机商城 · 实例管理 · 管理后台）</p>
      </div>
      <div class="shots">
        <figure v-for="s in shots" :key="s.src" class="shot">
          <div class="shot-bar">
            <span class="dot red"></span><span class="dot amber"></span><span class="dot green"></span>
            <span class="shot-name">{{ s.title }}</span>
          </div>
          <el-image :src="s.src" :preview-src-list="[s.src]" fit="cover" class="shot-img" />
          <figcaption>{{ s.desc }}</figcaption>
        </figure>
      </div>
    </section>

    <!-- 套餐 -->
    <section id="pricing" class="section alt">
      <div class="sec-head">
        <h2>简单透明的套餐定价</h2>
        <p>按周期计费，随时续费；库存、限购、节点绑定均可配置</p>
      </div>
      <div v-if="products.length" class="price-grid">
        <div v-for="(p, idx) in products" :key="p.id" class="price-card" :class="{ hot: idx === 1 }">
          <div v-if="idx === 1" class="hot-tag">最受欢迎</div>
          <h3>{{ p.name }}</h3>
          <div class="price"><small>￥</small><b>{{ yuan(p.price) }}</b><span>/ {{ p.period_days || 30 }}天</span></div>
          <p class="pdesc">{{ p.description || '稳定可靠的云主机实例' }}</p>
          <ul class="spec">
            <li><el-icon><Check /></el-icon> CPU {{ p.cpu }} 核</li>
            <li><el-icon><Check /></el-icon> 内存 {{ p.memory_mb }} MB</li>
            <li><el-icon><Check /></el-icon> 硬盘 {{ p.disk_gb }} GB</li>
            <li><el-icon><Check /></el-icon> 流量 {{ p.traffic_gb >= 9999999 ? '不限' : p.traffic_gb + ' GB' }}</li>
            <li><el-icon><Check /></el-icon> 镜像 {{ p.image || 'Debian/Ubuntu' }}</li>
          </ul>
          <el-button :type="idx === 1 ? 'primary' : 'default'" size="large" style="width:100%" round @click="buy">
            立即购买
          </el-button>
        </div>
      </div>
      <el-empty v-else description="暂无在售套餐" />
    </section>

    <!-- 用户评价 -->
    <section id="reviews" class="section">
      <div class="sec-head">
        <h2>他们在用 OneHost</h2>
        <p>来自站长、开发者与小型服务商的真实反馈</p>
      </div>
      <div class="rev-grid">
        <div v-for="r in reviews" :key="r.name" class="rev-card">
          <div class="stars">★★★★★</div>
          <p class="rev-text">{{ r.text }}</p>
          <div class="rev-who">
            <span class="rev-avatar">{{ r.name.slice(0, 1) }}</span>
            <div class="rev-meta"><b>{{ r.name }}</b><small>{{ r.role }}</small></div>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section id="faq" class="section alt">
      <div class="sec-head">
        <h2>常见问题</h2>
        <p>关于购买、支付与运维的常见疑问</p>
      </div>
      <div class="faq-wrap">
        <el-collapse v-model="activeFaq">
          <el-collapse-item v-for="(q, i) in faqs" :key="i" :title="q.q" :name="i">
            <div class="faq-a">{{ q.a }}</div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta">
      <div class="cta-inner">
        <h2>现在就开启你的第一台云主机</h2>
        <p>注册即用，余额或代金券都能付；下单后自动开通，开箱即可运维。</p>
        <el-button type="primary" size="large" round @click="start">免费注册</el-button>
      </div>
    </section>

    <SiteFooter />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import SiteHeader from '../components/SiteHeader.vue'
import SiteFooter from '../components/SiteFooter.vue'
import api from '../api'

const router = useRouter()
const user = useUserStore()
const products = ref([])
const activeFaq = ref(0)

const yuan = (c) => ((c || 0) / 100).toFixed(2)

const stats = [
  { val: '99.9%', label: '服务可用性' },
  { val: '秒级', label: '自动开通' },
  { val: 'LXD / Incus', label: '虚拟化后端' },
  { val: '7×24', label: '稳定运行' }
]

const features = [
  { icon: 'Goods', tone: '', title: '商品化上架', desc: '定价、库存、限购、节点绑定、上下架，产品管理一目了然。' },
  { icon: 'Lightning', tone: 'amber', title: '秒级自动开通', desc: '支付成功即异步创建实例、分配 IP、回填状态，无需人工介入。' },
  { icon: 'Wallet', tone: 'green', title: '钱包与充值', desc: '账户余额、在线充值、余额流水，支持未接网关时的模拟支付。' },
  { icon: 'Discount', tone: 'rose', title: '代金券', desc: '批量生成、兑换入账、防重复兑换，活动促销轻松搞。' },
  { icon: 'Monitor', tone: 'cyan', title: '自助运维', desc: '开机、关机、重启、续费、销毁，用户随时自助管理实例。' },
  { icon: 'Connection', tone: '', title: '多节点接入', desc: '对接 LXD / Incus（TLS 客户端证书），内置 Mock 便于本地试跑。' }
]

const steps = [
  { t: '注册账号', d: '邮箱即可注册，无需信用卡；登录后自动创建账户钱包。' },
  { t: '充值或兑换', d: '在线充值，或输入代金券券码兑换余额，二者都能直接下单。' },
  { t: '选择套餐下单', d: '在云主机商城挑配置，一键下单并支付，系统自动扣减库存。' },
  { t: '自动开通运维', d: '支付成功即异步创建实例并分配 IP，随时开机、重启、续费。' }
]

const shots = [
  { title: '营销首页', src: '/screenshots/home.png', desc: '产品、套餐与购买入口一目了然' },
  { title: '用户控制台', src: '/screenshots/console-dashboard.png', desc: '总览余额、实例与订单' },
  { title: '云主机商城', src: '/screenshots/console-products.png', desc: '按套餐自助下单' },
  { title: '我的实例', src: '/screenshots/console-instances.png', desc: '开机 / 关机 / 重启 / 续费 / 销毁' },
  { title: '我的订单', src: '/screenshots/console-orders.png', desc: '订单状态与支付流程' },
  { title: '管理后台', src: '/screenshots/admin-dashboard.png', desc: '平台运营数据总览' }
]

const reviews = [
  { name: '林工', role: '独立开发者', text: '从下单到拿到实例不到一分钟，开机、重启、续费都在一个页面搞定，省心。' },
  { name: 'VPS 小站', role: '小型服务商', text: '产品和节点都能在后台配置，库存和限购很实用，接到业务直接就能卖。' },
  { name: '阿岚', role: '运维工程师', text: '对接 LXD/Incus 的流程很清晰，本地还能用 Mock 节点先跑通，联调不折腾。' }
]

const faqs = [
  { q: '如何购买一台云主机？', a: '注册登录后进入「云主机商城」，选择套餐并下单，使用账户余额支付，系统会自动为你开通实例。' },
  { q: '怎么给账户充值？', a: '进入「钱包」创建充值订单；若平台已配置易支付网关即可在线支付，未配置时可使用「模拟支付」到账，方便测试。' },
  { q: '代金券怎么用？', a: '在「代金券」页面输入券码即可兑换为账户余额，余额可直接用于购买与续费。' },
  { q: '支持哪些虚拟化后端？', a: '支持 LXD / Incus 的 HTTPS API（客户端证书认证），并提供内置 Mock 节点用于本地无依赖试跑。' },
  { q: '实例到期怎么办？', a: '可在「我的订单」中随时续费，续费成功后实例有效期顺延。' }
]

function start() {
  router.push(user.token ? '/console/products' : '/register')
}
function goConsole() {
  router.push(user.token ? (user.isAdmin ? '/admin' : '/console') : '/login')
}
function buy() {
  router.push(user.token ? '/console/products' : '/register')
}

onMounted(async () => {
  try {
    const d = await api.get('/products')
    products.value = Array.isArray(d) ? d : (d.list || d.items || [])
  } catch (e) {}
  if (user.token && !user.user) { try { await user.fetchProfile() } catch (e) {} }
})
</script>

<style scoped>
.home { background: var(--bg); }

/* Hero */
.hero {
  background:
    radial-gradient(900px 480px at 88% -8%, rgba(139, 92, 246, .18), transparent 60%),
    radial-gradient(800px 500px at 0% 0%, rgba(99, 102, 241, .16), transparent 55%),
    linear-gradient(180deg, #f7f8fe, var(--bg));
  border-bottom: 1px solid var(--border);
}
.hero-inner {
  max-width: 1200px; margin: 0 auto; padding: 72px 22px 64px;
  display: grid; grid-template-columns: 1.05fr .95fr; gap: 48px; align-items: center;
}
.tag {
  display: inline-flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 999px;
  background: #fff; border: 1px solid var(--border-strong); color: var(--brand-1);
  font-size: 13px; font-weight: 600; box-shadow: var(--shadow-1);
}
.hero h1 { font-size: 44px; line-height: 1.22; margin: 20px 0 16px; letter-spacing: .5px; color: var(--text); }
.grad { background: var(--brand-grad); -webkit-background-clip: text; background-clip: text; color: transparent; }
.lead { color: var(--text-2); font-size: 16px; line-height: 1.85; max-width: 540px; }
.hero-cta { display: flex; gap: 12px; margin: 28px 0 18px; flex-wrap: wrap; }
.hero-meta { display: flex; gap: 20px; flex-wrap: wrap; color: var(--muted); font-size: 13.5px; }
.hero-meta span { display: inline-flex; align-items: center; gap: 6px; }
.hero-meta .el-icon { color: var(--success); }

/* Hero 视觉 */
.hero-visual { position: relative; }
.mock {
  background: #fff; border: 1px solid var(--border); border-radius: var(--radius-lg);
  box-shadow: 0 24px 60px rgba(79, 70, 229, .18); overflow: hidden;
}
.mock-top {
  display: flex; align-items: center; gap: 7px; padding: 13px 16px;
  background: #fafbff; border-bottom: 1px solid var(--border);
}
.mock-top .dot { width: 11px; height: 11px; border-radius: 50%; }
.dot.red { background: #ff5f57; } .dot.amber { background: #febc2e; } .dot.green { background: #28c840; }
.mock-title { margin-left: 10px; font-size: 12.5px; color: var(--muted); }
.mock-body { padding: 16px; }
.mock-stat { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 14px; }
.mock-stat > div {
  background: var(--brand-grad-soft); border-radius: 12px; padding: 12px 14px;
}
.mock-stat small { display: block; color: var(--muted); font-size: 12px; }
.mock-stat b { font-size: 20px; color: var(--text); }
.mock-row {
  display: flex; align-items: center; gap: 12px; padding: 12px;
  border: 1px solid var(--border); border-radius: 12px; margin-bottom: 10px;
}
.mi-dot { width: 9px; height: 9px; border-radius: 50%; background: var(--success); flex: none; }
.mi { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.mi b { font-size: 13.5px; color: var(--text); }
.mi small { color: var(--muted); font-size: 12px; }
.mi-ip { font-size: 12px; color: var(--text-2); font-family: ui-monospace, monospace; }
.mi-tag {
  font-size: 11px; color: var(--success); background: rgba(16,185,129,.12);
  padding: 3px 10px; border-radius: 999px; font-weight: 600;
}

/* 数据带 */
.band { background: #fff; border-bottom: 1px solid var(--border); }
.band-inner {
  max-width: 1200px; margin: 0 auto; padding: 26px 22px;
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px;
}
.band-item { text-align: center; }
.bi-val { font-size: 26px; font-weight: 800; background: var(--brand-grad); -webkit-background-clip: text; background-clip: text; color: transparent; }
.bi-label { color: var(--muted); font-size: 13px; margin-top: 4px; }

/* 通用区块 */
.section { max-width: 1200px; margin: 0 auto; padding: 68px 22px; }
.section.alt { max-width: none; background: #fff; border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.section.alt > .sec-head, .section.alt > .price-grid, .section.alt > .el-empty,
.section.alt > .steps, .section.alt > .faq-wrap { max-width: 1200px; margin-left: auto; margin-right: auto; }
.section.alt > .faq-wrap { max-width: 820px; }
.sec-head { text-align: center; margin-bottom: 42px; }
.sec-head h2 { font-size: 30px; margin: 0 0 10px; letter-spacing: .5px; }
.sec-head p { color: var(--muted); margin: 0; font-size: 15px; }

/* 特性 */
.feat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
.feat-card {
  background: #fff; border: 1px solid var(--border); border-radius: var(--radius);
  padding: 26px 24px; box-shadow: var(--shadow-1); transition: .18s;
}
.feat-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-2); }
.fi {
  width: 52px; height: 52px; border-radius: 14px; display: grid; place-items: center;
  font-size: 24px; color: #fff; background: var(--brand-grad); margin-bottom: 16px;
  box-shadow: 0 8px 18px rgba(99, 102, 241, .28);
}
.fi.green { background: linear-gradient(135deg, #10b981, #34d399); box-shadow: 0 8px 18px rgba(16,185,129,.28); }
.fi.amber { background: linear-gradient(135deg, #f59e0b, #fbbf24); box-shadow: 0 8px 18px rgba(245,158,11,.28); }
.fi.cyan { background: linear-gradient(135deg, #06b6d4, #22d3ee); box-shadow: 0 8px 18px rgba(6,182,212,.28); }
.fi.rose { background: linear-gradient(135deg, #f43f5e, #fb7185); box-shadow: 0 8px 18px rgba(244,63,94,.28); }
.feat-card h3 { margin: 0 0 8px; font-size: 17px; }
.feat-card p { margin: 0; color: var(--text-2); font-size: 14px; line-height: 1.7; }

/* 使用流程 */
.steps { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 20px; max-width: 1200px; margin: 0 auto; }
.step { position: relative; background: #fff; border: 1px solid var(--border); border-radius: var(--radius); padding: 26px 22px; box-shadow: var(--shadow-1); }
.step-no {
  width: 42px; height: 42px; border-radius: 12px; display: grid; place-items: center;
  font-size: 18px; font-weight: 800; color: #fff; background: var(--brand-grad);
  box-shadow: 0 8px 18px rgba(99,102,241,.3); margin-bottom: 14px;
}
.step h3 { margin: 0 0 6px; font-size: 16px; }
.step p { margin: 0; color: var(--text-2); font-size: 13.5px; line-height: 1.7; }

/* 界面预览 */
.shots { display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 24px; }
.shot {
  margin: 0; background: #fff; border: 1px solid var(--border); border-radius: var(--radius);
  overflow: hidden; box-shadow: var(--shadow-1); transition: .18s;
}
.shot:hover { transform: translateY(-3px); box-shadow: var(--shadow-2); }
.shot-bar {
  display: flex; align-items: center; gap: 6px; padding: 10px 14px;
  background: #fafbff; border-bottom: 1px solid var(--border);
}
.shot-bar .dot { width: 10px; height: 10px; border-radius: 50%; }
.shot-name { margin-left: 8px; font-size: 12px; color: var(--muted); }
.shot-img { display: block; width: 100%; height: 232px; background: #f4f6fb; cursor: zoom-in; }
.shot figcaption { padding: 12px 16px; font-size: 13.5px; color: var(--text-2); }

/* 套餐 */
.price-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 22px; }
.price-card {
  position: relative; background: #fff; border: 1px solid var(--border); border-radius: var(--radius-lg);
  padding: 30px 26px; box-shadow: var(--shadow-1); transition: .18s;
}
.price-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-2); }
.price-card.hot { border-color: var(--brand-1); box-shadow: 0 20px 46px rgba(99,102,241,.22); }
.hot-tag {
  position: absolute; top: -12px; left: 50%; transform: translateX(-50%);
  background: var(--brand-grad); color: #fff; font-size: 12px; font-weight: 700;
  padding: 5px 14px; border-radius: 999px; box-shadow: 0 8px 18px rgba(99,102,241,.4);
}
.price-card h3 { margin: 0 0 12px; font-size: 18px; }
.price { display: flex; align-items: baseline; gap: 4px; margin-bottom: 12px; }
.price small { font-size: 16px; color: var(--brand-1); font-weight: 700; }
.price b { font-size: 38px; font-weight: 800; color: var(--text); line-height: 1; }
.price span { color: var(--muted); font-size: 13px; margin-left: 4px; }
.pdesc { color: var(--muted); font-size: 13.5px; margin: 0 0 18px; min-height: 38px; }
.spec { list-style: none; padding: 0; margin: 0 0 22px; }
.spec li { display: flex; align-items: center; gap: 8px; padding: 6px 0; color: var(--text-2); font-size: 14px; }
.spec .el-icon { color: var(--success); }

/* 评价 */
.rev-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
.rev-card { background: #fff; border: 1px solid var(--border); border-radius: var(--radius); padding: 26px 24px; box-shadow: var(--shadow-1); }
.stars { color: #f5a623; letter-spacing: 2px; font-size: 15px; margin-bottom: 12px; }
.rev-text { margin: 0 0 18px; color: var(--text-2); font-size: 14.5px; line-height: 1.8; }
.rev-who { display: flex; align-items: center; gap: 12px; }
.rev-avatar {
  width: 42px; height: 42px; border-radius: 50%; display: grid; place-items: center;
  font-weight: 800; color: #fff; background: var(--brand-grad); font-size: 16px;
}
.rev-meta { display: flex; flex-direction: column; line-height: 1.3; }
.rev-meta b { font-size: 14.5px; color: var(--text); }
.rev-meta small { color: var(--muted); font-size: 12.5px; }

/* FAQ */
.faq-wrap { max-width: 820px; margin: 0 auto; }
.faq-a { color: var(--text-2); line-height: 1.85; font-size: 14.5px; }

/* CTA */
.cta { padding: 0 22px 72px; }
.cta-inner {
  max-width: 1200px; margin: 0 auto; text-align: center; color: #fff;
  background: var(--brand-grad); border-radius: var(--radius-lg);
  padding: 56px 24px; box-shadow: 0 24px 60px rgba(99,102,241,.32);
  position: relative; overflow: hidden;
}
.cta-inner h2 { margin: 0 0 12px; font-size: 28px; }
.cta-inner p { margin: 0 0 26px; opacity: .92; font-size: 15px; }
.cta-inner .el-button { background: #fff; color: var(--brand-1); border: none; font-weight: 700; }

/* 响应式 */
@media (max-width: 992px) {
  .hero-inner { grid-template-columns: 1fr; gap: 34px; padding: 48px 20px 44px; }
  .hero h1 { font-size: 34px; }
  .hero-visual { max-width: 520px; }
}
@media (max-width: 768px) {
  .band-inner { grid-template-columns: repeat(2, 1fr); gap: 22px; }
  .section { padding: 46px 16px; }
  .sec-head h2 { font-size: 23px; }
  .sec-head { margin-bottom: 30px; }
  .hero h1 { font-size: 29px; }
  .lead { font-size: 14.5px; }
  .price b { font-size: 32px; }
  .shots { grid-template-columns: 1fr; gap: 18px; }
  .shot-img { height: 200px; }
}
</style>
