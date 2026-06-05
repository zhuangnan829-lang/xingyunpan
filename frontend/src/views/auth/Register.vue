<template>
  <div class="register-page">
    <div class="page-shell">
      <header class="site-header">
        <div class="brand-lockup" aria-label="星云盘品牌">
          <div class="brand-mark" aria-hidden="true">
            <span class="brand-mark-core"></span>
            <span class="brand-mark-ring"></span>
            <span class="brand-mark-glow"></span>
          </div>
          <div class="brand-copy">
            <span class="brand-kicker">XINGYUNPAN</span>
            <strong class="brand-name">星云盘</strong>
          </div>
        </div>

        <nav class="header-links" aria-label="快捷导航">
          <a :href="supportLink" class="header-link">服务支持</a>
          <router-link to="/login" class="header-link">返回登录</router-link>
        </nav>
      </header>

      <main class="hero-layout">
        <section class="hero-panel" aria-labelledby="register-title">
          <div class="hero-badge">新用户注册</div>
          <h1 id="register-title" class="hero-title">创建您的星云盘账户</h1>
          <p class="hero-subtitle">通过邮箱验证码完成注册，快速启用统一、稳定、正式的账户体验。</p>

          <div class="hero-visual" aria-hidden="true">
            <div class="visual-grid"></div>
            <div class="visual-card visual-card-main">
              <div class="visual-card-header">
                <span class="visual-pill">Register</span>
                <span class="visual-status">已验证邮箱</span>
              </div>
              <div class="visual-lines">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <div class="visual-metrics">
                <article>
                  <strong>邮箱验证</strong>
                  <span>确认账户真实可用</span>
                </article>
                <article>
                  <strong>60s</strong>
                  <span>验证码发送冷却</span>
                </article>
              </div>
            </div>
            <div class="visual-card visual-card-side">
              <span class="visual-side-kicker">账户体系</span>
              <p>从注册、验证到登录，保持一致、清晰、专业的产品体验。</p>
            </div>
          </div>

          <div class="trust-grid" aria-label="注册价值">
            <article class="trust-card">
              <span class="trust-label">邮箱校验</span>
              <p>通过邮件验证码确认账户有效性，降低错误注册与后续找回风险。</p>
            </article>
            <article class="trust-card">
              <span class="trust-label">统一账户</span>
              <p>注册完成后即可进入同一套授权、服务与节点管理入口。</p>
            </article>
            <article class="trust-card">
              <span class="trust-label">正式体验</span>
              <p>从首次进入开始即传达清晰、可信、稳定的产品感受。</p>
            </article>
          </div>
        </section>

        <section class="form-panel" aria-label="注册表单区域">
          <div class="form-card">
            <div class="card-header">
              <p class="card-kicker">注册</p>
              <h2 class="card-title">创建新的账户</h2>
              <p class="card-description">请填写基本信息，并完成邮箱验证码验证后继续。</p>
            </div>

            <el-alert
              v-if="submitError"
              class="inline-alert"
              type="error"
              :closable="false"
              show-icon
              :title="submitError"
            />

            <el-form
              ref="registerFormRef"
              :model="registerForm"
              :rules="registerRules"
              class="register-form"
              label-position="top"
              status-icon
              @submit.prevent="handleRegister"
            >
              <el-form-item label="用户名" prop="username">
                <el-input
                  v-model.trim="registerForm.username"
                  class="form-input"
                  size="large"
                  placeholder="请输入用户名"
                  clearable
                >
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item label="邮箱" prop="email">
                <div class="email-row">
                  <el-input
                    v-model.trim="registerForm.email"
                    class="form-input"
                    size="large"
                    placeholder="请输入常用邮箱"
                    clearable
                  >
                    <template #prefix>
                      <el-icon><Message /></el-icon>
                    </template>
                  </el-input>

                  <el-button
                    class="send-code-button"
                    :loading="sendingCode"
                    :disabled="sendingCode || countdown > 0"
                    @click="handleSendCode"
                  >
                    {{ countdown > 0 ? `${countdown}s 后重发` : '发送验证码' }}
                  </el-button>
                </div>
              </el-form-item>

              <el-form-item label="邮箱验证码" prop="email_code">
                <el-input
                  v-model.trim="registerForm.email_code"
                  class="form-input"
                  size="large"
                  maxlength="6"
                  placeholder="请输入 6 位验证码"
                  clearable
                >
                  <template #prefix>
                    <el-icon><Key /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item label="密码" prop="password">
                <el-input
                  v-model.trim="registerForm.password"
                  class="form-input"
                  type="password"
                  size="large"
                  placeholder="请输入密码，至少 6 位"
                  show-password
                  clearable
                >
                  <template #prefix>
                    <el-icon><Lock /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item label="确认密码" prop="confirm_password">
                <el-input
                  v-model.trim="registerForm.confirm_password"
                  class="form-input"
                  type="password"
                  size="large"
                  placeholder="请再次输入密码"
                  show-password
                  clearable
                  @keyup.enter="handleRegister"
                >
                  <template #prefix>
                    <el-icon><CircleCheck /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <CaptchaRuntime
                ref="captchaRef"
                scene="register"
                path="/api/v1/user/register"
                :identity="registerForm.email"
                @update:payload="captchaPayload = $event"
              />

              <el-form-item class="submit-item">
                <el-button type="primary" :loading="loading" class="submit-button" @click="handleRegister">
                  {{ loading ? '正在创建账户...' : '注册' }}
                </el-button>
              </el-form-item>
            </el-form>

            <div class="conversion-panel">
              <div class="conversion-copy">
                <p class="conversion-title">已经有账户？</p>
                <p class="conversion-text">返回登录页，继续访问您的授权工作台与服务入口。</p>
              </div>
              <router-link to="/login" class="conversion-link">立即登录</router-link>
            </div>
          </div>
        </section>
      </main>

      <footer class="site-footer">
        <div class="footer-brand">
          <div class="brand-mark brand-mark-small" aria-hidden="true">
            <span class="brand-mark-core"></span>
            <span class="brand-mark-ring"></span>
            <span class="brand-mark-glow"></span>
          </div>
          <div>
            <p class="footer-brand-name">星云盘</p>
            <p class="footer-brand-subtitle">账户注册系统</p>
          </div>
        </div>

        <div class="footer-meta">
          <span>© {{ currentYear }} 星云盘. 保留所有权利</span>
          <a :href="supportLink" class="footer-link">服务支持</a>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { CircleCheck, Key, Lock, Message, User } from '@element-plus/icons-vue';
import { sendRegisterEmailCode } from '@/api/user';
import { useUserStore } from '@/stores/user';
import CaptchaRuntime from '@/components/CaptchaRuntime/index.vue';
import type { CaptchaPayload } from '@/api/captcha';

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();

const registerFormRef = ref<FormInstance>();
const loading = ref(false);
const sendingCode = ref(false);
const countdown = ref(0);
const submitError = ref('');
const captchaPayload = ref<CaptchaPayload>({});
const captchaRef = ref<{ reload: () => Promise<void> } | null>(null);
const currentYear = new Date().getFullYear();
const supportLink = 'mailto:support@xingyunpan.com';
let countdownTimer: ReturnType<typeof setInterval> | null = null;

const registerForm = reactive({
  username: '',
  email: '',
  email_code: '',
  password: '',
  confirm_password: '',
});

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const startCountdown = (seconds: number): void => {
  countdown.value = seconds;
  if (countdownTimer) clearInterval(countdownTimer);
  countdownTimer = setInterval(() => {
    countdown.value -= 1;
    if (countdown.value <= 0 && countdownTimer) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
  }, 1000);
};

const validateUsername = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入用户名'));
    return;
  }
  if (value.length < 3) {
    callback(new Error('用户名长度至少为 3 位'));
    return;
  }
  callback();
};

const validateEmail = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入邮箱'));
    return;
  }
  if (!emailRegex.test(value)) {
    callback(new Error('请输入有效的邮箱地址'));
    return;
  }
  callback();
};

const validateEmailCode = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入邮箱验证码'));
    return;
  }
  if (!/^\d{6}$/.test(value)) {
    callback(new Error('验证码应为 6 位数字'));
    return;
  }
  callback();
};

const validatePassword = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入密码'));
    return;
  }
  if (value.length < 6) {
    callback(new Error('密码长度至少为 6 位'));
    return;
  }
  callback();
};

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请再次输入密码'));
    return;
  }
  if (value !== registerForm.password) {
    callback(new Error('两次输入的密码不一致'));
    return;
  }
  callback();
};

const registerRules: FormRules = {
  username: [{ validator: validateUsername, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }],
  email_code: [{ validator: validateEmailCode, trigger: 'blur' }],
  password: [{ validator: validatePassword, trigger: 'blur' }],
  confirm_password: [{ validator: validateConfirmPassword, trigger: 'blur' }],
};

const handleSendCode = async (): Promise<void> => {
  const email = registerForm.email.trim();
  submitError.value = '';

  if (!emailRegex.test(email)) {
    ElMessage.warning('请先输入有效的邮箱地址');
    return;
  }
  if (countdown.value > 0 || sendingCode.value) {
    return;
  }

  try {
    sendingCode.value = true;
    await sendRegisterEmailCode(email, captchaPayload.value);
    ElMessage.success('验证码已发送，请注意查收邮箱');
    startCountdown(60);
    await captchaRef.value?.reload();
  } catch (error: any) {
    submitError.value = error?.message || '验证码发送失败，请稍后重试';
    ElMessage.error(submitError.value);
    await captchaRef.value?.reload();
  } finally {
    sendingCode.value = false;
  }
};

const handleRegister = async (): Promise<void> => {
  if (!registerFormRef.value || loading.value) return;
  submitError.value = '';

  try {
    await registerFormRef.value.validate();
    loading.value = true;

    await userStore.register({
      username: registerForm.username,
      email: registerForm.email,
      email_code: registerForm.email_code,
      password: registerForm.password,
      ...captchaPayload.value,
    });

    ElMessage.success('注册成功，请登录');
    router.push('/login');
  } catch (error: any) {
    if (error !== false) {
      submitError.value = error?.message || '注册失败，请稍后重试';
      ElMessage.error(submitError.value);
      await captchaRef.value?.reload();
    }
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  const email = Array.isArray(route.query.email) ? route.query.email[0] : route.query.email;
  const emailCode = Array.isArray(route.query.email_code) ? route.query.email_code[0] : route.query.email_code;

  if (typeof email === 'string' && email.trim()) {
    registerForm.email = email.trim();
  }
  if (typeof emailCode === 'string' && emailCode.trim()) {
    registerForm.email_code = emailCode.trim();
  }
});

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
});
</script>

<style scoped>
.register-page {
  --page-bg: #eef4fb;
  --surface: rgba(255, 255, 255, 0.84);
  --text-primary: #0f172a;
  --text-secondary: #5b6b86;
  --text-tertiary: #8290aa;
  --brand-500: #2962ea;
  --brand-600: #1f4fd0;
  --brand-700: #173f9b;
  --shadow-soft: 0 30px 80px rgba(15, 23, 42, 0.08);
  min-height: 100vh;
  background:
    radial-gradient(circle at 12% 6%, rgba(96, 165, 250, 0.16), transparent 26%),
    radial-gradient(circle at 88% 10%, rgba(186, 230, 253, 0.22), transparent 20%),
    linear-gradient(180deg, #fbfdff 0%, #f4f8fd 46%, #edf3fa 100%);
  color: var(--text-primary);
}

.page-shell {
  position: relative;
  min-height: 100vh;
  max-width: 1380px;
  margin: 0 auto;
  padding: 28px 32px 22px;
}

.page-shell::before,
.page-shell::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
  filter: blur(26px);
}

.page-shell::before {
  width: 320px;
  height: 320px;
  left: -100px;
  top: 90px;
  background: linear-gradient(180deg, rgba(191, 219, 254, 0.65), rgba(255, 255, 255, 0));
}

.page-shell::after {
  width: 260px;
  height: 260px;
  right: 36px;
  bottom: 120px;
  background: linear-gradient(180deg, rgba(147, 197, 253, 0.5), rgba(255, 255, 255, 0));
}

.site-header,
.hero-layout,
.site-footer {
  position: relative;
  z-index: 1;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 34px;
}

.brand-lockup,
.footer-brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-mark {
  position: relative;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(219, 234, 254, 0.92) 100%);
  border: 1px solid rgba(191, 219, 254, 0.86);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 20px 34px rgba(148, 163, 184, 0.12);
}

.brand-mark-small {
  width: 38px;
  height: 38px;
  border-radius: 14px;
}

.brand-mark-core,
.brand-mark-ring,
.brand-mark-glow {
  position: absolute;
  border-radius: 999px;
}

.brand-mark-core {
  inset: 10px;
  background: linear-gradient(135deg, #2563eb 0%, #38bdf8 100%);
}

.brand-mark-ring {
  inset: 6px;
  border: 1px solid rgba(255, 255, 255, 0.7);
}

.brand-mark-glow {
  width: 18px;
  height: 18px;
  right: 6px;
  top: 6px;
  background: rgba(255, 255, 255, 0.86);
}

.brand-copy {
  display: grid;
  gap: 4px;
}

.brand-kicker,
.card-kicker,
.hero-badge {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--brand-600);
  text-transform: uppercase;
}

.brand-name,
.footer-brand-name {
  font-size: 22px;
  line-height: 1.1;
  font-weight: 700;
}

.header-links {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-link,
.footer-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.header-link:hover,
.footer-link:hover {
  color: var(--brand-600);
}

.hero-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(430px, 0.92fr);
  gap: 44px;
  align-items: center;
  min-height: calc(100vh - 170px);
}

.hero-panel {
  padding: 24px 0 24px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 14px;
  border: 1px solid rgba(96, 165, 250, 0.24);
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.92));
  box-shadow: 0 10px 22px rgba(37, 99, 235, 0.08);
}

.hero-title {
  max-width: 10ch;
  margin: 18px 0 14px;
  font-size: clamp(42px, 5vw, 64px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  font-weight: 800;
}

.hero-subtitle {
  max-width: 620px;
  margin: 0 0 28px;
  color: var(--text-secondary);
  font-size: 20px;
  line-height: 1.8;
}

.hero-visual {
  position: relative;
  overflow: hidden;
  min-height: 328px;
  margin-bottom: 24px;
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 36px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(245, 249, 255, 0.9)),
    linear-gradient(135deg, rgba(219, 234, 254, 0.26), rgba(255, 255, 255, 0));
  box-shadow: 0 28px 80px rgba(148, 163, 184, 0.12);
  backdrop-filter: blur(12px);
}

.visual-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(219, 234, 254, 0.34) 1px, transparent 1px),
    linear-gradient(90deg, rgba(219, 234, 254, 0.34) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: linear-gradient(180deg, rgba(15, 23, 42, 0.22), transparent 84%);
}

.visual-card {
  position: absolute;
  border-radius: 28px;
  border: 1px solid rgba(219, 234, 254, 0.96);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(16px);
}

.visual-card-main {
  left: 34px;
  top: 34px;
  width: min(430px, calc(100% - 112px));
  padding: 24px;
  box-shadow: 0 26px 56px rgba(37, 99, 235, 0.14);
}

.visual-card-side {
  right: 34px;
  bottom: 32px;
  width: 220px;
  padding: 18px 18px 20px;
  box-shadow: 0 18px 40px rgba(59, 130, 246, 0.1);
}

.visual-card-header,
.visual-metrics {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.visual-pill,
.visual-status {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.visual-pill {
  color: var(--brand-700);
  background: rgba(219, 234, 254, 0.78);
}

.visual-status {
  color: #0f766e;
  background: rgba(204, 251, 241, 0.9);
}

.visual-lines {
  display: grid;
  gap: 12px;
  margin: 26px 0 24px;
}

.visual-lines span {
  display: block;
  height: 12px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.16), rgba(59, 130, 246, 0.04));
}

.visual-lines span:nth-child(1) {
  width: 72%;
}

.visual-lines span:nth-child(2) {
  width: 88%;
}

.visual-lines span:nth-child(3) {
  width: 54%;
}

.visual-metrics article {
  display: grid;
  gap: 6px;
}

.visual-metrics strong {
  font-size: 26px;
  line-height: 1.1;
  color: var(--text-primary);
}

.visual-metrics span,
.visual-card-side p,
.trust-card p,
.card-description,
.conversion-text,
.footer-brand-subtitle {
  color: var(--text-secondary);
}

.visual-card-side p {
  margin: 8px 0 0;
  line-height: 1.75;
  font-size: 14px;
}

.visual-side-kicker {
  color: var(--brand-600);
  font-size: 13px;
  font-weight: 700;
}

.trust-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.trust-card {
  padding: 18px 18px 20px;
  border: 1px solid rgba(226, 232, 240, 0.94);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 14px 32px rgba(148, 163, 184, 0.06);
  backdrop-filter: blur(12px);
}

.trust-label {
  display: inline-block;
  margin-bottom: 10px;
  color: var(--brand-700);
  font-size: 15px;
  font-weight: 700;
}

.trust-card p {
  margin: 0;
  font-size: 14px;
  line-height: 1.8;
}

.form-panel {
  display: flex;
  justify-content: flex-end;
}

.form-card {
  width: 100%;
  max-width: 560px;
  padding: 34px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  border-radius: 34px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 251, 255, 0.84));
  box-shadow: 0 36px 90px rgba(148, 163, 184, 0.16);
  backdrop-filter: blur(20px);
}

.card-header {
  margin-bottom: 18px;
}

.card-title {
  margin: 10px 0;
  font-size: 32px;
  line-height: 1.18;
  letter-spacing: -0.02em;
}

.card-description {
  margin: 0;
  line-height: 1.8;
}

.inline-alert {
  margin-bottom: 18px;
  border-radius: 18px;
}

.register-form :deep(.el-form-item__label) {
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 8px;
}

.register-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.form-input :deep(.el-input__wrapper) {
  min-height: 56px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 0 0 1px rgba(203, 213, 225, 0.74) inset;
  transition: box-shadow 0.2s ease, transform 0.2s ease, background 0.2s ease;
}

.form-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.9) inset;
}

.form-input :deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.98);
  box-shadow:
    0 0 0 1px rgba(41, 98, 234, 0.88) inset,
    0 0 0 4px rgba(41, 98, 234, 0.08);
  transform: translateY(-1px);
}

.form-input :deep(.el-input__inner) {
  color: var(--text-primary);
  font-size: 15px;
}

.form-input :deep(.el-input__prefix-inner),
.form-input :deep(.el-input__suffix-inner) {
  color: #7c8fad;
}

.email-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 154px;
  gap: 12px;
  width: 100%;
}

.send-code-button {
  min-height: 56px;
  border-radius: 18px;
  border: 1px solid rgba(191, 219, 254, 0.9);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 246, 255, 0.98));
  color: var(--brand-600);
  font-weight: 700;
}

.send-code-button:hover {
  border-color: rgba(37, 99, 235, 0.35);
  color: var(--brand-700);
  background: linear-gradient(180deg, #ffffff, #eef5ff);
}

.submit-item {
  margin-bottom: 10px;
}

.submit-button {
  width: 100%;
  min-height: 56px;
  border: none;
  border-radius: 18px;
  background: linear-gradient(135deg, #2b60e8 0%, #4d86ff 50%, #376df0 100%);
  box-shadow: 0 22px 42px rgba(41, 98, 234, 0.24);
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.submit-button:hover {
  background: linear-gradient(135deg, #285ce0 0%, #447ef8 50%, #3369e7 100%);
}

.conversion-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 4px;
  padding: 20px 22px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(252, 254, 255, 0.92), rgba(243, 247, 253, 0.9));
}

.conversion-copy {
  max-width: 320px;
}

.conversion-title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 700;
}

.conversion-text {
  margin: 0;
  font-size: 14px;
  line-height: 1.75;
}

.conversion-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 118px;
  min-height: 44px;
  padding: 0 18px;
  color: #1f4fd0;
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  border: 1px solid rgba(191, 219, 254, 0.9);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 10px 20px rgba(148, 163, 184, 0.08);
}

.site-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 16px 0 4px;
}

.footer-meta {
  display: flex;
  align-items: center;
  gap: 18px;
  color: var(--text-tertiary);
  font-size: 13px;
}

.footer-brand-name {
  margin: 0 0 4px;
  font-size: 15px;
}

.footer-brand-subtitle {
  margin: 0;
  font-size: 13px;
}

@media (max-width: 1180px) {
  .hero-layout {
    grid-template-columns: 1fr;
    gap: 28px;
    min-height: auto;
  }

  .hero-title {
    max-width: none;
  }

  .form-panel {
    justify-content: flex-start;
  }

  .form-card {
    max-width: none;
  }
}

@media (max-width: 768px) {
  .page-shell {
    padding: 16px;
  }

  .site-header {
    margin-bottom: 18px;
    align-items: flex-start;
    flex-direction: column;
  }

  .header-links {
    gap: 14px;
  }

  .hero-title {
    margin-top: 14px;
    font-size: 34px;
    line-height: 1.12;
  }

  .hero-subtitle {
    margin-bottom: 20px;
    font-size: 15px;
    line-height: 1.8;
  }

  .hero-visual,
  .form-card {
    border-radius: 24px;
  }

  .hero-visual {
    min-height: 278px;
  }

  .visual-card-main {
    left: 18px;
    top: 18px;
    width: calc(100% - 36px);
    padding: 18px;
  }

  .visual-card-side {
    right: 18px;
    bottom: 18px;
    width: calc(100% - 60px);
  }

  .visual-metrics {
    align-items: flex-start;
    flex-direction: column;
  }

  .trust-grid,
  .email-row {
    grid-template-columns: 1fr;
  }

  .form-card {
    padding: 22px;
  }

  .card-title {
    font-size: 26px;
  }

  .conversion-panel,
  .site-footer,
  .footer-meta {
    align-items: flex-start;
    flex-direction: column;
  }

  .conversion-link {
    width: 100%;
  }
}
</style>
