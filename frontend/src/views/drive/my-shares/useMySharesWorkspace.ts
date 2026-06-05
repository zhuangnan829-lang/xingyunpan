import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getMyShares, type MyShare } from '@/api/share';
import { resolveShareURL } from '@/utils/public-url';

export type MyShareSortMode = 'recent' | 'name' | 'expires' | 'visits';
export type MyShareViewMode = 'grid' | 'list' | 'compact';

export type MyShareItem = MyShare & {
  id: string;
  display_name: string;
  share_link: string;
  expired: boolean;
};

function normalizeShare(item: MyShare): MyShareItem {
  const token = item.share_token || item.share_id;
  const link = resolveShareURL(item.share_url || '', token);
  const expiresAt = item.expires_at ? new Date(item.expires_at).getTime() : null;

  return {
    ...item,
    id: String(item.share_id),
    display_name: item.file_names?.join(', ') || 'Untitled share',
    share_link: link,
    expired: expiresAt !== null && expiresAt <= Date.now(),
  };
}

export function useMySharesWorkspace() {
  const items = ref<MyShareItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<MyShareSortMode>('recent');
  const viewMode = ref<MyShareViewMode>('grid');
  const cardSize = ref(260);
  const selectedIds = ref<string[]>([]);

  const filteredItems = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? items.value.filter((item) =>
          `${item.display_name} ${item.share_link} ${item.expired ? 'expired' : 'active'}`
            .toLowerCase()
            .includes(query),
        )
      : items.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.display_name.localeCompare(b.display_name, 'zh-CN');
      if (sortMode.value === 'expires') {
        const at = a.expires_at ? new Date(a.expires_at).getTime() : Number.MAX_SAFE_INTEGER;
        const bt = b.expires_at ? new Date(b.expires_at).getTime() : Number.MAX_SAFE_INTEGER;
        return at - bt;
      }
      if (sortMode.value === 'visits') return (b.access_count || 0) - (a.access_count || 0);
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
  });

  const selectedItems = computed(() => items.value.filter((item) => selectedIds.value.includes(item.id)));

  function toggleSelect(item: MyShareItem) {
    selectedIds.value = selectedIds.value.includes(item.id)
      ? selectedIds.value.filter((id) => id !== item.id)
      : [...selectedIds.value, item.id];
  }

  function selectOnly(item: MyShareItem) {
    selectedIds.value = [item.id];
  }

  function clearSelection() {
    selectedIds.value = [];
  }

  async function refreshMyShares() {
    loading.value = true;
    try {
      items.value = (await getMyShares()).map(normalizeShare);
      selectedIds.value = selectedIds.value.filter((id) => items.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载我的分享失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshMyShares);

  return {
    cardSize,
    clearSelection,
    filteredItems,
    items,
    keyword,
    loading,
    refreshMyShares,
    selectedIds,
    selectedItems,
    selectOnly,
    sortMode,
    toggleSelect,
    viewMode,
  };
}
