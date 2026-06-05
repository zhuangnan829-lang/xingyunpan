import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { listMusic } from '@/api/music';
import type { FileItem } from '@/types/file';

export type MusicSortMode = 'recent' | 'name' | 'size';
export type MusicViewMode = 'grid' | 'list' | 'album';

const audioExtensions = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'oga', 'm4a', 'wma', 'ape', 'opus', 'amr', 'mid', 'midi'];

function isAudioFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return mime === 'audio' || mime.startsWith('audio/') || audioExtensions.includes(extension);
}

export function useMusicWorkspace() {
  const tracks = ref<FileItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<MusicSortMode>('recent');
  const viewMode = ref<MusicViewMode>('grid');
  const cardSize = ref(260);
  const selectedIds = ref<number[]>([]);
  const totalSize = ref(0);
  const totalTracks = ref(0);

  const filteredTracks = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? tracks.value.filter((item) => item.name.toLowerCase().includes(query))
      : tracks.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
      if (sortMode.value === 'size') return (b.size || 0) - (a.size || 0);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  });

  const selectedTracks = computed(() => tracks.value.filter((item) => selectedIds.value.includes(item.id)));
  const selectedSize = computed(() => selectedTracks.value.reduce((sum, item) => sum + (item.size || 0), 0));

  function toggleSelect(file: FileItem) {
    selectedIds.value = selectedIds.value.includes(file.id)
      ? selectedIds.value.filter((id) => id !== file.id)
      : [...selectedIds.value, file.id];
  }

  function selectOnly(file: FileItem) {
    selectedIds.value = [file.id];
  }

  function clearSelection() {
    selectedIds.value = [];
  }

  function removeTracksByIds(ids: number[]) {
    const idSet = new Set(ids);
    tracks.value = tracks.value.filter((item) => !idSet.has(item.id));
    selectedIds.value = selectedIds.value.filter((id) => !idSet.has(id));
  }

  function upsertTracks(items: FileItem[]) {
    const audioItems = items.filter(isAudioFile).filter((item) => item.id > 0);
    if (!audioItems.length) return;

    const byId = new Map(tracks.value.map((item) => [item.id, item]));
    audioItems.forEach((item) => byId.set(item.id, { ...byId.get(item.id), ...item }));
    tracks.value = [...byId.values()];
  }

  async function refreshMusic() {
    loading.value = true;
    try {
      const result = await listMusic({
        keyword: '',
        page: 1,
        page_size: 2000,
        sort: sortMode.value,
      });
      tracks.value = result.files.filter(isAudioFile);
      totalTracks.value = result.total;
      totalSize.value = result.total_size;
      selectedIds.value = selectedIds.value.filter((id) => tracks.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载音乐失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshMusic);

  return {
    cardSize,
    clearSelection,
    filteredTracks,
    keyword,
    loading,
    refreshMusic,
    removeTracksByIds,
    selectedIds,
    selectedSize,
    selectedTracks,
    selectOnly,
    sortMode,
    toggleSelect,
    totalSize,
    totalTracks,
    tracks,
    upsertTracks,
    viewMode,
  };
}
