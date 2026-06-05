<template>
  <div class="permission-list">
    <article v-for="permission in permissions" :key="permission.id" class="permission-row" :class="{ disabled: permission.required }">
      <label class="check-cell">
        <input v-model="permission.enabled" type="checkbox" :disabled="permission.required" />
        <span></span>
      </label>

      <component :is="iconMap[permission.icon]" class="permission-icon" />

      <div class="permission-copy">
        <div class="permission-title">
          <strong>{{ permission.title }}</strong>
          <span class="scope-chip">{{ permission.scope }}</span>
          <span v-if="permission.required" class="required-chip">必需</span>
        </div>
        <p>{{ permission.description }}</p>
      </div>

      <div v-if="permission.modeSwitch" class="mode-switch" aria-label="权限模式">
        <button type="button" :class="{ active: permission.mode === 'read' }" @click="permission.mode = 'read'">只读</button>
        <button type="button" :class="{ active: permission.mode === 'write' }" @click="permission.mode = 'write'">读写</button>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { Connection, Document, FolderOpened, Goods, Lock, Money, Share, Tools, User, Wallet } from '@element-plus/icons-vue';
import type { OAuthPermission } from '../types';

defineProps<{
  permissions: OAuthPermission[];
}>();

const iconMap = {
  document: Document,
  offline: Connection,
  user: User,
  lock: Lock,
  folder: FolderOpened,
  share: Share,
  task: Goods,
  finance: Money,
  dav: Wallet,
  admin: Tools,
};
</script>

<style scoped>
.permission-list {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.44);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(16px);
}

.permission-row {
  display: grid;
  grid-template-columns: 28px 34px minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  min-height: 86px;
  padding: 18px 22px;
  border-bottom: 1px solid rgba(220, 231, 244, 0.86);
}

.permission-row:last-child {
  border-bottom: 0;
}

.permission-row.disabled {
  opacity: 0.52;
}

.check-cell {
  position: relative;
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
}

.check-cell input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.check-cell input:disabled {
  cursor: not-allowed;
}

.check-cell span {
  width: 22px;
  height: 22px;
  border: 2px solid #7a8794;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.82);
  box-sizing: border-box;
}

.check-cell input:checked + span {
  border-color: #2485dd;
  background: #2485dd;
  box-shadow: 0 8px 18px rgba(36, 133, 221, 0.2);
}

.check-cell input:checked + span::after {
  content: '';
  display: block;
  width: 9px;
  height: 5px;
  margin: 5px 0 0 4px;
  border-left: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(-45deg);
}

.permission-icon {
  width: 22px;
  height: 22px;
  color: #758391;
}

.permission-copy {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.permission-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.permission-title strong {
  color: #172033;
  font-size: 16px;
  font-weight: 840;
}

.scope-chip,
.required-chip {
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 760;
  line-height: 24px;
}

.scope-chip {
  border: 1px solid rgba(204, 213, 224, 0.9);
  background: rgba(255, 255, 255, 0.74);
  color: #344154;
}

.required-chip {
  background: linear-gradient(135deg, #9ccff5, #6eb6ec);
  color: #fff;
}

.permission-copy p {
  margin: 0;
  color: #657487;
  font-size: 13px;
  line-height: 1.6;
}

.mode-switch {
  display: inline-flex;
  overflow: hidden;
  border: 1px solid rgba(255, 123, 44, 0.72);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
}

.mode-switch button {
  min-width: 58px;
  min-height: 34px;
  border: 0;
  background: transparent;
  color: #64748b;
  font-weight: 800;
  cursor: pointer;
}

.mode-switch button.active {
  background: rgba(255, 122, 38, 0.12);
  color: #f26b1d;
}

@media (max-width: 860px) {
  .permission-row {
    grid-template-columns: 28px 30px minmax(0, 1fr);
  }

  .mode-switch {
    grid-column: 3;
    width: max-content;
  }
}
</style>
