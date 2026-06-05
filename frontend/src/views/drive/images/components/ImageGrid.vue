<template>
  <section class="image-board" :class="`size-${cardSize}`" @click.self="$emit('clear-selection')">
    <div v-if="loading" class="empty-state">正在整理图片...</div>
    <div v-else-if="images.length === 0" class="empty-state">
      <Picture class="empty-icon" />
      <strong>还没有图片</strong>
      <span>上传图片后，这里会自动聚合展示。</span>
    </div>
    <div v-else class="image-grid" @click.self="$emit('clear-selection')">
      <article
        v-for="image in images"
        :key="image.id"
        class="image-card"
        :class="{ selected: selectedIds.includes(image.id) }"
        @click="handleCardClick(image)"
        @contextmenu.prevent.stop="handleCardContextMenu($event, image)"
        @dblclick="handleCardDoubleClick(image)"
      >
        <div class="image-card-head">
          <span class="type-thumb" aria-hidden="true">
            <img v-if="thumbnailUrl(image)" :src="thumbnailUrl(image)" alt="" loading="lazy" @error="markImageFailed" />
            <Picture class="type-fallback" />
          </span>
          <strong :title="image.name">{{ image.name }}</strong>
          <button class="select-button" type="button" @click.stop="$emit('toggle-select', image)">
            <Check v-if="selectedIds.includes(image.id)" />
          </button>
        </div>

        <div class="image-thumb">
          <img v-if="thumbnailUrl(image)" :src="thumbnailUrl(image)" :alt="image.name" loading="lazy" @error="markImageFailed" />
          <div class="thumb-fallback">
            <Picture />
            <span>{{ image.name }}</span>
          </div>
        </div>

        <div class="image-meta">
          <span>{{ formatFileSize(image.size || 0) }}</span>
          <span>{{ formatTimestamp(image.updated_at) }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onUnmounted, watch } from 'vue';
import { Check, Picture } from '@element-plus/icons-vue';
import type { FileItem } from '@/types/file';
import { formatFileSize, formatTimestamp } from '@/utils/format';
import type { ImageCardSize } from '../useImagesWorkspace';

const props = defineProps<{
  cardSize: ImageCardSize;
  ensureThumbnail: (file: FileItem) => void;
  images: FileItem[];
  loading: boolean;
  selectedIds: number[];
  thumbnailUrl: (file: FileItem) => string;
}>();

const emit = defineEmits<{
  (e: 'clear-selection'): void;
  (e: 'image-context', event: MouseEvent, image: FileItem): void;
  (e: 'preview', image: FileItem): void;
  (e: 'toggle-select', image: FileItem): void;
}>();

const clickTimers = new Map<number, number>();
const CLICK_DELAY = 220;

function clearCardClickTimer(fileId: number) {
  const timer = clickTimers.get(fileId);
  if (!timer) return;
  window.clearTimeout(timer);
  clickTimers.delete(fileId);
}

function handleCardClick(image: FileItem) {
  clearCardClickTimer(image.id);
  const timer = window.setTimeout(() => {
    clickTimers.delete(image.id);
    emit('toggle-select', image);
  }, CLICK_DELAY);
  clickTimers.set(image.id, timer);
}

function handleCardDoubleClick(image: FileItem) {
  clearCardClickTimer(image.id);
  emit('preview', image);
}

function handleCardContextMenu(event: MouseEvent, image: FileItem) {
  clearCardClickTimer(image.id);
  emit('image-context', event, image);
}

function markImageFailed(event: Event) {
  const image = event.currentTarget as HTMLImageElement;
  image.dataset.failed = 'true';
}

watch(
  () => props.images.map((image) => `${image.id}:${image.updated_at}`).join('|'),
  () => {
    props.images.forEach((image) => props.ensureThumbnail(image));
  },
  { immediate: true },
);

onUnmounted(() => {
  clickTimers.forEach((timer) => window.clearTimeout(timer));
  clickTimers.clear();
});
</script>

<style scoped>
.image-board {
  min-height: calc(100vh - 260px);
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 24px;
  background:
    radial-gradient(circle at 96% 0%, rgba(146, 223, 255, 0.22), transparent 30%),
    radial-gradient(circle at 4% 102%, rgba(255, 203, 224, 0.18), transparent 28%),
    rgba(255, 255, 255, 0.62);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 28px 70px rgba(88, 128, 176, 0.12);
  backdrop-filter: blur(22px);
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--image-card-size), 1fr));
  gap: 18px;
  align-items: start;
}

.size-compact {
  --image-card-size: 176px;
}

.size-comfortable {
  --image-card-size: 244px;
}

.size-large {
  --image-card-size: 328px;
}

.image-card {
  display: grid;
  gap: 14px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(246, 251, 255, 0.48)),
    radial-gradient(circle at 10% 0%, rgba(116, 220, 255, 0.14), transparent 36%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.94),
    0 18px 38px rgba(81, 124, 172, 0.1);
  cursor: pointer;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.image-card:hover,
.image-card.selected {
  transform: translateY(-2px);
  border-color: rgba(84, 163, 255, 0.5);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 24px 48px rgba(62, 137, 220, 0.16);
}

.image-card.selected {
  outline: 2px solid rgba(45, 112, 255, 0.42);
  outline-offset: 2px;
}

.image-card-head {
  display: grid;
  grid-template-columns: 22px minmax(0, 1fr) 28px;
  align-items: center;
  gap: 10px;
  min-height: 32px;
}

.image-card-head strong {
  overflow: hidden;
  color: #12213b;
  font-size: 14px;
  font-weight: 820;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.type-thumb {
  position: relative;
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 7px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.88), rgba(226, 242, 255, 0.72)),
    radial-gradient(circle at 30% 20%, rgba(84, 163, 255, 0.24), transparent 46%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 6px 14px rgba(64, 119, 180, 0.12);
}

.type-thumb img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.type-thumb img[data-failed='true'] {
  display: none;
}

.type-fallback {
  width: 15px;
  height: 15px;
  color: #2d70ff;
}

.select-button {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: #fff;
  cursor: pointer;
  opacity: 0;
  transition:
    opacity 0.16s ease,
    background 0.16s ease,
    border-color 0.16s ease;
}

.image-card:hover .select-button,
.image-card.selected .select-button {
  opacity: 1;
}

.image-card.selected .select-button {
  border-color: rgba(37, 99, 235, 0.72);
  background: #1d70da;
}

.image-thumb {
  position: relative;
  display: grid;
  place-items: center;
  aspect-ratio: 1 / 0.78;
  overflow: hidden;
  border-radius: 14px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(245, 250, 255, 0.76)),
    radial-gradient(circle at 50% 20%, rgba(255, 197, 220, 0.16), transparent 36%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.image-thumb img {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-thumb img[data-failed='true'] {
  display: none;
}

.thumb-fallback {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 10px;
  padding: 16px;
  color: #5d718d;
  text-align: center;
}

.thumb-fallback svg {
  width: 48px;
  height: 48px;
  color: #2d70ff;
}

.thumb-fallback span {
  max-width: 100%;
  overflow: hidden;
  font-size: 13px;
  font-weight: 760;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-meta {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  color: #63738d;
  font-size: 12px;
  font-weight: 720;
}

.empty-state {
  display: grid;
  min-height: 360px;
  place-items: center;
  align-content: center;
  gap: 10px;
  color: #61708a;
  text-align: center;
}

.empty-state strong {
  color: #13213a;
  font-size: 18px;
}

.empty-icon {
  width: 52px;
  height: 52px;
  color: #2f7df5;
}

@media (max-width: 720px) {
  .image-board {
    padding: 14px;
  }

  .image-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}
</style>
