import { ref } from 'vue';
import { loadMapRuntimeConfig, type MapRuntimeApplyResult } from '@/utils/map-runtime';
import type { FileSystemMapRuntimeConfig } from '@/api/file-system-settings';

export function useMapRuntime() {
  const config = ref<FileSystemMapRuntimeConfig | null>(null);
  const applied = ref<MapRuntimeApplyResult | null>(null);
  const loading = ref(false);
  const error = ref('');

  async function refresh() {
    loading.value = true;
    error.value = '';
    try {
      config.value = await loadMapRuntimeConfig();
      return config.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'load map runtime failed';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function markApplied(result: MapRuntimeApplyResult) {
    applied.value = result;
    return result;
  }

  return {
    config,
    applied,
    loading,
    error,
    refresh,
    markApplied,
  };
}
