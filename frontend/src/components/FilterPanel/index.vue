<template>
  <a-popover
    v-model:popup-visible="visible"
    trigger="click"
    position="bottom"
    :content-style="{ padding: 0 }"
  >
    <template #content>
      <div class="filter-panel">
        <div class="filter-header">
          <span class="filter-title">筛选条件</span>
          <a-button type="text" size="small" @click="handleReset">
            重置
          </a-button>
        </div>

        <!-- File Type Filter -->
        <div class="filter-section">
          <div class="filter-label">文件类型</div>
          <a-select
            v-model="localFilters.fileType"
            :style="{ width: '100%' }"
            placeholder="选择文件类型"
            allow-clear
            @change="handleFilterChange"
          >
            <a-option value="">全部</a-option>
            <a-option value="image">图片</a-option>
            <a-option value="video">视频</a-option>
            <a-option value="audio">音频</a-option>
            <a-option value="document">文档</a-option>
            <a-option value="archive">压缩包</a-option>
            <a-option value="other">其他</a-option>
          </a-select>
        </div>

        <a-divider :margin="0" />

        <!-- Date Range Filter -->
        <div class="filter-section">
          <div class="filter-label">修改日期</div>
          <a-range-picker
            v-model="dateRange"
            :style="{ width: '100%' }"
            format="YYYY-MM-DD"
            :shortcuts="dateShortcuts"
            @change="handleDateRangeChange"
          />
        </div>

        <a-divider :margin="0" />

        <!-- Size Range Filter -->
        <div class="filter-section">
          <div class="filter-label">文件大小</div>
          <div class="size-range-inputs">
            <a-input-number
              v-model="localFilters.sizeMin"
              :min="0"
              :precision="0"
              placeholder="最小值"
              :style="{ width: '100%' }"
              @change="handleFilterChange"
            >
              <template #suffix>MB</template>
            </a-input-number>
            <span class="range-separator">-</span>
            <a-input-number
              v-model="localFilters.sizeMax"
              :min="0"
              :precision="0"
              placeholder="最大值"
              :style="{ width: '100%' }"
              @change="handleFilterChange"
            >
              <template #suffix>MB</template>
            </a-input-number>
          </div>
        </div>
      </div>
    </template>

    <a-button type="outline">
      <template #icon>
        <icon-filter />
      </template>
      筛选
      <span v-if="activeFilterCount > 0" class="filter-badge">
        ({{ activeFilterCount }})
      </span>
    </a-button>
  </a-popover>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { IconFilter } from '@arco-design/web-vue/es/icon';
import { useSearchStore, type SearchFilters } from '@/stores/search';
import { useFileStore } from '@/stores/file';

// Stores
const searchStore = useSearchStore();
const fileStore = useFileStore();

// Local state
const visible = ref(false);
const localFilters = ref<SearchFilters>({
  fileType: null,
  sizeMin: null,
  sizeMax: null,
  dateFrom: null,
  dateTo: null,
});
const dateRange = ref<[string, string] | undefined>(undefined);

// Date shortcuts for quick selection
const dateShortcuts = [
  {
    label: '今天',
    value: () => {
      const today = new Date();
      return [today, today];
    },
  },
  {
    label: '最近 7 天',
    value: () => {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 7);
      return [start, end];
    },
  },
  {
    label: '最近 30 天',
    value: () => {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 30);
      return [start, end];
    },
  },
  {
    label: '最近 90 天',
    value: () => {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 90);
      return [start, end];
    },
  },
];

// Computed

/**
 * Count active filter conditions
 * Requirements: 2.3-2.5, 8.2
 */
const activeFilterCount = computed(() => {
  let count = 0;
  if (localFilters.value.fileType) count++;
  if (localFilters.value.sizeMin !== null || localFilters.value.sizeMax !== null) count++;
  if (localFilters.value.dateFrom || localFilters.value.dateTo) count++;
  return count;
});

// Initialize local filters from store on mount
onMounted(() => {
  localFilters.value = { ...searchStore.filters };
  if (searchStore.filters.dateFrom && searchStore.filters.dateTo) {
    dateRange.value = [searchStore.filters.dateFrom, searchStore.filters.dateTo];
  }
});

// Methods

/**
 * Handle date range change
 * Requirements: 2.4
 */
function handleDateRangeChange(dateStrings: [string, string] | undefined) {
  if (dateStrings && dateStrings.length === 2) {
    localFilters.value.dateFrom = dateStrings[0];
    localFilters.value.dateTo = dateStrings[1];
  } else {
    localFilters.value.dateFrom = null;
    localFilters.value.dateTo = null;
  }
  handleFilterChange();
}

/**
 * Handle filter change - apply to search store
 * Requirements: 2.3-2.5
 */
function handleFilterChange() {
  // Convert MB to bytes for API
  const filters: SearchFilters = {
    fileType: localFilters.value.fileType || null,
    sizeMin: localFilters.value.sizeMin ? localFilters.value.sizeMin * 1024 * 1024 : null,
    sizeMax: localFilters.value.sizeMax ? localFilters.value.sizeMax * 1024 * 1024 : null,
    dateFrom: localFilters.value.dateFrom,
    dateTo: localFilters.value.dateTo,
  };
  
  searchStore.setFilters(filters);
  
  // If there's an active search, re-run it with new filters
  if (searchStore.currentKeyword) {
    searchStore.search(searchStore.currentKeyword, filters);
  }
}

/**
 * Reset all filters
 * Requirements: 2.3-2.5, 8.2
 */
function handleReset() {
  localFilters.value = {
    fileType: null,
    sizeMin: null,
    sizeMax: null,
    dateFrom: null,
    dateTo: null,
  };
  dateRange.value = undefined;
  searchStore.resetFilters();
  
  // If there's an active search, re-run it without filters
  if (searchStore.currentKeyword) {
    searchStore.search(searchStore.currentKeyword);
  }
}
</script>

<style scoped>
.filter-panel {
  width: 320px;
  max-height: 500px;
  overflow-y: auto;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border-2);
}

.filter-title {
  font-weight: 500;
  font-size: 14px;
}

.filter-section {
  padding: 16px;
}

.filter-label {
  font-weight: 500;
  font-size: 13px;
  margin-bottom: 12px;
  color: var(--color-text-1);
}

.size-range-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.range-separator {
  color: var(--color-text-3);
  flex-shrink: 0;
}

.filter-badge {
  color: var(--color-primary-6);
  font-weight: 500;
}

/* Mobile responsive */
@media (max-width: 767px) {
  .filter-panel {
    width: 280px;
    max-height: 400px;
  }

  .filter-header {
    padding: 10px 12px;
  }

  .filter-title {
    font-size: 13px;
  }

  .filter-section {
    padding: 12px;
  }

  .filter-label {
    font-size: 12px;
    margin-bottom: 10px;
  }

  .size-range-inputs {
    flex-direction: column;
    gap: 6px;
  }

  .range-separator {
    display: none;
  }

  :deep(.arco-input-number) {
    font-size: 13px;
  }

  :deep(.arco-select-view-single) {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .filter-panel {
    width: calc(100vw - 32px);
    max-width: 280px;
  }

  .filter-header {
    padding: 8px 10px;
  }

  .filter-section {
    padding: 10px;
  }
}
</style>
