<template>
  <div class="site-info-panel">
    <div class="layout-grid">
      <aside class="preview-column">
        <section class="panel-card preview-card">
          <p class="eyebrow">品牌预览</p>
          <div class="brand-preview brand-preview-light">
            <strong>{{ form.siteName || '星云盘' }}</strong>
            <span>{{ form.tagline || '新一代私有云盘控制台' }}</span>
          </div>
          <div class="brand-preview brand-preview-dark">
            <strong>{{ form.siteName || '星云盘' }}</strong>
            <span>{{ form.logoDark || form.logoLight || '/assets/branding/logo-dark.svg' }}</span>
          </div>
        </section>

        <section class="panel-card quick-card">
          <p class="eyebrow">状态</p>
          <div class="quick-item">
            <span>主站地址</span>
            <strong>{{ form.primaryUrl || '未设置' }}</strong>
          </div>
          <div class="quick-item">
            <span>备选域名</span>
            <strong>{{ normalizedBackupCount }}</strong>
          </div>
          <div class="quick-item">
            <span>移动端引导</span>
            <strong>{{ form.mobileGuideEnabled ? '已开启' : '已关闭' }}</strong>
          </div>
          <div class="quick-item">
            <span>桌面端引导</span>
            <strong>{{ form.desktopGuideEnabled ? '已开启' : '已关闭' }}</strong>
          </div>
        </section>
      </aside>

      <div class="content-column">
        <section class="panel-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">基础信息</p>
              <h2>品牌文案</h2>
            </div>
          </div>

          <div class="form-grid two-column">
            <label class="field">
              <span>站点名称</span>
              <input v-model="form.siteName" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>站点标语</span>
              <input v-model="form.tagline" class="field-input" type="text" />
            </label>

            <label class="field field-wide">
              <span>站点描述</span>
              <textarea v-model="form.description" class="field-textarea" rows="4"></textarea>
            </label>

            <label class="field">
              <span>使用条款链接</span>
              <input v-model="form.termsUrl" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>隐私政策链接</span>
              <input v-model="form.privacyUrl" class="field-input" type="text" />
            </label>
          </div>
        </section>

        <section class="panel-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">域名设置</p>
              <h2>访问入口</h2>
            </div>
            <button class="ghost-button" type="button" @click="addBackupUrl">
              <el-icon><Plus /></el-icon>
              <span>新增备选域名</span>
            </button>
          </div>

          <div class="form-grid">
            <label class="field">
              <span>主站 URL</span>
              <input v-model="form.primaryUrl" class="field-input" type="text" />
            </label>

            <div class="backup-list">
              <label v-for="(_, index) in form.backupUrls" :key="index" class="field backup-item">
                <span>备选 URL {{ index + 1 }}</span>
                <div class="backup-row">
                  <input v-model="form.backupUrls[index]" class="field-input" type="text" />
                  <button class="icon-button" type="button" @click="removeBackupUrl(index)">
                    <el-icon><Delete /></el-icon>
                  </button>
                </div>
              </label>

              <p v-if="form.backupUrls.length === 0" class="empty-hint">
                当前没有备选域名，可按需新增。
              </p>
            </div>
          </div>
        </section>

        <section class="panel-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">品牌资源</p>
              <h2>图标与 Logo</h2>
            </div>
          </div>

          <div class="form-grid two-column">
            <label class="field">
              <span>浅色 Logo</span>
              <input v-model="form.logoLight" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>深色 Logo</span>
              <input v-model="form.logoDark" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>Favicon</span>
              <input v-model="form.favicon" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>192px 图标</span>
              <input v-model="form.logo192" class="field-input" type="text" />
            </label>

            <label class="field field-wide">
              <span>自定义注入代码</span>
              <textarea v-model="form.injectionCode" class="field-textarea code-textarea" rows="8"></textarea>
            </label>
          </div>
        </section>

        <section class="panel-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">客户端引导</p>
              <h2>移动端与桌面端</h2>
            </div>
          </div>

          <div class="toggle-grid">
            <label class="switch-card">
              <div>
                <strong>移动端引导</strong>
                <p>控制是否展示移动客户端入口、下载说明或帮助链接。</p>
              </div>
              <input v-model="form.mobileGuideEnabled" type="checkbox" />
            </label>

            <label class="switch-card">
              <div>
                <strong>桌面端引导</strong>
                <p>控制是否展示桌面客户端安装说明或社区入口。</p>
              </div>
              <input v-model="form.desktopGuideEnabled" type="checkbox" />
            </label>
          </div>

          <div class="form-grid two-column">
            <label class="field">
              <span>移动端反馈 URL</span>
              <input v-model="form.mobileFeedbackUrl" class="field-input" type="text" />
            </label>

            <label class="field">
              <span>桌面端社区 URL</span>
              <input v-model="form.desktopCommunityUrl" class="field-input" type="text" />
            </label>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { Delete, Plus } from '@element-plus/icons-vue';
import { getSiteSettings, updateSiteSettings, type SiteSettingsPayload } from '@/api/site-settings';

type SiteSettingsForm = {
  siteName: string;
  tagline: string;
  description: string;
  termsUrl: string;
  privacyUrl: string;
  primaryUrl: string;
  backupUrls: string[];
  logoLight: string;
  logoDark: string;
  favicon: string;
  logo192: string;
  injectionCode: string;
  mobileGuideEnabled: boolean;
  mobileFeedbackUrl: string;
  desktopGuideEnabled: boolean;
  desktopCommunityUrl: string;
};

const defaultFormState = (): SiteSettingsForm => ({
  siteName: '星云盘',
  tagline: '新一代私有云盘控制台',
  description: '面向多节点云盘场景的现代化后台，强调品牌控制、访问路由与多端引导体验。',
  termsUrl: 'http://localhost:8080/terms',
  privacyUrl: 'http://localhost:8080/privacy-policy',
  primaryUrl: 'http://localhost:8080',
  backupUrls: [],
  logoLight: '/assets/branding/logo-light.svg',
  logoDark: '/assets/branding/logo-dark.svg',
  favicon: '/favicon.ico',
  logo192: '/logo192.png',
  injectionCode: '',
  mobileGuideEnabled: true,
  mobileFeedbackUrl: 'http://localhost:8080/support/mobile-feedback',
  desktopGuideEnabled: true,
  desktopCommunityUrl: 'http://localhost:8080/community',
});

const form = reactive<SiteSettingsForm>(defaultFormState());
const lastSavedSnapshot = ref('');
const loading = ref(false);
const saving = ref(false);
const normalizedBackupCount = computed(() => form.backupUrls.map((item) => item.trim()).filter(Boolean).length);
const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);

function cloneFormState(source: SiteSettingsForm): SiteSettingsForm {
  return { ...source, backupUrls: [...source.backupUrls] };
}

function applyFormState(source: SiteSettingsForm) {
  const next = cloneFormState(source);
  form.siteName = next.siteName;
  form.tagline = next.tagline;
  form.description = next.description;
  form.termsUrl = next.termsUrl;
  form.privacyUrl = next.privacyUrl;
  form.primaryUrl = next.primaryUrl;
  form.backupUrls = next.backupUrls;
  form.logoLight = next.logoLight;
  form.logoDark = next.logoDark;
  form.favicon = next.favicon;
  form.logo192 = next.logo192;
  form.injectionCode = next.injectionCode;
  form.mobileGuideEnabled = next.mobileGuideEnabled;
  form.mobileFeedbackUrl = next.mobileFeedbackUrl;
  form.desktopGuideEnabled = next.desktopGuideEnabled;
  form.desktopCommunityUrl = next.desktopCommunityUrl;
}

function createSnapshot(source: SiteSettingsForm): string {
  return JSON.stringify({
    ...source,
    backupUrls: source.backupUrls.map((item) => item.trim()).filter(Boolean),
  });
}

function toFormState(payload: SiteSettingsPayload): SiteSettingsForm {
  return {
    siteName: payload.site_name ?? '',
    tagline: payload.tagline ?? '',
    description: payload.description ?? '',
    termsUrl: payload.terms_url ?? '',
    privacyUrl: payload.privacy_url ?? '',
    primaryUrl: payload.primary_url ?? '',
    backupUrls: Array.isArray(payload.backup_urls) ? payload.backup_urls : [],
    logoLight: payload.logo_light ?? '',
    logoDark: payload.logo_dark ?? '',
    favicon: payload.favicon ?? '',
    logo192: payload.logo_192 ?? '',
    injectionCode: payload.injection_code ?? '',
    mobileGuideEnabled: Boolean(payload.mobile_guide_enabled),
    mobileFeedbackUrl: payload.mobile_feedback_url ?? '',
    desktopGuideEnabled: Boolean(payload.desktop_guide_enabled),
    desktopCommunityUrl: payload.desktop_community_url ?? '',
  };
}

function mergeIntoPayload(source: SiteSettingsForm, current: SiteSettingsPayload): SiteSettingsPayload {
  return {
    ...current,
    site_name: source.siteName.trim(),
    tagline: source.tagline.trim(),
    description: source.description.trim(),
    terms_url: source.termsUrl.trim(),
    privacy_url: source.privacyUrl.trim(),
    primary_url: source.primaryUrl.trim(),
    backup_urls: source.backupUrls.map((item) => item.trim()).filter(Boolean),
    logo_light: source.logoLight.trim(),
    logo_dark: source.logoDark.trim(),
    favicon: source.favicon.trim(),
    logo_192: source.logo192.trim(),
    injection_code: source.injectionCode.trim(),
    mobile_guide_enabled: source.mobileGuideEnabled,
    mobile_feedback_url: source.mobileFeedbackUrl.trim(),
    desktop_guide_enabled: source.desktopGuideEnabled,
    desktop_community_url: source.desktopCommunityUrl.trim(),
  };
}

function addBackupUrl() {
  form.backupUrls.push('');
}

function removeBackupUrl(index: number) {
  form.backupUrls.splice(index, 1);
}

async function reload() {
  loading.value = true;
  try {
    const data = await getSiteSettings();
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载站点设置失败');
  } finally {
    loading.value = false;
  }
}

function reset() {
  applyFormState(defaultFormState());
  ElMessage.success('已恢复为默认值，记得保存');
}

async function save() {
  if (!form.siteName.trim()) {
    ElMessage.warning('站点名称不能为空');
    return;
  }

  saving.value = true;
  try {
    const current = await getSiteSettings();
    const data = await updateSiteSettings(mergeIntoPayload(form, current));
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('站点信息已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存站点设置失败');
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
.site-info-panel,
.content-column,
.preview-column {
  display: grid;
  gap: 20px;
}

.layout-grid {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 20px;
}

.panel-card,
.switch-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.05);
}

.panel-card {
  padding: 24px;
  border-radius: 24px;
}

.preview-card,
.quick-card {
  display: grid;
  gap: 14px;
}

.eyebrow {
  margin: 0;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h2,
p,
strong,
span {
  margin: 0;
}

h2,
strong,
.quick-item strong,
.field span {
  color: #0f172a;
}

.brand-preview {
  position: relative;
  overflow: hidden;
  display: grid;
  gap: 6px;
  padding: 18px;
  border-radius: 20px;
}

.brand-preview::before {
  content: '';
  position: absolute;
  inset: -35% -20%;
  pointer-events: none;
  background:
    radial-gradient(circle at 18% 20%, rgba(91, 192, 235, 0.28), transparent 34%),
    radial-gradient(circle at 84% 16%, rgba(255, 182, 193, 0.42), transparent 32%),
    radial-gradient(circle at 78% 86%, rgba(146, 197, 253, 0.18), transparent 34%);
  filter: blur(18px);
}

.brand-preview strong,
.brand-preview span {
  position: relative;
  z-index: 1;
}

.brand-preview span,
.quick-item span,
.empty-hint,
.switch-card p {
  color: #64748b;
}

.brand-preview-light {
  background: linear-gradient(180deg, #fff 0%, #eff6ff 100%);
}

.brand-preview-dark {
  border: 1px solid rgba(255, 255, 255, 0.82);
  background:
    linear-gradient(135deg, rgba(240, 249, 255, 0.92) 0%, rgba(255, 255, 255, 0.86) 44%, rgba(255, 241, 246, 0.9) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 18px 38px rgba(96, 165, 250, 0.16),
    0 14px 30px rgba(244, 114, 182, 0.12);
  backdrop-filter: blur(18px);
}

.brand-preview-dark strong {
  color: #102a43;
}

.brand-preview-dark span {
  color: #627d98;
}

.quick-item {
  display: grid;
  gap: 6px;
  padding: 14px 0;
  border-bottom: 1px solid rgba(226, 232, 240, 0.9);
}

.quick-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.form-grid {
  display: grid;
  gap: 16px;
}

.two-column,
.toggle-grid {
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
.field-textarea {
  width: 100%;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 16px;
  background: #fff;
  color: #0f172a;
}

.field-input {
  min-height: 48px;
  padding: 0 14px;
}

.field-textarea {
  min-height: 120px;
  padding: 14px;
  resize: vertical;
}

.code-textarea {
  font-family: Consolas, 'Courier New', monospace;
  line-height: 1.6;
}

.field-input:focus,
.field-textarea:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.8);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}

.ghost-button,
.icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: none;
  cursor: pointer;
}

.ghost-button {
  min-height: 42px;
  padding: 0 14px;
  border-radius: 14px;
  background: rgba(239, 246, 255, 0.95);
  color: #2563eb;
}

.backup-list {
  display: grid;
  gap: 14px;
}

.backup-item {
  padding: 14px;
  border-radius: 18px;
  background: #f8fafc;
}

.backup-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 48px;
  gap: 10px;
}

.icon-button {
  width: 48px;
  border-radius: 14px;
  background: rgba(226, 232, 240, 0.95);
  color: #475569;
}

.toggle-grid {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
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

@media (max-width: 1100px) {
  .layout-grid,
  .two-column,
  .toggle-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .panel-card {
    padding: 18px;
  }
}
</style>
