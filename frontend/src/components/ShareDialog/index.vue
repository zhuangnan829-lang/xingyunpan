<template>
  <el-dialog
    v-model="dialogVisible"
    width="520px"
    class="share-create-dialog"
    :show-close="false"
    :fullscreen="isMobile()"
    destroy-on-close
    @close="handleClose"
  >
    <div class="share-modal">
      <header class="share-header">
        <h2>创建分享链接</h2>
        <button type="button" aria-label="关闭" @click="handleClose">
          <el-icon><Close /></el-icon>
        </button>
      </header>

      <template v-if="!shareLink">
        <div v-if="hasFolders" class="warning-line">当前仅支持分享文件，已自动忽略文件夹。</div>

        <section class="file-row glass-surface">
          <el-icon><Document /></el-icon>
          <div>
            <span>分享文件</span>
            <strong>{{ validFileNames.join(', ') || '未选择可分享文件' }}</strong>
          </div>
        </section>

        <section class="option-list">
          <article class="share-option" :class="{ expanded: formData.passwordEnabled }">
            <button type="button" class="option-main" @click="formData.passwordEnabled = !formData.passwordEnabled">
              <el-icon><View /></el-icon>
              <span>使用密码保护链接</span>
              <el-checkbox v-model="formData.passwordEnabled" @click.stop />
            </button>
            <div v-if="formData.passwordEnabled" class="option-body">
              <input v-model="formData.password" maxlength="8" placeholder="可选，4-8 位字符" />
              <small>留空表示无需密码，勾选后建议设置访问密码。</small>
            </div>
          </article>

          <article class="share-option" :class="{ expanded: formData.expiryEnabled }">
            <button type="button" class="option-main" @click="formData.expiryEnabled = !formData.expiryEnabled">
              <el-icon><Timer /></el-icon>
              <span>超时自动过期</span>
              <el-checkbox v-model="formData.expiryEnabled" @click.stop />
            </button>
            <div v-if="formData.expiryEnabled" class="option-body is-inline">
              <input v-model.number="formData.expiryAmount" type="number" min="0.5" step="0.5" />
              <select v-model="formData.expiryUnit" aria-label="过期单位">
                <option value="hour">小时后过期</option>
                <option value="day">天后过期</option>
              </select>
            </div>
          </article>

          <article class="share-option" :class="{ expanded: formData.downloadExpiryEnabled }">
            <button
              type="button"
              class="option-main"
              @click="formData.downloadExpiryEnabled = !formData.downloadExpiryEnabled"
            >
              <el-icon><RefreshLeft /></el-icon>
              <span>下载后自动过期</span>
              <el-checkbox v-model="formData.downloadExpiryEnabled" @click.stop />
            </button>
            <div v-if="formData.downloadExpiryEnabled" class="option-body is-inline">
              <input v-model.number="formData.maxDownloads" type="number" min="1" max="9999" />
              <span>次下载后过期</span>
            </div>
          </article>
        </section>

        <footer class="share-footer">
          <button type="button" class="plain" @click="handleClose">取消</button>
          <button type="button" class="primary" :disabled="loading || validFileIds.length === 0" @click="handleGenerateLink">
            {{ loading ? '生成中...' : '确定' }}
          </button>
        </footer>
      </template>

      <template v-else>
        <section class="result-box glass-surface">
          <div>
            <span>分享链接</span>
            <strong>{{ shareLink }}</strong>
          </div>
          <button type="button" title="复制链接" @click="handleCopyLink">
            <el-icon><CopyDocument /></el-icon>
          </button>
        </section>

        <footer class="share-footer">
          <button type="button" class="plain" @click="handleCopyLink">复制</button>
          <button type="button" class="primary" @click="handleClose">关闭</button>
        </footer>
      </template>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { ElCheckbox, ElDialog, ElIcon, ElMessage } from 'element-plus';
import { Close, CopyDocument, Document, RefreshLeft, Timer, View } from '@element-plus/icons-vue';
import { useBreakpoint } from '@/composables/useBreakpoint';
import { useFileStore } from '@/stores/file';
import { useShareStore } from '@/stores/share';
import { resolveShareURL } from '@/utils/public-url';
import { copyToClipboard, validatePassword } from '@/utils/share-utils';

interface Props {
  visible: boolean;
  fileIds: string[];
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'share-created', shareLink: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  fileIds: () => [],
});

const emit = defineEmits<Emits>();
const shareStore = useShareStore();
const fileStore = useFileStore();
const { isMobile } = useBreakpoint();

const loading = ref(false);
const dialogVisible = ref(props.visible);
const shareLink = ref<string | null>(null);

const formData = reactive({
  passwordEnabled: false,
  password: '',
  expiryEnabled: false,
  expiryAmount: 1,
  expiryUnit: 'day' as 'hour' | 'day',
  downloadExpiryEnabled: false,
  maxDownloads: 1,
});

const validFileIds = computed(() =>
  props.fileIds.filter((id) => !fileStore.folders.some((folder) => folder.id.toString() === id)),
);

const validFileNames = computed(() =>
  validFileIds.value.map((id) => {
    const file = fileStore.files.find((entry) => entry.id.toString() === id);
    return file ? file.name : `文件 ${id}`;
  }),
);

const hasFolders = computed(() => props.fileIds.length > validFileIds.value.length);

watch(
  () => props.visible,
  (newVal) => {
    dialogVisible.value = newVal;
    if (newVal) resetForm();
  },
);

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
});

function resetForm(): void {
  shareLink.value = null;
  formData.passwordEnabled = false;
  formData.password = '';
  formData.expiryEnabled = false;
  formData.expiryAmount = 1;
  formData.expiryUnit = 'day';
  formData.downloadExpiryEnabled = false;
  formData.maxDownloads = 1;
}

function handleClose(): void {
  dialogVisible.value = false;
  resetForm();
}

function normalizeCreatedLink(url: string, token?: string): string {
  return resolveShareURL(url, token);
}

async function handleGenerateLink(): Promise<void> {
  if (validFileIds.value.length === 0) {
    ElMessage.warning('请选择至少一个可分享文件');
    return;
  }

  if (formData.passwordEnabled && !validatePassword(formData.password)) {
    ElMessage.warning('访问密码需要 4-8 位字符');
    return;
  }

  if (formData.expiryEnabled && (!formData.expiryAmount || formData.expiryAmount < 0.5)) {
    ElMessage.warning('请输入有效的过期时间');
    return;
  }

  if (formData.downloadExpiryEnabled && (!formData.maxDownloads || formData.maxDownloads < 1)) {
    ElMessage.warning('请输入有效的下载次数');
    return;
  }

  try {
    loading.value = true;
    const expirySeconds =
      formData.expiryUnit === 'hour' ? formData.expiryAmount * 60 * 60 : formData.expiryAmount * 24 * 60 * 60;
    const share = await shareStore.createShare(validFileIds.value, {
      expires_in: formData.expiryEnabled ? Math.round(expirySeconds) : null,
      access_code: formData.passwordEnabled ? formData.password : null,
      max_downloads: formData.downloadExpiryEnabled ? formData.maxDownloads : null,
    });

    shareLink.value = normalizeCreatedLink(share.share_url, share.share_token);
    ElMessage.success('分享链接已生成');
    emit('share-created', shareLink.value);
  } catch (error: any) {
    ElMessage.error(error.message || '生成分享链接失败');
  } finally {
    loading.value = false;
  }
}

async function handleCopyLink(): Promise<void> {
  if (!shareLink.value) return;
  try {
    await copyToClipboard(shareLink.value);
    ElMessage.success('分享链接已复制');
  } catch {
    ElMessage.error('复制失败，请手动复制');
  }
}
</script>

<style scoped>
:deep(.share-create-dialog .el-dialog) {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 22px;
  background:
    radial-gradient(circle at 8% 0%, rgba(186, 230, 253, 0.55), transparent 38%),
    radial-gradient(circle at 100% 8%, rgba(252, 231, 243, 0.52), transparent 40%),
    rgba(255, 255, 255, 0.82);
  box-shadow: 0 24px 62px rgba(30, 41, 59, 0.24), inset 0 1px 0 rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(22px);
}

:deep(.share-create-dialog .el-dialog__header) {
  display: none;
}

:deep(.share-create-dialog .el-dialog__body) {
  padding: 0;
}

.share-modal {
  padding: 22px 26px 20px;
  color: #10213b;
}

.share-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.share-header h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0;
}

.share-header button,
.result-box button {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border: 0;
  border-radius: 12px;
  color: #64748b;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
}

.glass-surface,
.share-option.expanded {
  border: 1px solid rgba(255, 255, 255, 0.72);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.72), rgba(240, 249, 255, 0.5)),
    rgba(255, 255, 255, 0.48);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.95), 0 12px 28px rgba(96, 165, 250, 0.1);
  backdrop-filter: blur(16px);
}

.warning-line {
  margin-bottom: 10px;
  color: #d97706;
  font-size: 14px;
}

.file-row {
  display: flex;
  gap: 12px;
  align-items: center;
  min-height: 58px;
  margin-bottom: 12px;
  padding: 12px 14px;
  border-radius: 14px;
}

.file-row .el-icon {
  color: #2f7df6;
  font-size: 18px;
}

.file-row div {
  min-width: 0;
}

.file-row span,
.result-box span,
.option-body small {
  display: block;
  color: #64748b;
  font-size: 12px;
}

.file-row strong {
  display: block;
  overflow: hidden;
  margin-top: 3px;
  color: #1e293b;
  font-size: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.option-list {
  display: grid;
  gap: 10px;
}

.share-option {
  overflow: hidden;
  border-radius: 14px;
  transition: box-shadow 0.2s ease, background 0.2s ease;
}

.option-main {
  display: grid;
  grid-template-columns: 28px 1fr auto;
  align-items: center;
  width: 100%;
  min-height: 50px;
  padding: 0 14px;
  border: 0;
  color: #1e293b;
  background: transparent;
  font-size: 15px;
  font-weight: 700;
  text-align: left;
  cursor: pointer;
}

.option-main .el-icon {
  color: #64748b;
  font-size: 18px;
}

.option-body {
  padding: 0 14px 14px 42px;
}

.option-body.is-inline {
  display: grid;
  grid-template-columns: minmax(96px, 1fr) minmax(122px, auto);
  gap: 12px;
  align-items: end;
}

.option-body input {
  width: 100%;
  height: 34px;
  border: 0;
  border-bottom: 2px solid rgba(15, 23, 42, 0.82);
  outline: none;
  color: #0f172a;
  background: transparent;
  font-size: 15px;
}

.option-body select {
  height: 34px;
  border: 0;
  border-radius: 10px;
  outline: none;
  color: #10213b;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.22);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.option-body small {
  margin-top: 6px;
}

.share-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 18px;
}

.share-footer button {
  min-width: 78px;
  height: 42px;
  border: 0;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
}

.share-footer .plain {
  color: #1976d2;
  background: transparent;
}

.share-footer .primary {
  color: #fff;
  background: linear-gradient(135deg, #2f7df6, #159de8);
  box-shadow: 0 10px 22px rgba(47, 125, 246, 0.24);
}

.share-footer .primary:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.result-box {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  min-height: 68px;
  padding: 14px;
  border-radius: 16px;
}

.result-box strong {
  display: block;
  overflow: hidden;
  margin-top: 3px;
  color: #0f172a;
  font-size: 16px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 720px) {
  .share-modal {
    padding: 20px 16px;
  }

  .share-header h2 {
    font-size: 20px;
  }
}
</style>
