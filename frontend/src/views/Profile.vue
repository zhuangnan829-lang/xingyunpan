<template>
  <section class="settings-page">
    <div class="settings-shell">
      <header class="settings-hero">
        <div>
          <p class="settings-kicker">{{ t('accountSettings') }}</p>
          <h1>{{ t('settings') }}</h1>
        </div>
        <div class="settings-orbit" aria-hidden="true">
          <span class="orbit-chip moon-chip"><Moon /></span>
          <span class="orbit-chip setting-chip"><Setting /></span>
          <span class="orbit-badge">
            <img v-if="avatarSrc" :src="avatarSrc" alt="avatar" />
            <span v-else>{{ avatarInitial }}</span>
          </span>
        </div>
      </header>

      <nav class="settings-tabs" aria-label="璁剧疆鍒嗙被">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="settings-tab"
          :class="{ active: activeTab === tab.key }"
          type="button"
          @click="activeTab = tab.key"
        >
          <component :is="tab.icon" />
          <span>{{ t(tab.labelKey) }}</span>
        </button>
      </nav>

      <div v-if="loading" class="settings-loading">
        <el-skeleton :rows="8" animated />
      </div>

      <template v-else-if="userStore.profile">
        <section v-show="activeTab === 'profile'" class="settings-panel profile-panel">
          <div class="glass-column profile-form-card">
            <label class="field-block">
              <span>{{ t('email') }}</span>
              <input :value="userStore.profile.email" type="email" disabled />
            </label>

            <label class="field-block">
              <span>{{ t('nickname') }}</span>
              <input v-model="nickname" type="text" />
              <small>{{ t('nicknameHint') }}</small>
            </label>

            <div class="profile-meta-grid">
              <div class="meta-item">
                <span>UID</span>
                <strong>{{ profileUid }}</strong>
              </div>
              <div class="meta-item">
                <span>{{ t('registeredAt') }}</span>
                <strong>{{ formatTimestamp(userStore.profile.created_at) }}</strong>
              </div>
              <div class="meta-item">
                <span>{{ t('userGroup') }}</span>
                <strong>{{ groupName }}</strong>
              </div>
              <label class="meta-item select-meta">
                <span>{{ t('homepage') }}</span>
                <select v-model="homeVisibility">
                  <option value="passwordless">{{ t('passwordlessOnly') }}</option>
                  <option value="all">{{ t('publicSharesAll') }}</option>
                  <option value="hidden">{{ t('hidden') }}</option>
                </select>
              </label>
            </div>

            <div class="settings-actions">
              <button class="primary-save-button" type="button" :disabled="profileSaving" @click="saveProfile">
                <span>{{ profileSaving ? t('saving') : t('saveProfile') }}</span>
              </button>
            </div>
          </div>

          <aside class="avatar-panel">
            <span>{{ t('avatar') }}</span>
            <div class="avatar-preview">
              <img v-if="avatarSrc" :src="avatarSrc" alt="avatar" />
              <span v-else>{{ avatarInitial }}</span>
            </div>
            <input
              ref="avatarInputRef"
              class="avatar-file-input"
              type="file"
              accept="image/png,image/jpeg,image/webp,image/gif"
              @change="handleAvatarSelected"
            />
            <button class="soft-button" type="button" :disabled="avatarUploading" @click="openAvatarPicker">
              <EditPen />
              <span>{{ avatarUploading ? t('uploading') : t('edit') }}</span>
            </button>
          </aside>
        </section>

        <section v-show="activeTab === 'preferences'" class="settings-panel preference-panel">
          <div class="preference-grid">
            <label class="field-block setting-card">
              <span>{{ t('language') }}</span>
              <select v-model="preferences.language">
                <option v-for="language in appLanguages" :key="language.code" :value="language.code">
                  {{ language.nativeLabel }}
                </option>
              </select>
              <small>{{ t('languageHint') }}</small>
            </label>

            <label class="field-block setting-card">
              <span>{{ t('timezone') }}</span>
              <select v-model="preferences.timezone">
                <option value="Asia/Shanghai">Asia/Shanghai</option>
                <option value="UTC">UTC</option>
                <option value="America/Los_Angeles">America/Los_Angeles</option>
              </select>
              <small>{{ t('timezoneHint') }}</small>
            </label>

            <div class="setting-card">
              <span class="group-title">{{ t('darkMode') }}</span>
              <div class="segmented-control">
                <button
                  v-for="mode in modes"
                  :key="mode.key"
                  :class="{ active: preferences.mode === mode.key }"
                  type="button"
                  @click="preferences.mode = mode.key"
                >
                  <component :is="mode.icon" />
                  <span>{{ mode.label }}</span>
                </button>
              </div>
            </div>

            <div class="setting-card theme-card">
              <span class="group-title">{{ t('themeColor') }}</span>
              <div class="theme-swatches">
                <button
                  v-for="theme in themes"
                  :key="theme.key"
                  class="theme-swatch"
                  :class="{ active: preferences.theme === theme.key }"
                  :style="{ background: theme.color }"
                  type="button"
                  :aria-label="theme.label"
                  @click="preferences.theme = theme.key"
                />
              </div>
            </div>

            <section class="version-card">
              <label class="check-row">
                <input v-model="preferences.keepVersions" type="checkbox" />
                <strong>{{ t('enableVersionRetention') }}</strong>
                <span>{{ t('versionRetentionHint') }}</span>
              </label>
              <label class="field-block compact-field">
                <input v-model="preferences.versionExtensions" type="text" :placeholder="t('enabledFileExtensions')" />
                <small>{{ isEnglish ? 'Press Enter to add extensions. Leave empty to enable all files.' : '按回车键添加，留空时会对所有文件启用' }}</small>
              </label>
              <label class="field-block compact-number">
                <input v-model.number="preferences.maxVersions" type="number" min="0" />
                <small>{{ isEnglish ? 'Maximum versions. 0 means unlimited.' : '最大版本数量，0 表示无限制' }}</small>
              </label>
            </section>

            <div class="setting-card">
              <span class="group-title">{{ t('viewSettings') }}</span>
              <div class="segmented-control">
                <button
                  :class="{ active: preferences.viewSync === 'server' }"
                  type="button"
                  @click="preferences.viewSync = 'server'"
                >
                  <FolderOpened />
                  <span>{{ t('syncToServer') }}</span>
                </button>
                <button
                  :class="{ active: preferences.viewSync === 'local' }"
                  type="button"
                  @click="preferences.viewSync = 'local'"
                >
                  <Close />
                  <span>{{ t('noSync') }}</span>
                </button>
              </div>
              <small>{{ isEnglish ? 'Remember each folder view and sync it to the server.' : '是否记住各目录视图设置，并同步到服务器' }}</small>
            </div>

            <div class="setting-card">
              <span class="group-title">{{ t('treeView') }}</span>
              <label class="check-row simple">
                <input v-model="preferences.expandTree" type="checkbox" />
                <strong>{{ t('expandTree') }}</strong>
              </label>
              <small>{{ t('treeViewHint') }}</small>
            </div>

            <div class="setting-card">
              <span class="group-title">{{ t('folderClickAction') }}</span>
              <div class="segmented-control">
                <button
                  :class="{ active: preferences.folderAction === 'open' }"
                  type="button"
                  @click="preferences.folderAction = 'open'"
                >
                  <FolderOpened />
                  <span>{{ t('open') }}</span>
                </button>
                <button
                  :class="{ active: preferences.folderAction === 'select' }"
                  type="button"
                  @click="preferences.folderAction = 'select'"
                >
                  <Select />
                  <span>{{ t('select') }}</span>
                </button>
              </div>
              <small>{{ isEnglish ? 'Choose what happens when clicking a folder.' : '选择点击文件夹时的行为' }}</small>
            </div>

            <div class="settings-actions preference-actions">
              <button class="primary-save-button" type="button" :disabled="preferencesSaving" @click="savePreferences">
                <span>{{ preferencesSaving ? t('saving') : t('savePreferences') }}</span>
              </button>
              <button class="soft-button" type="button" :disabled="preferencesLoading" @click="loadPreferences">
                <span>{{ preferencesLoading ? t('loading') : t('reload') }}</span>
              </button>
            </div>
          </div>
        </section>

        <section v-show="activeTab === 'security'" class="settings-panel security-panel">
          <div class="security-grid">
            <article class="security-card">
              <div class="security-icon"><Lock /></div>
              <div>
                <span class="group-title">{{ t('password') }}</span>
                <small>{{ isEnglish ? 'Update your login password regularly to keep the account safe.' : '定期更新登录密码，保护账号访问安全' }}</small>
              </div>
              <button class="soft-button" type="button" @click="showChangePasswordDialog">
                <EditPen />
                <span>{{ t('resetPassword') }}</span>
              </button>
            </article>

            <article class="security-card">
              <div class="security-icon violet"><Connection /></div>
              <div>
                <span class="group-title">{{ t('twoStep') }}</span>
                <small>{{ isEnglish ? 'Require an extra verification step when signing in.' : '启用后，登录时需要额外验证身份' }}</small>
              </div>
              <button class="soft-button" type="button">
                <Connection />
                <span>{{ t('enableTwoStep') }}</span>
              </button>
            </article>

            <article class="security-card passkey-section">
              <div class="security-icon mint"><Stamp /></div>
              <div>
                <span class="group-title">{{ t('passkey') }}</span>
                <small>{{ t('passkeyHint') }}</small>
              </div>
              <div class="passkey-card">
                <Stamp class="passkey-icon" />
                <span>{{ t('noPasskeys') }}</span>
              </div>
              <button class="soft-button" type="button">
                <Key />
                <span>{{ t('addCredential') }}</span>
              </button>
            </article>
          </div>
        </section>
      </template>

      <div v-else class="empty-panel">
        <el-empty :description="t('loadUserFailed')">
          <el-button type="primary" @click="loadProfile">{{ t('retry') }}</el-button>
        </el-empty>
      </div>
    </div>

    <ChangePasswordDialog v-model:visible="changePasswordDialogVisible" />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import {
  Close,
  Connection,
  EditPen,
  FolderOpened,
  Key,
  Lock,
  MagicStick,
  Moon,
  Select,
  Setting,
  Stamp,
  Sunny,
  User,
} from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useUserStore } from '@/stores/user';
import { formatTimestamp } from '@/utils/format';
import { isEnglish, languageOptions, normalizeLanguage, setAppLanguage, t } from '@/utils/language';
import ChangePasswordDialog from '@/components/ChangePasswordDialog/index.vue';
import type { UserPreferences } from '@/types/user';

type SettingsTab = 'profile' | 'preferences' | 'security';

const userStore = useUserStore();

const loading = ref(false);
const activeTab = ref<SettingsTab>('profile');
const nickname = ref('');
const homeVisibility = ref('passwordless');
const changePasswordDialogVisible = ref(false);
const profileSaving = ref(false);
const avatarUploading = ref(false);
const avatarInputRef = ref<HTMLInputElement | null>(null);
const preferencesLoading = ref(false);
const preferencesSaving = ref(false);

const preferences = reactive({
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  mode: 'light',
  theme: 'sky',
  keepVersions: true,
  versionExtensions: '',
  maxVersions: 10,
  viewSync: 'server',
  expandTree: true,
  folderAction: 'open',
});

const tabs = [
  { key: 'profile' as const, labelKey: 'profile' as const, icon: User },
  { key: 'preferences' as const, labelKey: 'preferences' as const, icon: MagicStick },
  { key: 'security' as const, labelKey: 'security' as const, icon: Lock },
];

const modes = computed(() => [
  { key: 'light', label: isEnglish.value ? 'Light' : '浅色', icon: Sunny },
  { key: 'system', label: isEnglish.value ? 'System' : '系统', icon: MagicStick },
  { key: 'dark', label: isEnglish.value ? 'Dark' : '黑暗', icon: Moon },
]);

const themes = [
  { key: 'sky', label: 'Aurora Blue', color: 'linear-gradient(135deg, #1287e8 0%, #68d9ff 100%)' },
  { key: 'violet', label: '绯栨灉绮夎摑', color: 'linear-gradient(135deg, #5364d8 0%, #ffb6d5 100%)' },
  { key: 'mint', label: 'Mint', color: 'linear-gradient(135deg, #13c6b4 0%, #83e8d6 100%)' },
];

const appLanguages = languageOptions;

const profileUid = computed(() => {
  const id = userStore.profile?.id ?? 0;
  return id ? id.toString(36).toUpperCase() : '2GUO';
});

const avatarInitial = computed(() => nickname.value.trim().charAt(0).toUpperCase() || '3');
const avatarSrc = computed(() => userStore.profile?.avatar_url || '');
const groupName = computed(() => (userStore.profile?.role === 'admin' ? 'Admin' : 'User'));

watch(
  () => userStore.profile?.username,
  (username) => {
    nickname.value = username || '';
  },
  { immediate: true },
);

watch(
  () => preferences.language,
  (language) => {
    preferences.language = setAppLanguage(language);
  },
);

async function loadProfile(): Promise<void> {
  loading.value = true;
  try {
    await userStore.fetchProfile();
  } catch (error) {
    ElMessage.error('\u52a0\u8f7d\u7528\u6237\u4fe1\u606f\u5931\u8d25');
    console.error('Failed to load profile:', error);
  } finally {
    loading.value = false;
  }
}

function showChangePasswordDialog(): void {
  changePasswordDialogVisible.value = true;
}

function openAvatarPicker(): void {
  if (avatarUploading.value) return;
  avatarInputRef.value?.click();
}

async function handleAvatarSelected(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  if (!file.type.startsWith('image/')) {
    ElMessage.error('\u8bf7\u9009\u62e9\u56fe\u7247\u6587\u4ef6');
    input.value = '';
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('\u5934\u50cf\u56fe\u7247\u4e0d\u80fd\u8d85\u8fc7 5MB');
    input.value = '';
    return;
  }

  avatarUploading.value = true;
  try {
    await userStore.uploadAvatar(file);
    ElMessage.success(isEnglish.value ? 'Avatar updated' : '头像已更新');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : (isEnglish.value ? 'Failed to upload avatar' : '上传头像失败'));
  } finally {
    avatarUploading.value = false;
    input.value = '';
  }
}

function toPreferencePayload(): UserPreferences {
  return {
    language: normalizeLanguage(preferences.language),
    timezone: preferences.timezone,
    mode: preferences.mode,
    theme: preferences.theme,
    keep_versions: preferences.keepVersions,
    version_extensions: preferences.versionExtensions,
    max_versions: preferences.maxVersions,
    view_sync: preferences.viewSync,
    expand_tree: preferences.expandTree,
    folder_action: preferences.folderAction,
    home_visibility: homeVisibility.value,
  };
}

function applyPreferences(payload: UserPreferences): void {
  preferences.language = setAppLanguage(payload.language || 'zh-CN');
  preferences.timezone = payload.timezone || 'Asia/Shanghai';
  preferences.mode = payload.mode || 'light';
  preferences.theme = payload.theme || 'sky';
  preferences.keepVersions = payload.keep_versions;
  preferences.versionExtensions = payload.version_extensions || '';
  preferences.maxVersions = payload.max_versions ?? 10;
  preferences.viewSync = payload.view_sync || 'server';
  preferences.expandTree = payload.expand_tree;
  preferences.folderAction = payload.folder_action || 'open';
  homeVisibility.value = payload.home_visibility || 'passwordless';
}

async function loadPreferences(): Promise<void> {
  preferencesLoading.value = true;
  try {
    await userStore.fetchPreferences();
    if (userStore.preferences) {
      applyPreferences(userStore.preferences);
    }
  } catch (error) {
    ElMessage.error('\u52a0\u8f7d\u504f\u597d\u8bbe\u7f6e\u5931\u8d25');
    console.error('Failed to load preferences:', error);
  } finally {
    preferencesLoading.value = false;
  }
}

async function saveProfile(): Promise<void> {
  profileSaving.value = true;
  try {
    await userStore.updateProfile({ username: nickname.value.trim() });
    await userStore.updatePreferences(toPreferencePayload());
    setAppLanguage(preferences.language);
    ElMessage.success(isEnglish.value ? 'Profile saved' : '个人资料已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : (isEnglish.value ? 'Failed to save profile' : '保存个人资料失败'));
  } finally {
    profileSaving.value = false;
  }
}

async function savePreferences(): Promise<void> {
  preferencesSaving.value = true;
  try {
    await userStore.updatePreferences(toPreferencePayload());
    setAppLanguage(preferences.language);
    ElMessage.success(isEnglish.value ? 'Preferences saved' : '偏好设置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : (isEnglish.value ? 'Failed to save preferences' : '保存偏好设置失败'));
  } finally {
    preferencesSaving.value = false;
  }
}

onMounted(async () => {
  if (!userStore.profile) {
    await loadProfile();
  }
  await loadPreferences();
});
</script>

<style scoped>
.settings-page {
  min-height: calc(100vh - 90px);
  padding: 0 0 22px;
}

.settings-shell {
  position: relative;
  overflow: hidden;
  width: min(100%, 1740px);
  min-height: calc(100vh - 118px);
  padding: 48px clamp(28px, 8vw, 164px) 58px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 34px;
  background:
    radial-gradient(circle at 85% 8%, rgba(104, 217, 255, 0.34), transparent 24%),
    radial-gradient(circle at 64% 2%, rgba(255, 184, 215, 0.28), transparent 23%),
    radial-gradient(circle at 18% 92%, rgba(124, 140, 255, 0.13), transparent 30%),
    linear-gradient(152deg, rgba(255, 255, 255, 0.78), rgba(246, 252, 255, 0.52) 48%, rgba(255, 247, 251, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.98),
    inset 0 -1px 0 rgba(158, 205, 236, 0.18),
    0 30px 90px rgba(86, 139, 191, 0.16);
  backdrop-filter: blur(30px) saturate(1.18);
}

.settings-shell::before,
.settings-shell::after {
  content: '';
  position: absolute;
  pointer-events: none;
}

.settings-shell::before {
  inset: 0;
  background:
    linear-gradient(115deg, rgba(255, 255, 255, 0.42), transparent 32%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.34), transparent 18%);
}

.settings-shell::after {
  right: 8%;
  top: 16%;
  width: 520px;
  height: 240px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(111, 205, 255, 0.18), rgba(255, 190, 221, 0.18));
  filter: blur(34px);
  transform: rotate(-8deg);
}

.settings-hero,
.settings-tabs,
.settings-panel,
.settings-loading,
.empty-panel {
  position: relative;
  z-index: 1;
}

.settings-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 36px;
}

.settings-kicker {
  margin: 0 0 10px;
  color: #4f88d2;
  font-size: 14px;
  font-weight: 850;
  letter-spacing: 0;
}

.settings-hero h1 {
  margin: 0;
  color: #172033;
  font-size: clamp(54px, 5vw, 76px);
  line-height: 0.98;
  font-weight: 920;
  letter-spacing: 0;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.8);
}

.settings-orbit {
  display: flex;
  gap: 18px;
  align-items: center;
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 34px;
  background:
    radial-gradient(circle at 72% 12%, rgba(101, 221, 255, 0.24), transparent 40%),
    rgba(255, 255, 255, 0.3);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 20px 54px rgba(92, 162, 219, 0.14);
  backdrop-filter: blur(22px);
}

.orbit-chip,
.orbit-badge {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  border-radius: 19px;
}

.orbit-chip {
  color: #172033;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.86), rgba(239, 249, 255, 0.58));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    inset 0 -12px 24px rgba(217, 238, 255, 0.38),
    0 18px 30px rgba(78, 135, 185, 0.12);
}

.orbit-chip svg {
  width: 25px;
  height: 25px;
}

.setting-chip {
  color: #2184e8;
  border: 1px solid rgba(122, 202, 255, 0.78);
}

.orbit-badge {
  overflow: hidden;
  border-radius: 50%;
  background: linear-gradient(135deg, #1fc9bb 0%, #38d7c9 100%);
  color: #ffffff;
  font-size: 23px;
  font-weight: 900;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 18px 32px rgba(20, 184, 166, 0.24);
}

.orbit-badge img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.settings-tabs {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 36px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(195, 214, 236, 0.68);
}

.settings-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid transparent;
  border-radius: 17px;
  background: transparent;
  color: #667286;
  font: inherit;
  font-size: 18px;
  font-weight: 780;
  cursor: pointer;
  transition: transform 0.2s ease, color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.settings-tab svg {
  width: 22px;
  height: 22px;
}

.settings-tab:hover,
.settings-tab.active {
  transform: translateY(-1px);
  color: #1379df;
  border-color: rgba(255, 255, 255, 0.76);
  background:
    radial-gradient(circle at 22% 0%, rgba(255, 255, 255, 0.92), transparent 44%),
    linear-gradient(180deg, rgba(225, 243, 255, 0.86), rgba(255, 255, 255, 0.48));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 14px 30px rgba(75, 155, 226, 0.13);
}

.settings-tab.active::after {
  content: '';
  position: absolute;
  left: 14px;
  right: 14px;
  bottom: -17px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, #1787ef, #6ed9ff, #ffbfd8);
  box-shadow: 0 0 14px rgba(67, 176, 255, 0.44);
}

.settings-panel {
  max-width: 1220px;
}

.profile-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 58px;
}

.glass-column,
.setting-card,
.version-card,
.security-card,
.settings-loading,
.empty-panel {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.56), rgba(248, 252, 255, 0.34)),
    radial-gradient(circle at 100% 0%, rgba(113, 214, 255, 0.13), transparent 35%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    inset 0 -1px 0 rgba(166, 209, 239, 0.16),
    0 18px 44px rgba(97, 132, 177, 0.1);
  backdrop-filter: blur(20px);
}

.profile-form-card {
  display: grid;
  gap: 30px;
  padding: 28px;
  border-radius: 30px;
}

.field-block,
.setting-card,
.security-card {
  display: grid;
  gap: 12px;
}

.field-block > span,
.group-title,
.avatar-panel > span {
  color: #121b2f;
  font-size: 19px;
  font-weight: 860;
}

input,
select {
  width: 100%;
  min-height: 62px;
  box-sizing: border-box;
  border: 1px solid rgba(173, 197, 223, 0.78);
  border-radius: 19px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.46));
  color: #162033;
  font: inherit;
  font-size: 17px;
  font-weight: 650;
  outline: none;
  padding: 0 22px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    inset 0 -10px 22px rgba(226, 242, 255, 0.24),
    0 13px 30px rgba(100, 133, 176, 0.08);
}

input:focus,
select:focus {
  border-color: rgba(77, 177, 249, 0.9);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 0 0 5px rgba(63, 174, 255, 0.13),
    0 18px 38px rgba(100, 133, 176, 0.1);
}

input:disabled {
  color: #8b97aa;
  background: rgba(255, 255, 255, 0.38);
}

small {
  color: #617089;
  font-size: 15px;
  line-height: 1.7;
}

.profile-meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 26px 58px;
  padding-top: 6px;
}

.meta-item {
  display: grid;
  gap: 10px;
  min-height: 78px;
  padding: 16px 18px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.34);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.78);
}

.meta-item span {
  color: #121b2f;
  font-size: 18px;
  font-weight: 860;
}

.meta-item strong {
  color: #526073;
  font-size: 18px;
  font-weight: 620;
}

.select-meta select {
  min-height: 34px;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  color: #526073;
}

.avatar-panel {
  display: grid;
  align-content: start;
  gap: 16px;
}

.avatar-preview {
  position: relative;
  display: grid;
  place-items: center;
  width: 360px;
  height: 360px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.42);
  border-radius: 26px;
  background:
    radial-gradient(circle at 28% 20%, rgba(255, 255, 255, 0.34), transparent 22%),
    radial-gradient(circle at 88% 86%, rgba(85, 230, 215, 0.54), transparent 28%),
    linear-gradient(135deg, #18c8b4 0%, #17c3b4 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.34),
    inset 0 -20px 42px rgba(7, 148, 148, 0.12),
    0 28px 54px rgba(17, 197, 178, 0.24);
}

.avatar-preview::after {
  content: '';
  position: absolute;
  inset: 16px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.16);
}

.avatar-preview img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-preview span {
  color: #ffffff;
  font-size: 138px;
  font-weight: 300;
  text-shadow: 0 12px 26px rgba(0, 117, 117, 0.12);
}

.avatar-file-input {
  display: none;
}

.soft-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: fit-content;
  min-height: 56px;
  padding: 0 22px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 19px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(245, 250, 255, 0.48));
  color: #48576b;
  font: inherit;
  font-size: 17px;
  font-weight: 780;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.94),
    0 16px 34px rgba(113, 144, 180, 0.14);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, color 0.2s ease;
}

.soft-button:hover {
  transform: translateY(-1px);
  color: #1677df;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.98),
    0 20px 40px rgba(74, 156, 226, 0.16);
}

.soft-button svg {
  width: 22px;
  height: 22px;
}

.settings-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 2px;
}

.preference-actions {
  grid-column: 1 / -1;
}

.primary-save-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 52px;
  padding: 0 22px;
  border: 1px solid rgba(66, 153, 255, 0.58);
  border-radius: 18px;
  background: linear-gradient(135deg, #237dff 0%, #1eb8ee 100%);
  color: #ffffff;
  font: inherit;
  font-size: 16px;
  font-weight: 830;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 18px 34px rgba(45, 127, 240, 0.22);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
}

.primary-save-button:hover {
  transform: translateY(-1px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.32),
    0 22px 42px rgba(45, 127, 240, 0.27);
}

.primary-save-button:disabled,
.soft-button:disabled {
  cursor: wait;
  opacity: 0.68;
  transform: none;
}

.preference-panel,
.security-panel {
  max-width: 1040px;
}

.preference-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.setting-card {
  align-content: start;
  min-height: 148px;
  padding: 22px;
  border-radius: 26px;
}

.theme-card {
  min-height: 118px;
}

.segmented-control {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  overflow: hidden;
  border: 1px solid rgba(206, 222, 239, 0.82);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.38);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 12px 24px rgba(96, 139, 182, 0.08);
}

.segmented-control button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 48px;
  padding: 0 16px;
  border: 0;
  border-right: 1px solid rgba(206, 222, 239, 0.72);
  background: transparent;
  color: #5f6b7d;
  font: inherit;
  font-size: 16px;
  font-weight: 760;
  cursor: pointer;
}

.segmented-control button:last-child {
  border-right: 0;
}

.segmented-control button.active {
  color: #147fe8;
  background:
    radial-gradient(circle at 20% 0%, rgba(255, 255, 255, 0.9), transparent 42%),
    linear-gradient(180deg, rgba(218, 240, 255, 0.96), rgba(237, 248, 255, 0.62));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.95);
}

.segmented-control svg {
  width: 20px;
  height: 20px;
}

.theme-swatches {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-swatch {
  position: relative;
  width: 46px;
  height: 46px;
  border: 1px solid rgba(255, 255, 255, 0.52);
  border-radius: 50%;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.42),
    0 14px 28px rgba(80, 120, 200, 0.2);
  cursor: pointer;
}

.theme-swatch.active {
  box-shadow:
    0 0 0 8px rgba(226, 243, 255, 0.8),
    0 15px 30px rgba(49, 133, 226, 0.24);
}

.theme-swatch.active::after {
  content: '';
  position: absolute;
  inset: 15px;
  border: 3px solid rgba(255, 255, 255, 0.92);
  border-radius: 50%;
}

.version-card {
  display: grid;
  grid-column: 1 / -1;
  gap: 14px;
  padding: 22px;
  border-radius: 26px;
}

.check-row {
  display: flex;
  align-items: center;
  gap: 14px;
  color: #111827;
}

.check-row input {
  width: 21px;
  min-height: 21px;
  accent-color: #1677df;
  box-shadow: none;
}

.check-row strong {
  font-size: 17px;
  white-space: nowrap;
}

.check-row span {
  color: #5f6b7d;
}

.check-row.simple {
  width: fit-content;
}

.compact-field input,
.compact-number input {
  min-height: 54px;
}

.compact-number {
  max-width: 240px;
}

.security-grid {
  display: grid;
  gap: 18px;
}

.security-card {
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  min-height: 116px;
  padding: 22px;
  border-radius: 28px;
}

.security-card small {
  display: block;
  margin-top: 6px;
}

.security-icon {
  display: grid;
  place-items: center;
  width: 62px;
  height: 62px;
  border-radius: 22px;
  color: #1677df;
  background:
    linear-gradient(180deg, rgba(221, 241, 255, 0.84), rgba(255, 255, 255, 0.46));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 14px 28px rgba(70, 157, 232, 0.14);
}

.security-icon.violet {
  color: #6671d9;
  background: linear-gradient(180deg, rgba(232, 233, 255, 0.9), rgba(255, 255, 255, 0.48));
}

.security-icon.mint {
  color: #14b8a6;
  background: linear-gradient(180deg, rgba(216, 255, 248, 0.9), rgba(255, 255, 255, 0.48));
}

.security-icon svg {
  width: 28px;
  height: 28px;
}

.passkey-section {
  grid-template-columns: auto minmax(0, 1fr);
}

.passkey-section .soft-button {
  grid-column: 2;
}

.passkey-card {
  grid-column: 1 / -1;
  display: grid;
  place-items: center;
  gap: 8px;
  min-height: 126px;
  border: 1px solid rgba(205, 221, 238, 0.8);
  border-radius: 24px;
  background:
    radial-gradient(circle at 50% 8%, rgba(255, 255, 255, 0.9), transparent 28%),
    rgba(255, 255, 255, 0.38);
  color: #607089;
  text-align: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.passkey-icon {
  width: 48px;
  height: 48px;
  color: #667085;
}

.settings-loading,
.empty-panel {
  max-width: 940px;
  padding: 36px;
  border-radius: 24px;
}

/* Compact desktop density: keep the glass look without making settings feel oversized. */
.settings-shell {
  padding: 34px clamp(24px, 6vw, 128px) 42px;
  border-radius: 30px;
}

.settings-hero {
  margin-bottom: 26px;
}

.settings-kicker {
  margin-bottom: 8px;
  font-size: 13px;
}

.settings-hero h1 {
  font-size: clamp(44px, 4vw, 58px);
}

.settings-orbit {
  gap: 14px;
  padding: 12px;
  border-radius: 28px;
}

.orbit-chip,
.orbit-badge {
  width: 46px;
  height: 46px;
}

.orbit-chip {
  border-radius: 16px;
}

.orbit-chip svg {
  width: 21px;
  height: 21px;
}

.orbit-badge {
  font-size: 20px;
}

.settings-tabs {
  gap: 18px;
  margin-bottom: 28px;
  padding-bottom: 12px;
}

.settings-tab {
  min-height: 40px;
  padding: 0 12px;
  border-radius: 14px;
  font-size: 16px;
}

.settings-tab svg {
  width: 19px;
  height: 19px;
}

.settings-tab.active::after {
  bottom: -13px;
}

.profile-panel {
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 40px;
}

.profile-form-card {
  gap: 22px;
  padding: 22px;
  border-radius: 24px;
}

.field-block > span,
.group-title,
.avatar-panel > span {
  font-size: 17px;
}

input,
select {
  min-height: 52px;
  border-radius: 16px;
  padding: 0 18px;
  font-size: 15px;
}

small {
  font-size: 14px;
}

.profile-meta-grid {
  gap: 18px 30px;
}

.meta-item {
  min-height: 62px;
  padding: 13px 15px;
  border-radius: 18px;
}

.meta-item span,
.meta-item strong {
  font-size: 16px;
}

.avatar-preview {
  width: 300px;
  height: 300px;
  border-radius: 22px;
}

.avatar-preview span {
  font-size: 108px;
}

.soft-button {
  min-height: 48px;
  padding: 0 18px;
  border-radius: 16px;
  font-size: 15px;
}

.soft-button svg {
  width: 19px;
  height: 19px;
}

.primary-save-button {
  min-height: 46px;
  padding: 0 18px;
  border-radius: 16px;
  font-size: 15px;
}

.preference-grid,
.security-grid {
  gap: 14px;
}

.setting-card {
  min-height: 126px;
  padding: 18px;
  border-radius: 22px;
}

.theme-card {
  min-height: 102px;
}

.segmented-control button {
  min-height: 42px;
  padding: 0 14px;
  font-size: 14px;
}

.theme-swatch {
  width: 38px;
  height: 38px;
}

.theme-swatch.active::after {
  inset: 12px;
}

.version-card {
  gap: 12px;
  padding: 18px;
  border-radius: 22px;
}

.compact-field input,
.compact-number input {
  min-height: 48px;
}

.security-card {
  min-height: 96px;
  padding: 18px;
  border-radius: 22px;
}

.security-icon {
  width: 52px;
  height: 52px;
  border-radius: 18px;
}

.security-icon svg {
  width: 24px;
  height: 24px;
}

.passkey-card {
  min-height: 100px;
  border-radius: 20px;
}

.passkey-icon {
  width: 40px;
  height: 40px;
}

@media (max-width: 1180px) {
  .profile-panel {
    grid-template-columns: 1fr;
  }

  .avatar-preview {
    width: min(300px, 100%);
    aspect-ratio: 1;
    height: auto;
  }
}

@media (max-width: 980px) {
  .settings-shell {
    padding: 30px 22px 42px;
  }

  .preference-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .settings-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .settings-tabs {
    gap: 8px;
    overflow-x: auto;
  }

  .profile-meta-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .settings-orbit {
    display: none;
  }

  .segmented-control {
    width: 100%;
  }

  .segmented-control button {
    flex: 1;
    justify-content: center;
    padding: 0 10px;
  }

  .check-row,
  .security-card,
  .passkey-section {
    grid-template-columns: 1fr;
    align-items: flex-start;
  }

  .check-row {
    flex-wrap: wrap;
  }

  .passkey-section .soft-button {
    grid-column: auto;
  }
}
</style>
