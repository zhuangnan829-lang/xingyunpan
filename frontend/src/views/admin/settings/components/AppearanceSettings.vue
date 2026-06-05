<template>
  <section class="appearance-settings">
    <div class="appearance-shell">
      <div class="settings-workspace">
        <div class="main-column">
          <section class="content-card theme-options-card">
            <div class="section-head">
              <div>
                <h3>主题选项</h3>
                <p>按品牌主色、辅助色与暗色模式色值统一控制登录页和后台界面的视觉参数。</p>
              </div>
            </div>

            <div class="theme-form-stack">
              <article
                v-for="theme in themeOptions"
                :key="theme.id"
                class="theme-form-card"
                :class="{ active: selectedThemeId === theme.id }"
              >
                <div class="theme-form-head">
                  <div>
                    <strong>主题方案 {{ theme.id }}</strong>
                    <span>{{ theme.isDefault ? '当前默认主题' : '可切换主题' }}</span>
                  </div>
                  <button type="button" class="default-pill" @click="selectTheme(theme.id)">
                    <span class="checkbox" :class="{ active: theme.isDefault }" />
                    <span>设为默认</span>
                  </button>
                </div>

                <label class="setting-row">
                  <span class="setting-label">主色调</span>
                  <div class="color-field" :style="{ '--glow': theme.light.primary }">
                    <input v-model="theme.light.primary" class="hex-input" type="text" maxlength="7" />
                    <input v-model="theme.light.primary" class="picker-input" type="color" aria-label="选择主色调" />
                  </div>
                  <small>用于主按钮、活跃页签和后台重点操作，建议保持高识别度蓝色。</small>
                </label>

                <label class="setting-row">
                  <span class="setting-label">次色调</span>
                  <div class="color-field" :style="{ '--glow': theme.light.secondary }">
                    <input v-model="theme.light.secondary" class="hex-input" type="text" maxlength="7" />
                    <input v-model="theme.light.secondary" class="picker-input" type="color" aria-label="选择次色调" />
                  </div>
                  <small>用于辅助强调、渐变过渡和轻量装饰色，当前示例为 #9c27b0。</small>
                </label>

                <label class="setting-row">
                  <span class="setting-label">暗色主色调</span>
                  <div class="color-field" :style="{ '--glow': theme.dark.primary }">
                    <input v-model="theme.dark.primary" class="hex-input" type="text" maxlength="7" />
                    <input v-model="theme.dark.primary" class="picker-input" type="color" aria-label="选择暗色主色调" />
                  </div>
                  <small>暗色模式中的按钮、链接与焦点态颜色，应比亮色主色更柔和。</small>
                </label>

                <label class="setting-row">
                  <span class="setting-label">暗色次色调</span>
                  <div class="color-field" :style="{ '--glow': theme.dark.secondary }">
                    <input v-model="theme.dark.secondary" class="hex-input" type="text" maxlength="7" />
                    <input v-model="theme.dark.secondary" class="picker-input" type="color" aria-label="选择暗色次色调" />
                  </div>
                  <small>暗色模式下的辅助高亮色，用于徽标、进度条和浮层边缘光。</small>
                </label>

                <div class="row-actions">
                  <button type="button" class="icon-button" @click="openEditor(theme.id)">编辑完整 JSON</button>
                </div>
              </article>
            </div>

            <button class="add-button" type="button" @click="addThemeOption">
              <span>+</span>
              <strong>添加主题选项</strong>
            </button>
          </section>

          <section class="content-card feature-card">
            <div class="section-head">
              <div>
                <h3>自定义 UI</h3>
                <p>你可以打开更接近设计稿的霓虹细节、玻璃拟态、动态过渡和数据可视化增强。</p>
              </div>
            </div>

            <div class="feature-list">
              <label v-for="feature in featureOptions" :key="feature.key" class="feature-item">
                <input v-model="feature.enabled" type="checkbox" />
                <span class="feature-box"><span class="feature-core" /></span>
                <span class="feature-copy">
                  <strong>{{ feature.label }}</strong>
                  <small>{{ feature.description }}</small>
                </span>
              </label>
            </div>
          </section>
        </div>

        <aside class="preview-column" :style="previewVars">
          <section class="login-preview-card">
            <div class="preview-title">
              <span>Login Page Preview</span>
              <strong>登录页预览</strong>
            </div>
            <div class="phone-shell">
              <div class="phone-status">
                <span />
                <i />
              </div>
              <div class="phone-screen">
                <div class="brand-row">
                  <div class="brand-mark"><span class="brand-wave wave-a" /><span class="brand-wave wave-b" /></div>
                  <strong>星云盘</strong>
                </div>
                <h4>Welcome back</h4>
                <div class="preview-input">Email address</div>
                <div class="preview-input muted">Password</div>
                <button class="preview-button" type="button">Sign In</button>
                <div class="powered-by">Xingyunpan V2</div>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>

    <el-dialog v-model="editorVisible" class="theme-dialog" width="1180px" append-to-body destroy-on-close>
      <template #header>
        <div class="dialog-head">
          <div>
            <p class="dialog-kicker">编辑主题选项</p>
            <h3>编辑主题选项</h3>
          </div>
        </div>
      </template>

      <div v-if="editorTheme" class="dialog-grid">
        <section class="dialog-editor">
          <div class="dialog-section-title">主题配置</div>
            <div class="editor-shell">
              <div class="editor-gutter">
                <span v-for="line in 14" :key="line">{{ line }}</span>
              </div>
              <div class="editor-code">
                <div v-for="(line, index) in editorLines" :key="index" class="code-line">
                  <span v-for="(token, tokenIndex) in line" :key="`${index}-${tokenIndex}`" class="token" :class="`token-${token.type}`">
                    {{ token.text }}
                  </span>
                </div>
              </div>
          </div>
          <p class="editor-help">完整的可配置项请参考你的主题设计规范。</p>
        </section>

        <section class="dialog-preview">
          <div class="dialog-section-title">主题预览</div>
          <div class="preview-stack">
            <div class="mini-preview light" :style="dialogPreviewVars">
              <p>亮色主题</p>
              <div class="mini-card">
                <h4>预览标题</h4>
                <button type="button">预览标题</button>
                <div class="mini-input">输入字段</div>
                <span class="mini-chip">主色调</span>
                <span class="mini-badge">10</span>
              </div>
            </div>
            <div class="mini-preview dark" :style="dialogPreviewVars">
              <p>暗色主题</p>
              <div class="mini-card dark-card">
                <h4>预览标题</h4>
                <button type="button">预览标题</button>
                <div class="mini-input dark-input">输入字段</div>
                <span class="mini-chip dark-chip">主色调</span>
                <span class="mini-badge">10</span>
              </div>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="dialog-actions">
          <button class="dialog-button cancel" type="button" @click="editorVisible = false">取消</button>
          <button class="dialog-button confirm" type="button" @click="editorVisible = false">确定</button>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  getAppearanceSettings,
  updateAppearanceSettings,
  type AppearanceFeatureOptionPayload,
  type AppearanceThemeOptionPayload,
} from '@/api/appearance-settings';

type ThemePalette = { primary: string; secondary: string };
type ThemeOption = { id: number; isDefault: boolean; light: ThemePalette; dark: ThemePalette };
type FeatureOption = { key: string; label: string; description: string; enabled: boolean };
const defaultThemes: ThemeOption[] = [
  { id: 1, isDefault: true, light: { primary: '#1976d2', secondary: '#9c27b0' }, dark: { primary: '#90caf9', secondary: '#ce93d8' } },
  { id: 2, isDefault: false, light: { primary: '#3f51b5', secondary: '#f50057' }, dark: { primary: '#9fa8da', secondary: '#ff4081' } },
];
const defaultFeatures: FeatureOption[] = [
  { key: 'neon', label: '启用霓虹细节 (Neon Accents)', description: '为色块、按钮和边框加入更亮的蓝色辉光。', enabled: true },
  { key: 'glass', label: '玻璃拟态元素 (Glassmorphism Elements)', description: '增强纯白界面中的轻玻璃感和层次感。', enabled: true },
  { key: 'motion', label: '动态过渡动画 (Animated Transitions)', description: '增加卡片切换与控件悬停时的细腻反馈。', enabled: true },
  { key: 'data-viz', label: '数据可视化优化', description: '强化预览卡片中的进度条与发光效果。', enabled: true },
];

const themeOptions = ref<ThemeOption[]>(clone(defaultThemes));
const featureOptions = ref<FeatureOption[]>(clone(defaultFeatures));
const selectedThemeId = ref<number>(1);
const editorVisible = ref(false);
const editorThemeId = ref<number | null>(null);
const loading = ref(false);
const saving = ref(false);

const snapshot = computed(() => JSON.stringify({ themeOptions: themeOptions.value, featureOptions: featureOptions.value, selectedThemeId: selectedThemeId.value }));
const pristineSnapshot = ref(snapshot.value);
const isDirty = computed(() => snapshot.value !== pristineSnapshot.value);
const activeTheme = computed(() => themeOptions.value.find((item) => item.id === selectedThemeId.value) ?? themeOptions.value[0]);
const editorTheme = computed(() => themeOptions.value.find((item) => item.id === editorThemeId.value) ?? null);
const previewVars = computed(() => ({
  '--accent-primary': activeTheme.value.light.primary,
  '--accent-secondary': activeTheme.value.light.secondary,
  '--accent-dark': activeTheme.value.dark.primary,
  '--accent-pink': activeTheme.value.dark.secondary,
  '--neon-opacity': featureOptions.value.find((item) => item.key === 'neon')?.enabled ? '1' : '.35',
}));
const dialogPreviewVars = computed(() => ({
  '--dialog-primary': editorTheme.value?.light.primary ?? '#1976d2',
  '--dialog-secondary': editorTheme.value?.light.secondary ?? '#9c27b0',
  '--dialog-dark': editorTheme.value?.dark.primary ?? '#90caf9',
  '--dialog-pink': editorTheme.value?.dark.secondary ?? '#ce93d8',
}));
const editorLines = computed(() => {
  const lightPrimary = editorTheme.value?.light.primary ?? '#1976d2';
  const lightSecondary = editorTheme.value?.light.secondary ?? '#9c27b0';
  const darkPrimary = editorTheme.value?.dark.primary ?? '#90caf9';
  const darkSecondary = editorTheme.value?.dark.secondary ?? '#ce93d8';

  return [
    [{ text: '{', type: 'brace' }],
    [{ text: '  ', type: 'space' }, { text: '"light"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{', type: 'brace' }],
    [{ text: '    ', type: 'space' }, { text: '"palette"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{', type: 'brace' }],
    [{ text: '      ', type: 'space' }, { text: '"primary"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{ ', type: 'brace' }, { text: '"main"', type: 'key' }, { text: ': ', type: 'colon' }, { text: `"${lightPrimary}"`, type: 'string' }, { text: ' }', type: 'brace' }],
    [{ text: '      ', type: 'space' }, { text: '"secondary"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{ ', type: 'brace' }, { text: '"main"', type: 'key' }, { text: ': ', type: 'colon' }, { text: `"${lightSecondary}"`, type: 'string' }, { text: ' }', type: 'brace' }],
    [{ text: '    ', type: 'space' }, { text: '}', type: 'brace' }],
    [{ text: '  ', type: 'space' }, { text: '},', type: 'brace' }],
    [{ text: '  ', type: 'space' }, { text: '"dark"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{', type: 'brace' }],
    [{ text: '    ', type: 'space' }, { text: '"palette"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{', type: 'brace' }],
    [{ text: '      ', type: 'space' }, { text: '"primary"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{ ', type: 'brace' }, { text: '"main"', type: 'key' }, { text: ': ', type: 'colon' }, { text: `"${darkPrimary}"`, type: 'string' }, { text: ' }', type: 'brace' }],
    [{ text: '      ', type: 'space' }, { text: '"secondary"', type: 'key' }, { text: ': ', type: 'colon' }, { text: '{ ', type: 'brace' }, { text: '"main"', type: 'key' }, { text: ': ', type: 'colon' }, { text: `"${darkSecondary}"`, type: 'string' }, { text: ' }', type: 'brace' }],
    [{ text: '    ', type: 'space' }, { text: '}', type: 'brace' }],
    [{ text: '  ', type: 'space' }, { text: '}', type: 'brace' }],
    [{ text: '}', type: 'brace' }],
  ];
});

function clone<T>(value: T): T { return JSON.parse(JSON.stringify(value)); }
function applyPayload(payload: { theme_options: AppearanceThemeOptionPayload[]; feature_options: AppearanceFeatureOptionPayload[]; selected_theme_id: number }) {
  themeOptions.value = payload.theme_options.map((item) => ({
    id: item.id,
    isDefault: item.is_default,
    light: { primary: item.light.primary, secondary: item.light.secondary },
    dark: { primary: item.dark.primary, secondary: item.dark.secondary },
  }));
  featureOptions.value = payload.feature_options.map((item) => ({
    key: item.key,
    label: item.label,
    description: item.description,
    enabled: item.enabled,
  }));
  selectedThemeId.value = payload.selected_theme_id;
}
function toPayload() {
  return {
    theme_options: themeOptions.value.map((item) => ({
      id: item.id,
      is_default: item.isDefault,
      light: { primary: item.light.primary, secondary: item.light.secondary },
      dark: { primary: item.dark.primary, secondary: item.dark.secondary },
    })),
    feature_options: featureOptions.value.map((item) => ({
      key: item.key,
      label: item.label,
      description: item.description,
      enabled: item.enabled,
    })),
    selected_theme_id: selectedThemeId.value,
  };
}
function selectTheme(id: number) {
  selectedThemeId.value = id;
  themeOptions.value = themeOptions.value.map((item) => ({ ...item, isDefault: item.id === id }));
}
function openEditor(id: number) { editorThemeId.value = id; editorVisible.value = true; }
function addThemeOption() {
  const nextId = Math.max(...themeOptions.value.map((item) => item.id)) + 1;
  themeOptions.value.push({ id: nextId, isDefault: false, light: { primary: '#1e88ff', secondary: '#7c4dff' }, dark: { primary: '#74b6ff', secondary: '#b388ff' } });
  selectTheme(nextId);
  openEditor(nextId);
  ElMessage.success('已添加新的主题选项');
}
async function reload() {
  loading.value = true;
  try {
    const data = await getAppearanceSettings();
    applyPayload(data);
    pristineSnapshot.value = snapshot.value;
    ElMessage.success('外观设置已重新加载');
  } finally { loading.value = false; }
}
async function reset() {
  themeOptions.value = clone(defaultThemes);
  featureOptions.value = clone(defaultFeatures);
  selectedThemeId.value = 1;
  ElMessage.success('外观设置已恢复默认值');
}
async function save() {
  saving.value = true;
  try {
    const data = await updateAppearanceSettings(toPayload());
    applyPayload(data);
    pristineSnapshot.value = snapshot.value;
    ElMessage.success('外观设置已保存');
  } finally { saving.value = false; }
}

defineExpose({ isDirty, loading, saving, reload, reset, save });

onMounted(async () => {
  await reload();
});
</script>

<style scoped>
.appearance-settings {
  color: #16324d;
}

.appearance-settings h3,
.appearance-settings h4,
.appearance-settings p,
.appearance-settings strong,
.appearance-settings span,
.appearance-settings small {
  margin: 0;
}

.settings-workspace {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 330px);
  gap: 22px;
}

.main-column {
  display: grid;
  gap: 22px;
}

.content-card,
.login-preview-card {
  border: 1px solid rgba(224, 231, 240, 0.96);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 14px 34px rgba(122, 176, 218, 0.1);
  backdrop-filter: blur(18px);
}

.content-card {
  padding: 22px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.section-head h3 {
  font-size: 24px;
  line-height: 1.1;
  letter-spacing: 0;
}

.section-head p {
  margin-top: 8px;
  color: #7388a0;
  font-size: 14px;
  line-height: 1.75;
}

.theme-form-stack {
  display: grid;
  gap: 18px;
}

.theme-form-card {
  display: grid;
  gap: 15px;
  padding: 18px;
  border: 1px solid rgba(222, 232, 242, 0.96);
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(250, 253, 255, 0.9)),
    radial-gradient(circle at 94% 12%, rgba(25, 118, 210, 0.08), transparent 26%);
}

.theme-form-card.active {
  border-color: rgba(25, 118, 210, 0.38);
  box-shadow: 0 0 0 4px rgba(25, 118, 210, 0.08), 0 18px 34px rgba(25, 118, 210, 0.12);
}

.theme-form-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.theme-form-head > div {
  display: grid;
  gap: 6px;
}

.theme-form-head strong {
  color: #16324d;
  font-size: 16px;
}

.theme-form-head span,
.setting-row small {
  color: #8a9aae;
}

.setting-row {
  display: grid;
  gap: 8px;
}

.setting-label {
  color: #172033;
  font-size: 15px;
  font-weight: 800;
}

.color-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 48px;
  align-items: center;
  gap: 12px;
  min-height: 52px;
  padding: 6px 8px 6px 18px;
  border: 1px solid rgba(214, 226, 239, 0.96);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: inset 0 1px 2px rgba(100, 116, 139, 0.08), 0 0 26px color-mix(in srgb, var(--glow) 18%, transparent);
}

.hex-input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #263f5b;
  font: 800 15px/1.2 Consolas, 'Courier New', monospace;
  letter-spacing: 0;
}

.picker-input {
  width: 36px;
  height: 36px;
  padding: 0;
  border: 4px solid #fff;
  border-radius: 50%;
  overflow: hidden;
  background: transparent;
  box-shadow: 0 0 0 1px rgba(203, 213, 225, 0.9), 0 0 22px var(--glow);
  cursor: pointer;
}

.picker-input::-webkit-color-swatch-wrapper {
  padding: 0;
}

.picker-input::-webkit-color-swatch {
  border: 0;
  border-radius: 50%;
}

.default-pill,
.icon-button,
.add-button,
.preview-button,
.dialog-button {
  border: none;
  cursor: pointer;
}

.default-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 40px;
  padding: 0 14px;
  border: 1px solid rgba(205, 226, 249, 0.95);
  border-radius: 999px;
  background: #fff;
  color: #2877c9;
  font-weight: 800;
}

.checkbox {
  width: 18px;
  height: 18px;
  border-radius: 6px;
  background: #fff;
  box-shadow: inset 0 0 0 2px rgba(152, 177, 203, 0.58);
}

.checkbox.active {
  background: linear-gradient(135deg, #1976d2, #53b9ff);
  box-shadow: 0 0 16px rgba(44, 136, 255, 0.28);
}

.row-actions {
  display: flex;
  justify-content: flex-start;
}

.icon-button {
  min-height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  background: rgba(25, 118, 210, 0.08);
  color: #1976d2;
  font-size: 12px;
  font-weight: 800;
}

.add-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 46px;
  margin-top: 20px;
  padding: 0 18px;
  border-radius: 999px;
  background: linear-gradient(135deg, #57b8ea, #f0a2a4);
  color: #fff;
  box-shadow: 0 16px 28px rgba(25, 118, 210, 0.22);
}

.add-button span {
  font-size: 24px;
  line-height: 1;
}

.feature-list {
  display: grid;
  gap: 12px;
}

.feature-item {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr);
  gap: 14px;
  align-items: center;
  padding: 15px 16px;
  border-radius: 18px;
  background: #f9fbff;
  box-shadow: inset 0 0 0 1px rgba(221, 232, 243, 0.96);
}

.feature-item input {
  position: absolute;
  opacity: 0;
}

.feature-box {
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border-radius: 8px;
  background: #fff;
  box-shadow: inset 0 0 0 1px rgba(162, 194, 225, 0.9);
}

.feature-core {
  width: 14px;
  height: 14px;
  border-radius: 4px;
  background: linear-gradient(135deg, #38a5ff, #6ad7ff);
  box-shadow: 0 0 16px rgba(59, 161, 255, 0.5);
}

.feature-copy {
  display: grid;
  gap: 4px;
}

.feature-copy strong {
  font-size: 15px;
}

.feature-copy small {
  color: #758aa2;
  line-height: 1.7;
}

.preview-column {
  display: grid;
  align-items: start;
}

.login-preview-card {
  position: sticky;
  top: 18px;
  align-self: start;
  overflow: hidden;
  margin-top: 70px;
  padding: 18px;
}

.login-preview-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg, rgba(25, 118, 210, 0.7), rgba(110, 209, 255, 0.12), rgba(156, 39, 176, 0.28));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

.preview-title {
  display: grid;
  gap: 4px;
  margin-bottom: 18px;
}

.preview-title span {
  color: #1976d2;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.preview-title strong {
  font-size: 17px;
}

.phone-shell {
  width: min(100%, 252px);
  margin-left: auto;
  padding: 12px;
  border-radius: 30px;
  background: #172033;
  box-shadow: 0 26px 46px rgba(23, 32, 51, 0.18);
}

.phone-status {
  display: flex;
  justify-content: space-between;
  padding: 2px 12px 10px;
}

.phone-status span,
.phone-status i {
  display: block;
  height: 4px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.45);
}

.phone-status span {
  width: 48px;
}

.phone-status i {
  width: 18px;
}

.phone-screen {
  display: grid;
  gap: 12px;
  min-height: 390px;
  padding: 24px 18px 18px;
  border-radius: 22px;
  background:
    radial-gradient(circle at 50% 0%, color-mix(in srgb, var(--accent-primary) 18%, transparent), transparent 42%),
    linear-gradient(180deg, #fff, #f8fbff);
}

.brand-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-mark {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, #69e1ff, var(--accent-primary) 60%, #205fd6);
  overflow: hidden;
  box-shadow: 0 0 calc(24px * var(--neon-opacity)) rgba(45, 145, 255, 0.28);
}

.brand-wave {
  position: absolute;
  border-radius: 50%;
}

.wave-a {
  inset: 7px 5px 11px 14px;
  background: #fff;
  opacity: 0.92;
}

.wave-b {
  left: 4px;
  right: 9px;
  bottom: 4px;
  height: 16px;
  border-radius: 20px 20px 8px 8px;
  background: rgba(138, 226, 255, 0.72);
}

.phone-screen h4 {
  margin-top: 10px;
  color: #142238;
  font-size: 25px;
  line-height: 1.05;
  letter-spacing: 0;
}

.preview-input {
  display: flex;
  align-items: center;
  min-height: 46px;
  padding: 0 16px;
  border: 1px solid rgba(212, 225, 239, 0.96);
  border-radius: 14px;
  background: #fff;
  color: #8fa2b6;
}

.preview-input.muted {
  color: #b3bfca;
}

.preview-button {
  min-height: 48px;
  border-radius: 14px;
  background: linear-gradient(135deg, #57b8ea, var(--accent-primary));
  color: #fff;
  font-size: 15px;
  font-weight: 900;
  box-shadow: 0 0 calc(26px * var(--neon-opacity)) rgba(44, 136, 255, 0.24);
}

.powered-by {
  align-self: end;
  padding-top: 26px;
  color: #b1b9c5;
  text-align: center;
  font-size: 12px;
  font-weight: 800;
}
.dialog-head h3,.dialog-kicker,.dialog-section-title,.editor-help{margin:0}.dialog-kicker{color:#3a8ff6;font-size:12px;font-weight:800;letter-spacing:.14em;text-transform:uppercase}.dialog-head h3{margin-top:6px;font-size:2rem;letter-spacing:-.04em}.dialog-grid{display:grid;grid-template-columns:minmax(0,1.08fr) minmax(380px,.92fr);gap:28px}.dialog-section-title{margin-bottom:14px;color:#1f3d5a;font-size:1.35rem;font-weight:700}.editor-shell{display:grid;grid-template-columns:60px minmax(0,1fr);overflow:hidden;border:1px solid rgba(223,231,240,.96);border-radius:18px;background:#fff;min-height:602px}.editor-gutter{display:grid;align-content:start;gap:8px;padding:18px 12px;background:#f8fbff;color:#5d7ea5;text-align:right;font-weight:700}.editor-code{padding:18px 20px;background:#fff}.code-line{display:flex;min-height:38px;align-items:center;white-space:pre-wrap;font-family:Consolas,'Courier New',monospace;font-size:15px;line-height:1.7}.token-space{color:transparent}.token-brace{color:#2563eb;font-weight:700}.token-key{color:#c026d3}.token-colon{color:#64748b}.token-string{display:inline-flex;align-items:center;min-height:28px;padding:0 8px;border-radius:8px;background:rgba(239,68,68,.08);color:#dc2626;box-shadow:inset 0 0 0 1px rgba(248,113,113,.14)}.editor-help{margin-top:12px;color:#7288a1}.preview-stack{display:grid;gap:18px}.mini-preview{padding:18px;border-radius:18px;background:#f7f9fc}.mini-preview p{margin-bottom:12px;color:#516a86;font-weight:700}.mini-card{position:relative;display:grid;gap:16px;padding:20px;border-radius:18px;background:#fff;box-shadow:inset 0 0 0 1px rgba(226,233,243,.96)}.mini-card h4{font-size:1.7rem}.mini-card button,.mini-chip{display:inline-flex;align-items:center;justify-content:center;min-height:40px;padding:0 16px;border:none;border-radius:14px;background:linear-gradient(135deg,var(--dialog-primary),var(--dialog-dark));color:#fff;font-weight:700;justify-self:start}.mini-card button{width:100%;justify-content:flex-start;padding:0 18px;font-size:15px}.mini-input{display:flex;align-items:center;min-height:50px;padding:0 16px;border-radius:14px;border:1px solid rgba(220,229,239,.96);color:#7e93a9;font-size:15px}.mini-chip{min-height:42px;padding:0 18px;font-size:15px}.mini-badge{position:absolute;left:94px;bottom:20px;display:inline-flex;align-items:center;justify-content:center;min-width:32px;height:28px;padding:0 10px;border-radius:999px;background:var(--dialog-secondary);color:#fff;font-size:12px;font-weight:700;box-shadow:0 0 20px rgba(156,39,176,.2)}.dark{background:#292929}.dark-card{background:#262626;box-shadow:none}.dark-card h4,.dark-card p{color:#fff}.dark-input{border-color:#4b5563;color:#b6c0cf;background:#2a2a2a}.dark-chip,.dark-card button{background:linear-gradient(135deg,#32485f,var(--dialog-dark))}.dialog-actions{display:flex;justify-content:flex-end;gap:12px}.dialog-button{min-height:42px;padding:0 18px;border-radius:14px}.dialog-button.cancel{background:#eef3f8;color:#6b7f95}.dialog-button.confirm{background:#258ff6;color:#fff}
:deep(.theme-dialog .el-dialog){border-radius:24px}:deep(.theme-dialog .el-dialog__body){padding-top:8px}@keyframes flow{0%{filter:hue-rotate(0deg)}100%{filter:hue-rotate(18deg)}}@media (max-width:1180px){.settings-workspace,.dialog-grid{grid-template-columns:1fr}.login-preview-card{position:relative;top:auto;margin-top:0}.phone-shell{margin-left:0}}@media (max-width:720px){.content-card,.login-preview-card{padding:18px}.theme-form-head{align-items:flex-start;flex-direction:column}.color-field{grid-template-columns:minmax(0,1fr) 44px}.phone-screen{min-height:420px}}
</style>
