<template>
  <section class="dashboard-page">
    <div class="aurora aurora-blue"></div>
    <div class="aurora aurora-coral"></div>
    <div class="aurora aurora-pink"></div>

    <div class="dashboard-shell">
      <header class="dashboard-hero glass-card">
        <div class="hero-copy">
          <p class="eyebrow">Xingyunpan V2 Console</p>
          <h1>面板首页</h1>
          <p>
            统一观察存储、流量、在线用户与节点健康，让星云盘后台的关键态势在第一屏完成判断。
          </p>
        </div>

        <div class="hero-toolbar" aria-label="控制台状态">
          <button class="icon-action" type="button" :disabled="dashboardLoading" aria-label="刷新概览" @click="loadDashboard">
            <Refresh />
          </button>
          <button class="icon-action" type="button" aria-label="系统设置">
            <Setting />
          </button>
          <div class="notify-dot" aria-label="3 条未读通知">3</div>
        </div>
      </header>

      <section class="metric-grid" aria-label="核心指标">
        <article
          v-for="item in metricCards"
          :key="item.label"
          class="metric-card glass-card"
          :class="item.tone"
        >
          <div class="metric-head">
            <span class="metric-icon">
              <component :is="item.icon" />
            </span>
            <span class="metric-change">{{ item.change }}</span>
          </div>
          <strong>{{ item.value }}</strong>
          <span>{{ item.label }}</span>
          <small>{{ item.detail }}</small>
        </article>
      </section>

      <section class="bento-grid">
        <article class="storage-panel glass-card">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Storage Used</p>
              <h2>存储空间</h2>
            </div>
            <span class="panel-chip">{{ formatBytes(dashboard?.storage.used_bytes || 0) }} / {{ formatBytes(dashboard?.storage.total_bytes || 0) }}</span>
          </div>

          <div class="donut-wrap">
            <svg class="storage-donut" viewBox="0 0 220 220" :aria-label="`存储使用率 ${storagePercent}%`">
              <defs>
                <linearGradient id="storageRingGradient" x1="32" y1="38" x2="190" y2="188">
                  <stop offset="0%" stop-color="#7dd3fc" />
                  <stop offset="48%" stop-color="#1d9bf0" />
                  <stop offset="100%" stop-color="#2563eb" />
                </linearGradient>
                <filter id="storageGlow" x="-40%" y="-40%" width="180%" height="180%">
                  <feGaussianBlur stdDeviation="4" result="blur" />
                  <feMerge>
                    <feMergeNode in="blur" />
                    <feMergeNode in="SourceGraphic" />
                  </feMerge>
                </filter>
              </defs>
              <circle class="donut-track" cx="110" cy="110" r="78" />
              <circle
                class="donut-progress"
                cx="110"
                cy="110"
                r="78"
                :stroke-dasharray="storageCircumference"
                :stroke-dashoffset="storageOffset"
              />
            </svg>
            <div class="donut-center">
              <strong>{{ storagePercentLabel }}</strong>
              <span>已使用</span>
            </div>
          </div>

          <div class="storage-breakdown">
            <div v-for="item in storageBreakdown" :key="item.name" class="breakdown-row">
              <span class="breakdown-dot" :class="item.tone"></span>
              <span>{{ item.name }}</span>
              <strong>{{ formatBytes(item.size_bytes) }}</strong>
            </div>
          </div>
        </article>

        <article class="traffic-panel glass-card">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Network Traffic Trends</p>
              <h2>今日流量趋势</h2>
            </div>
            <div class="legend">
              <span class="legend-blue">入站</span>
              <span class="legend-amber">出站</span>
            </div>
          </div>

          <div class="traffic-chart">
            <svg viewBox="0 0 880 340" preserveAspectRatio="none" aria-hidden="true">
              <defs>
                <linearGradient id="inboundArea" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.28" />
                  <stop offset="100%" stop-color="#38bdf8" stop-opacity="0" />
                </linearGradient>
                <linearGradient id="outboundArea" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#f59e0b" stop-opacity="0.22" />
                  <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
                </linearGradient>
                <filter id="blueLineGlow" x="-10%" y="-40%" width="120%" height="180%">
                  <feGaussianBlur stdDeviation="4" result="blur" />
                  <feMerge>
                    <feMergeNode in="blur" />
                    <feMergeNode in="SourceGraphic" />
                  </feMerge>
                </filter>
                <filter id="amberLineGlow" x="-10%" y="-40%" width="120%" height="180%">
                  <feGaussianBlur stdDeviation="4" result="blur" />
                  <feMerge>
                    <feMergeNode in="blur" />
                    <feMergeNode in="SourceGraphic" />
                  </feMerge>
                </filter>
              </defs>

              <g class="chart-grid">
                <line v-for="y in chartGridY" :key="`y-${y}`" x1="48" :y1="y" x2="842" :y2="y" />
                <line v-for="x in chartGridX" :key="`x-${x}`" :x1="x" y1="34" :x2="x" y2="286" />
              </g>
              <path class="area inbound" :d="trafficAreas.inbound" />
              <path class="area outbound" :d="trafficAreas.outbound" />
              <path class="traffic-line inbound" :d="trafficLines.inbound" />
              <path class="traffic-line outbound" :d="trafficLines.outbound" />
              <circle
                v-for="(point, index) in plottedTraffic.inbound"
                :key="`in-${index}`"
                class="line-dot inbound"
                :cx="point.x"
                :cy="point.y"
                r="4"
              />
              <circle
                v-for="(point, index) in plottedTraffic.outbound"
                :key="`out-${index}`"
                class="line-dot outbound"
                :cx="point.x"
                :cy="point.y"
                r="4"
              />
            </svg>

            <div class="chart-callout">
              <span>峰值吞吐</span>
              <strong>{{ peakTrafficLabel }}</strong>
              <small>{{ dashboard?.traffic.peak_window || '--' }}</small>
            </div>
          </div>

          <div class="time-axis">
            <span v-for="item in trafficLabels" :key="item">{{ item }}</span>
          </div>
        </article>

        <article class="online-panel glass-card">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Online Users</p>
              <h2>当前在线用户</h2>
            </div>
            <span class="live-pill"><i></i> Live</span>
          </div>

          <div class="online-main">
            <strong>{{ formatCompact(dashboard?.online.current_sessions || 0) }}</strong>
            <span>近 5 分钟活跃</span>
          </div>

          <div class="user-stack" aria-hidden="true">
            <span v-for="item in userOrbs" :key="item" :class="item"></span>
          </div>

          <div class="online-list">
            <div v-for="item in onlineGroups" :key="item.name" class="online-row">
              <span>{{ item.name }}</span>
              <strong>{{ formatCompact(item.value) }}</strong>
              <em :style="{ width: `${item.percent}%` }"></em>
            </div>
          </div>
        </article>

        <article class="node-panel glass-card">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Nodes</p>
              <h2>节点健康</h2>
            </div>
            <span class="panel-chip success">{{ dashboard?.nodes.online || 0 }} / {{ dashboard?.nodes.total || 0 }} 在线</span>
          </div>

          <div class="node-list">
            <div v-for="node in nodeHealth" :key="node.name" class="node-row">
              <span class="node-status"></span>
              <div>
                <strong>{{ node.name }}</strong>
                <small>{{ node.region }} · {{ node.latency_ms > 0 ? `${node.latency_ms}ms` : '暂无延迟' }}</small>
              </div>
              <span>{{ node.load }}%</span>
            </div>
          </div>
        </article>

        <article class="task-panel glass-card">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Tasks</p>
              <h2>后台任务</h2>
            </div>
            <span class="panel-chip warm">{{ runningTaskCount }} 运行中</span>
          </div>

          <div class="task-list">
            <div v-for="task in tasks" :key="task.name" class="task-row">
              <div>
                <strong>{{ task.name }}</strong>
                <span>{{ task.detail }}</span>
              </div>
              <small>{{ task.progress }}%</small>
              <em :style="{ width: `${task.progress}%` }"></em>
            </div>
          </div>
        </article>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  Files,
  Refresh,
  Setting,
  Timer,
  Upload,
  User,
} from '@element-plus/icons-vue';
import {
  getDashboardOverview,
  type DashboardNodePayload,
  type DashboardOverviewPayload,
  type DashboardTaskPayload,
} from '@/api/dashboard';

type MetricCard = {
  label: string;
  value: string;
  detail: string;
  change: string;
  tone: string;
  icon: unknown;
};

type Point = {
  x: number;
  y: number;
};

const dashboard = ref<DashboardOverviewPayload | null>(null);
const dashboardLoading = ref(false);

const storageRadius = 78;
const storageCircumference = 2 * Math.PI * storageRadius;
const storagePercent = computed(() => dashboard.value?.storage.percent || 0);
const storagePercentLabel = computed(() => `${formatNumber(storagePercent.value)}%`);
const storageOffset = computed(() => storageCircumference * (1 - storagePercent.value / 100));

const metricCards = computed<MetricCard[]>(() => {
  const metrics = dashboard.value?.metrics;
  return [
    {
      label: '今日上传',
      value: formatCompact(metrics?.today_upload_count || 0),
      detail: `${formatBytes(metrics?.today_upload_bytes || 0)} 写入`,
      change: formatPercent(metrics?.upload_change_pct || 0),
      tone: 'tone-blue',
      icon: Upload,
    },
    {
      label: '平均延迟',
      value: metrics?.average_latency_ms ? `${metrics.average_latency_ms}ms` : '--',
      detail: metrics?.average_latency_ms ? '来自节点巡检记录' : '暂无节点延迟数据',
      change: formatPercent(metrics?.latency_change_pct || 0),
      tone: 'tone-cyan',
      icon: Timer,
    },
    {
      label: '活跃用户',
      value: formatCompact(metrics?.active_users || 0),
      detail: '当前启用账户数',
      change: formatPercent(metrics?.active_users_change_pct || 0),
      tone: 'tone-coral',
      icon: User,
    },
    {
      label: '文件对象',
      value: formatCompact(metrics?.blob_count || 0),
      detail: `${formatCompact(metrics?.file_count || 0)} 文件 / ${formatCompact(metrics?.folder_count || 0)} 目录`,
      change: formatPercent(metrics?.blob_count_change_pct || 0),
      tone: 'tone-violet',
      icon: Files,
    },
  ];
});

const storageBreakdown = computed(() => dashboard.value?.storage.breakdown || []);
const trafficLabels = computed(() => dashboard.value?.traffic.labels || ['00:00', '03:00', '06:00', '09:00', '12:00', '15:00', '18:00', '21:00']);
const trafficSeries = computed(() => ({
  inbound: normalizeSeries(dashboard.value?.traffic.inbound),
  outbound: normalizeSeries(dashboard.value?.traffic.outbound),
}));

const chartOriginX = 48;
const chartOriginY = 34;
const chartWidth = 794;
const chartHeight = 252;
const chartBaseline = chartOriginY + chartHeight;
const chartGridY = Array.from({ length: 5 }, (_, index) => chartOriginY + (chartHeight / 4) * index);
const chartGridX = computed(() => {
  const count = Math.max(trafficLabels.value.length, 2);
  return Array.from({ length: trafficLabels.value.length }, (_, index) => chartOriginX + (chartWidth / (count - 1)) * index);
});

function projectSeries(series: number[]): Point[] {
  const allValues = [...trafficSeries.value.inbound, ...trafficSeries.value.outbound];
  const max = Math.max(...allValues, 1);
  const min = Math.min(...allValues, 0);
  const spread = Math.max(max - min, 1);

  return series.map((value, index) => {
    const count = Math.max(series.length, 2);
    const x = chartOriginX + (chartWidth / (count - 1)) * index;
    const y = chartOriginY + chartHeight - ((value - min) / spread) * (chartHeight - 34);
    return { x: Number(x.toFixed(2)), y: Number(y.toFixed(2)) };
  });
}

function createSmoothLine(points: Point[]) {
  if (!points.length) return '';
  return points.slice(1).reduce((path, point, index) => {
    const previous = points[index];
    const controlX = (previous.x + point.x) / 2;
    return `${path} C ${controlX} ${previous.y}, ${controlX} ${point.y}, ${point.x} ${point.y}`;
  }, `M ${points[0].x} ${points[0].y}`);
}

function createAreaPath(points: Point[], linePath: string) {
  if (!points.length || !linePath) return '';
  const start = points[0];
  const end = points[points.length - 1];
  return `${linePath} L ${end.x} ${chartBaseline} L ${start.x} ${chartBaseline} Z`;
}

const plottedTraffic = computed(() => ({
  inbound: projectSeries(trafficSeries.value.inbound),
  outbound: projectSeries(trafficSeries.value.outbound),
}));

const trafficLines = computed(() => ({
  inbound: createSmoothLine(plottedTraffic.value.inbound),
  outbound: createSmoothLine(plottedTraffic.value.outbound),
}));

const trafficAreas = computed(() => ({
  inbound: createAreaPath(plottedTraffic.value.inbound, trafficLines.value.inbound),
  outbound: createAreaPath(plottedTraffic.value.outbound, trafficLines.value.outbound),
}));

const userOrbs = ['orb-a', 'orb-b', 'orb-c', 'orb-d', 'orb-e'];
const onlineGroups = computed(() => dashboard.value?.online.groups || []);
const peakTrafficLabel = computed(() => formatBytes(dashboard.value?.traffic.peak_value || 0));

const nodeHealth = computed<DashboardNodePayload[]>(() => dashboard.value?.nodes.items || []);
const tasks = computed<DashboardTaskPayload[]>(() => dashboard.value?.tasks || []);
const runningTaskCount = computed(() => tasks.value.reduce((total, task) => total + (task.processing || 0), 0));

function normalizeSeries(series?: number[]) {
  const normalized = series && series.length ? series : Array.from({ length: trafficLabels.value.length }, () => 0);
  return normalized.map((value) => Number.isFinite(value) ? Number(value) : 0);
}

function formatCompact(value: number) {
  return new Intl.NumberFormat('zh-CN', { notation: value >= 10000 ? 'compact' : 'standard', maximumFractionDigits: 1 }).format(value);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 }).format(value);
}

function formatPercent(value: number) {
  const sign = value > 0 ? '+' : '';
  return `${sign}${formatNumber(value)}%`;
}

function formatBytes(bytes: number) {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  let value = bytes;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${new Intl.NumberFormat('zh-CN', { maximumFractionDigits: value >= 10 || index === 0 ? 0 : 1 }).format(value)} ${units[index]}`;
}

async function loadDashboard() {
  dashboardLoading.value = true;
  try {
    dashboard.value = await getDashboardOverview();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载面板首页数据失败');
  } finally {
    dashboardLoading.value = false;
  }
}

onMounted(loadDashboard);
</script>

<style scoped>
.dashboard-page {
  position: relative;
  min-height: calc(100vh - 96px);
  overflow: hidden;
  padding: 14px 24px 24px;
  border-radius: 32px;
  background:
    radial-gradient(circle at 10% 8%, rgba(125, 211, 252, 0.28), transparent 28%),
    radial-gradient(circle at 94% 2%, rgba(255, 177, 191, 0.3), transparent 26%),
    radial-gradient(circle at 80% 96%, rgba(250, 204, 21, 0.12), transparent 24%),
    linear-gradient(135deg, #fff9ef 0%, #f7fbff 48%, #fff4f6 100%);
}

.aurora {
  position: absolute;
  border-radius: 999px;
  filter: blur(36px);
  opacity: 0.72;
  pointer-events: none;
}

.aurora-blue {
  top: 8%;
  left: 24%;
  width: 360px;
  height: 220px;
  background: rgba(56, 189, 248, 0.26);
}

.aurora-coral {
  right: -80px;
  top: 18%;
  width: 300px;
  height: 300px;
  background: rgba(251, 113, 133, 0.2);
}

.aurora-pink {
  left: -90px;
  bottom: -90px;
  width: 360px;
  height: 320px;
  background: rgba(216, 180, 254, 0.18);
}

.dashboard-shell {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 18px;
}

.glass-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 28px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.76), rgba(255, 255, 255, 0.48));
  box-shadow:
    0 24px 58px rgba(67, 93, 121, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    inset 0 0 0 1px rgba(255, 255, 255, 0.22);
  backdrop-filter: blur(18px);
}

.glass-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.7), transparent 36%, rgba(255, 255, 255, 0.32));
  opacity: 0.72;
  pointer-events: none;
}

.dashboard-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  min-height: 168px;
  padding: 28px;
}

.hero-copy,
.hero-toolbar,
.panel-head,
.metric-card > *,
.donut-wrap,
.storage-breakdown,
.traffic-chart,
.time-axis,
.online-main,
.user-stack,
.online-list,
.node-list,
.task-list {
  position: relative;
  z-index: 1;
}

.eyebrow {
  margin: 0 0 10px;
  color: #0f8fd8;
  font-size: 12px;
  font-weight: 850;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.hero-copy h1,
.panel-head h2 {
  margin: 0;
  color: #132238;
}

.hero-copy h1 {
  font-size: clamp(42px, 5vw, 72px);
  line-height: 1;
  font-weight: 860;
}

.hero-copy p:last-child {
  max-width: 720px;
  margin: 14px 0 0;
  color: #526173;
  font-size: 15px;
  line-height: 1.8;
}

.hero-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-action {
  display: inline-grid;
  place-items: center;
  width: 46px;
  height: 46px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.58);
  color: #2563eb;
  box-shadow: 0 12px 26px rgba(59, 130, 246, 0.12);
  cursor: pointer;
}

.icon-action :deep(svg) {
  width: 20px;
  height: 20px;
}

.notify-dot {
  display: inline-grid;
  place-items: center;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: linear-gradient(135deg, #fb7185, #fda4af);
  color: #fff;
  font-weight: 850;
  box-shadow:
    0 0 0 8px rgba(251, 113, 133, 0.12),
    0 14px 30px rgba(251, 113, 133, 0.24);
  animation: coralPulse 2.5s ease-in-out infinite;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

.metric-card {
  display: grid;
  gap: 10px;
  min-height: 178px;
  padding: 20px;
}

.metric-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.metric-icon {
  display: inline-grid;
  place-items: center;
  width: 48px;
  height: 48px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.62);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.metric-icon :deep(svg) {
  width: 22px;
  height: 22px;
}

.metric-change {
  padding: 7px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.58);
  color: #2f4c68;
  font-size: 12px;
  font-weight: 800;
}

.metric-card strong {
  color: #122235;
  font-size: 34px;
  line-height: 1;
  font-weight: 860;
  text-shadow: 0 0 18px rgba(56, 189, 248, 0.14);
}

.metric-card span:not(.metric-icon, .metric-change),
.metric-card small {
  color: #5f6d7d;
}

.metric-card span:not(.metric-icon, .metric-change) {
  font-weight: 760;
}

.metric-card small {
  font-size: 13px;
}

.tone-blue {
  color: #1687d9;
  box-shadow:
    0 22px 52px rgba(56, 189, 248, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.95);
}

.tone-cyan {
  color: #0891b2;
}

.tone-coral {
  color: #fb7185;
}

.tone-violet {
  color: #7c3aed;
}

.bento-grid {
  display: grid;
  grid-template-columns: minmax(310px, 0.92fr) minmax(0, 1.48fr) minmax(280px, 0.86fr);
  grid-auto-rows: minmax(220px, auto);
  gap: 18px;
}

.storage-panel,
.traffic-panel,
.online-panel,
.node-panel,
.task-panel {
  padding: 24px;
}

.traffic-panel {
  grid-column: span 2;
}

.storage-panel {
  display: grid;
  align-content: space-between;
  min-height: 540px;
}

.online-panel {
  min-height: 540px;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-head h2 {
  font-size: 24px;
}

.panel-chip,
.live-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 36px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.62);
  color: #416078;
  font-size: 12px;
  font-weight: 820;
  white-space: nowrap;
}

.panel-chip.success {
  color: #059669;
  background: rgba(209, 250, 229, 0.62);
}

.panel-chip.warm {
  color: #b45309;
  background: rgba(254, 243, 199, 0.7);
}

.donut-wrap {
  display: grid;
  place-items: center;
  margin: 14px 0 10px;
  min-height: 250px;
}

.storage-donut {
  width: min(260px, 100%);
  filter: drop-shadow(0 0 16px rgba(56, 189, 248, 0.26));
  transform: rotate(-90deg);
}

.donut-track,
.donut-progress {
  fill: none;
  stroke-width: 20;
}

.donut-track {
  stroke: rgba(148, 163, 184, 0.16);
}

.donut-progress {
  stroke: url(#storageRingGradient);
  stroke-linecap: round;
  filter: url(#storageGlow);
  animation: breatheStroke 3.2s ease-in-out infinite;
}

.donut-center {
  position: absolute;
  display: grid;
  gap: 6px;
  text-align: center;
}

.donut-center strong {
  color: #132238;
  font-size: 44px;
  line-height: 1;
  font-weight: 880;
}

.donut-center span {
  color: #64748b;
  font-size: 13px;
  font-weight: 760;
}

.storage-breakdown,
.online-list,
.node-list,
.task-list {
  display: grid;
  gap: 12px;
}

.breakdown-row,
.online-row,
.node-row {
  display: grid;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.64);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.48);
}

.breakdown-row {
  grid-template-columns: auto minmax(0, 1fr) auto;
}

.breakdown-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.breakdown-dot.blue {
  background: #38bdf8;
  box-shadow: 0 0 14px rgba(56, 189, 248, 0.58);
}

.breakdown-dot.pink {
  background: #fb7185;
  box-shadow: 0 0 14px rgba(251, 113, 133, 0.5);
}

.breakdown-dot.amber {
  background: #f59e0b;
  box-shadow: 0 0 14px rgba(245, 158, 11, 0.46);
}

.breakdown-row span,
.online-row span,
.node-row small,
.task-row span,
.time-axis {
  color: #64748b;
}

.breakdown-row strong,
.online-row strong,
.node-row strong,
.task-row strong {
  color: #132238;
}

.legend {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.legend span {
  position: relative;
  padding-left: 16px;
  color: #64748b;
  font-size: 13px;
  font-weight: 760;
}

.legend span::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 9px;
  height: 9px;
  border-radius: 999px;
  transform: translateY(-50%);
}

.legend-blue::before {
  background: #38bdf8;
  box-shadow: 0 0 10px rgba(56, 189, 248, 0.7);
}

.legend-amber::before {
  background: #f59e0b;
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.6);
}

.traffic-chart {
  margin-top: 18px;
  min-height: 340px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.42), rgba(255, 255, 255, 0.22));
}

.traffic-chart svg {
  width: 100%;
  height: 340px;
}

.chart-grid line {
  stroke: rgba(100, 116, 139, 0.13);
  stroke-width: 1;
}

.area.inbound {
  fill: url(#inboundArea);
}

.area.outbound {
  fill: url(#outboundArea);
}

.traffic-line {
  fill: none;
  stroke-width: 5;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.traffic-line.inbound {
  stroke: #35bdf8;
  filter: url(#blueLineGlow);
}

.traffic-line.outbound {
  stroke: #f59e0b;
  filter: url(#amberLineGlow);
}

.line-dot {
  fill: #fff;
  stroke-width: 3;
}

.line-dot.inbound {
  stroke: #38bdf8;
}

.line-dot.outbound {
  stroke: #f59e0b;
}

.chart-callout {
  position: absolute;
  top: 94px;
  right: 34px;
  display: grid;
  gap: 4px;
  min-width: 138px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.64);
  border-radius: 18px;
  background: rgba(19, 34, 56, 0.78);
  color: #fff;
  box-shadow: 0 18px 36px rgba(19, 34, 56, 0.18);
}

.chart-callout span,
.chart-callout small {
  color: rgba(255, 255, 255, 0.7);
}

.chart-callout strong {
  font-size: 22px;
}

.time-axis {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 8px;
  margin-top: 10px;
  padding: 0 14px;
  font-size: 12px;
  text-align: center;
}

.live-pill {
  gap: 8px;
  color: #0f9f8f;
}

.live-pill i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #14b8a6;
  box-shadow: 0 0 0 6px rgba(20, 184, 166, 0.12);
}

.online-main {
  display: grid;
  gap: 8px;
  margin-top: 28px;
  text-align: center;
}

.online-main strong {
  color: #132238;
  font-size: 62px;
  line-height: 1;
  font-weight: 880;
  text-shadow: 0 0 24px rgba(20, 184, 166, 0.14);
}

.online-main span {
  color: #64748b;
  font-weight: 760;
}

.user-stack {
  display: flex;
  justify-content: center;
  margin: 28px 0;
}

.user-stack span {
  width: 48px;
  height: 48px;
  margin-left: -10px;
  border: 3px solid rgba(255, 255, 255, 0.8);
  border-radius: 50%;
  box-shadow: 0 12px 26px rgba(72, 93, 116, 0.14);
}

.user-stack span:first-child {
  margin-left: 0;
}

.orb-a { background: linear-gradient(135deg, #38bdf8, #2563eb); }
.orb-b { background: linear-gradient(135deg, #fb7185, #fbcfe8); }
.orb-c { background: linear-gradient(135deg, #14b8a6, #99f6e4); }
.orb-d { background: linear-gradient(135deg, #f59e0b, #fde68a); }
.orb-e { background: linear-gradient(135deg, #8b5cf6, #ddd6fe); }

.online-row {
  position: relative;
  grid-template-columns: minmax(0, 1fr) auto;
  overflow: hidden;
}

.online-row em,
.task-row em {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, #38bdf8, #fb7185);
  box-shadow: 0 0 12px rgba(56, 189, 248, 0.42);
}

.node-list,
.task-list {
  margin-top: 20px;
}

.node-row {
  grid-template-columns: auto minmax(0, 1fr) auto;
}

.node-status {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 0 7px rgba(34, 197, 94, 0.12), 0 0 16px rgba(34, 197, 94, 0.5);
}

.node-row div {
  display: grid;
  gap: 4px;
}

.task-row {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  overflow: hidden;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.64);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.48);
}

.task-row div {
  display: grid;
  gap: 5px;
}

.task-row small {
  color: #0f8fd8;
  font-weight: 840;
}

@keyframes coralPulse {
  0%,
  100% {
    box-shadow:
      0 0 0 8px rgba(251, 113, 133, 0.12),
      0 14px 30px rgba(251, 113, 133, 0.24);
  }

  50% {
    box-shadow:
      0 0 0 13px rgba(251, 113, 133, 0.04),
      0 18px 38px rgba(251, 113, 133, 0.3);
  }
}

@keyframes breatheStroke {
  0%,
  100% {
    opacity: 0.92;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 1380px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bento-grid {
    grid-template-columns: minmax(300px, 0.92fr) minmax(0, 1fr);
  }

  .traffic-panel {
    grid-column: span 1;
  }

  .online-panel,
  .storage-panel {
    min-height: 500px;
  }
}

@media (max-width: 980px) {
  .dashboard-page {
    padding: 18px;
  }

  .dashboard-hero,
  .panel-head {
    flex-direction: column;
  }

  .hero-toolbar {
    align-self: flex-start;
  }

  .bento-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 680px) {
  .dashboard-page {
    padding: 14px;
    border-radius: 24px;
  }

  .metric-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-hero,
  .storage-panel,
  .traffic-panel,
  .online-panel,
  .node-panel,
  .task-panel,
  .metric-card {
    padding: 18px;
  }

  .time-axis {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .chart-callout {
    position: static;
    margin: 0 12px 12px;
  }

  .traffic-chart,
  .traffic-chart svg {
    min-height: 280px;
    height: 280px;
  }
}
</style>
