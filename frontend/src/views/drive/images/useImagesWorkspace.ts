import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { searchFiles } from '@/api/search';
import { downloadFile as apiDownloadFile, getAuthenticatedFileDownloadUrl } from '@/api/file';
import type { FileItem } from '@/types/file';

export type ImageSortMode = 'recent' | 'name' | 'size';
export type ImageCardSize = 'compact' | 'comfortable' | 'large';
export type WorkspaceFileKind = 'file' | 'markdown' | 'text' | 'drawio' | 'dwb' | 'excalidraw';

const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'avif'];

function isImageFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return mime === 'image' || mime.startsWith('image/') || imageExtensions.includes(extension);
}

export function useImagesWorkspace() {
  const images = ref<FileItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<ImageSortMode>('recent');
  const cardSize = ref<ImageCardSize>('comfortable');
  const selectedIds = ref<number[]>([]);
  const thumbnailUrls = reactive<Record<number, string>>({});
  const loadingThumbnailIds = new Set<number>();
  const failedThumbnailIds = new Set<number>();

  const filteredImages = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? images.value.filter((item) => item.name.toLowerCase().includes(query))
      : images.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
      if (sortMode.value === 'size') return b.size - a.size;
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  });

  const totalSize = computed(() => images.value.reduce((sum, item) => sum + (item.size || 0), 0));
  const selectedImages = computed(() => images.value.filter((item) => selectedIds.value.includes(item.id)));
  const selectedSize = computed(() => selectedImages.value.reduce((sum, item) => sum + (item.size || 0), 0));

  function imageUrl(file: FileItem) {
    const url = new URL(getAuthenticatedFileDownloadUrl(file.id, true));
    url.searchParams.set('v', file.updated_at || String(file.id));
    return url.toString();
  }

  function thumbnailUrl(file: FileItem) {
    return thumbnailUrls[file.id] || '';
  }

  async function ensureThumbnail(file: FileItem) {
    if (thumbnailUrls[file.id] || loadingThumbnailIds.has(file.id) || failedThumbnailIds.has(file.id)) {
      return;
    }

    loadingThumbnailIds.add(file.id);
    try {
      const blob = await apiDownloadFile(file.id);
      if (blob.size <= 0) {
        failedThumbnailIds.add(file.id);
        return;
      }

      file.size = blob.size;
      thumbnailUrls[file.id] = URL.createObjectURL(blob);
    } catch (error) {
      failedThumbnailIds.add(file.id);
    } finally {
      loadingThumbnailIds.delete(file.id);
    }
  }

  function revokeStaleThumbnails(activeIds: Set<number>) {
    Object.entries(thumbnailUrls).forEach(([rawId, url]) => {
      const id = Number(rawId);
      if (activeIds.has(id)) return;
      URL.revokeObjectURL(url);
      delete thumbnailUrls[id];
      failedThumbnailIds.delete(id);
    });
  }

  function revokeAllThumbnails() {
    Object.values(thumbnailUrls).forEach((url) => URL.revokeObjectURL(url));
    Object.keys(thumbnailUrls).forEach((id) => delete thumbnailUrls[Number(id)]);
    loadingThumbnailIds.clear();
    failedThumbnailIds.clear();
  }

  function removeImagesByIds(ids: number[]) {
    const idSet = new Set(ids);
    if (idSet.size === 0) return;

    images.value = images.value.filter((item) => !idSet.has(item.id));
    selectedIds.value = selectedIds.value.filter((id) => !idSet.has(id));

    idSet.forEach((id) => {
      if (thumbnailUrls[id]) {
        URL.revokeObjectURL(thumbnailUrls[id]);
        delete thumbnailUrls[id];
      }
      loadingThumbnailIds.delete(id);
      failedThumbnailIds.delete(id);
    });
  }

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

  function uniqueImageName(baseName: string) {
    const existingNames = new Set(images.value.map((item) => item.name));
    if (!existingNames.has(baseName)) return baseName;

    const dotIndex = baseName.lastIndexOf('.');
    const stem = dotIndex > 0 ? baseName.slice(0, dotIndex) : baseName;
    const extension = dotIndex > 0 ? baseName.slice(dotIndex) : '';
    let index = 2;
    let nextName = `${stem} ${index}${extension}`;
    while (existingNames.has(nextName)) {
      index += 1;
      nextName = `${stem} ${index}${extension}`;
    }
    return nextName;
  }

  async function refreshImages() {
    loading.value = true;
    try {
      const result = await searchFiles({
        keyword: '',
        file_type: null,
        page: 1,
        page_size: 2000,
      });
      images.value = result.files.filter(isImageFile);
      revokeStaleThumbnails(new Set(images.value.map((item) => item.id)));
      selectedIds.value = selectedIds.value.filter((id) => images.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载图片失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshImages);
  onUnmounted(revokeAllThumbnails);

  return {
    cardSize,
    clearSelection,
    filteredImages,
    ensureThumbnail,
    imageUrl,
    images,
    keyword,
    loading,
    refreshImages,
    removeImagesByIds,
    selectedIds,
    selectedImages,
    selectedSize,
    selectOnly,
    sortMode,
    thumbnailUrl,
    toggleSelect,
    totalSize,
    uniqueImageName,
  };
}
