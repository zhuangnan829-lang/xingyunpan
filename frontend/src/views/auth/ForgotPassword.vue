<template>
  <div class="forgot-page">
    <div class="page-shell">
      <header class="site-header">
        <div class="brand-block" aria-label="星云盘品牌">
          <div class="brand-mark" aria-hidden="true">
            <span class="brand-mark-core"></span>
            <span class="brand-mark-glow"></span>
          </div>
          <div class="brand-copy">
            <span class="brand-kicker">XINGYUNPAN</span>
            <strong class="brand-name">星云盘</strong>
          </div>
        </div>

        <nav class="header-links" aria-label="快捷导航">
          <router-link to="/login" class="header-link">返回登录</router-link>
          <router-link to="/register" class="header-link">创建账号</router-link>
        </nav>
      </header>

      <main class="hero-layout">
        <section class="hero-copy" aria-labelledby="forgot-title">
          <div class="eyebrow">找回密码</div>
          <h1 id="forgot-title" class="hero-title">通过邮箱验证码重置密码</h1>
          <p class="hero-subtitle">
            为已注册邮箱发送安全验证码，完成校验后即可设置新的登录密码，整个过程清晰、正式、可信。
          </p>

          <div class="hero-panel">
            <div class="hero-panel-content">
              <span class="panel-label">安全重置流程</span>
              <p class="panel-title">已注册邮箱专属验证，重置过程更稳妥</p>
              <p class="panel-text">
                系统仅向已注册邮箱发送验证码，并在短时有效期内完成校验，兼顾找回效率与账户安全。
              </p>
            </div>

            <div class="feature-grid" aria-label="重置特点">
              <article class="feature-card">
                <span class="feature-value">邮箱校验</span>
                <span class="feature-label">仅已注册邮箱可接收验证码</span>
              </article>
              <article class="feature-card">
                <span class="feature-value">短时有效</span>
                <span class="feature-label">验证码按时效限制自动失效</span>
              </article>
              <article class="feature-card">
                <span class="feature-value">立即生效</span>
                <span class="feature-label">重置完成后可直接使用新密码登录</span>
              </article>
            </div>
          </div>
        </section>

        <section class="form-panel" aria-label="找回密码表单区域">
          <div class="form-card">
            <div class="card-top">
              <p class="card-kicker">账户恢复</p>
              <h2 class="card-title">设置新的登录密码</h2>
              <p class="card-description">请输入已注册邮箱，完成验证码校验后重置密码。</p>
            </div>

            <el-form
              ref="formRef"
              :model="form"
              :rules="rules"
              class="forgot-form"
              label-position="top"
              status-icon
              @submit.prevent="handleSubmit"
            >
              <el-form-item label="邮箱" prop="email">
                <div class="email-row">
                  <el-input
                    v-model.trim="form.email"
                    class="form-input"
                    size="large"
                    placeholder="请输入已注册邮箱"
                    clearable
                  >
                    <template #prefix>
                      <el-icon><Message /></el-icon>
                    </template>
                  </el-input>

                  <el-button
                    class="send-code-button"
                    :loading="sendingCode"
                    :disabled="countdown > 0"
                    @click="handleSendCode"
                  >
                    {{ countdown > 0 ? `${countdown}s 后重发` : '发送验证码' }}
                  </el-button>
                </div>
              </el-form-item>

              <el-form-item label="邮箱验证码" prop="email_code">
                <el-input
                  v-model.trim="form.email_code"
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

              <el-form-item label="新密码" prop="new_password">
                <el-input
                  v-model.trim="form.new_password"
                  class="form-input"
                  type="password"
                  size="large"
                  placeholder="请输入新密码，至少 6 位"
                  show-password
                  clearable
                >
                  <template #prefix>
                    <el-icon><Lock /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item label="确认新密码" prop="confirm_password">
                <el-input
                  v-model.trim="form.confirm_password"
                  class="form-input"
                  type="password"
                  size="large"
                  placeholder="请再次输入新密码"
                  show-password
                  clearable
                  @keyup.enter="handleSubmit"
                >
                  <template #prefix>
                    <el-icon><Check /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <CaptchaRuntime
                scene="reset_password"
                path="/api/v1/user/password/reset"
                :identity="form.email"
                @update:payload="captchaPayload = $event"
              />

              <el-form-item class="submit-item">
                <el-button type="primary" :loading="loading" class="submit-button" @click="handleSubmit">
                  {{ loading ? '正在重置...' : '确认重置密码' }}
                </el-button>
              </el-form-item>
            </el-form>

            <div class="conversion-panel">
              <p class="conversion-title">想起密码了？</p>
              <p class="conversion-text">返回登录页，继续访问您的账户与服务管理入口。</p>
              <div class="conversion-actions">
                <router-link to="/login" class="conversion-link">
                  返回登录
                  <el-icon><ArrowRight /></el-icon>
                </router-link>
                <router-link to="/register" class="secondary-link">注册账号</router-link>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { ArrowRight, Check, Key, Lock, Message } from '@element-plus/icons-vue';
import { resetPasswordByEmailCode, sendResetPasswordEmailCode } from '@/api/user';
import CaptchaRuntime from '@/components/CaptchaRuntime/index.vue';
import type { CaptchaPayload } from '@/api/captcha';

const router = useRouter();
const route = useRoute();
const formRef = ref<FormInstance>();
const loading = ref(false);
const sendingCode = ref(false);
const countdown = ref(0);
const captchaPayload = ref<CaptchaPayload>({});
let countdownTimer: ReturnType<typeof setInterval> | null = null;

const form = reactive({
  email: '',
  email_code: '',
  new_password: '',
  confirm_password: '',
});

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const startCountdown = (seconds: number): void => {
  countdown.value = seconds;
  if (countdownTimer) {
    clearInterval(countdownTimer);
  }
  countdownTimer = setInterval(() => {
    countdown.value -= 1;
    if (countdown.value <= 0 && countdownTimer) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
  }, 1000);
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
    callback(new Error('请输入新密码'));
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
    callback(new Error('请再次输入新密码'));
    return;
  }
  if (value !== form.new_password) {
    callback(new Error('两次输入的密码不一致'));
    return;
  }
  callback();
};

const rules: FormRules = {
  email: [{ validator: validateEmail, trigger: 'blur' }],
  email_code: [{ validator: validateEmailCode, trigger: 'blur' }],
  new_password: [{ validator: validatePassword, trigger: 'blur' }],
  confirm_password: [{ validator: validateConfirmPassword, trigger: 'blur' }],
};

const handleSendCode = async (): Promise<void> => {
  const email = form.email.trim();
  if (!emailRegex.test(email)) {
    ElMessage.warning('请先输入有效的已注册邮箱');
    return;
  }
  if (countdown.value > 0 || sendingCode.value) {
    return;
  }

  try {
    sendingCode.value = true;
    await sendResetPasswordEmailCode(email, captchaPayload.value);
    ElMessage.success('验证码已发送，请查收邮箱');
    startCountdown(60);
  } catch (error: any) {
    ElMessage.error(error?.message || '验证码发送失败，请稍后重试');
  } finally {
    sendingCode.value = false;
  }
};

const handleSubmit = async (): Promise<void> => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    loading.value = true;

    await resetPasswordByEmailCode({
      email: form.email,
      email_code: form.email_code,
      new_password: form.new_password,
      ...captchaPayload.value,
    });

    ElMessage.success('密码重置成功，请重新登录');
    router.push('/login');
  } catch (error: any) {
    if (error !== false) {
      ElMessage.error(error?.message || '密码重置失败，请稍后重试');
    }
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  const email = Array.isArray(route.query.email) ? route.query.email[0] : route.query.email;
  const emailCode = Array.isArray(route.query.email_code) ? route.query.email_code[0] : route.query.email_code;

  if (typeof email === 'string' && email.trim()) {
    form.email = email.trim();
  }
  if (typeof emailCode === 'string' && emailCode.trim()) {
    form.email_code = emailCode.trim();
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
.forgot-page {
  --page-bg: #f4f7fb;
  --surface: rgba(255, 255, 255, 0.88);
  --text-primary: #0f172a;
  --text-secondary: #52607a;
  --text-tertiary: #7e8aa3;
  --brand-500: #2563eb;
  --brand-600: #1d4ed8;
  --brand-700: #163ea7;
  --shadow-soft: 0 24px 60px rgba(15, 23, 42, 0.08);
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(96, 165, 250, 0.18), transparent 32%),
    radial-gradient(circle at 85% 20%, rgba(59, 130, 246, 0.12), transparent 18%),
    linear-gradient(180deg, #fbfdff 0%, var(--page-bg) 100%);
  color: var(--text-primary);
}

.page-shell {
  position: relative;
  min-height: 100vh;
  padding: 24px;
  max-width: 1320px;
  margin: 0 auto;
}

.site-header,
.hero-layout {
  position: relative;
  z-index: 1;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 28px;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-mark {
  position: relative;
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid rgba(96, 165, 250, 0.2);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.brand-mark-core,
.brand-mark-glow {
  position: absolute;
  border-radius: 999px;
}

.brand-mark-core {
  inset: 11px;
  background: linear-gradient(135deg, #2563eb 0%, #60a5fa 100%);
}

.brand-mark-glow {
  width: 14px;
  height: 14px;
  right: 6px;
  top: 6px;
  background: rgba(147, 197, 253, 0.8);
  filter: blur(4px);
}

.brand-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.brand-kicker {
  font-size: 12px;
  letter-spacing: 0.18em;
  color: var(--text-tertiary);
}

.brand-name {
  font-size: 20px;
  font-weight: 700;
}

.header-links {
  display: flex;
  align-items: center;
  gap: 18px;
}

.header-link,
.forgot-link,
.secondary-link,
.conversion-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.header-link:hover,
.secondary-link:hover,
.forgot-link:hover {
  color: var(--brand-600);
}

.hero-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(420px, 480px);
  gap: 40px;
  align-items: center;
  min-height: calc(100vh - 140px);
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.08);
  color: var(--brand-600);
  font-size: 13px;
  font-weight: 600;
}

.hero-title {
  margin: 18px 0 14px;
  font-size: clamp(36px, 4vw, 54px);
  line-height: 1.08;
  letter-spacing: -0.03em;
}

.hero-subtitle {
  margin: 0;
  max-width: 580px;
  font-size: 17px;
  line-height: 1.8;
  color: var(--text-secondary);
}

.hero-panel {
  margin-top: 32px;
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.58);
  border: 1px solid rgba(148, 163, 184, 0.15);
  box-shadow: 0 20px 55px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(18px);
}

.panel-label {
  display: inline-block;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--brand-600);
}

.panel-title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.panel-text {
  margin: 14px 0 0;
  font-size: 15px;
  line-height: 1.8;
  color: var(--text-secondary);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 26px;
}

.feature-card {
  padding: 18px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.feature-value {
  display: block;
  margin-bottom: 8px;
  color: var(--brand-700);
  font-size: 15px;
  font-weight: 700;
}

.feature-label {
  color: var(--text-secondary);
  line-height: 1.7;
  font-size: 13px;
}

.form-card {
  padding: 34px;
  border-radius: 28px;
  background: var(--surface);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: var(--shadow-soft);
  backdrop-filter: blur(18px);
}

.card-top {
  margin-bottom: 24px;
}

.card-kicker {
  margin: 0 0 8px;
  color: var(--brand-600);
  font-size: 13px;
  font-weight: 600;
}

.card-title {
  margin: 0;
  font-size: 28px;
}

.card-description {
  margin: 10px 0 0;
  color: var(--text-secondary);
  line-height: 1.75;
}

.email-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 136px;
  gap: 12px;
  width: 100%;
}

:deep(.form-input .el-input__wrapper) {
  min-height: 52px;
  padding: 0 16px;
  border-radius: 16px;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.22);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

:deep(.form-input .el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(96, 165, 250, 0.38);
}

:deep(.form-input.is-focus .el-input__wrapper) {
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.55), 0 0 0 4px rgba(37, 99, 235, 0.08);
  transform: translateY(-1px);
}

.send-code-button,
.submit-button {
  border: none;
  border-radius: 16px;
}

.send-code-button {
  min-height: 52px;
  color: var(--brand-700);
  background: rgba(37, 99, 235, 0.08);
}

.send-code-button:hover {
  color: #fff;
  background: linear-gradient(135deg, var(--brand-500), var(--brand-600));
}

.submit-button {
  width: 100%;
  min-height: 54px;
  font-size: 16px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--brand-500), var(--brand-600));
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.22);
}

.submit-button:hover {
  background: linear-gradient(135deg, #2b6bf0, var(--brand-700));
}

.conversion-panel {
  margin-top: 26px;
  padding-top: 22px;
  border-top: 1px solid rgba(148, 163, 184, 0.16);
}

.conversion-title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
}

.conversion-text {
  margin: 8px 0 0;
  color: var(--text-secondary);
  line-height: 1.75;
}

.conversion-actions {
  display: flex;
  align-items: center;
  gap: 18px;
  margin-top: 16px;
}

.conversion-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--brand-600);
  font-weight: 600;
}

@media (max-width: 1100px) {
  .hero-layout {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .form-panel {
    max-width: 560px;
  }
}

@media (max-width: 767px) {
  .page-shell {
    padding: 18px;
  }

  .site-header,
  .header-links,
  .conversion-actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-title {
    font-size: 34px;
  }

  .hero-panel,
  .form-card {
    padding: 22px;
    border-radius: 24px;
  }

  .feature-grid,
  .email-row {
    grid-template-columns: 1fr;
  }

  .send-code-button {
    width: 100%;
  }
}
</style>
