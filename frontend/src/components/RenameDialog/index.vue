<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isFolder ? '重命名文件夹' : '重命名文件'"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="80px"
    >
      <el-form-item label="新名称" prop="name">
        <el-input
          v-model="formData.name"
          placeholder="请输入新名称"
          @keyup.enter="handleSubmit"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import type { FileItem } from '@/types/file';
import type { FolderItem } from '@/types/folder';

interface Props {
  visible: boolean;
  item?: FileItem | FolderItem | null;
  isFolder?: boolean;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'confirm', name: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  item: null,
  isFolder: false
});

const emit = defineEmits<Emits>();

const formRef = ref<FormInstance>();
const loading = ref(false);
const dialogVisible = ref(props.visible);

const formData = reactive({
  name: ''
});

// Validation rules
const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { min: 1, max: 255, message: '名称长度应在 1 到 255 个字符之间', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        // Check for invalid characters
        const invalidChars = /[<>:"/\\|?*]/;
        if (invalidChars.test(value)) {
          callback(new Error('名称不能包含以下字符: < > : " / \\ | ? *'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

// Watch for visibility changes
watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (newVal && props.item) {
    // Pre-fill with current name
    formData.name = props.item.name;
  }
});

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
});

// Handle close
const handleClose = () => {
  dialogVisible.value = false;
  formRef.value?.resetFields();
};

// Handle submit
const handleSubmit = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    
    // Check if name is unchanged
    if (props.item && formData.name === props.item.name) {
      handleClose();
      return;
    }

    loading.value = true;
    emit('confirm', formData.name);
    
    // Note: Parent component should handle success/error and close dialog
  } catch (error) {
    // Validation failed
    console.error('Validation failed:', error);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
:deep(.el-dialog__body) {
  padding: 20px 20px 10px;
}
</style>
