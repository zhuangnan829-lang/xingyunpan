<template>
  <section class="storage-policy-page">
    <input ref="importFileInputRef" class="hidden-file-input" type="file" accept="application/json,.json" @change="handleImportFileChange" />
    <div v-if="migrationDialog.visible" class="migration-backdrop">
      <section class="migration-dialog">
        <header>
          <div>
            <p class="section-eyebrow">用户组迁移</p>
            <h2>先迁移绑定用户组</h2>
          </div>
          <button class="migration-close" type="button" aria-label="关闭" @click="closeMigrationDialog">×</button>
        </header>
        <p class="migration-summary">
          “{{ resolvePolicyName(migrationDialog.sourcePolicy?.name || '') }}”仍影响 {{ migrationDialog.userCount }} 位用户，需要迁移后才能删除。
        </p>
        <div class="migration-group-list">
          <article v-for="group in migrationDialog.groups" :key="group.id">
            <strong>{{ group.name }}</strong>
            <span>{{ group.user_count }} 位用户 · {{ group.file_count }} 个文件</span>
          </article>
        </div>
        <label class="field-block">
          <span>迁移到</span>
          <label class="unit-select migration-select">
            <select v-model.number="migrationDialog.targetPolicyId">
              <option v-for="policy in migrationTargetPolicies" :key="policy.id" :value="policy.id">{{ resolvePolicyName(policy.name) }}</option>
            </select>
            <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg>
          </label>
        </label>
        <footer>
          <button class="ghost-button" type="button" :disabled="migrationDialog.loading" @click="closeMigrationDialog">取消</button>
          <button class="primary-button" type="button" :disabled="migrationDialog.loading || !migrationDialog.targetPolicyId" @click="migrateAndDeletePolicy">
            {{ migrationDialog.loading ? '迁移中' : '迁移并删除' }}
          </button>
        </footer>
      </section>
    </div>

    <div v-if="!activePolicy" class="policy-shell">
      <header class="policy-hero">
        <div class="policy-hero-copy">
          <p class="section-eyebrow">存储策略</p>
          <h1>存储策略</h1>
          <p>按用户组绑定策略，统一管理上传、分片、CDN 和加密配置。</p>
        </div>
        <div class="hero-orb" aria-hidden="true">
          <span class="led led-online"></span>
          <strong>{{ policies.length || 1 }}</strong>
          <small>策略数量</small>
        </div>
      </header>

      <section class="policy-toolbar">
        <button class="toolbar-button" type="button" :disabled="loading" @click="reloadPolicies">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M16.2 9.2A6.2 6.2 0 1 0 15 13.8" />
            <path d="M16.2 4.8v4.7h-4.7" />
          </svg>
          <span>{{ loading ? '刷新中' : '刷新' }}</span>
        </button>
        <button class="toolbar-button" type="button" :disabled="loading || saving" @click="triggerImport()">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M10 3v10" /><path d="m6.5 6.5 3.5-3.5 3.5 3.5" /><path d="M4 13v2.5A1.5 1.5 0 0 0 5.5 17h9A1.5 1.5 0 0 0 16 15.5V13" />
          </svg>
          <span>导入策略</span>
        </button>

        <label class="filter-select">
          <select v-model="selectedFilter">
            <option v-for="item in policyFilters" :key="item.value" :value="item.value">{{ item.label }}</option>
          </select>
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg>
        </label>
      </section>

      <section class="policy-overview">
        <article v-for="item in policyInsights" :key="item.label" class="overview-card">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <small>{{ item.detail }}</small>
        </article>
      </section>

      <section class="policy-layout">
        <div class="policy-board">
          <button class="add-policy-card" type="button" @click="createPolicy">
            <span class="add-policy-icon">+</span>
            <span>添加存储策略</span>
          </button>

          <article v-for="policy in pagedPolicies" :key="policy.id" class="policy-card" @click="openPolicy(policy.id)">
            <div class="policy-card-header">
              <span class="health-line"><span class="led led-online"></span> 生效中</span>
              <span class="type-chip">{{ resolveTypeLabel(policy.type) }}</span>
            </div>
            <div class="policy-card-body">
              <div class="policy-card-copy">
                <strong>{{ resolvePolicyName(policy.name) }}</strong>
                <p>{{ resolvePolicySummary(policy) }}</p>
                <div class="policy-chip-row">
                  <span v-for="group in policy.groups" :key="group" class="policy-chip">{{ resolveGroupLabel(group) }}</span>
                </div>
              </div>
              <div class="policy-card-art" :class="`is-${policy.type}`">
                <div class="storage-cube"><span></span><span></span><span></span></div>
                <small>{{ resolveNodeName(policy.type) }}</small>
              </div>
            </div>
            <div class="policy-card-footer">
              <div class="policy-card-footer-actions">
                <button class="policy-stats-link" type="button" @click.stop="openPolicy(policy.id)">查看配置</button>
                <button class="policy-stats-link" type="button" @click.stop="duplicatePolicy(policy)">复制</button>
                <button class="policy-stats-link" type="button" @click.stop="downloadPolicyJson(policy)">导出</button>
              </div>
              <span>{{ policy.chunkSize }}{{ policy.chunkSizeUnit }} 分片 · {{ policy.parallelChunkCount }} 路并行</span>
            </div>
            <button v-if="canDeletePolicy(policy)" class="policy-delete-button" type="button" title="删除" aria-label="删除" @click.stop="removePolicy(policy)">
              <svg viewBox="0 0 20 20" aria-hidden="true">
                <path d="M6.5 6.5v9" /><path d="M10 6.5v9" /><path d="M13.5 6.5v9" /><path d="M4.5 5.5h11" />
                <path d="M7.5 5.5V4a1 1 0 0 1 1-1h3a1 1 0 0 1 1 1v1.5" /><path d="M6 5.5V16a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V5.5" />
              </svg>
            </button>
          </article>
        </div>

        <aside class="node-panel">
          <div class="panel-title"><span class="section-eyebrow">存储节点</span><strong>节点健康矩阵</strong></div>
          <div class="node-list">
            <article v-for="node in storageNodes" :key="node.name" class="node-tile">
              <span class="led" :class="node.online ? 'led-online' : 'led-idle'"></span>
              <div><strong>{{ node.name }}</strong><small>{{ node.detail }}</small></div>
              <span>{{ node.latency }}</span>
            </article>
          </div>
        </aside>
      </section>

      <footer class="policy-footer">
        <div class="pager">
          <button class="pager-arrow" type="button" :disabled="page === 1" @click="page = Math.max(1, page - 1)">
            <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m11.5 5.5-4.5 4.5 4.5 4.5" /></svg>
          </button>
          <span class="pager-index">{{ page }}</span>
          <button class="pager-arrow" type="button" :disabled="page >= totalPages" @click="page = Math.min(totalPages, page + 1)">
            <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m8.5 5.5 4.5 4.5-4.5 4.5" /></svg>
          </button>
        </div>
        <label class="page-size-select">
          <select v-model.number="pageSize"><option :value="11">每页 11 条</option><option :value="20">每页 20 条</option><option :value="50">每页 50 条</option></select>
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg>
        </label>
      </footer>
    </div>

    <div v-else class="policy-editor-shell">
      <header class="policy-editor-hero">
        <div class="policy-editor-copy"><p class="section-eyebrow">策略编辑器</p><h1>编辑 {{ resolvePolicyName(activePolicy.name) }}</h1></div>
        <span class="health-line"><span class="led led-online"></span> 配置生效</span>
      </header>

      <section v-if="auditPreview" class="policy-audit-panel">
        <article><span>用户组</span><strong>{{ auditPreview.groups.length }}</strong><small>{{ auditPreview.groups.map((group) => group.name).join(', ') || '暂无绑定用户组' }}</small></article>
        <article><span>影响用户</span><strong>{{ auditPreview.user_count }}</strong><small>按用户组绑定实时统计</small></article>
        <article><span>已有文件</span><strong>{{ auditPreview.existing_file_count }}</strong><small>历史 Blob 元数据保持不变</small></article>
        <article><span>新上传</span><strong>{{ auditPreview.new_upload_config.chunk_size }}{{ auditPreview.new_upload_config.chunk_size_unit }}</strong><small>{{ auditPreview.new_upload_config.parallel_chunk_count }} 路并行 · {{ auditPreview.new_upload_config.enable_cdn ? 'CDN' : '本地下载' }} · {{ auditPreview.new_upload_config.enable_encryption ? '加密' : '明文' }}</small></article>
      </section>

      <section class="policy-audit-history">
        <div class="section-heading">
          <h2>最近变更记录</h2>
          <small>创建、修改、删除都会记录操作者、影响范围和配置快照。</small>
        </div>
        <article v-for="record in auditRecords" :key="record.id" class="audit-record">
          <div class="audit-record-main">
            <div>
              <strong>{{ resolveAuditAction(record.action) }}</strong>
              <span>{{ record.operator_name || `#${record.operator_id}` }}</span>
            </div>
            <div class="audit-record-actions">
              <button class="audit-text-button" type="button" @click="toggleAuditDiff(record)">查看差异</button>
              <button v-if="canRollbackAudit(record)" class="audit-text-button danger" type="button" :disabled="saving" @click="rollbackToAudit(record)">回滚到此版本</button>
            </div>
          </div>
          <p>{{ formatAuditTime(record.created_at) }} · {{ record.user_count }} 位用户 · {{ record.groups.map((group) => group.name).join(', ') || '暂无绑定用户组' }}{{ record.source_audit_id ? ` · 来源 #${record.source_audit_id}` : '' }}</p>
          <div v-if="expandedAuditId === record.id" class="audit-diff-list">
            <div v-for="diff in resolveAuditDiffs(record)" :key="diff.key" class="audit-diff-row">
              <span>{{ diff.label }}</span>
              <code>{{ diff.before }}</code>
              <strong>→</strong>
              <code>{{ diff.after }}</code>
            </div>
            <p v-if="!resolveAuditDiffs(record).length">本次没有可显示的配置差异。</p>
          </div>
        </article>
        <p v-if="!auditRecords.length" class="empty-audit-text">暂无变更记录</p>
      </section>

      <section class="policy-hit-history">
        <div class="section-heading">
          <h2>最近命中记录</h2>
          <small>上传、分片、下载、预览和分享下载实际命中的策略。</small>
        </div>
        <article v-for="hit in hitRecords" :key="hit.id" class="hit-record">
          <div class="hit-record-main">
            <div>
              <strong>{{ resolveHitAction(hit.action) }}</strong>
              <span>{{ hit.username || `#${hit.user_id}` }} · {{ hit.user_group_name || '未绑定用户组' }}</span>
            </div>
            <span class="hit-type-chip">{{ resolveHitType(hit.hit_type) }}</span>
          </div>
          <p>{{ hit.file_name || hit.resource_id || '未记录文件' }} · {{ formatFileSize(hit.file_size) }} · {{ formatAuditTime(hit.created_at) }}</p>
          <div class="hit-config-row">
            <span v-for="item in resolveHitConfig(hit)" :key="item">{{ item }}</span>
          </div>
        </article>
        <p v-if="!hitRecords.length" class="empty-audit-text">暂无命中记录</p>
      </section>

      <div class="policy-form">
        <section class="form-section">
          <h2>基础信息</h2>
          <label class="field-block"><span>名称</span><input v-model="form.name" type="text" class="field-input" /><small>存储策略展示名。</small></label>
        </section>

        <section class="form-section">
          <h2>存储与上传</h2>
          <label class="field-block"><span>Blob 存储目录</span><input v-model="form.blobPath" type="text" class="field-input" /><small>新 Blob 的物理目录，已有文件不迁移。</small></label>
          <label class="field-block"><span>Blob 名称</span><input v-model="form.blobNamePattern" type="text" class="field-input" /><small>新 Blob 的命名模式。</small></label>
          <label class="field-block"><span>文件大小限制</span><div class="split-input"><input v-model.number="form.maxFileSize" type="number" class="field-input" /><label class="unit-select"><select v-model="form.maxFileSizeUnit"><option value="KB">KB</option><option value="MB">MB</option><option value="GB">GB</option></select><svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg></label></div><small>0 表示不限制。</small></label>
          <label class="field-block"><span>文件扩展名限制</span><div class="rule-row"><label class="prefix-select"><select v-model="form.extensionMode"><option value="allow">允许</option><option value="deny">拒绝</option></select><svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg></label><input v-model="form.extensions" type="text" class="field-input" placeholder="无限制" /></div><small>多个扩展名用半角逗号分隔。</small></label>
          <label class="field-block"><span>文件名正则规则</span><div class="rule-row"><label class="prefix-select"><select v-model="form.nameRuleMode"><option value="allow">允许</option><option value="deny">拒绝</option></select><svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg></label><input v-model="form.nameRegex" type="text" class="field-input" placeholder="无限制" /></div><small>留空表示无限制。</small></label>
          <label class="field-block"><span>上传分片大小</span><div class="split-input"><input v-model.number="form.chunkSize" type="number" class="field-input" /><label class="unit-select"><select v-model="form.chunkSizeUnit"><option value="KB">KB</option><option value="MB">MB</option><option value="GB">GB</option></select><svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5.5 7.5 4.5 5 4.5-5" /></svg></label></div><small>0 表示不启用分片。</small></label>
          <label class="toggle-field"><button class="switch-button" :class="{ active: form.preAllocate }" type="button" @click="form.preAllocate = !form.preAllocate"><span></span></button><div><strong>预分配硬盘空间</strong><small>用于并行分片上传前的空间预留。</small></div></label>
          <label class="field-block"><span>并行上传分片数</span><input v-model.number="form.parallelChunkCount" type="number" class="field-input" /><small>Web 端同时上传的分片数量。</small></label>
        </section>

        <section class="form-section">
          <h2>下载</h2>
          <label class="checkbox-field"><input v-model="form.enableCdn" type="checkbox" /><div><strong>使用 CDN 加速下载</strong><small>开启后安全场景会优先重定向 CDN。</small></div></label>
          <label class="field-block"><span>下载 CDN</span><input v-model="form.downloadCdn" type="text" class="field-input" placeholder="https://cdn.example.com" /><small>留空则继续后端流式下载。</small></label>
        </section>

        <section class="form-section">
          <div class="section-heading"><h2>文件加密</h2><button class="help-badge" type="button" aria-label="File encryption help">?</button></div>
          <label class="toggle-field"><button class="switch-button muted" :class="{ active: form.enableEncryption }" type="button" @click="form.enableEncryption = !form.enableEncryption"><span></span></button><div><strong>启用文件加密</strong><small>新 Blob 会加密存储，历史 Blob 不改变。</small></div></label>
          <label v-if="form.enableEncryption" class="field-block"><span>加密密钥标识</span><input v-model="form.encryptionKeyId" type="text" class="field-input" placeholder="default-master-key" /><small>写入物理 Blob 元数据，供下载解密使用。</small></label>
        </section>
      </div>

      <footer class="editor-actions">
        <button class="ghost-button" type="button" @click="goBackToList">返回列表</button>
        <button class="ghost-button" type="button" :disabled="saving" @click="duplicatePolicy(form)">复制策略</button>
        <button class="ghost-button" type="button" :disabled="saving" @click="downloadPolicyJson(form)">导出 JSON</button>
        <button class="ghost-button" type="button" :disabled="saving" @click="triggerImport(activePolicyId)">导入覆盖</button>
        <button class="ghost-button" type="button" @click="resetForm">恢复当前配置</button>
        <button class="primary-button" type="button" :disabled="saving" @click="savePolicy">{{ saving ? '保存中' : '保存存储策略' }}</button>
      </footer>
    </div>
  </section>
</template>
<script setup lang="ts">
import {
  copyStoragePolicy,
  createStoragePolicy,
  deleteStoragePolicy,
  exportStoragePolicy,
  getStoragePolicyAudit,
  getStoragePolicy,
  importStoragePolicy,
  listStoragePolicyHits,
  listStoragePolicyAudits,
  listStoragePolicies,
  migrateStoragePolicyGroups,
  previewStoragePolicy,
  rollbackStoragePolicy,
  updateStoragePolicy,
  type StoragePolicyAuditPayload,
  type StoragePolicyGroupCoverage,
  type StoragePolicyHitLogPayload,
  type StoragePolicyPreviewPayload,
  type StoragePolicyPayload,
  type StoragePolicyType,
} from '@/api/storage-policy';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

type StoragePolicy = {
  id: number;
  name: string;
  type: StoragePolicyType;
  groups: string[];
  blobPath: string;
  blobNamePattern: string;
  maxFileSize: number;
  maxFileSizeUnit: 'KB' | 'MB' | 'GB';
  extensionMode: 'allow' | 'deny';
  extensions: string;
  nameRuleMode: 'allow' | 'deny';
  nameRegex: string;
  chunkSize: number;
  chunkSizeUnit: 'KB' | 'MB' | 'GB';
  preAllocate: boolean;
  parallelChunkCount: number;
  enableCdn: boolean;
  downloadCdn: string;
  enableEncryption: boolean;
  encryptionKeyId: string;
  effectiveUserCount: number;
  effectiveFileCount: number;
};

const route = useRoute();
const router = useRouter();

const policyFilters = [
  { label: '全部', value: 'all' },
  { label: '本地存储', value: 'local' },
  { label: '从机存储', value: 'remote' },
  { label: '阿里云 OSS', value: 'oss' },
  { label: 'OneDrive', value: 'onedrive' },
  { label: '腾讯云 COS', value: 'cos' },
  { label: 'S3 / MinIO', value: 's3' },
  { label: '华为云 OBS', value: 'obs' },
  { label: '负载均衡', value: 'balance' },
] as const;

const selectedFilter = ref<(typeof policyFilters)[number]['value']>('all');
const page = ref(1);
const pageSize = ref(11);
const policies = ref<StoragePolicy[]>([]);
const activePolicyDetail = ref<StoragePolicy | null>(null);
const loading = ref(false);
const saving = ref(false);
const auditPreview = ref<StoragePolicyPreviewPayload | null>(null);
const auditRecords = ref<StoragePolicyAuditPayload[]>([]);
const hitRecords = ref<StoragePolicyHitLogPayload[]>([]);
const expandedAuditId = ref<number | null>(null);
const importFileInputRef = ref<HTMLInputElement | null>(null);
const importOverwriteId = ref<number | null>(null);
const migrationDialog = reactive<{
  visible: boolean;
  loading: boolean;
  sourcePolicy: StoragePolicy | null;
  targetPolicyId: number;
  groups: StoragePolicyGroupCoverage[];
  userCount: number;
}>({
  visible: false,
  loading: false,
  sourcePolicy: null,
  targetPolicyId: 0,
  groups: [],
  userCount: 0,
});

const form = reactive<StoragePolicy>(clonePolicy(createDefaultPolicy()));
const originalForm = ref<StoragePolicy>(clonePolicy(form));

const activePolicyId = computed(() => Number(route.params.policyId || 0));
const activePolicy = computed(() => {
  const id = activePolicyId.value;
  if (id > 0 && activePolicyDetail.value?.id === id) {
    return activePolicyDetail.value;
  }
  return policies.value.find((item) => item.id === id) ?? null;
});
const filteredPolicies = computed(() =>
  selectedFilter.value === 'all' ? policies.value : policies.value.filter((item) => item.type === selectedFilter.value),
);
const totalPages = computed(() => Math.max(1, Math.ceil(filteredPolicies.value.length / pageSize.value)));
const pagedPolicies = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return filteredPolicies.value.slice(start, start + pageSize.value);
});
const migrationTargetPolicies = computed(() =>
  policies.value.filter((item) => item.id > 0 && item.id !== migrationDialog.sourcePolicy?.id),
);

const policyInsights = computed(() => [
  { label: '本地策略', value: policies.value.filter((item) => item.type === 'local').length || 1, detail: '本地服务器集群' },
  { label: '对象存储', value: policies.value.filter((item) => ['s3', 'oss', 'cos', 'obs'].includes(item.type)).length, detail: 'MinIO / OSS / COS / OBS' },
  { label: '加密策略', value: policies.value.filter((item) => item.enableEncryption).length, detail: '启用 Blob 加密的策略' },
]);

const storageNodes = computed(() =>
  [
    { name: '本地服务器集群', detail: '本地存储 · 主节点', latency: '12 ms', online: true },
    { name: 'MinIO 对象存储', detail: 'S3 兼容 · 对象存储', latency: '28 ms', online: policies.value.some((item) => item.type === 's3') },
    { name: '阿里云 OSS', detail: '华东区域 · 冷热分层', latency: '35 ms', online: policies.value.some((item) => item.type === 'oss') },
    { name: '边缘工作节点', detail: '分布式边缘节点 · 负载均衡', latency: '18 ms', online: policies.value.some((item) => item.type === 'balance') },
  ].filter((node) => node.online),
);

watch(
  () => pageSize.value,
  () => {
    page.value = 1;
  },
);

watch(
  () => selectedFilter.value,
  () => {
    page.value = 1;
  },
);

watch(
  () => totalPages.value,
  (value) => {
    if (page.value > value) {
      page.value = value;
    }
  },
  { immediate: true },
);

watch(
  activePolicy,
  (policy) => {
    if (policy) {
      Object.assign(form, clonePolicy(policy));
      originalForm.value = clonePolicy(policy);
      return;
    }

    auditPreview.value = null;
    auditRecords.value = [];
    hitRecords.value = [];
    Object.assign(form, clonePolicy(createDefaultPolicy()));
    originalForm.value = clonePolicy(form);
  },
  { immediate: true },
);

watch(
  () => activePolicyId.value,
  (id) => {
    if (id > 0) {
      fetchPolicyDetail(id);
    } else {
      activePolicyDetail.value = null;
    }
  },
  { immediate: true },
);

function createDefaultPolicy(): StoragePolicy {
  return {
    id: 0,
    name: '默认存储策略',
    type: 'local',
    groups: ['Admin', 'User'],
    blobPath: '/cloudreve/data/uploads/{uid}/{path}',
    blobNamePattern: '{uid}_{randomkey8}_{originname}',
    maxFileSize: 0,
    maxFileSizeUnit: 'MB',
    extensionMode: 'allow',
    extensions: '',
    nameRuleMode: 'allow',
    nameRegex: '',
    chunkSize: 25,
    chunkSizeUnit: 'MB',
    preAllocate: true,
    parallelChunkCount: 1,
    enableCdn: false,
    downloadCdn: '',
    enableEncryption: false,
    encryptionKeyId: '',
    effectiveUserCount: 0,
    effectiveFileCount: 0,
  };
}

function clonePolicy(policy: StoragePolicy): StoragePolicy {
  return JSON.parse(JSON.stringify(policy));
}

function mapPolicyFromPayload(payload: StoragePolicyPayload): StoragePolicy {
  return {
    id: Number(payload.id || 0),
    name: payload.name,
    type: payload.type,
    groups: Array.isArray(payload.groups) ? payload.groups : [],
    blobPath: payload.blob_path,
    blobNamePattern: payload.blob_name_pattern,
    maxFileSize: payload.max_file_size,
    maxFileSizeUnit: payload.max_file_size_unit,
    extensionMode: payload.extension_mode,
    extensions: payload.extensions,
    nameRuleMode: payload.name_rule_mode,
    nameRegex: payload.name_regex,
    chunkSize: payload.chunk_size,
    chunkSizeUnit: payload.chunk_size_unit,
    preAllocate: payload.pre_allocate,
    parallelChunkCount: payload.parallel_chunk_count,
    enableCdn: payload.enable_cdn,
    downloadCdn: payload.download_cdn,
    enableEncryption: payload.enable_encryption,
    encryptionKeyId: payload.encryption_key_id,
    effectiveUserCount: Number(payload.effective_user_count || 0),
    effectiveFileCount: Number(payload.effective_file_count || 0),
  };
}

function mapPolicyToPayload(policy: StoragePolicy): StoragePolicyPayload {
  return {
    id: policy.id || undefined,
    name: policy.name,
    type: policy.type,
    groups: policy.groups,
    blob_path: policy.blobPath,
    blob_name_pattern: policy.blobNamePattern,
    max_file_size: policy.maxFileSize,
    max_file_size_unit: policy.maxFileSizeUnit,
    extension_mode: policy.extensionMode,
    extensions: policy.extensions,
    name_rule_mode: policy.nameRuleMode,
    name_regex: policy.nameRegex,
    chunk_size: policy.chunkSize,
    chunk_size_unit: policy.chunkSizeUnit,
    pre_allocate: policy.preAllocate,
    parallel_chunk_count: policy.parallelChunkCount,
    enable_cdn: policy.enableCdn,
    download_cdn: policy.downloadCdn,
    enable_encryption: policy.enableEncryption,
    encryption_key_id: policy.encryptionKeyId,
    effective_user_count: policy.effectiveUserCount,
    effective_file_count: policy.effectiveFileCount,
  };
}

function resolveTypeLabel(type: StoragePolicyType) {
  return policyFilters.find((item) => item.value === type)?.label || '本地存储';
}

function resolveNodeName(type: StoragePolicyType) {
  const names: Record<StoragePolicyType, string> = {
    local: '本地服务器集群',
    remote: '从机存储节点',
    oss: '阿里云 OSS',
    onedrive: 'OneDrive 网关',
    cos: '腾讯云 COS',
    s3: 'MinIO 对象存储',
    obs: '华为云 OBS',
    balance: '边缘均衡池',
  };
  return names[type];
}

function resolvePolicyName(name: string) {
  return name;
}

function resolveGroupLabel(group: string) {
  const groups: Record<string, string> = {
    Admin: 'Admin',
    User: 'User',
  };
  return groups[group] || group;
}

function resolvePolicySummary(policy: StoragePolicy) {
  const size = policy.maxFileSize > 0 ? `${policy.maxFileSize}${policy.maxFileSizeUnit}` : '不限大小';
  const encryption = policy.enableEncryption ? '已加密' : '未加密';
  return `${size} · ${encryption} · ${policy.effectiveUserCount} 位用户 · ${policy.effectiveFileCount} 个文件`;
}

async function fetchPolicies(silent = false) {
  loading.value = true;
  try {
    const data = await listStoragePolicies();
    policies.value = data.map(mapPolicyFromPayload);
    if (!silent) {
      ElMessage.success('存储策略已刷新');
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载存储策略失败');
  } finally {
    loading.value = false;
  }
}

async function fetchPolicyDetail(id: number) {
  try {
    const [data, preview, audits, hits] = await Promise.all([
      getStoragePolicy(id),
      previewStoragePolicy(id),
      listStoragePolicyAudits(id),
      listStoragePolicyHits(id),
    ]);
    const policy = mapPolicyFromPayload(data);
    activePolicyDetail.value = policy;
    auditPreview.value = preview;
    auditRecords.value = audits;
    hitRecords.value = hits;
    const index = policies.value.findIndex((item) => item.id === id);
    if (index >= 0) {
      policies.value.splice(index, 1, policy);
    } else {
      policies.value.push(policy);
    }
    Object.assign(form, clonePolicy(policy));
    originalForm.value = clonePolicy(policy);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载存储策略详情失败');
    goBackToList();
  }
}

function reloadPolicies() {
  fetchPolicies();
}

function openPolicy(id: number) {
  router.push(`/admin/storage-policy/${id}`);
}
function goBackToList() {
  router.push('/admin/storage-policy');
}

async function createPolicy() {
  try {
    const policy = createDefaultPolicy();
    policy.name = '新增存储策略';
    policy.type = selectedFilter.value === 'all' ? 'local' : (selectedFilter.value as StoragePolicyType);
    const created = await createStoragePolicy(mapPolicyToPayload(policy));
    const mapped = mapPolicyFromPayload(created);
    policies.value.unshift(mapped);
    ElMessage.success('存储策略已创建');
    openPolicy(mapped.id);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建存储策略失败');
  }
}

async function duplicatePolicy(policy: StoragePolicy) {
  if (!policy.id) {
    return;
  }
  saving.value = true;
  try {
    const created = await copyStoragePolicy(policy.id);
    const mapped = mapPolicyFromPayload(created);
    policies.value.unshift(mapped);
    ElMessage.success('存储策略副本已创建');
    openPolicy(mapped.id);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '复制存储策略失败');
  } finally {
    saving.value = false;
  }
}

async function downloadPolicyJson(policy: StoragePolicy) {
  if (!policy.id) {
    return;
  }
  try {
    const exported = await exportStoragePolicy(policy.id);
    const content = JSON.stringify(exported, null, 2);
    const blob = new Blob([content], { type: 'application/json;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${safeExportFileName(resolvePolicyName(policy.name)) || 'storage-policy'}.json`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
    ElMessage.success('存储策略 JSON 已导出');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '导出存储策略失败');
  }
}

function triggerImport(overwriteId?: number) {
  importOverwriteId.value = overwriteId && overwriteId > 0 ? overwriteId : null;
  if (importFileInputRef.value) {
    importFileInputRef.value.value = '';
    importFileInputRef.value.click();
  }
}

async function handleImportFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    return;
  }

  try {
    const text = await file.text();
    const parsed = JSON.parse(text) as StoragePolicyPayload | { policy?: StoragePolicyPayload } | null;
    if (!parsed || typeof parsed !== 'object') {
      throw new Error('导入文件不是有效的存储策略 JSON');
    }
    const payload = normalizeImportedPolicyPayload('policy' in parsed && parsed.policy ? parsed.policy : parsed);
    if (importOverwriteId.value) {
      await ElMessageBox.confirm('确认用导入 JSON 覆盖当前存储策略？覆盖会写入审计记录，已有 Blob 元数据不会被迁移。', '导入覆盖策略', {
        type: 'warning',
        confirmButtonText: '覆盖',
        cancelButtonText: '取消',
        distinguishCancelAndClose: true,
      });
    }

    saving.value = true;
    const imported = await importStoragePolicy(payload, importOverwriteId.value || undefined);
    const mapped = mapPolicyFromPayload(imported);
    const index = policies.value.findIndex((item) => item.id === mapped.id);
    if (index >= 0) {
      policies.value.splice(index, 1, mapped);
    } else {
      policies.value.unshift(mapped);
    }
    activePolicyDetail.value = activePolicyId.value === mapped.id ? mapped : activePolicyDetail.value;
    ElMessage.success(importOverwriteId.value ? '存储策略已导入覆盖' : '存储策略已导入');
    await fetchPolicies(true);
    openPolicy(mapped.id);
  } catch (error) {
    if (error instanceof Error) {
      ElMessage.error(error.message);
    } else {
      ElMessage.error('导入存储策略失败');
    }
  } finally {
    saving.value = false;
    importOverwriteId.value = null;
    input.value = '';
  }
}

function normalizeImportedPolicyPayload(value: StoragePolicyPayload | Record<string, unknown>): StoragePolicyPayload {
  if (!value || typeof value !== 'object') {
    throw new Error('导入文件不是有效的存储策略 JSON');
  }
  const payload = value as StoragePolicyPayload;
  if (!payload.name || !payload.type || !payload.blob_path || !payload.blob_name_pattern) {
    throw new Error('导入文件缺少必要的存储策略字段');
  }
  return payload;
}

function safeExportFileName(value: string) {
  return value
    .trim()
    .replace(/[\\/:*?"<>|]+/g, '-')
    .replace(/\s+/g, '-')
    .slice(0, 80);
}

function canDeletePolicy(policy: StoragePolicy) {
  return policy.id > 0 && policy.name !== '默认存储策略';
}

async function removePolicy(policy: StoragePolicy) {
  try {
    await ElMessageBox.confirm(`确认删除“${resolvePolicyName(policy.name)}”？此操作不可撤销。`, '删除存储策略', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      distinguishCancelAndClose: true,
    });
    await deleteStoragePolicy(policy.id);
    policies.value = policies.value.filter((item) => item.id !== policy.id);
    ElMessage.success('存储策略已删除');
    if (activePolicyId.value === policy.id) {
      goBackToList();
    }
  } catch (error) {
    if (error instanceof Error && (await openMigrationDialogIfBound(policy))) {
      return;
    }
    if (error instanceof Error) {
      ElMessage.error(error.message);
    }
  }
}

async function openMigrationDialogIfBound(policy: StoragePolicy) {
  try {
    const preview = await previewStoragePolicy(policy.id);
    if (!preview.groups.length) {
      return false;
    }
    migrationDialog.sourcePolicy = policy;
    migrationDialog.groups = preview.groups;
    migrationDialog.userCount = preview.user_count;
    migrationDialog.targetPolicyId = migrationTargetPolicies.value[0]?.id || 0;
    migrationDialog.visible = true;
    return true;
  } catch {
    return false;
  }
}

function closeMigrationDialog() {
  if (migrationDialog.loading) {
    return;
  }
  migrationDialog.visible = false;
  migrationDialog.sourcePolicy = null;
  migrationDialog.targetPolicyId = 0;
  migrationDialog.groups = [];
  migrationDialog.userCount = 0;
}

async function migrateAndDeletePolicy() {
  const source = migrationDialog.sourcePolicy;
  if (!source || !migrationDialog.targetPolicyId) {
    return;
  }

  migrationDialog.loading = true;
  try {
    await migrateStoragePolicyGroups(source.id, {
      target_policy_id: migrationDialog.targetPolicyId,
      group_ids: migrationDialog.groups.map((group) => group.id),
    });
    await deleteStoragePolicy(source.id);
    policies.value = policies.value.filter((item) => item.id !== source.id);
    ElMessage.success('用户组已迁移，原存储策略已删除');
    migrationDialog.loading = false;
    closeMigrationDialog();
    await fetchPolicies(true);
    if (activePolicyId.value === source.id) {
      goBackToList();
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '迁移用户组失败');
  } finally {
    migrationDialog.loading = false;
  }
}

function resetForm() {
  Object.assign(form, clonePolicy(originalForm.value));
  ElMessage.success('已恢复当前配置');
}

async function savePolicy() {
  if (!activePolicyId.value) {
    return;
  }

  saving.value = true;
  try {
    const saved = await updateStoragePolicy(activePolicyId.value, mapPolicyToPayload(form));
    const [preview, audits, hits] = await Promise.all([
      previewStoragePolicy(activePolicyId.value),
      listStoragePolicyAudits(activePolicyId.value),
      listStoragePolicyHits(activePolicyId.value),
    ]);
    auditPreview.value = preview;
    auditRecords.value = audits;
    hitRecords.value = hits;
    const mapped = mapPolicyFromPayload(saved);
    activePolicyDetail.value = mapped;
    const index = policies.value.findIndex((item) => item.id === activePolicyId.value);
    if (index >= 0) {
      policies.value.splice(index, 1, mapped);
    }
    Object.assign(form, clonePolicy(mapped));
    originalForm.value = clonePolicy(mapped);
    ElMessage.success('存储策略已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存存储策略失败');
  } finally {
    saving.value = false;
  }
}

function resolveAuditAction(action: string) {
  const labels: Record<string, string> = {
    create: '创建策略',
    update: '修改策略',
    delete: '删除策略',
    copy: '复制策略',
    import: '导入策略',
    import_overwrite: '导入覆盖',
    rollback: '回滚策略',
    migrate_groups: '迁移用户组',
    repair_legacy: '修复历史数据',
  };
  return labels[action] || action;
}

function resolveHitAction(action: string) {
  const labels: Record<string, string> = {
    upload: '直接上传',
    multipart_init: '分片初始化',
    multipart_complete: '分片合并',
    download: '文件下载',
    preview: '文件预览',
    share_download: '分享下载',
  };
  return labels[action] || action;
}

function resolveHitType(type: string) {
  const labels: Record<string, string> = {
    user_group_policy: '用户组策略',
    global_default: '全局默认',
  };
  return labels[type] || type;
}

function resolveHitConfig(hit: StoragePolicyHitLogPayload) {
  const config = hit.config || {};
  const parts: string[] = [];
  if (typeof config.chunk_size === 'number' && typeof config.chunk_size_unit === 'string' && config.chunk_size_unit) {
    parts.push(`${config.chunk_size}${String(config.chunk_size_unit || '')} 分片`);
  } else if (typeof config.chunk_size === 'number' && config.chunk_size > 0) {
    parts.push(`${formatFileSize(config.chunk_size)} 分片`);
  }
  if (typeof config.parallel_chunk_count === 'number' && config.parallel_chunk_count > 0) {
    parts.push(`${config.parallel_chunk_count} 路并行`);
  }
  if (typeof config.delivery === 'string') {
    parts.push(resolveDeliveryLabel(config.delivery));
  } else if (config.enable_cdn === true) {
    parts.push('CDN 已开启');
  }
  if (config.enable_encryption === true) {
    parts.push('加密');
  }
  if (config.pre_allocate === true) {
    parts.push('预分配');
  }
  return parts.slice(0, 5);
}

function resolveDeliveryLabel(value: string) {
  const labels: Record<string, string> = {
    cdn: 'CDN',
    local_stream: '本地流式',
    converted_pdf: 'PDF 预览',
    zip_stream: '打包下载',
  };
  return labels[value] || value;
}

function formatFileSize(size: number) {
  if (!Number.isFinite(size) || size <= 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value >= 10 || index === 0 ? value.toFixed(0) : value.toFixed(1)} ${units[index]}`;
}

function canRollbackAudit(record: StoragePolicyAuditPayload) {
  return record.action !== 'delete' && Boolean(record.after);
}

async function toggleAuditDiff(record: StoragePolicyAuditPayload) {
  if (expandedAuditId.value === record.id) {
    expandedAuditId.value = null;
    return;
  }

  expandedAuditId.value = record.id;
  if (!activePolicyId.value) {
    return;
  }

  try {
    const detail = await getStoragePolicyAudit(activePolicyId.value, record.id);
    const index = auditRecords.value.findIndex((item) => item.id === record.id);
    if (index >= 0) {
      auditRecords.value.splice(index, 1, detail);
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载策略变更详情失败');
  }
}

async function rollbackToAudit(record: StoragePolicyAuditPayload) {
  if (!activePolicyId.value || !canRollbackAudit(record)) {
    return;
  }

  try {
    await ElMessageBox.confirm(`确认回滚到 #${record.id} 的策略版本？当前配置会被新的历史版本覆盖。`, '回滚存储策略', {
      type: 'warning',
      confirmButtonText: '回滚',
      cancelButtonText: '取消',
      distinguishCancelAndClose: true,
    });

    saving.value = true;
    const saved = await rollbackStoragePolicy(activePolicyId.value, record.id);
    const [preview, audits, hits] = await Promise.all([
      previewStoragePolicy(activePolicyId.value),
      listStoragePolicyAudits(activePolicyId.value),
      listStoragePolicyHits(activePolicyId.value),
    ]);
    auditPreview.value = preview;
    auditRecords.value = audits;
    hitRecords.value = hits;
    expandedAuditId.value = null;

    const mapped = mapPolicyFromPayload(saved);
    activePolicyDetail.value = mapped;
    const index = policies.value.findIndex((item) => item.id === activePolicyId.value);
    if (index >= 0) {
      policies.value.splice(index, 1, mapped);
    }
    Object.assign(form, clonePolicy(mapped));
    originalForm.value = clonePolicy(mapped);
    ElMessage.success('存储策略已回滚');
  } catch (error) {
    if (error instanceof Error) {
      ElMessage.error(error.message);
    }
  } finally {
    saving.value = false;
  }
}

function resolveAuditDiffs(record: StoragePolicyAuditPayload) {
  const before = record.before;
  const after = record.after;
  const fields: Array<{ key: keyof StoragePolicyPayload; label: string }> = [
    { key: 'name', label: '名称' },
    { key: 'type', label: '类型' },
    { key: 'groups', label: '策略用户组' },
    { key: 'blob_path', label: 'Blob 存储目录' },
    { key: 'blob_name_pattern', label: 'Blob 名称' },
    { key: 'max_file_size', label: '文件大小限制' },
    { key: 'max_file_size_unit', label: '文件大小单位' },
    { key: 'extension_mode', label: '扩展名模式' },
    { key: 'extensions', label: '扩展名规则' },
    { key: 'name_rule_mode', label: '文件名规则模式' },
    { key: 'name_regex', label: '文件名正则' },
    { key: 'chunk_size', label: '上传分片大小' },
    { key: 'chunk_size_unit', label: '分片大小单位' },
    { key: 'pre_allocate', label: '预分配硬盘空间' },
    { key: 'parallel_chunk_count', label: '并行上传分片数' },
    { key: 'enable_cdn', label: '使用 CDN 加速下载' },
    { key: 'download_cdn', label: '下载 CDN' },
    { key: 'enable_encryption', label: '启用文件加密' },
    { key: 'encryption_key_id', label: '加密密钥标识' },
  ];

  return fields
    .map((field) => ({
      key: field.key,
      label: field.label,
      before: formatAuditValue(before?.[field.key]),
      after: formatAuditValue(after?.[field.key]),
    }))
    .filter((item) => item.before !== item.after);
}

function formatAuditValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.length ? value.join(', ') : '-';
  }
  if (typeof value === 'boolean') {
    return value ? '开启' : '关闭';
  }
  if (value === undefined || value === null || value === '') {
    return '-';
  }
  return String(value);
}

function formatAuditTime(value: string) {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

onMounted(() => {
  fetchPolicies(true);
});
</script>

<style scoped>
.storage-policy-page {
  min-height: calc(100vh - 104px);
  padding: 4px 2px 20px;
  color: #172033;
}

.hidden-file-input {
  display: none;
}

.migration-backdrop {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: grid;
  place-items: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.28);
  backdrop-filter: blur(12px);
}

.migration-dialog {
  display: grid;
  gap: 18px;
  width: min(620px, 100%);
  padding: 24px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.92), rgba(245, 250, 255, 0.86));
  box-shadow: 0 30px 80px rgba(15, 23, 42, 0.18), inset 0 1px 18px rgba(255, 255, 255, 0.72);
}

.migration-dialog header,
.migration-dialog footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.migration-dialog h2 {
  margin: 6px 0 0;
  color: #172033;
  font-size: 28px;
  font-weight: 900;
}

.migration-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: 1px solid rgba(218, 228, 240, 0.88);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  color: #5c6d83;
  font-size: 24px;
  cursor: pointer;
}

.migration-summary {
  margin: 0;
  color: #5c6d83;
  font-size: 14px;
  font-weight: 800;
  line-height: 1.75;
}

.migration-group-list {
  display: grid;
  gap: 10px;
  max-height: 240px;
  overflow: auto;
}

.migration-group-list article {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(218, 228, 240, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.62);
}

.migration-group-list strong {
  color: #172033;
  font-size: 14px;
  font-weight: 900;
}

.migration-group-list span {
  color: #74849b;
  font-size: 12px;
  font-weight: 800;
}

.migration-select {
  width: 100%;
}

.policy-shell,
.policy-editor-shell {
  position: relative;
  overflow: hidden;
  min-height: calc(100vh - 120px);
  padding: 30px;
  border: 1px solid rgba(226, 235, 246, 0.88);
  border-radius: 30px;
  background:
    radial-gradient(circle at 8% 4%, rgba(186, 230, 253, 0.58), transparent 28%),
    radial-gradient(circle at 96% 18%, rgba(251, 207, 232, 0.42), transparent 30%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.86), rgba(248, 252, 255, 0.76));
  box-shadow: 0 26px 70px rgba(52, 75, 118, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(22px);
}

.policy-shell::before,
.policy-editor-shell::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(120deg, rgba(255, 255, 255, 0.55), transparent 38%, rgba(219, 234, 254, 0.32));
  pointer-events: none;
}

.policy-hero,
.policy-editor-hero,
.policy-toolbar,
.policy-overview,
.policy-layout,
.policy-footer,
.policy-audit-panel,
  .policy-audit-history,
  .policy-hit-history,
  .policy-form,
.editor-actions {
  position: relative;
  z-index: 1;
}

.policy-hero,
.policy-editor-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 26px;
}

.section-eyebrow {
  margin: 0;
  color: #2f80d8;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.policy-hero-copy h1,
.policy-editor-copy h1 {
  margin: 8px 0 0;
  color: #172033;
  font-size: 56px;
  line-height: 1.02;
  font-weight: 900;
  letter-spacing: -0.04em;
}

.policy-hero-copy p:not(.section-eyebrow) {
  max-width: 760px;
  margin: 14px 0 0;
  color: #66768f;
  font-size: 15px;
  line-height: 1.85;
}

.hero-orb {
  display: grid;
  place-items: center;
  width: 132px;
  height: 132px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 32px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(235, 248, 255, 0.54));
  box-shadow: 0 22px 45px rgba(96, 165, 250, 0.16), inset 0 1px 16px rgba(255, 255, 255, 0.8);
}

.hero-orb strong {
  color: #172033;
  font-size: 34px;
}

.hero-orb small {
  color: #74849b;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.policy-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
}

.toolbar-button,
.filter-select,
.page-size-select,
.ghost-button,
.primary-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
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
  padding: 0 16px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.toolbar-button:hover,
.ghost-button:hover,
.primary-button:hover,
.policy-card:hover,
.node-tile:hover,
.add-policy-card:hover {
  transform: translateY(-1px);
}

.toolbar-button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.toolbar-button svg,
.filter-select svg,
.page-size-select svg,
.policy-delete-button svg,
.pager-arrow svg,
.unit-select svg,
.prefix-select svg {
  width: 18px;
  height: 18px;
  flex: none;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.filter-select,
.page-size-select,
.unit-select,
.prefix-select {
  position: relative;
  padding-right: 40px;
}

.filter-select select,
.page-size-select select,
.unit-select select,
.prefix-select select {
  appearance: none;
  min-height: 48px;
  padding: 0 16px;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  outline: none;
  cursor: pointer;
}

.filter-select svg,
.page-size-select svg,
.unit-select svg,
.prefix-select svg {
  position: absolute;
  right: 14px;
  pointer-events: none;
}

.policy-overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 18px;
}

.policy-audit-panel {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  max-width: 980px;
  margin: -8px 0 22px;
}

.policy-audit-panel article {
  display: grid;
  gap: 8px;
  min-width: 0;
  padding: 16px 18px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 18px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.58));
  box-shadow: 0 18px 38px rgba(87, 109, 143, 0.1), inset 0 1px 14px rgba(255, 255, 255, 0.62);
}

.policy-audit-panel span,
.policy-audit-panel small {
  overflow-wrap: anywhere;
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.policy-audit-panel strong {
  color: #172033;
  font-size: 24px;
  font-weight: 900;
}

.policy-audit-history {
  display: grid;
  gap: 12px;
  max-width: 980px;
  margin: 0 0 22px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.58));
  box-shadow: 0 18px 38px rgba(87, 109, 143, 0.1), inset 0 1px 14px rgba(255, 255, 255, 0.62);
}

.policy-hit-history {
  display: grid;
  gap: 12px;
  max-width: 980px;
  margin: 0 0 22px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.58));
  box-shadow: 0 18px 38px rgba(87, 109, 143, 0.1), inset 0 1px 14px rgba(255, 255, 255, 0.62);
}

.policy-hit-history .section-heading {
  align-items: flex-end;
  justify-content: space-between;
}

.policy-hit-history .section-heading small,
.hit-record p {
  margin: 0;
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.hit-record {
  display: grid;
  gap: 8px;
  padding: 12px 14px;
  border: 1px solid rgba(218, 228, 240, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.62);
}

.hit-record-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.hit-record-main > div {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.hit-record strong {
  color: #172033;
  font-size: 14px;
  font-weight: 900;
}

.hit-record span {
  color: #506179;
  font-size: 12px;
  font-weight: 800;
}

.hit-type-chip,
.hit-config-row span {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  min-height: 24px;
  padding: 0 9px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.76);
  color: #2f80d8;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.hit-config-row {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.policy-audit-history .section-heading {
  align-items: flex-end;
  justify-content: space-between;
}

.policy-audit-history .section-heading small,
.empty-audit-text,
.audit-record p {
  margin: 0;
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.audit-record {
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  border: 1px solid rgba(218, 228, 240, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.62);
}

.audit-record-main,
.audit-record-main > div,
.audit-record-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.audit-record-main > div,
.audit-record-actions {
  flex-wrap: wrap;
  justify-content: flex-start;
}

.audit-record strong {
  color: #172033;
  font-size: 14px;
  font-weight: 900;
}

.audit-record span {
  color: #506179;
  font-size: 12px;
  font-weight: 800;
}

.audit-text-button {
  padding: 0;
  border: 0;
  background: transparent;
  color: #2f80d8;
  font-size: 12px;
  font-weight: 900;
  cursor: pointer;
}

.audit-text-button.danger {
  color: #dc2626;
}

.audit-text-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.audit-diff-list {
  display: grid;
  gap: 8px;
  padding-top: 8px;
}

.audit-diff-row {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid rgba(218, 228, 240, 0.72);
  border-radius: 12px;
  background: rgba(248, 252, 255, 0.72);
}

.audit-diff-row span,
.audit-diff-row strong {
  color: #506179;
  font-size: 12px;
  font-weight: 900;
}

.audit-diff-row code {
  overflow-wrap: anywhere;
  padding: 6px 8px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.78);
  color: #172033;
  font-family: inherit;
  font-size: 12px;
  font-weight: 800;
}

.overview-card,
.node-panel,
.add-policy-card,
.policy-card,
.form-section {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.58));
  box-shadow: 0 20px 46px rgba(87, 109, 143, 0.12), inset 0 1px 18px rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(18px);
}

.overview-card {
  display: grid;
  gap: 8px;
  padding: 18px 20px;
  border-radius: 20px;
}

.overview-card span,
.overview-card small {
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.overview-card strong {
  color: #172033;
  font-size: 28px;
}

.policy-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 18px;
  align-items: start;
}

.policy-board {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

.add-policy-card,
.policy-card {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
}

.add-policy-card {
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

.add-policy-icon {
  font-size: 40px;
  font-weight: 300;
  line-height: 1;
}

.policy-card {
  display: grid;
  gap: 18px;
  padding: 22px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.policy-card::before {
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

.policy-card:hover {
  border-color: rgba(251, 207, 232, 0.9);
  box-shadow: 0 24px 54px rgba(232, 121, 179, 0.13), 0 14px 34px rgba(96, 165, 250, 0.12);
}

.policy-card:hover::before {
  opacity: 1;
}

.policy-card-header,
.policy-card-body,
.policy-card-footer {
  position: relative;
  z-index: 1;
}

.policy-card-header,
.policy-card-body,
.policy-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.health-line,
.type-chip {
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

.type-chip {
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.62);
  color: #3b638d;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.led {
  position: relative;
  width: 10px;
  height: 10px;
  flex: none;
  border-radius: 999px;
}

.led::after {
  content: '';
  position: absolute;
  inset: -7px;
  border-radius: inherit;
  opacity: 0.72;
  animation: ledPulse 1.9s ease-in-out infinite;
}

.led-online {
  background: #26d67f;
  box-shadow: 0 0 14px rgba(38, 214, 127, 0.92);
}

.led-online::after {
  background: radial-gradient(circle, rgba(38, 214, 127, 0.34), transparent 68%);
}

.led-idle {
  background: #a7b7c8;
  box-shadow: 0 0 12px rgba(148, 163, 184, 0.45);
}

.led-idle::after {
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

.policy-card-copy {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.policy-card-copy strong {
  color: #172033;
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.policy-card-copy p {
  margin: 0;
  color: #687991;
  font-size: 13px;
  line-height: 1.7;
}

.policy-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.policy-chip {
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

.policy-stats-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: #2f80d8;
  font-size: 14px;
  font-weight: 900;
  cursor: pointer;
}

.policy-card-footer-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.policy-card-footer span {
  padding-right: 58px;
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.policy-card-art {
  display: grid;
  justify-items: center;
  gap: 8px;
  min-width: 150px;
  color: #5c6d83;
  font-size: 12px;
  font-weight: 900;
  text-align: center;
}

.storage-cube {
  position: relative;
  display: grid;
  gap: 7px;
  width: 86px;
  padding: 14px 12px;
  border: 1px solid rgba(255, 255, 255, 0.88);
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.76), rgba(226, 242, 255, 0.56));
  box-shadow: 0 18px 36px rgba(73, 112, 170, 0.14), inset 0 1px 18px rgba(255, 255, 255, 0.82);
}

.storage-cube span {
  height: 9px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(147, 197, 253, 0.72), rgba(251, 207, 232, 0.72));
}

.storage-cube span:nth-child(2) {
  width: 72%;
}

.storage-cube span:nth-child(3) {
  width: 52%;
}

.policy-delete-button {
  position: absolute;
  right: 18px;
  bottom: 18px;
  z-index: 2;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border: 1px solid rgba(218, 228, 240, 0.8);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  color: #606a78;
  cursor: pointer;
  box-shadow: 0 14px 28px rgba(84, 105, 138, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.82);
  transition: transform 0.2s ease, color 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.policy-delete-button:hover,
.policy-delete-button:focus-visible {
  border-color: rgba(248, 113, 113, 0.36);
  background: rgba(255, 241, 242, 0.78);
  color: #dc2626;
  transform: translateY(-1px);
  outline: none;
}

.node-panel {
  position: sticky;
  top: 18px;
  display: grid;
  gap: 16px;
  padding: 20px;
  border-radius: 24px;
}

.panel-title {
  display: grid;
  gap: 8px;
}

.panel-title strong {
  color: #172033;
  font-size: 20px;
  font-weight: 900;
}

.node-list {
  display: grid;
  gap: 10px;
}

.node-tile {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.54);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.node-tile strong {
  display: block;
  color: #223044;
  font-size: 14px;
}

.node-tile small,
.node-tile > span:last-child {
  color: #74849b;
  font-size: 12px;
  font-weight: 700;
}

.policy-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 18px;
}

.pager {
  display: inline-flex;
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
  opacity: 0.35;
  cursor: not-allowed;
}

.pager-index {
  background: rgba(255, 255, 255, 0.7);
  color: #172033;
  font-weight: 900;
}

.policy-form {
  display: grid;
  gap: 22px;
  max-width: 980px;
}

.form-section {
  display: grid;
  gap: 20px;
  padding: 26px;
  border-radius: 24px;
}

.form-section h2,
.section-heading h2 {
  margin: 0;
  color: #172033;
  font-size: 28px;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.field-block,
.toggle-field,
.checkbox-field {
  display: grid;
  gap: 10px;
}

.field-block > span,
.toggle-field strong,
.checkbox-field strong {
  color: #223044;
  font-size: 16px;
  font-weight: 900;
}

.field-block small,
.toggle-field small,
.checkbox-field small {
  color: #6f7f94;
  font-size: 13px;
  line-height: 1.75;
}

.field-input {
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
}

.field-input:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 4px rgba(147, 197, 253, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.split-input,
.rule-row,
.section-heading,
.editor-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

.split-input .field-input,
.rule-row .field-input {
  flex: 1;
}

.unit-select,
.prefix-select {
  flex: none;
  min-width: 126px;
  min-height: 54px;
  border: 1px solid rgba(210, 222, 236, 0.95);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.68);
  color: #2f3b4d;
  font-size: 15px;
  font-weight: 800;
  box-sizing: border-box;
}

.prefix-select {
  min-width: 142px;
}

.toggle-field,
.checkbox-field {
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 12px 14px;
}

.switch-button {
  position: relative;
  width: 60px;
  height: 34px;
  margin-top: 4px;
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
  border-radius: 50%;
  background: linear-gradient(145deg, #fff, #eef7ff);
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.2), inset 0 1px 0 #fff;
  transition: transform 0.2s ease;
}

.switch-button.active {
  background: linear-gradient(135deg, rgba(125, 211, 252, 0.9), rgba(251, 207, 232, 0.9));
  box-shadow: 0 0 0 4px rgba(251, 207, 232, 0.16), inset 0 2px 8px rgba(59, 130, 246, 0.16);
}

.switch-button.active span {
  transform: translateX(26px);
}

.switch-button.muted.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.72), rgba(20, 184, 166, 0.72));
}

.checkbox-field {
  grid-template-columns: 24px minmax(0, 1fr);
}

.checkbox-field input {
  width: 22px;
  height: 22px;
  margin-top: 4px;
  accent-color: #38bdf8;
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

.editor-actions {
  justify-content: flex-end;
  max-width: 980px;
  padding-top: 4px;
}

.primary-button {
  border: 0;
  background: linear-gradient(135deg, #38bdf8, #f9a8d4);
  color: #fff;
  box-shadow: 0 18px 34px rgba(56, 189, 248, 0.2), 0 14px 28px rgba(249, 168, 212, 0.18);
}

.primary-button:disabled,
.ghost-button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

@media (max-width: 1180px) {
  .policy-layout {
    grid-template-columns: 1fr;
  }

  .node-panel {
    position: static;
  }
}

@media (max-width: 900px) {
  .policy-shell,
  .policy-editor-shell {
    padding: 22px 18px 24px;
    border-radius: 24px;
  }

  .policy-hero,
  .policy-editor-hero,
  .policy-card-body,
  .policy-card-footer {
    display: grid;
    grid-template-columns: 1fr;
  }

  .policy-hero-copy h1,
  .policy-editor-copy h1 {
    font-size: 38px;
  }

  .hero-orb {
    width: 100%;
    height: auto;
    min-height: 96px;
  }

  .policy-overview {
    grid-template-columns: 1fr;
  }

  .policy-audit-panel {
    grid-template-columns: 1fr;
  }

  .audit-record-main,
  .audit-diff-row {
    grid-template-columns: 1fr;
    display: grid;
  }

  .policy-toolbar,
  .policy-footer,
  .split-input,
  .rule-row,
  .editor-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .toggle-field,
  .checkbox-field {
    grid-template-columns: 1fr;
  }

  .policy-card-art {
    justify-items: start;
    min-width: 0;
  }
}
</style>

