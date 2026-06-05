import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { listVideos } from '@/api/video';
import type { FileItem } from '@/types/file';

export type VideoSortMode = 'recent' | 'name' | 'size';
export type VideoViewMode = 'grid' | 'list' | 'cinema';

const videoExtensions = [
  'mp4',
  'mov',
  'mkv',
  'webm',
  'avi',
  'wmv',
  'flv',
  'm4v',
  'mpeg',
  'mpg',
  '3gp',
  '3g2',
  'ts',
  'mts',
  'm2ts',
  'rm',
  'rmvb',
  'vob',
  'ogv',
  'asf',
  'divx',
];

function isVideoFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return mime === 'video' || mime.startsWith('video/') || videoExtensions.includes(extension);
}

export function useVideosWorkspace() {
  const videos = ref<FileItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<VideoSortMode>('recent');
  const viewMode = ref<VideoViewMode>('grid');
  const posterEnabled = ref(true);
  const tileSize = ref(280);
  const selectedIds = ref<number[]>([]);
  const totalSize = ref(0);
  const totalVideos = ref(0);

  const filteredVideos = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? videos.value.filter((item) => item.name.toLowerCase().includes(query))
      : videos.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
      if (sortMode.value === 'size') return (b.size || 0) - (a.size || 0);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  });

  const selectedVideos = computed(() => videos.value.filter((item) => selectedIds.value.includes(item.id)));
  const selectedSize = computed(() => selectedVideos.value.reduce((sum, item) => sum + (item.size || 0), 0));

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

  function removeVideosByIds(ids: number[]) {
    const idSet = new Set(ids);
    videos.value = videos.value.filter((item) => !idSet.has(item.id));
    selectedIds.value = selectedIds.value.filter((id) => !idSet.has(id));
  }

  function upsertVideos(items: FileItem[]) {
    const videoItems = items.filter(isVideoFile).filter((item) => item.id > 0);
    if (!videoItems.length) return;

    const byId = new Map(videos.value.map((item) => [item.id, item]));
    videoItems.forEach((item) => byId.set(item.id, { ...byId.get(item.id), ...item }));
    videos.value = [...byId.values()];
  }

  async function refreshVideos() {
    loading.value = true;
    try {
      const result = await listVideos({
        keyword: '',
        page: 1,
        page_size: 2000,
        sort: sortMode.value,
      });
      videos.value = result.files.filter(isVideoFile);
      totalVideos.value = result.total;
      totalSize.value = result.total_size;
      selectedIds.value = selectedIds.value.filter((id) => videos.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载视频失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshVideos);

  return {
    clearSelection,
    filteredVideos,
    keyword,
    loading,
    posterEnabled,
    refreshVideos,
    removeVideosByIds,
    selectedIds,
    selectedSize,
    selectedVideos,
    selectOnly,
    sortMode,
    tileSize,
    toggleSelect,
    totalSize,
    totalVideos,
    upsertVideos,
    videos,
    viewMode,
  };
}
