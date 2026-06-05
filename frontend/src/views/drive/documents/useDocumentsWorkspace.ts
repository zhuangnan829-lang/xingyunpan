import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { listDocuments } from '@/api/documents';
import type { FileItem } from '@/types/file';

export type DocumentSortMode = 'recent' | 'name' | 'size';
export type DocumentViewMode = 'grid' | 'list' | 'compact';

const documentExtensions = [
  'pdf',
  'doc',
  'docx',
  'xls',
  'xlsx',
  'ppt',
  'pptx',
  'txt',
  'md',
  'markdown',
  'csv',
  'json',
  'xml',
  'yaml',
  'yml',
  'rtf',
  'odt',
  'ods',
  'odp',
  'epub',
  'html',
  'htm',
];

function isDocumentFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return (
    mime === 'document' ||
    mime.includes('pdf') ||
    mime.includes('word') ||
    mime.includes('excel') ||
    mime.includes('spreadsheet') ||
    mime.includes('powerpoint') ||
    mime.includes('presentation') ||
    mime.startsWith('text/') ||
    ['application/json', 'application/xml'].includes(mime) ||
    documentExtensions.includes(extension)
  );
}

export function useDocumentsWorkspace() {
  const documents = ref<FileItem[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const sortMode = ref<DocumentSortMode>('recent');
  const viewMode = ref<DocumentViewMode>('grid');
  const cardSize = ref(260);
  const selectedIds = ref<number[]>([]);
  const totalSize = ref(0);
  const totalDocuments = ref(0);

  const filteredDocuments = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    const source = query
      ? documents.value.filter((item) => item.name.toLowerCase().includes(query))
      : documents.value;

    return [...source].sort((a, b) => {
      if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
      if (sortMode.value === 'size') return (b.size || 0) - (a.size || 0);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  });

  const selectedDocuments = computed(() => documents.value.filter((item) => selectedIds.value.includes(item.id)));
  const selectedSize = computed(() => selectedDocuments.value.reduce((sum, item) => sum + (item.size || 0), 0));

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

  function removeDocumentsByIds(ids: number[]) {
    const idSet = new Set(ids);
    documents.value = documents.value.filter((item) => !idSet.has(item.id));
    selectedIds.value = selectedIds.value.filter((id) => !idSet.has(id));
  }

  function upsertDocuments(items: FileItem[]) {
    const documentItems = items.filter(isDocumentFile).filter((item) => item.id > 0);
    if (!documentItems.length) return;

    const byId = new Map(documents.value.map((item) => [item.id, item]));
    documentItems.forEach((item) => byId.set(item.id, { ...byId.get(item.id), ...item }));
    documents.value = [...byId.values()];
  }

  async function refreshDocuments() {
    loading.value = true;
    try {
      const result = await listDocuments({
        keyword: '',
        page: 1,
        page_size: 2000,
        sort: sortMode.value,
      });
      documents.value = result.files.filter(isDocumentFile);
      totalDocuments.value = result.total;
      totalSize.value = result.total_size;
      selectedIds.value = selectedIds.value.filter((id) => documents.value.some((item) => item.id === id));
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载文档失败');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refreshDocuments);

  return {
    cardSize,
    clearSelection,
    documents,
    filteredDocuments,
    keyword,
    loading,
    refreshDocuments,
    removeDocumentsByIds,
    selectedDocuments,
    selectedIds,
    selectedSize,
    selectOnly,
    sortMode,
    totalDocuments,
    totalSize,
    toggleSelect,
    upsertDocuments,
    viewMode,
  };
}
