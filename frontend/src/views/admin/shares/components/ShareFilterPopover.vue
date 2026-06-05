<template>
  <el-popover trigger="click" placement="bottom-start" width="280" popper-class="share-filter-popover">
    <template #reference>
      <button class="toolbar-button" type="button">
        <Filter />
        <span>过滤</span>
      </button>
    </template>

    <div class="filter-panel">
      <label>
        <span>状态</span>
        <el-select v-model="filters.status" size="large">
          <el-option label="全部分享" value="all" />
          <el-option label="有效" value="active" />
          <el-option label="已失效" value="expired" />
          <el-option label="带密码" value="protected" />
          <el-option label="全部不可用" value="unavailable" />
          <el-option label="下载达上限" value="download_limit_reached" />
        </el-select>
      </label>

      <label>
        <span>最低下载次数</span>
        <el-input-number v-model="filters.minDownloads" :min="0" :step="5" controls-position="right" />
      </label>

      <el-checkbox v-model="filters.expiringOnly">仅看 3 天内过期</el-checkbox>

      <button class="reset-button" type="button" @click="$emit('reset')">重置筛选</button>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { Filter } from '@element-plus/icons-vue';
import type { ShareFilters } from '../types';

defineProps<{
  filters: ShareFilters;
}>();

defineEmits<{
  reset: [];
}>();
</script>

<style scoped>
.toolbar-button,
.reset-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 42px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.56);
  color: #29435c;
  font-weight: 760;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.toolbar-button {
  padding: 0 15px;
}

.toolbar-button svg {
  width: 17px;
  height: 17px;
}

.filter-panel {
  display: grid;
  gap: 14px;
}

.filter-panel label {
  display: grid;
  gap: 8px;
  color: #526173;
  font-size: 13px;
  font-weight: 760;
}

.reset-button {
  width: 100%;
}
</style>
