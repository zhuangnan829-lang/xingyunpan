<template>
  <div class="share-page">
    <!-- Loading state -->
    <div v-if="pageState === 'loading'" class="state-container">
      <el-skeleton :rows="5" animated />
    </div>

    <!-- Error: share not found -->
    <div v-else-if="pageState === 'not-found'" class="state-container">
      <el-result
        icon="warning"
        title="分享链接不存在"
        sub-title="该分享链接不存在或已被删除"
      >
        <template #extra>
          <el-button type="primary" @click="goHome">返回首页</el-button>
        </template>
      </el-result>
    </div>

    <!-- Error: share expired -->
    <div v-else-if="pageState === 'expired'" class="state-container">
      <el-result
        icon="warning"
        title="分享链接已过期"
        sub-title="该分享链接已过期，请联系分享者重新生成"
      >
        <template #extra>
          <el-button type="primary" @click="goHome">返回首页</el-button>
        </template>
      </el-result>
    </div>

    <!-- Password input state -->
    <div v-else-if="pageState === 'password'" class="state-container">
      <div class="password-card">
        <div class="password-header">
          <el-icon class="lock-icon"><Lock /></el-icon>
          <h2>访问密码</h2>
          <p class="password-hint">该分享链接已设置访问密码，请输入密码后访问</p>
        </div>

        <el-form @submit.prevent="handlePasswordSubmit" class="password-form">
          <el-form-item :error="passwordError">
            <el-input
              v-model="passwordInput"
              type="password"
              placeholder="请输入访问密码（4-8位）"
              show-password
              :disabled="isLocked"
              maxlength="8"
              @keyup.enter="handlePasswordSubmit"
            />
          </el-form-item>

          <div v-if="isLocked" class="lock-warning">
            <el-icon><Warning /></el-icon>
            密码错误次数过多，请 {{ lockCountdown }} 秒后重试
          </div>

          <el-button
            type="primary"
            :loading="verifying"
            :disabled="isLocked || !passwordInput"
            class="submit-btn"
            @click="handlePasswordSubmit"
          >
            确认访问
          </el-button>
        </el-form>
      </div>
    </div>

    <!-- File info state (access granted) -->
    <div v-else-if="pageState === 'file-info'" class="state-container">
      <div class="file-card">
        <div class="file-card-header">
          <el-icon class="share-icon"><Share /></el-icon>
          <h2>分享文件</h2>
        </div>

        <div class="file-info-meta">
          <div class="meta-item">
            <span class="meta-label">分享者：</span>
            <span class="meta-value">{{ shareInfo!.creator_name }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">创建时间：</span>
            <span class="meta-value">{{ formatExpiry(shareInfo!.created_at) }}</span>
          </div>
        </div>

        <div class="file-list">
          <div
            v-for="(fileName, index) in shareInfo!.file_names"
            :key="index"
            class="file-item"
          >
            <el-icon class="file-icon"><Document /></el-icon>
            <div class="file-details">
              <span class="file-name">{{ fileName }}</span>
            </div>
          </div>
        </div>

        <div class="share-meta">
          <div class="meta-item">
            <span class="meta-label">有效期至：</span>
            <span class="meta-value">{{ formatExpiry(shareInfo!.expires_at) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">下载次数：</span>
            <span class="meta-value">{{ shareInfo!.download_count }}</span>
          </div>
        </div>

        <div class="action-area">
          <el-button
            type="primary"
            size="large"
            :icon="Download"
            @click="handleActualDownload"
          >
            下载文件
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { Lock, Warning, Share, Document, Download } from '@element-plus/icons-vue';
import { useShareStore } from '@/stores/share';
import { getShareDownloadUrl } from '@/api/share';
import type { ShareInfo } from '@/api/share';

type PageState = 'loading' | 'not-found' | 'expired' | 'password' | 'file-info';

const MAX_PASSWORD_ATTEMPTS = 3;
const LOCK_DURATION_SECONDS = 300; // 5 minutes

const route = useRoute();
const router = useRouter();
const shareStore = useShareStore();

const pageState = ref<PageState>('loading');
const shareInfo = ref<ShareInfo | null>(null);
const passwordInput = ref('');
const passwordError = ref('');
const verifying = ref(false);
const failedAttempts = ref(0);
const isLocked = ref(false);
const lockCountdown = ref(0);

let lockTimer: ReturnType<typeof setInterval> | null = null;

// shareId is set once in onMounted and used by handlers
let currentShareId = '';

function formatExpiry(expiresAt: string | null): string {
  if (!expiresAt) return '永久有效';
  return new Date(expiresAt).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function goHome() {
  router.push('/');
}

function startLockTimer() {
  isLocked.value = true;
  lockCountdown.value = LOCK_DURATION_SECONDS;

  lockTimer = setInterval(() => {
    lockCountdown.value -= 1;
    if (lockCountdown.value <= 0) {
      isLocked.value = false;
      failedAttempts.value = 0;
      passwordError.value = '';
      if (lockTimer) {
        clearInterval(lockTimer);
        lockTimer = null;
      }
    }
  }, 1000);
}

async function handlePasswordSubmit() {
  if (isLocked.value || !passwordInput.value || !shareInfo.value) return;

  verifying.value = true;
  passwordError.value = '';

  try {
    const isValid = await shareStore.verifyPassword(currentShareId, passwordInput.value);
    
    if (isValid) {
      // Password verified, show file info
      pageState.value = 'file-info';
    } else {
      throw new Error('密码错误');
    }
  } catch (error: any) {
    failedAttempts.value += 1;

    if (failedAttempts.value >= MAX_PASSWORD_ATTEMPTS) {
      startLockTimer();
      passwordError.value = `密码错误次数过多，已锁定 ${LOCK_DURATION_SECONDS / 60} 分钟`;
    } else {
      passwordError.value = `密码错误，还剩 ${MAX_PASSWORD_ATTEMPTS - failedAttempts.value} 次机会`;
    }
    passwordInput.value = '';
  } finally {
    verifying.value = false;
  }
}

function handleActualDownload() {
  if (!shareInfo.value) return;

  const link = document.createElement('a');
  link.href = getShareDownloadUrl(currentShareId);
  link.rel = 'noopener';
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

async function handleDownload() {
  if (!shareInfo.value) return;

  try {
    // Increment download count
    await shareStore.incrementDownloadCount(currentShareId);
    
    // Show download message for each file
    shareInfo.value.file_names.forEach((fileName) => {
      ElMessage.success(`正在下载：${fileName}`);
    });
    
    // In a real implementation, you would trigger actual file downloads here
    // For now, we just show a message
  } catch (error: any) {
    ElMessage.error(error.message || '下载失败');
  }
}

onMounted(async () => {
  const shareId = (route.params.shareToken || route.params.shareId) as string;

  if (!shareId) {
    pageState.value = 'not-found';
    return;
  }

  currentShareId = shareId;

  try {
    const info = await shareStore.getShareInfo(shareId);
    shareInfo.value = info;

    // Check if expired
    if (shareStore.isExpired(info)) {
      pageState.value = 'expired';
      return;
    }

    if (info.has_password) {
      pageState.value = 'password';
    } else {
      // No password required, show file info directly
      pageState.value = 'file-info';
    }
  } catch (error: any) {
    const msg: string = error?.message ?? '';
    if (msg.includes('不存在') || msg.includes('已被删除') || msg.includes('404')) {
      pageState.value = 'not-found';
    } else if (msg.includes('已过期')) {
      pageState.value = 'expired';
    } else {
      pageState.value = 'not-found';
      ElMessage.error(error.message || '加载分享信息失败');
    }
  }
});

onUnmounted(() => {
  if (lockTimer) {
    clearInterval(lockTimer);
    lockTimer = null;
  }
  // Clear current share info when leaving the page
  shareStore.clearCurrentShareInfo();
});
</script>

<style scoped>
.share-page {
  min-height: 100vh;
  background-color: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.state-container {
  width: 100%;
  max-width: 560px;
}

/* Password card */
.password-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.password-header {
  margin-bottom: 32px;
}

.lock-icon {
  font-size: 48px;
  color: var(--el-color-primary);
  margin-bottom: 12px;
}

.password-header h2 {
  margin: 8px 0;
  font-size: 22px;
  color: #303133;
}

.password-hint {
  color: #909399;
  font-size: 14px;
  margin: 0;
}

.password-form {
  text-align: left;
}

.lock-warning {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-color-danger);
  font-size: 13px;
  margin-bottom: 12px;
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
}

/* File card */
.file-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.file-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 24px;
}

.share-icon {
  font-size: 28px;
  color: var(--el-color-primary);
}

.file-card-header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.file-info-meta {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.file-list {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 20px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid #f0f2f5;
}

.file-item:last-child {
  border-bottom: none;
}

.file-icon {
  font-size: 20px;
  color: var(--el-color-primary);
  flex-shrink: 0;
}

.file-details {
  flex: 1;
  overflow: hidden;
}

.file-name {
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.share-meta {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.meta-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
}

.meta-item:last-child {
  margin-bottom: 0;
}

.meta-label {
  color: #909399;
  width: 80px;
  flex-shrink: 0;
}

.meta-value {
  color: #303133;
}

.action-area {
  text-align: center;
}

.action-area .el-button {
  min-width: 160px;
}

/* Mobile responsive */
@media (max-width: 600px) {
  .share-page {
    padding: 16px;
    align-items: flex-start;
    padding-top: 40px;
  }

  .password-card,
  .file-card {
    padding: 24px 20px;
  }

  .password-header h2,
  .file-card-header h2 {
    font-size: 18px;
  }

  .password-hint {
    font-size: 13px;
  }

  .lock-icon,
  .share-icon {
    font-size: 36px;
  }

  .meta-label {
    width: 70px;
    font-size: 12px;
  }

  .meta-value {
    font-size: 12px;
  }

  .file-name {
    font-size: 13px;
  }

  .action-area .el-button {
    width: 100%;
    min-width: auto;
  }
}

@media (max-width: 400px) {
  .share-page {
    padding: 12px;
    padding-top: 30px;
  }

  .password-card,
  .file-card {
    padding: 20px 16px;
  }

  .password-header h2,
  .file-card-header h2 {
    font-size: 16px;
  }

  .lock-icon,
  .share-icon {
    font-size: 32px;
  }
}
</style>
