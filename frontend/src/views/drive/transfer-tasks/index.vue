<template>
  <main class="connect-page" @click="closeFloatingLayers" @contextmenu="showBlankMenu">
    <section class="connect-hero glass-panel">
      <div>
        <p>外部客户端接入</p>
        <h1>连接与挂载</h1>
      </div>
      <button class="primary-connect" type="button" @click.stop="createAccount">
        <el-icon><Plus /></el-icon>
        创建新账号
      </button>
    </section>

    <section class="connect-tabs glass-panel">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        type="button"
        :class="{ active: activeTab === tab.value }"
        @click.stop="activeTab = tab.value"
      >
        <el-icon><component :is="tab.icon" /></el-icon>
        {{ tab.label }}
      </button>
    </section>

    <section class="connect-toolbar glass-panel">
      <label class="search-field">
        <el-icon><Search /></el-icon>
        <input v-model="keyword" type="search" placeholder="筛选账号名称、目录、权限或状态" />
      </label>
      <div class="toolbar-actions">
        <button class="tool-button" type="button" title="刷新" @click.stop="refreshAccounts">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="view-button" type="button" :class="{ active: viewMenuVisible }" @click.stop="toggleViewMenu">
          <el-icon><Grid /></el-icon>
          <span>{{ viewMode === 'grid' ? '卡片' : '列表' }}</span>
        </button>
        <button class="sort-button" type="button" :class="{ active: sortMenuVisible }" @click.stop="toggleSortMenu">
          <el-icon><Sort /></el-icon>
          <span>{{ currentSortLabel }}</span>
        </button>
        <span class="result-count">{{ filteredAccounts.length }} / {{ accounts.length }}</span>
      </div>
    </section>

    <section class="connect-board glass-panel" :class="`is-${viewMode}`">
      <div class="board-head">
        <div class="section-title">
          <el-icon><Connection /></el-icon>
          <span>{{ currentTabTitle }}</span>
        </div>
        <div class="board-actions">
          <button type="button" @click.stop="refreshAccounts">
            <el-icon><Refresh /></el-icon>
            刷新
          </button>
          <button type="button" @click.stop="toggleSortMenu">
            <el-icon><Sort /></el-icon>
            排序
          </button>
        </div>
      </div>

      <section v-if="activeTab !== 'webdav'" class="client-guide">
        <div class="guide-icon">
          <el-icon><component :is="activeTab === 'ios' ? Iphone : Monitor" /></el-icon>
        </div>
        <div>
          <strong>{{ currentTabTitle }}</strong>
          <span>{{ activeTab === 'ios' ? '使用文件 App 或支持 WebDAV 的 iOS 客户端连接星云盘。' : '在资源管理器中映射网络位置，或使用 RaiDrive、Mountain Duck 等客户端挂载。' }}</span>
        </div>
        <button type="button" @click.stop="copyEndpoint">复制服务器地址</button>
      </section>

      <div v-if="activeTab === 'webdav' && !filteredAccounts.length" class="empty-state">
        <div class="empty-visual">
          <el-icon><Connection /></el-icon>
        </div>
        <strong>没有连接账号</strong>
        <span>创建 WebDAV 账号后，可以在这里管理挂载目录、权限、代理与客户端访问地址。</span>
        <button type="button" @click.stop="createAccount">
          <el-icon><Plus /></el-icon>
          创建新账号
        </button>
      </div>

      <article
        v-for="account in filteredAccounts"
        v-else-if="activeTab === 'webdav'"
        :key="account.id"
        class="connect-card"
        :class="{ selected: selectedIds.includes(account.id), offline: account.status !== 'active' }"
        @click.stop="toggleSelect(account)"
        @dblclick.stop="openAccount(account)"
        @contextmenu.prevent.stop="showItemMenu($event, account)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${account.name}`" @click.stop="toggleSelect(account)">
          <el-icon v-if="selectedIds.includes(account.id)"><Check /></el-icon>
        </button>
        <div class="connect-cover">
          <span class="status-chip" :class="{ offline: account.status !== 'active' }">
            {{ account.status === 'active' ? '可连接' : '已停用' }}
          </span>
          <el-icon><Connection /></el-icon>
        </div>
        <div class="connect-info">
          <strong :title="account.name">{{ account.name }}</strong>
          <span>{{ account.endpoint }}</span>
          <dl>
            <div>
              <dt>相对目录</dt>
              <dd>{{ account.root }}</dd>
            </div>
            <div>
              <dt>权限</dt>
              <dd>{{ permissionLabel(account.permission) }}</dd>
            </div>
            <div>
              <dt>反向代理</dt>
              <dd>{{ account.proxy ? '已开启' : '未开启' }}</dd>
            </div>
            <div>
              <dt>创建于</dt>
              <dd>{{ formatDate(account.createdAt) }}</dd>
            </div>
          </dl>
        </div>
        <div class="card-actions">
          <button type="button" title="打开" @click.stop="openAccount(account)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="复制地址" @click.stop="copyAccountEndpoint(account)">
            <el-icon><CopyDocument /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showItemMenu($event, account)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <section v-if="selectedAccounts.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedAccounts.length }} 个连接账号</span>
      <div>
        <button v-if="selectedAccounts.length === 1" type="button" @click.stop="openAccount(selectedAccounts[0])">打开</button>
        <button type="button" @click.stop="copySelectedEndpoints">复制地址</button>
        <button type="button" class="danger" @click.stop="deleteSelected">删除</button>
        <button type="button" @click.stop="clearSelection">取消</button>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="viewMenuVisible" class="floating-menu view-menu" :style="menuStyle(viewMenuPosition)" @click.stop>
        <p>显示方式</p>
        <button type="button" :class="{ active: viewMode === 'grid' }" @click="setViewMode('grid')">
          <el-icon><Grid /></el-icon>
          卡片视图
        </button>
        <button type="button" :class="{ active: viewMode === 'list' }" @click="setViewMode('list')">
          <el-icon><List /></el-icon>
          列表视图
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="sortMenuVisible" class="floating-menu sort-menu" :style="menuStyle(sortMenuPosition)" @click.stop>
        <p>排序方式</p>
        <button
          v-for="option in sortOptions"
          :key="option.value"
          type="button"
          :class="{ active: sortMode === option.value }"
          @click="setSortMode(option.value)"
        >
          <el-icon><component :is="option.icon" /></el-icon>
          {{ option.label }}
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="itemMenuVisible" class="floating-menu context-menu" :style="menuStyle(itemMenuPosition)" @click.stop @contextmenu.prevent.stop>
        <button type="button" @click="openContextAccount">
          <el-icon><View /></el-icon>
          打开访问地址
        </button>
        <button type="button" @click="copyContextEndpoint">
          <el-icon><CopyDocument /></el-icon>
          复制连接地址
        </button>
        <button type="button" @click="renameContextAccount">
          <el-icon><EditPen /></el-icon>
          重命名
        </button>
        <button type="button" @click="showContextDetails">
          <el-icon><InfoFilled /></el-icon>
          连接详情
        </button>
        <button type="button" @click="resetContextSecret">
          <el-icon><Key /></el-icon>
          重置密钥
        </button>
        <hr />
        <button type="button" class="danger" @click="deleteContextAccount">
          <el-icon><Delete /></el-icon>
          删除账号
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="blankMenuVisible" class="floating-menu context-menu" :style="menuStyle(blankMenuPosition)" @click.stop @contextmenu.prevent.stop>
        <button type="button" @click="createFromMenu">
          <el-icon><Plus /></el-icon>
          创建新账号
        </button>
        <button type="button" @click="refreshFromMenu">
          <el-icon><Refresh /></el-icon>
          刷新
        </button>
        <button type="button" @click="selectAll">
          <el-icon><Grid /></el-icon>
          全选
        </button>
        <button type="button" @click="invertSelection">
          <el-icon><Operation /></el-icon>
          反选
        </button>
        <button type="button" @click="clearSelection">
          <el-icon><Close /></el-icon>
          取消选择
        </button>
      </div>
    </Teleport>
  </main>
</template>

<script setup lang="ts">
import { computed, markRaw, onMounted, reactive, ref } from 'vue';
import {
  Calendar,
  Check,
  Close,
  Connection,
  CopyDocument,
  Delete,
  EditPen,
  FolderOpened,
  Grid,
  InfoFilled,
  Iphone,
  Key,
  List,
  Monitor,
  MoreFilled,
  Operation,
  Plus,
  Refresh,
  Search,
  Sort,
  View,
} from '@element-plus/icons-vue';
import type { Component } from 'vue';
import { ElIcon, ElMessage, ElMessageBox } from 'element-plus';
import {
  createDavAccount,
  deleteDavAccount,
  listDavAccounts,
  resetDavAccountSecret,
  updateDavAccount,
  type DavAccount,
  type DavAccountPermission,
} from '@/api/dav-account';

type ConnectTab = 'webdav' | 'ios' | 'windows';
type ViewMode = 'grid' | 'list';
type SortMode = 'recent' | 'name' | 'root' | 'status';

type MountAccount = {
  id: number;
  name: string;
  root: string;
  permission: DavAccountPermission;
  proxy: boolean;
  status: 'active' | 'disabled';
  endpoint: string;
  createdAt: string;
  description: string;
};

const activeTab = ref<ConnectTab>('webdav');
const keyword = ref('');
const viewMode = ref<ViewMode>('grid');
const sortMode = ref<SortMode>('recent');
const accounts = ref<MountAccount[]>([]);
const selectedIds = ref<number[]>([]);
const itemMenuTarget = ref<MountAccount | null>(null);

const viewMenuVisible = ref(false);
const sortMenuVisible = ref(false);
const itemMenuVisible = ref(false);
const blankMenuVisible = ref(false);
const viewMenuPosition = reactive({ x: 0, y: 0 });
const sortMenuPosition = reactive({ x: 0, y: 0 });
const itemMenuPosition = reactive({ x: 0, y: 0 });
const blankMenuPosition = reactive({ x: 0, y: 0 });

const tabs = [
  { value: 'webdav' as const, label: 'WebDAV 账号管理', icon: markRaw(Connection) },
  { value: 'ios' as const, label: 'iOS/iPadOS 客户端', icon: markRaw(Iphone) },
  { value: 'windows' as const, label: 'Windows 客户端', icon: markRaw(Monitor) },
];

const sortOptions: Array<{ value: SortMode; label: string; icon: Component }> = [
  { value: 'recent', label: '最新创建', icon: Calendar },
  { value: 'name', label: '账号名称', icon: Connection },
  { value: 'root', label: '相对目录', icon: FolderOpened },
  { value: 'status', label: '连接状态', icon: View },
];

const currentSortLabel = computed(() => sortOptions.find((item) => item.value === sortMode.value)?.label || '排序');
const currentTabTitle = computed(() => tabs.find((tab) => tab.value === activeTab.value)?.label || '连接与挂载');
const selectedAccounts = computed(() => accounts.value.filter((account) => selectedIds.value.includes(account.id)));

const filteredAccounts = computed(() => {
  const query = keyword.value.trim().toLowerCase();
  const source = query
    ? accounts.value.filter((account) =>
        `${account.name} ${account.root} ${account.permission} ${account.status} ${account.endpoint}`.toLowerCase().includes(query),
      )
    : accounts.value;

  return [...source].sort((a, b) => {
    if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
    if (sortMode.value === 'root') return a.root.localeCompare(b.root, 'zh-CN');
    if (sortMode.value === 'status') return a.status.localeCompare(b.status);
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
  });
});

onMounted(refreshAccounts);

function normalizeAccount(account: DavAccount): MountAccount {
  return {
    id: account.id,
    name: account.name,
    root: account.root_path,
    permission: account.permission,
    proxy: account.reverse_proxy,
    status: account.status,
    endpoint: account.endpoint,
    createdAt: account.created_at,
    description: account.description || '',
  };
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(
    new Date(value),
  );
}

function permissionLabel(value: DavAccountPermission) {
  return value === 'read' ? '只读' : '读写';
}

function menuStyle(position: { x: number; y: number }) {
  return { left: `${position.x}px`, top: `${position.y}px` };
}

function clampMenu(event: MouseEvent, position: { x: number; y: number }, width: number, height: number) {
  position.x = Math.max(12, Math.min(event.clientX, window.innerWidth - width - 12));
  position.y = Math.max(12, Math.min(event.clientY, window.innerHeight - height - 12));
}

function closeFloatingLayers() {
  viewMenuVisible.value = false;
  sortMenuVisible.value = false;
  itemMenuVisible.value = false;
  blankMenuVisible.value = false;
}

function toggleViewMenu(event: MouseEvent) {
  closeFloatingLayers();
  clampMenu(event, viewMenuPosition, 220, 156);
  viewMenuVisible.value = true;
}

function toggleSortMenu(event: MouseEvent) {
  closeFloatingLayers();
  clampMenu(event, sortMenuPosition, 230, 238);
  sortMenuVisible.value = true;
}

function setViewMode(mode: ViewMode) {
  viewMode.value = mode;
  viewMenuVisible.value = false;
}

function setSortMode(mode: SortMode) {
  sortMode.value = mode;
  sortMenuVisible.value = false;
}

function toggleSelect(account: MountAccount) {
  selectedIds.value = selectedIds.value.includes(account.id)
    ? selectedIds.value.filter((id) => id !== account.id)
    : [...selectedIds.value, account.id];
}

function selectOnly(account: MountAccount) {
  selectedIds.value = [account.id];
}

function clearSelection() {
  selectedIds.value = [];
  closeFloatingLayers();
}

async function createAccount() {
  closeFloatingLayers();
  const result = await createDavAccount({
    name: `WebDAV 账号 ${accounts.value.length + 1}`,
    root_path: '/',
    permission: 'write',
    reverse_proxy: false,
    status: 'active',
  });
  const account = normalizeAccount(result.account);
  accounts.value = [account, ...accounts.value.filter((item) => item.id !== account.id)];
  selectedIds.value = [account.id];
  await ElMessageBox.alert(
    `账号已创建。\n\n连接地址：${account.endpoint}\n账号名：${result.account.account_token}\n密钥：${result.secret}\n\n密钥仅显示一次，请妥善保存。`,
    'WebDAV 账号密钥',
    { confirmButtonText: '知道了' },
  );
}

async function refreshAccounts() {
  closeFloatingLayers();
  accounts.value = (await listDavAccounts()).map(normalizeAccount);
  selectedIds.value = selectedIds.value.filter((id) => accounts.value.some((account) => account.id === id));
  ElMessage.success('连接列表已刷新');
}

function openAccount(account: MountAccount) {
  selectOnly(account);
  closeFloatingLayers();
  window.open(account.endpoint, '_blank', 'noopener,noreferrer');
}

async function copyText(text: string, message: string) {
  await navigator.clipboard.writeText(text);
  ElMessage.success(message);
}

async function copyAccountEndpoint(account: MountAccount) {
  await copyText(account.endpoint, '连接地址已复制');
}

async function copySelectedEndpoints() {
  if (!selectedAccounts.value.length) {
    ElMessage.info('请先选择连接账号');
    return;
  }
  await copyText(selectedAccounts.value.map((account) => account.endpoint).join('\n'), '已复制选中连接地址');
}

async function copyEndpoint() {
  await copyText(`${window.location.origin}/dav`, '服务器地址已复制');
}

function showItemMenu(event: MouseEvent, account: MountAccount) {
  closeFloatingLayers();
  selectOnly(account);
  itemMenuTarget.value = account;
  clampMenu(event, itemMenuPosition, 236, 294);
  itemMenuVisible.value = true;
}

function showBlankMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.connect-card') ||
    target?.closest('.floating-menu') ||
    target?.closest('button') ||
    target?.closest('input') ||
    target?.closest('.el-overlay')
  ) {
    return;
  }

  event.preventDefault();
  closeFloatingLayers();
  clampMenu(event, blankMenuPosition, 220, 264);
  blankMenuVisible.value = true;
}

function getContextAccount() {
  return itemMenuTarget.value;
}

function openContextAccount() {
  const account = getContextAccount();
  if (account) openAccount(account);
}

async function copyContextEndpoint() {
  const account = getContextAccount();
  closeFloatingLayers();
  if (account) await copyAccountEndpoint(account);
}

async function renameContextAccount() {
  const account = getContextAccount();
  closeFloatingLayers();
  if (!account) return;
  const { value } = await ElMessageBox.prompt('输入新的账号备注名', '重命名连接账号', {
    inputValue: account.name,
    confirmButtonText: '保存',
    cancelButtonText: '取消',
  });
  const updated = await updateDavAccount(account.id, {
    name: value || account.name,
    root_path: account.root,
    permission: account.permission,
    reverse_proxy: account.proxy,
    status: account.status,
    description: account.description,
  });
  Object.assign(account, normalizeAccount(updated));
  ElMessage.success('账号名称已更新');
}

async function resetContextSecret() {
  const account = getContextAccount();
  closeFloatingLayers();
  if (!account) return;
  try {
    await ElMessageBox.confirm('重置后旧密钥会立即失效，已挂载的客户端需要更新密钥。', '重置 WebDAV 密钥', {
      confirmButtonText: '重置',
      cancelButtonText: '取消',
      type: 'warning',
    });
    const result = await resetDavAccountSecret(account.id);
    Object.assign(account, normalizeAccount(result.account));
    await ElMessageBox.alert(
      `新密钥：${result.secret}\n\n密钥仅显示一次，请妥善保存。`,
      'WebDAV 新密钥',
      { confirmButtonText: '知道了' },
    );
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '重置失败');
  }
}

function showContextDetails() {
  const account = getContextAccount();
  closeFloatingLayers();
  if (!account) return;
  ElMessageBox.alert(
    `账号：${account.name}\n地址：${account.endpoint}\n目录：${account.root}\n权限：${permissionLabel(account.permission)}\n反向代理：${account.proxy ? '已开启' : '未开启'}\n状态：${account.status === 'active' ? '可连接' : '已停用'}`,
    '连接详情',
    { confirmButtonText: '知道了' },
  );
}

async function deleteAccounts(targets: MountAccount[]) {
  if (!targets.length) return;
  await ElMessageBox.confirm(`确定删除 ${targets.length} 个连接账号吗？客户端将无法继续使用这些地址挂载。`, '删除连接账号', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  });
  const ids = new Set(targets.map((account) => account.id));
  await Promise.all(targets.map((account) => deleteDavAccount(account.id)));
  accounts.value = accounts.value.filter((account) => !ids.has(account.id));
  clearSelection();
  ElMessage.success('连接账号已删除');
}

async function deleteContextAccount() {
  const account = getContextAccount();
  closeFloatingLayers();
  if (!account) return;
  try {
    await deleteAccounts([account]);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

async function deleteSelected() {
  try {
    await deleteAccounts(selectedAccounts.value);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

function createFromMenu() {
  void createAccount();
}

function refreshFromMenu() {
  void refreshAccounts();
}

function selectAll() {
  selectedIds.value = filteredAccounts.value.map((account) => account.id);
  closeFloatingLayers();
}

function invertSelection() {
  const visible = new Set(filteredAccounts.value.map((account) => account.id));
  const selected = new Set(selectedIds.value);
  selectedIds.value = [
    ...selectedIds.value.filter((id) => !visible.has(id)),
    ...filteredAccounts.value.filter((account) => !selected.has(account.id)).map((account) => account.id),
  ];
  closeFloatingLayers();
}
</script>

<style scoped>
.connect-page {
  position: relative;
  isolation: isolate;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: calc(100vh - 112px);
  padding: 14px 26px 28px;
  overflow: visible;
}

.connect-page::before,
.connect-page::after {
  content: '';
  position: fixed;
  z-index: -1;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(10px);
}

.connect-page::before {
  width: 46vw;
  height: 38vw;
  right: 4vw;
  top: 0;
  background: radial-gradient(circle, rgba(191, 219, 254, 0.44), rgba(252, 231, 243, 0.2) 58%, transparent 72%);
}

.connect-page::after {
  width: 38vw;
  height: 34vw;
  left: 18vw;
  bottom: 0;
  background: radial-gradient(circle, rgba(186, 230, 253, 0.34), rgba(255, 214, 226, 0.2) 56%, transparent 72%);
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 28px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.5), transparent 42%),
    radial-gradient(circle at 100% 8%, rgba(252, 231, 243, 0.48), transparent 42%),
    rgba(255, 255, 255, 0.6);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 22px 64px rgba(115, 145, 190, 0.13);
  backdrop-filter: blur(24px);
}

.connect-hero,
.connect-tabs,
.connect-toolbar,
.board-head,
.selection-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.connect-hero {
  min-height: 110px;
  padding: 24px 34px;
}

.connect-hero p {
  margin: 0 0 8px;
  color: #64748b;
  font-weight: 820;
}

.connect-hero h1 {
  margin: 0;
  color: #10203d;
  font-size: 38px;
  line-height: 1.1;
  font-weight: 920;
}

.primary-connect,
.tool-button,
.view-button,
.sort-button,
.board-actions button,
.client-guide button,
.empty-state button,
.selection-strip button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 48px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  color: #172642;
  font-weight: 820;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82), 0 14px 34px rgba(99, 132, 174, 0.1);
  cursor: pointer;
}

.primary-connect {
  padding: 0 22px;
  border-color: rgba(47, 125, 245, 0.58);
  background: linear-gradient(135deg, #2f72ff 0%, #1bb6e8 100%);
  color: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.28), 0 18px 34px rgba(45, 127, 240, 0.26);
}

.connect-tabs {
  justify-content: flex-start;
  min-height: 76px;
  padding: 12px 18px;
}

.connect-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 50px;
  padding: 0 18px;
  border: 0;
  border-radius: 16px;
  background: transparent;
  color: #64748b;
  font-size: 16px;
  font-weight: 840;
  cursor: pointer;
}

.connect-tabs button.active,
.connect-tabs button:hover,
.view-button.active,
.sort-button.active,
.tool-button:hover,
.board-actions button:hover {
  background: rgba(255, 255, 255, 0.76);
  color: #1d72ed;
}

.connect-toolbar {
  min-height: 92px;
  padding: 18px 34px;
}

.search-field {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 16px;
  min-width: 240px;
  color: #2f7df5;
}

.search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #10203d;
  font-size: 18px;
  font-weight: 760;
}

.search-field input::placeholder {
  color: rgba(71, 85, 105, 0.7);
}

.toolbar-actions,
.board-actions,
.selection-strip div,
.card-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.tool-button {
  width: 54px;
  padding: 0;
}

.view-button,
.sort-button {
  padding: 0 18px;
}

.result-count {
  color: #64748b;
  font-size: 17px;
  font-weight: 840;
  white-space: nowrap;
}

.connect-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(286px, 1fr));
  align-content: start;
  gap: 18px;
  min-height: 500px;
  padding: 28px;
}

.board-head,
.client-guide,
.empty-state {
  grid-column: 1 / -1;
}

.board-head {
  min-height: 64px;
  padding: 0 6px 12px;
}

.section-title {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: #10203d;
  font-size: 24px;
  font-weight: 920;
}

.client-guide {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr) auto;
  align-items: center;
  gap: 18px;
  min-height: 148px;
  padding: 24px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.48);
}

.guide-icon,
.empty-visual {
  display: grid;
  place-items: center;
  border-radius: 24px;
  background:
    radial-gradient(circle at 25% 20%, rgba(186, 230, 253, 0.62), transparent 48%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 48%),
    rgba(255, 255, 255, 0.68);
  color: #2f7df5;
}

.guide-icon {
  width: 92px;
  height: 92px;
}

.guide-icon .el-icon,
.empty-visual .el-icon {
  width: 48px;
  height: 48px;
}

.client-guide strong,
.client-guide span {
  display: block;
}

.client-guide strong {
  color: #10203d;
  font-size: 22px;
  font-weight: 920;
}

.client-guide span {
  margin-top: 8px;
  color: #64748b;
  font-weight: 720;
}

.connect-card {
  position: relative;
  display: grid;
  grid-template-rows: 132px auto auto;
  gap: 14px;
  min-height: 360px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.64);
  border-radius: 24px;
  background:
    linear-gradient(160deg, rgba(255, 255, 255, 0.66), rgba(255, 255, 255, 0.42)),
    radial-gradient(circle at 0% 0%, rgba(191, 219, 254, 0.24), transparent 48%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.22), transparent 48%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88), 0 18px 42px rgba(105, 133, 178, 0.08);
  backdrop-filter: blur(18px);
  cursor: default;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.connect-card:hover,
.connect-card.selected {
  border-color: rgba(47, 125, 245, 0.52);
  box-shadow: 0 20px 52px rgba(77, 129, 225, 0.16);
  transform: translateY(-1px);
}

.connect-card.offline {
  opacity: 0.72;
}

.select-dot {
  position: absolute;
  z-index: 2;
  top: 14px;
  left: 14px;
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  border: 1px solid rgba(47, 125, 245, 0.42);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.78);
  color: #1d72ed;
  cursor: pointer;
}

.connect-cover {
  position: relative;
  display: grid;
  min-height: 132px;
  place-items: center;
  border-radius: 18px;
  background:
    radial-gradient(circle at 22% 20%, rgba(186, 230, 253, 0.64), transparent 44%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 46%),
    rgba(255, 255, 255, 0.62);
  color: #2f7df5;
}

.connect-cover .el-icon {
  width: 66px;
  height: 66px;
}

.status-chip {
  position: absolute;
  top: 12px;
  right: 12px;
  border-radius: 999px;
  padding: 6px 12px;
  background: rgba(34, 197, 94, 0.13);
  color: #15803d;
  font-size: 12px;
  font-weight: 900;
}

.status-chip.offline {
  background: rgba(100, 116, 139, 0.12);
  color: #64748b;
}

.connect-info {
  min-width: 0;
}

.connect-info strong,
.connect-info span,
.connect-info dt,
.connect-info dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connect-info strong,
.connect-info span {
  display: block;
}

.connect-info strong {
  color: #10203d;
  font-size: 18px;
  font-weight: 920;
}

.connect-info span {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
  font-weight: 720;
}

.connect-info dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 14px 0 0;
}

.connect-info dl div {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.48);
}

.connect-info dt,
.connect-info dd {
  margin: 0;
}

.connect-info dt {
  color: #8b9ab0;
  font-size: 12px;
  font-weight: 760;
}

.connect-info dd {
  margin-top: 4px;
  color: #172642;
  font-size: 13px;
  font-weight: 860;
}

.card-actions {
  justify-content: flex-end;
}

.card-actions button {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.74);
  color: #172642;
  cursor: pointer;
}

.card-actions button:hover {
  color: #1d72ed;
}

.connect-board.is-list {
  grid-template-columns: 1fr;
}

.connect-board.is-list .connect-card {
  grid-template-columns: 76px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 122px;
}

.connect-board.is-list .connect-cover {
  min-height: 76px;
}

.connect-board.is-list .connect-cover .el-icon {
  width: 36px;
  height: 36px;
}

.connect-board.is-list .status-chip {
  display: none;
}

.connect-board.is-list .connect-info dl {
  grid-template-columns: repeat(4, minmax(94px, 1fr));
}

.empty-state {
  display: grid;
  min-height: 360px;
  place-items: center;
  align-content: center;
  gap: 14px;
  color: #64748b;
  text-align: center;
}

.empty-visual {
  width: 104px;
  height: 104px;
}

.empty-state strong {
  color: #10203d;
  font-size: 26px;
  font-weight: 920;
}

.selection-strip {
  min-height: 74px;
  padding: 0 28px;
}

.selection-strip span {
  color: #172642;
  font-weight: 840;
}

.selection-strip .danger:hover,
.floating-menu .danger:hover {
  color: #ef4444;
}

.floating-menu {
  position: fixed;
  z-index: 5200;
  display: grid;
  gap: 8px;
  width: 230px;
  padding: 14px 12px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 8% 0%, rgba(186, 230, 253, 0.58), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 44%),
    rgba(255, 255, 255, 0.9);
  box-shadow: 0 24px 72px rgba(92, 120, 166, 0.28);
  backdrop-filter: blur(24px);
}

.floating-menu p {
  margin: 4px 8px 8px;
  color: #64748b;
  font-size: 14px;
  font-weight: 900;
}

.floating-menu button {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 42px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #172642;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  text-align: left;
}

.floating-menu button:hover,
.floating-menu button.active {
  background: rgba(255, 255, 255, 0.72);
  color: #2f7df5;
}

.floating-menu hr {
  width: 100%;
  height: 1px;
  margin: 8px 0;
  border: 0;
  background: rgba(170, 190, 215, 0.48);
}

@media (max-width: 980px) {
  .connect-hero,
  .connect-toolbar,
  .client-guide,
  .selection-strip {
    align-items: stretch;
    flex-direction: column;
  }

  .toolbar-actions,
  .board-actions,
  .selection-strip div {
    flex-wrap: wrap;
  }

  .client-guide {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .connect-page {
    padding: 12px;
  }

  .glass-panel,
  .connect-card {
    border-radius: 22px;
  }

  .connect-hero,
  .connect-toolbar,
  .connect-board {
    padding: 18px;
  }

  .connect-hero h1 {
    font-size: 30px;
  }
}
</style>
