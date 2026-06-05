<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="image-context-menu"
      :style="{ left: `${x}px`, top: `${y}px` }"
      role="menu"
      tabindex="-1"
      @click.stop
      @contextmenu.prevent
      @mouseleave="drawioSubmenuVisible = false"
    >
      <div class="context-menu-section">
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('upload-file')">
          <Upload class="context-menu-icon" />
          <span>上传文件</span>
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('upload-folder')">
          <FolderAdd class="context-menu-icon" />
          <span>上传目录</span>
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('upload-clipboard')">
          <CopyDocument class="context-menu-icon" />
          <span>从剪贴板上传</span>
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('offline-download')">
          <Download class="context-menu-icon" />
          <span>离线下载</span>
        </button>
      </div>

      <div class="context-menu-section">
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-folder')">
          <FolderAdd class="context-menu-icon" />
          <span>创建文件夹</span>
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'file')">
          <DocumentAdd class="context-menu-icon" />
          <span>创建文件</span>
        </button>
      </div>

      <div class="context-menu-section">
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'markdown')">
          <span class="context-app-icon markdown">MD</span>
          <span>Markdown (.md)</span>
        </button>
        <button
          class="context-menu-item has-submenu"
          :class="{ active: drawioSubmenuVisible }"
          type="button"
          role="menuitem"
          @mouseenter="drawioSubmenuVisible = true"
          @click.prevent
        >
          <span class="context-app-icon drawio">IO</span>
          <span>draw.io</span>
          <ArrowRight class="context-submenu-arrow" />
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'text')">
          <Tickets class="context-app-icon text" />
          <span>文本 (.txt)</span>
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'excalidraw')">
          <EditPen class="context-app-icon excalidraw" />
          <span>Excalidraw (.excalidraw)</span>
        </button>
      </div>

      <div class="context-menu-section">
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('refresh')">
          <RefreshRight class="context-menu-icon" />
          <span>刷新</span>
        </button>
      </div>

      <div v-if="drawioSubmenuVisible" class="image-context-submenu" @mouseleave="drawioSubmenuVisible = false">
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'drawio')">
          图表 (.drawio)
        </button>
        <button class="context-menu-item" type="button" role="menuitem" @click="$emit('create-file', 'dwb')">
          白板 (.dwb)
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import {
  ArrowRight,
  CopyDocument,
  DocumentAdd,
  Download,
  EditPen,
  FolderAdd,
  RefreshRight,
  Tickets,
  Upload,
} from '@element-plus/icons-vue';
import type { WorkspaceFileKind } from '../useImagesWorkspace';

const props = defineProps<{
  visible: boolean;
  x: number;
  y: number;
}>();

defineEmits<{
  (e: 'upload-file'): void;
  (e: 'upload-folder'): void;
  (e: 'upload-clipboard'): void;
  (e: 'offline-download'): void;
  (e: 'create-folder'): void;
  (e: 'create-file', kind: WorkspaceFileKind): void;
  (e: 'refresh'): void;
}>();

const drawioSubmenuVisible = ref(false);

watch(
  () => props.visible,
  (visible) => {
    if (!visible) drawioSubmenuVisible.value = false;
  },
);
</script>

<style scoped>
.image-context-menu,
.image-context-submenu {
  position: fixed;
  z-index: 3200;
  width: 252px;
  padding: 6px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background:
    radial-gradient(circle at 8% 0%, rgba(118, 196, 255, 0.22), transparent 34%),
    radial-gradient(circle at 100% 20%, rgba(255, 196, 221, 0.2), transparent 32%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.9), rgba(246, 252, 255, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    inset 0 -1px 0 rgba(197, 224, 247, 0.5),
    0 18px 42px rgba(63, 114, 174, 0.18),
    0 8px 20px rgba(255, 168, 205, 0.08);
  color: #23272f;
  font-family: "Microsoft YaHei", "PingFang SC", system-ui, sans-serif;
  backdrop-filter: blur(22px) saturate(1.25);
}

.image-context-submenu {
  position: absolute;
  left: calc(100% + 8px);
  top: 242px;
  width: 172px;
}

.context-menu-section {
  padding: 4px 0;
  border-bottom: 1px solid rgba(206, 221, 238, 0.72);
}

.context-menu-section:last-of-type {
  border-bottom: 0;
}

.context-menu-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 14px;
  align-items: center;
  width: 100%;
  min-height: 34px;
  padding: 0 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #22304a;
  font-size: 14px;
  font-weight: 720;
  line-height: 1.2;
  text-align: left;
  cursor: pointer;
  transition:
    background 0.14s ease,
    box-shadow 0.14s ease,
    transform 0.14s ease;
}

.context-menu-item span {
  min-width: 0;
}

.context-menu-item:hover,
.context-menu-item.active {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.82), rgba(232, 245, 255, 0.68)),
    rgba(255, 255, 255, 0.52);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 8px 18px rgba(70, 136, 210, 0.12);
  transform: translateX(1px);
}

.context-menu-icon,
.context-app-icon {
  width: 20px;
  height: 20px;
  color: #6f7b8d;
}

.context-app-icon {
  display: inline-grid;
  place-items: center;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 900;
  line-height: 1;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.32),
    0 5px 10px rgba(55, 89, 130, 0.12);
}

.context-app-icon.markdown {
  background: #3f3f46;
  color: #fff;
}

.context-app-icon.drawio {
  background: linear-gradient(135deg, #ff9c1a, #f06c00);
  color: #fff;
}

.context-app-icon.text {
  color: #1682c5;
}

.context-app-icon.excalidraw {
  color: #6957d9;
}

.context-submenu-arrow {
  width: 14px;
  height: 14px;
  justify-self: end;
  color: #686c72;
}

.image-context-submenu .context-menu-item {
  grid-template-columns: minmax(0, 1fr);
  padding: 0 14px;
}
</style>
