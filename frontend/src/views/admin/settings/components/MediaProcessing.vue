<template>
  <section class="media-center">
    <div class="media-shell">
      <AestheticsBar
        title="媒体资产引擎"
        description="把缩略图、抽帧预览、元数据提取与 Office 审阅，变成一套有温度的数字感官体验。"
        :engines="engineStatuses"
      />

      <div class="feature-grid">
        <FeatureCard
          eyebrow="Card A"
          title="图像视觉引擎"
          description="全格式图像超清预览。让 RAW 原片、封面图与复杂设计素材，在点开之前就已足够动人。"
          :icon="Picture"
          tone="sky"
          :badges="imageBadges"
          :items="imageCapabilities"
        >
          <div class="panel-surface">
            <div class="control-block">
              <div class="control-meta">
                <span>最大原始文件尺寸</span>
                <strong>{{ form.imageMaxSizeGb }} GB</strong>
              </div>
              <input v-model="form.imageMaxSizeGb" class="range-slider" type="range" min="1" max="20" />
              <p class="control-note">大文件也能被温柔预览，不必在细节与速度之间做选择。</p>
            </div>

            <div class="control-block">
              <div class="control-meta">
                <span>缩略图质量</span>
                <strong>{{ form.imageQuality }}%</strong>
              </div>
              <input v-model="form.imageQuality" class="range-slider" type="range" min="40" max="100" />
              <div class="control-scale">
                <span>轻量传输</span>
                <span>保留更多层次</span>
              </div>
            </div>
          </div>
        </FeatureCard>

        <FeatureCard
          eyebrow="Card B"
          title="流媒体直觉"
          description="视频秒开与智能预览。精彩不需要先下载，情绪也不必等待缓冲。"
          :icon="VideoPlay"
          tone="violet"
          :metrics="videoMetrics"
          dark
        >
          <div class="dark-preview">
            <div>
              <p class="preview-label">抽帧预览起始点</p>
              <strong class="preview-time">{{ formattedPreviewTime }}</strong>
            </div>
            <div class="strategy-pill">
              <span>FFmpeg</span>
              <strong>{{ currentVideoStrategy.label }}</strong>
            </div>
          </div>

          <div class="dark-panel">
            <input
              v-model="form.videoPreviewSecond"
              class="range-slider range-slider-dark"
              type="range"
              min="0"
              max="12"
              step="0.25"
            />
            <div class="control-scale control-scale-dark">
              <span>起始瞬间</span>
              <span>更稳定画面</span>
            </div>
          </div>

          <label class="field-card">
            <span class="field-label">视频预加载策略</span>
            <select v-model="form.videoStrategy" class="field-input">
              <option v-for="option in videoStrategies" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
            <span class="field-note">{{ currentVideoStrategy.description }}</span>
          </label>
        </FeatureCard>

        <FeatureCard
          eyebrow="Card C"
          title="数字考古家"
          description="自动文件故事挖掘。把沉默的文件名，变成能被理解、被回忆、被搜索的收藏。"
          :icon="MagicStick"
          tone="amber"
          :items="metadataCards"
          compact
        >
          <div class="panel-surface">
            <div class="toggle-row">
              <div>
                <p class="toggle-title">深度扫描</p>
                <p class="toggle-copy">遇到非常规 EXIF 与嵌入式信息时，继续向文件深处追索。</p>
              </div>

              <button
                type="button"
                class="toggle-button"
                :class="{ 'is-active': form.metadataDeepScan }"
                @click="form.metadataDeepScan = !form.metadataDeepScan"
              >
                <span />
              </button>
            </div>
          </div>

          <div class="annotation-card">
            <p class="annotation-title">智能资源克制</p>
            <p class="annotation-copy">
              由 Golang 驱动的后台调度与内存回收机制，在处理万千文件时依旧保持轻盈。
              复杂留给系统，流畅留给你。
            </p>
          </div>
        </FeatureCard>

        <FeatureCard
          eyebrow="Card D"
          title="文档工作流"
          description="Office 文档免下载审阅。还没保存到本地，内容已经先一步抵达眼前。"
          :icon="Document"
          tone="emerald"
          :items="docFeatures"
          compact
        >
          <div class="panel-surface panel-surface-emerald">
            <div class="status-row">
              <div>
                <p class="toggle-title">LibreOffice 引擎状态</p>
                <p class="toggle-copy">路径检测成功后，即可自动生成文档预览与封面。</p>
              </div>

              <span class="status-pill" :class="{ 'is-ready': libreOfficeReady, 'is-waiting': !libreOfficeReady }">
                <i />
                {{ libreOfficeReady ? '已就绪' : '待配置' }}
              </span>
            </div>

            <label class="field-card field-card-white">
              <span class="field-label">LibreOffice 路径</span>
              <input
                v-model.trim="form.libreofficePath"
                class="field-input"
                type="text"
                placeholder="例如：soffice 或 /usr/bin/soffice"
              />
              <div class="inline-actions">
                <button type="button" class="action-button" @click="checkLibreOffice">检测路径</button>
                <span class="inline-badge">支持 docx / xlsx / pptx / pdf 等常见工作文件</span>
              </div>
            </label>
          </div>
        </FeatureCard>
      </div>

      <UpgradeBanner
        title="激活星云盘 V2 专业版，释放万千媒体资产的终极视觉潜力。"
        description="为视频秒开、RAW 预览、故事化元数据与无感资源调度，提供一整套优雅而强大的用户体验。"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  Camera,
  Document,
  Location,
  MagicStick,
  Microphone,
  Picture,
  VideoCameraFilled,
  VideoPlay,
} from '@element-plus/icons-vue';

import { getMediaSettings, updateMediaSettings, type MediaSettingsPayload } from '@/api/media-settings';
import AestheticsBar from './AestheticsBar.vue';
import FeatureCard from './FeatureCard.vue';
import UpgradeBanner from './UpgradeBanner.vue';
import type { EngineStatus, FeatureCardBadge, FeatureCardItem, FeatureCardMetric } from './media-center.types';

type ImageMode = 'quality' | 'compatibility';
type VideoStrategy = 'smooth' | 'balanced' | 'quality';

type MediaProcessingForm = {
  imageMode: ImageMode;
  imageMaxSizeGb: number;
  imageQuality: number;
  videoPreviewSecond: number;
  videoStrategy: VideoStrategy;
  metadataDeepScan: boolean;
  libreofficePath: string;
};

const defaultFormState = (): MediaProcessingForm => ({
  imageMode: 'quality',
  imageMaxSizeGb: 8,
  imageQuality: 88,
  videoPreviewSecond: 1,
  videoStrategy: 'balanced',
  metadataDeepScan: true,
  libreofficePath: 'soffice',
});

const form = reactive<MediaProcessingForm>(defaultFormState());
const loading = ref(false);
const saving = ref(false);
const lastSavedSnapshot = ref(createSnapshot(defaultFormState()));

const isDirty = computed(() => createSnapshot(form) !== lastSavedSnapshot.value);
const libreOfficeReady = computed(() => form.libreofficePath.trim().length > 0);

const engineStatuses = computed<EngineStatus[]>(() => [
  { name: 'VIPS / LibreRaw', label: '图像预览引擎', ready: true, progress: 100 },
  { name: 'FFmpeg', label: '流媒体抽帧引擎', ready: true, progress: 100 },
  { name: 'Metadata Core', label: '文件故事挖掘', ready: true, progress: form.metadataDeepScan ? 100 : 86 },
  {
    name: 'LibreOffice',
    label: '文档审阅能力',
    ready: libreOfficeReady.value,
    progress: libreOfficeReady.value ? 100 : 58,
  },
]);

const imageBadges = computed<FeatureCardBadge[]>(() => [
  { label: '质量优先', active: form.imageMode === 'quality' },
  { label: '兼容性优先', active: form.imageMode === 'compatibility' },
  { label: '歌曲封面提取', active: true },
]);

const imageCapabilities: FeatureCardItem[] = [
  { title: '4K 与超高分辨率', desc: '大图依然清晰、有层次', icon: Picture },
  { title: 'RAW 原片还原', desc: '保留摄影作品的质感', icon: Camera },
  { title: '音乐封面提取', desc: '让歌单不再只是文件名', icon: Microphone },
];

const videoStrategies = [
  { label: '流畅', value: 'smooth' as VideoStrategy, description: '优先秒开与快速预览，适合移动网络与大规模内容浏览。' },
  { label: '平衡', value: 'balanced' as VideoStrategy, description: '在加载速度与预览细节之间取得自然平衡，适合日常使用。' },
  { label: '高质量', value: 'quality' as VideoStrategy, description: '为更稳定的画面与更完整的细节保留更多缓冲空间。' },
];

const currentVideoStrategy = computed(() => {
  return videoStrategies.find((item) => item.value === form.videoStrategy) ?? videoStrategies[1];
});

const formattedPreviewTime = computed(() => {
  const total = Number(form.videoPreviewSecond);
  const minutes = Math.floor(total / 60);
  const seconds = Math.floor(total % 60);
  const hundredths = Math.round((total - Math.floor(total)) * 100);
  return `00:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}.${String(hundredths).padStart(2, '0')}`;
});

const videoMetrics = computed<FeatureCardMetric[]>(() => {
  const values = {
    smooth: { open: '98%', detail: '72%', stability: '84%' },
    balanced: { open: '92%', detail: '85%', stability: '88%' },
    quality: { open: '84%', detail: '96%', stability: '94%' },
  }[form.videoStrategy];

  return [
    { label: '秒开感知', value: values.open, width: values.open },
    { label: '画面细节', value: values.detail, width: values.detail },
    { label: '播放稳定', value: values.stability, width: values.stability },
  ];
});

const metadataCards: FeatureCardItem[] = [
  { title: '相机参数', desc: '快门、光圈、镜头', icon: Camera },
  { title: '位置记忆', desc: '地点与拍摄足迹', icon: Location },
  { title: '乐曲信息', desc: '专辑、封面、艺术家', icon: Microphone },
  { title: '视频流信息', desc: '编码、时长、分辨率', icon: VideoCameraFilled },
];

const docFeatures: FeatureCardItem[] = [
  { title: '免下载审阅', desc: '打开链接之前，内容已经被预览成可阅读的模样。', icon: Document },
  { title: '工作流友好', desc: '文档、表格、演示稿统一收纳，团队与个人都更顺手。', icon: Picture },
  { title: '轻盈后台', desc: '预览生成在后台安静发生，不打扰你的每一次操作。', icon: MagicStick },
];

function createSnapshot(source: MediaProcessingForm) {
  return JSON.stringify({
    ...source,
    libreofficePath: source.libreofficePath.trim(),
    imageMaxSizeGb: Number(source.imageMaxSizeGb),
    imageQuality: Number(source.imageQuality),
    videoPreviewSecond: Number(source.videoPreviewSecond),
  });
}

function applyFormState(source: MediaProcessingForm) {
  form.imageMode = source.imageMode;
  form.imageMaxSizeGb = source.imageMaxSizeGb;
  form.imageQuality = source.imageQuality;
  form.videoPreviewSecond = source.videoPreviewSecond;
  form.videoStrategy = source.videoStrategy;
  form.metadataDeepScan = source.metadataDeepScan;
  form.libreofficePath = source.libreofficePath;
}

function normalizeForSave(source: MediaProcessingForm): MediaProcessingForm {
  return {
    imageMode: source.imageMode,
    imageMaxSizeGb: Math.min(20, Math.max(1, Number(source.imageMaxSizeGb) || 1)),
    imageQuality: Math.min(100, Math.max(40, Number(source.imageQuality) || 40)),
    videoPreviewSecond: Math.min(12, Math.max(0, Number(source.videoPreviewSecond) || 0)),
    videoStrategy: source.videoStrategy,
    metadataDeepScan: source.metadataDeepScan,
    libreofficePath: source.libreofficePath.trim(),
  };
}

function validateForm(source: MediaProcessingForm) {
  if (!source.libreofficePath) {
    throw new Error('LibreOffice 路径不能为空');
  }
}

function toFormState(payload: MediaSettingsPayload): MediaProcessingForm {
  return {
    imageMode: payload.image_mode ?? 'quality',
    imageMaxSizeGb: payload.image_max_size_gb ?? 8,
    imageQuality: payload.image_quality ?? 88,
    videoPreviewSecond: payload.video_preview_second ?? 1,
    videoStrategy: payload.video_strategy ?? 'balanced',
    metadataDeepScan: Boolean(payload.metadata_deep_scan),
    libreofficePath: payload.libreoffice_path ?? 'soffice',
  };
}

function toPayload(source: MediaProcessingForm): MediaSettingsPayload {
  return {
    image_mode: source.imageMode,
    image_max_size_gb: source.imageMaxSizeGb,
    image_quality: source.imageQuality,
    video_preview_second: source.videoPreviewSecond,
    video_strategy: source.videoStrategy,
    metadata_deep_scan: source.metadataDeepScan,
    libreoffice_path: source.libreofficePath.trim(),
  };
}

async function reload() {
  loading.value = true;
  try {
    const data = await getMediaSettings();
    const next = toFormState(data);
    applyFormState(next);
    lastSavedSnapshot.value = createSnapshot(next);
    ElMessage.success('媒体处理配置已重新加载');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载媒体处理配置失败');
  } finally {
    loading.value = false;
  }
}

function reset() {
  const next = defaultFormState();
  applyFormState(next);
  ElMessage.success('媒体处理配置已恢复默认值，记得保存');
}

async function save() {
  saving.value = true;
  try {
    const next = normalizeForSave(form);
    validateForm(next);
    const data = await updateMediaSettings(toPayload(next));
    const saved = toFormState(data);
    applyFormState(saved);
    lastSavedSnapshot.value = createSnapshot(saved);
    ElMessage.success('媒体处理配置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存媒体处理配置失败');
  } finally {
    saving.value = false;
  }
}

function checkLibreOffice() {
  if (libreOfficeReady.value) {
    ElMessage.success('LibreOffice 路径看起来可用');
    return;
  }

  ElMessage.warning('请输入可用的 LibreOffice 路径');
}

defineExpose({ isDirty, loading, saving, reload, reset, save });

onMounted(async () => {
  await reload();
});
</script>

<style scoped>
.media-center {
  min-height: 100%;
  padding: 8px 0 0;
}

.media-shell {
  display: grid;
  gap: 24px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.panel-surface,
.dark-panel,
.field-card,
.annotation-card {
  display: grid;
  gap: 14px;
  padding: 18px;
  border-radius: 24px;
}

.panel-surface {
  background: #f8fafc;
}

.panel-surface-emerald {
  background: linear-gradient(180deg, #ecfdf5, #ffffff);
}

.control-block {
  display: grid;
  gap: 10px;
}

.control-meta,
.status-row,
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.control-meta span,
.field-label,
.preview-label,
.toggle-title {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.control-meta strong,
.preview-time {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.preview-time {
  display: block;
  margin-top: 10px;
  font-size: 34px;
  letter-spacing: -0.04em;
  color: #fff;
}

.control-note,
.control-scale span,
.field-note,
.toggle-copy,
.annotation-copy,
.inline-badge,
.preview-label {
  color: #64748b;
  font-size: 12px;
  line-height: 1.8;
}

.control-scale,
.inline-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.range-slider {
  width: 100%;
  height: 8px;
  appearance: none;
  border-radius: 999px;
  background: #dbe4ee;
  cursor: pointer;
}

.range-slider::-webkit-slider-thumb {
  width: 18px;
  height: 18px;
  appearance: none;
  border: 4px solid #fff;
  border-radius: 50%;
  background: #0f172a;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.16);
}

.range-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border: 4px solid #fff;
  border-radius: 50%;
  background: #0f172a;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.16);
}

.dark-preview,
.strategy-pill {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.strategy-pill {
  display: grid;
  width: fit-content;
  padding: 12px 14px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.08);
}

.strategy-pill span {
  color: rgba(255, 255, 255, 0.56);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.strategy-pill strong {
  margin-top: 4px;
  color: #fff;
  font-size: 14px;
}

.dark-panel {
  background: rgba(255, 255, 255, 0.06);
}

.range-slider-dark {
  background: rgba(255, 255, 255, 0.18);
}

.range-slider-dark::-webkit-slider-thumb,
.range-slider-dark::-moz-range-thumb {
  background: #fff;
}

.control-scale-dark span,
.field-card .field-note {
  color: rgba(255, 255, 255, 0.72);
}

.field-card {
  background: rgba(255, 255, 255, 0.08);
}

.field-card-white {
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(226, 232, 240, 0.92);
}

.field-card-white .field-note,
.field-card-white .field-label {
  color: #0f172a;
}

.field-input {
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid rgba(203, 213, 225, 0.9);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.96);
  color: #334155;
  outline: none;
}

.field-card:not(.field-card-white) .field-input {
  background: rgba(255, 255, 255, 0.94);
}

.toggle-button {
  position: relative;
  width: 56px;
  height: 32px;
  border: none;
  border-radius: 999px;
  background: #cbd5e1;
  cursor: pointer;
  transition: background 0.2s ease;
}

.toggle-button span {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.18);
  transition: transform 0.2s ease;
}

.toggle-button.is-active {
  background: #0f172a;
}

.toggle-button.is-active span {
  transform: translateX(24px);
}

.annotation-card {
  border: 1px solid rgba(253, 230, 138, 0.9);
  background: rgba(255, 251, 235, 0.95);
}

.annotation-title {
  margin: 0;
  color: #92400e;
  font-size: 14px;
  font-weight: 700;
}

.annotation-copy {
  color: #78350f;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 700;
}

.status-pill i {
  width: 9px;
  height: 9px;
  border-radius: 50%;
}

.status-pill.is-ready {
  background: #d1fae5;
  color: #047857;
}

.status-pill.is-ready i {
  background: #10b981;
}

.status-pill.is-waiting {
  background: #fee2e2;
  color: #b91c1c;
}

.status-pill.is-waiting i {
  background: #ef4444;
}

.action-button {
  min-height: 40px;
  padding: 0 16px;
  border: none;
  border-radius: 999px;
  background: #0f172a;
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.inline-badge {
  display: inline-flex;
  align-items: center;
  padding: 10px 14px;
  border-radius: 999px;
  background: #f8fafc;
  color: #475569;
}

@media (max-width: 1080px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .dark-preview,
  .strategy-pill,
  .status-row,
  .toggle-row {
    grid-template-columns: 1fr;
  }

  .dark-preview,
  .status-row,
  .toggle-row {
    display: grid;
  }

  .preview-time {
    font-size: 28px;
  }
}
</style>
