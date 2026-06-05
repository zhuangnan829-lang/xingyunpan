/**
 * 错误处理 Composable
 * 提供统一的错误处理和消息显示
 */

import { ElMessage } from 'element-plus';
import { handleAPIError } from '@/api/request';
import { errorLogger } from '@/utils/error-logger';

export function useErrorHandler() {
  /**
   * 处理错误并显示消息
   */
  const handleError = (error: any, context?: string): void => {
    const message = error instanceof Error ? error.message : handleAPIError(error);
    
    // 记录错误日志
    errorLogger.error(
      context ? `${context}: ${message}` : message,
      error instanceof Error ? error : undefined,
      { context }
    );
    
    // 显示错误消息
    ElMessage.error({
      message,
      duration: 3000,
      showClose: true,
    });
  };

  /**
   * 显示成功消息
   */
  const showSuccess = (message: string): void => {
    ElMessage.success({
      message,
      duration: 2000,
      showClose: true,
    });
  };

  /**
   * 显示警告消息
   */
  const showWarning = (message: string): void => {
    ElMessage.warning({
      message,
      duration: 3000,
      showClose: true,
    });
  };

  /**
   * 显示信息消息
   */
  const showInfo = (message: string): void => {
    ElMessage.info({
      message,
      duration: 2000,
      showClose: true,
    });
  };

  /**
   * 包装异步操作，自动处理错误
   */
  const withErrorHandling = async <T>(
    operation: () => Promise<T>,
    options?: {
      context?: string;
      successMessage?: string;
      onError?: (error: any) => void;
    }
  ): Promise<T | undefined> => {
    try {
      const result = await operation();
      
      if (options?.successMessage) {
        showSuccess(options.successMessage);
      }
      
      return result;
    } catch (error) {
      handleError(error, options?.context);
      
      if (options?.onError) {
        options.onError(error);
      }
      
      return undefined;
    }
  };

  return {
    handleError,
    showSuccess,
    showWarning,
    showInfo,
    withErrorHandling,
  };
}
