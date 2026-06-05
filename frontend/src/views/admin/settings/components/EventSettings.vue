<template>
  <section class="event-settings">
    <div class="event-shell">
      <header class="event-hero">
        <div class="hero-copy">
          <p class="eyebrow">Event Registry</p>
          <div class="title-row">
            <h2>事件</h2>
            <span class="pro-badge">Pro</span>
          </div>
          <p>
            配置哪些事件应该被记录。部分事件可用于额外功能，例如文件活动、登录活动、Webhook 和审计追踪。
          </p>
        </div>

        <div class="hero-metrics">
          <article>
            <span>事件总数</span>
            <strong>{{ totalCount }}</strong>
          </article>
          <article>
            <span>已启用</span>
            <strong>{{ selectedCount }}</strong>
          </article>
          <article>
            <span>启用分类</span>
            <strong>{{ enabledCategoryCount }}</strong>
          </article>
        </div>
      </header>

      <section class="event-board">
        <div class="board-toolbar">
          <div>
            <p class="eyebrow">记录范围</p>
            <strong>{{ boardSummary }}</strong>
          </div>
          <div class="toolbar-actions">
            <button type="button" class="soft-button" :disabled="panelLocked" @click="setAllEvents(true)">启用全部</button>
            <button type="button" class="soft-button" :disabled="panelLocked" @click="setAllEvents(false)">禁用全部</button>
          </div>
        </div>

        <EventCategorySection
          v-for="category in eventCategories"
          :key="category.key"
          :category="category"
          :model="form"
          :all-checked="isCategoryEnabled(category.key)"
          :selected-count="categorySelectedCount(category.key)"
          :disabled="panelLocked"
          @toggle-category="toggleCategory"
          @toggle-event="toggleEvent"
        />
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  getEventSettings,
  resetEventSettings,
  toggleAllEventSettings,
  toggleEventCategory,
  toggleEventSetting,
  updateEventSettings,
  type EventSettingsPayload,
} from '@/api/event-settings';
import EventCategorySection from './EventCategorySection.vue';
import { createDefaultEventState, eventCategories } from './event-settings.data';

type EventState = Record<string, boolean>;

function cloneState(source: EventState): EventState {
  return { ...createDefaultEventState(), ...source };
}

function createSnapshot(source: EventState) {
  const ordered: EventState = {};
  for (const category of eventCategories) {
    for (const item of category.items) {
      ordered[item.key] = Boolean(source[item.key]);
    }
  }
  return JSON.stringify(ordered);
}

const form = reactive<EventState>(createDefaultEventState());
const loading = ref(false);
const saving = ref(false);
const lastSavedSnapshot = ref(createSnapshot(form));

const totalCount = computed(() => eventCategories.reduce((sum, category) => sum + category.items.length, 0));
const selectedCount = computed(() => Object.values(form).filter(Boolean).length);
const enabledCategoryCount = computed(() => eventCategories.filter((category) => categorySelectedCount(category.key) > 0).length);
const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);
const panelLocked = computed(() => loading.value || saving.value);
const boardSummary = computed(() => {
  if (selectedCount.value === 0) return '当前没有启用事件记录';
  if (selectedCount.value === totalCount.value) return '当前已启用所有事件记录';
  return `当前启用了 ${selectedCount.value} 个事件记录`;
});

function categoryByKey(categoryKey: string) {
  return eventCategories.find((category) => category.key === categoryKey);
}

function categorySelectedCount(categoryKey: string) {
  const category = categoryByKey(categoryKey);
  if (!category) return 0;
  return category.items.filter((item) => form[item.key]).length;
}

function isCategoryEnabled(categoryKey: string) {
  const category = categoryByKey(categoryKey);
  return Boolean(category?.items.length && category.items.every((item) => form[item.key]));
}

function applyPayload(payload: EventSettingsPayload) {
  applyState(payload.events || {});
  lastSavedSnapshot.value = createSnapshot(form);
}

async function toggleCategory(categoryKey: string, checked: boolean) {
  saving.value = true;
  try {
    const payload = await toggleEventCategory(categoryKey, checked);
    applyPayload(payload);
    ElMessage.success(checked ? '事件分类已启用' : '事件分类已禁用');
  } finally {
    saving.value = false;
  }
}

async function toggleEvent(eventKey: string, checked: boolean) {
  saving.value = true;
  try {
    const payload = await toggleEventSetting(eventKey, checked);
    applyPayload(payload);
    ElMessage.success(checked ? '事件已启用' : '事件已禁用');
  } finally {
    saving.value = false;
  }
}

async function setAllEvents(checked: boolean) {
  saving.value = true;
  try {
    const payload = await toggleAllEventSettings(checked);
    applyPayload(payload);
    ElMessage.success(checked ? '已启用全部事件' : '已禁用全部事件');
  } finally {
    saving.value = false;
  }
}

function applyState(source: EventState) {
  const next = cloneState(source);
  for (const key of Object.keys(createDefaultEventState())) {
    form[key] = Boolean(next[key]);
  }
}

async function reload() {
  loading.value = true;
  try {
    const payload = await getEventSettings();
    applyPayload(payload);
    ElMessage.success('事件设置已重新加载');
  } finally {
    loading.value = false;
  }
}

async function reset() {
  saving.value = true;
  try {
    const payload = await resetEventSettings();
    applyPayload(payload);
    ElMessage.success('事件设置已恢复默认值');
  } finally {
    saving.value = false;
  }
}

async function save() {
  saving.value = true;
  try {
    const payload = await updateEventSettings({ events: cloneState(form) });
    applyPayload(payload);
    lastSavedSnapshot.value = createSnapshot(form);
    ElMessage.success('事件设置已保存到后端');
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  void reload();
});

defineExpose({ isDirty, loading, saving, reload, reset, save });
</script>

<style scoped>
.event-settings {
  color: #18293d;
}

.event-shell {
  display: grid;
  gap: 22px;
}

.event-hero,
.event-board {
  border: 1px solid rgba(255, 255, 255, 0.82);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(247, 252, 255, 0.6));
  box-shadow:
    0 22px 54px rgba(75, 133, 180, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.84);
  backdrop-filter: blur(18px);
}

.event-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(340px, 440px);
  gap: 28px;
  align-items: end;
  padding: 30px 34px;
  border-radius: 28px;
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.88) 0%, rgba(232, 247, 255, 0.72) 52%, rgba(255, 237, 241, 0.68) 100%);
}

.hero-copy {
  display: grid;
  gap: 12px;
}

.eyebrow,
h2,
p,
strong,
span {
  margin: 0;
}

.eyebrow {
  color: #2488dc;
  font-size: 12px;
  font-weight: 820;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

h2 {
  color: #142339;
  font-size: clamp(2.4rem, 4vw, 3.6rem);
  line-height: 1.05;
  font-weight: 840;
  letter-spacing: 0;
}

.pro-badge {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, #2c92ef 0%, #49c4df 100%);
  color: #fff;
  font-size: 12px;
  font-weight: 820;
  box-shadow: 0 12px 22px rgba(44, 146, 239, 0.22);
}

.hero-copy p {
  max-width: 840px;
  color: #66768a;
  line-height: 1.8;
  font-size: 15px;
  font-weight: 620;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.hero-metrics article {
  display: grid;
  gap: 7px;
  min-height: 96px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 12px 26px rgba(66, 139, 190, 0.08);
}

.hero-metrics span {
  color: #6f8094;
  font-size: 13px;
  font-weight: 700;
}

.hero-metrics strong {
  color: #16283d;
  font-size: 28px;
  line-height: 1;
}

.event-board {
  display: grid;
  padding: 28px 40px 36px;
  border-radius: 28px;
}

.board-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(207, 222, 232, 0.72);
}

.board-toolbar strong {
  display: block;
  margin-top: 7px;
  color: #142339;
  font-size: 18px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.soft-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 16px;
  border: 1px solid rgba(199, 222, 238, 0.86);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.64);
  color: #37657f;
  font-weight: 720;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    color 0.18s ease;
}

.soft-button:hover {
  color: #2388dc;
  transform: translateY(-1px);
  box-shadow:
    0 12px 22px rgba(66, 181, 230, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.soft-button:disabled {
  cursor: not-allowed;
  opacity: 0.58;
  transform: none;
}

@media (max-width: 1180px) {
  .event-hero {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 780px) {
  .event-hero,
  .event-board {
    padding: 22px;
    border-radius: 24px;
  }

  .hero-metrics {
    grid-template-columns: 1fr;
  }

  .board-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
