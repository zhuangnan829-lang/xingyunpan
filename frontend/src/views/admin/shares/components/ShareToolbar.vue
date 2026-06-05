<template>
  <div class="share-toolbar">
    <div class="search-box">
      <Search />
      <input v-model="filters.keyword" type="search" placeholder="搜索源文件、分享者或 ID" />
    </div>

    <div class="toolbar-actions">
      <ShareFilterPopover :filters="filters" @reset="$emit('reset-filters')" />
      <button class="toolbar-button" type="button" :disabled="loading" @click="$emit('refresh')">
        <Refresh />
        <span>刷新</span>
      </button>
      <button
        class="toolbar-button danger"
        type="button"
        :disabled="selectedCount === 0 || loading"
        @click="$emit('delete-selected')"
      >
        <Delete />
        <span>删除 {{ selectedCount || '' }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Delete, Refresh, Search } from '@element-plus/icons-vue';
import ShareFilterPopover from './ShareFilterPopover.vue';
import type { ShareFilters } from '../types';

defineProps<{
  filters: ShareFilters;
  loading: boolean;
  selectedCount: number;
}>();

defineEmits<{
  refresh: [];
  'delete-selected': [];
  'reset-filters': [];
}>();
</script>

<style scoped>
.share-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.42);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(16px);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 260px;
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.58);
}

.search-box svg {
  width: 18px;
  height: 18px;
  color: #148dd3;
}

.search-box input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #1f2d3d;
  font: inherit;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.toolbar-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 15px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.56);
  color: #29435c;
  font-weight: 760;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.toolbar-button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.toolbar-button svg {
  width: 17px;
  height: 17px;
}

.toolbar-button.danger {
  color: #d94a5f;
}

@media (max-width: 720px) {
  .share-toolbar,
  .toolbar-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .search-box,
  .toolbar-button {
    width: 100%;
  }
}
</style>
