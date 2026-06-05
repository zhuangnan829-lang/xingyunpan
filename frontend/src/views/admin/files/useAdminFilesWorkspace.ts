import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { deleteAdminFile, listAdminFiles, type AdminFilePayload } from '@/api/admin-files';
import { listStoragePolicies, type StoragePolicyPayload } from '@/api/storage-policy';

export interface FileFiltersState {
  ownerId: string;
  keyword: string;
  storagePolicyId: number | 0;
}

export interface DisplayFileItem {
  id: string;
  rawId: number;
  fileName: string;
  fileSize: number;
  occupiedSize: number;
  filePath: string;
  ownerId: number;
  ownerName: string;
  storagePolicyName: string;
  createdAt: string;
  updatedAt: string;
  hasShareLink: boolean;
  hasDirectLink: boolean;
  uploading: boolean;
  extension: string;
  isFolder: boolean;
  displayIcon: string;
  displayIconTint: string;
  displayIconLabel: string;
  displayIconSource: string;
}

const createDefaultFilters = (): FileFiltersState => ({
  ownerId: '',
  keyword: '',
  storagePolicyId: 0,
});

const extensionFromName = (name: string) => {
  const parts = name.split('.');
  return parts.length > 1 ? parts[parts.length - 1].toLowerCase() : 'file';
};

const mapAdminFileToDisplay = (file: AdminFilePayload): DisplayFileItem => ({
  id: `live-${file.id}`,
  rawId: file.id,
  fileName: file.file_name,
  fileSize: file.file_size,
  occupiedSize: file.occupied_size,
  filePath: file.file_path,
  ownerId: file.owner_id,
  ownerName: file.owner_username,
  storagePolicyName: file.storage_policy_name || '未分配存储策略',
  createdAt: file.created_at,
  updatedAt: file.updated_at,
  hasShareLink: file.has_share_link,
  hasDirectLink: file.has_direct_link,
  uploading: file.uploading,
  extension: extensionFromName(file.file_name),
  isFolder: file.is_folder,
  displayIcon: file.display_icon || '',
  displayIconTint: file.display_icon_tint || '',
  displayIconLabel: file.display_icon_label || '',
  displayIconSource: file.display_icon_source || '',
});

export function useAdminFilesWorkspace() {
  const filterAnchorRef = ref<HTMLElement | null>(null);
  const filterVisible = ref(false);
  const loading = ref(false);
  const files = ref<AdminFilePayload[]>([]);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const selectedIds = ref<string[]>([]);
  const storagePolicies = ref<StoragePolicyPayload[]>([]);
  const sortDirection = ref<'desc' | 'asc'>('desc');
  const filters = ref<FileFiltersState>(createDefaultFilters());

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));
  const displayFiles = computed<DisplayFileItem[]>(() =>
    [...files.value]
      .map(mapAdminFileToDisplay)
      .sort((a, b) => (sortDirection.value === 'desc' ? b.rawId - a.rawId : a.rawId - b.rawId)),
  );

  const pagedFiles = computed(() => displayFiles.value);
  const pageFileSize = computed(() => pagedFiles.value.reduce((sum, file) => sum + file.fileSize, 0));
  const allPageSelected = computed(() => pagedFiles.value.length > 0 && pagedFiles.value.every((file) => selectedIds.value.includes(file.id)));

  const formatSize = (value: number): string => {
    if (!value) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = value;
    let unitIndex = 0;
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex += 1;
    }
    const digits = size >= 10 || unitIndex === 0 ? 0 : 1;
    return `${size.toFixed(digits)} ${units[unitIndex]}`;
  };

  const formatDate = (value: string): string => {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return '-';
    }
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: 'numeric',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
  };

  const getFileIconLabel = (file: DisplayFileItem): string => {
    if (file.displayIcon) return file.displayIcon;
    if (file.isFolder) return 'DIR';
    const ext = file.extension || 'file';
    if (ext === 'docx') return 'DOC';
    if (ext === 'xlsx') return 'XLS';
    if (ext.length <= 4) return ext.toUpperCase();
    return 'FILE';
  };

  const getFileIconTone = (file: DisplayFileItem): string => {
    if (file.displayIcon && file.displayIconTint) return 'is-custom';
    if (file.isFolder) return 'is-folder';
    const ext = file.extension;
    if (ext === 'pdf') return 'is-pdf';
    if (['doc', 'docx'].includes(ext)) return 'is-docx';
    if (['xls', 'xlsx', 'csv'].includes(ext)) return 'is-xlsx';
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext)) return 'is-image';
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return 'is-archive';
    return 'is-file';
  };

  const getFileIconStyle = (file: DisplayFileItem): Record<string, string> => {
    if (!file.displayIconTint) return {};
    return {
      background: file.displayIconTint,
    };
  };

  const getFileIconTitle = (file: DisplayFileItem): string => file.displayIconLabel || file.extension || 'file';

  const fetchFiles = async () => {
    loading.value = true;
    try {
      const response = await listAdminFiles({
        page: page.value,
        page_size: pageSize.value,
        owner_id: filters.value.ownerId ? Number(filters.value.ownerId) : undefined,
        keyword: filters.value.keyword || undefined,
        storage_policy_id: filters.value.storagePolicyId || undefined,
      });
      files.value = response.list;
      total.value = response.total;
      selectedIds.value = [];
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载文件列表失败');
    } finally {
      loading.value = false;
    }
  };

  const fetchMeta = async () => {
    try {
      storagePolicies.value = await listStoragePolicies();
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载存储策略失败');
    }
  };

  const toggleSort = () => {
    sortDirection.value = sortDirection.value === 'desc' ? 'asc' : 'desc';
  };

  const toggleSelect = (id: string, checked: boolean) => {
    if (checked) {
      if (!selectedIds.value.includes(id)) {
        selectedIds.value = [...selectedIds.value, id];
      }
      return;
    }
    selectedIds.value = selectedIds.value.filter((item) => item !== id);
  };

  const toggleSelectAll = (checked: boolean) => {
    if (checked) {
      selectedIds.value = Array.from(new Set([...selectedIds.value, ...pagedFiles.value.map((item) => item.id)]));
      return;
    }
    const currentIds = new Set(pagedFiles.value.map((item) => item.id));
    selectedIds.value = selectedIds.value.filter((id) => !currentIds.has(id));
  };

  const previewFile = (file: DisplayFileItem) => {
    ElMessageBox.alert(
      [
        `所有者：${file.ownerName || '-'} (#${file.ownerId})`,
        `大小：${formatSize(file.fileSize)}`,
        `占用空间：${formatSize(file.occupiedSize)}`,
        `存储策略：${file.storagePolicyName}`,
        `路径：${file.filePath || '-'}`,
        `创建时间：${formatDate(file.createdAt)}`,
      ].join('\n'),
      file.fileName,
      { confirmButtonText: '知道了' },
    );
  };

  const quickFilterOwner = async (ownerIdValue: number) => {
    filters.value.ownerId = String(ownerIdValue);
    page.value = 1;
    await fetchFiles();
  };

  const removeFile = async (file: DisplayFileItem) => {
    try {
      await ElMessageBox.confirm(`确定删除“${file.fileName}”吗？此操作无法撤销。`, '删除文件', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
      });
      await deleteAdminFile(file.rawId);
      ElMessage.success('文件已删除');
      await fetchFiles();
    } catch (error) {
      if (error === 'cancel' || error === 'close') return;
      ElMessage.error(error instanceof Error ? error.message : '删除文件失败');
    }
  };

  const handleAction = (action: 'download' | 'share' | 'rename', file: DisplayFileItem) => {
    const labelMap = {
      download: '下载',
      share: '分享',
      rename: '重命名',
    };
    ElMessage.info(`${labelMap[action]}“${file.fileName}”的真实接口还可以继续接入。`);
  };

  const applyFilters = async () => {
    page.value = 1;
    filterVisible.value = false;
    await fetchFiles();
  };

  const resetFilters = async () => {
    filters.value = createDefaultFilters();
    page.value = 1;
    filterVisible.value = false;
    await fetchFiles();
  };

  const triggerImport = () => {
    ElMessage.info('导入接口尚未接入，这一版先保留文件系统管理界面。');
  };

  const handleDocumentClick = (event: MouseEvent) => {
    if (!filterVisible.value) return;
    const target = event.target as Node | null;
    if (target && filterAnchorRef.value?.contains(target)) return;
    filterVisible.value = false;
  };

  watch([page, pageSize], () => {
    fetchFiles();
  });

  watch(totalPages, () => {
    if (page.value > totalPages.value) {
      page.value = totalPages.value;
    }
  });

  onMounted(async () => {
    document.addEventListener('mousedown', handleDocumentClick);
    await Promise.all([fetchMeta(), fetchFiles()]);
  });

  onBeforeUnmount(() => {
    document.removeEventListener('mousedown', handleDocumentClick);
  });

  return {
    filterAnchorRef,
    filterVisible,
    loading,
    page,
    pageSize,
    storagePolicies,
    selectedIds,
    sortDirection,
    filters,
    total,
    totalPages,
    pageFileSize,
    displayFiles,
    pagedFiles,
    allPageSelected,
    formatSize,
    formatDate,
    getFileIconLabel,
    getFileIconTone,
    getFileIconStyle,
    getFileIconTitle,
    fetchFiles,
    toggleSort,
    toggleSelect,
    toggleSelectAll,
    previewFile,
    quickFilterOwner,
    removeFile,
    handleAction,
    applyFilters,
    resetFilters,
    triggerImport,
  };
}
