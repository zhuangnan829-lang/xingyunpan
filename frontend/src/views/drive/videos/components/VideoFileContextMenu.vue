<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="video-file-menu"
      :style="{ left: `${x}px`, top: `${y}px` }"
      role="menu"
      @click.stop
      @contextmenu.prevent
    >
      <span class="menu-arrow" aria-hidden="true"></span>
      <button class="menu-item" type="button" @click="$emit('rename')">
        <EditPen class="menu-icon" />
        <span>重命名</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('download')">
        <Download class="menu-icon" />
        <span>下载</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('share')">
        <Share class="menu-icon" />
        <span>分享</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('history')">
        <Clock class="menu-icon" />
        <span>版本历史</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('collaboration')">
        <User class="menu-icon" />
        <span>协作管理</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('move')">
        <Folder class="menu-icon" />
        <span>移动到</span>
      </button>
      <button class="menu-item" type="button" @click="$emit('copy')">
        <CopyDocument class="menu-icon" />
        <span>复制到</span>
      </button>
      <div class="menu-separator"></div>
      <button class="menu-item danger" type="button" @click="$emit('delete')">
        <Delete class="menu-icon" />
        <span>删除</span>
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { Clock, CopyDocument, Delete, Download, EditPen, Folder, Share, User } from '@element-plus/icons-vue';

defineProps<{
  visible: boolean;
  x: number;
  y: number;
}>();

defineEmits<{
  (e: 'rename'): void;
  (e: 'download'): void;
  (e: 'share'): void;
  (e: 'history'): void;
  (e: 'collaboration'): void;
  (e: 'move'): void;
  (e: 'copy'): void;
  (e: 'delete'): void;
}>();
</script>

<style scoped>
.video-file-menu {
  position: fixed;
  z-index: 3400;
  width: 188px;
  padding: 18px 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 22px;
  background:
    radial-gradient(circle at 0% 0%, rgba(135, 214, 255, 0.18), transparent 38%),
    radial-gradient(circle at 100% 8%, rgba(255, 203, 225, 0.18), transparent 34%),
    rgba(255, 255, 255, 0.84);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 28px 62px rgba(63, 108, 168, 0.22);
  color: #1a2944;
  backdrop-filter: blur(24px) saturate(1.2);
}

.menu-arrow {
  position: absolute;
  top: -8px;
  left: 50%;
  width: 16px;
  height: 16px;
  border-left: 1px solid rgba(255, 255, 255, 0.82);
  border-top: 1px solid rgba(255, 255, 255, 0.82);
  background: rgba(255, 255, 255, 0.86);
  transform: translateX(-50%) rotate(45deg);
}

.menu-item {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  align-items: center;
  width: 100%;
  min-height: 42px;
  padding: 0 10px;
  border: 0;
  border-radius: 12px;
  background: transparent;
  color: #1a2944;
  font-size: 16px;
  font-weight: 840;
  text-align: left;
  cursor: pointer;
  transition:
    background 0.16s ease,
    color 0.16s ease,
    transform 0.16s ease;
}

.menu-item:hover {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.78), rgba(228, 244, 255, 0.72)),
    rgba(255, 255, 255, 0.54);
  color: #1d70da;
  transform: translateX(1px);
}

.menu-item.danger:hover {
  color: #ef4444;
}

.menu-icon {
  width: 18px;
  height: 18px;
  color: currentColor;
}

.menu-separator {
  height: 1px;
  margin: 12px 0;
  background: rgba(199, 217, 237, 0.82);
}
</style>
