<template>
  <div class="breadcrumb-container">
    <el-breadcrumb separator="/">
      <!-- Root directory -->
      <el-breadcrumb-item>
        <a @click.prevent="handleNavigate(null)">我的文件</a>
      </el-breadcrumb-item>

      <!-- Path segments with ellipsis for long paths -->
      <template v-if="displayPath.length > 0">
        <!-- Show ellipsis if path is truncated -->
        <el-breadcrumb-item v-if="showEllipsis">
          <span class="ellipsis">...</span>
        </el-breadcrumb-item>

        <!-- Display visible path segments -->
        <el-breadcrumb-item
          v-for="item in displayPath"
          :key="item.id"
        >
          <a @click.prevent="handleNavigate(item.id)">{{ item.name }}</a>
        </el-breadcrumb-item>
      </template>
    </el-breadcrumb>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { FolderPathItem } from '@/types/folder';

interface Props {
  path?: FolderPathItem[];
}

interface Emits {
  (e: 'navigate', folderId: number | null): void;
}

const props = withDefaults(defineProps<Props>(), {
  path: () => []
});

const emit = defineEmits<Emits>();

// Maximum number of path segments to display before truncating
const MAX_VISIBLE_SEGMENTS = 4;

// Compute whether to show ellipsis
const showEllipsis = computed(() => {
  return props.path.length > MAX_VISIBLE_SEGMENTS;
});

// Compute the visible path segments
const displayPath = computed(() => {
  if (props.path.length <= MAX_VISIBLE_SEGMENTS) {
    return props.path;
  }

  // Show first segment and last 2 segments with ellipsis in between
  const firstSegment = props.path[0];
  const lastSegments = props.path.slice(-2);

  return [firstSegment, ...lastSegments];
});

// Handle navigation click
const handleNavigate = (folderId: number | null) => {
  emit('navigate', folderId);
};
</script>

<style scoped>
.breadcrumb-container {
  padding: 12px 0;
  background-color: #fff;
}

.ellipsis {
  color: #909399;
  cursor: default;
  user-select: none;
}

:deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #303133;
  font-weight: 500;
}

:deep(.el-breadcrumb__item a) {
  color: #409eff;
  transition: color 0.3s;
}

:deep(.el-breadcrumb__item a:hover) {
  color: #66b1ff;
}
</style>
