<template>
  <div class="user-session-panel">
    <section class="panel-card">
      <div class="section-head">
        <div>
          <p class="eyebrow">Registration and Login</p>
          <h2>注册与登录</h2>
        </div>
        <span class="section-badge">4 项设置</span>
      </div>

      <div class="switch-list">
        <button type="button" class="switch-row" @click="toggleField('allowRegistration')">
          <span class="switch-control" :class="{ 'is-on': form.allowRegistration }" aria-hidden="true">
            <span class="switch-thumb"></span>
          </span>
          <span class="switch-copy">
            <strong>允许新用户注册</strong>
            <span>关闭后，无法再通过前台注册新的用户。</span>
          </span>
        </button>

        <button type="button" class="switch-row" @click="toggleField('emailActivation')">
          <span class="switch-control" :class="{ 'is-on': form.emailActivation }" aria-hidden="true">
            <span class="switch-thumb"></span>
          </span>
          <span class="switch-copy">
            <strong>邮件激活</strong>
            <span>
              开启后，新用户注册需要点击邮件中的激活链接才能完成。需确认
              <span class="inline-link">邮件发信设置</span>
              是否正确，否则激活邮件无法送达。
            </span>
          </span>
        </button>

        <button type="button" class="switch-row" @click="toggleField('passkeyLogin')">
          <span class="switch-control" :class="{ 'is-on': form.passkeyLogin }" aria-hidden="true">
            <span class="switch-thumb"></span>
          </span>
          <span class="switch-copy">
            <strong>使用通行密钥登录</strong>
            <span>
              是否允许用户使用绑定的硬件认证设备登录，比如：人脸、指纹或 USB 密钥；站点必须启用 HTTPS
              才能使用。
            </span>
          </span>
        </button>
      </div>

      <div class="setting-item">
        <div class="setting-copy">
          <strong>默认用户组</strong>
          <p>用户注册后的初始用户组。</p>
        </div>
        <label class="select-shell">
          <select v-model="form.defaultGroup" class="select-input">
            <option v-for="option in groupOptions" :key="option" :value="option">
              {{ option }}
            </option>
          </select>
        </label>
      </div>
    </section>

    <section class="panel-card">
      <div class="section-head">
        <div>
          <p class="eyebrow">Avatar</p>
          <h2>头像</h2>
        </div>
        <span class="section-badge">点击条目即可修改</span>
      </div>

      <div class="avatar-edit-list">
        <article
          v-for="item in avatarItems"
          :key="item.key"
          class="avatar-edit-card"
          :class="{ 'is-active': activeAvatarField === item.key }"
        >
          <button type="button" class="avatar-trigger" @click="toggleAvatarField(item.key)">
            <span class="avatar-trigger-copy">
              <strong>{{ item.label }}</strong>
              <span>{{ item.description }}</span>
            </span>
            <span class="avatar-trigger-meta">
              <em>{{ item.summary }}</em>
              <span class="avatar-trigger-action">{{ activeAvatarField === item.key ? '收起' : '修改' }}</span>
            </span>
          </button>

          <div v-if="activeAvatarField === item.key" class="avatar-editor">
            <label class="editor-field">
              <span>{{ item.label }}</span>

              <input
                v-if="item.key === 'avatarPath'"
                v-model="form.avatarPath"
                class="text-input"
                type="text"
                autocomplete="off"
              />

              <div v-else-if="item.key === 'avatarSizeLimitMb'" class="input-with-unit input-with-select">
                <input
                  v-model.number="avatarSizeDisplayValue"
                  class="text-input"
                  type="number"
                  min="0"
                  step="1"
                />
                <select v-model="avatarSizeUnit" class="unit-select">
                  <option v-for="unit in sizeUnitOptions" :key="unit" :value="unit">
                    {{ unit }}
                  </option>
                </select>
              </div>

              <div v-else-if="item.key === 'avatarDimension'" class="input-with-unit">
                <input
                  v-model.number="form.avatarDimension"
                  class="text-input"
                  type="number"
                  min="64"
                  step="1"
                />
                <span class="input-unit">px</span>
              </div>

              <input
                v-else
                v-model="form.gravatarServer"
                class="text-input"
                type="text"
                autocomplete="off"
              />
            </label>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { getSiteSettings, updateSiteSettings, type SiteSettingsPayload } from '@/api/site-settings';

type UserSessionForm = {
  allowRegistration: boolean;
  emailActivation: boolean;
  passkeyLogin: boolean;
  defaultGroup: string;
  avatarPath: string;
  avatarSizeLimitMb: number;
  avatarDimension: number;
  gravatarServer: string;
};

type SwitchFieldKey = 'allowRegistration' | 'emailActivation' | 'passkeyLogin';
type AvatarFieldKey = 'avatarPath' | 'avatarSizeLimitMb' | 'avatarDimension' | 'gravatarServer';
type SizeUnit = 'B' | 'KB' | 'MB' | 'GB' | 'TB';

const groupOptions = ['User', 'Admin', 'Guest'];
const sizeUnitOptions: SizeUnit[] = ['B', 'KB', 'MB', 'GB', 'TB'];
const sizeUnitToMbMap: Record<SizeUnit, number> = {
  B: 1 / (1024 * 1024),
  KB: 1 / 1024,
  MB: 1,
  GB: 1024,
  TB: 1024 * 1024,
};

const defaultFormState = (): UserSessionForm => ({
  allowRegistration: true,
  emailActivation: false,
  passkeyLogin: false,
  defaultGroup: 'User',
  avatarPath: 'avatar',
  avatarSizeLimitMb: 4,
  avatarDimension: 200,
  gravatarServer: 'https://www.gravatar.com/',
});

const form = reactive<UserSessionForm>(defaultFormState());
const loading = ref(false);
const saving = ref(false);
const lastSavedSnapshot = ref('');
const activeAvatarField = ref<AvatarFieldKey | null>('avatarPath');
const avatarSizeUnit = ref<SizeUnit>('MB');

const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);
const avatarSizeDisplayValue = computed({
  get() {
    const factor = sizeUnitToMbMap[avatarSizeUnit.value];
    const value = form.avatarSizeLimitMb / factor;
    return Number.isInteger(value) ? value : Number(value.toFixed(2));
  },
  set(value: number | string) {
    const numericValue = Number(value);
    if (!Number.isFinite(numericValue) || numericValue < 0) {
      form.avatarSizeLimitMb = 0;
      return;
    }

    form.avatarSizeLimitMb = Number((numericValue * sizeUnitToMbMap[avatarSizeUnit.value]).toFixed(4));
  },
});

const avatarItems = computed(() => [
  {
    key: 'avatarPath' as const,
    label: '头像存储路径',
    description: '用户上传自定义头像的存储路径，相对于星云盘数据目录。',
    summary: form.avatarPath || '未设置',
  },
  {
    key: 'avatarSizeLimitMb' as const,
    label: '头像文件大小限制',
    description: '用户可上传头像文件的最大大小。',
    summary: `${form.avatarSizeLimitMb} MB`,
  },
  {
    key: 'avatarDimension' as const,
    label: '图像尺寸 (px)',
    description: '用户所上传头像会被裁剪到指定的尺寸，单位为像素。',
    summary: `${form.avatarDimension} px`,
  },
  {
    key: 'gravatarServer' as const,
    label: 'Gravatar 服务器',
    description: 'Gravatar 服务器地址，可选择使用国内镜像。',
    summary: form.gravatarServer || '未设置',
  },
]);

function createSnapshot(source: UserSessionForm): string {
  return JSON.stringify({
    ...source,
    defaultGroup: source.defaultGroup.trim(),
    avatarPath: source.avatarPath.trim(),
    gravatarServer: source.gravatarServer.trim(),
  });
}

function cloneFormState(source: UserSessionForm): UserSessionForm {
  return {
    allowRegistration: source.allowRegistration,
    emailActivation: source.emailActivation,
    passkeyLogin: source.passkeyLogin,
    defaultGroup: source.defaultGroup,
    avatarPath: source.avatarPath,
    avatarSizeLimitMb: source.avatarSizeLimitMb,
    avatarDimension: source.avatarDimension,
    gravatarServer: source.gravatarServer,
  };
}

function applyFormState(source: UserSessionForm) {
  const next = cloneFormState(source);
  form.allowRegistration = next.allowRegistration;
  form.emailActivation = next.emailActivation;
  form.passkeyLogin = next.passkeyLogin;
  form.defaultGroup = next.defaultGroup;
  form.avatarPath = next.avatarPath;
  form.avatarSizeLimitMb = next.avatarSizeLimitMb;
  form.avatarDimension = next.avatarDimension;
  form.gravatarServer = next.gravatarServer;
}

function toFormState(payload: SiteSettingsPayload): UserSessionForm {
  return {
    allowRegistration: Boolean(payload.allow_registration),
    emailActivation: Boolean(payload.email_activation),
    passkeyLogin: Boolean(payload.passkey_login_enabled),
    defaultGroup: payload.default_group ?? 'User',
    avatarPath: payload.avatar_path ?? 'avatar',
    avatarSizeLimitMb: payload.avatar_size_limit_mb ?? 4,
    avatarDimension: payload.avatar_dimension ?? 200,
    gravatarServer: payload.gravatar_server ?? 'https://www.gravatar.com/',
  };
}

function mergeIntoPayload(source: UserSessionForm, current: SiteSettingsPayload): SiteSettingsPayload {
  return {
    ...current,
    allow_registration: source.allowRegistration,
    email_activation: source.emailActivation,
    passkey_login_enabled: source.passkeyLogin,
    default_group: source.defaultGroup.trim(),
    avatar_path: source.avatarPath.trim(),
    avatar_size_limit_mb: Number(source.avatarSizeLimitMb),
    avatar_dimension: Number(source.avatarDimension),
    gravatar_server: source.gravatarServer.trim(),
  };
}

function toggleField(key: SwitchFieldKey) {
  form[key] = !form[key];
}

function toggleAvatarField(key: AvatarFieldKey) {
  activeAvatarField.value = activeAvatarField.value === key ? null : key;
}

async function reload() {
  loading.value = true;
  try {
    const data = await getSiteSettings();
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载用户会话设置失败');
  } finally {
    loading.value = false;
  }
}

async function reset() {
  applyFormState(defaultFormState());
  activeAvatarField.value = 'avatarPath';
  ElMessage.success('已恢复为默认值，记得保存');
}

async function save() {
  if (!form.defaultGroup.trim()) {
    ElMessage.warning('默认用户组不能为空');
    return;
  }

  if (!form.avatarPath.trim()) {
    ElMessage.warning('头像存储路径不能为空');
    return;
  }

  if (!form.gravatarServer.trim()) {
    ElMessage.warning('Gravatar 服务器不能为空');
    return;
  }

  if (Number(form.avatarSizeLimitMb) <= 0) {
    ElMessage.warning('头像文件大小限制必须大于 0');
    return;
  }

  if (Number(form.avatarDimension) <= 0) {
    ElMessage.warning('图像尺寸必须大于 0');
    return;
  }

  saving.value = true;
  try {
    const current = await getSiteSettings();
    const data = await updateSiteSettings(mergeIntoPayload(form, current));
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('用户会话设置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存用户会话设置失败');
  } finally {
    saving.value = false;
  }
}

watch(avatarSizeUnit, (nextUnit, previousUnit) => {
  if (!previousUnit) {
    return;
  }

  const valueInMb = form.avatarSizeLimitMb;
  const nextDisplay = valueInMb / sizeUnitToMbMap[nextUnit];
  avatarSizeDisplayValue.value = Number.isInteger(nextDisplay) ? nextDisplay : Number(nextDisplay.toFixed(2));
});

defineExpose({ isDirty, loading, saving, reload, reset, save });

onMounted(async () => {
  await reload();
});
</script>

<style scoped>
.user-session-panel {
  display: grid;
  gap: 20px;
}

.panel-card {
  padding: 24px;
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.05);
}

.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 22px;
}

.eyebrow,
h2,
strong,
p,
span,
em {
  margin: 0;
}

.eyebrow {
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h2,
.switch-copy strong,
.setting-copy strong,
.avatar-trigger-copy strong {
  color: #0f172a;
}

h2 {
  margin-top: 6px;
  font-size: 28px;
  line-height: 1.2;
}

.section-badge {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(239, 246, 255, 0.95);
  color: #2563eb;
  font-size: 13px;
  font-weight: 700;
}

.switch-list,
.avatar-edit-list {
  display: grid;
  gap: 16px;
}

.switch-row,
.avatar-trigger {
  width: 100%;
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  text-align: left;
}

.switch-row {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.switch-control {
  position: relative;
  display: inline-flex;
  align-items: center;
  width: 44px;
  height: 24px;
  padding: 2px;
  margin-top: 4px;
  border-radius: 999px;
  background: #d1d5db;
  transition: background-color 0.2s ease;
}

.switch-control.is-on {
  background: #60a5fa;
}

.switch-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.18);
  transition: transform 0.2s ease;
}

.switch-control.is-on .switch-thumb {
  transform: translateX(20px);
}

.switch-copy,
.setting-copy,
.avatar-trigger-copy {
  display: grid;
  gap: 8px;
}

.switch-copy strong,
.setting-copy strong,
.avatar-trigger-copy strong {
  font-size: 16px;
  font-weight: 700;
}

.switch-copy span,
.setting-copy p,
.avatar-trigger-copy span,
.avatar-trigger-action,
.avatar-trigger-meta em {
  color: #64748b;
  line-height: 1.75;
  font-size: 14px;
}

.inline-link {
  color: #2563eb;
  font-weight: 600;
}

.setting-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 240px;
  gap: 18px;
  align-items: center;
  padding-top: 20px;
  margin-top: 20px;
  border-top: 1px solid rgba(226, 232, 240, 0.92);
}

.select-shell,
.input-with-unit {
  position: relative;
  display: block;
}

.input-with-select .text-input {
  padding-right: 92px;
}

.select-input,
.text-input,
.unit-select {
  width: 100%;
  min-height: 48px;
  padding: 0 14px;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 14px;
  background: #fff;
  color: #0f172a;
  font: inherit;
}

.select-input {
  appearance: none;
  padding-right: 38px;
  background-image:
    linear-gradient(45deg, transparent 50%, #64748b 50%),
    linear-gradient(135deg, #64748b 50%, transparent 50%);
  background-position:
    calc(100% - 20px) calc(50% - 2px),
    calc(100% - 14px) calc(50% - 2px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
}

.text-input:focus,
.select-input:focus,
.unit-select:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.8);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}

.input-unit {
  position: absolute;
  top: 50%;
  right: 14px;
  transform: translateY(-50%);
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.unit-select {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 82px;
  min-height: 40px;
  padding: 0 28px 0 14px;
  border: none;
  border-left: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 12px;
  background-color: #f8fafc;
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, #64748b 50%),
    linear-gradient(135deg, #64748b 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 2px),
    calc(100% - 12px) calc(50% - 2px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
}

.avatar-edit-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 20px;
  background: linear-gradient(180deg, #fff 0%, #f8fbff 100%);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.avatar-edit-card.is-active {
  border-color: rgba(96, 165, 250, 0.75);
  box-shadow: 0 12px 26px rgba(37, 99, 235, 0.08);
}

.avatar-trigger {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 20px;
}

.avatar-trigger-meta {
  display: grid;
  justify-items: end;
  gap: 6px;
  min-width: 140px;
}

.avatar-trigger-meta em {
  max-width: 280px;
  font-style: normal;
  color: #0f172a;
  line-height: 1.5;
  text-align: right;
  word-break: break-all;
}

.avatar-trigger-action {
  color: #2563eb;
  font-weight: 700;
}

.avatar-editor {
  padding: 0 20px 20px;
}

.editor-field {
  display: grid;
  gap: 10px;
}

.editor-field > span {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

@media (max-width: 900px) {
  .section-head {
    flex-direction: column;
  }

  .setting-item {
    grid-template-columns: 1fr;
  }

  .avatar-trigger {
    flex-direction: column;
  }

  .avatar-trigger-meta {
    justify-items: start;
    min-width: 0;
  }

  .avatar-trigger-meta em {
    max-width: none;
    text-align: left;
  }
}

@media (max-width: 720px) {
  .panel-card {
    padding: 18px;
    border-radius: 20px;
  }

  h2 {
    font-size: 24px;
  }
}
</style>
