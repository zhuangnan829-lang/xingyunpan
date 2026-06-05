<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="document-file-menu"
      :style="{ left: `${x}px`, top: `${y}px` }"
      @click.stop
      @contextmenu.prevent.stop
    >
      <button type="button" @click="$emit('open')">
        <el-icon><View /></el-icon>
        打开预览
      </button>
      <button type="button" @click="$emit('rename')">
        <el-icon><EditPen /></el-icon>
        重命名
      </button>
      <button type="button" @click="$emit('download')">
        <el-icon><Download /></el-icon>
        下载
      </button>
      <button type="button" @click="$emit('share')">
        <el-icon><Share /></el-icon>
        分享
      </button>
      <button type="button" @click="$emit('history')">
        <el-icon><Clock /></el-icon>
        版本历史
      </button>
      <button type="button" @click="$emit('collaboration')">
        <el-icon><User /></el-icon>
        协作管理
      </button>
      <button type="button" @click="$emit('move')">
        <el-icon><Folder /></el-icon>
        移动到
      </button>
      <button type="button" @click="$emit('copy')">
        <el-icon><CopyDocument /></el-icon>
        复制到
      </button>
      <hr />
      <button type="button" class="danger" @click="$emit('delete')">
        <el-icon><Delete /></el-icon>
        删除
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { Clock, CopyDocument, Delete, Download, EditPen, Folder, Share, User, View } from '@element-plus/icons-vue';

defineProps<{
  visible: boolean;
  x: number;
  y: number;
}>();

defineEmits<{
  (e: 'open'): void;
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
.document-file-menu {
  position: fixed;
  z-index: 3200;
  display: grid;
  width: 196px;
  padding: 14px 12px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 8% 0%, rgba(186, 230, 253, 0.58), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 44%),
    rgba(255, 255, 255, 0.84);
  box-shadow: 0 22px 58px rgba(92, 120, 166, 0.26);
  backdrop-filter: blur(24px);
}

button {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 42px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #172642;
  font-size: 15px;
  font-weight: 780;
  cursor: pointer;
  text-align: left;
}

button:hover {
  background: rgba(255, 255, 255, 0.72);
  color: #2f7df5;
}

.el-icon {
  width: 18px;
  height: 18px;
}

hr {
  width: 100%;
  height: 1px;
  margin: 10px 0;
  border: 0;
  background: rgba(170, 190, 215, 0.48);
}

.danger:hover {
  background: rgba(255, 240, 240, 0.86);
  color: #ef4444;
}
</style>
