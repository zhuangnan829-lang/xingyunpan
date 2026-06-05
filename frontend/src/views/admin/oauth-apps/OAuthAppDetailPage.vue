<template>
  <section class="oauth-detail-page">
    <div v-if="app" v-loading="loading" class="detail-shell">
      <header class="detail-hero">
        <button class="back-button" type="button" @click="router.push('/admin/oauth')">
          <ArrowLeft />
          <span>返回应用列表</span>
        </button>
        <div class="hero-content">
          <div class="detail-mark" aria-hidden="true">
            <span></span>
          </div>
          <div>
            <p class="hero-kicker">OAuth Application</p>
            <h1>编辑 {{ app.name }}</h1>
            <p>{{ app.description }}</p>
          </div>
        </div>
      </header>

      <section class="detail-card">
        <div class="section-head">
          <h2>基本信息</h2>
          <p>控制应用展示名称、图标、客户端凭据和启用状态。</p>
        </div>

        <div class="system-notice">
          <InfoFilled />
          <span>这是系统 OAuth 应用。</span>
        </div>

        <label class="toggle-row">
          <input v-model="form.enabled" type="checkbox" />
          <span class="toggle-track"></span>
          <strong>启用 OAuth 应用</strong>
          <small>禁用后，此 OAuth 应用将无法用于身份验证。</small>
        </label>

        <div class="form-grid">
          <label class="field-row">
            <span>应用名称</span>
            <input v-model="form.appName" type="text" />
            <small>此 OAuth 应用的显示名称。可使用 i18next 翻译键进行本地化。</small>
          </label>

          <label class="field-row icon-field">
            <span>应用图标</span>
            <div class="icon-input-row">
              <div class="icon-preview" aria-hidden="true"><span></span></div>
              <input v-model="form.iconPath" type="text" />
            </div>
            <small>此 OAuth 应用的图标。支持图片 URL 或 Iconify 图标 ID。</small>
          </label>

          <label class="field-row">
            <span>客户端 ID</span>
            <div class="copy-input">
              <input v-model="form.clientId" type="text" readonly />
              <button type="button" title="复制客户端 ID" @click="copy(form.clientId)">
                <CopyDocument />
              </button>
            </div>
            <small>此 OAuth 应用的唯一标识符。配置应用程序时使用此 ID。</small>
          </label>

          <label class="field-row">
            <span>客户端密钥</span>
            <input v-model="form.clientSecret" type="password" />
            <small>用于验证此 OAuth 应用的密钥。请妥善保管，切勿公开泄露。</small>
          </label>
        </div>
      </section>

      <section class="detail-card">
        <div class="section-head">
          <h2>OAuth 配置</h2>
          <p>设置授权后允许的回调地址和刷新令牌有效期。</p>
        </div>

        <div class="form-grid compact">
          <label class="field-row span-2">
            <span>重定向 URI</span>
            <textarea v-model="redirectUrisText" rows="4"></textarea>
            <small>授权后允许的回调 URL。每行输入一个 URL。OAuth 流程中必须完全匹配。</small>
          </label>

          <label class="field-row unit-field">
            <span>刷新令牌有效期</span>
            <div class="unit-input">
              <input v-model.number="form.refreshTokenTtlSeconds" type="number" min="0" />
              <em>秒</em>
            </div>
            <small>刷新令牌的有效时长。设为 0 则遵循全局设置。</small>
          </label>
        </div>
      </section>

      <section class="detail-card">
        <div class="section-head">
          <h2>权限</h2>
          <p>选择此 OAuth 应用可以请求的权限。用户在授权时将看到这些权限。</p>
        </div>
        <OAuthPermissionList :permissions="form.permissions" />
      </section>

      <OAuthCredentialPanel :app="credentialApp" />

      <div class="sticky-actions">
        <button class="save-button" type="button" @click="save">
          <FolderChecked />
          <span>保存</span>
        </button>
        <button class="reset-button" type="button" @click="reset">
          <RefreshLeft />
          <span>撤销更改</span>
        </button>
      </div>
    </div>

    <div v-else class="missing-card">
      <h1>未找到 OAuth 应用</h1>
      <button type="button" @click="router.push('/admin/oauth')">返回应用列表</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ArrowLeft, CopyDocument, FolderChecked, InfoFilled, RefreshLeft } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { getOAuthApp, updateOAuthApp } from '@/api/admin-oauth-apps';
import OAuthCredentialPanel from './components/OAuthCredentialPanel.vue';
import OAuthPermissionList from './components/OAuthPermissionList.vue';
import type { OAuthApp } from './types';

const route = useRoute();
const router = useRouter();

const createEmptyApp = (): OAuthApp => ({
  id: '',
  name: '',
  description: '',
  appName: '',
  iconPath: '',
  clientId: '',
  clientSecret: '',
  redirectUris: [],
  scopes: [],
  isSystem: false,
  enabled: true,
  createdAt: '',
  updatedAt: '',
  tokenTtl: '7 天',
  refreshTokenTtlSeconds: 604800,
  permissions: [],
});

const app = ref<OAuthApp | null>(null);
const loading = ref(false);
const form = reactive<OAuthApp>(createEmptyApp());
const redirectUrisText = ref(form.redirectUris.join('\n'));

const credentialApp = computed<OAuthApp>(() => ({
  ...form,
  redirectUris: redirectUrisText.value
    .split('\n')
    .map((uri) => uri.trim())
    .filter(Boolean),
}));

const loadApp = async () => {
  const appId = String(route.params.appId || '');
  if (!appId) return;
  loading.value = true;
  try {
    const data = await getOAuthApp(appId);
    app.value = data;
    Object.assign(form, cloneApp(data));
    redirectUrisText.value = data.redirectUris.join('\n');
  } catch (error) {
    app.value = null;
    ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用加载失败');
  } finally {
    loading.value = false;
  }
};

watch(() => route.params.appId, loadApp);

function cloneApp(value: OAuthApp): OAuthApp {
  return {
    ...value,
    redirectUris: [...value.redirectUris],
    scopes: [...value.scopes],
    permissions: value.permissions.map((permission) => ({ ...permission })),
  };
}

const copy = async (value: string) => {
  try {
    await navigator.clipboard.writeText(value);
    ElMessage.success('已复制到剪贴板');
  } catch {
    ElMessage.error('复制失败，请手动复制');
  }
};

const save = async () => {
  form.redirectUris = credentialApp.value.redirectUris;
  form.scopes = form.permissions.filter((permission) => permission.enabled || permission.required).map((permission) => permission.scope);
  try {
    const updated = await updateOAuthApp(form);
    app.value = updated;
    Object.assign(form, cloneApp(updated));
    redirectUrisText.value = updated.redirectUris.join('\n');
    ElMessage.success('OAuth 应用配置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用配置保存失败');
  }
};

const reset = () => {
  if (!app.value) return;
  Object.assign(form, cloneApp(app.value));
  redirectUrisText.value = app.value.redirectUris.join('\n');
  ElMessage.info('已撤销未保存更改');
};

onMounted(loadApp);
</script>

<style scoped>
.oauth-detail-page {
  min-height: calc(100vh - 96px);
  color: #172033;
}

.detail-shell {
  position: relative;
  display: grid;
  gap: 22px;
  min-height: calc(100vh - 96px);
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 30px;
  background:
    radial-gradient(circle at 10% 4%, rgba(125, 211, 252, 0.32), transparent 30%),
    radial-gradient(circle at 88% 6%, rgba(252, 188, 202, 0.28), transparent 26%),
    radial-gradient(circle at 18% 100%, rgba(205, 183, 255, 0.16), transparent 28%),
    linear-gradient(135deg, rgba(248, 252, 255, 0.92), rgba(255, 249, 252, 0.84) 52%, rgba(246, 251, 255, 0.92));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.96), 0 28px 68px rgba(91, 145, 186, 0.13);
}

.detail-hero,
.detail-card,
.missing-card {
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 26px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.44));
  box-shadow: 0 22px 54px rgba(81, 120, 154, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(18px);
}

.detail-hero {
  display: grid;
  gap: 22px;
  padding: 26px;
}

.back-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: max-content;
  min-height: 40px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.56);
  color: #37516a;
  font-weight: 800;
  cursor: pointer;
}

.back-button svg {
  width: 17px;
  height: 17px;
}

.hero-content {
  display: flex;
  align-items: flex-start;
  gap: 18px;
}

.detail-mark,
.icon-preview {
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 18px;
  background:
    radial-gradient(circle at 72% 20%, #31d5ef 0 20%, transparent 21%),
    linear-gradient(135deg, #1f6fe8 0%, #67d9ff 58%, #fff8fb 100%);
  box-shadow: 0 14px 30px rgba(44, 150, 226, 0.22), inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.detail-mark {
  width: 64px;
  height: 64px;
}

.detail-mark span,
.icon-preview span {
  width: 42px;
  height: 18px;
  border-radius: 999px 999px 12px 12px;
  background: rgba(255, 255, 255, 0.8);
  transform: translateY(8px);
}

.hero-kicker,
.detail-hero h1,
.detail-hero p {
  margin: 0;
}

.hero-kicker {
  color: #6c7c90;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.detail-hero h1 {
  margin-top: 10px;
  color: #152235;
  font-size: clamp(34px, 4vw, 56px);
  line-height: 1.08;
  font-weight: 880;
}

.detail-hero p:not(.hero-kicker) {
  max-width: 780px;
  margin-top: 12px;
  color: #66768a;
  font-size: 16px;
  line-height: 1.75;
}

.detail-card {
  display: grid;
  gap: 22px;
  padding: 26px;
}

.section-head {
  display: grid;
  gap: 10px;
}

.section-head h2 {
  margin: 0;
  color: #172033;
  font-size: 28px;
  font-weight: 860;
}

.section-head p {
  margin: 0;
  color: #617187;
  font-size: 15px;
  line-height: 1.7;
}

.system-notice {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  width: min(100%, 720px);
  min-height: 58px;
  padding: 0 18px;
  border-radius: 18px;
  background: rgba(219, 242, 255, 0.82);
  color: #096b9d;
  font-weight: 820;
}

.system-notice svg {
  width: 22px;
  height: 22px;
  color: #1198dc;
}

.toggle-row {
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  align-items: center;
  column-gap: 14px;
  row-gap: 6px;
  width: max-content;
  max-width: 100%;
}

.toggle-row input {
  display: none;
}

.toggle-track {
  position: relative;
  width: 52px;
  height: 24px;
  border-radius: 999px;
  background: #bfd9f2;
  cursor: pointer;
}

.toggle-track::after {
  content: '';
  position: absolute;
  top: -3px;
  left: 18px;
  width: 30px;
  height: 30px;
  border-radius: 999px;
  background: #2385dd;
  box-shadow: 0 8px 18px rgba(35, 133, 221, 0.28);
  transition: left 0.2s ease, background 0.2s ease;
}

.toggle-row input:not(:checked) + .toggle-track::after {
  left: 0;
  background: #94a3b8;
}

.toggle-row strong {
  color: #172033;
  font-size: 17px;
  font-weight: 840;
}

.toggle-row small {
  grid-column: 2;
  color: #66768a;
  font-size: 13px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  max-width: 1120px;
}

.form-grid.compact {
  max-width: 760px;
}

.field-row {
  display: grid;
  gap: 9px;
}

.field-row.span-2 {
  grid-column: 1 / -1;
}

.field-row > span {
  color: #172033;
  font-size: 15px;
  font-weight: 840;
}

.field-row input,
.field-row textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(211, 224, 238, 0.96);
  border-radius: 16px;
  outline: 0;
  background: rgba(255, 255, 255, 0.72);
  color: #162336;
  font: inherit;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.13);
}

.field-row input {
  min-height: 52px;
  padding: 0 18px;
}

.field-row textarea {
  min-height: 132px;
  padding: 16px 18px;
  resize: vertical;
  line-height: 1.7;
}

.field-row small {
  color: #657487;
  font-size: 13px;
  line-height: 1.65;
}

.icon-input-row {
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
}

.icon-preview {
  width: 52px;
  height: 52px;
  border-radius: 16px;
}

.icon-preview span {
  width: 34px;
  height: 15px;
  transform: translateY(6px);
}

.copy-input,
.unit-input {
  position: relative;
}

.copy-input button {
  position: absolute;
  right: 12px;
  top: 11px;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 0;
  background: transparent;
  color: #7a8794;
  cursor: pointer;
}

.copy-input input {
  padding-right: 54px;
}

.unit-input em {
  position: absolute;
  right: 18px;
  top: 15px;
  color: #66768a;
  font-style: normal;
  font-weight: 840;
}

.unit-input input {
  padding-right: 56px;
}

.sticky-actions {
  position: sticky;
  bottom: 18px;
  z-index: 5;
  display: flex;
  gap: 12px;
  width: max-content;
  max-width: calc(100vw - 48px);
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.76);
  box-shadow: 0 18px 42px rgba(76, 113, 144, 0.16), inset 0 1px 0 rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(18px);
}

.save-button,
.reset-button,
.missing-card button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 46px;
  padding: 0 18px;
  border-radius: 15px;
  font-weight: 840;
  cursor: pointer;
}

.save-button {
  border: 0;
  background: linear-gradient(135deg, #1f74e8, #18bddf);
  color: #fff;
  box-shadow: 0 12px 24px rgba(33, 135, 219, 0.22);
}

.reset-button,
.missing-card button {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background: rgba(255, 255, 255, 0.62);
  color: #5d6b7c;
}

.save-button svg,
.reset-button svg {
  width: 18px;
  height: 18px;
}

.missing-card {
  display: grid;
  gap: 18px;
  padding: 32px;
}

.missing-card h1 {
  margin: 0;
}

@media (max-width: 920px) {
  .detail-shell {
    padding: 16px;
    border-radius: 24px;
  }

  .hero-content {
    flex-direction: column;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .sticky-actions {
    width: auto;
  }
}
</style>
