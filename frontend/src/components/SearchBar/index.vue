<template>
  <div class="search-bar">
    <a-input-search
      v-model="localKeyword"
      :placeholder="placeholder"
      :style="{ width: '100%' }"
      allow-clear
      @search="handleSearch"
      @clear="handleClear"
      @focus="showHistoryDropdown = true"
      @blur="handleBlur"
    >
      <template #prefix>
        <icon-search />
      </template>
      <template #suffix>
        <span v-if="!localKeyword" class="shortcut-hint">Ctrl+F</span>
      </template>
    </a-input-search>

    <!-- Search History Dropdown -->
    <div
      v-if="showHistoryDropdown && recentHistory.length > 0"
      class="search-history-dropdown"
    >
      <div class="search-history-header">
        <span class="search-history-title">搜索历史</span>
        <a-button
          type="text"
          size="mini"
          @click="handleClearHistory"
        >
          清除历史
        </a-button>
      </div>
      <div class="search-history-list">
        <div
          v-for="item in recentHistory"
          :key="item.keyword"
          class="search-history-item"
          @mousedown.prevent="handleHistoryClick(item.keyword)"
        >
          <icon-clock-circle class="history-icon" />
          <span class="history-keyword">{{ item.keyword }}</span>
        </div>
      </div>
    </div>

    <!-- Search Suggestions Dropdown -->
    <div
      v-if="showSuggestionsDropdown && suggestions.length > 0"
      class="search-suggestions-dropdown"
    >
      <div class="search-suggestions-header">
        <span class="search-suggestions-title">搜索建议</span>
      </div>
      <div class="search-suggestions-list">
        <div
          v-for="item in suggestions"
          :key="item.keyword"
          class="search-suggestion-item"
          @mousedown.prevent="handleSuggestionClick(item.keyword)"
        >
          <icon-search class="suggestion-icon" />
          <span class="suggestion-keyword">{{ item.keyword }}</span>
          <span v-if="item.count > 0" class="suggestion-count">{{ item.count }} 个结果</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { IconSearch, IconClockCircle } from '@arco-design/web-vue/es/icon';
import { useSearchStore } from '@/stores/search';
import { useFileStore } from '@/stores/file';
import { debounce } from '@/utils/debounce';

// Props
interface Props {
  placeholder?: string;
  showHistory?: boolean;
  debounceMs?: number;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '搜索文件',
  showHistory: true,
  debounceMs: 300,
});

// Emits
interface Emits {
  (e: 'search', keyword: string): void;
  (e: 'clear'): void;
}

const emit = defineEmits<Emits>();

// Stores
const searchStore = useSearchStore();
const fileStore = useFileStore();

// Local state
const localKeyword = ref('');
const showHistoryDropdown = ref(false);
const showSuggestionsDropdown = ref(false);
const suggestions = ref<Array<{ keyword: string; count: number }>>([]);
const isLoadingSuggestions = ref(false);
const inputRef = ref<HTMLInputElement | null>(null);

// Computed
const recentHistory = computed(() => {
  if (!props.showHistory) {
    return [];
  }
  return searchStore.getRecentHistory(5);
});

const showDropdown = computed(() => {
  return showHistoryDropdown.value || showSuggestionsDropdown.value;
});

// Watch fileStore searchKeyword to sync with local state
watch(
  () => fileStore.searchKeyword,
  (newValue) => {
    if (newValue !== localKeyword.value) {
      localKeyword.value = newValue;
    }
  },
  { immediate: true }
);

// Watch local keyword for debounced suggestions
watch(localKeyword, (newValue) => {
  if (newValue && newValue.trim().length > 0) {
    debouncedLoadSuggestions(newValue.trim());
  } else {
    suggestions.value = [];
    showSuggestionsDropdown.value = false;
  }
});

// Methods

/**
 * Load search suggestions with debounce
 * Requirements: 2.1, 2.9
 */
const debouncedLoadSuggestions = debounce(async (keyword: string) => {
  if (!keyword || keyword.length < 2) {
    suggestions.value = [];
    showSuggestionsDropdown.value = false;
    return;
  }

  try {
    isLoadingSuggestions.value = true;
    const results = await searchStore.getSuggestions(keyword);
    suggestions.value = results;
    showSuggestionsDropdown.value = results.length > 0;
    showHistoryDropdown.value = false;
  } catch (error) {
    console.error('Failed to load suggestions:', error);
    suggestions.value = [];
    showSuggestionsDropdown.value = false;
  } finally {
    isLoadingSuggestions.value = false;
  }
}, props.debounceMs);

/**
 * Handle search action (Enter key or search button click)
 * Requirements: 4.1, 4.2, 4.5
 */
function handleSearch(keyword: string): void {
  const trimmedKeyword = keyword.trim();
  
  // If keyword is empty, clear search
  if (!trimmedKeyword) {
    handleClear();
    return;
  }
  
  // Add to search history
  searchStore.addHistory(trimmedKeyword);
  
  // Update file store search keyword
  fileStore.setSearchKeyword(trimmedKeyword);
  
  // Hide dropdown
  showHistoryDropdown.value = false;
  
  // Emit search event
  emit('search', trimmedKeyword);
}

/**
 * Handle clear search
 * Requirements: 4.6
 */
function handleClear(): void {
  localKeyword.value = '';
  fileStore.setSearchKeyword('');
  showHistoryDropdown.value = false;
  emit('clear');
}

/**
 * Handle search history item click
 * Requirements: 4.4, 4.5
 */
function handleHistoryClick(keyword: string): void {
  localKeyword.value = keyword;
  handleSearch(keyword);
}

/**
 * Handle suggestion item click
 * Requirements: 2.9
 */
function handleSuggestionClick(keyword: string): void {
  localKeyword.value = keyword;
  handleSearch(keyword);
}

/**
 * Handle clear all search history
 * Requirements: 4.8
 */
function handleClearHistory(): void {
  searchStore.clearHistory();
  showHistoryDropdown.value = false;
}

/**
 * Handle input blur with delay to allow history item clicks
 */
function handleBlur(): void {
  // Delay hiding dropdown to allow mousedown event on history items
  setTimeout(() => {
    showHistoryDropdown.value = false;
  }, 200);
}

/**
 * Focus the search input (used by keyboard shortcut)
 */
function focusInput(): void {
  // Find the underlying input element inside a-input-search
  const el = document.querySelector('.search-bar input') as HTMLInputElement | null;
  if (el) {
    el.focus();
    el.select();
  }
  showHistoryDropdown.value = true;
}

/**
 * Global keyboard shortcut handler
 * Ctrl+F / Cmd+F: focus search input
 * Esc: clear search and hide dropdown
 */
function handleKeydown(event: KeyboardEvent): void {
  // Ctrl+F or Cmd+F to focus search
  if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
    // Only intercept if not already in an input/textarea
    const target = event.target as HTMLElement;
    if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
      event.preventDefault();
      focusInput();
    }
  }
  
  // Esc to clear search when input is focused
  if (event.key === 'Escape') {
    const target = event.target as HTMLElement;
    if (target.closest('.search-bar')) {
      handleClear();
      (target as HTMLInputElement).blur?.();
    }
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<style scoped>
.search-bar {
  position: relative;
  width: 100%;
}

.search-history-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--color-bg-popup);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-medium);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  max-height: 300px;
  overflow-y: auto;
}

.search-history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
}

.search-history-title {
  font-size: 12px;
  color: var(--color-text-3);
  font-weight: 500;
}

.search-history-list {
  padding: 4px 0;
}

.search-history-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.search-history-item:hover {
  background-color: var(--color-fill-2);
}

.history-icon {
  font-size: 14px;
  color: var(--color-text-3);
  margin-right: 8px;
  flex-shrink: 0;
}

.history-keyword {
  font-size: 14px;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shortcut-hint {
  font-size: 11px;
  color: var(--color-text-4);
  background: var(--color-fill-2);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  padding: 1px 4px;
  line-height: 1.4;
  pointer-events: none;
  user-select: none;
}

/* Search Suggestions Dropdown */
.search-suggestions-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--color-bg-popup);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-medium);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  max-height: 300px;
  overflow-y: auto;
}

.search-suggestions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
}

.search-suggestions-title {
  font-size: 12px;
  color: var(--color-text-3);
  font-weight: 500;
}

.search-suggestions-list {
  padding: 4px 0;
}

.search-suggestion-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.search-suggestion-item:hover {
  background-color: var(--color-fill-2);
}

.suggestion-icon {
  font-size: 14px;
  color: var(--color-text-3);
  margin-right: 8px;
  flex-shrink: 0;
}

.suggestion-keyword {
  font-size: 14px;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.suggestion-count {
  font-size: 12px;
  color: var(--color-text-3);
  margin-left: 8px;
  flex-shrink: 0;
}

/* Mobile responsive */
@media (max-width: 767px) {
  .search-history-dropdown,
  .search-suggestions-dropdown {
    max-height: 250px;
  }

  .search-history-header,
  .search-suggestions-header {
    padding: 6px 10px;
  }

  .search-history-title,
  .search-suggestions-title {
    font-size: 11px;
  }

  .search-history-item,
  .search-suggestion-item {
    padding: 6px 10px;
  }

  .history-keyword,
  .suggestion-keyword {
    font-size: 13px;
  }

  .suggestion-count {
    font-size: 11px;
  }

  .shortcut-hint {
    display: none; /* Hide keyboard shortcut hint on mobile */
  }
}

@media (max-width: 480px) {
  .search-history-dropdown,
  .search-suggestions-dropdown {
    max-height: 200px;
  }

  .history-keyword,
  .suggestion-keyword {
    font-size: 12px;
  }
}
</style>
