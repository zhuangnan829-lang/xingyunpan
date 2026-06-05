<template>
  <div class="console-layout" :class="{ 'is-admin-layout': isAdminArea }">
    <aside class="sidebar" :class="{ 'is-admin-sidebar': isAdminArea }">
      <div class="brand-panel">
        <router-link class="brand-lockup" to="/drive/my-files">
          <div class="brand-mark" aria-hidden="true">
            <span class="brand-core"></span>
            <span class="brand-ring"></span>
          </div>
          <div class="brand-copy">
            <strong>星云盘</strong>
            <span>{{ isAdminArea ? t('adminSubtitle') : t('brandSubtitle') }}</span>
          </div>
        </router-link>
      </div>

      <template v-if="!isAdminArea">
        <nav class="menu-group drive-menu-group" aria-label="主导航">
          <router-link
            v-for="item in driveLinks"
            :key="item.to"
            :to="item.to"
            class="menu-link"
            :class="{ 'is-active': isActive(item.to) }"
          >
            <component :is="item.icon" class="menu-icon" />
            <span>{{ t(item.labelKey) }}</span>
          </router-link>
        </nav>

        <div class="sidebar-footer">
          <div class="storage-card">
            <div class="storage-card-head">
              <strong>{{ t('storage') }}</strong>
              <span>{{ userStore.storageUsagePercent }}%</span>
            </div>
            <div class="storage-bar">
              <span :style="{ width: `${userStore.storageUsagePercent}%` }"></span>
            </div>
            <p>{{ formatSize(userStore.profile?.used_size || 0) }} / {{ formatSize(userStore.profile?.capacity || 0) }}</p>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="sidebar-scroll admin-sidebar-scroll">
          <nav class="menu-group admin-menu-group" aria-label="管理导航">
            <router-link
              v-for="item in adminLinks"
              :key="item.to"
              :to="item.to"
              class="menu-link"
              :class="{ 'is-active': isActive(item.to) }"
            >
              <component :is="item.icon" class="menu-icon" />
            <span>{{ t(item.labelKey) }}</span>
            </router-link>
          </nav>
        </div>
      </template>
    </aside>

    <div class="content-shell">
      <header v-if="!isAdminArea" class="topbar">
        <div class="create-action-wrap">
          <button class="primary-action" type="button" @click.stop="toggleCreateMenu">
            <Plus class="button-icon" />
          <span>{{ t('new') }}</span>
          </button>

          <div v-if="createMenuVisible" class="create-menu" role="menu" @click.stop>
            <button class="create-menu-item" type="button" role="menuitem" @click="triggerFileUpload">
              <Upload class="menu-action-icon" />
              <span>{{ t('uploadFiles') }}</span>
            </button>
            <button class="create-menu-item" type="button" role="menuitem" @click="openCreateFolderDialog">
              <FolderOpened class="menu-action-icon" />
              <span>{{ t('newFolder') }}</span>
            </button>
          </div>

          <input
            ref="fileInputRef"
            class="hidden-file-input"
            type="file"
            multiple
            @change="handleFileSelect"
          />
        </div>

        <form class="search-shell" role="search" @submit.prevent="runTopbarSearch">
          <button class="search-submit" type="submit" :disabled="searchStore.isSearching" aria-label="搜索">
            <Search class="search-icon" />
          </button>
          <input
            ref="searchInputRef"
            v-model="topbarKeyword"
            type="search"
            :placeholder="t('searchPlaceholder')"
            :disabled="searchStore.isSearching"
            @input="handleSearchInput"
          />
          <button v-if="topbarKeyword" class="search-clear" type="button" aria-label="清除搜索" @click="clearTopbarSearch">
            ×
          </button>
          <kbd v-else>/</kbd>
        </form>

        <div class="toolbar-actions">
          <button class="icon-button" type="button" :aria-label="t('darkMode')">
            <Moon class="toolbar-icon" />
          </button>
          <button
            class="icon-button"
            :class="{ 'is-settings-active': route.path === '/profile' }"
            type="button"
            :aria-label="t('settings')"
            @click="navigateToSettings"
          >
            <Setting class="toolbar-icon" />
          </button>
          <span class="counter-badge">
            <img v-if="avatarSrc" :src="avatarSrc" alt="avatar" />
            <span v-else>{{ avatarInitial }}</span>
          </span>
        </div>
      </header>

      <main class="content-main">
        <slot />
      </main>
    </div>

    <UploadDialog :folder-id="uploadTargetFolderId" />
    <CreateFolderDialog v-model:visible="createFolderVisible" @confirm="handleCreateFolder" />
  </div>
</template>

<script setup lang="ts">
import {
  Connection,
  DataAnalysis,
  Delete,
  Document,
  Files,
  FolderOpened,
  Goods,
  Headset,
  House,
  Monitor,
  Moon,
  Picture,
  Plus,
  Search,
  Setting,
  Share,
  Upload,
  User,
  UserFilled,
  VideoPlay,
} from '@element-plus/icons-vue';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import UploadDialog from '@/components/UploadDialog/index.vue';
import CreateFolderDialog from '@/components/CreateFolderDialog/index.vue';
import { useUserStore } from '@/stores/user';
import { useFileStore } from '@/stores/file';
import { useSearchStore } from '@/stores/search';
import { useUploadStore } from '@/stores/upload';
import { normalizeLanguage, setAppLanguage, t, type MessageKey } from '@/utils/language';

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const fileStore = useFileStore();
const searchStore = useSearchStore();
const uploadStore = useUploadStore();

const createMenuVisible = ref(false);
const createFolderVisible = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);
const searchInputRef = ref<HTMLInputElement | null>(null);
const topbarKeyword = ref('');
const lastCompletedUploadCount = ref(0);

const isAdminArea = computed(() => route.path.startsWith('/admin'));
const uploadTargetFolderId = computed(() => fileStore.currentFolderId ?? undefined);
const avatarSrc = computed(() => userStore.profile?.avatar_url || '');
const avatarInitial = computed(() => userStore.profile?.username?.trim().charAt(0).toUpperCase() || '3');

const driveLinks = [
  { labelKey: 'myFiles' as MessageKey, to: '/drive/my-files', icon: House },
  { labelKey: 'images' as MessageKey, to: '/drive/images', icon: Picture },
  { labelKey: 'videos' as MessageKey, to: '/drive/videos', icon: VideoPlay },
  { labelKey: 'music' as MessageKey, to: '/drive/music', icon: Headset },
  { labelKey: 'documents' as MessageKey, to: '/drive/documents', icon: Document },
  { labelKey: 'sharedWithMe' as MessageKey, to: '/drive/shared-with-me', icon: Share },
  { labelKey: 'recycle' as MessageKey, to: '/recycle', icon: Delete },
  { labelKey: 'myShares' as MessageKey, to: '/drive/my-shares', icon: Share },
  { labelKey: 'connections' as MessageKey, to: '/drive/transfer-tasks', icon: Connection },
  { labelKey: 'backgroundJobs' as MessageKey, to: '/drive/background-jobs', icon: Monitor },
  { labelKey: 'offlineDownloads' as MessageKey, to: '/drive/offline-downloads', icon: Upload },
  { labelKey: 'adminPanel' as MessageKey, to: '/admin/dashboard', icon: Setting },
];

const adminLinks = [
  { labelKey: 'dashboard' as MessageKey, to: '/admin/dashboard', icon: DataAnalysis },
  { labelKey: 'parameterSettings' as MessageKey, to: '/admin/settings', icon: Setting },
  { labelKey: 'fileSystem' as MessageKey, to: '/admin/file-system', icon: Connection },
  { labelKey: 'storagePolicy' as MessageKey, to: '/admin/storage-policy', icon: Monitor },
  { labelKey: 'nodes' as MessageKey, to: '/admin/nodes', icon: Monitor },
  { labelKey: 'userGroups' as MessageKey, to: '/admin/user-groups', icon: UserFilled },
  { labelKey: 'users' as MessageKey, to: '/admin/users', icon: User },
  { labelKey: 'files' as MessageKey, to: '/admin/files', icon: FolderOpened },
  { labelKey: 'fileBlobs' as MessageKey, to: '/admin/blobs', icon: Files },
  { labelKey: 'shares' as MessageKey, to: '/admin/shares', icon: Share },
  { labelKey: 'backgroundJobs' as MessageKey, to: '/admin/tasks', icon: Goods },
  { labelKey: 'oauthApps' as MessageKey, to: '/admin/oauth', icon: Connection },
  { labelKey: 'backHome' as MessageKey, to: '/drive/my-files', icon: House },
];

const isActive = (path: string): boolean => route.path === path || route.path.startsWith(`${path}/`);

const formatSize = (size: number): string => {
  if (!size) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
};

const closeCreateMenu = () => {
  createMenuVisible.value = false;
};

const toggleCreateMenu = () => {
  createMenuVisible.value = !createMenuVisible.value;
};

const triggerFileUpload = () => {
  closeCreateMenu();
  fileInputRef.value?.click();
};

const openCreateFolderDialog = () => {
  closeCreateMenu();
  createFolderVisible.value = true;
};

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const files = Array.from(target.files || []);
  target.value = '';

  if (!files.length) return;

  try {
    for (const file of files) {
      await uploadStore.addTask(file, uploadTargetFolderId.value);
    }
    ElMessage.success(`已添加 ${files.length} 个文件到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '添加上传任务失败');
  }
};

const handleCreateFolder = async (name: string) => {
  try {
    await fileStore.createFolder(name);
    createFolderVisible.value = false;
    ElMessage.success('文件夹已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件夹失败');
  }
};

const handleSearchInput = () => {
  if (topbarKeyword.value.trim()) return;
  if (!fileStore.isSearchResult && !fileStore.searchKeyword) return;
  clearTopbarSearch();
};

const runTopbarSearch = async () => {
  const keyword = topbarKeyword.value.trim();
  if (!keyword) {
    await clearTopbarSearch();
    return;
  }

  try {
    await searchStore.search(keyword, undefined, 1);
    fileStore.applySearchResults(
      searchStore.searchResults,
      searchStore.folderResults,
      searchStore.totalResults,
      keyword,
    );
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '搜索失败，请稍后重试');
  }
};

const navigateToSettings = () => {
  router.push('/profile');
};

const clearTopbarSearch = async () => {
  topbarKeyword.value = '';
  searchStore.clearSearch();
  fileStore.clearSearchResults();
  if (!isAdminArea.value) {
    await fileStore.refresh();
  }
};

const handleGlobalSearchShortcut = async (event: KeyboardEvent) => {
  const target = event.target as HTMLElement | null;
  const isTypingTarget = target?.closest('input, textarea, [contenteditable="true"]');

  if (event.key === '/' && !isTypingTarget && !isAdminArea.value) {
    event.preventDefault();
    await nextTick();
    searchInputRef.value?.focus();
  }

  if (event.key === 'Escape' && document.activeElement === searchInputRef.value) {
    searchInputRef.value?.blur();
  }
};

onMounted(async () => {
  document.addEventListener('click', closeCreateMenu);
  document.addEventListener('keydown', handleGlobalSearchShortcut);

  if (!userStore.profile && userStore.isAuthenticated) {
    try {
      await userStore.fetchProfile();
    } catch {
      // keep layout renderable
    }
  }

  if (userStore.isAuthenticated) {
    try {
      await userStore.fetchPreferences();
      setAppLanguage(normalizeLanguage(userStore.preferences?.language));
    } catch {
      // keep the locally stored language when preferences are unavailable
    }
  }
});

onBeforeUnmount(() => {
  document.removeEventListener('click', closeCreateMenu);
  document.removeEventListener('keydown', handleGlobalSearchShortcut);
});

watch(
  () => uploadStore.completedTasks.filter((task) => task.status === 'completed').length,
  async (completedCount) => {
    if (completedCount <= lastCompletedUploadCount.value) {
      lastCompletedUploadCount.value = completedCount;
      return;
    }

    lastCompletedUploadCount.value = completedCount;
    if (!isAdminArea.value) {
      await fileStore.refresh();
    }
  },
);
</script>

<style scoped>
.console-layout {
  display: grid;
  grid-template-columns: 292px minmax(0, 1fr);
  height: 100vh;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 10%, rgba(111, 205, 255, 0.26), transparent 26%),
    radial-gradient(circle at 72% 8%, rgba(255, 198, 221, 0.2), transparent 24%),
    radial-gradient(circle at 88% 82%, rgba(96, 165, 250, 0.16), transparent 30%),
    linear-gradient(135deg, #f9fcff 0%, #eef6ff 46%, #fff7fb 100%);
}

.console-layout.is-admin-layout {
  grid-template-columns: 318px minmax(0, 1fr);
  align-items: start;
  background:
    radial-gradient(circle at 18% 10%, rgba(120, 218, 255, 0.28), transparent 28%),
    radial-gradient(circle at 72% 8%, rgba(255, 205, 226, 0.24), transparent 26%),
    radial-gradient(circle at 86% 78%, rgba(255, 226, 136, 0.18), transparent 32%),
    linear-gradient(135deg, #f8fcff 0%, #eef7ff 48%, #fff5f8 100%);
}

.sidebar {
  position: sticky;
  top: 0;
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 18px 18px 20px;
  height: 100vh;
  max-height: 100vh;
  min-height: 100vh;
  box-sizing: border-box;
  border-right: 1px solid rgba(255, 255, 255, 0.78);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.8), rgba(246, 251, 255, 0.64)),
    radial-gradient(circle at 24% 6%, rgba(109, 213, 255, 0.22), transparent 28%);
  box-shadow:
    inset -1px 0 0 rgba(203, 222, 245, 0.42),
    20px 0 50px rgba(91, 143, 190, 0.08);
  backdrop-filter: blur(22px);
  overflow: hidden;
}

.sidebar.is-admin-sidebar {
  position: sticky;
  top: 0;
  isolation: isolate;
  align-self: start;
  gap: 12px;
  margin: 0;
  height: 100vh;
  min-height: 0;
  max-height: 100vh;
  padding: 26px 22px 24px;
  border: 0;
  border-radius: 0 0 44px 0;
  background:
    radial-gradient(circle at 12% 8%, rgba(134, 218, 255, 0.42), transparent 32%),
    radial-gradient(circle at 76% 44%, rgba(255, 217, 229, 0.5), transparent 34%),
    radial-gradient(circle at 18% 86%, rgba(255, 232, 229, 0.4), transparent 34%),
    linear-gradient(168deg, rgba(212, 240, 255, 0.92) 0%, rgba(242, 249, 255, 0.84) 50%, rgba(255, 239, 241, 0.88) 100%);
  box-shadow:
    inset -12px 0 30px rgba(255, 255, 255, 0.5),
    22px 0 56px rgba(122, 176, 218, 0.14);
  backdrop-filter: blur(24px);
  overflow: hidden;
}

.sidebar.is-admin-sidebar::before {
  content: '';
  position: absolute;
  inset: 0 0 0 auto;
  z-index: -1;
  width: 118px;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0), rgba(239, 248, 255, 0.26) 42%, rgba(239, 248, 255, 0)),
    linear-gradient(180deg, rgba(203, 235, 255, 0.22), rgba(255, 229, 228, 0.22));
  filter: blur(10px);
  transform: translateX(18px);
  pointer-events: none;
}

.sidebar.is-admin-sidebar::after {
  display: none;
}

.brand-panel {
  display: grid;
  gap: 10px;
  padding: 8px 10px 0;
}

.is-admin-sidebar .brand-panel {
  padding: 0;
}

.brand-lockup {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: #111827;
  text-decoration: none;
}

.is-admin-sidebar .brand-lockup {
  gap: 12px;
  transform: translateX(0);
}

.brand-mark {
  position: relative;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background:
    radial-gradient(circle at 30% 24%, rgba(255, 255, 255, 0.98), transparent 34%),
    conic-gradient(from 215deg, #2f7df5, #72dbff, #ffbbcd, #2f7df5);
  box-shadow:
    0 12px 26px rgba(78, 156, 232, 0.18),
    inset 0 0 0 1px rgba(255, 255, 255, 0.78);
  overflow: hidden;
}

.is-admin-sidebar .brand-mark {
  width: 48px;
  height: 48px;
  background:
    radial-gradient(circle at 50% 52%, rgba(255, 255, 255, 0.98) 0 46%, transparent 47%),
    conic-gradient(from 210deg, #91d7ff, #ffccb9, #fbf6df, #91d7ff);
  box-shadow:
    0 12px 28px rgba(92, 157, 210, 0.16),
    0 0 0 6px rgba(255, 255, 255, 0.34),
    inset 0 0 0 1px rgba(183, 146, 125, 0.18);
}

.brand-core {
  position: absolute;
  inset: 8px;
  border-radius: 50%;
  background:
    radial-gradient(circle at 34% 34%, #7dd3fc 0 28%, transparent 29%),
    radial-gradient(circle at 68% 38%, #3b82f6 0 36%, transparent 37%),
    radial-gradient(circle at 52% 68%, #38bdf8 0 42%, #2563eb 43%, #2563eb 58%, transparent 59%),
    linear-gradient(180deg, #dcefff 0%, #ffffff 100%);
}

.is-admin-sidebar .brand-core {
  inset: 10px;
  box-shadow: 0 8px 16px rgba(61, 137, 214, 0.2);
}

.brand-ring {
  position: absolute;
  inset: 0;
  border: 1px solid rgba(203, 213, 225, 0.74);
  border-radius: 50%;
}

.brand-copy {
  display: grid;
  gap: 2px;
}

.brand-copy strong {
  font-size: 24px;
  line-height: 1;
  font-weight: 820;
  color: #10213f;
}

.is-admin-sidebar .brand-copy strong {
  color: #a88478;
  font-size: 23px;
  font-weight: 800;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.72);
}

.brand-copy span {
  color: #94a3b8;
  font-size: 12px;
  font-weight: 600;
}

.is-admin-sidebar .brand-copy span {
  color: #607182;
  font-size: 12px;
  font-weight: 650;
}

.sidebar-scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  padding-right: 4px;
}

.admin-sidebar-scroll {
  padding-top: 24px;
  padding-right: 0;
  overflow: hidden;
  scrollbar-width: none;
}

.admin-sidebar-scroll::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.admin-sidebar-scroll::-webkit-scrollbar-track {
  background: transparent;
}

.admin-sidebar-scroll::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: linear-gradient(180deg, #64b4dd, #ef9298);
}

.menu-group {
  display: grid;
  gap: 8px;
}

.drive-menu-group {
  flex: 1 1 auto;
  min-height: 0;
  align-content: start;
  overflow: auto;
  padding-right: 4px;
  scrollbar-width: none;
}

.drive-menu-group::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.admin-menu-group {
  display: flex;
  min-height: calc(100vh - 178px);
  flex-direction: column;
  justify-content: space-between;
  gap: 0;
}

.menu-link {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 50px;
  padding: 0 18px;
  border-radius: 999px;
  color: #23334d;
  font-size: 16px;
  font-weight: 720;
  text-decoration: none;
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.is-admin-sidebar .menu-link {
  position: relative;
  min-height: 54px;
  padding: 0 22px;
  border-radius: 999px;
  color: #263341;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.58);
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.menu-link:hover {
  background: rgba(255, 255, 255, 0.66);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 12px 26px rgba(106, 159, 209, 0.1);
  transform: translateX(2px);
}

.is-admin-sidebar .menu-link:hover {
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.62), rgba(255, 239, 244, 0.5));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.82),
    0 14px 30px rgba(98, 172, 217, 0.17);
  transform: translateX(2px);
}

.menu-link.is-active {
  background:
    linear-gradient(90deg, rgba(199, 227, 255, 0.96), rgba(218, 241, 255, 0.92) 55%, rgba(255, 221, 232, 0.74));
  color: #12213a;
  font-weight: 820;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    inset 0 -1px 0 rgba(95, 156, 224, 0.16),
    0 16px 30px rgba(71, 141, 215, 0.16);
}

.is-admin-sidebar .menu-link.is-active {
  background:
    radial-gradient(circle at 14% 50%, rgba(255, 255, 255, 0.68), transparent 34%),
    linear-gradient(90deg, rgba(226, 249, 255, 0.96) 0%, rgba(255, 214, 223, 0.98) 100%);
  color: #24313f;
  font-weight: 820;
  box-shadow:
    0 20px 38px rgba(91, 184, 228, 0.3),
    0 10px 28px rgba(255, 157, 174, 0.2),
    inset 0 0 0 1px rgba(255, 255, 255, 0.86),
    inset 0 2px 12px rgba(255, 255, 255, 0.72);
}

.menu-icon {
  width: 20px;
  height: 20px;
  color: #66758a;
  flex: 0 0 auto;
}

.is-admin-sidebar .menu-icon {
  width: 23px;
  height: 23px;
  color: #b48883;
  filter:
    drop-shadow(0 0 4px rgba(255, 255, 255, 0.96))
    drop-shadow(0 6px 12px rgba(188, 141, 135, 0.18));
}

.menu-link.is-active .menu-icon,
.menu-link:hover .menu-icon {
  color: currentColor;
}

.is-admin-sidebar .menu-link.is-active .menu-icon,
.is-admin-sidebar .menu-link:hover .menu-icon {
  color: #b88782;
}

.sidebar-footer {
  margin-top: auto;
  flex: 0 0 auto;
}

.storage-card {
  display: grid;
  gap: 12px;
  padding: 16px 18px;
  border: 1px solid rgba(255, 255, 255, 0.84);
  border-radius: 22px;
  background:
    radial-gradient(circle at 90% 0%, rgba(116, 220, 255, 0.18), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(248, 252, 255, 0.58));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 20px 42px rgba(108, 151, 196, 0.14);
  backdrop-filter: blur(18px);
}

.storage-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.storage-card strong {
  color: #0f172a;
  font-size: 15px;
  font-weight: 700;
}

.storage-card-head span,
.storage-card p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
}

.storage-bar {
  overflow: hidden;
  height: 8px;
  border-radius: 999px;
  background: rgba(219, 231, 244, 0.9);
}

.storage-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #2f7df5 0%, #5fd3ff 58%, #ffc0d7 100%);
  box-shadow: 0 0 16px rgba(95, 211, 255, 0.54);
}

@media (max-height: 930px) and (min-width: 1181px) {
  .sidebar:not(.is-admin-sidebar) {
    gap: 14px;
    padding-top: 14px;
    padding-bottom: 14px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-panel {
    padding-top: 2px;
  }

  .drive-menu-group {
    gap: 5px;
  }

  .sidebar:not(.is-admin-sidebar) .menu-link {
    min-height: 44px;
  }

  .storage-card {
    gap: 8px;
    padding: 12px 16px;
  }
}

@media (max-height: 760px) and (min-width: 1181px) {
  .sidebar:not(.is-admin-sidebar) {
    gap: 10px;
    padding-top: 10px;
    padding-bottom: 10px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-panel {
    gap: 6px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-mark {
    width: 42px;
    height: 42px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-core {
    inset: 7px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-copy strong {
    font-size: 21px;
  }

  .sidebar:not(.is-admin-sidebar) .brand-copy span {
    font-size: 11px;
  }

  .sidebar:not(.is-admin-sidebar) .menu-link {
    min-height: 40px;
    font-size: 15px;
  }

  .drive-menu-group {
    gap: 3px;
  }

  .storage-card {
    gap: 7px;
    padding: 10px 16px;
  }

  .storage-card p {
    font-size: 12px;
  }
}

.content-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-width: 0;
  overflow: auto;
}

.topbar {
  position: relative;
  z-index: 200;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 28px 36px 0 0;
}

.create-action-wrap {
  position: relative;
  z-index: 30;
  display: inline-flex;
}

.create-menu {
  position: absolute;
  top: calc(100% + 12px);
  left: 0;
  z-index: 40;
  display: grid;
  gap: 8px;
  min-width: 176px;
  padding: 10px;
  border: 1px solid rgba(255, 255, 255, 0.84);
  border-radius: 20px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.18), transparent 36%),
    rgba(255, 255, 255, 0.82);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 22px 44px rgba(73, 112, 160, 0.2);
  backdrop-filter: blur(18px);
}

.create-menu-item {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 0 12px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #172846;
  font: inherit;
  font-weight: 780;
  cursor: pointer;
  transition:
    background 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.create-menu-item:hover {
  color: #1d70da;
  background: linear-gradient(180deg, rgba(232, 245, 255, 0.96), rgba(211, 234, 255, 0.78));
  transform: translateY(-1px);
}

.menu-action-icon {
  width: 18px;
  height: 18px;
  flex: none;
}

.hidden-file-input {
  display: none;
}

.is-admin-topbar {
  align-items: flex-start;
}

.admin-topbar-spacer {
  flex: 1;
  min-height: 44px;
}

.primary-action,
.icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 54px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  color: #0f172a;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 14px 34px rgba(99, 132, 174, 0.12);
  backdrop-filter: blur(18px);
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.primary-action:hover,
.icon-button:hover,
.icon-button.is-settings-active {
  transform: translateY(-1px);
  border-color: rgba(128, 205, 255, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.98),
    0 18px 36px rgba(56, 150, 225, 0.16);
}

.icon-button.is-settings-active {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.84), rgba(234, 247, 255, 0.66));
  color: #1677df;
}

.primary-action {
  min-width: 96px;
  padding: 0 24px;
  border-color: rgba(60, 139, 255, 0.7);
  background: linear-gradient(135deg, #2f72ff 0%, #1daee9 100%);
  color: #ffffff;
  font-size: 15px;
  font-weight: 820;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 18px 34px rgba(45, 127, 240, 0.28);
}

.button-icon,
.toolbar-icon,
.search-icon {
  width: 18px;
  height: 18px;
}

.search-shell {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 350px;
  min-height: 54px;
  padding: 0 18px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.62);
  color: #64748b;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 14px 36px rgba(99, 132, 174, 0.1);
  backdrop-filter: blur(18px);
}

.search-shell:focus-within {
  border-color: rgba(96, 165, 250, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 0 0 4px rgba(37, 99, 235, 0.1),
    0 14px 36px rgba(99, 132, 174, 0.12);
}

.search-shell input {
  flex: 1;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #172846;
  font: inherit;
  font-weight: 720;
}

.search-shell input::placeholder {
  color: #64748b;
}

.search-icon {
  color: #2563eb;
}

.search-submit,
.search-clear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #2563eb;
  cursor: pointer;
}

.search-submit:disabled {
  cursor: wait;
  opacity: 0.56;
}

.search-clear {
  color: #64748b;
  font-size: 22px;
  line-height: 1;
}

.search-submit:hover,
.search-clear:hover {
  background: rgba(37, 99, 235, 0.08);
}

.search-shell kbd {
  min-width: 24px;
  padding: 2px 7px;
  border: 1px solid rgba(203, 222, 245, 0.9);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.76);
  color: #94a3b8;
  text-align: center;
  box-shadow: inset 0 -1px 0 rgba(148, 163, 184, 0.12);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.icon-button {
  width: 66px;
  min-height: 66px;
  padding: 0;
  border-radius: 22px;
}

.counter-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  width: 58px;
  height: 58px;
  border-radius: 999px;
  background: linear-gradient(135deg, #11c9b4 0%, #31d7c8 100%);
  color: #ffffff;
  font-size: 18px;
  font-weight: 800;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 14px 28px rgba(20, 184, 166, 0.24);
}

.counter-badge img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.content-main {
  position: relative;
  z-index: 1;
  flex: 1;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow-x: hidden;
  padding: 20px 28px 28px 0;
}

.is-admin-layout .topbar {
  padding-right: 36px;
  padding-left: 0;
}

.is-admin-layout .content-main {
  padding-top: 12px;
  padding-left: 0;
  padding-right: 36px;
}

@media (max-width: 1180px) {
  .console-layout {
    grid-template-columns: 1fr;
    height: auto;
    overflow: visible;
  }

  .console-layout.is-admin-layout {
    grid-template-columns: 292px minmax(0, 1fr);
    height: 100vh;
    overflow: hidden;
  }

  .sidebar {
    position: relative;
    top: auto;
    align-self: stretch;
    height: auto;
    max-height: none;
    min-height: auto;
    overflow: visible;
    border-right: none;
    border-bottom: 1px solid rgba(214, 221, 232, 0.92);
  }

  .sidebar.is-admin-sidebar {
    position: sticky;
    top: 0;
    margin: 0;
    min-height: 100vh;
    height: 100vh;
    max-height: 100vh;
    padding: 18px;
    border-radius: 0 0 34px 0;
    overflow: hidden;
  }

  .sidebar.is-admin-sidebar::after {
    display: none;
  }

  .content-main,
  .topbar {
    padding-left: 24px;
  }

  .content-shell {
    height: auto;
    overflow: visible;
  }
}

@media (max-width: 768px) {
  .topbar {
    align-items: stretch;
    flex-wrap: wrap;
    padding: 16px;
  }

  .content-main {
    padding: 10px 16px 16px;
  }

  .search-shell,
  .admin-topbar-spacer {
    min-width: 0;
    width: 100%;
  }

  .toolbar-actions {
    margin-left: 0;
  }
}
</style>
