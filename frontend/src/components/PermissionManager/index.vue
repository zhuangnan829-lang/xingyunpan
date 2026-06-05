<template>
  <div class="permission-manager">
    <!-- Current User Permission Display -->
    <div v-if="currentPermission" class="current-permission">
      <div class="permission-header">
        <icon-safe class="permission-icon" />
        <span class="permission-title">您的权限</span>
      </div>
      <div class="permission-badge" :class="permissionClass">
        {{ permissionLabel }}
      </div>
    </div>

    <!-- Permission Matrix -->
    <div class="permission-matrix">
      <h4>权限矩阵</h4>
      <a-table
        :columns="columns"
        :data="permissionData"
        :pagination="false"
        :bordered="true"
        size="small"
      >
        <template #permission="{ record }">
          <strong>{{ record.permission }}</strong>
        </template>
        <template #view="{ record }">
          <icon-check v-if="record.view" class="check-icon" />
          <icon-close v-else class="close-icon" />
        </template>
        <template #download="{ record }">
          <icon-check v-if="record.download" class="check-icon" />
          <icon-close v-else class="close-icon" />
        </template>
        <template #edit="{ record }">
          <icon-check v-if="record.edit" class="check-icon" />
          <icon-close v-else class="close-icon" />
        </template>
        <template #delete="{ record }">
          <icon-check v-if="record.delete" class="check-icon" />
          <icon-close v-else class="close-icon" />
        </template>
      </a-table>
    </div>

    <!-- Permission-Based UI Rendering Info -->
    <div v-if="showActions" class="permission-actions">
      <h4>可用操作</h4>
      <div class="actions-list">
        <div v-if="canView" class="action-item">
          <icon-eye class="action-icon" />
          <span>查看文件信息和预览</span>
        </div>
        <div v-if="canDownload" class="action-item">
          <icon-download class="action-icon" />
          <span>下载文件</span>
        </div>
        <div v-if="canEdit" class="action-item">
          <icon-edit class="action-icon" />
          <span>上传新版本和重命名</span>
        </div>
        <div v-if="isOwner" class="action-item owner">
          <icon-user class="action-icon" />
          <span>管理协作者和删除文件</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import {
  IconSafe,
  IconCheck,
  IconClose,
  IconEye,
  IconDownload,
  IconEdit,
  IconUser
} from '@arco-design/web-vue/es/icon';
import type { PermissionCheck } from '@/api/collaboration';

interface Props {
  permission?: PermissionCheck | null;
  showActions?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  permission: null,
  showActions: true
});

// Table columns
const columns = [
  {
    title: '权限级别',
    dataIndex: 'permission',
    slotName: 'permission',
    width: 120
  },
  {
    title: '查看',
    dataIndex: 'view',
    slotName: 'view',
    align: 'center' as const,
    width: 80
  },
  {
    title: '下载',
    dataIndex: 'download',
    slotName: 'download',
    align: 'center' as const,
    width: 80
  },
  {
    title: '编辑',
    dataIndex: 'edit',
    slotName: 'edit',
    align: 'center' as const,
    width: 80
  },
  {
    title: '删除',
    dataIndex: 'delete',
    slotName: 'delete',
    align: 'center' as const,
    width: 80
  }
];

// Permission matrix data
const permissionData = [
  {
    key: 'view',
    permission: '仅查看',
    view: true,
    download: false,
    edit: false,
    delete: false
  },
  {
    key: 'download',
    permission: '下载',
    view: true,
    download: true,
    edit: false,
    delete: false
  },
  {
    key: 'edit',
    permission: '编辑',
    view: true,
    download: true,
    edit: true,
    delete: false
  },
  {
    key: 'owner',
    permission: '所有者',
    view: true,
    download: true,
    edit: true,
    delete: true
  }
];

// Current permission
const currentPermission = computed(() => props.permission);

// Permission checks
const canView = computed(() => currentPermission.value?.can_view || false);
const canDownload = computed(() => currentPermission.value?.can_download || false);
const canEdit = computed(() => currentPermission.value?.can_edit || false);
const isOwner = computed(() => currentPermission.value?.is_owner || false);

// Permission label
const permissionLabel = computed(() => {
  if (!currentPermission.value) return '未知';
  
  if (isOwner.value) return '所有者';
  if (canEdit.value) return '编辑';
  if (canDownload.value) return '下载';
  if (canView.value) return '仅查看';
  
  return '无权限';
});

// Permission class for styling
const permissionClass = computed(() => {
  if (!currentPermission.value) return 'permission-none';
  
  if (isOwner.value) return 'permission-owner';
  if (canEdit.value) return 'permission-edit';
  if (canDownload.value) return 'permission-download';
  if (canView.value) return 'permission-view';
  
  return 'permission-none';
});
</script>

<style scoped>
.permission-manager {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.current-permission {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: var(--color-fill-2);
  border-radius: 4px;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.permission-icon {
  font-size: 18px;
  color: var(--color-text-2);
}

.permission-title {
  font-weight: 500;
  color: var(--color-text-1);
}

.permission-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 500;
}

.permission-owner {
  background-color: rgb(var(--primary-6));
  color: white;
}

.permission-edit {
  background-color: rgb(var(--success-6));
  color: white;
}

.permission-download {
  background-color: rgb(var(--warning-6));
  color: white;
}

.permission-view {
  background-color: rgb(var(--gray-6));
  color: white;
}

.permission-none {
  background-color: var(--color-fill-3);
  color: var(--color-text-3);
}

.permission-matrix h4 {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
}

.check-icon {
  color: rgb(var(--success-6));
  font-size: 18px;
}

.close-icon {
  color: rgb(var(--danger-6));
  font-size: 18px;
}

.permission-actions h4 {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
}

.actions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: var(--color-fill-2);
  border-radius: 4px;
  font-size: 13px;
}

.action-item.owner {
  background-color: rgb(var(--primary-1));
  border: 1px solid rgb(var(--primary-3));
}

.action-icon {
  font-size: 16px;
  color: var(--color-text-2);
}

.action-item.owner .action-icon {
  color: rgb(var(--primary-6));
}

/* Responsive design */
@media (max-width: 768px) {
  .permission-manager {
    gap: 16px;
  }

  .current-permission {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    padding: 10px 12px;
  }

  .permission-header {
    gap: 6px;
  }

  .permission-icon {
    font-size: 16px;
  }

  .permission-title {
    font-size: 13px;
  }

  .permission-badge {
    align-self: flex-start;
    font-size: 12px;
    padding: 3px 10px;
  }

  .permission-matrix h4,
  .permission-actions h4 {
    font-size: 13px;
    margin-bottom: 10px;
  }

  .action-item {
    padding: 6px 10px;
    font-size: 12px;
  }

  .action-icon {
    font-size: 14px;
  }

  :deep(.arco-table) {
    font-size: 12px;
  }

  :deep(.arco-table-th),
  :deep(.arco-table-td) {
    padding: 8px 6px;
  }

  .check-icon,
  .close-icon {
    font-size: 16px;
  }
}

@media (max-width: 480px) {
  .permission-manager {
    gap: 12px;
  }

  .current-permission {
    padding: 8px 10px;
  }

  .permission-title {
    font-size: 12px;
  }

  .permission-badge {
    font-size: 11px;
    padding: 2px 8px;
  }

  .permission-matrix h4,
  .permission-actions h4 {
    font-size: 12px;
  }

  .action-item {
    padding: 5px 8px;
    font-size: 11px;
  }

  .action-icon {
    font-size: 13px;
  }

  :deep(.arco-table) {
    font-size: 11px;
  }

  :deep(.arco-table-th),
  :deep(.arco-table-td) {
    padding: 6px 4px;
  }

  .check-icon,
  .close-icon {
    font-size: 14px;
  }
}
</style>
