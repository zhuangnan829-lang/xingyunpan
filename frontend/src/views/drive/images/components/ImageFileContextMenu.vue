<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="image-file-menu"
      :style="{ left: `${x}px`, top: `${y}px` }"
      role="menu"
      @click.stop
      @contextmenu.prevent
      @mouseleave="activeSubmenu = null"
    >
      <div class="menu-section">
        <button class="menu-item" type="button" @click="$emit('open')">
          <Open class="menu-icon" />
          <span>打开</span>
        </button>
        <button
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'openWith' }"
          type="button"
          @mouseenter="activeSubmenu = 'openWith'"
          @click.prevent
        >
          <Grid class="menu-icon" />
          <span>打开方式</span>
          <ArrowRight class="submenu-arrow" />
        </button>
        <button class="menu-item" type="button" @click="$emit('download')">
          <Download class="menu-icon" />
          <span>下载</span>
        </button>
      </div>

      <div class="menu-section">
        <button class="menu-item" type="button" @click="$emit('share')">
          <Share class="menu-icon" />
          <span>分享</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('rename')">
          <EditPen class="menu-icon" />
          <span>重命名</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('copy')">
          <CopyDocument class="menu-icon" />
          <span>复制</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('copy-link')">
          <Link class="menu-icon" />
          <span>获取直链</span>
        </button>
      </div>

      <div class="menu-section">
        <button
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'tags' }"
          type="button"
          @mouseenter="activeSubmenu = 'tags'"
          @click.prevent
        >
          <CollectionTag class="menu-icon" />
          <span>标签</span>
          <ArrowRight class="submenu-arrow" />
        </button>
        <button
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'organize' }"
          type="button"
          @mouseenter="activeSubmenu = 'organize'"
          @click.prevent
        >
          <FolderOpened class="menu-icon accent" />
          <span>整理</span>
          <ArrowRight class="submenu-arrow" />
        </button>
        <button
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'more' }"
          type="button"
          @mouseenter="activeSubmenu = 'more'"
          @click.prevent
        >
          <Tools class="menu-icon" />
          <span>更多操作</span>
          <ArrowRight class="submenu-arrow" />
        </button>
      </div>

      <div class="menu-section">
        <button class="menu-item" type="button" @click="$emit('locate')">
          <Folder class="menu-icon" />
          <span>转到所在目录</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('details')">
          <InfoFilled class="menu-icon" />
          <span>详细信息</span>
        </button>
      </div>

      <div class="menu-section danger-section">
        <button class="menu-item danger" type="button" @click="$emit('delete')">
          <Delete class="menu-icon" />
          <span>删除</span>
        </button>
      </div>

      <div v-if="activeSubmenu === 'openWith'" class="image-file-submenu open-with">
        <button class="menu-item" type="button" @click="$emit('open')">
          <Picture class="menu-icon app-icon image-viewer" />
          <span>图片查看器</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('open-photopea')">
          <span class="photopea-icon">P</span>
          <span>Photopea</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('choose-app')">
          <Grid class="menu-icon" />
          <span>选择应用...</span>
        </button>
      </div>

      <div v-if="activeSubmenu === 'tags'" class="image-file-submenu tags">
        <button class="menu-item" type="button" @click="$emit('manage-tags')">
          <CollectionTag class="menu-icon" />
          <span>管理标签</span>
        </button>
      </div>

      <div v-if="activeSubmenu === 'organize'" class="image-file-submenu organize">
        <button class="menu-item" type="button" @click="$emit('move')">
          <FolderOpened class="menu-icon" />
          <span>移动</span>
        </button>
        <button class="menu-item" type="button" @click="$emit('custom-icon')">
          <PictureRounded class="menu-icon" />
          <span>自定义图标</span>
        </button>
      </div>

      <div v-if="activeSubmenu === 'more'" class="image-file-submenu more">
        <button class="menu-item" type="button" @click="$emit('details')">
          <InfoFilled class="menu-icon" />
          <span>查看属性</span>
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import {
  ArrowRight,
  CollectionTag,
  CopyDocument,
  Delete,
  Download,
  EditPen,
  Folder,
  FolderOpened,
  Grid,
  InfoFilled,
  Link,
  Open,
  Picture,
  PictureRounded,
  Share,
  Tools,
} from '@element-plus/icons-vue';

const props = defineProps<{
  visible: boolean;
  x: number;
  y: number;
}>();

defineEmits<{
  (e: 'open'): void;
  (e: 'open-photopea'): void;
  (e: 'choose-app'): void;
  (e: 'download'): void;
  (e: 'share'): void;
  (e: 'rename'): void;
  (e: 'copy'): void;
  (e: 'copy-link'): void;
  (e: 'manage-tags'): void;
  (e: 'move'): void;
  (e: 'custom-icon'): void;
  (e: 'locate'): void;
  (e: 'details'): void;
  (e: 'delete'): void;
}>();

const activeSubmenu = ref<'openWith' | 'tags' | 'organize' | 'more' | null>(null);

watch(
  () => props.visible,
  (visible) => {
    if (!visible) activeSubmenu.value = null;
  },
);
</script>

<style scoped>
.image-file-menu {
  position: fixed;
  z-index: 3400;
  width: 242px;
  padding: 8px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 14px;
  background:
    radial-gradient(circle at 8% 0%, rgba(118, 196, 255, 0.16), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.96), rgba(247, 252, 255, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 38px rgba(31, 45, 72, 0.18);
  color: #20242c;
  backdrop-filter: blur(20px) saturate(1.18);
}

.image-file-submenu {
  position: absolute;
  left: calc(100% + 8px);
  z-index: 3401;
  width: 214px;
  padding: 8px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 14px;
  background:
    radial-gradient(circle at 8% 0%, rgba(118, 196, 255, 0.16), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.96), rgba(247, 252, 255, 0.9));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 14px 30px rgba(31, 45, 72, 0.18);
  backdrop-filter: blur(20px) saturate(1.18);
}

.image-file-submenu.open-with {
  top: 42px;
}

.image-file-submenu.tags {
  top: 278px;
}

.image-file-submenu.organize {
  top: 322px;
}

.image-file-submenu.more {
  top: 366px;
}

.menu-section {
  padding: 4px 0;
  border-bottom: 1px solid rgba(213, 222, 233, 0.82);
}

.menu-section:last-of-type {
  border-bottom: 0;
}

.menu-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 14px;
  align-items: center;
  width: 100%;
  min-height: 38px;
  padding: 0 10px;
  border: 0;
  border-radius: 10px;
  color: #20242c;
  background: transparent;
  font-size: 15px;
  font-weight: 650;
  text-align: left;
  cursor: pointer;
}

.menu-item:hover,
.menu-item.active {
  background: rgba(15, 23, 42, 0.06);
}

.menu-item.danger {
  color: #1f2937;
}

.menu-icon {
  width: 21px;
  height: 21px;
  color: #767d86;
}

.menu-icon.accent,
.menu-item:hover .menu-icon,
.menu-item.active .menu-icon {
  color: #1677ff;
}

.submenu-arrow {
  width: 14px;
  height: 14px;
  justify-self: end;
  color: #686c72;
}

.photopea-icon,
.app-icon {
  display: inline-grid;
  place-items: center;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  color: #fff;
  font-size: 13px;
  font-weight: 900;
}

.photopea-icon {
  background: linear-gradient(135deg, #10c9a7, #168bff);
}

.image-viewer {
  color: #ef3340;
}
</style>
