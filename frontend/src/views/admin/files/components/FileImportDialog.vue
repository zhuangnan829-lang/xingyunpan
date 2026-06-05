<template>
  <el-dialog
    :model-value="visible"
    width="720px"
    append-to-body
    destroy-on-close
    class="file-import-dialog"
    @close="emit('update:visible', false)"
  >
    <template #header>
      <div class="dialog-header">
        <h2>导入外部目录</h2>
      </div>
    </template>

    <div class="dialog-body">
      <section class="tip-card info-card">
        <div class="tip-icon">i</div>
        <p>你可以将存储策略中已有文件、目录结构导入到 Cloudreve 中，导入操作不会额外占用物理存储空间，但仍会正常扣除用户已用容量空间。</p>
      </section>

      <section class="tip-card warn-card">
        <div class="tip-icon">!</div>
        <div>
          <h3>注意事项</h3>
          <ul>
            <li>导入后，物理文件将由 Cloudreve 接管，后续请不要在外部修改此文件；</li>
            <li>不要重复导入相同的文件；</li>
            <li>如果用户文件冲突，此文件会被跳过；</li>
          </ul>
        </div>
      </section>

      <label class="field-block">
        <span>存储策略</span>
        <select v-model="form.storagePolicyId" class="field-input field-select">
          <option :value="0">请选择存储策略</option>
          <option v-for="policy in policies" :key="policy.id" :value="policy.id">
            {{ policy.name }}
          </option>
        </select>
        <small>选择要导入文件目前存储所在的存储策略。</small>
      </label>

      <label class="field-block">
        <span>原始目录路径</span>
        <input v-model.trim="form.sourcePath" class="field-input" type="text" placeholder="/path/to/folder" />
        <small>要导入的目录在存储端的路径。</small>
      </label>

      <label class="field-block">
        <span>目标用户</span>
        <select v-model="form.ownerId" class="field-input field-select">
          <option :value="0">选择用户</option>
          <option v-for="user in users" :key="user.id" :value="user.id">
            {{ user.username }} ({{ user.email }})
          </option>
        </select>
        <small>选择要将文件导入到哪个用户的文件系统中。</small>
      </label>

      <label class="field-block">
        <span>目的目录路径</span>
        <input v-model.trim="form.targetPath" class="field-input" type="text" placeholder="/" />
        <small>要将目录导入到用户文件系统中的路径。</small>
      </label>

      <label class="switch-row">
        <input v-model="form.recursive" type="checkbox" />
        <div>
          <strong>递归导入子目录</strong>
          <small>是否将目录下的所有子目录递归导入。</small>
        </div>
      </label>

      <label class="switch-row">
        <input v-model="form.extractMedia" type="checkbox" />
        <div>
          <strong>提取媒体信息</strong>
          <small>是否在导入文件的同时尝试提取每个文件的媒体信息。</small>
        </div>
      </label>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <button class="cancel-button" type="button" @click="emit('update:visible', false)">取消</button>
        <button class="confirm-button" type="button" :disabled="submitting" @click="submit">
          {{ submitting ? '处理中...' : '确定' }}
        </button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue';
import { ElMessage } from 'element-plus';
import type { AdminUserPayload } from '@/api/admin-users';
import type { StoragePolicyPayload } from '@/api/storage-policy';

const props = defineProps<{
  visible: boolean;
  policies: StoragePolicyPayload[];
  users: AdminUserPayload[];
  submitting?: boolean;
}>();

const emit = defineEmits<{
  (event: 'update:visible', value: boolean): void;
  (event: 'submit', value: {
    storagePolicyId: number;
    sourcePath: string;
    ownerId: number;
    targetPath: string;
    recursive: boolean;
    extractMedia: boolean;
  }): void;
}>();

const form = reactive({
  storagePolicyId: 0,
  sourcePath: '',
  ownerId: 0,
  targetPath: '/',
  recursive: true,
  extractMedia: false,
});

const reset = () => {
  form.storagePolicyId = 0;
  form.sourcePath = '';
  form.ownerId = 0;
  form.targetPath = '/';
  form.recursive = true;
  form.extractMedia = false;
};

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      reset();
    }
  },
);

const submit = () => {
  if (!form.storagePolicyId) {
    ElMessage.warning('请选择存储策略');
    return;
  }
  if (!form.sourcePath) {
    ElMessage.warning('请输入原始目录路径');
    return;
  }
  if (!form.ownerId) {
    ElMessage.warning('请选择目标用户');
    return;
  }
  if (!form.targetPath) {
    ElMessage.warning('请输入目的目录路径');
    return;
  }

  emit('submit', {
    storagePolicyId: form.storagePolicyId,
    sourcePath: form.sourcePath,
    ownerId: form.ownerId,
    targetPath: form.targetPath,
    recursive: form.recursive,
    extractMedia: form.extractMedia,
  });
};
</script>

<style scoped>
.dialog-header h2 {
  margin: 0;
  color: #202530;
  font-size: 24px;
  font-weight: 900;
}

.dialog-body {
  display: grid;
  gap: 18px;
  padding-bottom: 6px;
}

.tip-card {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 14px;
  padding: 18px 20px;
  border-radius: 20px;
}

.info-card {
  background: #e7f5ff;
  color: #1d5e8a;
}

.warn-card {
  background: #fff3df;
  color: #8f5412;
}

.tip-icon {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  font-weight: 900;
}

.warn-card h3 {
  margin: 2px 0 10px;
  font-size: 18px;
}

.warn-card ul {
  margin: 0;
  padding-left: 20px;
}

.warn-card li + li {
  margin-top: 6px;
}

.field-block {
  display: grid;
  gap: 8px;
}

.field-block > span {
  color: #202530;
  font-size: 16px;
  font-weight: 800;
}

.field-block small,
.switch-row small {
  color: #727b88;
  font-size: 13px;
  line-height: 1.7;
}

.field-input {
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid #d5deea;
  border-radius: 16px;
  background: #fff;
  color: #202530;
  font-size: 15px;
  outline: none;
}

.field-select {
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, #7a7f87 50%),
    linear-gradient(135deg, #7a7f87 50%, transparent 50%);
  background-position:
    calc(100% - 26px) calc(50% - 3px),
    calc(100% - 16px) calc(50% - 3px);
  background-size: 10px 10px, 10px 10px;
  background-repeat: no-repeat;
}

.switch-row {
  display: grid;
  grid-template-columns: 20px 1fr;
  gap: 12px;
  align-items: start;
}

.switch-row input {
  width: 20px;
  height: 20px;
  margin-top: 2px;
  accent-color: #2a7de1;
}

.switch-row strong {
  display: block;
  margin-bottom: 4px;
  color: #202530;
  font-size: 15px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.cancel-button,
.confirm-button {
  min-width: 84px;
  min-height: 44px;
  padding: 0 20px;
  border-radius: 15px;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
}

.cancel-button {
  border: 0;
  background: transparent;
  color: #2a7de1;
}

.confirm-button {
  border: 0;
  background: linear-gradient(135deg, #1f7ae0, #2f8af0);
  color: #fff;
}

.confirm-button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

:deep(.file-import-dialog .el-dialog) {
  border-radius: 28px;
  overflow: hidden;
}

:deep(.file-import-dialog .el-dialog__header) {
  margin: 0;
  padding: 24px 28px 8px;
}

:deep(.file-import-dialog .el-dialog__body) {
  padding: 12px 28px 18px;
}

:deep(.file-import-dialog .el-dialog__footer) {
  padding: 0 28px 24px;
}
</style>
