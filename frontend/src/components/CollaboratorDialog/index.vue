<template>
  <a-modal
    v-model:visible="visibleModel"
    title="管理协作者"
    width="600px"
    :footer="false"
    @cancel="handleClose"
  >
    <!-- Add Collaborator Section -->
    <div class="add-collaborator-section">
      <h4>添加协作者</h4>
      <a-form :model="form" layout="vertical">
        <a-form-item label="用户名" required>
          <a-input
            v-model="form.username"
            placeholder="输入用户名"
            :disabled="isLoading"
          />
        </a-form-item>
        <a-form-item label="权限级别" required>
          <a-select
            v-model="form.permission"
            placeholder="选择权限级别"
            :disabled="isLoading"
          >
            <a-option value="view">
              <div class="permission-option">
                <span class="permission-label">仅查看</span>
                <span class="permission-desc">可以查看文件信息和预览</span>
              </div>
            </a-option>
            <a-option value="download">
              <div class="permission-option">
                <span class="permission-label">下载</span>
                <span class="permission-desc">可以查看和下载文件</span>
              </div>
            </a-option>
            <a-option value="edit">
              <div class="permission-option">
                <span class="permission-label">编辑</span>
                <span class="permission-desc">可以查看、下载、上传新版本和重命名</span>
              </div>
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-button
            type="primary"
            :loading="isLoading"
            :disabled="!form.username || !form.permission"
            @click="handleAddCollaborator"
          >
            添加协作者
          </a-button>
        </a-form-item>
      </a-form>
    </div>

    <a-divider />

    <!-- Current Collaborators List -->
    <div class="collaborators-list-section">
      <h4>当前协作者 ({{ collaboratorsList.length }})</h4>
      <div v-if="isLoadingCollaborators" class="loading-container">
        <a-spin />
      </div>
      <div v-else-if="collaboratorsList.length === 0" class="empty-state">
        <a-empty description="暂无协作者" />
      </div>
      <div v-else class="collaborators-list">
        <div
          v-for="collaborator in collaboratorsList"
          :key="collaborator.user_id"
          class="collaborator-item"
        >
          <div class="collaborator-info">
            <div class="collaborator-name">
              <icon-user />
              <span>{{ collaborator.username }}</span>
            </div>
            <div class="collaborator-meta">
              添加于 {{ formatDate(collaborator.added_at) }}
            </div>
          </div>
          <div class="collaborator-actions">
            <a-select
              v-model="collaborator.permission"
              size="small"
              :disabled="isLoading"
              @change="handleUpdatePermission(collaborator)"
            >
              <a-option value="view">仅查看</a-option>
              <a-option value="download">下载</a-option>
              <a-option value="edit">编辑</a-option>
            </a-select>
            <a-button
              type="text"
              status="danger"
              size="small"
              :loading="isLoading"
              @click="handleRemoveCollaborator(collaborator)"
            >
              <template #icon>
                <icon-delete />
              </template>
            </a-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Permission Descriptions -->
    <div class="permission-descriptions">
      <h4>权限说明</h4>
      <div class="permission-desc-list">
        <div class="permission-desc-item">
          <strong>仅查看：</strong>可以查看文件信息和预览，但不能下载或编辑
        </div>
        <div class="permission-desc-item">
          <strong>下载：</strong>可以查看和下载文件，但不能编辑
        </div>
        <div class="permission-desc-item">
          <strong>编辑：</strong>可以查看、下载、上传新版本和重命名文件
        </div>
        <div class="permission-desc-item">
          <strong>注意：</strong>只有文件所有者可以删除文件和管理协作者
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Message } from '@arco-design/web-vue';
import { IconUser, IconDelete } from '@arco-design/web-vue/es/icon';
import { useCollaborationStore } from '@/stores/collaboration';
import type { Collaborator, PermissionLevel } from '@/api/collaboration';

interface Props {
  visible: boolean;
  fileId: string;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'success'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const visibleModel = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
});

const collaborationStore = useCollaborationStore();

// Form state
const form = ref({
  username: '',
  permission: '' as PermissionLevel | ''
});

// Loading states
const isLoading = ref(false);
const isLoadingCollaborators = ref(false);

// Collaborators list
const collaboratorsList = computed(() => {
  return collaborationStore.getCollaborators(props.fileId);
});

// Watch visible prop to load collaborators
watch(
  () => props.visible,
  async (newVisible) => {
    if (newVisible && props.fileId) {
      await loadCollaborators();
    }
  },
  { immediate: true }
);

/**
 * Load collaborators for the file
 */
async function loadCollaborators() {
  if (!props.fileId) return;

  isLoadingCollaborators.value = true;
  try {
    await collaborationStore.loadCollaborators(props.fileId);
  } catch (error: any) {
    Message.error(error.message || '加载协作者列表失败');
  } finally {
    isLoadingCollaborators.value = false;
  }
}

/**
 * Handle adding a collaborator
 */
async function handleAddCollaborator() {
  if (!form.value.username || !form.value.permission) {
    Message.warning('请填写用户名和选择权限级别');
    return;
  }

  isLoading.value = true;
  try {
    await collaborationStore.addCollaborator(
      props.fileId,
      form.value.username,
      form.value.permission as PermissionLevel
    );

    Message.success('协作者添加成功');
    
    // Reset form
    form.value.username = '';
    form.value.permission = '';

    emit('success');
  } catch (error: any) {
    if (error.message?.includes('不存在')) {
      Message.error('用户不存在，请检查用户名');
    } else if (error.message?.includes('已是协作者')) {
      Message.error('该用户已是协作者');
    } else {
      Message.error(error.message || '添加协作者失败');
    }
  } finally {
    isLoading.value = false;
  }
}

/**
 * Handle updating collaborator permission
 */
async function handleUpdatePermission(collaborator: Collaborator) {
  isLoading.value = true;
  try {
    await collaborationStore.updatePermission(
      props.fileId,
      collaborator.user_id,
      collaborator.permission
    );

    Message.success('权限更新成功');
    emit('success');
  } catch (error: any) {
    Message.error(error.message || '更新权限失败');
    // Reload collaborators to revert the change
    await loadCollaborators();
  } finally {
    isLoading.value = false;
  }
}

/**
 * Handle removing a collaborator
 */
async function handleRemoveCollaborator(collaborator: Collaborator) {
  isLoading.value = true;
  try {
    await collaborationStore.removeCollaborator(props.fileId, collaborator.user_id);

    Message.success('协作者已移除');
    emit('success');
  } catch (error: any) {
    Message.error(error.message || '移除协作者失败');
  } finally {
    isLoading.value = false;
  }
}

/**
 * Format date string
 */
function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  });
}

/**
 * Handle dialog close
 */
function handleClose() {
  emit('update:visible', false);
}
</script>

<style scoped>
.add-collaborator-section {
  margin-bottom: 16px;
}

.add-collaborator-section h4 {
  margin-bottom: 16px;
  font-size: 14px;
  font-weight: 600;
}

.permission-option {
  display: flex;
  flex-direction: column;
}

.permission-label {
  font-weight: 500;
}

.permission-desc {
  font-size: 12px;
  color: var(--color-text-3);
  margin-top: 2px;
}

.collaborators-list-section {
  margin-bottom: 16px;
}

.collaborators-list-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
}

.loading-container {
  display: flex;
  justify-content: center;
  padding: 24px;
}

.empty-state {
  padding: 24px;
}

.collaborators-list {
  max-height: 300px;
  overflow-y: auto;
}

.collaborator-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--color-border-2);
  border-radius: 4px;
  margin-bottom: 8px;
}

.collaborator-item:last-child {
  margin-bottom: 0;
}

.collaborator-info {
  flex: 1;
}

.collaborator-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  margin-bottom: 4px;
}

.collaborator-meta {
  font-size: 12px;
  color: var(--color-text-3);
}

.collaborator-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.permission-descriptions {
  margin-top: 16px;
  padding: 12px;
  background-color: var(--color-fill-2);
  border-radius: 4px;
}

.permission-descriptions h4 {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
}

.permission-desc-list {
  font-size: 12px;
  line-height: 1.6;
}

.permission-desc-item {
  margin-bottom: 4px;
}

.permission-desc-item:last-child {
  margin-bottom: 0;
  margin-top: 8px;
  color: var(--color-text-3);
}

/* Mobile responsive */
@media (max-width: 767px) {
  .add-collaborator-section h4,
  .collaborators-list-section h4,
  .permission-descriptions h4 {
    font-size: 13px;
  }

  .collaborators-list {
    max-height: 250px;
  }

  .collaborator-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 10px;
  }

  .collaborator-info {
    width: 100%;
  }

  .collaborator-name {
    font-size: 13px;
  }

  .collaborator-meta {
    font-size: 11px;
  }

  .collaborator-actions {
    width: 100%;
    justify-content: space-between;
  }

  .collaborator-actions .arco-select {
    flex: 1;
  }

  .permission-option {
    font-size: 13px;
  }

  .permission-desc {
    font-size: 11px;
  }

  .permission-descriptions {
    padding: 10px;
  }

  .permission-desc-list {
    font-size: 11px;
  }

  :deep(.arco-modal-body) {
    padding: 16px;
  }

  :deep(.arco-form-item) {
    margin-bottom: 12px;
  }
}

@media (max-width: 480px) {
  .add-collaborator-section h4,
  .collaborators-list-section h4,
  .permission-descriptions h4 {
    font-size: 12px;
  }

  .collaborators-list {
    max-height: 200px;
  }

  .collaborator-item {
    padding: 8px;
  }

  .collaborator-name {
    font-size: 12px;
  }

  .collaborator-meta {
    font-size: 10px;
  }

  .permission-option {
    font-size: 12px;
  }

  .permission-desc {
    font-size: 10px;
  }

  .permission-descriptions {
    padding: 8px;
  }

  .permission-desc-list {
    font-size: 10px;
  }

  :deep(.arco-modal-body) {
    padding: 12px;
  }

  :deep(.arco-btn) {
    font-size: 13px;
  }
}
</style>
