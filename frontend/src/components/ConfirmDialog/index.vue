<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    width="450px"
    @close="handleClose"
  >
    <div class="confirm-content">
      <el-icon :size="48" :color="iconColor" class="confirm-icon">
        <component :is="iconComponent" />
      </el-icon>
      <div class="confirm-message">
        {{ message }}
      </div>
      <div v-if="subMessage" class="confirm-sub-message">
        {{ subMessage }}
      </div>
    </div>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button :type="confirmType" :loading="loading" @click="handleConfirm">
        {{ confirmText }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { WarningFilled, QuestionFilled, InfoFilled } from '@element-plus/icons-vue';

interface Props {
  visible: boolean;
  title?: string;
  message: string;
  subMessage?: string;
  type?: 'warning' | 'danger' | 'info';
  confirmText?: string;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '确认操作',
  subMessage: '',
  type: 'warning',
  confirmText: '确定'
});

const emit = defineEmits<Emits>();

const loading = ref(false);
const dialogVisible = ref(props.visible);

// Computed properties
const iconComponent = computed(() => {
  switch (props.type) {
    case 'danger':
      return WarningFilled;
    case 'info':
      return InfoFilled;
    default:
      return QuestionFilled;
  }
});

const iconColor = computed(() => {
  switch (props.type) {
    case 'danger':
      return '#f56c6c';
    case 'info':
      return '#409eff';
    default:
      return '#e6a23c';
  }
});

const confirmType = computed(() => {
  switch (props.type) {
    case 'danger':
      return 'danger';
    case 'info':
      return 'primary';
    default:
      return 'warning';
  }
});

// Watch for visibility changes
watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
});

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
  if (!newVal) {
    loading.value = false;
  }
});

// Handle close
const handleClose = () => {
  dialogVisible.value = false;
  emit('cancel');
};

// Handle confirm
const handleConfirm = () => {
  loading.value = true;
  emit('confirm');
  
  // Note: Parent component should handle success/error and close dialog
};
</script>

<style scoped>
.confirm-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
}

.confirm-icon {
  margin-bottom: 16px;
}

.confirm-message {
  font-size: 16px;
  color: #303133;
  text-align: center;
  margin-bottom: 8px;
}

.confirm-sub-message {
  font-size: 14px;
  color: #909399;
  text-align: center;
}

:deep(.el-dialog__body) {
  padding: 20px 20px 10px;
}
</style>
