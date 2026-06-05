<template>
  <section class="user-groups-page">
    <div class="page-shell">
      <header class="page-header">
        <div class="page-header-copy">
          <p class="page-eyebrow">User Group Console</p>
          <h1>用户组</h1>
          <p class="page-subtitle">
            统一管理用户组的存储策略、成员规模与容量上限，并在这里设置默认用户组。
          </p>
        </div>
      </header>

      <section class="metric-grid">
        <article v-for="metric in metrics" :key="metric.label" class="metric-card">
          <span class="metric-label">{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
          <small>{{ metric.detail }}</small>
        </article>
      </section>

      <section class="toolbar">
        <button class="primary-action" type="button" @click="openCreateDialog">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M10 4.5v11" />
            <path d="M4.5 10h11" />
          </svg>
          <span>新建用户组</span>
        </button>

        <button class="ghost-action" type="button" :disabled="loading" @click="refreshGroups">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M16.2 9.2A6.2 6.2 0 1 0 15 13.8" />
            <path d="M16.2 4.8v4.7h-4.7" />
          </svg>
          <span>{{ loading ? '刷新中' : '刷新' }}</span>
        </button>
      </section>

      <section class="table-card">
        <div class="table-head">
          <div class="table-row table-row-head">
            <div class="col-index">#</div>
            <div class="col-name">名称</div>
            <div class="col-policy">存储策略</div>
            <div class="col-users">成员数</div>
            <div class="col-capacity">容量上限</div>
            <div class="col-actions">操作</div>
          </div>
        </div>

        <div class="table-body">
          <article v-for="group in pagedGroups" :key="group.id" class="table-row table-row-body">
            <div class="col-index">
              <span class="index-pill">{{ group.id }}</span>
            </div>

            <div class="col-name">
              <div class="group-cell">
                <div class="group-title-row">
                  <span class="group-name">{{ group.name }}</span>
                  <span v-if="group.isDefault" class="default-badge">默认组</span>
                </div>
                <small>{{ group.description || '未填写分组说明' }}</small>
              </div>
            </div>

            <div class="col-policy">
              <span class="policy-pill">{{ group.policy }}</span>
            </div>

            <div class="col-users">
              <button class="member-stack-button" type="button" @click="openMembersDialog(group)">
                <span class="avatar-stack" aria-hidden="true">
                  <span
                    v-for="avatar in avatarStack(group)"
                    :key="avatar"
                    class="stack-avatar"
                    :class="`tone-${avatar % 4}`"
                  >
                    {{ group.name.slice(avatar, avatar + 1).toUpperCase() || avatar + 1 }}
                  </span>
                </span>
                <strong>{{ group.userCount }}</strong>
              </button>
            </div>

            <div class="col-capacity">
              <span class="capacity-text">{{ formatCapacity(group.capacityBytes) }}</span>
            </div>

            <div class="col-actions">
              <button
                class="line-button subtle"
                type="button"
                :disabled="group.isDefault || defaultGroupSavingId === group.id"
                @click="setAsDefaultGroup(group)"
              >
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M10 2.8 12.2 7l4.7.7-3.4 3.3.8 4.7-4.3-2.2-4.3 2.2.8-4.7-3.4-3.3 4.7-.7Z" />
                </svg>
                <span>{{ group.isDefault ? '默认中' : defaultGroupSavingId === group.id ? '设置中' : '设为默认' }}</span>
              </button>

              <button class="line-button" type="button" @click="openEditDialog(group)">
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M4 13.5V16h2.5L15 7.5 12.5 5 4 13.5Z" />
                  <path d="m11.8 5.7 2.5 2.5" />
                </svg>
                <span>编辑</span>
              </button>

              <button class="icon-button" type="button" aria-label="删除用户组" @click="openDeleteDialog(group)">
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M6.5 6.5v8.5" />
                  <path d="M10 6.5v8.5" />
                  <path d="M13.5 6.5v8.5" />
                  <path d="M4.5 5.5h11" />
                  <path d="M7.75 5.5V4.2a1 1 0 0 1 1-1h2.5a1 1 0 0 1 1 1v1.3" />
                  <path d="M6 5.5V16a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V5.5" />
                </svg>
              </button>
            </div>
          </article>
        </div>
      </section>

      <footer class="table-footer">
        <div class="pager">
          <button class="pager-arrow" type="button" :disabled="page === 1" @click="page = Math.max(1, page - 1)">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="m11.5 5.5-4.5 4.5 4.5 4.5" />
            </svg>
          </button>
          <span class="pager-index">{{ page }}</span>
          <button
            class="pager-arrow"
            type="button"
            :disabled="page >= totalPages"
            @click="page = Math.min(totalPages, page + 1)"
          >
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="m8.5 5.5 4.5 4.5-4.5 4.5" />
            </svg>
          </button>
        </div>

        <label class="page-size-select">
          <select v-model.number="pageSize">
            <option :value="10">每页 10 条</option>
            <option :value="20">每页 20 条</option>
            <option :value="50">每页 50 条</option>
          </select>
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="m5.5 7.5 4.5 5 4.5-5" />
          </svg>
        </label>
      </footer>
    </div>

    <div v-if="dialogOpen" class="dialog-mask" @click.self="closeDialog">
      <div class="dialog-panel">
        <div class="dialog-header">
          <div>
            <p class="dialog-eyebrow">{{ dialogMode === 'create' ? 'Create Group' : 'Edit Group' }}</p>
            <h2>{{ dialogMode === 'create' ? '新建用户组' : `编辑 ${form.name || '用户组'}` }}</h2>
          </div>
          <button class="dialog-close" type="button" aria-label="关闭" @click="closeDialog">x</button>
        </div>

        <div class="dialog-body">
          <label class="field-block">
            <span>用户组名称</span>
            <input v-model.trim="form.name" type="text" class="field-input" placeholder="例如：企业高级版" />
            <small>用于在注册分组、套餐面板和策略绑定中展示。</small>
          </label>

          <label class="field-block">
            <span>分组说明</span>
            <textarea
              v-model.trim="form.description"
              class="field-textarea"
              rows="4"
              placeholder="描述该用户组适合的人群、能力边界和定位"
            ></textarea>
            <small>让运营和管理员能快速识别这个分组的用途。</small>
          </label>

          <label class="field-block">
            <span>关联存储策略</span>
            <div class="select-shell">
              <select v-model.number="form.storagePolicyId">
                <option v-if="storagePolicies.length === 0" :value="0" disabled>
                  暂无可用存储策略
                </option>
                <option v-for="policy in storagePolicies" :key="policy.id" :value="policy.id">
                  {{ policy.name }}
                </option>
              </select>
              <svg viewBox="0 0 20 20" aria-hidden="true">
                <path d="m5.5 7.5 4.5 5 4.5-5" />
              </svg>
            </div>
            <small>保存后，该用户组的新成员会使用这套存储能力与上传约束。</small>
          </label>

          <label class="field-block">
            <span>容量上限</span>
            <div class="capacity-row">
              <input
                v-model.number="form.capacityValue"
                type="number"
                min="0"
                class="field-input"
                :disabled="form.unlimited"
                placeholder="50"
              />
              <div class="select-shell unit-shell" :class="{ 'is-disabled': form.unlimited }">
                <select v-model="form.capacityUnit" :disabled="form.unlimited">
                  <option value="MB">MB</option>
                  <option value="GB">GB</option>
                  <option value="TB">TB</option>
                </select>
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="m5.5 7.5 4.5 5 4.5-5" />
                </svg>
              </div>

            </div>
            <small>设为 0 或开启无限制时，表示该用户组不限制总容量。</small>
          </label>

          <label class="toggle-field">
            <button class="switch-button" :class="{ active: form.unlimited }" type="button" @click="toggleUnlimitedCapacity">
              <span></span>
            </button>
            <div>
              <strong>无限制容量</strong>
              <small>适合管理员、VIP 套餐或特殊运维分组。</small>
            </div>
          </label>

          <label v-if="dialogMode === 'edit'" class="toggle-field">
            <button
              class="switch-button"
              :class="{ active: form.syncMemberCapacity }"
              type="button"
              @click="form.syncMemberCapacity = !form.syncMemberCapacity"
            >
              <span></span>
            </button>
            <div>
              <strong>同步已有成员容量</strong>
              <small>开启后，本组所有成员的个人容量会更新为当前用户组容量上限；已用容量不会改变。</small>
            </div>
          </label>
        </div>

        <div class="dialog-actions">
          <button class="ghost-action dialog-button" type="button" @click="closeDialog">取消</button>
          <button class="primary-action dialog-button" type="button" :disabled="saving" @click="submitDialog">
            {{ saving ? '保存中' : dialogMode === 'create' ? '创建用户组' : '保存修改' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="membersDialogOpen" class="dialog-mask" @click.self="closeMembersDialog">
      <div class="dialog-panel members-panel">
        <div class="dialog-header">
          <div>
            <p class="dialog-eyebrow">Group Members</p>
            <h2>{{ membersDialogTitle }}</h2>
          </div>
          <button class="dialog-close" type="button" aria-label="关闭" @click="closeMembersDialog">x</button>
        </div>

        <div class="members-toolbar">
          <label class="search-shell">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <circle cx="9" cy="9" r="5.5" />
              <path d="m13.5 13.5 3 3" />
            </svg>
            <input v-model.trim="membersSearch" type="text" placeholder="搜索用户名或邮箱" />
          </label>

          <label class="select-shell filter-shell">
            <select v-model="membersRoleFilter">
              <option value="all">全部角色</option>
              <option value="admin">管理员</option>
              <option value="user">用户</option>
            </select>
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="m5.5 7.5 4.5 5 4.5-5" />
            </svg>
          </label>
        </div>

        <div class="members-shell">
          <div class="members-head">
            <span>用户名</span>
            <span>邮箱</span>
            <span>角色</span>
            <span>容量进度</span>
          </div>

          <div v-if="membersLoading" class="empty-state">正在加载成员列表...</div>
          <div v-else-if="!filteredMembers.length" class="empty-state">当前筛选条件下没有成员</div>

          <div v-else class="members-list">
            <article v-for="member in filteredMembers" :key="member.id" class="member-row">
              <div class="member-main">
                <strong>{{ member.username }}</strong>
                <small>#{{ member.id }}</small>
              </div>
              <div class="member-email">{{ member.email }}</div>
              <div>
                <span class="role-pill" :class="`is-${member.role}`">{{ roleLabel(member.role) }}</span>
              </div>
              <div class="member-storage">
                <div class="member-storage-copy">
                  <span>{{ formatCapacity(member.usedSize) }} / {{ formatCapacity(member.capacity) }}</span>
                  <strong>{{ usagePercent(member) }}%</strong>
                </div>
                <div class="progress-track">
                  <span class="progress-fill" :style="{ width: `${usagePercent(member)}%` }"></span>
                </div>
              </div>
            </article>
          </div>
        </div>

        <div class="dialog-actions">
          <button class="ghost-action dialog-button" type="button" @click="closeMembersDialog">关闭</button>
        </div>
      </div>
    </div>

    <div v-if="deleteDialogOpen && pendingDeleteGroup" class="dialog-mask" @click.self="closeDeleteDialog">
      <div class="dialog-panel confirm-panel">
        <div class="dialog-header">
          <div>
            <p class="dialog-eyebrow">Delete Confirm</p>
            <h2>删除 {{ pendingDeleteGroup.name }}</h2>
          </div>
          <button class="dialog-close" type="button" aria-label="关闭" @click="closeDeleteDialog">x</button>
        </div>

        <div class="confirm-copy">
          <p>这会删除该用户组配置。若该组仍有成员，后端会阻止删除。</p>
          <small>建议先查看成员，确认无误后再执行删除。</small>
        </div>

        <div class="dialog-actions">
          <button class="ghost-action dialog-button" type="button" @click="closeDeleteDialog">取消</button>
          <button class="danger-action dialog-button" type="button" :disabled="deleting" @click="removeGroupConfirmed">
            {{ deleting ? '删除中' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { getSiteSettings, updateSiteSettings, type SiteSettingsPayload } from '@/api/site-settings';
import { listStoragePolicies, type StoragePolicyPayload } from '@/api/storage-policy';
import { bytesToCapacity, capacityToBytes, type CapacityUnit } from '@/utils/capacity';
import {
  createUserGroup,
  deleteUserGroup,
  getUserGroupSummary,
  listUserGroupMembers,
  listUserGroups,
  updateUserGroup,
  type UserGroupMemberPayload,
  type UserGroupPayload,
  type UserGroupSummaryPayload,
} from '@/api/user-groups';

type UserGroupRecord = {
  id: number;
  name: string;
  description: string;
  policy: string;
  storagePolicyId: number;
  userCount: number;
  capacityBytes: number;
  isDefault: boolean;
};

type UserGroupMemberRecord = {
  id: number;
  username: string;
  email: string;
  role: string;
  capacity: number;
  usedSize: number;
};

type StoragePolicyOption = {
  id: number;
  name: string;
};

type DialogMode = 'create' | 'edit';

const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const saving = ref(false);
const deleting = ref(false);
const defaultGroupSavingId = ref<number | null>(null);
const dialogOpen = ref(false);
const deleteDialogOpen = ref(false);
const membersDialogOpen = ref(false);
const membersLoading = ref(false);
const dialogMode = ref<DialogMode>('create');
const editingGroupId = ref<number | null>(null);
const membersDialogTitle = ref('\u6210\u5458\u5217\u8868');
const pendingDeleteGroup = ref<UserGroupRecord | null>(null);
const groups = ref<UserGroupRecord[]>([]);
const groupSummary = ref<UserGroupSummaryPayload | null>(null);
const members = ref<UserGroupMemberRecord[]>([]);
const storagePolicies = ref<StoragePolicyOption[]>([]);
const storagePoliciesLoadFailed = ref(false);
const membersSearch = ref('');
const membersRoleFilter = ref<'all' | 'admin' | 'user'>('all');
const siteSettings = ref<SiteSettingsPayload | null>(null);

const form = reactive({
  name: '',
  description: '',
  storagePolicyId: 0,
  capacityValue: 50,
  capacityUnit: 'GB' as CapacityUnit,
  unlimited: false,
  syncMemberCapacity: false,
});

const totalPages = computed(() => Math.max(1, Math.ceil(groups.value.length / pageSize.value)));
const pagedGroups = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return groups.value.slice(start, start + pageSize.value);
});

const filteredMembers = computed(() => {
  const keyword = membersSearch.value.trim().toLowerCase();
  return members.value.filter((member) => {
    const roleMatch = membersRoleFilter.value === 'all' || member.role === membersRoleFilter.value;
    const keywordMatch =
      !keyword ||
      member.username.toLowerCase().includes(keyword) ||
      member.email.toLowerCase().includes(keyword);
    return roleMatch && keywordMatch;
  });
});

const metrics = computed(() => {
  const summary = groupSummary.value;
  const threshold = summary?.high_capacity_bytes || 200 * 1024 * 1024 * 1024;
  return [
    { label: '用户组总数', value: `${summary?.total_groups ?? 0}`, detail: '后端实时统计的分组数量' },
    { label: '覆盖用户数', value: `${summary?.total_users ?? 0}`, detail: '按 users.user_group_id 汇总的成员规模' },
    { label: '默认用户组', value: summary?.default_group || '未配置', detail: '新用户注册后默认归属的分组' },
    { label: '高配/无限制', value: `${summary?.high_capacity_groups ?? 0} / ${summary?.unlimited_groups ?? 0}`, detail: `容量 >= ${formatCapacity(threshold)} / max_capacity=0` },
  ];
});

watch(totalPages, (value) => {
  if (page.value > value) {
    page.value = value;
  }
});

watch(pageSize, () => {
  page.value = 1;
});

function mapGroup(payload: UserGroupPayload): UserGroupRecord {
  const defaultGroupName = siteSettings.value?.default_group?.trim() || '';
  return {
    id: Number(payload.id || 0),
    name: payload.name,
    description: payload.description,
    policy: payload.storage_policy_name || '未绑定策略',
    storagePolicyId: Number(payload.storage_policy_id || 0),
    userCount: Number(payload.user_count || 0),
    capacityBytes: Number(payload.max_capacity || 0),
    isDefault: defaultGroupName !== '' && payload.name === defaultGroupName,
  };
}

function mapMember(payload: UserGroupMemberPayload): UserGroupMemberRecord {
  return {
    id: Number(payload.id || 0),
    username: payload.username,
    email: payload.email,
    role: payload.role,
    capacity: Number(payload.capacity || 0),
    usedSize: Number(payload.used_size || 0),
  };
}

function mapPolicy(payload: StoragePolicyPayload): StoragePolicyOption {
  return {
    id: Number(payload.id || 0),
    name: payload.name,
  };
}

function roleLabel(role: string) {
  return role === 'admin' ? '管理员' : '用户';
}

function avatarStack(group: UserGroupRecord) {
  const fallback = group.isDefault ? 3 : 2;
  const size = Math.min(4, Math.max(1, group.userCount || fallback));
  return Array.from({ length: size }, (_, index) => index);
}

function formatCapacity(bytes: number) {
  if (!bytes || bytes <= 0) {
    return '无限制';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  const precision = value >= 100 || unitIndex === 0 ? 0 : value >= 10 ? 1 : 2;
  return `${value.toFixed(precision)} ${units[unitIndex]}`;
}

function usagePercent(member: UserGroupMemberRecord) {
  if (!member.capacity || member.capacity <= 0) {
    return 100;
  }
  return Math.max(0, Math.min(100, Math.round((member.usedSize / member.capacity) * 100)));
}

function resetForm() {
  form.name = '';
  form.description = '';
  form.storagePolicyId = storagePolicies.value[0]?.id || 0;
  form.capacityValue = 50;
  form.capacityUnit = 'GB';
  form.unlimited = false;
  form.syncMemberCapacity = false;
}

function getFormCapacityBytes() {
  return capacityToBytes(form.capacityValue, form.capacityUnit, form.unlimited);
}

function toggleUnlimitedCapacity() {
  form.unlimited = !form.unlimited;
  if (!form.unlimited && (!form.capacityValue || form.capacityValue <= 0)) {
    form.capacityValue = 50;
    form.capacityUnit = 'GB';
  }
}

function fillForm(group: UserGroupRecord) {
  const capacity = bytesToCapacity(group.capacityBytes);
  form.name = group.name;
  form.description = group.description;
  form.storagePolicyId = group.storagePolicyId || storagePolicies.value[0]?.id || 0;
  form.capacityValue = capacity.value;
  form.capacityUnit = capacity.unit;
  form.unlimited = capacity.unlimited;
  form.syncMemberCapacity = false;
}

async function fetchSiteSettings() {
  siteSettings.value = await getSiteSettings();
}

async function fetchGroups(showSuccess = false) {
  loading.value = true;
  try {
    const data = await listUserGroups();
    groups.value = data.map(mapGroup);
    try {
      groupSummary.value = await getUserGroupSummary();
    } catch {
      groupSummary.value = buildLocalSummary(groups.value);
    }
    if (showSuccess) {
      ElMessage.success('\u7528\u6237\u7ec4\u5217\u8868\u5df2\u5237\u65b0');
    }
  } finally {
    loading.value = false;
  }
}

function buildLocalSummary(items: UserGroupRecord[]): UserGroupSummaryPayload {
  const highCapacityBytes = 200 * 1024 * 1024 * 1024;
  return {
    total_groups: items.length,
    total_users: items.reduce((sum, item) => sum + item.userCount, 0),
    default_group: siteSettings.value?.default_group?.trim() || items.find((item) => item.isDefault)?.name || '',
    high_capacity_groups: items.filter((item) => item.capacityBytes >= highCapacityBytes).length,
    unlimited_groups: items.filter((item) => item.capacityBytes <= 0).length,
    high_capacity_bytes: highCapacityBytes,
  };
}

async function fetchStoragePolicies() {
  try {
    const data = await listStoragePolicies();
    storagePolicies.value = data.map(mapPolicy);
    storagePoliciesLoadFailed.value = false;
    if (!form.storagePolicyId) {
      form.storagePolicyId = storagePolicies.value[0]?.id || 0;
    }
  } catch (error) {
    storagePolicies.value = [];
    storagePoliciesLoadFailed.value = true;
    form.storagePolicyId = 0;
    ElMessage.error(error instanceof Error ? error.message : '存储策略加载失败');
    throw error;
  }
}

async function refreshGroups() {
  try {
    await Promise.all([fetchSiteSettings(), fetchStoragePolicies()]);
    await fetchGroups(true);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '刷新用户组失败');
  }
}

function openCreateDialog() {
  dialogMode.value = 'create';
  editingGroupId.value = null;
  resetForm();
  dialogOpen.value = true;
}

function openEditDialog(group: UserGroupRecord) {
  dialogMode.value = 'edit';
  editingGroupId.value = group.id;
  fillForm(group);
  dialogOpen.value = true;
}

function closeDialog() {
  if (saving.value) return;
  dialogOpen.value = false;
  editingGroupId.value = null;
}

async function submitDialog() {
  if (!form.name.trim()) {
    ElMessage.warning('用户组名称不能为空');
    return;
  }
  if (!form.storagePolicyId) {
    ElMessage.warning('请选择存储策略');
    return;
  }
  if (storagePoliciesLoadFailed.value || storagePolicies.value.length === 0) {
    ElMessage.warning('存储策略未成功加载，不能保存用户组');
    return;
  }
  if (!storagePolicies.value.some((policy) => policy.id === form.storagePolicyId)) {
    ElMessage.warning('当前选择的存储策略不存在，请刷新后重新选择');
    return;
  }

  const payload = {
    name: form.name.trim(),
    description: form.description.trim(),
    storage_policy_id: form.storagePolicyId,
    max_capacity: getFormCapacityBytes(),
    sync_member_capacity: dialogMode.value === 'edit' && form.syncMemberCapacity,
  };

  saving.value = true;
  try {
    if (dialogMode.value === 'create') {
      const created = await createUserGroup(payload);
      groups.value.unshift(mapGroup(created));
      ElMessage.success('用户组已创建');
    } else if (editingGroupId.value) {
      const wasDefault = groups.value.some((item) => item.id === editingGroupId.value && item.isDefault);
      const updated = await updateUserGroup(editingGroupId.value, payload);
      if (wasDefault && siteSettings.value) {
        siteSettings.value = {
          ...siteSettings.value,
          default_group: updated.name,
        };
      }
      const next = mapGroup(updated);
      const index = groups.value.findIndex((item) => item.id === editingGroupId.value);
      if (index >= 0) {
        groups.value.splice(index, 1, next);
      }
      ElMessage.success('\u7528\u6237\u7ec4\u4fee\u6539\u5df2\u4fdd\u5b58');
    }
    await fetchGroups();
    closeDialog();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存用户组失败');
  } finally {
    saving.value = false;
  }
}

async function setAsDefaultGroup(group: UserGroupRecord) {
  if (!siteSettings.value || group.isDefault) {
    return;
  }
  defaultGroupSavingId.value = group.id;
  try {
    const nextSettings: SiteSettingsPayload = {
      ...siteSettings.value,
      default_group: group.name,
    };
    siteSettings.value = await updateSiteSettings(nextSettings);
    groups.value = groups.value.map((item) => ({
      ...item,
      isDefault: item.id === group.id,
    }));
    ElMessage.success(`已将 ${group.name} 设为默认用户组`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '设置默认用户组失败');
  } finally {
    defaultGroupSavingId.value = null;
  }
}

async function openMembersDialog(group: UserGroupRecord) {
  membersDialogTitle.value = `${group.name} 的成员（${group.userCount}）`;
  membersDialogOpen.value = true;
  membersLoading.value = true;
  members.value = [];
  membersSearch.value = '';
  membersRoleFilter.value = 'all';
  try {
    const data = await listUserGroupMembers(group.id);
    members.value = data.map(mapMember);
    membersDialogTitle.value = `${group.name} 的成员（${members.value.length}）`;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载成员列表失败');
  } finally {
    membersLoading.value = false;
  }
}

function closeMembersDialog() {
  membersDialogOpen.value = false;
  members.value = [];
  membersSearch.value = '';
  membersRoleFilter.value = 'all';
}

function openDeleteDialog(group: UserGroupRecord) {
  pendingDeleteGroup.value = group;
  deleteDialogOpen.value = true;
}

function closeDeleteDialog() {
  if (deleting.value) return;
  deleteDialogOpen.value = false;
  pendingDeleteGroup.value = null;
}

async function removeGroupConfirmed() {
  if (!pendingDeleteGroup.value) return;
  const groupName = pendingDeleteGroup.value.name;
  deleting.value = true;
  try {
    await deleteUserGroup(pendingDeleteGroup.value.id);
    await fetchGroups();
    ElMessage.success(`已删除用户组：${groupName}`);
    closeDeleteDialog();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除用户组失败');
  } finally {
    deleting.value = false;
  }
}

onMounted(async () => {
  try {
    await Promise.all([fetchSiteSettings(), fetchStoragePolicies()]);
    await fetchGroups();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '用户组页面加载失败');
  }
});
</script>

<style scoped>
.user-groups-page {
  min-height: calc(100vh - 104px);
  padding: 4px 2px 24px;
}

.page-shell {
  display: grid;
  gap: 20px;
  min-height: calc(100vh - 120px);
  padding: 28px 28px 30px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 26px;
  background:
    radial-gradient(circle at 12% 0%, rgba(251, 207, 232, 0.22), transparent 26%),
    radial-gradient(circle at 92% 8%, rgba(96, 165, 250, 0.18), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.94) 0%, rgba(248, 251, 255, 0.88) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 24px 58px rgba(15, 23, 42, 0.07);
  backdrop-filter: blur(18px);
}

.page-header-copy h1 {
  margin: 0;
  color: #1b2430;
  font-size: 56px;
  font-weight: 800;
  letter-spacing: -0.04em;
}

.page-eyebrow,
.dialog-eyebrow {
  margin: 0 0 12px;
  color: #3b82f6;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.page-subtitle {
  max-width: 760px;
  margin: 14px 0 0;
  color: #718096;
  font-size: 15px;
  line-height: 1.9;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  gap: 8px;
  padding: 20px 22px;
  border: 1px solid rgba(227, 232, 240, 0.95);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 12px 28px rgba(148, 163, 184, 0.08);
}

.metric-label {
  color: #94a3b8;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.metric-card strong {
  color: #111827;
  font-size: 28px;
  letter-spacing: -0.03em;
}

.metric-card small {
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
}

.toolbar,
.table-footer,
.capacity-row,
.dialog-actions,
.members-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.primary-action,
.ghost-action,
.danger-action,
.page-size-select,
.line-button,
.dialog-button,
.select-shell {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 48px;
  padding: 0 18px;
  border-radius: 16px;
  font-size: 15px;
  font-weight: 700;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.primary-action {
  border: 1px solid transparent;
  background: linear-gradient(135deg, #2563eb 0%, #38bdf8 100%);
  box-shadow: 0 18px 34px rgba(37, 99, 235, 0.24);
  color: #fff;
  cursor: pointer;
}

.danger-action {
  border: 1px solid transparent;
  background: linear-gradient(135deg, #ef4444 0%, #f97316 100%);
  box-shadow: 0 18px 34px rgba(239, 68, 68, 0.22);
  color: #fff;
  cursor: pointer;
}

.ghost-action,
.page-size-select,
.line-button,
.select-shell {
  border: 1px solid #e3e8ef;
  background: #fff;
  color: #334155;
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.12);
}

.line-button.subtle {
  color: #2563eb;
  background: rgba(239, 246, 255, 0.9);
  border-color: rgba(147, 197, 253, 0.42);
}

.primary-action:hover,
.ghost-action:hover,
.danger-action:hover,
.icon-button:hover,
.pager-arrow:hover,
.page-size-select:hover,
.line-button:hover,
.select-shell:hover,
.count-link.interactive:hover {
  transform: translateY(-1px);
}

.primary-action:disabled,
.ghost-action:disabled,
.danger-action:disabled,
.dialog-button:disabled,
.line-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.primary-action svg,
.ghost-action svg,
.icon-button svg,
.pager-arrow svg,
.page-size-select svg,
.line-button svg,
.select-shell svg,
.search-shell svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.table-card {
  overflow: hidden;
  border: 1px solid rgba(228, 233, 241, 0.82);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(249, 252, 255, 0.9)),
    radial-gradient(circle at 14% 0%, rgba(251, 207, 232, 0.2), transparent 28%),
    radial-gradient(circle at 94% 10%, rgba(125, 211, 252, 0.18), transparent 30%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.78),
    0 18px 42px rgba(15, 23, 42, 0.05);
  backdrop-filter: blur(18px);
}

.table-row {
  display: grid;
  grid-template-columns: 80px minmax(220px, 1.4fr) minmax(160px, 1fr) 120px 140px 280px;
  align-items: center;
  gap: 10px;
}

.table-row-head {
  min-height: 62px;
  padding: 0 24px;
  border-bottom: 1px solid #eef2f7;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 800;
}

.table-row-body {
  min-height: 92px;
  padding: 0 24px;
  border-bottom: 1px solid #f1f5f9;
  background: rgba(255, 255, 255, 0.42);
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease;
}

.table-row-body:hover {
  background: rgba(255, 255, 255, 0.8);
  box-shadow: inset 3px 0 0 rgba(59, 130, 246, 0.55);
}

.table-row-body:last-child {
  border-bottom: none;
}

.index-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 34px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f3f7fd;
  color: #2563eb;
  font-size: 13px;
  font-weight: 800;
}

.group-cell {
  display: grid;
  gap: 6px;
}

.group-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.group-name {
  color: #111827;
  font-size: 18px;
  font-weight: 800;
}

.default-badge {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(244, 114, 182, 0.88), rgba(251, 146, 160, 0.88));
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  box-shadow: 0 10px 22px rgba(244, 114, 182, 0.2);
}

.group-cell small {
  color: #7b8794;
  font-size: 13px;
  line-height: 1.7;
}

.policy-pill,
.role-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 800;
}

.policy-pill {
  background: rgba(219, 234, 254, 0.72);
  color: #1d4ed8;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.role-pill.is-admin {
  background: linear-gradient(135deg, rgba(30, 64, 175, 0.92), rgba(15, 23, 42, 0.92));
  color: #fff;
  box-shadow: 0 12px 26px rgba(37, 99, 235, 0.22);
}

.role-pill.is-user {
  background: linear-gradient(135deg, rgba(244, 114, 182, 0.86), rgba(251, 146, 160, 0.86));
  color: #fff;
  box-shadow: 0 12px 26px rgba(244, 114, 182, 0.2);
}

.count-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: #2563eb;
  font-size: 16px;
  font-weight: 800;
}

.count-link.interactive {
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: rgba(37, 99, 235, 0.26);
  text-underline-offset: 4px;
}

.member-stack-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 0;
  border: none;
  background: transparent;
  color: #1d4ed8;
  cursor: pointer;
}

.member-stack-button strong {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 34px;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.7);
  font-size: 14px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.avatar-stack {
  display: inline-flex;
  align-items: center;
  padding-left: 8px;
}

.stack-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  margin-left: -8px;
  border: 2px solid rgba(255, 255, 255, 0.88);
  border-radius: 50%;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.36),
    0 10px 22px rgba(59, 130, 246, 0.16);
}

.stack-avatar.tone-0 {
  background: linear-gradient(135deg, #38bdf8, #2563eb);
}

.stack-avatar.tone-1 {
  background: linear-gradient(135deg, #f9a8d4, #f472b6);
}

.stack-avatar.tone-2 {
  background: linear-gradient(135deg, #93c5fd, #0f172a);
}

.stack-avatar.tone-3 {
  background: linear-gradient(135deg, #67e8f9, #a78bfa);
}

.capacity-text {
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
}

.col-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-wrap: wrap;
}

.line-button {
  min-height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  cursor: pointer;
}

.line-button span {
  white-space: nowrap;
}

.icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #7c8795;
  cursor: pointer;
}

.table-footer {
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
  width: 40px;
  height: 40px;
  border-radius: 999px;
}

.pager-arrow {
  border: 1px solid #e4e9f1;
  background: #fff;
  color: #64748b;
  cursor: pointer;
}

.pager-arrow:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.pager-index {
  background: #f1f5f9;
  color: #0f172a;
  font-weight: 700;
}

.page-size-select {
  padding-right: 44px;
}

.select-shell {
  padding: 0;
  overflow: visible;
}

.page-size-select select,
.select-shell select {
  position: relative;
  z-index: 1;
  min-height: 48px;
  width: 100%;
  padding: 0 44px 0 16px;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 15px;
  font-weight: 700;
  appearance: none;
  outline: none;
  cursor: pointer;
}

.page-size-select svg,
.select-shell svg {
  position: absolute;
  z-index: 0;
  right: 16px;
  pointer-events: none;
}

.select-shell.is-disabled,
.field-input:disabled {
  opacity: 0.62;
  cursor: not-allowed;
}

.select-shell.is-disabled select {
  cursor: not-allowed;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.28);
  backdrop-filter: blur(16px);
}

.dialog-panel {
  width: min(760px, 100%);
  display: grid;
  gap: 22px;
  padding: 28px;
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 30px 80px rgba(15, 23, 42, 0.18);
}

.members-panel {
  width: min(980px, 100%);
}

.confirm-panel {
  width: min(560px, 100%);
}

.dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.dialog-header h2 {
  margin: 0;
  color: #111827;
  font-size: 34px;
  letter-spacing: -0.03em;
}

.dialog-close {
  width: 38px;
  height: 38px;
  border: 1px solid #d9e2ec;
  border-radius: 999px;
  background: #fff;
  color: #64748b;
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
}

.dialog-body {
  display: grid;
  gap: 18px;
}

.field-block {
  display: grid;
  gap: 10px;
}

.field-block > span,
.toggle-field strong {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
}

.field-block small,
.toggle-field small,
.confirm-copy small {
  color: #6b7280;
  font-size: 13px;
  line-height: 1.8;
}

.field-input,
.field-textarea {
  width: 100%;
  min-height: 52px;
  padding: 0 16px;
  border: 1px solid #d8e1eb;
  border-radius: 16px;
  background: #fff;
  color: #111827;
  font-size: 15px;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
  box-sizing: border-box;
}

.field-textarea {
  min-height: 116px;
  padding: 14px 16px;
  resize: vertical;
}

.field-input:focus,
.field-textarea:focus,
.select-shell:focus-within,
.search-shell:focus-within {
  border-color: #60a5fa;
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.14);
}

.capacity-row .field-input {
  flex: 1;
  min-width: 0;
}

.unit-shell {
  flex: 0 0 128px;
  min-width: 128px;
}

.toggle-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 14px;
  align-items: start;
  padding: 4px 0;
}

.switch-button {
  position: relative;
  width: 56px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: #dbe5f0;
  cursor: pointer;
  transition: background 0.2s ease;
}

.switch-button span {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: #fff;
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.14);
  transition: transform 0.2s ease;
}

.switch-button.active {
  background: linear-gradient(135deg, #60a5fa, #2563eb);
}

.switch-button.active span {
  transform: translateX(24px);
}

.members-toolbar {
  justify-content: space-between;
}

.search-shell {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  flex: 1;
  padding: 0 16px;
  border: 1px solid #d8e1eb;
  border-radius: 16px;
  background: #fff;
  color: #64748b;
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.08);
}

.search-shell input {
  width: 100%;
  border: none;
  background: transparent;
  color: #111827;
  font-size: 15px;
  outline: none;
}

.filter-shell {
  min-width: 150px;
}

.members-shell {
  display: grid;
  gap: 10px;
}

.members-head,
.member-row {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) minmax(220px, 1.1fr) 120px minmax(220px, 1.1fr);
  gap: 12px;
  align-items: center;
}

.members-head {
  padding: 0 0 10px;
  border-bottom: 1px solid #ecf1f6;
  color: #94a3b8;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.members-list {
  display: grid;
  gap: 10px;
}

.member-row {
  padding: 14px 16px;
  border: 1px solid #e8edf3;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.88);
}

.member-main {
  display: grid;
  gap: 4px;
}

.member-main strong {
  color: #111827;
  font-size: 15px;
}

.member-main small,
.member-email,
.empty-state,
.confirm-copy p {
  color: #64748b;
  font-size: 14px;
  line-height: 1.7;
}

.member-storage {
  display: grid;
  gap: 8px;
}

.member-storage-copy {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: #475569;
  font-size: 13px;
}

.member-storage-copy strong {
  color: #1d4ed8;
  font-size: 13px;
}

.progress-track {
  position: relative;
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: #e6edf5;
}

.progress-fill {
  position: absolute;
  inset: 0 auto 0 0;
  border-radius: 999px;
  background: linear-gradient(90deg, #38bdf8, #2563eb);
}

.empty-state {
  padding: 18px 16px;
  border: 1px dashed #d8e1eb;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.9);
}

.confirm-copy {
  display: grid;
  gap: 8px;
}

.confirm-copy p {
  margin: 0;
}

.dialog-actions {
  justify-content: flex-end;
}

.dialog-button {
  min-width: 132px;
}

@media (max-width: 1280px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .table-row {
    grid-template-columns: 70px minmax(220px, 1.3fr) minmax(150px, 1fr) 110px 130px 270px;
  }
}

@media (max-width: 900px) {
  .page-shell,
  .dialog-panel {
    padding: 22px 18px 24px;
  }

  .page-header-copy h1,
  .dialog-header h2 {
    font-size: 38px;
  }

  .metric-grid {
    grid-template-columns: 1fr;
  }

  .toolbar,
  .table-footer,
  .dialog-actions,
  .capacity-row,
  .members-toolbar {
    display: grid;
    grid-template-columns: 1fr;
  }

  .table-card {
    overflow-x: auto;
  }

  .table-row,
  .members-head,
  .member-row {
    min-width: 1080px;
  }

  .toggle-field {
    grid-template-columns: 1fr;
  }
}
</style>

