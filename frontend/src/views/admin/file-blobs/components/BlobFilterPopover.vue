<template>
  <div v-if="visible" class="popover-card">
    <label class="field">
      <span class="field-label">所有者 ID</span>
      <input
        :value="modelValue.ownerId"
        type="text"
        class="field-input"
        placeholder="留空表示不过滤此项。"
        @input="updateField('ownerId', ($event.target as HTMLInputElement).value)"
      />
    </label>

    <label class="field">
      <span class="field-label">类型</span>
      <select :value="modelValue.kind" class="field-select" @change="updateField('kind', ($event.target as HTMLSelectElement).value)">
        <option value="all">全部</option>
        <option value="version">版本</option>
        <option value="thumbnail">缩略图</option>
        <option value="live-photo">Live Photo</option>
      </select>
    </label>

    <label class="field">
      <span class="field-label">存储策略</span>
      <select
        :value="String(modelValue.storagePolicyId)"
        class="field-select"
        @change="updateStoragePolicy(($event.target as HTMLSelectElement).value)"
      >
        <option value="all">全部</option>
        <option v-for="policy in policies" :key="policy.id" :value="String(policy.id)">
          {{ policy.name }}
        </option>
      </select>
    </label>

    <div class="actions">
      <button class="ghost-button" type="button" @click="$emit('reset')">重置</button>
      <button class="primary-button" type="button" @click="$emit('apply')">应用筛选</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BlobFilterState } from '../types';

interface PolicyOption {
  id: number;
  name: string;
}

const props = defineProps<{
  modelValue: BlobFilterState;
  visible: boolean;
  policies: PolicyOption[];
}>();

const emit = defineEmits<{
  'update:modelValue': [value: BlobFilterState];
  apply: [];
  reset: [];
}>();

function updateField<K extends keyof BlobFilterState>(key: K, value: BlobFilterState[K] | string) {
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: value,
  });
}

function updateStoragePolicy(value: string) {
  emit('update:modelValue', {
    ...props.modelValue,
    storagePolicyId: value === 'all' ? 'all' : Number(value),
  });
}
</script>

<style scoped>
.popover-card {
  position: absolute;
  top: calc(100% + 14px);
  left: 0;
  z-index: 30;
  display: grid;
  gap: 18px;
  width: min(428px, calc(100vw - 48px));
  padding: 22px;
  border: 1px solid #dde5ef;
  border-radius: 22px;
  background: #ffffff;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.16);
}

.field {
  display: grid;
  gap: 10px;
}

.field-label {
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
}

.field-input,
.field-select {
  min-height: 58px;
  padding: 0 18px;
  border: 1px solid #cfd8e3;
  border-radius: 18px;
  background: #ffffff;
  color: #111827;
  font-size: 15px;
  outline: none;
}

.field-input:focus,
.field-select:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.ghost-button,
.primary-button {
  min-height: 44px;
  padding: 0 18px;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.ghost-button {
  border: 1px solid #d6deea;
  background: #ffffff;
  color: #475569;
}

.primary-button {
  border: 0;
  background: linear-gradient(135deg, #2f7ed8, #2ab3ff);
  color: #ffffff;
}
</style>
