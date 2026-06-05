<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="my-share-blank-menu"
      :style="{ left: `${x}px`, top: `${y}px` }"
      @click.stop
      @contextmenu.prevent.stop
    >
      <button type="button" @click="$emit('refresh')">
        <el-icon><Refresh /></el-icon>
        刷新
      </button>
      <button type="button" @click="$emit('select-all')">
        <el-icon><Grid /></el-icon>
        全选
      </button>
      <button type="button" @click="$emit('clear')">
        <el-icon><Close /></el-icon>
        取消选择
      </button>
      <button type="button" @click="$emit('invert')">
        <el-icon><Operation /></el-icon>
        反选
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { Close, Grid, Operation, Refresh } from '@element-plus/icons-vue';

defineProps<{
  visible: boolean;
  x: number;
  y: number;
}>();

defineEmits<{
  (e: 'refresh'): void;
  (e: 'select-all'): void;
  (e: 'clear'): void;
  (e: 'invert'): void;
}>();
</script>

<style scoped>
.my-share-blank-menu {
  position: fixed;
  z-index: 5200;
  display: grid;
  width: 210px;
  padding: 14px 12px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 8% 0%, rgba(186, 230, 253, 0.58), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 44%),
    rgba(255, 255, 255, 0.86);
  box-shadow: 0 22px 58px rgba(92, 120, 166, 0.28);
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
</style>
