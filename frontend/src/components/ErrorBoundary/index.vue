<template>
  <slot v-if="!hasError" />
  <div v-else class="error-boundary">
    <el-result
      icon="error"
      title="页面出现错误"
      :sub-title="errorMessage"
    >
      <template #extra>
        <el-button type="primary" @click="reset">重新加载</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue';

const hasError = ref(false);
const errorMessage = ref('发生了意外错误，请尝试刷新页面');

onErrorCaptured((error: Error) => {
  hasError.value = true;
  errorMessage.value = error.message || '发生了意外错误，请尝试刷新页面';
  console.error('[ErrorBoundary] Caught error:', error);
  // Return false to prevent the error from propagating further
  return false;
});

function reset(): void {
  hasError.value = false;
  errorMessage.value = '发生了意外错误，请尝试刷新页面';
}
</script>

<style scoped>
.error-boundary {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  padding: 40px;
}
</style>
