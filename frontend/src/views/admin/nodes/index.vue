<template>
  <section class="nodes-page">
    <div v-if="!activeNode" class="nodes-shell">
      <header class="nodes-hero">
        <div>
          <p class="nodes-eyebrow">Node Management</p>
          <h1>节点管理</h1>
          <p class="nodes-subtitle">
            管理主节点、边缘工作节点与离线下载能力，确保存储策略能够在健康节点上稳定调度。
          </p>
        </div>
        <button class="hero-help" type="button" aria-label="节点帮助">?</button>
      </header>

      <section class="metric-grid">
        <article v-for="metric in metrics" :key="metric.label" class="metric-card">
          <span class="metric-label">{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
          <small>{{ metric.detail }}</small>
        </article>
      </section>

      <section class="nodes-toolbar">
        <button class="toolbar-button" type="button" :disabled="loading" @click="refreshNodes">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M16.25 10a6.25 6.25 0 1 1-1.52-4.1" />
            <path d="M16.25 4.75v4.5h-4.5" />
          </svg>
          <span>{{ loading ? '刷新中' : '刷新' }}</span>
        </button>
        <button class="toolbar-button" type="button" :disabled="healthCheckingAll" @click="runHealthChecks">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M10 3.5v2.5" />
            <path d="M10 14v2.5" />
            <path d="M4.4 5.4 6.2 7.2" />
            <path d="m13.8 12.8 1.8 1.8" />
            <path d="M3.5 10H6" />
            <path d="M14 10h2.5" />
            <path d="m4.4 14.6 1.8-1.8" />
            <path d="m13.8 7.2 1.8-1.8" />
            <circle cx="10" cy="10" r="2.4" />
          </svg>
          <span>{{ healthCheckingAll ? '巡检中' : '巡检节点' }}</span>
        </button>
      </section>

      <section class="node-board">
        <button class="create-node-card" type="button" @click="createNode">
          <span class="create-node-plus">+</span>
          <span>新建节点</span>
        </button>

        <article
          v-for="node in pagedNodes"
          :key="node.id"
          class="node-card"
          :class="{ 'is-disabled': !node.enabled }"
          @click="openNode(node.id)"
        >
          <div class="node-card-main">
            <div class="node-card-header">
              <span class="node-health-label">
                <span class="status-led" :class="healthLedClass(node.health.status)"></span>
                {{ healthStatusLabel(node.health.status) }}
              </span>
              <span class="node-role-badge" :class="`is-${node.type}`">
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="m10 2.5 2.2 4.46 4.93.72-3.56 3.47.84 4.91L10 13.72 5.6 16.06l.84-4.91-3.56-3.47 4.93-.72Z" />
                </svg>
                {{ node.typeLabel }}
              </span>
            </div>

            <div class="node-name-row">
              <strong>{{ node.name }}</strong>
              <small>权重 {{ node.weight || 1 }} · {{ node.isBuiltIn ? '内置节点' : '可调度节点' }}</small>
            </div>

            <div class="node-chip-row">
              <span v-for="feature in enabledFeatureLabels(node)" :key="feature" class="node-chip">
                {{ feature }}
              </span>
              <span v-if="!enabledFeatureLabels(node).length" class="node-chip is-muted">无扩展能力</span>
            </div>

            <div class="node-flow">
              <span></span>
              <span></span>
              <span></span>
            </div>

            <div class="node-health-row">
              <span class="health-pill" :class="`is-${node.health.status}`">{{ healthStatusLabel(node.health.status) }}</span>
              <small>{{ node.health.message || '暂无状态说明' }}</small>
              <small>最后心跳 {{ formatDateTime(node.health.lastHeartbeatAt) }}</small>
              <small>最近检测 {{ formatDateTime(node.health.lastCheckedAt) }}</small>
            </div>
          </div>

          <button
            v-if="!node.isBuiltIn"
            class="node-delete-button"
            type="button"
            aria-label="删除节点"
            @click.stop="removeNode(node.id)"
          >
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="M6.5 6.5v8.5" />
              <path d="M10 6.5v8.5" />
              <path d="M13.5 6.5v8.5" />
              <path d="M4.5 5.5h11" />
              <path d="M7.75 5.5V4.2a1 1 0 0 1 1-1h2.5a1 1 0 0 1 1 1v1.3" />
              <path d="M6 5.5V16a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V5.5" />
            </svg>
          </button>
        </article>
      </section>

      <footer class="nodes-footer">
        <div class="pager">
          <button class="pager-arrow" type="button" :disabled="page === 1" @click="page = Math.max(1, page - 1)">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="m11.5 5.5-4.5 4.5 4.5 4.5" />
            </svg>
          </button>
          <span class="pager-index">{{ page }}</span>
          <button class="pager-arrow" type="button" :disabled="page >= totalPages" @click="page = Math.min(totalPages, page + 1)">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="m8.5 5.5 4.5 4.5-4.5 4.5" />
            </svg>
          </button>
        </div>

        <label class="page-size-select">
          <select v-model.number="pageSize">
            <option :value="11">每页 11 条</option>
            <option :value="20">每页 20 条</option>
            <option :value="50">每页 50 条</option>
          </select>
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="m5.5 7.5 4.5 5 4.5-5" />
          </svg>
        </label>
      </footer>
    </div>

    <div v-else class="editor-shell">
      <header class="editor-hero">
        <div class="editor-hero-copy">
          <p class="nodes-eyebrow">Node Editor</p>
          <h1>编辑节点 {{ activeNode.name }}</h1>
        </div>
        <span class="node-health-label">
          <span class="status-led" :class="healthLedClass(form.health.status)"></span>
          {{ healthStatusLabel(form.health.status) }}
        </span>
      </header>

      <div class="editor-surface">
        <section class="editor-section">
          <h2>基础信息</h2>

          <div class="health-summary-card" :class="`is-${activeNode.health.status}`">
            <div>
              <strong>{{ healthStatusLabel(activeNode.health.status) }}</strong>
              <p>{{ activeNode.health.message || '暂无巡检结果' }}</p>
              <small>最后心跳：{{ formatDateTime(activeNode.health.lastHeartbeatAt) }}</small>
              <small>最近巡检：{{ formatDateTime(activeNode.health.lastCheckedAt) }}</small>
            </div>
            <button class="ghost-button" type="button" :disabled="healthCheckingOne" @click="checkCurrentNodeHealth">
              {{ healthCheckingOne ? '检测中' : '检测节点状态' }}
            </button>
          </div>

          <label class="toggle-field">
            <button class="switch-button" :class="{ active: form.enabled }" type="button" @click="form.enabled = !form.enabled">
              <span></span>
            </button>
            <div>
              <strong>启用节点</strong>
              <small>启用后，节点会接受已开启能力对应的处理任务。</small>
            </div>
          </label>

          <label class="field-block">
            <span>名称</span>
            <input v-model="form.name" type="text" class="field-input" />
            <small>节点名称，也用于向管理员展示。</small>
          </label>

          <label class="field-block">
            <span>类型</span>
            <input :value="activeNode.typeLabel" type="text" class="field-input" disabled />
          </label>

          <label class="field-block">
            <span>负载均衡权重</span>
            <input v-model.number="form.weight" type="number" min="1" class="field-input" />
            <small>数值越高，节点在负载均衡中被选中的概率越大。</small>
          </label>
        </section>

        <section class="editor-section">
          <h2>已启用能力</h2>

          <label v-for="item in featureFields" :key="item.key" class="toggle-field feature-toggle">
            <button
              class="switch-button accent"
              :class="{ active: featureModel(item.key) }"
              type="button"
              @click="toggleFeature(item.key)"
            >
              <span></span>
            </button>
            <div>
              <strong>{{ item.label }}</strong>
              <small>{{ item.description }}</small>
              <span class="effect-state" :class="featureEffectState(item.key).tone">
                {{ featureEffectState(item.key).text }}
              </span>
            </div>
          </label>
        </section>

        <section v-if="form.features.offlineDownload" class="editor-section">
          <div class="section-heading">
            <h2>离线下载</h2>
            <button class="help-badge" type="button" aria-label="离线下载帮助">?</button>
          </div>

          <div class="effect-grid">
            <article v-for="item in offlineEffectStates" :key="item.label" class="effect-card" :class="item.tone">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
              <small>{{ item.detail }}</small>
            </article>
          </div>

          <label class="field-block">
            <span>下载器</span>
            <label class="select-shell">
              <select v-model="form.offline.downloader">
                <option value="Aria2">Aria2</option>
                <option value="qBittorrent">qBittorrent</option>
                <option value="Transmission">Transmission</option>
              </select>
              <svg viewBox="0 0 20 20" aria-hidden="true">
                <path d="m5.5 7.5 4.5 5 4.5-5" />
              </svg>
            </label>
            <small>在目标节点服务器上配置下载器，让离线任务被稳定接管。</small>
          </label>

          <label class="field-block">
            <span>RPC 服务地址</span>
            <input v-model="form.offline.rpcUrl" type="text" class="field-input" />
            <small>包含端口的完整 RPC 地址，例如 <code>http://127.0.0.1:6800/jsonrpc</code>。</small>
          </label>

          <label class="field-block">
            <span>RPC 授权令牌</span>
            <input v-model="form.offline.rpcSecret" type="text" class="field-input" />
            <small>与下载器配置中的密钥保持一致，未设置请留空。</small>
          </label>

          <label class="field-block">
            <span>下载器任务参数</span>
            <textarea v-model="form.offline.taskOptions" class="field-textarea" rows="7"></textarea>
            <small>以 JSON 键值对格式填写额外下载配置，用于按任务覆盖默认行为。</small>
          </label>

          <label class="field-block">
            <span>临时下载目录</span>
            <input v-model="form.offline.tempDir" type="text" class="field-input" />
            <small>节点用于存放离线下载文件的目录，需要确保 Cloudreve 进程具备读写权限。</small>
          </label>

          <label class="field-block">
            <span>状态刷新间隔（秒）</span>
            <input v-model.number="form.offline.refreshInterval" type="number" min="1" class="field-input" />
            <small>Cloudreve 向下载器请求任务状态的轮询间隔。</small>
          </label>

          <label class="toggle-field">
            <button
              class="switch-button muted"
              :class="{ active: form.offline.waitForSeeding }"
              type="button"
              @click="form.offline.waitForSeeding = !form.offline.waitForSeeding"
            >
              <span></span>
            </button>
            <div>
              <strong>等待做种完成</strong>
              <small>开启后，下载任务完成后仍可继续保留做种状态，直到满足下载器中的结束条件。</small>
            </div>
          </label>

          <div class="test-card">
            <div>
              <strong>下载器通信检测</strong>
              <p>{{ offlineConnectionHint }}</p>
              <dl v-if="offlineTestResult" class="test-result-list">
                <div>
                  <dt>下载器</dt>
                  <dd>{{ offlineTestResult.downloader }}</dd>
                </div>
                <div>
                  <dt>版本/结果</dt>
                  <dd>{{ offlineTestResult.version || '未知' }}</dd>
                </div>
                <div>
                  <dt>RPC 地址</dt>
                  <dd>{{ offlineTestResult.rpcUrl }}</dd>
                </div>
                <div>
                  <dt>测试时间</dt>
                  <dd>{{ formatDateTime(offlineTestResult.testedAt) }}</dd>
                </div>
                <div>
                  <dt>结果</dt>
                  <dd>{{ offlineTestResult.message }}</dd>
                </div>
              </dl>
            </div>
            <button class="ghost-button" type="button" :disabled="testingOffline" @click="testOfflineDownload">
              {{ testingOffline ? '测试中' : '测试下载器通信' }}
            </button>
          </div>
        </section>
      </div>

      <footer class="editor-actions">
        <button class="ghost-button" type="button" @click="goBackToList">返回列表</button>
        <button class="ghost-button" type="button" @click="resetForm">恢复当前配置</button>
        <button class="primary-button" type="button" :disabled="saving" @click="saveNode">
          {{ saving ? '保存中' : '保存节点' }}
        </button>
      </footer>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  checkNodeHealth,
  createNode as createAdminNode,
  deleteNode as deleteAdminNode,
  getNode,
  listNodes,
  testNodeOfflineConnectivity,
  type NodeOfflineConnectivityResult,
  type NodeDownloaderType,
  type NodePayload,
  updateNode as updateAdminNode,
} from '@/api/nodes';

type FeatureKey = 'createArchive' | 'extractArchive' | 'offlineDownload';

type NodeRecord = {
  id: number;
  name: string;
  type: 'master' | 'worker';
  typeLabel: string;
  enabled: boolean;
  weight: number;
  isBuiltIn: boolean;
  health: {
    status: string;
    message: string;
    lastHeartbeatAt: string | null;
    lastCheckedAt: string | null;
  };
  features: {
    createArchive: boolean;
    extractArchive: boolean;
    offlineDownload: boolean;
  };
  offline: {
    downloader: NodeDownloaderType;
    rpcUrl: string;
    rpcSecret: string;
    taskOptions: string;
    tempDir: string;
    refreshInterval: number;
    waitForSeeding: boolean;
  };
};

type NodeForm = Omit<NodeRecord, 'id' | 'typeLabel'>;

type OfflineTestResult = {
  success: boolean;
  message: string;
  downloader: NodeDownloaderType;
  rpcUrl: string;
  version: string;
  testedAt: string;
};

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const testingOffline = ref(false);
const healthCheckingAll = ref(false);
const healthCheckingOne = ref(false);
const nodes = ref<NodeRecord[]>([]);
const originalForm = ref<NodeForm | null>(null);
const offlineTestResult = ref<OfflineTestResult | null>(null);
const page = ref(1);
const pageSize = ref(11);

const form = reactive<NodeForm>({
  name: '',
  type: 'worker',
  enabled: true,
  weight: 1,
  isBuiltIn: false,
  health: {
    status: 'unknown',
    message: '',
    lastHeartbeatAt: null,
    lastCheckedAt: null,
  },
  features: {
    createArchive: false,
    extractArchive: false,
    offlineDownload: false,
  },
  offline: {
    downloader: 'Aria2',
    rpcUrl: 'http://127.0.0.1:6800/jsonrpc',
    rpcSecret: '',
    taskOptions: '{\n  "max-connection-per-server": "8"\n}',
    tempDir: '',
    refreshInterval: 5,
    waitForSeeding: false,
  },
});

const featureFields = [
  { key: 'createArchive', label: '创建压缩文件', description: '接受创建压缩文件的任务请求。' },
  { key: 'extractArchive', label: '解压缩', description: '接受解压文件的任务请求。' },
  { key: 'offlineDownload', label: '离线下载', description: '接受离线下载任务，启用后会显示下载器配置。' },
] as const;

const totalPages = computed(() => Math.max(1, Math.ceil(nodes.value.length / pageSize.value)));
const pagedNodes = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return nodes.value.slice(start, start + pageSize.value);
});

const activeNodeId = computed(() => {
  const raw = route.params.nodeId;
  if (!raw) return null;
  const value = Number(raw);
  return Number.isFinite(value) ? value : null;
});

const activeNode = computed(() => nodes.value.find((item) => item.id === activeNodeId.value) ?? null);

const metrics = computed(() => {
  const enabledCount = nodes.value.filter((item) => item.enabled).length;
  const onlineCount = nodes.value.filter((item) => item.health.status === 'online').length;
  const offlineCount = nodes.value.filter((item) => item.features.offlineDownload).length;
  const healthyCount = nodes.value.filter((item) => item.health.status === 'online' || item.health.status === 'idle').length;
  return [
    { label: '健康节点', value: `${healthyCount}/${nodes.value.length || 1}`, detail: '可参与策略调度的节点数量' },
    { label: '离线下载能力', value: `${offlineCount}`, detail: '已开启下载器联动的节点' },
    { label: '最高权重', value: `${Math.max(1, ...nodes.value.map((item) => item.weight || 1))}`, detail: '当前负载均衡权重峰值' },
    { label: '在线节点', value: `${onlineCount}/${enabledCount}`, detail: '最近检测真实在线的服务实例' },
  ];
});

const offlineConnectionHint = computed(() => {
  if (!form.features.offlineDownload) {
    return '当前节点尚未启用离线下载能力。';
  }
  return `将尝试以 ${form.offline.downloader} 对接 ${form.offline.rpcUrl || '未配置 RPC 地址'}。`;
});

const supportsRealOfflineSubmit = computed(() => form.offline.downloader === 'Aria2');

const participatesInOfflineDispatch = computed(() => {
  return form.enabled && form.features.offlineDownload && supportsRealOfflineSubmit.value;
});

const offlineEffectStates = computed(() => [
  {
    label: '调度参与',
    value: participatesInOfflineDispatch.value ? '会参与' : '不会参与',
    detail: offlineDispatchReason(),
    tone: participatesInOfflineDispatch.value ? 'is-ok' : 'is-warn',
  },
  {
    label: '真实提交',
    value: supportsRealOfflineSubmit.value ? '已接入' : '未接入',
    detail: supportsRealOfflineSubmit.value
      ? '当前下载器会通过 aria2.addUri 提交真实离线任务。'
      : `${form.offline.downloader} 目前只支持通信测试，真实任务提交尚未接入。`,
    tone: supportsRealOfflineSubmit.value ? 'is-ok' : 'is-danger',
  },
  {
    label: '最近通信',
    value: offlineTestResult.value ? (offlineTestResult.value.success ? '成功' : '失败') : '未验证',
    detail: offlineTestResult.value
      ? `${offlineTestResult.value.downloader} ${offlineTestResult.value.version || 'unknown'} @ ${formatDateTime(offlineTestResult.value.testedAt)}`
      : '尚未在本页面完成下载器通信测试。',
    tone: offlineTestResult.value ? (offlineTestResult.value.success ? 'is-ok' : 'is-danger') : 'is-warn',
  },
]);

watch(pageSize, () => {
  page.value = 1;
});

watch(totalPages, (value) => {
  if (page.value > value) {
    page.value = value;
  }
});

watch(
  activeNode,
  (node) => {
    if (node) {
      syncFormFromNode(node);
      loadOfflineTestResult(node.id);
    } else {
      offlineTestResult.value = null;
    }
  },
  { immediate: true },
);

const cloneFeatures = (features: NodeRecord['features']): NodeRecord['features'] => ({
  createArchive: features.createArchive,
  extractArchive: features.extractArchive,
  offlineDownload: features.offlineDownload,
});

const cloneOffline = (offline: NodeRecord['offline']): NodeRecord['offline'] => ({
  downloader: offline.downloader,
  rpcUrl: offline.rpcUrl,
  rpcSecret: offline.rpcSecret,
  taskOptions: offline.taskOptions,
  tempDir: offline.tempDir,
  refreshInterval: offline.refreshInterval,
  waitForSeeding: offline.waitForSeeding,
});

const cloneNodeForm = (node: NodeRecord): NodeForm => ({
  name: node.name,
  type: node.type,
  enabled: node.enabled,
  weight: node.weight,
  isBuiltIn: node.isBuiltIn,
  health: {
    status: node.health.status,
    message: node.health.message,
    lastHeartbeatAt: node.health.lastHeartbeatAt,
    lastCheckedAt: node.health.lastCheckedAt,
  },
  features: cloneFeatures(node.features),
  offline: cloneOffline(node.offline),
});

const mapNodeFromPayload = (payload: NodePayload): NodeRecord => ({
  id: payload.id || 0,
  name: payload.name,
  type: payload.type,
  typeLabel: payload.type === 'master' ? '主机' : '从机',
  enabled: payload.enabled,
  weight: payload.weight,
  isBuiltIn: payload.is_built_in,
  health: {
    status: payload.health?.status || 'unknown',
    message: payload.health?.message || '',
    lastHeartbeatAt: payload.health?.last_heartbeat_at || null,
    lastCheckedAt: payload.health?.last_checked_at || null,
  },
  features: {
    createArchive: payload.features.create_archive,
    extractArchive: payload.features.extract_archive,
    offlineDownload: payload.features.offline_download,
  },
  offline: {
    downloader: payload.offline.downloader,
    rpcUrl: payload.offline.rpc_url,
    rpcSecret: payload.offline.rpc_secret,
    taskOptions: payload.offline.task_options,
    tempDir: payload.offline.temp_dir,
    refreshInterval: payload.offline.refresh_interval,
    waitForSeeding: payload.offline.wait_for_seeding,
  },
});

const mapFormToPayload = (state: NodeForm): NodePayload => ({
  name: state.name.trim(),
  type: state.type,
  enabled: state.enabled,
  weight: Number(state.weight) || 1,
  is_built_in: state.isBuiltIn,
  health: {
    status: state.health.status,
    message: state.health.message,
    last_heartbeat_at: state.health.lastHeartbeatAt,
    last_checked_at: state.health.lastCheckedAt,
  },
  features: {
    create_archive: state.features.createArchive,
    extract_archive: state.features.extractArchive,
    offline_download: state.features.offlineDownload,
  },
  offline: {
    downloader: state.offline.downloader,
    rpc_url: state.offline.rpcUrl.trim(),
    rpc_secret: state.offline.rpcSecret.trim(),
    task_options: state.offline.taskOptions,
    temp_dir: state.offline.tempDir.trim(),
    refresh_interval: Number(state.offline.refreshInterval) || 5,
    wait_for_seeding: state.offline.waitForSeeding,
  },
});

const syncFormFromNode = (node: NodeRecord | null) => {
  if (!node) return;
  Object.assign(form, cloneNodeForm(node));
  originalForm.value = cloneNodeForm(node);
};

const featureModel = (key: FeatureKey) => form.features[key];

const toggleFeature = (key: FeatureKey) => {
  form.features[key] = !form.features[key];
};

const featureEffectState = (key: FeatureKey) => {
  if (!form.enabled) {
    return { tone: 'is-warn', text: '节点已停用，保存后不会参与任何调度。' };
  }
  if (!form.features[key]) {
    return { tone: 'is-warn', text: '能力已关闭，保存后不会参与该能力调度。' };
  }
  if (key === 'createArchive') {
    return form.type === 'master'
      ? { tone: 'is-ok', text: '已接入真实任务：压缩包下载会走本机 ZIP 创建流程。' }
      : { tone: 'is-danger', text: '尚未接入真实任务：远端节点压缩执行暂不支持。' };
  }
  if (key === 'extractArchive') {
    return form.type === 'master'
      ? { tone: 'is-ok', text: '已接入真实任务：ZIP 解压会走本机解压流程。' }
      : { tone: 'is-danger', text: '尚未接入真实任务：远端节点解压执行暂不支持。' };
  }
  return supportsRealOfflineSubmit.value
    ? { tone: 'is-ok', text: '已接入真实任务：Aria2 会接收离线下载提交。' }
    : { tone: 'is-danger', text: `${form.offline.downloader} 尚未接入真实任务提交。` };
};

const offlineDispatchReason = () => {
  if (!form.enabled) return '节点已停用，离线下载调度会跳过它。';
  if (!form.features.offlineDownload) return '离线下载能力已关闭，调度会跳过它。';
  if (!supportsRealOfflineSubmit.value) return `${form.offline.downloader} 尚未支持真实提交任务，调度前会被拦截。`;
  if (!form.offline.rpcUrl.trim()) return 'RPC 地址为空，真实提交会失败。';
  return '节点启用、能力开启且下载器支持真实提交。';
};

const validateTaskOptions = () => {
  const raw = form.offline.taskOptions.trim();
  if (!raw) {
    form.offline.taskOptions = '{}';
    return true;
  }
  try {
    const parsed = JSON.parse(raw);
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      ElMessage.error('下载器任务参数必须是 JSON 对象，例如 {"max-connection-per-server":"8"}');
      return false;
    }
    form.offline.taskOptions = JSON.stringify(parsed, null, 2);
    return true;
  } catch (error) {
    const message = error instanceof Error ? error.message : 'JSON 格式错误';
    ElMessage.error(`下载器任务参数不是合法 JSON：${message}`);
    return false;
  }
};

const offlineTestStorageKey = (nodeID: number) => `xingyunpan.node.${nodeID}.offlineConnectivity`;

const mapOfflineTestResult = (result: NodeOfflineConnectivityResult): OfflineTestResult => ({
  success: result.success,
  message: result.message,
  downloader: result.downloader || form.offline.downloader,
  rpcUrl: result.rpc_url || form.offline.rpcUrl.trim(),
  version: result.version || 'unknown',
  testedAt: result.tested_at || new Date().toISOString(),
});

const loadOfflineTestResult = (nodeID: number) => {
  try {
    const raw = window.localStorage.getItem(offlineTestStorageKey(nodeID));
    offlineTestResult.value = raw ? (JSON.parse(raw) as OfflineTestResult) : null;
  } catch {
    offlineTestResult.value = null;
  }
};

const persistOfflineTestResult = (nodeID: number, result: OfflineTestResult) => {
  offlineTestResult.value = result;
  window.localStorage.setItem(offlineTestStorageKey(nodeID), JSON.stringify(result));
};

const updateNodeInList = (node: NodeRecord) => {
  const index = nodes.value.findIndex((item) => item.id === node.id);
  if (index >= 0) {
    nodes.value.splice(index, 1, node);
  }
};

const enabledFeatureLabels = (node: NodeRecord) => {
  const labels: string[] = [];
  if (node.features.createArchive) labels.push('创建压缩文件');
  if (node.features.extractArchive) labels.push('解压缩');
  if (node.features.offlineDownload) labels.push('离线下载');
  return labels;
};

const healthStatusLabel = (status: string) => {
  switch (status) {
    case 'online':
      return '在线';
    case 'offline':
      return '离线';
    case 'idle':
      return '待机';
    case 'disabled':
      return '停用';
    default:
      return '未知';
  }
};

const healthLedClass = (status: string) => {
  switch (status) {
    case 'online':
      return 'is-online';
    case 'offline':
      return 'is-offline';
    default:
      return 'is-idle';
  }
};

const formatDateTime = (value: string | null) => {
  if (!value) return '暂无';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '暂无';
  return date.toLocaleString('zh-CN', { hour12: false });
};

async function fetchNodes(showSuccess = false) {
  loading.value = true;
  try {
    const data = await listNodes();
    nodes.value = data.map(mapNodeFromPayload);
    if (showSuccess) {
      ElMessage.success('节点列表已刷新');
    }
    if (activeNodeId.value && !nodes.value.some((item) => item.id === activeNodeId.value)) {
      router.push('/admin/nodes');
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点列表加载失败');
  } finally {
    loading.value = false;
  }
}

async function refreshNodes() {
  await fetchNodes(true);
}

async function runHealthChecks() {
  if (!nodes.value.length) return;
  healthCheckingAll.value = true;
  try {
    let failed = 0;
    const results = await Promise.all(
      nodes.value.map(async (item) => {
        try {
          const checked = await checkNodeHealth(item.id);
          return mapNodeFromPayload(checked);
        } catch {
          failed += 1;
          return item;
        }
      }),
    );
    nodes.value = results;
    const summary = results.map((item) => `${item.name}: ${healthStatusLabel(item.health.status)}`).join('，');
    ElMessage.success(failed > 0 ? `节点巡检完成，${failed} 个接口请求失败；${summary}` : `节点巡检完成：${summary}`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点巡检失败');
  } finally {
    healthCheckingAll.value = false;
  }
}

async function createNode() {
  try {
    const nextIndex = nodes.value.length + 1;
    const created = await createAdminNode({
      name: `节点 ${nextIndex}`,
      type: 'worker',
      enabled: true,
      weight: 1,
      is_built_in: false,
      health: {
        status: 'unknown',
        message: '',
        last_heartbeat_at: null,
        last_checked_at: null,
      },
      features: {
        create_archive: true,
        extract_archive: false,
        offline_download: false,
      },
      offline: {
        downloader: 'Aria2',
        rpc_url: 'http://127.0.0.1:6800/jsonrpc',
        rpc_secret: '',
        task_options: '{\n  "max-connection-per-server": "8"\n}',
        temp_dir: '',
        refresh_interval: 5,
        wait_for_seeding: false,
      },
    });
    const mapped = mapNodeFromPayload(created);
    nodes.value.unshift(mapped);
    ElMessage.success('节点已创建');
    router.push(`/admin/nodes/${mapped.id}`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点创建失败');
  }
}

const openNode = (id: number) => {
  router.push(`/admin/nodes/${id}`);
};

async function removeNode(id: number) {
  try {
    await deleteAdminNode(id);
    nodes.value = nodes.value.filter((item) => item.id !== id);
    ElMessage.success('节点已删除');
    if (activeNodeId.value === id) {
      router.push('/admin/nodes');
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点删除失败');
  }
}

const goBackToList = () => {
  router.push('/admin/nodes');
};

function resetForm() {
  if (!originalForm.value) return;
  Object.assign(form, JSON.parse(JSON.stringify(originalForm.value)));
  ElMessage.success('已恢复当前配置');
}

async function saveNode() {
  if (!activeNode.value?.id) return;
  if (!validateTaskOptions()) return;
  const nodeID = activeNode.value.id;
  const payload = mapFormToPayload(form);
  saving.value = true;
  try {
    await updateAdminNode(nodeID, payload);
    const confirmed = await getNode(nodeID);
    const mapped = mapNodeFromPayload(confirmed);
    updateNodeInList(mapped);
    syncFormFromNode(mapped);
    ElMessage.success('节点已保存，并已从后端重新确认配置。');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点保存失败');
  } finally {
    saving.value = false;
  }
}

async function checkCurrentNodeHealth() {
  if (!activeNode.value?.id) return;
  healthCheckingOne.value = true;
  try {
    const checked = await checkNodeHealth(activeNode.value.id);
    const mapped = mapNodeFromPayload(checked);
    updateNodeInList(mapped);
    syncFormFromNode(mapped);
    ElMessage.success(`节点状态：${healthStatusLabel(mapped.health.status)}；${mapped.health.message || '暂无检测详情'}`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '节点状态检测失败');
  } finally {
    healthCheckingOne.value = false;
  }
}

async function testOfflineDownload() {
  if (!validateTaskOptions()) return;
  testingOffline.value = true;
  try {
    form.offline.taskOptions = form.offline.taskOptions.trim() || '{}';
    const result = await testNodeOfflineConnectivity({
      downloader: form.offline.downloader,
      rpc_url: form.offline.rpcUrl.trim(),
      rpc_secret: form.offline.rpcSecret.trim(),
      task_options: form.offline.taskOptions,
      temp_dir: form.offline.tempDir.trim(),
      refresh_interval: Number(form.offline.refreshInterval) || 5,
      wait_for_seeding: form.offline.waitForSeeding,
    });
    const mapped = mapOfflineTestResult(result);
    if (activeNode.value?.id) {
      persistOfflineTestResult(activeNode.value.id, mapped);
    } else {
      offlineTestResult.value = mapped;
    }
    ElMessage.success(result.message);
  } catch (error) {
    const message = error instanceof Error ? error.message : '下载器通信测试失败';
    const failedResult: OfflineTestResult = {
      success: false,
      message,
      downloader: form.offline.downloader,
      rpcUrl: form.offline.rpcUrl.trim(),
      version: 'unavailable',
      testedAt: new Date().toISOString(),
    };
    if (activeNode.value?.id) {
      persistOfflineTestResult(activeNode.value.id, failedResult);
    } else {
      offlineTestResult.value = failedResult;
    }
    ElMessage.error(message);
  } finally {
    testingOffline.value = false;
  }
}

onMounted(() => {
  fetchNodes();
});
</script>

<style scoped>
.nodes-page {
  min-height: calc(100vh - 96px);
  color: #172033;
}

.nodes-shell,
.editor-shell {
  display: grid;
  gap: 22px;
  padding: 30px;
  border: 1px solid rgba(226, 235, 246, 0.88);
  border-radius: 30px;
  background:
    radial-gradient(circle at 8% 4%, rgba(186, 230, 253, 0.56), transparent 28%),
    radial-gradient(circle at 96% 16%, rgba(251, 207, 232, 0.4), transparent 30%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.86), rgba(248, 252, 255, 0.76));
  box-shadow: 0 26px 70px rgba(52, 75, 118, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(22px);
}

.nodes-hero,
.editor-hero,
.metric-card,
.create-node-card,
.node-card,
.editor-section {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.58));
  box-shadow: 0 20px 46px rgba(87, 109, 143, 0.12), inset 0 1px 18px rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(18px);
}

.nodes-hero,
.editor-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 28px;
  border-radius: 26px;
}

.nodes-eyebrow {
  margin: 0 0 10px;
  color: #2f80d8;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.nodes-hero h1,
.editor-hero h1 {
  margin: 0;
  color: #172033;
  font-size: 52px;
  line-height: 1.02;
  font-weight: 900;
  letter-spacing: -0.04em;
}

.nodes-subtitle {
  max-width: 760px;
  margin: 14px 0 0;
  color: #66768f;
  font-size: 15px;
  line-height: 1.85;
}

.hero-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: 1px solid rgba(210, 222, 236, 0.95);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.68);
  color: #728094;
  font-size: 20px;
  font-weight: 900;
  cursor: pointer;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  gap: 8px;
  padding: 18px 20px;
  border-radius: 20px;
}

.metric-label,
.metric-card small {
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.metric-label {
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.metric-card strong {
  color: #172033;
  font-size: 28px;
}

.nodes-toolbar,
.nodes-footer,
.editor-actions,
.section-heading,
.test-card {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-button,
.ghost-button,
.primary-button,
.page-size-select,
.select-shell {
  display: inline-flex;
  align-items: center;
  min-height: 48px;
  border: 1px solid rgba(218, 228, 240, 0.9);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  color: #263241;
  font-size: 15px;
  font-weight: 800;
  box-shadow: 0 12px 26px rgba(84, 105, 138, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(16px);
}

.toolbar-button,
.ghost-button,
.primary-button {
  justify-content: center;
  gap: 10px;
  padding: 0 16px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.toolbar-button:hover,
.ghost-button:hover,
.primary-button:hover,
.page-size-select:hover,
.select-shell:hover,
.node-card:hover,
.create-node-card:hover {
  transform: translateY(-1px);
}

.toolbar-button:disabled,
.ghost-button:disabled,
.primary-button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.toolbar-button svg,
.node-delete-button svg,
.pager-arrow svg,
.page-size-select svg,
.select-shell svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.node-board {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

.create-node-card,
.node-card {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
}

.create-node-card {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 112px;
  gap: 14px;
  border-style: dashed;
  color: #617086;
  font-size: 20px;
  font-weight: 800;
  cursor: pointer;
}

.create-node-plus {
  color: #4f8fe9;
  font-size: 40px;
  font-weight: 300;
  line-height: 1;
}

.node-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 16px;
  padding: 22px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.node-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 90% 16%, rgba(251, 207, 232, 0.34), transparent 28%),
    radial-gradient(circle at 18% 100%, rgba(125, 211, 252, 0.28), transparent 30%);
  opacity: 0;
  transition: opacity 0.22s ease;
  pointer-events: none;
}

.node-card:hover {
  border-color: rgba(251, 207, 232, 0.9);
  box-shadow: 0 24px 54px rgba(232, 121, 179, 0.13), 0 14px 34px rgba(96, 165, 250, 0.12);
}

.node-card:hover::before {
  opacity: 1;
}

.node-card.is-disabled {
  opacity: 0.78;
}

.node-card-main {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 14px;
  min-width: 0;
}

.node-card-header,
.node-name-row,
.node-health-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.node-health-label {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-height: 30px;
  color: #52705f;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.status-led {
  position: relative;
  width: 10px;
  height: 10px;
  border-radius: 999px;
}

.status-led::after {
  content: '';
  position: absolute;
  inset: -7px;
  border-radius: inherit;
  opacity: 0.72;
  animation: ledPulse 1.9s ease-in-out infinite;
}

.status-led.is-online {
  background: #26d67f;
  box-shadow: 0 0 14px rgba(38, 214, 127, 0.92);
}

.status-led.is-online::after {
  background: radial-gradient(circle, rgba(38, 214, 127, 0.34), transparent 68%);
}

.status-led.is-offline {
  background: #ef4444;
  box-shadow: 0 0 14px rgba(239, 68, 68, 0.72);
}

.status-led.is-offline::after {
  background: radial-gradient(circle, rgba(239, 68, 68, 0.3), transparent 68%);
}

.status-led.is-idle {
  background: #a7b7c8;
  box-shadow: 0 0 12px rgba(148, 163, 184, 0.45);
}

.status-led.is-idle::after {
  background: radial-gradient(circle, rgba(148, 163, 184, 0.22), transparent 68%);
}

@keyframes ledPulse {
  0%,
  100% {
    transform: scale(0.75);
    opacity: 0.45;
  }

  50% {
    transform: scale(1.18);
    opacity: 0.9;
  }
}

.node-role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
}

.node-role-badge svg {
  width: 14px;
  height: 14px;
  fill: currentColor;
}

.node-role-badge.is-master {
  background: rgba(147, 197, 253, 0.24);
  color: #2563eb;
}

.node-role-badge.is-worker {
  background: rgba(45, 212, 191, 0.2);
  color: #0f766e;
}

.node-name-row {
  align-items: flex-end;
}

.node-name-row strong {
  color: #172033;
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.node-name-row small,
.node-health-row small {
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.node-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.node-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 11px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.66);
  color: #38475c;
  font-size: 12px;
  font-weight: 800;
}

.node-chip.is-muted {
  color: #8290a4;
}

.node-flow {
  display: grid;
  grid-template-columns: 1fr 0.72fr 0.42fr;
  gap: 8px;
  max-width: 420px;
}

.node-flow span {
  height: 9px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(147, 197, 253, 0.68), rgba(251, 207, 232, 0.74));
  box-shadow: 0 8px 18px rgba(147, 197, 253, 0.18);
}

.node-health-row {
  justify-content: flex-start;
  flex-wrap: wrap;
  padding-top: 14px;
  border-top: 1px solid rgba(226, 235, 246, 0.74);
}

.health-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: fit-content;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
}

.health-pill.is-online {
  background: rgba(34, 197, 94, 0.14);
  color: #15803d;
}

.health-pill.is-offline {
  background: rgba(239, 68, 68, 0.12);
  color: #b91c1c;
}

.health-pill.is-idle {
  background: rgba(59, 130, 246, 0.13);
  color: #2563eb;
}

.health-pill.is-disabled,
.health-pill.is-unknown {
  background: rgba(148, 163, 184, 0.16);
  color: #475569;
}

.node-delete-button {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #7b8794;
  cursor: pointer;
}

.nodes-footer {
  justify-content: space-between;
}

.pager {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pager-arrow,
.pager-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 999px;
}

.pager-arrow {
  border: 1px solid rgba(226, 235, 246, 0.9);
  background: rgba(255, 255, 255, 0.68);
  color: #74849b;
  cursor: pointer;
}

.pager-arrow:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.pager-index {
  background: rgba(255, 255, 255, 0.7);
  color: #172033;
  font-weight: 900;
}

.page-size-select,
.select-shell {
  position: relative;
  padding-right: 42px;
}

.page-size-select select,
.select-shell select {
  min-height: 48px;
  width: 100%;
  padding: 0 16px;
  border: none;
  background: transparent;
  color: #172033;
  font: inherit;
  appearance: none;
  outline: none;
}

.page-size-select svg,
.select-shell svg {
  position: absolute;
  right: 16px;
  pointer-events: none;
}

.editor-surface {
  display: grid;
  gap: 22px;
}

.editor-section {
  display: grid;
  gap: 22px;
  padding: 26px;
  border-radius: 24px;
}

.editor-section h2 {
  margin: 0;
  color: #172033;
  font-size: 28px;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.health-summary-card,
.test-card {
  padding: 18px 20px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.54);
}

.effect-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.effect-card {
  display: grid;
  gap: 7px;
  padding: 14px 16px;
  border: 1px solid rgba(218, 228, 240, 0.82);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.58);
}

.effect-card span,
.effect-card small {
  color: #74849b;
  font-size: 12px;
  font-weight: 800;
}

.effect-card strong {
  color: #172033;
  font-size: 18px;
  font-weight: 900;
}

.effect-card.is-ok,
.effect-state.is-ok {
  border-color: rgba(34, 197, 94, 0.24);
  color: #15803d;
}

.effect-card.is-warn,
.effect-state.is-warn {
  border-color: rgba(245, 158, 11, 0.28);
  color: #b45309;
}

.effect-card.is-danger,
.effect-state.is-danger {
  border-color: rgba(239, 68, 68, 0.24);
  color: #b91c1c;
}

.effect-state {
  display: block;
  width: fit-content;
  max-width: 100%;
  margin-top: 8px;
  padding: 6px 9px;
  border: 1px solid currentColor;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.54);
  font-size: 12px;
  font-weight: 800;
  line-height: 1.5;
}

.health-summary-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.health-summary-card strong,
.test-card strong {
  display: block;
  margin-bottom: 6px;
  color: #172033;
  font-size: 16px;
  font-weight: 900;
}

.health-summary-card p,
.health-summary-card small,
.test-card p {
  margin: 0;
  color: #6f7f94;
  font-size: 13px;
  line-height: 1.75;
}

.health-summary-card small {
  display: block;
}

.health-summary-card.is-online {
  border-color: rgba(34, 197, 94, 0.24);
}

.health-summary-card.is-offline {
  border-color: rgba(239, 68, 68, 0.22);
}

.test-result-list {
  display: grid;
  gap: 8px;
  margin: 12px 0 0;
}

.test-result-list div {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  gap: 10px;
}

.test-result-list dt,
.test-result-list dd {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
}

.test-result-list dt {
  color: #74849b;
  font-weight: 900;
}

.test-result-list dd {
  min-width: 0;
  color: #263241;
  font-weight: 800;
  overflow-wrap: anywhere;
}

.field-block {
  display: grid;
  gap: 10px;
}

.field-block > span,
.toggle-field strong {
  color: #223044;
  font-size: 16px;
  font-weight: 900;
}

.field-block small,
.toggle-field small {
  color: #6f7f94;
  font-size: 13px;
  line-height: 1.75;
}

.field-input,
.field-textarea {
  width: 100%;
  min-height: 54px;
  padding: 0 18px;
  border: 1px solid rgba(210, 222, 236, 0.95);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.68);
  color: #172033;
  font-size: 16px;
  outline: none;
  box-sizing: border-box;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.field-textarea {
  min-height: 180px;
  padding: 16px 18px;
  resize: vertical;
}

.field-input:focus,
.field-textarea:focus,
.select-shell:focus-within {
  border-color: #93c5fd;
  box-shadow: 0 0 0 4px rgba(147, 197, 253, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.field-input[disabled] {
  color: #8290a4;
}

.toggle-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 14px;
  align-items: start;
}

.feature-toggle {
  padding: 4px 0 18px;
  border-bottom: 1px solid rgba(226, 235, 246, 0.74);
}

.feature-toggle:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.switch-button {
  position: relative;
  width: 60px;
  height: 34px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.84);
  border-radius: 999px;
  background: linear-gradient(145deg, rgba(219, 229, 240, 0.92), rgba(245, 248, 252, 0.82));
  box-shadow: inset 0 2px 8px rgba(117, 135, 160, 0.22), 0 10px 18px rgba(80, 103, 138, 0.1);
  cursor: pointer;
  transition: background 0.2s ease, box-shadow 0.2s ease;
}

.switch-button span {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: linear-gradient(145deg, #fff, #eef7ff);
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.2), inset 0 1px 0 #fff;
  transition: transform 0.2s ease;
}

.switch-button.active {
  background: linear-gradient(135deg, rgba(125, 211, 252, 0.9), rgba(251, 207, 232, 0.9));
  box-shadow: 0 0 0 4px rgba(251, 207, 232, 0.16), inset 0 2px 8px rgba(59, 130, 246, 0.16);
}

.switch-button.accent.active {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.78), rgba(56, 189, 248, 0.84));
}

.switch-button.muted.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.72), rgba(20, 184, 166, 0.72));
}

.switch-button.active span {
  transform: translateX(26px);
}

.help-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(210, 222, 236, 0.95);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.68);
  color: #728094;
  font-size: 18px;
  font-weight: 900;
}

.test-card {
  justify-content: space-between;
}

.editor-actions {
  justify-content: flex-end;
}

.primary-button {
  border: 0;
  background: linear-gradient(135deg, #38bdf8, #f9a8d4);
  color: #fff;
  box-shadow: 0 18px 34px rgba(56, 189, 248, 0.2), 0 14px 28px rgba(249, 168, 212, 0.18);
}

code {
  padding: 2px 6px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.72);
  color: #2563eb;
  font-size: 12px;
  font-family: Consolas, 'Liberation Mono', Menlo, monospace;
}

@media (max-width: 1180px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .nodes-shell,
  .editor-shell {
    padding: 20px;
    border-radius: 24px;
  }

  .nodes-hero,
  .editor-hero,
  .editor-section,
  .node-card {
    padding: 22px 18px;
    border-radius: 22px;
  }

  .nodes-hero,
  .editor-hero,
  .nodes-footer,
  .editor-actions,
  .test-card,
  .node-card-meta,
  .node-card-header,
  .node-name-row,
  .section-heading,
  .health-summary-card {
    flex-direction: column;
    align-items: stretch;
  }

  .nodes-hero h1,
  .editor-hero h1 {
    font-size: 36px;
  }

  .metric-grid {
    grid-template-columns: 1fr;
  }

  .effect-grid {
    grid-template-columns: 1fr;
  }

  .node-card {
    grid-template-columns: 1fr;
  }

  .toggle-field {
    grid-template-columns: 1fr;
  }
}
</style>
