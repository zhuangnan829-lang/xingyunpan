<template>
  <section class="settings-page">
    <div class="page-shell">
      <header class="hero-card">
        <div class="hero-wireframe" aria-hidden="true">
          <span class="wire-orbit orbit-a" />
          <span class="wire-orbit orbit-b" />
          <span class="wire-orbit orbit-c" />
          <span class="wire-node node-a" />
          <span class="wire-node node-b" />
          <span class="wire-node node-c" />
        </div>
        <div class="hero-copy">
          <p class="eyebrow">Admin Settings</p>
          <h1>参数设置</h1>
          <p class="hero-text">
            统一管理站点信息、会话安全与服务参数
          </p>
        </div>

        <div class="hero-summary">
          <article class="summary-card">
            <span>当前面板</span>
            <strong>{{ activeTabMeta.label }}</strong>
          </article>
          <article class="summary-card">
            <span>可配置模块</span>
            <strong>{{ tabs.length }}</strong>
          </article>
        </div>
      </header>

      <section class="panel-card">
        <div class="tabs-scroll">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            type="button"
            class="tab-button"
            :class="{ 'is-active': activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            <el-icon class="tab-icon">
              <component :is="tab.icon" />
            </el-icon>
            <span>{{ tab.label }}</span>
          </button>
        </div>

        <div class="tab-panel">
          <component :is="activeTabMeta.component" :ref="setPanelRef" />
        </div>
      </section>
    </div>

    <footer class="action-bar">
      <div class="action-copy">
        <strong>{{ actionTitle }}</strong>
        <span>{{ actionDescription }}</span>
      </div>

      <div class="action-group">
        <button class="secondary-button" type="button" :disabled="panelBusy" @click="handleReload">
          <el-icon><RefreshRight /></el-icon>
          <span>重新加载</span>
        </button>

        <button class="secondary-button" type="button" :disabled="panelBusy" @click="handleReset">
          恢复默认
        </button>

        <button class="primary-button" type="button" :disabled="panelBusy" @click="handleSave">
          {{ saveButtonText }}
        </button>
      </div>
    </footer>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import {
  Brush,
  Camera,
  Connection,
  Crop,
  Message,
  Monitor,
  RefreshRight,
  Setting,
  SwitchButton,
  Tickets,
} from '@element-plus/icons-vue';

import AppearanceSettings from './components/AppearanceSettings.vue';
import Captcha from './components/Captcha.vue';
import EventSettings from './components/EventSettings.vue';
import MailSettings from './components/MailSettings.vue';
import MediaProcessing from './components/MediaProcessing.vue';
import QueueSettings from './components/QueueSettings.vue';
import ServerSettings from './components/ServerSettings.vue';
import SiteInfo from './components/SiteInfo.vue';
import UserSessions from './components/UserSessions.vue';

type SettingsTabKey =
  | 'site-info'
  | 'user-sessions'
  | 'captcha'
  | 'media-processing'
  | 'mail'
  | 'queue'
  | 'appearance'
  | 'events'
  | 'server';

type SettingsPanelExpose = {
  isDirty?: boolean;
  loading?: boolean;
  saving?: boolean;
  save?: () => Promise<void> | void;
  reload?: () => Promise<void> | void;
  reset?: () => Promise<void> | void;
};

type SettingsTabItem = {
  key: SettingsTabKey;
  label: string;
  icon: unknown;
  component: unknown;
};

const tabs: SettingsTabItem[] = [
  { key: 'site-info', label: '站点信息', icon: Setting, component: SiteInfo },
  { key: 'user-sessions', label: '会话安全', icon: SwitchButton, component: UserSessions },
  { key: 'captcha', label: '验证码', icon: Camera, component: Captcha },
  { key: 'media-processing', label: '媒体处理', icon: Crop, component: MediaProcessing },
  { key: 'mail', label: '邮件', icon: Message, component: MailSettings },
  { key: 'queue', label: '队列', icon: Tickets, component: QueueSettings },
  { key: 'appearance', label: '外观', icon: Brush, component: AppearanceSettings },
  { key: 'events', label: '事件', icon: Connection, component: EventSettings },
  { key: 'server', label: '服务器', icon: Monitor, component: ServerSettings },
];

const route = useRoute();
const router = useRouter();
const activeTab = ref<SettingsTabKey>('site-info');
const activePanelRef = ref<SettingsPanelExpose | null>(null);

const activeTabMeta = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0]);
const panelBusy = computed(() => Boolean(activePanelRef.value?.loading || activePanelRef.value?.saving));
const saveButtonText = computed(() => (activePanelRef.value?.saving ? '保存中...' : '保存更改'));

const actionTitle = computed(() => {
  if (activePanelRef.value?.saving) return `正在保存${activeTabMeta.value.label}`;
  if (activePanelRef.value?.loading) return `正在加载${activeTabMeta.value.label}`;
  if (activePanelRef.value?.isDirty) return `${activeTabMeta.value.label}有未保存修改`;
  return `${activeTabMeta.value.label}已与当前配置同步`;
});

const actionDescription = computed(() => `当前底部操作条将作用于“${activeTabMeta.value.label}”面板。`);

function setPanelRef(instance: unknown) {
  activePanelRef.value = (instance as SettingsPanelExpose | null) ?? null;
}

function resolveTab(value: unknown): SettingsTabKey {
  const matched = tabs.find((tab) => tab.key === value);
  return (matched?.key ?? 'site-info') as SettingsTabKey;
}

async function handleReload() {
  if (!activePanelRef.value?.reload) {
    ElMessage.info(`“${activeTabMeta.value.label}”暂未接入重新加载逻辑`);
    return;
  }
  await activePanelRef.value.reload();
}

async function handleReset() {
  if (!activePanelRef.value?.reset) {
    ElMessage.info(`“${activeTabMeta.value.label}”暂未接入恢复默认逻辑`);
    return;
  }
  await activePanelRef.value.reset();
}

async function handleSave() {
  if (!activePanelRef.value?.save) {
    ElMessage.info(`“${activeTabMeta.value.label}”暂未接入保存逻辑`);
    return;
  }
  await activePanelRef.value.save();
}

watch(
  () => route.query.tab,
  (value) => {
    const next = resolveTab(value);
    if (activeTab.value !== next) {
      activeTab.value = next;
    }
  },
  { immediate: true },
);

watch(activeTab, async (value) => {
  if (route.query.tab === value) return;
  await router.replace({
    query: {
      ...route.query,
      tab: value,
    },
  });
});
</script>

<style scoped>
.settings-page {
  position: relative;
  min-height: calc(100vh - 96px);
  padding: 18px 20px 108px;
  background:
    radial-gradient(circle at 7% 8%, rgba(117, 196, 235, 0.18), transparent 26%),
    radial-gradient(circle at 34% 92%, rgba(255, 175, 169, 0.14), transparent 28%),
    linear-gradient(180deg, #f3f8fe 0%, #eef5fb 100%);
  overflow: hidden;
}

.page-shell {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 18px;
}

.hero-card,
.panel-card,
.action-bar,
.summary-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: rgba(255, 255, 255, 0.72);
  box-shadow:
    0 16px 38px rgba(122, 176, 218, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px);
}

.hero-card {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 20px;
  min-height: 188px;
  padding: 30px 36px 28px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background:
    radial-gradient(circle at 12% 0%, rgba(255, 255, 255, 0.86), transparent 30%),
    radial-gradient(circle at 80% 6%, rgba(139, 210, 255, 0.16), transparent 32%),
    linear-gradient(142deg, rgba(255, 255, 255, 0.9) 0%, rgba(245, 251, 255, 0.78) 54%, rgba(255, 239, 238, 0.62) 100%);
}

.hero-copy,
.hero-summary {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 12px;
}

.hero-summary {
  align-content: start;
}

.summary-card {
  display: grid;
  gap: 7px;
  padding: 16px 18px;
  border: 1px solid rgba(222, 233, 244, 0.72);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(248, 252, 255, 0.58));
}

.hero-wireframe {
  position: absolute;
  top: 8px;
  right: 28px;
  width: min(38vw, 430px);
  height: 210px;
  opacity: 0.24;
  pointer-events: none;
}

.hero-wireframe::before,
.hero-wireframe::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(31deg, transparent 48%, rgba(79, 152, 212, 0.22) 49%, rgba(79, 152, 212, 0.22) 50%, transparent 51%),
    linear-gradient(148deg, transparent 44%, rgba(185, 133, 129, 0.18) 45%, rgba(185, 133, 129, 0.18) 46%, transparent 47%),
    linear-gradient(90deg, transparent 49%, rgba(79, 152, 212, 0.13) 50%, transparent 51%);
  clip-path: polygon(4% 76%, 26% 16%, 62% 6%, 94% 52%, 70% 90%, 28% 84%);
}

.hero-wireframe::after {
  inset: 18px 34px 10px 50px;
  opacity: 0.7;
  transform: rotate(-7deg);
}

.wire-orbit {
  position: absolute;
  border: 1px solid rgba(91, 157, 210, 0.24);
  border-radius: 50%;
  transform: rotate(-14deg);
}

.orbit-a {
  inset: 18px 74px 34px 24px;
}

.orbit-b {
  inset: 52px 26px 52px 104px;
  transform: rotate(18deg);
}

.orbit-c {
  inset: 70px 146px 70px 44px;
  transform: rotate(38deg);
}

.wire-node {
  position: absolute;
  width: 9px;
  height: 9px;
  border-radius: 999px;
  background: #5aaee4;
  box-shadow: 0 0 18px rgba(90, 174, 228, 0.28);
}

.node-a { top: 40px; right: 160px; }
.node-b { right: 42px; bottom: 82px; }
.node-c { left: 86px; bottom: 42px; }

.eyebrow {
  margin: 0;
  color: #3a93df;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

h1,
p,
strong,
span {
  margin: 0;
}

h1,
strong {
  color: #193e5d;
}

h1 {
  font-size: clamp(2.75rem, 3.8vw, 4rem);
  line-height: 1.05;
  font-weight: 820;
  letter-spacing: 0;
}

.hero-text,
.summary-card span,
.action-copy span {
  color: #72879b;
}

.hero-text {
  max-width: 840px;
  line-height: 1.75;
  font-size: 15px;
  font-weight: 650;
}

.summary-card strong {
  font-size: 22px;
}

.panel-card {
  position: relative;
  overflow: visible;
  padding: 18px 28px 28px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(249, 253, 255, 0.66));
}

.tabs-scroll {
  position: sticky;
  top: 12px;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 18px;
  min-height: 72px;
  padding: 10px 12px 16px;
  margin: 0 0 26px;
  overflow-x: auto;
  border-bottom: 1px solid rgba(216, 228, 238, 0.76);
  border-radius: 24px 24px 0 0;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(247, 252, 255, 0.42));
  backdrop-filter: blur(16px) saturate(145%);
  scrollbar-width: thin;
}

.tab-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-height: 54px;
  padding: 0 18px;
  border: 1px solid transparent;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.28);
  color: #70859b;
  cursor: pointer;
  white-space: nowrap;
  font-size: 17px;
  font-weight: 780;
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.tab-button:hover {
  color: #263341;
  background: rgba(255, 255, 255, 0.56);
  transform: translateY(-1px);
}

.tab-button.is-active {
  color: #2387dc;
  font-weight: 900;
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 18% 18%, rgba(108, 214, 244, 0.26), transparent 42%),
    linear-gradient(90deg, rgba(219, 246, 255, 0.94) 0%, rgba(255, 211, 220, 0.9) 100%);
  box-shadow:
    0 12px 22px rgba(90, 181, 225, 0.2),
    inset 0 0 0 1px rgba(255, 255, 255, 0.62),
    inset 0 2px 8px rgba(255, 255, 255, 0.52);
}

.tab-button.is-active::after {
  display: none;
}

.tab-icon {
  color: #b98581;
  font-size: 18px;
}

.tab-panel {
  min-height: 520px;
}

.action-bar {
  position: fixed;
  right: 24px;
  bottom: 24px;
  left: 344px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px);
}

.action-copy {
  display: grid;
  gap: 6px;
}

.action-group {
  display: flex;
  gap: 10px;
}

.primary-button,
.secondary-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 44px;
  border: none;
  border-radius: 999px;
  cursor: pointer;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.secondary-button {
  padding: 0 18px;
  background: rgba(255, 255, 255, 0.64);
  color: #5d748c;
  border: 1px solid rgba(214, 229, 240, 0.95);
}

.primary-button {
  padding: 0 22px;
  background: linear-gradient(135deg, #57b8ea 0%, #5ad0dc 58%, #f0a2a4 100%);
  color: #fff;
  box-shadow: 0 14px 28px rgba(90, 181, 225, 0.22);
}

.primary-button:hover,
.secondary-button:hover {
  transform: translateY(-1px);
}

.primary-button:disabled,
.secondary-button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  transform: none;
}

@media (max-width: 1100px) {
  .hero-card {
    grid-template-columns: 1fr;
  }

  .action-bar {
    left: 24px;
    flex-direction: column;
    align-items: stretch;
  }

  .action-group {
    flex-wrap: wrap;
  }
}

@media (max-width: 720px) {
  .settings-page {
    padding: 16px 16px 160px;
  }

  .hero-card,
  .panel-card {
    padding: 18px;
  }

  h1 {
    font-size: 40px;
  }

  .hero-wireframe {
    width: 340px;
    right: -130px;
    opacity: 0.28;
  }

  .action-group {
    display: grid;
  }
}
</style>
