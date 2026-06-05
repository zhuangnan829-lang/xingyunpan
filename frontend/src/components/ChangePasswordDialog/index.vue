<template>
  <el-dialog
    v-model="dialogVisible"
    title="修改密码"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
      @submit.prevent="handleSubmit"
    >
      <el-form-item label="旧密码" prop="old_password">
        <el-input
          v-model="formData.old_password"
          type="password"
          placeholder="请输入旧密码"
          show-password
          clearable
        />
      </el-form-item>

      <el-form-item label="新密码" prop="new_password">
        <el-input
          v-model="formData.new_password"
          type="password"
          placeholder="请输入新密码（至少 6 位）"
          show-password
          clearable
        />
      </el-form-item>

      <el-form-item label="确认密码" prop="confirm_password">
        <el-input
          v-model="formData.confirm_password"
          type="password"
          placeholder="请再次输入新密码"
          show-password
          clearable
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { useUserStore } from '@/stores/user';
import type { ChangePasswordRequest } from '@/types/user';

interface Props {
  visible: boolean;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const userStore = useUserStore();

const formRef = ref<FormInstance>();
const loading = ref(false);
const dialogVisible = ref(props.visible);

const formData = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
});

watch(
  () => props.visible,
  (newVal) => {
    dialogVisible.value = newVal;
  },
);

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
});

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!value) {
    callback(new Error('请再次输入新密码'));
    return;
  }
  if (value !== formData.new_password) {
    callback(new Error('两次输入的密码不一致'));
    return;
  }
  callback();
};

const formRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
};

async function handleSubmit(): Promise<void> {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    loading.value = true;

    const requestData: ChangePasswordRequest = {
      old_password: formData.old_password,
      new_password: formData.new_password,
    };

    await userStore.changePassword(requestData);
    ElMessage.success('密码修改成功，请重新登录');
    handleClose();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改密码失败，请重试');
    console.error('Failed to change password:', error);
  } finally {
    loading.value = false;
  }
}

function handleClose(): void {
  formRef.value?.resetFields();
  formData.old_password = '';
  formData.new_password = '';
  formData.confirm_password = '';
  dialogVisible.value = false;
}
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
