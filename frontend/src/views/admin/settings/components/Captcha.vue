<template>
  <div class="captcha-panel">
    <section class="panel-card hero-card">
      <div>
        <p class="eyebrow">Security Check</p>
        <h2>验证码</h2>
        <p class="hero-text">
          配置登录、注册、找回密码等场景的验证码策略。保存后会直接写入后台验证码设置接口。
        </p>
      </div>

      <div class="hero-status">
        <article class="status-card">
          <span>模块状态</span>
          <strong>{{ securityLevelLabel }}</strong>
        </article>
        <article class="status-card">
          <span>启用场景</span>
          <strong>{{ enabledSceneCount }}</strong>
        </article>
      </div>
    </section>

    <section class="panel-card">
      <div class="section-head">
        <div>
          <p class="eyebrow">基础策略</p>
          <h3>启用场景</h3>
        </div>
      </div>

      <div class="toggle-grid">
        <label class="switch-card">
          <div>
            <strong>登录启用验证码</strong>
            <p>用户登录时触发验证码校验，降低撞库与爆破风险。</p>
          </div>
          <input v-model="form.loginEnabled" type="checkbox" />
        </label>

        <label class="switch-card">
          <div>
            <strong>注册启用验证码</strong>
            <p>用户注册时启用验证码，减少恶意注册和垃圾账户。</p>
          </div>
          <input v-model="form.registerEnabled" type="checkbox" />
        </label>

        <label class="switch-card">
          <div>
            <strong>找回密码启用验证码</strong>
            <p>找回密码流程增加验证码校验，提高账户恢复安全性。</p>
          </div>
          <input v-model="form.resetPasswordEnabled" type="checkbox" />
        </label>
      </div>
    </section>

    <section class="panel-card">
      <div class="section-head">
        <div>
          <p class="eyebrow">服务配置</p>
          <h3>验证码供应商</h3>
        </div>
      </div>

      <div class="form-grid two-column">
        <label class="field">
          <span>验证码类型</span>
          <select v-model="form.provider" class="field-select">
            <option value="image">图形验证码</option>
            <option value="slider">滑块验证码</option>
            <option value="turnstile">Turnstile</option>
            <option value="recaptcha">reCAPTCHA</option>
          </select>
        </label>

        <label class="field">
          <span>难度等级</span>
          <select v-model="form.securityLevel" class="field-select">
            <option value="balanced">平衡</option>
            <option value="strict">严格</option>
            <option value="relaxed">宽松</option>
          </select>
        </label>

        <label class="field">
          <span>站点 Key</span>
          <input v-model.trim="form.siteKey" class="field-input" type="text" placeholder="请输入站点 Key" />
        </label>

        <label class="field">
          <span>服务端 Secret</span>
          <input v-model.trim="form.secretKey" class="field-input" type="password" placeholder="请输入服务端 Secret" />
        </label>

        <label class="field">
          <span>失败阈值</span>
          <input v-model.number="form.failureThreshold" class="field-input" type="number" min="1" max="10" />
        </label>

        <label class="field">
          <span>冷却时间（秒）</span>
          <input v-model.number="form.cooldownSeconds" class="field-input" type="number" min="0" max="3600" />
        </label>

        <label class="field field-wide">
          <span>白名单路径</span>
          <textarea
            v-model="form.whitelistPaths"
            class="field-textarea"
            rows="4"
            placeholder="每行一个路径，例如：/api/v1/health"
          ></textarea>
        </label>
      </div>
    </section>

    <section class="panel-card summary-card">
      <p class="eyebrow">运行说明</p>
      <ul class="summary-list">
        <li>当前验证码配置通过后台接口统一读写，便于后续部署到不同环境。</li>
        <li>第三方供应商模式下会校验站点 Key 和 Secret，避免误保存空配置。</li>
        <li>白名单路径按每行一项保存，后端会自动清理空行和多余空格。</li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  getCaptchaSettings,
  updateCaptchaSettings,
  type CaptchaSettingsPayload,
} from '@/api/captcha-settings';

type CaptchaProvider = 'image' | 'slider' | 'turnstile' | 'recaptcha';
type SecurityLevel = 'balanced' | 'strict' | 'relaxed';

type CaptchaForm = {
  loginEnabled: boolean;
  registerEnabled: boolean;
  resetPasswordEnabled: boolean;
  provider: CaptchaProvider;
  securityLevel: SecurityLevel;
  siteKey: string;
  secretKey: string;
  failureThreshold: number;
  cooldownSeconds: number;
  whitelistPaths: string;
};

const defaultFormState = (): CaptchaForm => ({
  loginEnabled: true,
  registerEnabled: true,
  resetPasswordEnabled: true,
  provider: 'image',
  securityLevel: 'balanced',
  siteKey: '',
  secretKey: '',
  failureThreshold: 3,
  cooldownSeconds: 60,
  whitelistPaths: '/api/v1/health',
});

const form = reactive<CaptchaForm>(defaultFormState());
const loading = ref(false);
const saving = ref(false);
const lastSavedSnapshot = ref(createSnapshot(defaultFormState()));

const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);
const enabledSceneCount = computed(() => [form.loginEnabled, form.registerEnabled, form.resetPasswordEnabled].filter(Boolean).length);
const securityLevelLabel = computed(() => {
  if (form.securityLevel === 'strict') return '严格';
  if (form.securityLevel === 'relaxed') return '宽松';
  return '平衡';
});

function createSnapshot(source: CaptchaForm) {
  return JSON.stringify({
    ...source,
    siteKey: source.siteKey.trim(),
    secretKey: source.secretKey.trim(),
    whitelistPaths: normalizeMultiline(source.whitelistPaths),
  });
}

function normalizeMultiline(value: string) {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)
    .join('\n');
}

function applyFormState(source: CaptchaForm) {
  form.loginEnabled = source.loginEnabled;
  form.registerEnabled = source.registerEnabled;
  form.resetPasswordEnabled = source.resetPasswordEnabled;
  form.provider = source.provider;
  form.securityLevel = source.securityLevel;
  form.siteKey = source.siteKey;
  form.secretKey = source.secretKey;
  form.failureThreshold = source.failureThreshold;
  form.cooldownSeconds = source.cooldownSeconds;
  form.whitelistPaths = source.whitelistPaths;
}

function normalizeForSave(source: CaptchaForm): CaptchaForm {
  return {
    ...source,
    siteKey: source.siteKey.trim(),
    secretKey: source.secretKey.trim(),
    whitelistPaths: normalizeMultiline(source.whitelistPaths),
    failureThreshold: Math.min(10, Math.max(1, Number(source.failureThreshold) || 1)),
    cooldownSeconds: Math.min(3600, Math.max(0, Number(source.cooldownSeconds) || 0)),
  };
}

function validateForm(source: CaptchaForm) {
  const requiresCredential = source.provider === 'turnstile' || source.provider === 'recaptcha';

  if (!source.loginEnabled && !source.registerEnabled && !source.resetPasswordEnabled) {
    throw new Error('至少需要启用一个验证码场景');
  }

  if (requiresCredential && !source.siteKey) {
    throw new Error('当前验证码类型要求填写站点 Key');
  }

  if (requiresCredential && !source.secretKey) {
    throw new Error('当前验证码类型要求填写服务端 Secret');
  }
}

function toFormState(payload: CaptchaSettingsPayload): CaptchaForm {
  return {
    loginEnabled: Boolean(payload.login_enabled),
    registerEnabled: Boolean(payload.register_enabled),
    resetPasswordEnabled: Boolean(payload.reset_password_enabled),
    provider: payload.provider ?? 'image',
    securityLevel: payload.security_level ?? 'balanced',
    siteKey: payload.site_key ?? '',
    secretKey: payload.secret_key ?? '',
    failureThreshold: payload.failure_threshold ?? 3,
    cooldownSeconds: payload.cooldown_seconds ?? 60,
    whitelistPaths: Array.isArray(payload.whitelist_paths) ? payload.whitelist_paths.join('\n') : '',
  };
}

function toPayload(source: CaptchaForm): CaptchaSettingsPayload {
  return {
    login_enabled: source.loginEnabled,
    register_enabled: source.registerEnabled,
    reset_password_enabled: source.resetPasswordEnabled,
    provider: source.provider,
    security_level: source.securityLevel,
    site_key: source.siteKey.trim(),
    secret_key: source.secretKey.trim(),
    failure_threshold: source.failureThreshold,
    cooldown_seconds: source.cooldownSeconds,
    whitelist_paths: normalizeMultiline(source.whitelistPaths)
      .split('\n')
      .map((item) => item.trim())
      .filter(Boolean),
  };
}

async function reload() {
  loading.value = true;
  try {
    const data = await getCaptchaSettings();
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('验证码配置已重新加载');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载验证码配置失败');
  } finally {
    loading.value = false;
  }
}

function reset() {
  const next = defaultFormState();
  applyFormState(next);
  ElMessage.success('验证码配置已恢复默认值，记得保存');
}

async function save() {
  saving.value = true;
  try {
    const next = normalizeForSave(form);
    validateForm(next);
    const data = await updateCaptchaSettings(toPayload(next));
    const saved = toFormState(data);
    applyFormState(saved);
    lastSavedSnapshot.value = createSnapshot(saved);
    ElMessage.success('验证码配置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存验证码配置失败');
  } finally {
    saving.value = false;
  }
}

defineExpose({ isDirty, loading, saving, reload, reset, save });

onMounted(async () => {
  await reload();
});
</script>

<style scoped>
.captcha-panel {
  display: grid;
  gap: 20px;
}

.panel-card,
.switch-card,
.status-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.05);
}

.panel-card {
  padding: 24px;
  border-radius: 24px;
}

.hero-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 20px;
}

.hero-status {
  display: grid;
  gap: 12px;
  align-content: start;
}

.status-card {
  display: grid;
  gap: 8px;
  padding: 18px;
  border-radius: 20px;
}

.eyebrow {
  margin: 0 0 10px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h2,
h3,
p,
strong,
span,
ul {
  margin: 0;
}

h2,
h3,
strong,
.field span {
  color: #0f172a;
}

.hero-text,
.status-card span,
.switch-card p,
.summary-list {
  color: #64748b;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.toggle-grid {
  display: grid;
  gap: 16px;
}

.switch-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px;
  border-radius: 20px;
}

.switch-card strong {
  display: block;
  margin-bottom: 6px;
}

.switch-card input {
  width: 18px;
  height: 18px;
}

.form-grid {
  display: grid;
  gap: 16px;
}

.two-column {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.field {
  display: grid;
  gap: 8px;
}

.field span {
  font-size: 14px;
  font-weight: 700;
}

.field-wide {
  grid-column: 1 / -1;
}

.field-input,
.field-select,
.field-textarea {
  width: 100%;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 16px;
  background: #fff;
  color: #0f172a;
}

.field-input,
.field-select {
  min-height: 48px;
  padding: 0 14px;
}

.field-textarea {
  min-height: 120px;
  padding: 14px;
  resize: vertical;
}

.field-input:focus,
.field-select:focus,
.field-textarea:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.8);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}

.summary-card {
  background: linear-gradient(180deg, #fff 0%, #f8fbff 100%);
}

.summary-list {
  padding-left: 18px;
  line-height: 1.9;
}

@media (max-width: 1100px) {
  .hero-card,
  .two-column {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .panel-card {
    padding: 18px;
  }
}
</style>
