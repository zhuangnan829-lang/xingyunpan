<template>
  <article class="oauth-app-card" role="button" tabindex="0" @click="$emit('open', app)" @keydown.enter="$emit('open', app)">
    <div class="card-main">
      <div class="app-mark" aria-hidden="true">
        <span></span>
      </div>
      <div class="app-copy">
        <div class="title-row">
          <h3>{{ app.name }}</h3>
          <span v-if="app.isSystem" class="system-badge">系统</span>
        </div>
        <p>{{ app.description }}</p>
        <div class="scope-row">
          <span v-for="scope in visibleScopes" :key="scope" class="scope-chip">{{ scope }}</span>
          <span v-if="hiddenScopeCount > 0" class="scope-chip">+{{ hiddenScopeCount }}</span>
        </div>
      </div>
    </div>

    <div class="meta-grid">
      <div>
        <span>Client ID</span>
        <strong>{{ app.clientId }}</strong>
      </div>
      <div>
        <span>Token 有效期</span>
        <strong>{{ app.tokenTtl }}</strong>
      </div>
    </div>

    <div class="callback-box">
      <span>回调地址</span>
      <strong>{{ app.redirectUris[0] }}</strong>
    </div>

    <footer class="card-actions">
      <button class="status-button" type="button" :class="{ enabled: app.enabled }" @click.stop="$emit('toggle', app)">
        <CircleCheck v-if="app.enabled" />
        <CircleClose v-else />
        <span>{{ app.enabled ? '已启用' : '已停用' }}</span>
      </button>
      <button class="icon-action" type="button" title="编辑应用" @click.stop="$emit('open', app)">
        <EditPen />
      </button>
      <button class="icon-action danger" type="button" title="删除应用" :disabled="app.isSystem" @click.stop="$emit('delete', app)">
        <Delete />
      </button>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { CircleCheck, CircleClose, Delete, EditPen } from '@element-plus/icons-vue';
import type { OAuthApp } from '../types';

const props = defineProps<{
  app: OAuthApp;
}>();

defineEmits<{
  open: [app: OAuthApp];
  toggle: [app: OAuthApp];
  delete: [app: OAuthApp];
}>();

const visibleScopes = computed(() => props.app.scopes.slice(0, 3));
const hiddenScopeCount = computed(() => Math.max(props.app.scopes.length - visibleScopes.value.length, 0));
</script>

<style scoped>
.oauth-app-card {
  display: grid;
  gap: 16px;
  min-height: 260px;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 22px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.74), rgba(255, 255, 255, 0.44));
  box-shadow: 0 22px 48px rgba(84, 118, 148, 0.13), inset 0 1px 0 rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(18px);
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease;
}

.oauth-app-card:hover,
.oauth-app-card:focus-visible {
  border-color: rgba(125, 211, 252, 0.72);
  box-shadow: 0 26px 54px rgba(57, 135, 208, 0.16), inset 0 1px 0 rgba(255, 255, 255, 0.94);
  transform: translateY(-2px);
  outline: 0;
}

.card-main {
  display: flex;
  gap: 14px;
  min-width: 0;
}

.app-mark {
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  width: 46px;
  height: 46px;
  border-radius: 16px;
  background:
    radial-gradient(circle at 72% 20%, #31d5ef 0 20%, transparent 21%),
    linear-gradient(135deg, #1f6fe8 0%, #67d9ff 58%, #fff8fb 100%);
  box-shadow: 0 12px 26px rgba(44, 150, 226, 0.22), inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.app-mark span {
  width: 30px;
  height: 14px;
  border-radius: 999px 999px 10px 10px;
  background: rgba(255, 255, 255, 0.78);
  transform: translateY(5px);
}

.app-copy {
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-row h3 {
  margin: 0;
  overflow: hidden;
  color: #162336;
  font-size: 18px;
  font-weight: 860;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.system-badge {
  flex: 0 0 auto;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, #2176e8, #1bbce6);
  color: #fff;
  font-size: 13px;
  font-weight: 850;
  line-height: 28px;
}

.app-copy p {
  display: -webkit-box;
  margin: 8px 0 0;
  overflow: hidden;
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.scope-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.scope-chip {
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.56);
  color: #1e293b;
  font-size: 13px;
  line-height: 28px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.meta-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(120px, 0.46fr);
  gap: 10px;
}

.meta-grid div,
.callback-box {
  display: grid;
  gap: 5px;
  min-width: 0;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.38);
}

.meta-grid span,
.callback-box span {
  color: #7a8a9b;
  font-size: 12px;
  font-weight: 780;
}

.meta-grid strong,
.callback-box strong {
  overflow: hidden;
  color: #25364a;
  font-size: 13px;
  font-weight: 780;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid rgba(226, 236, 246, 0.88);
}

.status-button,
.icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  background: rgba(255, 255, 255, 0.5);
  color: #6b7b8f;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.84);
}

.status-button {
  gap: 8px;
  padding: 0 13px;
  border-radius: 13px;
  font-weight: 820;
}

.status-button.enabled {
  color: #218b45;
}

.icon-action {
  width: 38px;
  border-radius: 13px;
}

.icon-action.danger {
  margin-left: auto;
  color: #c8a5aa;
}

.icon-action.danger:not(:disabled) {
  color: #dc5266;
}

.icon-action:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.status-button svg,
.icon-action svg {
  width: 17px;
  height: 17px;
}
</style>
