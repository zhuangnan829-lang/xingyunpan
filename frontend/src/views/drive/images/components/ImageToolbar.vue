<template>
  <section class="image-toolbar">
    <div class="location-pill">
      <Picture class="toolbar-icon" />
      <span>图片</span>
    </div>

    <div class="toolbar-actions">
      <button class="icon-action" type="button" title="刷新" :disabled="loading" @click="$emit('refresh')">
        <RefreshRight class="toolbar-icon" />
      </button>
      <button class="icon-action" type="button" title="更多" @click="$emit('open-menu', $event)">
        <MoreFilled class="toolbar-icon" />
      </button>
      <div class="segmented-control">
        <button
          class="segment-button"
          :class="{ active: cardSize === 'compact' }"
          type="button"
          @click="$emit('update:cardSize', 'compact')"
        >
          紧凑
        </button>
        <button
          class="segment-button"
          :class="{ active: cardSize === 'comfortable' }"
          type="button"
          @click="$emit('update:cardSize', 'comfortable')"
        >
          视图
        </button>
        <button
          class="segment-button"
          :class="{ active: cardSize === 'large' }"
          type="button"
          @click="$emit('update:cardSize', 'large')"
        >
          大图
        </button>
      </div>
      <select class="sort-select" :value="sortMode" @change="handleSortChange">
        <option value="recent">最近更新</option>
        <option value="name">名称排序</option>
        <option value="size">大小排序</option>
      </select>
    </div>
  </section>
</template>

<script setup lang="ts">
import { MoreFilled, Picture, RefreshRight } from '@element-plus/icons-vue';
import type { ImageCardSize, ImageSortMode } from '../useImagesWorkspace';

defineProps<{
  cardSize: ImageCardSize;
  loading: boolean;
  sortMode: ImageSortMode;
}>();

const emit = defineEmits<{
  (e: 'refresh'): void;
  (e: 'open-menu', event: MouseEvent): void;
  (e: 'update:cardSize', value: ImageCardSize): void;
  (e: 'update:sortMode', value: ImageSortMode): void;
}>();

function handleSortChange(event: Event) {
  emit('update:sortMode', (event.target as HTMLSelectElement).value as ImageSortMode);
}
</script>

<style scoped>
.image-toolbar {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) auto;
  gap: 12px;
  align-items: center;
}

.location-pill,
.toolbar-actions,
.segmented-control {
  min-height: 58px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(246, 251, 255, 0.56)),
    radial-gradient(circle at 16% 0%, rgba(121, 205, 255, 0.18), transparent 34%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 16px 34px rgba(86, 126, 171, 0.1);
  backdrop-filter: blur(18px);
}

.location-pill {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  color: #15233d;
  font-size: 15px;
  font-weight: 860;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 7px;
}

.icon-action,
.segment-button,
.sort-select {
  border: 0;
  background: transparent;
  color: #172846;
  font: inherit;
  font-weight: 820;
  cursor: pointer;
}

.icon-action {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 14px;
}

.icon-action:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.icon-action:hover:not(:disabled),
.segment-button.active,
.segment-button:hover,
.sort-select:hover {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(222, 241, 255, 0.78)),
    rgba(255, 255, 255, 0.62);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 22px rgba(66, 142, 220, 0.12);
}

.segmented-control {
  display: inline-flex;
  min-height: 44px;
  overflow: hidden;
  border-radius: 14px;
}

.segment-button {
  min-width: 64px;
  padding: 0 14px;
  border-radius: 12px;
}

.sort-select {
  min-height: 44px;
  padding: 0 14px;
  border-radius: 14px;
  outline: none;
}

.toolbar-icon {
  width: 18px;
  height: 18px;
}

@media (max-width: 920px) {
  .image-toolbar {
    grid-template-columns: 1fr;
  }

  .toolbar-actions {
    flex-wrap: wrap;
  }
}
</style>
