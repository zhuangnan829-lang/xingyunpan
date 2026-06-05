<template>
  <div class="login-page">
    <div class="page-shell">
      <header class="site-header">
        <div class="brand-group">
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
        </div>

        <el-dropdown trigger="click" class="header-menu">
          <button class="menu-button" type="button" aria-label="打开快捷菜单">
            <span></span>
            <span></span>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item>
                <a :href="docsLink" class="dropdown-link" target="_blank" rel="noreferrer">产品文档</a>
              </el-dropdown-item>
              <el-dropdown-item>
                <router-link to="/register" class="dropdown-link">创建账号</router-link>
              </el-dropdown-item>
              <el-dropdown-item>
                <router-link to="/forgot-password" class="dropdown-link">找回密码</router-link>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <main class="hero-layout">
        <section class="hero-panel" aria-labelledby="login-hero-title">
          <div class="hero-badge">官方授权入口</div>
          <h1 id="login-hero-title" class="hero-title">星云盘授权管理系统</h1>
          <p class="hero-subtitle">登录以管理您的授权、节点接入与账户服务，保持统一、稳定、正式的产品体验。</p>

          <div class="hero-visual" aria-hidden="true">
            <div class="visual-grid"></div>
            <div class="visual-card visual-card-main">
              <div class="visual-card-header">
                <span class="visual-pill">授权状态</span>
                <span class="visual-status">已验证</span>
              </div>
              <div class="visual-lines">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <div class="visual-metrics">
                <article>
                  <strong>授权版</strong>
                  <span>商业授权状态</span>
                </article>
                <article>
                  <strong>99.9%</strong>
                  <span>服务可用性目标</span>
                </article>
              </div>
            </div>
            <div class="visual-card visual-card-side">
              <span class="visual-side-kicker">产品概览</span>
              <p>统一的账户入口与清晰的状态反馈，适合长期稳定使用。</p>
            </div>
            <router-link to="/register" class="visual-card visual-card-entry">
              <span class="visual-entry-kicker">新用户入口</span>
              <strong class="visual-entry-title">创建账户</strong>
              <span class="visual-entry-text">完成邮箱验证后即可启用统一账户与授权服务。</span>
              <span class="visual-entry-action">立即开始</span>
            </router-link>
          </div>

          <div class="trust-grid" aria-label="系统特性">
            <article class="trust-card">
              <span class="trust-label">统一入口</span>
              <p>与官网保持一致的账户入口和服务结构，降低认知成本。</p>
            </article>
            <article class="trust-card">
              <span class="trust-label">正式体验</span>
              <p>视觉与交互克制有序，强化系统稳定、正规、可信的感受。</p>
            </article>
            <article class="trust-card">
              <span class="trust-label">安全校验</span>
              <p>通过标准化验证流程与清晰反馈机制，提升使用过程的可靠性。</p>
            </article>
          </div>
        </section>

        <section class="form-panel" aria-label="登录表单区域">
          <div class="form-card">
            <div class="card-header">
              <p class="card-kicker">登录</p>
              <h2 class="card-title">进入您的授权工作台</h2>
              <p class="card-description">请输入账户信息完成登录。系统围绕正式、稳定、清晰的产品体验进行设计。</p>
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
              ref="loginFormRef"
              :model="loginForm"
              :rules="loginRules"
              class="login-form"
              label-position="top"
              status-icon
              @submit.prevent="handleLogin"
            >
              <el-form-item label="邮箱" prop="email">
                <el-input
                  v-model.trim="loginForm.email"
                  class="form-input"
                  size="large"
                  placeholder="请输入您的邮箱"
                  clearable
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <el-icon><Message /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <div class="label-row">
                <span class="field-label">密码</span>
                <router-link to="/forgot-password" class="forgot-link">忘记密码？</router-link>
              </div>
              <el-form-item prop="password">
                <el-input
                  v-model.trim="loginForm.password"
                  class="form-input"
                  type="password"
                  size="large"
                  placeholder="请输入账户密码"
                  show-password
                  clearable
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <el-icon><Lock /></el-icon>
                  </template>
                </el-input>
              </el-form-item>

              <CaptchaRuntime
                ref="captchaRef"
                scene="login"
                path="/api/v1/user/login"
                :identity="loginForm.email"
                @update:payload="captchaPayload = $event"
              />

              <el-form-item class="submit-item">
                <el-button type="primary" :loading="loading" class="submit-button" @click="handleLogin">
                  {{ loading ? '正在验证身份...' : '登录' }}
                </el-button>
              </el-form-item>
            </el-form>

            <div class="register-spotlight">
              <div class="register-copy">
                <p class="register-title">首次使用星云盘？</p>
                <p class="register-text">立即创建账户，完成邮箱验证后即可进入统一的授权与服务管理入口。</p>
              </div>
              <router-link to="/register" class="register-link">立即注册</router-link>
            </div>

            <div class="conversion-panel">
              <div class="conversion-copy">
                <p class="conversion-title">面向长期使用的正式产品入口</p>
                <p class="conversion-text">统一的授权体系、清晰的状态信息与更稳定的账户体验，让登录本身也保持应有的专业感。</p>
              </div>
              <span class="conversion-badge">Official</span>
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
            <p class="footer-brand-subtitle">授权管理系统</p>
          </div>
        </div>

        <div class="footer-meta">
          <span>© {{ currentYear }} 星云盘. 保留所有权利</span>
          <a :href="docsLink" class="footer-link" target="_blank" rel="noreferrer">文档</a>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { Key, Lock, Message } from '@element-plus/icons-vue';
import { useUserStore } from '@/stores/user';
import CaptchaRuntime from '@/components/CaptchaRuntime/index.vue';
import type { CaptchaPayload } from '@/api/captcha';

const router = useRouter();
const userStore = useUserStore();

const loginFormRef = ref<FormInstance>();
const loading = ref(false);
const submitError = ref('');
const captchaSeed = ref('');
const captchaPayload = ref<CaptchaPayload>({});
const captchaRef = ref<{ reload: () => Promise<void> } | null>(null);
const currentYear = new Date().getFullYear();
const docsLink = 'mailto:support@xingyunpan.com?subject=产品文档咨询';

const loginForm = reactive({
  email: '',
  password: '',
  captcha: '',
});

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const captchaChars = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';

const buildCaptchaSvg = (text: string): string => {
  const noise = Array.from({ length: 18 }, (_, index) => {
    const cx = 10 + index * 10 + ((index * 7) % 13);
    const cy = 8 + ((index * 17) % 34);
    const radius = 0.9 + ((index * 11) % 12) / 10;
    return `<circle cx="${cx}" cy="${cy}" r="${radius}" fill="rgba(59, 130, 246, 0.18)" />`;
  }).join('');

  const letters = text
    .split('')
    .map((char, index) => {
      const x = 18 + index * 26;
      const y = 34 + (index % 2 === 0 ? 3 : -3);
      const rotate = index % 2 === 0 ? -10 : 11;
      return `<text x="${x}" y="${y}" font-size="24" font-family="Arial, sans-serif" font-weight="700" fill="#1f3f75" transform="rotate(${rotate} ${x} ${y})">${char}</text>`;
    })
    .join('');

  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="132" height="52" viewBox="0 0 132 52" role="img" aria-label="验证码">
      <rect width="132" height="52" rx="16" fill="#f8fbff"/>
      <path d="M6 34 C24 12, 38 42, 58 20 S93 14, 126 34" stroke="rgba(37, 99, 235, 0.22)" stroke-width="1.4" fill="none"/>
      <path d="M8 20 C24 40, 46 8, 66 28 S102 42, 124 18" stroke="rgba(14, 116, 144, 0.16)" stroke-width="1.2" fill="none"/>
      ${noise}
      ${letters}
    </svg>
  `;

  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
};

const captchaImage = computed(() => buildCaptchaSvg(captchaSeed.value));
const captchaHint = computed(() => captchaSeed.value.split('').join(' '));

const randomCaptcha = (): string =>
  Array.from({ length: 4 }, () => captchaChars[Math.floor(Math.random() * captchaChars.length)]).join('');

const refreshCaptcha = (): void => {
  captchaSeed.value = randomCaptcha();
  loginForm.captcha = '';
};

const validateEmail = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入邮箱地址'));
    return;
  }
  if (!emailRegex.test(value)) {
    callback(new Error('请输入有效的邮箱地址'));
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

const validateCaptcha = (_rule: unknown, value: string, callback: (error?: Error) => void): void => {
  if (!value) {
    callback(new Error('请输入图形验证码'));
    return;
  }
  if (value.trim().toUpperCase() !== captchaSeed.value) {
    callback(new Error('验证码不正确，请重新输入'));
    return;
  }
  callback();
};

const loginRules: FormRules = {
  email: [{ validator: validateEmail, trigger: 'blur' }],
  password: [{ validator: validatePassword, trigger: 'blur' }],
};

const handleLogin = async (): Promise<void> => {
  if (!loginFormRef.value || loading.value) return;

  submitError.value = '';

  const valid = await loginFormRef.value.validate().catch(() => false);
  if (!valid) {
    return;
  }

  try {
    loading.value = true;

    await userStore.login({
      username: loginForm.email,
      password: loginForm.password,
      ...captchaPayload.value,
    });

    ElMessage.success('登录成功，正在进入授权工作台');
    router.push('/');
  } catch (error: any) {
    submitError.value = error?.message || '登录失败，请稍后重试';
    ElMessage.error(submitError.value);
    await captchaRef.value?.reload();
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  refreshCaptcha();
});
</script>

<style scoped>
.login-page {
  --page-bg: #eef4fb;
  --surface: rgba(255, 255, 255, 0.82);
  --surface-strong: #ffffff;
  --text-primary: #0f172a;
  --text-secondary: #5b6b86;
  --text-tertiary: #8290aa;
  --line-soft: rgba(148, 163, 184, 0.16);
  --line-strong: rgba(191, 219, 254, 0.74);
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

.menu-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 48px;
  height: 48px;
  border: 1px solid rgba(191, 219, 254, 0.85);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 18px 35px rgba(15, 23, 42, 0.04);
  backdrop-filter: blur(16px);
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.menu-button span {
  display: block;
  width: 13px;
  height: 2px;
  border-radius: 999px;
  background: #284579;
}

.menu-button:hover {
  transform: translateY(-1px);
  border-color: rgba(37, 99, 235, 0.4);
  background: rgba(255, 255, 255, 0.92);
}

.dropdown-link {
  color: var(--text-primary);
  text-decoration: none;
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

.visual-card-entry {
  left: 58%;
  top: 56%;
  width: 246px;
  padding: 18px 18px 20px;
  text-decoration: none;
  box-shadow:
    0 22px 48px rgba(41, 98, 234, 0.16),
    0 0 0 1px rgba(255, 255, 255, 0.55) inset;
  background:
    radial-gradient(circle at top left, rgba(96, 165, 250, 0.18), transparent 36%),
    linear-gradient(160deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.94));
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
}

.visual-card-entry:hover {
  transform: translateY(-4px);
  border-color: rgba(96, 165, 250, 0.42);
  box-shadow:
    0 28px 60px rgba(41, 98, 234, 0.2),
    0 0 0 1px rgba(255, 255, 255, 0.68) inset;
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
  font-size: 28px;
  line-height: 1;
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

.visual-entry-kicker {
  display: inline-flex;
  margin-bottom: 10px;
  color: var(--brand-600);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.visual-entry-title {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
  font-size: 22px;
  line-height: 1.15;
  letter-spacing: -0.02em;
}

.visual-entry-text {
  display: block;
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.7;
}

.visual-entry-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 112px;
  min-height: 38px;
  margin-top: 14px;
  padding: 0 14px;
  color: #ffffff;
  font-size: 13px;
  font-weight: 800;
  border-radius: 999px;
  background: linear-gradient(135deg, #2f66ee 0%, #5c95ff 100%);
  box-shadow:
    0 14px 24px rgba(41, 98, 234, 0.26),
    0 0 24px rgba(93, 145, 255, 0.26);
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
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 251, 255, 0.84));
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

.login-form :deep(.el-form-item__label) {
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 8px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 8px;
}

.field-label {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
}

.forgot-link,
.footer-link {
  text-decoration: none;
}

.forgot-link,
.footer-link {
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.forgot-link:hover,
.footer-link:hover {
  color: var(--brand-600);
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

.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 142px;
  gap: 12px;
  width: 100%;
}

.captcha-box {
  display: grid;
  place-items: center;
  gap: 4px;
  padding: 6px;
  border: 1px solid rgba(203, 213, 225, 0.84);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(246, 250, 255, 0.94));
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.captcha-box:hover {
  transform: translateY(-1px);
  border-color: rgba(37, 99, 235, 0.34);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.08);
}

.captcha-image {
  width: 132px;
  height: 52px;
  border-radius: 14px;
}

.captcha-refresh {
  color: var(--text-tertiary);
  font-size: 12px;
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
  background: linear-gradient(135deg, #295fe7 0%, #4485ff 52%, #356ff1 100%);
}

.submit-button:active {
  transform: translateY(1px);
}

.register-spotlight {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin: 6px 0 16px;
  padding: 18px 20px;
  border: 1px solid rgba(96, 165, 250, 0.18);
  border-radius: 24px;
  background:
    radial-gradient(circle at left top, rgba(191, 219, 254, 0.42), transparent 34%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(238, 246, 255, 0.94));
  box-shadow: 0 18px 36px rgba(59, 130, 246, 0.08);
}

.register-copy {
  max-width: 310px;
}

.register-title {
  margin: 0 0 6px;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 800;
}

.register-text {
  margin: 0;
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.75;
}

.register-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 120px;
  min-height: 46px;
  padding: 0 18px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 800;
  text-decoration: none;
  border-radius: 999px;
  background:
    radial-gradient(circle at 30% 20%, rgba(255, 255, 255, 0.34), transparent 32%),
    linear-gradient(135deg, #2d63ea 0%, #5a93ff 100%);
  box-shadow:
    0 18px 32px rgba(41, 98, 234, 0.28),
    0 0 28px rgba(93, 145, 255, 0.28);
  transition: transform 0.2s ease, box-shadow 0.2s ease, filter 0.2s ease;
  position: relative;
  overflow: hidden;
}

.register-link:hover {
  transform: translateY(-1px);
  filter: saturate(1.05);
  box-shadow:
    0 22px 38px rgba(41, 98, 234, 0.32),
    0 0 36px rgba(93, 145, 255, 0.34);
}

.register-link::before {
  content: '';
  position: absolute;
  inset: -40% auto -40% -20%;
  width: 42px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0), rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0));
  transform: rotate(22deg);
  opacity: 0.85;
  animation: registerShine 3.2s ease-in-out infinite;
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
  background:
    linear-gradient(180deg, rgba(252, 254, 255, 0.92), rgba(243, 247, 253, 0.9));
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

.conversion-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 118px;
  min-height: 44px;
  padding: 0 18px;
  color: #1f4fd0;
  font-size: 14px;
  font-weight: 700;
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
  }

  .hero-panel {
    padding-top: 6px;
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

  .visual-card-entry {
    left: 18px;
    top: auto;
    bottom: 18px;
    width: calc(100% - 36px);
  }

  .visual-metrics {
    align-items: flex-start;
    flex-direction: column;
  }

  .trust-grid,
  .captcha-row {
    grid-template-columns: 1fr;
  }

  .form-card {
    padding: 22px;
  }

  .card-title {
    font-size: 26px;
  }

  .register-spotlight,
  .conversion-panel,
  .site-footer,
  .footer-meta {
    align-items: flex-start;
    flex-direction: column;
  }

  .conversion-panel {
    padding: 18px;
  }

  .register-link,
  .conversion-badge {
    width: 100%;
  }
}

@keyframes registerShine {
  0% {
    transform: translateX(-120%) rotate(22deg);
    opacity: 0;
  }
  20% {
    opacity: 0.9;
  }
  55% {
    transform: translateX(360%) rotate(22deg);
    opacity: 0.9;
  }
  100% {
    transform: translateX(420%) rotate(22deg);
    opacity: 0;
  }
}
</style>
