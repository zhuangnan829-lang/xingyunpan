<template>
  <div class="progress-bar">
    <el-progress
      :percentage="progress"
      :status="progressStatus"
      :stroke-width="8"
      :show-text="true"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  progress: number;  // 0-100
  status: 'uploading' | 'completed' | 'failed' | 'pending' | 'hashing' | 'cancelled';
}

const props = defineProps<Props>();

// Map status to Element Plus progress status
const progressStatus = computed(() => {
  switch (props.status) {
    case 'completed':
      return 'success';
    case 'failed':
    case 'cancelled':
      return 'exception';
    case 'uploading':
    case 'hashing':
      return undefined; // Default blue color
    case 'pending':
      return undefined;
    default:
      return undefined;
  }
});
</script>

<style scoped>
.progress-bar {
  width: 100%;
}
</style>
