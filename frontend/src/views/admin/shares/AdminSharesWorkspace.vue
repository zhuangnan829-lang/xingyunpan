<template>
  <section class="admin-shares-page">
    <div class="shares-shell">
      <header class="shares-hero">
        <div>
          <p class="eyebrow">Admin Console</p>
          <h1>分享</h1>
          <p>统一查看分享记录、访问热度、下载转化和失效策略，方便后续接入全局审计与风险处理。</p>
        </div>
      </header>

      <ShareSummaryCards :metrics="metrics" />

      <ShareToolbar
        :filters="filters"
        :loading="isLoading"
        :selected-count="selectedRecords.length"
        @refresh="loadShares"
        @delete-selected="deleteSelected"
        @reset-filters="resetFilters"
      />

      <ShareRecordsTable
        :records="filteredRecords"
        :loading="isLoading"
        @copy="copyLink"
        @delete="deleteRecord"
        @selection-change="updateSelection"
      />

      <footer class="shares-pagination">
        <span>第 {{ pagination.page }} 页 · 共 {{ pagination.total }} 条</span>
        <div class="pager-actions">
          <button type="button" :disabled="!canGoPrevious || isLoading" @click="goPreviousPage">上一页</button>
          <button type="button" :disabled="!canGoNext || isLoading" @click="goNextPage">下一页</button>
        </div>
      </footer>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import ShareRecordsTable from './components/ShareRecordsTable.vue';
import ShareSummaryCards from './components/ShareSummaryCards.vue';
import ShareToolbar from './components/ShareToolbar.vue';
import { useAdminSharesWorkspace } from './useAdminSharesWorkspace';

const {
  copyLink,
  deleteRecord,
  deleteSelected,
  filteredRecords,
  filters,
  canGoNext,
  canGoPrevious,
  goNextPage,
  goPreviousPage,
  isLoading,
  loadShares,
  metrics,
  pagination,
  resetFilters,
  selectedRecords,
  updateSelection,
} = useAdminSharesWorkspace();

onMounted(loadShares);
</script>

<style scoped>
.admin-shares-page {
  position: relative;
  min-height: calc(100vh - 96px);
  overflow: hidden;
  padding: 24px;
  border-radius: 32px;
  background:
    radial-gradient(circle at 10% 8%, rgba(125, 211, 252, 0.28), transparent 28%),
    radial-gradient(circle at 94% 2%, rgba(255, 177, 191, 0.3), transparent 26%),
    radial-gradient(circle at 18% 96%, rgba(216, 180, 254, 0.15), transparent 24%),
    linear-gradient(135deg, #f8fcff 0%, #fff8fb 52%, #f7fbff 100%);
}

.shares-shell {
  display: grid;
  gap: 16px;
}

.shares-hero {
  min-height: 168px;
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 28px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.48));
  box-shadow: 0 24px 58px rgba(67, 93, 121, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(18px);
}

.eyebrow {
  margin: 0 0 10px;
  color: #697789;
  font-size: 12px;
  font-weight: 850;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.shares-hero h1 {
  margin: 0;
  color: #132238;
  font-size: clamp(42px, 5vw, 72px);
  line-height: 1;
  font-weight: 860;
}

.shares-hero p:last-child {
  max-width: 780px;
  margin: 18px 0 0;
  color: #526173;
  font-size: 16px;
  line-height: 1.8;
}

.shares-pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 14px;
  padding: 14px 16px;
  color: #526173;
  font-weight: 720;
}

.pager-actions {
  display: flex;
  gap: 10px;
}

.pager-actions button {
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.58);
  color: #29435c;
  font: inherit;
  cursor: pointer;
}

.pager-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

@media (max-width: 680px) {
  .admin-shares-page {
    padding: 14px;
    border-radius: 24px;
  }

  .shares-hero {
    padding: 20px;
  }

  .shares-pagination {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
