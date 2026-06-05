<template>
  <el-dialog :model-value="modelValue" width="620px" class="oauth-create-dialog" @update:model-value="$emit('update:modelValue', $event)">
    <template #header>
      <div class="dialog-head">
        <span>新建 OAuth 应用</span>
        <small>为星云盘外部客户端创建授权入口</small>
      </div>
    </template>

    <form class="create-form" @submit.prevent="$emit('submit')">
      <label>
        <span>应用名称</span>
        <input v-model="draft.name" type="text" placeholder="例如：企业同步客户端" />
      </label>
      <label>
        <span>应用说明</span>
        <textarea v-model="draft.description" rows="3" placeholder="描述这个 OAuth 应用的使用场景"></textarea>
      </label>
      <label>
        <span>回调地址</span>
        <input v-model="draft.redirectUri" type="url" placeholder="https://example.com/oauth/callback" />
      </label>
      <div class="scope-field">
        <span>授权范围</span>
        <div class="scope-options">
          <label v-for="scope in scopeOptions" :key="scope" class="scope-option">
            <input v-model="draft.scopes" type="checkbox" :value="scope" />
            <span>{{ scope }}</span>
          </label>
        </div>
      </div>
    </form>

    <template #footer>
      <button class="dialog-button ghost" type="button" @click="$emit('update:modelValue', false)">取消</button>
      <button class="dialog-button primary" type="button" @click="$emit('submit')">创建应用</button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { OAuthAppDraft } from '../types';

defineProps<{
  modelValue: boolean;
  draft: OAuthAppDraft;
  scopeOptions: string[];
}>();

defineEmits<{
  'update:modelValue': [value: boolean];
  submit: [];
}>();
</script>

<style scoped>
.dialog-head {
  display: grid;
  gap: 6px;
}

.dialog-head span {
  color: #162336;
  font-size: 20px;
  font-weight: 860;
}

.dialog-head small {
  color: #7b8a9d;
  font-size: 13px;
}

.create-form {
  display: grid;
  gap: 16px;
}

.create-form label,
.scope-field {
  display: grid;
  gap: 8px;
}

.create-form label > span,
.scope-field > span {
  color: #314257;
  font-size: 13px;
  font-weight: 820;
}

.create-form input,
.create-form textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #dce8f5;
  border-radius: 14px;
  outline: 0;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  color: #172033;
  font: inherit;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.12);
}

.create-form input {
  min-height: 44px;
  padding: 0 13px;
}

.create-form textarea {
  resize: vertical;
  padding: 12px 13px;
}

.scope-options {
  display: flex;
  flex-wrap: wrap;
  gap: 9px;
}

.scope-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 0 11px;
  border: 1px solid rgba(199, 215, 234, 0.92);
  border-radius: 12px;
  background: #fff;
  color: #34465b;
}

.scope-option input {
  width: 15px;
  min-height: auto;
  accent-color: #1687e0;
}

.dialog-button {
  min-height: 42px;
  padding: 0 16px;
  border-radius: 14px;
  font-weight: 820;
  cursor: pointer;
}

.dialog-button.ghost {
  border: 1px solid #d7e3f0;
  background: #fff;
  color: #526173;
}

.dialog-button.primary {
  border: 0;
  background: linear-gradient(135deg, #1f74e8, #19bddf);
  color: #fff;
  box-shadow: 0 12px 24px rgba(40, 139, 219, 0.22);
}
</style>
