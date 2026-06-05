<template>
  <section class="users-page">
    <div class="page-shell">
      <header class="hero-card">
        <div class="hero-copy">
          <p class="hero-eyebrow">User Console</p>
          <h1>用户管理</h1>
          <p>
            统一查看账号状态、用户组归属、管理员比例与容量占用，支持批量启用、批量禁用、批量改组和角色切换。
          </p>
        </div>

        <div class="hero-actions">
          <button class="primary-button" type="button" @click="openCreateDialog">新建用户</button>
          <button class="ghost-button" type="button" :disabled="loading" @click="refreshUsers">
            {{ loading ? '刷新中...' : '刷新列表' }}
          </button>
        </div>
      </header>

      <section class="metric-grid">
        <button
          v-for="metric in metrics"
          :key="metric.key"
          class="metric-card"
          :class="{ active: activeMetricKey === metric.key }"
          type="button"
          @click="applyMetricFilter(metric.key)"
        >
          <span class="metric-label">{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
          <small>{{ metric.detail }}</small>
          <em>{{ metric.actionText }}</em>
        </button>
      </section>

      <section class="workspace-card">
        <div class="workspace-main">
          <section class="panel-card filter-card">
            <div class="section-head">
              <div>
                <h2>筛选与批量操作</h2>
                <p>通过搜索、角色、状态和用户组快速定位账号，再执行批量动作。</p>
              </div>

              <button v-if="hasActiveFilters" class="ghost-button slim-button" type="button" @click="resetFilters">
                清空筛选
              </button>
            </div>

            <div class="filter-grid">
              <label class="field-block field-span-2">
                <span>搜索用户</span>
                <input v-model.trim="searchKeyword" class="field-input" type="text" placeholder="输入用户名、邮箱或用户组" />
                <small>支持按用户名、邮箱、用户组名称模糊搜索。</small>
              </label>

              <label class="field-block">
                <span>角色筛选</span>
                <select v-model="roleFilter" class="field-input field-select">
                  <option value="all">全部角色</option>
                  <option value="admin">管理员</option>
                  <option value="user">普通用户</option>
                </select>
                <small>区分后台管理账号与普通成员。</small>
              </label>

              <label class="field-block">
                <span>状态筛选</span>
                <select v-model="statusFilter" class="field-input field-select">
                  <option value="all">全部状态</option>
                  <option value="enabled">已启用</option>
                  <option value="disabled">已禁用</option>
                </select>
                <small>快速定位停用或可用账号。</small>
              </label>

              <label class="field-block field-span-2">
                <span>用户组筛选</span>
                <select v-model="groupFilter" class="field-input field-select">
                  <option value="all">全部用户组</option>
                  <option v-for="group in groupOptions" :key="group.id" :value="group.name">
                    {{ group.name }}
                  </option>
                </select>
                <small>联动查看各用户组当前成员分布。</small>
              </label>
            </div>

            <div v-if="selectedIds.length" class="batch-bar">
              <div class="batch-copy">
                <strong>已选中 {{ selectedIds.length }} 个用户</strong>
                <small>支持一次性改组、切换角色、启用、禁用和删除。</small>
              </div>

              <div class="batch-actions">
                <select v-model.number="batchGroupId" class="field-input compact-input field-select">
                  <option :value="0">选择用户组</option>
                  <option v-for="group in groupOptions" :key="group.id" :value="group.id">
                    {{ group.name }}
                  </option>
                </select>
                <button class="ghost-button compact-button" type="button" :disabled="batchSaving" @click="applyBatchGroup">
                  {{ batchSaving ? '应用中...' : '批量改组' }}
                </button>

                <select v-model="batchRole" class="field-input compact-input field-select">
                  <option value="user">普通用户</option>
                  <option value="admin">管理员</option>
                </select>
                <button class="ghost-button compact-button" type="button" :disabled="batchRoleSaving" @click="applyBatchRole">
                  {{ batchRoleSaving ? '切换中...' : '批量角色' }}
                </button>

                <button class="ghost-button compact-button" type="button" :disabled="batchStatusSaving" @click="applyBatchStatus(true)">
                  {{ batchStatusSaving && pendingBatchStatus === true ? '启用中...' : '批量启用' }}
                </button>
                <button class="ghost-button compact-button warning-button" type="button" :disabled="batchStatusSaving" @click="applyBatchStatus(false)">
                  {{ batchStatusSaving && pendingBatchStatus === false ? '禁用中...' : '批量禁用' }}
                </button>

                <button class="danger-button compact-button" type="button" :disabled="batchDeleting" @click="openBatchDeleteDialog">
                  {{ batchDeleting ? '删除中...' : '批量删除' }}
                </button>
              </div>
            </div>
          </section>

          <section class="table-card">
            <div class="table-toolbar">
              <div>
                <h3>用户列表</h3>
                <p>共 {{ totalUsers }} 个匹配结果，点击用户名可在右侧查看详情。</p>
              </div>
            </div>

            <div class="table-row table-head">
              <div class="col-check">
                <label class="check-shell">
                  <input :checked="allPageSelected" type="checkbox" @change="toggleSelectPage($event)" />
                  <span></span>
                </label>
              </div>
              <div class="col-user">用户</div>
              <div class="col-contact">联系信息</div>
              <div class="col-role">角色</div>
              <div class="col-status">状态</div>
              <div class="col-group">用户组</div>
              <div class="col-usage">容量使用</div>
              <div class="col-created">创建时间</div>
              <div class="col-actions">操作</div>
            </div>

            <div v-if="loading" class="state-shell">正在加载用户列表...</div>
            <div v-else-if="!pagedUsers.length" class="state-shell">当前筛选条件下没有匹配用户。</div>

            <div v-else>
              <article v-for="user in pagedUsers" :key="user.id" class="table-row table-body-row">
                <div class="col-check">
                  <label class="check-shell">
                    <input :checked="selectedIds.includes(user.id)" type="checkbox" @change="toggleSelectUser(user.id, $event)" />
                    <span></span>
                  </label>
                </div>

                <div class="col-user">
                  <button class="user-cell" type="button" @click="openUserDetail(user)">
                    <span class="avatar" :class="`is-${user.role}`">{{ user.username.slice(0, 1).toUpperCase() }}</span>
                    <span class="user-copy">
                      <strong>{{ user.username }}</strong>
                      <small>ID #{{ user.id }}</small>
                    </span>
                  </button>
                </div>

                <div class="col-contact">
                  <div class="info-stack">
                    <strong>{{ user.email }}</strong>
                    <small>{{ formatDate(user.updated_at, true) }} 更新</small>
                  </div>
                </div>

                <div class="col-role">
                  <div class="access-stack">
                    <span class="pill" :class="`role-${user.role}`">{{ roleLabel(user.role) }}</span>
                    <span class="access-chip" :class="`is-${userTierClass(user)}`">{{ userTierLabel(user) }}</span>
                  </div>
                </div>

                <div class="col-status">
                  <button
                    class="status-toggle"
                    :class="{ active: user.enabled }"
                    type="button"
                    :disabled="statusUpdatingId === user.id"
                    @click="toggleUserStatus(user)"
                  >
                    <span class="status-toggle-track">
                      <span class="status-toggle-knob"></span>
                    </span>
                    <span class="status-text">{{ user.enabled ? '正常' : '受限' }}</span>
                  </button>
                </div>

                <div class="col-group">
                  <button class="group-pill" type="button" @click="groupFilter = user.user_group_name">
                    {{ user.user_group_name }}
                  </button>
                </div>

                <div class="col-usage">
                  <div class="usage-stack">
                    <div class="usage-top">
                      <span>{{ formatCapacity(user.used_size) }} / {{ formatCapacity(user.capacity) }}</span>
                      <strong>{{ usagePercent(user) }}%</strong>
                    </div>
                    <div class="progress-track">
                      <span class="progress-fill" :style="{ width: `${usagePercent(user)}%` }"></span>
                    </div>
                  </div>
                </div>

                <div class="col-created">
                  <span class="date-text">{{ formatDate(user.created_at) }}</span>
                </div>

                <div class="col-actions">
                  <div class="control-dock" aria-label="用户操作">
                    <button class="dock-button" type="button" title="详情" aria-label="查看详情" @click="openUserDetail(user)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <circle cx="10" cy="10" r="3.2" />
                        <path d="M2.8 10s2.6-5 7.2-5 7.2 5 7.2 5-2.6 5-7.2 5-7.2-5-7.2-5Z" />
                      </svg>
                    </button>
                    <button class="dock-button" type="button" title="编辑" aria-label="编辑用户" @click="openEditDialog(user)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M4 13.5V16h2.5L15 7.5 12.5 5 4 13.5Z" />
                        <path d="m11.8 5.7 2.5 2.5" />
                      </svg>
                    </button>
                    <button class="dock-button" type="button" title="重置密码" aria-label="重置密码" @click="openResetPasswordDialog(user)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M6.5 8V6.5a3.5 3.5 0 0 1 7 0V8" />
                        <path d="M5 8h10v8H5z" />
                        <path d="M10 11.5v2" />
                      </svg>
                    </button>
                    <button class="dock-button danger" type="button" title="删除" aria-label="删除用户" @click="openDeleteDialog(user)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M6.5 6.5v8.5" />
                        <path d="M10 6.5v8.5" />
                        <path d="M13.5 6.5v8.5" />
                        <path d="M4.5 5.5h11" />
                        <path d="M7.8 5.5V4.2h4.4v1.3" />
                      </svg>
                    </button>
                  </div>
                </div>
              </article>
            </div>
          </section>

          <footer class="footer-bar">
            <div class="pager">
              <button class="pager-button" type="button" :disabled="page === 1" @click="page = Math.max(1, page - 1)">上一页</button>
              <span class="pager-index">{{ page }} / {{ totalPages }}</span>
              <button class="pager-button" type="button" :disabled="page >= totalPages" @click="page = Math.min(totalPages, page + 1)">下一页</button>
            </div>

            <select v-model.number="pageSize" class="field-input footer-select field-select">
              <option :value="10">每页 10 条</option>
              <option :value="20">每页 20 条</option>
              <option :value="50">每页 50 条</option>
            </select>
          </footer>
        </div>
      </section>

      <section class="detail-bottom-card">
        <aside class="detail-drawer bottom-layout" :class="{ empty: !detailUser }">
          <template v-if="detailUser">
            <div class="detail-band-top">
              <div class="drawer-head">
                <div class="drawer-user">
                  <span class="avatar large" :class="`is-${detailUser.role}`">{{ detailUser.username.slice(0, 1).toUpperCase() }}</span>
                  <div>
                    <p class="hero-eyebrow drawer-eyebrow">User Detail</p>
                    <h3>{{ detailUser.username }}</h3>
                    <p>{{ detailUser.email }}</p>
                  </div>
                </div>
                <button class="drawer-close" type="button" @click="closeUserDetail">×</button>
              </div>

              <div class="drawer-highlight">
                <div class="highlight-item">
                  <span>角色</span>
                  <strong>{{ roleLabel(detailUser.role) }}</strong>
                </div>
                <div class="highlight-item">
                  <span>状态</span>
                  <strong>{{ detailUser.enabled ? '已启用' : '已禁用' }}</strong>
                </div>
                <div class="highlight-item">
                  <span>用户组</span>
                  <strong>{{ detailUser.user_group_name }}</strong>
                </div>
              </div>

              <div class="drawer-section compact-usage">
                <div class="drawer-section-head">
                  <h4>容量概览</h4>
                  <strong>{{ usagePercent(detailUser) }}%</strong>
                </div>
                <div class="usage-card">
                  <div class="usage-top">
                    <span>{{ formatCapacity(detailUser.used_size) }}</span>
                    <span>{{ formatCapacity(detailUser.capacity) }}</span>
                  </div>
                  <div class="progress-track large">
                    <span class="progress-fill" :style="{ width: `${usagePercent(detailUser)}%` }"></span>
                  </div>
                </div>
              </div>

              <div class="drawer-actions">
                <button class="toolbar-action" type="button" @click="openEditDialog(detailUser)">
                  <span class="toolbar-icon">E</span>
                  <span>编辑用户</span>
                </button>
                <button class="toolbar-action" type="button" @click="openResetPasswordDialog(detailUser)">
                  <span class="toolbar-icon">P</span>
                  <span>重置密码</span>
                </button>
                <button class="toolbar-action" type="button" @click="toggleUserStatus(detailUser)">
                  <span class="toolbar-icon">{{ detailUser.enabled ? 'O' : 'I' }}</span>
                  <span>{{ detailUser.enabled ? '禁用账号' : '启用账号' }}</span>
                </button>
              </div>
            </div>

            <div class="drawer-grid metrics-strip">
              <article class="detail-card">
                <span>创建时间</span>
                <strong>{{ formatDate(detailUser.created_at, true) }}</strong>
              </article>
              <article class="detail-card">
                <span>最近更新</span>
                <strong>{{ formatDate(detailUser.updated_at, true) }}</strong>
              </article>
              <article class="detail-card">
                <span>用户 ID</span>
                <strong>#{{ detailUser.id }}</strong>
              </article>
              <article class="detail-card">
                <span>存储配额</span>
                <strong>{{ formatCapacity(detailUser.capacity) }}</strong>
              </article>
              <article class="detail-card">
                <span>已用空间</span>
                <strong>{{ formatCapacity(detailUser.used_size) }}</strong>
              </article>
              <article class="detail-card">
                <span>账号状态</span>
                <strong>{{ detailUser.enabled ? '正常可用' : '已被禁用' }}</strong>
              </article>
            </div>
          </template>

          <template v-else>
            <div class="drawer-empty">
              <span class="drawer-empty-icon">U</span>
              <h3>选择一个用户</h3>
              <p>点击左侧列表中的任意用户，即可在这里查看详细信息、容量状态和快捷操作。</p>
            </div>
          </template>
        </aside>
      </section>
    </div>

    <div v-if="dialogOpen" class="dialog-mask" @click.self="closeDialog">
      <div class="dialog-panel form-panel">
        <div class="dialog-head">
          <div>
            <p class="dialog-eyebrow">{{ dialogMode === 'create' ? 'Create User' : 'Edit User' }}</p>
            <h3>{{ dialogMode === 'create' ? '新建用户' : `编辑 ${form.username || '用户'}` }}</h3>
          </div>
          <button class="dialog-close" type="button" @click="closeDialog">×</button>
        </div>

        <div class="form-grid">
          <label class="field-block">
            <span>用户名</span>
            <input v-model.trim="form.username" class="field-input" type="text" placeholder="例如 demo_user" />
            <small>用于登录与后台识别。</small>
          </label>

          <label class="field-block">
            <span>邮箱</span>
            <input v-model.trim="form.email" class="field-input" type="email" placeholder="例如 demo@example.com" />
            <small>用于登录、通知与找回密码。</small>
          </label>

          <label class="field-block">
            <span>{{ dialogMode === 'create' ? '登录密码' : '重置密码' }}</span>
            <input
              v-model.trim="form.password"
              class="field-input"
              type="password"
              :placeholder="dialogMode === 'create' ? '至少 6 位密码' : '留空则保持当前密码不变'"
            />
            <small>{{ dialogMode === 'create' ? '新建用户时必填。' : '仅在需要重设密码时填写。' }}</small>
          </label>

          <label class="field-block">
            <span>角色</span>
            <select v-model="form.role" class="field-input field-select">
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
            <small>管理员拥有后台管理权限。</small>
          </label>

          <label class="field-block">
            <span>启用状态</span>
            <select v-model="form.enabledText" class="field-input field-select">
              <option value="enabled">已启用</option>
              <option value="disabled">已禁用</option>
            </select>
            <small>禁用后将阻止该账号登录。</small>
          </label>

          <label class="field-block">
            <span>用户组</span>
            <select v-model.number="form.user_group_id" class="field-input field-select">
              <option v-for="group in groupOptions" :key="group.id" :value="group.id">
                {{ group.name }}
              </option>
            </select>
            <small>决定默认容量和归属策略。</small>
          </label>

          <label class="field-block field-span-2">
            <span>容量配额</span>
            <div class="capacity-mode-row">
              <label class="mode-pill">
                <input v-model="form.capacityMode" type="radio" value="group" />
                <span>跟随用户组容量</span>
              </label>
              <label class="mode-pill">
                <input v-model="form.capacityMode" type="radio" value="custom" />
                <span>自定义容量</span>
              </label>
            </div>
            <div class="capacity-row">
              <input v-model.number="form.capacityValue" class="field-input" type="number" min="0" placeholder="50" :disabled="form.capacityMode === 'group'" />
              <select v-model="form.capacityUnit" class="field-input field-select unit-select" :disabled="form.capacityMode === 'group'">
                <option value="MB">MB</option>
                <option value="GB">GB</option>
                <option value="TB">TB</option>
              </select>
            </div>
            <small>{{ capacityModeHint }}</small>
          </label>
        </div>

        <div class="dialog-actions">
          <button class="ghost-button" type="button" @click="closeDialog">取消</button>
          <button class="primary-button" type="button" :disabled="saving" @click="submitDialog">
            {{ saving ? '保存中...' : dialogMode === 'create' ? '创建用户' : '保存修改' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="resetDialogOpen && resetTargetUser" class="dialog-mask" @click.self="closeResetPasswordDialog">
      <div class="dialog-panel confirm-panel">
        <div class="dialog-head">
          <div>
            <p class="dialog-eyebrow">Reset Password</p>
            <h3>重置 {{ resetTargetUser.username }} 的密码</h3>
          </div>
          <button class="dialog-close" type="button" @click="closeResetPasswordDialog">×</button>
        </div>

        <label class="field-block">
          <span>新密码</span>
          <input v-model.trim="resetPasswordValue" class="field-input" type="password" placeholder="请输入至少 6 位的新密码" />
          <small>提交后立即写入后端。</small>
        </label>

        <div class="dialog-actions">
          <button class="ghost-button" type="button" @click="closeResetPasswordDialog">取消</button>
          <button class="primary-button" type="button" :disabled="resettingPassword" @click="submitResetPassword">
            {{ resettingPassword ? '提交中...' : '确认重置' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="deleteDialogOpen && pendingDeleteUser" class="dialog-mask" @click.self="closeDeleteDialog">
      <div class="dialog-panel confirm-panel">
        <div class="dialog-head">
          <div>
            <p class="dialog-eyebrow">Delete User</p>
            <h3>删除用户 {{ pendingDeleteUser.username }}</h3>
          </div>
          <button class="dialog-close" type="button" @click="closeDeleteDialog">×</button>
        </div>

        <p class="confirm-copy">安全模式会先检查该用户是否仍有文件、分享、任务或外部账号等资产。</p>
        <div v-if="deletePreviewLoading" class="state-shell compact-state">正在检查账号资产...</div>
        <div v-else-if="pendingDeletePreview" class="delete-preview" :class="{ blocked: pendingDeletePreview.has_blocking_assets }">
          <strong>{{ pendingDeletePreview.has_blocking_assets ? '该账号仍有资产，不能删除' : '该账号无阻塞资产，可以删除' }}</strong>
          <div class="delete-preview-grid">
            <span>文件 {{ pendingDeletePreview.file_count }}</span>
            <span>分享 {{ pendingDeletePreview.share_count }}</span>
            <span>回收站 {{ pendingDeletePreview.recycle_bin_count }}</span>
            <span>离线任务 {{ pendingDeletePreview.offline_download_task_count }}</span>
            <span>分片任务 {{ pendingDeletePreview.multipart_upload_count }}</span>
            <span>DAV {{ pendingDeletePreview.dav_account_count }}</span>
            <span>OAuth {{ pendingDeletePreview.oauth_credential_count }}</span>
            <span>协作 {{ pendingDeletePreview.collaboration_owned_count + pendingDeletePreview.collaboration_shared_count }}</span>
            <span>已用 {{ formatCapacity(pendingDeletePreview.used_size) }}</span>
          </div>
        </div>

        <div class="dialog-actions">
          <button class="ghost-button" type="button" @click="closeDeleteDialog">取消</button>
          <button class="danger-button" type="button" :disabled="deleting || deletePreviewLoading || !!pendingDeletePreview?.has_blocking_assets" @click="removeUserConfirmed">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="batchDeleteDialogOpen" class="dialog-mask" @click.self="closeBatchDeleteDialog">
      <div class="dialog-panel confirm-panel">
        <div class="dialog-head">
          <div>
            <p class="dialog-eyebrow">Batch Delete</p>
            <h3>批量删除所选用户</h3>
          </div>
          <button class="dialog-close" type="button" @click="closeBatchDeleteDialog">×</button>
        </div>

        <p class="confirm-copy">将删除当前选中的 {{ selectedIds.length }} 个用户，请确认后再继续。</p>

        <div class="dialog-actions">
          <button class="ghost-button" type="button" @click="closeBatchDeleteDialog">取消</button>
          <button class="danger-button" type="button" :disabled="batchDeleting" @click="removeUsersBatchConfirmed">
            {{ batchDeleting ? '删除中...' : '确认批量删除' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import {
  batchDeleteAdminUsers,
  batchUpdateAdminUserStatus,
  batchUpdateAdminUsersGroup,
  batchUpdateAdminUsersRole,
  createAdminUser,
  deleteAdminUser,
  getAdminUserDeletePreview,
  listAdminUsers,
  resetAdminUserPassword,
  updateAdminUser,
  updateAdminUserStatus,
  type AdminUserDeletePreviewPayload,
  type AdminUserPayload,
  type ListAdminUsersParams,
} from '@/api/admin-users';
import { listUserGroups, type UserGroupPayload } from '@/api/user-groups';

type UserRecord = AdminUserPayload;
type DialogMode = 'create' | 'edit';
type GroupOption = UserGroupPayload;
type MetricKey = 'all' | 'admins' | 'disabled' | 'group-hot';

const loading = ref(false);
const saving = ref(false);
const deleting = ref(false);
const deletePreviewLoading = ref(false);
const batchSaving = ref(false);
const batchDeleting = ref(false);
const batchRoleSaving = ref(false);
const batchStatusSaving = ref(false);
const resettingPassword = ref(false);
const statusUpdatingId = ref<number | null>(null);
const pendingBatchStatus = ref<boolean | null>(null);
const page = ref(1);
const pageSize = ref(10);
const totalUsers = ref(0);
const users = ref<UserRecord[]>([]);
const groups = ref<GroupOption[]>([]);
const selectedIds = ref<number[]>([]);
const batchGroupId = ref(0);
const batchRole = ref<'admin' | 'user'>('user');
const searchKeyword = ref('');
const roleFilter = ref<'all' | 'admin' | 'user'>('all');
const statusFilter = ref<'all' | 'enabled' | 'disabled'>('all');
const groupFilter = ref('all');
const activeMetricKey = ref<MetricKey>('all');
const detailUser = ref<UserRecord | null>(null);
const dialogOpen = ref(false);
const dialogMode = ref<DialogMode>('create');
const editingUserId = ref<number | null>(null);
const deleteDialogOpen = ref(false);
const pendingDeleteUser = ref<UserRecord | null>(null);
const pendingDeletePreview = ref<AdminUserDeletePreviewPayload | null>(null);
const resetDialogOpen = ref(false);
const resetTargetUser = ref<UserRecord | null>(null);
const resetPasswordValue = ref('');
const batchDeleteDialogOpen = ref(false);

const form = reactive({
  username: '',
  email: '',
  password: '',
  role: 'user' as 'admin' | 'user',
  enabledText: 'enabled' as 'enabled' | 'disabled',
  user_group_id: 0,
  capacityMode: 'group' as 'group' | 'custom',
  capacityValue: 0,
  capacityUnit: 'GB' as 'MB' | 'GB' | 'TB',
});

const groupOptions = computed(() => groups.value);
const selectedFormGroup = computed(() => groupOptions.value.find((group) => group.id === form.user_group_id));
const capacityModeHint = computed(() => {
  if (form.capacityMode === 'group') {
    return `跟随用户组容量：${formatCapacity(selectedFormGroup.value?.max_capacity || 0)}。用户组容量为 0 时表示无限容量。`;
  }
  return '自定义容量：填写 0 表示无限容量，不会被用户组容量覆盖。';
});

const groupMemberMap = computed(() => {
  const map = new Map<string, number>();
  users.value.forEach((user) => {
    map.set(user.user_group_name, (map.get(user.user_group_name) || 0) + 1);
  });
  return map;
});

const hottestGroup = computed(() => {
  let currentName = '';
  let currentCount = 0;
  groupMemberMap.value.forEach((count, name) => {
    if (count > currentCount) {
      currentName = name;
      currentCount = count;
    }
  });
  return { name: currentName, count: currentCount };
});

const totalPages = computed(() => Math.max(1, Math.ceil(totalUsers.value / pageSize.value)));
const pagedUsers = computed(() => users.value);
const allPageSelected = computed(
  () => pagedUsers.value.length > 0 && pagedUsers.value.every((item) => selectedIds.value.includes(item.id)),
);

const hasActiveFilters = computed(
  () =>
    searchKeyword.value.trim() !== '' ||
    roleFilter.value !== 'all' ||
    statusFilter.value !== 'all' ||
    groupFilter.value !== 'all' ||
    activeMetricKey.value !== 'all',
);

const metrics = computed(() => {
  const total = totalUsers.value;
  const admins = users.value.filter((user) => user.role === 'admin').length;
  const disabled = users.value.filter((user) => !user.enabled).length;
  const topGroup = hottestGroup.value;

  return [
    {
      key: 'all' as MetricKey,
      label: '全部用户',
      value: total,
      detail: `当前结果 · ${groups.value.length} 个用户组正在使用`,
      actionText: '查看所有账号',
    },
    {
      key: 'admins' as MetricKey,
      label: '管理员人数',
      value: admins,
      detail: '当前结果 · 点击后仅显示管理员账号',
      actionText: '筛选管理员',
    },
    {
      key: 'disabled' as MetricKey,
      label: '禁用人数',
      value: disabled,
      detail: '当前结果 · 点击后仅显示已禁用账号',
      actionText: '筛选禁用用户',
    },
    {
      key: 'group-hot' as MetricKey,
      label: '最大用户组',
      value: topGroup.count || 0,
      detail: topGroup.name ? `当前结果 · ${topGroup.name} 成员最多` : '当前结果 · 暂无分组数据',
      actionText: topGroup.name ? `查看 ${topGroup.name}` : '等待数据',
    },
  ];
});

watch([searchKeyword, roleFilter, statusFilter, groupFilter], () => {
  if (page.value !== 1) {
    page.value = 1;
    return;
  }
  void fetchUsers();
});

watch(pageSize, () => {
  if (page.value !== 1) {
    page.value = 1;
    return;
  }
  void fetchUsers();
});

watch(page, () => {
  void fetchUsers();
});

watch(totalPages, (value) => {
  if (page.value > value) {
    page.value = value;
  }
});

watch(
  () => [form.capacityMode, form.user_group_id] as const,
  () => {
    if (form.capacityMode === 'group') {
      fillCapacity(selectedFormGroup.value?.max_capacity || 0);
    }
  },
);

function roleLabel(role: string) {
  return role === 'admin' ? '管理员' : '普通用户';
}

function userTierClass(user: UserRecord) {
  if (!user.enabled) {
    return 'restricted';
  }
  if (user.role === 'admin') {
    return 'admin';
  }
  if (/pro|vip|会员|高级|高配|套餐/i.test(user.user_group_name || '')) {
    return 'pro';
  }
  return 'normal';
}

function userTierLabel(user: UserRecord) {
  const tier = userTierClass(user);
  if (tier === 'admin') return 'Administrator';
  if (tier === 'pro') return 'Pro Member';
  if (tier === 'restricted') return '受限';
  return '正常';
}

function formatDate(value: string, withTime = false) {
  if (!value) return '--';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return withTime ? date.toLocaleString('zh-CN') : date.toLocaleDateString('zh-CN');
}

function formatCapacity(value: number) {
  if (!value) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  return `${size >= 100 ? size.toFixed(0) : size.toFixed(size >= 10 ? 1 : 2)} ${units[index]}`;
}

function usagePercent(user: UserRecord) {
  if (!user.capacity || user.capacity <= 0) return 0;
  return Math.min(100, Math.max(0, Math.round((user.used_size / user.capacity) * 100)));
}

function bytesFromUnit(value: number, unit: 'MB' | 'GB' | 'TB') {
  const base = unit === 'TB' ? 1024 ** 4 : unit === 'GB' ? 1024 ** 3 : 1024 ** 2;
  return Math.max(0, Math.round((value || 0) * base));
}

function fillCapacity(value: number) {
  if (!value) {
    form.capacityValue = 0;
    form.capacityUnit = 'GB';
    return;
  }
  if (value % 1024 ** 4 === 0) {
    form.capacityValue = value / 1024 ** 4;
    form.capacityUnit = 'TB';
    return;
  }
  if (value % 1024 ** 3 === 0) {
    form.capacityValue = value / 1024 ** 3;
    form.capacityUnit = 'GB';
    return;
  }
  form.capacityValue = Number((value / 1024 ** 2).toFixed(2));
  form.capacityUnit = 'MB';
}

function syncUpdatedUser(updated: UserRecord) {
  users.value = users.value.map((item) => (item.id === updated.id ? updated : item));
  if (detailUser.value?.id === updated.id) {
    detailUser.value = updated;
  }
  if (pendingDeleteUser.value?.id === updated.id) {
    pendingDeleteUser.value = updated;
  }
  if (resetTargetUser.value?.id === updated.id) {
    resetTargetUser.value = updated;
  }
}

function syncUpdatedUsers(updated: UserRecord[]) {
  updated.forEach(syncUpdatedUser);
}

function resetForm() {
  form.username = '';
  form.email = '';
  form.password = '';
  form.role = 'user';
  form.enabledText = 'enabled';
  form.user_group_id = groupOptions.value[0]?.id || 0;
  form.capacityMode = 'group';
  fillCapacity(selectedFormGroup.value?.max_capacity || 0);
}

function fillForm(user: UserRecord) {
  form.username = user.username;
  form.email = user.email;
  form.password = '';
  form.role = user.role;
  form.enabledText = user.enabled ? 'enabled' : 'disabled';
  form.user_group_id = user.user_group_id;
  form.capacityMode = user.capacity === (selectedFormGroup.value?.max_capacity || 0) ? 'group' : 'custom';
  fillCapacity(user.capacity);
}

function buildUpsertPayload() {
  return {
    username: form.username.trim(),
    email: form.email.trim(),
    password: form.password.trim(),
    role: form.role,
    enabled: form.enabledText === 'enabled',
    user_group_id: form.user_group_id,
    capacity: bytesFromUnit(form.capacityValue, form.capacityUnit),
    follow_group_capacity: form.capacityMode === 'group',
  };
}

function clearSelection() {
  selectedIds.value = [];
  batchGroupId.value = 0;
  batchRole.value = 'user';
}

function resetFilters() {
  activeMetricKey.value = 'all';
  searchKeyword.value = '';
  roleFilter.value = 'all';
  statusFilter.value = 'all';
  groupFilter.value = 'all';
}

function applyMetricFilter(key: MetricKey) {
  activeMetricKey.value = key;

  if (key === 'all') {
    resetFilters();
    activeMetricKey.value = 'all';
    return;
  }

  searchKeyword.value = '';
  roleFilter.value = 'all';
  statusFilter.value = 'all';
  groupFilter.value = 'all';

  if (key === 'admins') {
    roleFilter.value = 'admin';
  }
  if (key === 'disabled') {
    statusFilter.value = 'disabled';
  }
  if (key === 'group-hot' && hottestGroup.value.name) {
    groupFilter.value = hottestGroup.value.name;
  }
}

async function fetchGroups() {
  const data = await listUserGroups();
  groups.value = data;
  if (!form.user_group_id && data.length) {
    form.user_group_id = data[0].id || 0;
  }
}

async function fetchUsers(showSuccess = false) {
  loading.value = true;
  try {
    const params: ListAdminUsersParams = {
      keyword: searchKeyword.value.trim() || undefined,
      role: roleFilter.value,
      status: statusFilter.value,
      user_group_name: groupFilter.value === 'all' ? undefined : groupFilter.value,
      page: page.value,
      page_size: pageSize.value,
    };
    const data = await listAdminUsers(params);
    users.value = data.items || [];
    totalUsers.value = data.total || 0;
    page.value = data.page || page.value;
    pageSize.value = data.page_size || pageSize.value;
    selectedIds.value = selectedIds.value.filter((id) => users.value.some((item) => item.id === id));
    if (!detailUser.value && users.value.length) {
      detailUser.value = users.value[0];
    }
    if (detailUser.value) {
      const latest = users.value.find((item) => item.id === detailUser.value?.id);
      detailUser.value = latest || users.value[0] || null;
    }
    if (showSuccess) {
      ElMessage.success('用户列表已刷新');
    }
  } finally {
    loading.value = false;
  }
}

async function refreshUsers() {
  try {
    await Promise.all([fetchGroups(), fetchUsers(true)]);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '刷新用户列表失败');
  }
}

function toggleSelectUser(id: number, event: Event) {
  const checked = (event.target as HTMLInputElement).checked;
  if (checked) {
    selectedIds.value = Array.from(new Set([...selectedIds.value, id]));
    return;
  }
  selectedIds.value = selectedIds.value.filter((item) => item !== id);
}

function toggleSelectPage(event: Event) {
  const checked = (event.target as HTMLInputElement).checked;
  const pageIds = pagedUsers.value.map((item) => item.id);
  if (checked) {
    selectedIds.value = Array.from(new Set([...selectedIds.value, ...pageIds]));
    return;
  }
  selectedIds.value = selectedIds.value.filter((id) => !pageIds.includes(id));
}

async function applyBatchGroup() {
  if (!selectedIds.value.length) {
    ElMessage.warning('请先勾选至少一个用户');
    return;
  }
  if (!batchGroupId.value) {
    ElMessage.warning('请选择目标用户组');
    return;
  }

  batchSaving.value = true;
  try {
    const updated = await batchUpdateAdminUsersGroup({
      ids: selectedIds.value,
      user_group_id: batchGroupId.value,
    });
    syncUpdatedUsers(updated);
    clearSelection();
    await fetchUsers();
    ElMessage.success('已批量更新用户组');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '批量改组失败');
  } finally {
    batchSaving.value = false;
  }
}

async function applyBatchRole() {
  if (!selectedIds.value.length) {
    ElMessage.warning('请先勾选至少一个用户');
    return;
  }

  batchRoleSaving.value = true;
  try {
    const updated = await batchUpdateAdminUsersRole({
      ids: selectedIds.value,
      role: batchRole.value,
    });
    syncUpdatedUsers(updated);
    clearSelection();
    await fetchUsers();
    ElMessage.success('已批量切换用户角色');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '批量切换角色失败');
  } finally {
    batchRoleSaving.value = false;
  }
}

async function applyBatchStatus(enabled: boolean) {
  if (!selectedIds.value.length) {
    ElMessage.warning('请先勾选至少一个用户');
    return;
  }

  pendingBatchStatus.value = enabled;
  batchStatusSaving.value = true;
  try {
    const updated = await batchUpdateAdminUserStatus({
      ids: selectedIds.value,
      enabled,
    });
    syncUpdatedUsers(updated);
    clearSelection();
    await fetchUsers();
    ElMessage.success(enabled ? '已批量启用用户' : '已批量禁用用户');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '批量更新用户状态失败');
  } finally {
    batchStatusSaving.value = false;
    pendingBatchStatus.value = null;
  }
}

async function toggleUserStatus(user: UserRecord) {
  statusUpdatingId.value = user.id;
  try {
    const updated = await updateAdminUserStatus(user.id, { enabled: !user.enabled });
    syncUpdatedUser(updated);
    await fetchUsers();
    ElMessage.success(updated.enabled ? '用户已启用' : '用户已禁用');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新用户状态失败');
  } finally {
    statusUpdatingId.value = null;
  }
}

function openUserDetail(user: UserRecord) {
  detailUser.value = user;
}

function closeUserDetail() {
  detailUser.value = null;
}

function openCreateDialog() {
  dialogMode.value = 'create';
  editingUserId.value = null;
  resetForm();
  dialogOpen.value = true;
}

function openEditDialog(user: UserRecord) {
  dialogMode.value = 'edit';
  editingUserId.value = user.id;
  fillForm(user);
  dialogOpen.value = true;
}

function closeDialog() {
  if (saving.value) return;
  dialogOpen.value = false;
  editingUserId.value = null;
}

async function submitDialog() {
  if (!form.username.trim()) {
    ElMessage.warning('请输入用户名');
    return;
  }
  if (!form.email.trim()) {
    ElMessage.warning('请输入邮箱');
    return;
  }
  if (dialogMode.value === 'create' && !form.password.trim()) {
    ElMessage.warning('新建用户时必须设置密码');
    return;
  }
  if (!form.user_group_id) {
    ElMessage.warning('请选择用户组');
    return;
  }

  saving.value = true;
  try {
    const payload = buildUpsertPayload();
    if (dialogMode.value === 'create') {
      const created = await createAdminUser(payload);
      users.value.unshift(created);
      detailUser.value = created;
      await fetchUsers();
      ElMessage.success('用户已创建并写入后端');
    } else if (editingUserId.value) {
      const updated = await updateAdminUser(editingUserId.value, payload);
      syncUpdatedUser(updated);
      await fetchUsers();
      ElMessage.success('用户修改已保存');
    }
    dialogOpen.value = false;
    editingUserId.value = null;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存用户失败');
  } finally {
    saving.value = false;
  }
}

function openResetPasswordDialog(user: UserRecord) {
  resetTargetUser.value = user;
  resetPasswordValue.value = '';
  resetDialogOpen.value = true;
}

function closeResetPasswordDialog() {
  if (resettingPassword.value) return;
  resetDialogOpen.value = false;
  resetTargetUser.value = null;
  resetPasswordValue.value = '';
}

async function submitResetPassword() {
  if (!resetTargetUser.value) return;
  if (resetPasswordValue.value.trim().length < 6) {
    ElMessage.warning('新密码至少需要 6 位');
    return;
  }

  resettingPassword.value = true;
  try {
    await resetAdminUserPassword(resetTargetUser.value.id, resetPasswordValue.value.trim());
    ElMessage.success(`已重置 ${resetTargetUser.value.username} 的密码`);
    resetDialogOpen.value = false;
    resetTargetUser.value = null;
    resetPasswordValue.value = '';
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重置密码失败');
  } finally {
    resettingPassword.value = false;
  }
}

async function openDeleteDialog(user: UserRecord) {
  pendingDeleteUser.value = user;
  pendingDeletePreview.value = null;
  deleteDialogOpen.value = true;
  deletePreviewLoading.value = true;
  try {
    pendingDeletePreview.value = await getAdminUserDeletePreview(user.id);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取删除预览失败');
  } finally {
    deletePreviewLoading.value = false;
  }
}

function closeDeleteDialog() {
  if (deleting.value) return;
  deleteDialogOpen.value = false;
  pendingDeleteUser.value = null;
  pendingDeletePreview.value = null;
}

async function removeUserConfirmed() {
  if (!pendingDeleteUser.value) return;
  if (pendingDeletePreview.value?.has_blocking_assets) {
    ElMessage.warning('该用户仍有账号资产，安全模式下不能删除');
    return;
  }
  deleting.value = true;
  try {
    await deleteAdminUser(pendingDeleteUser.value.id);
    users.value = users.value.filter((item) => item.id !== pendingDeleteUser.value?.id);
    selectedIds.value = selectedIds.value.filter((id) => id !== pendingDeleteUser.value?.id);
    if (detailUser.value?.id === pendingDeleteUser.value.id) {
      detailUser.value = users.value[0] || null;
    }
    await fetchUsers();
    ElMessage.success(`已删除用户：${pendingDeleteUser.value.username}`);
    deleteDialogOpen.value = false;
    pendingDeleteUser.value = null;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除用户失败');
  } finally {
    deleting.value = false;
  }
}

function openBatchDeleteDialog() {
  if (!selectedIds.value.length) {
    ElMessage.warning('请先勾选至少一个用户');
    return;
  }
  batchDeleteDialogOpen.value = true;
}

function closeBatchDeleteDialog() {
  if (batchDeleting.value) return;
  batchDeleteDialogOpen.value = false;
}

async function removeUsersBatchConfirmed() {
  if (!selectedIds.value.length) return;
  batchDeleting.value = true;
  try {
    await batchDeleteAdminUsers({ ids: selectedIds.value });
    const idSet = new Set(selectedIds.value);
    users.value = users.value.filter((item) => !idSet.has(item.id));
    if (detailUser.value && idSet.has(detailUser.value.id)) {
      detailUser.value = users.value[0] || null;
    }
    clearSelection();
    await fetchUsers();
    batchDeleteDialogOpen.value = false;
    ElMessage.success('已批量删除所选用户');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '批量删除失败');
  } finally {
    batchDeleting.value = false;
  }
}

onMounted(async () => {
  try {
    await Promise.all([fetchGroups(), fetchUsers()]);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '用户页面加载失败');
  }
});
</script>

<style scoped>
.users-page {
  min-height: calc(100vh - 104px);
  padding: 8px 2px 28px;
}

.page-shell {
  display: grid;
  gap: 18px;
  padding: 28px;
  border: 1px solid rgba(224, 231, 255, 0.78);
  border-radius: 30px;
  background:
    radial-gradient(circle at top right, rgba(56, 189, 248, 0.18), transparent 24%),
    radial-gradient(circle at left top, rgba(37, 99, 235, 0.1), transparent 28%),
    linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 28px 68px rgba(15, 23, 42, 0.08);
}

.hero-card,
.workspace-card,
.panel-card,
.table-card,
.detail-drawer,
.metric-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 36px rgba(148, 163, 184, 0.1);
}

.hero-card {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  padding: 28px 32px;
  background:
    radial-gradient(circle at 86% 12%, rgba(125, 211, 252, 0.18), transparent 28%),
    radial-gradient(circle at 6% 0%, rgba(251, 207, 232, 0.18), transparent 30%),
    rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(18px);
}

.hero-eyebrow,
.dialog-eyebrow {
  margin: 0 0 10px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.hero-copy h1,
.dialog-head h3,
.drawer-head h3 {
  margin: 0;
  color: #0f172a;
  letter-spacing: -0.03em;
}

.hero-copy h1 {
  font-size: 50px;
  font-weight: 800;
}

.hero-copy p:last-child,
.section-head p,
.metric-card small,
.metric-card em,
.field-block small,
.batch-copy small,
.info-stack small,
.confirm-copy,
.drawer-head p,
.drawer-empty p,
.table-toolbar p {
  color: #64748b;
  font-size: 13px;
  line-height: 1.75;
}

.hero-actions,
.dialog-actions,
.batch-actions,
.footer-bar,
.pager,
.usage-top,
.dialog-head,
.capacity-row,
.table-toolbar,
.drawer-head,
.drawer-user,
.drawer-section-head,
.drawer-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hero-actions,
.footer-bar,
.dialog-actions,
.table-toolbar,
.drawer-head,
.drawer-section-head {
  justify-content: space-between;
}

.primary-button,
.ghost-button,
.danger-button,
.line-button,
.pager-button,
.status-toggle,
.group-pill,
.metric-card {
  border-radius: 16px;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    opacity 0.2s ease,
    border-color 0.2s ease;
}

.primary-button,
.ghost-button,
.danger-button,
.line-button,
.pager-button {
  min-height: 46px;
  padding: 0 18px;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.primary-button {
  border: 1px solid transparent;
  color: #fff;
  background: linear-gradient(135deg, #2563eb 0%, #0ea5e9 100%);
  box-shadow: 0 18px 34px rgba(37, 99, 235, 0.24);
}

.ghost-button,
.line-button,
.pager-button {
  border: 1px solid #dbe4ee;
  color: #334155;
  background: #fff;
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.1);
}

.danger-button {
  border: 1px solid transparent;
  color: #fff;
  background: linear-gradient(135deg, #ef4444 0%, #f97316 100%);
  box-shadow: 0 18px 34px rgba(239, 68, 68, 0.2);
}

.warning-button {
  color: #b45309;
  border-color: rgba(251, 191, 36, 0.35);
  background: linear-gradient(180deg, #fffdf5 0%, #fff7ed 100%);
}

.danger-text {
  color: #dc2626;
}

.slim-button {
  min-height: 40px;
  padding: 0 14px;
}

.primary-button:hover,
.ghost-button:hover,
.danger-button:hover,
.line-button:hover,
.pager-button:hover,
.status-toggle:hover,
.group-pill:hover,
.metric-card:hover {
  transform: translateY(-1px);
}

.primary-button:disabled,
.ghost-button:disabled,
.danger-button:disabled,
.line-button:disabled,
.pager-button:disabled,
.status-toggle:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  gap: 8px;
  width: 100%;
  min-height: 122px;
  padding: 18px 20px;
  text-align: left;
  border: 1px solid rgba(223, 231, 241, 0.96);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(249, 252, 255, 0.78)),
    radial-gradient(circle at 95% 10%, rgba(96, 165, 250, 0.12), transparent 32%);
  cursor: pointer;
  backdrop-filter: blur(14px);
}

.metric-card.active {
  border-color: rgba(37, 99, 235, 0.34);
  background: linear-gradient(135deg, rgba(239, 246, 255, 0.96), rgba(255, 255, 255, 0.98));
  box-shadow: 0 22px 44px rgba(37, 99, 235, 0.12);
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
}

.metric-card em {
  font-style: normal;
  color: #2563eb;
  font-weight: 700;
}

.workspace-card {
  display: block;
}

.workspace-main {
  display: grid;
  gap: 18px;
  min-width: 0;
  overflow: hidden;
}

.panel-card {
  display: grid;
  gap: 16px;
  padding: 22px 24px;
}

.section-head h2,
.table-toolbar h3,
.drawer-section-head h4 {
  margin: 0;
  color: #111827;
  font-size: 22px;
  font-weight: 800;
}

.section-head p,
.table-toolbar p {
  margin: 6px 0 0;
}

.filter-grid,
.form-grid,
.drawer-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.filter-card .filter-grid {
  grid-template-columns: minmax(260px, 1.4fr) minmax(150px, 0.75fr) minmax(150px, 0.75fr) minmax(210px, 1fr);
  align-items: start;
}

.field-block {
  display: grid;
  gap: 10px;
}

.field-span-2 {
  grid-column: span 2;
}

.filter-card .field-span-2 {
  grid-column: span 1;
}

.field-block > span,
.detail-card > span,
.highlight-item span {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.field-input {
  width: 100%;
  min-height: 46px;
  padding: 0 15px;
  border: 1px solid #d8e2ec;
  border-radius: 15px;
  background: #fff;
  color: #111827;
  font-size: 14px;
  box-sizing: border-box;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.filter-card .field-block small {
  min-height: 20px;
  color: #64748b;
  font-size: 12px;
  line-height: 1.6;
}

.field-select {
  appearance: none;
  cursor: pointer;
}

.field-input:focus {
  border-color: #60a5fa;
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.14);
}

.field-input:disabled {
  cursor: not-allowed;
  background: #f8fafc;
  color: #64748b;
}

.capacity-mode-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.mode-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0 12px;
  border: 1px solid #d8e2ec;
  border-radius: 999px;
  background: #fff;
  color: #334155;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.mode-pill input {
  accent-color: #2563eb;
}

.compact-input {
  min-width: 148px;
}

.compact-button {
  min-height: 44px;
}

.batch-bar {
  display: grid;
  gap: 16px;
  padding: 18px 20px;
  border: 1px solid rgba(147, 197, 253, 0.4);
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(239, 246, 255, 0.96), rgba(248, 250, 252, 0.98));
}

.batch-copy {
  display: grid;
  gap: 6px;
}

.batch-copy strong {
  color: #1d4ed8;
  font-size: 16px;
}

.batch-actions {
  flex-wrap: wrap;
}

.compact-state {
  min-height: 48px;
  padding: 12px;
}

.delete-preview {
  display: grid;
  gap: 12px;
  padding: 14px;
  border: 1px solid rgba(34, 197, 94, 0.32);
  border-radius: 16px;
  background: rgba(240, 253, 244, 0.72);
}

.delete-preview.blocked {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(254, 242, 242, 0.78);
}

.delete-preview strong {
  color: #0f172a;
}

.delete-preview-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.delete-preview-grid span {
  min-height: 30px;
  padding: 6px 8px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.72);
  color: #475569;
  font-size: 12px;
  font-weight: 700;
}

.table-card {
  overflow: hidden;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(249, 252, 255, 0.92)),
    radial-gradient(circle at 8% 0%, rgba(248, 187, 208, 0.2), transparent 28%),
    radial-gradient(circle at 92% 8%, rgba(96, 165, 250, 0.18), transparent 30%);
  backdrop-filter: blur(18px);
}

.table-toolbar {
  padding: 22px 22px 12px;
}

.table-row {
  display: grid;
  grid-template-columns: 38px minmax(150px, 1fr) minmax(190px, 1.05fr) minmax(132px, 0.78fr) 118px 106px minmax(164px, 0.95fr) 104px 150px;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.table-head {
  min-height: 62px;
  border-top: 1px solid #eef2f7;
  border-bottom: 1px solid #eef2f7;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 800;
}

.table-body-row {
  min-height: 86px;
  border-bottom: 1px solid #f1f5f9;
  background: rgba(255, 255, 255, 0.48);
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease;
}

.table-body-row:hover {
  background: rgba(255, 255, 255, 0.82);
  box-shadow: inset 3px 0 0 rgba(59, 130, 246, 0.58);
}

.table-body-row:last-child {
  border-bottom: none;
}

.state-shell {
  padding: 34px 24px;
  color: #64748b;
  font-size: 14px;
}

.check-shell {
  position: relative;
  display: inline-flex;
  width: 20px;
  height: 20px;
}

.check-shell input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.check-shell span {
  width: 20px;
  height: 20px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: #fff;
}

.check-shell input:checked + span {
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  border-color: #2563eb;
}

.check-shell input:checked + span::after {
  content: '';
  position: absolute;
  left: 7px;
  top: 3px;
  width: 4px;
  height: 9px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 0;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border: 2px solid rgba(255, 255, 255, 0.84);
  border-radius: 50%;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.42),
    0 14px 26px rgba(59, 130, 246, 0.18);
}

.avatar.large {
  width: 62px;
  height: 62px;
  font-size: 22px;
  border-radius: 22px;
}

.avatar.is-admin {
  background: linear-gradient(135deg, #1d4ed8, #0f172a);
}

.avatar.is-user {
  background: linear-gradient(135deg, #38bdf8, #f9a8d4);
}

.user-copy,
.info-stack {
  display: grid;
  gap: 4px;
}

.user-copy strong,
.info-stack strong,
.detail-card strong,
.highlight-item strong,
.date-text {
  color: #111827;
  font-size: 14px;
  font-weight: 700;
}

.pill,
.group-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 800;
}

.pill.role-admin {
  color: #fff;
  background: linear-gradient(135deg, rgba(30, 64, 175, 0.9), rgba(15, 23, 42, 0.92));
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.22);
}

.pill.role-user {
  color: #33506f;
  background: rgba(219, 234, 254, 0.72);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.access-stack {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.access-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid rgba(255, 255, 255, 0.42);
  border-radius: 999px;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.36),
    0 10px 22px rgba(15, 23, 42, 0.08);
}

.access-chip.is-admin {
  background: linear-gradient(135deg, rgba(30, 64, 175, 0.9), rgba(15, 23, 42, 0.92));
  box-shadow: 0 12px 28px rgba(37, 99, 235, 0.26);
}

.access-chip.is-pro {
  background: linear-gradient(135deg, rgba(244, 114, 182, 0.9), rgba(251, 146, 160, 0.9));
  box-shadow: 0 12px 28px rgba(244, 114, 182, 0.24);
}

.access-chip.is-normal {
  color: #166534;
  background: rgba(220, 252, 231, 0.78);
}

.access-chip.is-restricted {
  color: #fff;
  background: linear-gradient(135deg, rgba(100, 116, 139, 0.86), rgba(51, 65, 85, 0.9));
}

.group-pill {
  border: none;
  max-width: 100%;
  color: #1d4ed8;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.11), rgba(56, 189, 248, 0.12));
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 5px 10px 5px 6px;
  border: 1px solid rgba(203, 213, 225, 0.72);
  background: rgba(248, 250, 252, 0.72);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.74);
  cursor: pointer;
}

.status-toggle.active {
  border-color: rgba(34, 197, 94, 0.32);
  background: rgba(240, 253, 244, 0.76);
}

.status-toggle-track {
  position: relative;
  width: 38px;
  height: 22px;
  border-radius: 999px;
  background: #cbd5e1;
  transition: background 0.2s ease;
}

.status-toggle.active .status-toggle-track {
  background: linear-gradient(90deg, #22c55e, #38bdf8);
}

.status-toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  transition: left 0.2s ease;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.18);
}

.status-toggle.active .status-toggle-knob {
  left: 19px;
}

.status-text {
  color: #334155;
  font-size: 13px;
  font-weight: 800;
}

.status-toggle.active .status-text {
  color: #166534;
}

.control-dock {
  display: inline-flex;
  justify-content: flex-end;
  gap: 8px;
  width: 100%;
}

.dock-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(211, 224, 239, 0.82);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.78);
  color: #334155;
  cursor: pointer;
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.1);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease;
}

.dock-button:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.55);
  color: #1d4ed8;
  box-shadow: 0 14px 28px rgba(59, 130, 246, 0.16);
}

.dock-button.danger:hover {
  border-color: rgba(248, 113, 113, 0.55);
  color: #dc2626;
}

.dock-button svg {
  width: 17px;
  height: 17px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.usage-stack,
.drawer-section,
.drawer-empty {
  display: grid;
  gap: 10px;
}

.usage-top {
  justify-content: space-between;
  color: #475569;
  font-size: 13px;
}

.usage-top strong,
.drawer-section-head strong {
  color: #2563eb;
}

.progress-track {
  position: relative;
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: #e8eef5;
}

.progress-track.large {
  height: 12px;
}

.progress-fill {
  position: absolute;
  inset: 0 auto 0 0;
  border-radius: 999px;
  background: linear-gradient(90deg, #38bdf8, #2563eb);
}

.col-actions {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
  min-width: 0;
  overflow: visible;
}

.col-actions .line-button {
  min-height: 36px;
  padding: 0 12px;
  font-size: 13px;
  flex: 0 0 auto;
  white-space: nowrap;
}

.table-head .col-actions {
  display: block;
  text-align: left;
}

.footer-bar {
  padding: 0 4px;
}

.pager-index {
  min-width: 92px;
  text-align: center;
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.footer-select {
  width: 140px;
}

.detail-drawer {
  position: relative;
  display: grid;
  align-content: start;
  gap: 18px;
  width: 360px;
  max-width: 100%;
  min-height: 100%;
  padding: 22px;
  box-sizing: border-box;
  overflow: hidden;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.08), transparent 20%),
    linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.detail-bottom-card {
  border: 1px solid rgba(226, 232, 240, 0.82);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 251, 255, 0.88)),
    radial-gradient(circle at 8% 0%, rgba(251, 207, 232, 0.16), transparent 28%),
    radial-gradient(circle at 92% 8%, rgba(96, 165, 250, 0.14), transparent 30%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    0 18px 36px rgba(148, 163, 184, 0.1);
  overflow: hidden;
}

.detail-drawer.bottom-layout {
  width: 100%;
  max-width: none;
  min-height: 0;
  padding: 20px;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: stretch;
}

.detail-drawer.bottom-layout.empty {
  display: grid;
  grid-template-columns: 1fr;
}

.detail-band-top {
  display: grid;
  grid-template-columns: minmax(220px, 300px) minmax(300px, 0.95fr) minmax(240px, 1fr) minmax(132px, 0.42fr);
  gap: 12px;
  align-items: stretch;
}

.detail-drawer.bottom-layout .drawer-head {
  min-height: 132px;
  align-items: flex-start;
  padding: 16px 18px;
  border: 1px solid #e6edf5;
  border-radius: 22px;
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.12), transparent 30%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.99), rgba(246, 250, 255, 0.96));
  box-shadow: 0 18px 36px rgba(148, 163, 184, 0.12);
}

.detail-drawer.bottom-layout .drawer-highlight {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-self: stretch;
}

.detail-drawer.bottom-layout .drawer-section {
  min-width: 0;
  min-height: 132px;
}

.detail-drawer.bottom-layout .drawer-actions {
  min-height: 132px;
  justify-content: flex-end;
  align-content: stretch;
  flex-wrap: nowrap;
  flex-direction: column;
}

.detail-drawer.bottom-layout .drawer-actions .toolbar-action {
  width: 100%;
}

.detail-drawer.bottom-layout .drawer-grid {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.detail-drawer.empty {
  align-content: center;
}

.drawer-head h3 {
  font-size: 28px;
  font-weight: 800;
}

.drawer-eyebrow {
  margin-bottom: 6px;
}

.drawer-user {
  align-items: flex-start;
}

.drawer-close {
  width: 36px;
  height: 36px;
  border: 1px solid #dbe4ee;
  border-radius: 999px;
  background: #fff;
  color: #64748b;
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
}

.drawer-highlight {
  display: grid;
  gap: 12px;
}

.highlight-item,
.detail-card,
.usage-card {
  display: grid;
  gap: 7px;
  padding: 14px 16px;
  border: 1px solid #e6edf5;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.84);
}

.highlight-item {
  min-height: 132px;
  grid-template-rows: auto 1fr;
  align-content: stretch;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(245, 248, 255, 0.96));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.highlight-item strong {
  align-self: end;
  font-size: 20px;
  letter-spacing: -0.02em;
}

.drawer-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metrics-strip .detail-card {
  min-height: 88px;
  position: relative;
  overflow: hidden;
  border-color: rgba(191, 219, 254, 0.62);
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.18), transparent 34%),
    linear-gradient(135deg, rgba(239, 246, 255, 0.96), rgba(255, 255, 255, 0.98));
  box-shadow: 0 16px 30px rgba(59, 130, 246, 0.08);
}

.metrics-strip .detail-card::after {
  content: '';
  position: absolute;
  inset: auto 16px 0 16px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, #2563eb, #38bdf8);
  opacity: 0.9;
}

.compact-usage .usage-card {
  min-height: 100%;
  border-radius: 22px;
  background:
    radial-gradient(circle at top right, rgba(56, 189, 248, 0.18), transparent 34%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98), rgba(239, 246, 255, 0.96));
  box-shadow: 0 18px 36px rgba(56, 189, 248, 0.12);
  align-content: space-between;
}

.drawer-user > div,
.drawer-head,
.drawer-section,
.drawer-grid,
.drawer-actions {
  min-width: 0;
}

.drawer-user p,
.detail-card strong,
.highlight-item strong {
  word-break: break-word;
}

.drawer-actions {
  flex-wrap: wrap;
}

.toolbar-action {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  min-height: 40px;
  padding: 0 12px;
  border: 1px solid #dbe4ee;
  border-radius: 16px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(246, 249, 255, 0.98));
  color: #334155;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 12px 24px rgba(148, 163, 184, 0.12);
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.toolbar-action:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.4);
  box-shadow: 0 16px 28px rgba(59, 130, 246, 0.12);
}

.toolbar-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 10px;
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.04em;
  box-shadow: 0 10px 18px rgba(37, 99, 235, 0.22);
}

.drawer-empty {
  justify-items: center;
  text-align: center;
  padding: 12px;
}

.drawer-empty-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 68px;
  height: 68px;
  border-radius: 24px;
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  color: #fff;
  font-size: 28px;
  font-weight: 800;
  box-shadow: 0 18px 36px rgba(37, 99, 235, 0.18);
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
  width: min(860px, 100%);
  display: grid;
  gap: 22px;
  padding: 28px;
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 30px 80px rgba(15, 23, 42, 0.18);
}

.form-panel {
  width: min(980px, 100%);
}

.confirm-panel {
  width: min(540px, 100%);
}

.dialog-close {
  width: 38px;
  height: 38px;
  border: 1px solid #d9e2ec;
  border-radius: 999px;
  background: #fff;
  color: #64748b;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
}

.unit-select {
  width: 120px;
}

@media (max-width: 1500px) {
  .detail-drawer {
    width: 100%;
    min-height: 0;
  }

  .detail-band-top,
  .detail-drawer.bottom-layout {
    grid-template-columns: 1fr;
  }

  .detail-drawer.bottom-layout .drawer-highlight,
  .detail-drawer.bottom-layout .drawer-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .detail-drawer.bottom-layout .drawer-actions {
    justify-content: flex-start;
    min-height: 0;
    flex-direction: row;
    flex-wrap: wrap;
  }
}

@media (max-width: 1360px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filter-grid,
  .form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filter-card .filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .table-card {
    overflow-x: auto;
  }

  .table-row {
    min-width: 1260px;
  }
}

@media (max-width: 900px) {
  .page-shell,
  .dialog-panel,
  .workspace-card {
    padding: 20px 18px 24px;
  }

  .hero-card,
  .hero-actions,
  .batch-actions,
  .footer-bar,
  .dialog-actions,
  .capacity-row,
  .drawer-head,
  .drawer-user {
    display: grid;
    grid-template-columns: 1fr;
  }

  .hero-copy h1 {
    font-size: 38px;
  }

  .metric-grid,
  .filter-grid,
  .form-grid,
  .drawer-grid {
    grid-template-columns: 1fr;
  }

  .detail-drawer.bottom-layout .drawer-highlight,
  .detail-drawer.bottom-layout .drawer-grid {
    grid-template-columns: 1fr;
  }

  .field-span-2 {
    grid-column: span 1;
  }
}
</style>
