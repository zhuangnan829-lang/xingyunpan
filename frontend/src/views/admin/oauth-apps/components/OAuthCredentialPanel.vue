<template>
  <section class="credential-panel">
    <div class="section-head">
      <h2>凭据</h2>
      <p>使用这些凭据和端点配置 OAuth 应用程序。</p>
    </div>

    <div class="credential-grid">
      <label v-for="item in credentials" :key="item.label" class="credential-field" :class="{ textarea: item.multiline }">
        <span>{{ item.label }}</span>
        <textarea v-if="item.multiline" :value="item.value" readonly rows="4"></textarea>
        <input v-else :value="item.value" readonly />
        <button type="button" :title="`复制${item.label}`" @click="copy(item.value)">
          <CopyDocument />
        </button>
      </label>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { CopyDocument } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import type { OAuthApp } from '../types';

const props = defineProps<{
  app: OAuthApp;
}>();

const origin = computed(() => (typeof window === 'undefined' ? 'http://117.24.15.9:5212' : window.location.origin));
const enabledScopes = computed(() =>
  props.app.permissions
    .filter((permission) => permission.enabled || permission.required)
    .map((permission) => permission.scope)
    .join(' '),
);

const credentials = computed(() => [
  { label: '客户端 ID', value: props.app.clientId },
  { label: '授权端点', value: `${origin.value}/session/authorize` },
  { label: '令牌端点', value: `${origin.value}/api/v4/session/oauth/token` },
  { label: '刷新端点', value: `${origin.value}/api/v4/session/token/refresh` },
  { label: '用户信息端点', value: `${origin.value}/api/v4/session/oauth/userinfo` },
  { label: '权限范围', value: enabledScopes.value, multiline: true },
]);

const copy = async (value: string) => {
  try {
    await navigator.clipboard.writeText(value);
    ElMessage.success('已复制到剪贴板');
  } catch {
    ElMessage.error('复制失败，请手动复制');
  }
};
</script>

<style scoped>
.credential-panel {
  display: grid;
  gap: 18px;
}

.section-head {
  display: grid;
  gap: 10px;
}

.section-head h2 {
  margin: 0;
  color: #172033;
  font-size: 28px;
  font-weight: 860;
}

.section-head p {
  margin: 0;
  color: #617187;
  font-size: 15px;
}

.credential-grid {
  display: grid;
  gap: 10px;
  max-width: 720px;
}

.credential-field {
  position: relative;
  display: grid;
}

.credential-field span {
  position: absolute;
  top: -9px;
  left: 18px;
  z-index: 1;
  padding: 0 8px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 252, 255, 0.96));
  color: #68788b;
  font-size: 12px;
  font-weight: 800;
}

.credential-field input,
.credential-field textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(211, 224, 238, 0.96);
  border-radius: 16px;
  outline: 0;
  background: rgba(255, 255, 255, 0.72);
  color: #162336;
  font: inherit;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.13);
}

.credential-field input {
  min-height: 52px;
  padding: 0 52px 0 18px;
}

.credential-field textarea {
  min-height: 116px;
  resize: vertical;
  padding: 18px 52px 14px 18px;
  line-height: 1.6;
}

.credential-field button {
  position: absolute;
  right: 12px;
  top: 11px;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 0;
  background: transparent;
  color: #7a8794;
  cursor: pointer;
}

.credential-field button svg {
  width: 18px;
  height: 18px;
}

.credential-field.textarea button {
  top: 42px;
}
</style>
