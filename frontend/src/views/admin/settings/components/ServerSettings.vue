<template>
  <section class="server-settings">
    <div class="server-shell">
      <header class="server-header">
        <div>
          <p class="section-tag">参数设置</p>
          <h2>服务器设置</h2>
        </div>
      </header>

      <div class="server-layout">
        <div class="server-main">
          <section class="settings-card">
            <div class="card-head">
              <div>
                <p class="section-tag">服务器</p>
                <h3>服务器设置</h3>
              </div>
            </div>

            <div class="field-stack">
              <label v-for="field in textFields" :key="field.key" class="field-block">
                <span class="field-label">{{ field.label }}</span>
                <input v-model="form[field.key]" class="field-input" :type="field.type ?? 'text'" />
                <p class="field-hint">{{ field.hint }}</p>
              </label>

              <div class="field-block">
                <span class="field-label">主密钥</span>
                <div class="secret-row">
                  <button class="ghost-action" type="button" :disabled="saving" @click="rotateMasterKey">
                    <el-icon><RefreshRight /></el-icon>
                    <span>轮转主密钥</span>
                  </button>
                  <span class="secret-pill">最近轮转 {{ form.lastRotationAt }}</span>
                </div>
                <p class="field-hint">
                  用于加密用户令牌、签名的主密钥。轮转后，所有用户令牌、签名都将失效。保存后重启服务生效。
                </p>
              </div>
            </div>
          </section>
        </div>

        <aside class="server-side">
          <section class="summary-card">
            <p class="section-tag">运行概览</p>
            <div class="summary-list">
              <article class="summary-item">
                <span>会话窗口</span>
                <strong>{{ accessTokenHours }} 小时</strong>
              </article>
              <article class="summary-item">
                <span>刷新窗口</span>
                <strong>{{ refreshTokenDays }} 天</strong>
              </article>
              <article class="summary-item">
                <span>回收频率</span>
                <strong>{{ form.gcInterval }}</strong>
              </article>
            </div>
          </section>

          <section class="summary-card tone-muted">
            <p class="section-tag">注意事项</p>
            <ul class="tip-list">
              <li>修改 HashID 盐值会让现有直链和分享链接全部失效。</li>
              <li>访问令牌和刷新令牌 TTL 建议保持成对调整，避免登录状态不一致。</li>
              <li>这里先接入前端持久化，等后端接口补齐后可以直接替换为真实保存。</li>
            </ul>
          </section>
        </aside>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { RefreshRight } from '@element-plus/icons-vue';

type ServerSettingsForm = {
  tempPath: string;
  siteId: string;
  hashIdSalt: string;
  accessTokenTTL: string;
  refreshTokenTTL: string;
  gcInterval: string;
  lastRotationAt: string;
};

type FieldKey = keyof Pick<
  ServerSettingsForm,
  'tempPath' | 'siteId' | 'hashIdSalt' | 'accessTokenTTL' | 'refreshTokenTTL' | 'gcInterval'
>;

type FieldItem = {
  key: FieldKey;
  label: string;
  hint: string;
  type?: string;
};

const storageKey = 'xingyunpan.admin.server-settings';

const textFields: FieldItem[] = [
  {
    key: 'tempPath',
    label: '临时路径',
    hint: '存储临时文件的目录，相对于应用数据目录。修改前请确保没有正在运行的队列任务。',
  },
  {
    key: 'siteId',
    label: '站点 ID',
    hint: '用于标识站点的唯一 ID，一般无需修改。',
  },
  {
    key: 'hashIdSalt',
    label: 'HashID 盐值',
    hint: '用于生成 HashID 的盐值，请谨慎更改，更改后会导致现有的直链、分享链接等全部失效。',
  },
  {
    key: 'accessTokenTTL',
    label: '访问令牌 TTL',
    hint: '访问令牌的有效期，单位为秒。',
    type: 'number',
  },
  {
    key: 'refreshTokenTTL',
    label: '刷新令牌 TTL',
    hint: '刷新令牌的有效期，单位为秒。影响用户登录状态的保持时间。',
    type: 'number',
  },
  {
    key: 'gcInterval',
    label: '垃圾回收扫描间隔',
    hint: '设置多久扫描并回收临时文件和 KV 存储中的过期数据，此处需要填写正确的 Cron 表达式。',
  },
];

function createDefaultForm(): ServerSettingsForm {
  return {
    tempPath: 'temp',
    siteId: 'd2468e0f-2d42-4385-9a83-99bb61201c2a',
    hashIdSalt: 'jL99DwxhOfbNjGG7qOT4VYzjKwMCR5eP5DGrBOD0XBNF8bdL9ABUPrUP6QBC',
    accessTokenTTL: '3600',
    refreshTokenTTL: '1209600',
    gcInterval: '@every 30m',
    lastRotationAt: '今天',
  };
}

function cloneForm(source: ServerSettingsForm): ServerSettingsForm {
  return { ...source };
}

function createSnapshot(source: ServerSettingsForm): string {
  return JSON.stringify(source);
}

function loadFromStorage(): ServerSettingsForm {
  if (typeof window === 'undefined') return createDefaultForm();
  try {
    const raw = window.localStorage.getItem(storageKey);
    if (!raw) return createDefaultForm();
    return { ...createDefaultForm(), ...JSON.parse(raw) };
  } catch {
    return createDefaultForm();
  }
}

function saveToStorage(source: ServerSettingsForm) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(storageKey, JSON.stringify(source));
}

const form = reactive<ServerSettingsForm>(loadFromStorage());
const lastSavedSnapshot = ref(createSnapshot(form));
const loading = ref(false);
const saving = ref(false);
const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);
const accessTokenHours = computed(() => (Number(form.accessTokenTTL || 0) / 3600).toFixed(1).replace('.0', ''));
const refreshTokenDays = computed(() => (Number(form.refreshTokenTTL || 0) / 86400).toFixed(1).replace('.0', ''));

function applyForm(source: ServerSettingsForm) {
  Object.assign(form, cloneForm(source));
}

function generateRandomString(length: number) {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789';
  return Array.from({ length }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
}

function rotateMasterKey() {
  form.hashIdSalt = generateRandomString(64);
  form.lastRotationAt = '刚刚';
  ElMessage.success('已生成新的主密钥，记得点击保存');
}

async function reload() {
  loading.value = true;
  try {
    const next = loadFromStorage();
    applyForm(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('服务器设置已重新加载');
  } finally {
    loading.value = false;
  }
}

function reset() {
  const next = createDefaultForm();
  applyForm(next);
  ElMessage.success('服务器设置已恢复默认值，记得保存');
}

async function save() {
  if (!form.tempPath.trim()) {
    ElMessage.warning('临时路径不能为空');
    return;
  }

  if (!form.siteId.trim()) {
    ElMessage.warning('站点 ID 不能为空');
    return;
  }

  saving.value = true;
  try {
    const next = cloneForm(form);
    saveToStorage(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('服务器设置已保存到本地配置');
  } finally {
    saving.value = false;
  }
}

defineExpose({ isDirty, loading, saving, reload, reset, save });
</script>

<style scoped>
.server-settings,
.server-shell,
.server-main,
.server-side,
.field-stack,
.summary-list {
  display: grid;
  gap: 20px;
}

.server-shell {
  gap: 24px;
}

.server-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.server-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 24px;
  align-items: start;
}

.settings-card,
.summary-card {
  border: 1px solid rgba(226, 232, 240, 0.92);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.05);
}

.settings-card {
  padding: 28px 32px;
}

.summary-card {
  padding: 22px;
}

.tone-muted {
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.section-tag {
  margin: 0;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h2,
h3,
p,
span,
strong,
ul {
  margin: 0;
}

h2,
h3,
.field-label,
.summary-item strong {
  color: #0f172a;
}

h2 {
  margin-top: 8px;
  font-size: clamp(2.2rem, 4vw, 3rem);
  line-height: 1.05;
  letter-spacing: -0.04em;
}

h3 {
  margin-top: 6px;
  font-size: 28px;
}

.card-head {
  margin-bottom: 18px;
}

.field-block {
  display: grid;
  gap: 10px;
}

.field-label {
  font-size: 16px;
  font-weight: 800;
}

.field-input {
  width: min(100%, 640px);
  min-height: 46px;
  padding: 0 16px;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 16px;
  background: #fff;
  color: #0f172a;
}

.field-input:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.82);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}

.field-hint,
.summary-item span,
.tip-list {
  color: #64748b;
  line-height: 1.75;
}

.secret-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.ghost-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 44px;
  padding: 0 18px;
  border: none;
  border-radius: 16px;
  background: #f3f4f6;
  color: #475569;
  cursor: pointer;
}

.ghost-action:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.secret-pill {
  display: inline-flex;
  align-items: center;
  min-height: 36px;
  padding: 0 12px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 13px;
  font-weight: 700;
}

.summary-item {
  display: grid;
  gap: 8px;
  padding: 14px 0;
  border-bottom: 1px solid rgba(226, 232, 240, 0.9);
}

.summary-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.summary-item strong {
  font-size: 22px;
}

.tip-list {
  padding-left: 18px;
}

@media (max-width: 1100px) {
  .server-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .settings-card,
  .summary-card {
    padding: 18px;
  }

  .field-input {
    width: 100%;
  }
}
</style>
