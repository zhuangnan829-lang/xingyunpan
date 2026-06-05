<template>
  <div v-if="visible" ref="panelRef" class="filter-popover">
    <label class="field-block">
      <span>所有者 ID</span>
      <input
        :value="draft.ownerId"
        class="field-input"
        type="text"
        placeholder="留空表示不过滤此项。"
        @input="updateOwnerId(($event.target as HTMLInputElement).value)"
      />
    </label>

    <label class="field-block">
      <span>搜索文件名</span>
      <input
        :value="draft.keyword"
        class="field-input"
        type="text"
        placeholder="留空表示不过滤此项。"
        @input="updateKeyword(($event.target as HTMLInputElement).value)"
      />
    </label>

    <label class="field-block">
      <span>存储策略</span>
      <select
        :value="String(draft.storagePolicyId)"
        class="field-input field-select"
        @change="updateStoragePolicy(($event.target as HTMLSelectElement).value)"
      >
        <option value="0">全部</option>
        <option v-for="policy in policies" :key="policy.id" :value="String(policy.id)">
          {{ policy.name }}
        </option>
      </select>
    </label>

    <div class="field-block">
      <span>其他条件</span>
      <label class="check-row">
        <input v-model="draft.hasShareLink" type="checkbox" />
        <span>存在分享链接</span>
      </label>
      <label class="check-row">
        <input v-model="draft.hasDirectLink" type="checkbox" />
        <span>存在中转直链</span>
      </label>
      <label class="check-row">
        <input v-model="draft.uploading" type="checkbox" />
        <span>上传中</span>
      </label>
    </div>

    <div class="action-row">
      <button class="ghost-button" type="button" @click="reset">重置</button>
      <button class="primary-button" type="button" @click="apply">应用</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { StoragePolicyPayload } from '@/api/storage-policy';

export interface FileFilters {
  ownerId: string;
  keyword: string;
  storagePolicyId: number;
  hasShareLink: boolean;
  hasDirectLink: boolean;
  uploading: boolean;
}

const props = defineProps<{
  visible: boolean;
  modelValue: FileFilters;
  policies: StoragePolicyPayload[];
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: FileFilters): void;
  (event: 'apply', value: FileFilters): void;
  (event: 'reset', value: FileFilters): void;
}>();

const panelRef = ref<HTMLElement | null>(null);

const emptyFilters = (): FileFilters => ({
  ownerId: '',
  keyword: '',
  storagePolicyId: 0,
  hasShareLink: false,
  hasDirectLink: false,
  uploading: false,
});

const draft = ref<FileFilters>(emptyFilters());

watch(
  () => props.modelValue,
  (value) => {
    draft.value = { ...value };
  },
  { immediate: true, deep: true },
);

const normalizedDraft = computed<FileFilters>(() => ({
  ownerId: draft.value.ownerId.trim(),
  keyword: draft.value.keyword.trim(),
  storagePolicyId: Number.isFinite(draft.value.storagePolicyId) ? draft.value.storagePolicyId : 0,
  hasShareLink: draft.value.hasShareLink,
  hasDirectLink: draft.value.hasDirectLink,
  uploading: draft.value.uploading,
}));

const updateOwnerId = (value: string) => {
  draft.value.ownerId = value.replace(/[^\d]/g, '');
};

const updateKeyword = (value: string) => {
  draft.value.keyword = value;
};

const updateStoragePolicy = (value: string) => {
  draft.value.storagePolicyId = Number(value) || 0;
};

const apply = () => {
  const value = normalizedDraft.value;
  emit('update:modelValue', value);
  emit('apply', value);
};

const reset = () => {
  const value = emptyFilters();
  draft.value = value;
  emit('update:modelValue', value);
  emit('reset', value);
};

defineExpose({
  panelRef,
});
</script>

<style scoped>
.filter-popover {
  position: absolute;
  top: calc(100% + 14px);
  left: 0;
  z-index: 20;
  width: min(430px, calc(100vw - 48px));
  display: grid;
  gap: 18px;
  padding: 24px;
  border: 1px solid rgba(223, 231, 243, 0.96);
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(250, 252, 255, 0.98));
  box-shadow: 0 22px 48px rgba(109, 137, 178, 0.22);
}

.field-block {
  display: grid;
  gap: 10px;
}

.field-block > span {
  color: #1f2937;
  font-size: 15px;
  font-weight: 800;
}

.field-input {
  min-height: 58px;
  padding: 0 20px;
  border: 1px solid #d8e1ec;
  border-radius: 19px;
  background: #fff;
  color: #1f2937;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.field-input:focus {
  border-color: #2a7de1;
  box-shadow: 0 0 0 4px rgba(42, 125, 225, 0.14);
}

.field-select {
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, #7a7f87 50%),
    linear-gradient(135deg, #7a7f87 50%, transparent 50%);
  background-position:
    calc(100% - 28px) calc(50% - 3px),
    calc(100% - 18px) calc(50% - 3px);
  background-size: 10px 10px, 10px 10px;
  background-repeat: no-repeat;
}

.check-row {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #3a4351;
  font-size: 14px;
}

.check-row input {
  width: 22px;
  height: 22px;
  accent-color: #2a7de1;
}

.action-row {
  display: flex;
  justify-content: space-between;
  gap: 14px;
}

.ghost-button,
.primary-button {
  min-width: 98px;
  min-height: 44px;
  padding: 0 20px;
  border-radius: 16px;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
}

.ghost-button {
  border: 1px solid #8eb8f3;
  background: #fff;
  color: #1f7ae0;
}

.primary-button {
  border: 0;
  background: linear-gradient(135deg, #1f7ae0, #2f8af0);
  color: #fff;
}
</style>
