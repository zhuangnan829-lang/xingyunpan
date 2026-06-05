<template>
  <div class="oauth-toolbar">
    <div class="search-box">
      <Search />
      <input
        :value="keyword"
        type="search"
        placeholder="搜索应用、Client ID 或授权范围"
        @input="$emit('update:keyword', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <div class="toolbar-actions">
      <select
        :value="status"
        class="status-select"
        @change="$emit('update:status', ($event.target as HTMLSelectElement).value as 'all' | 'enabled' | 'disabled')"
      >
        <option value="all">全部状态</option>
        <option value="enabled">已启用</option>
        <option value="disabled">已停用</option>
      </select>
      <button class="toolbar-button" type="button" @click="$emit('refresh')">
        <Refresh />
        <span>刷新</span>
      </button>
      <button class="toolbar-button primary" type="button" @click="$emit('create')">
        <Plus />
        <span>新建应用</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Plus, Refresh, Search } from '@element-plus/icons-vue';

defineProps<{
  keyword: string;
  status: 'all' | 'enabled' | 'disabled';
}>();

defineEmits<{
  'update:keyword': [value: string];
  'update:status': [value: 'all' | 'enabled' | 'disabled'];
  refresh: [];
  create: [];
}>();
</script>

<style scoped>
.oauth-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.42);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(16px);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 280px;
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.58);
}

.search-box svg {
  width: 18px;
  height: 18px;
  color: #1687e0;
}

.search-box input,
.status-select {
  border: 0;
  outline: 0;
  background: transparent;
  color: #1f2d3d;
  font: inherit;
}

.search-box input {
  width: 100%;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.status-select {
  min-height: 42px;
  padding: 0 36px 0 13px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.56);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
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
  font-weight: 800;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.toolbar-button.primary {
  border: 0;
  background: linear-gradient(135deg, #1f74e8, #18bddf);
  color: #fff;
  box-shadow: 0 12px 24px rgba(33, 135, 219, 0.22);
}

.toolbar-button svg {
  width: 17px;
  height: 17px;
}

@media (max-width: 760px) {
  .oauth-toolbar,
  .toolbar-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .search-box,
  .toolbar-button,
  .status-select {
    width: 100%;
  }
}
</style>
