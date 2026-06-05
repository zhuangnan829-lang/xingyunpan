import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { listSharedWithMe, type SharedWithMeItem } from '@/api/shared-with-me';

export type SharedSortMode = 'recent' | 'name' | 'owner';
export type SharedViewMode = 'grid' | 'list' | 'compact';

export function useSharedWithMeWorkspace() {
  const items = ref<SharedWithMeItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<SharedSortMode>('recent');
  const viewMode = ref<SharedViewMode>('grid');
  const cardSize = ref(260);
  const selectedIds = ref<string[]>([]);

  const filteredItems = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? items.value.filter((item) =>
          `${item.file_name} ${item.owner_name} ${item.permission}`.toLowerCase().includes(query),
        )
      : items.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.file_name.localeCompare(b.file_name, 'zh-CN');
      if (sortMode.value === 'owner') return a.owner_name.localeCompare(b.owner_name, 'zh-CN');
      return new Date(b.shared_at).getTime() - new Date(a.shared_at).getTime();
    });
  });

  const selectedItems = computed(() => items.value.filter((item) => selectedIds.value.includes(item.id)));
  const selectedSize = computed(() => selectedItems.value.reduce((sum, item) => sum + (item.file_size || 0), 0));

  function toggleSelect(item: SharedWithMeItem) {
    selectedIds.value = selectedIds.value.includes(item.id)
      ? selectedIds.value.filter((id) => id !== item.id)
      : [...selectedIds.value, item.id];
  }

  function selectOnly(item: SharedWithMeItem) {
    selectedIds.value = [item.id];
  }

  function clearSelection() {
    selectedIds.value = [];
  }

  async function refreshSharedWithMe() {
    loading.value = true;
    try {
      items.value = await listSharedWithMe();
      selectedIds.value = selectedIds.value.filter((id) => items.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载与我分享失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshSharedWithMe);

  return {
    cardSize,
    clearSelection,
    filteredItems,
    items,
    keyword,
    loading,
    refreshSharedWithMe,
    selectedIds,
    selectedItems,
    selectedSize,
    selectOnly,
    sortMode,
    toggleSelect,
    viewMode,
  };
}
