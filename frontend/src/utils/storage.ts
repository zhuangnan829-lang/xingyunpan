/**
 * localStorage 工具函数
 * 提供带错误处理和配额检测的 localStorage 操作
 */

export type StorageErrorType = 'quota_exceeded' | 'parse_error' | 'write_error' | 'read_error';

export interface StorageError {
  type: StorageErrorType;
  key: string;
  message: string;
}

/** Registered error handler for storage errors */
let errorHandler: ((error: StorageError) => void) | null = null;

/**
 * Register a global storage error handler (e.g., to show Toast notifications)
 */
export function registerStorageErrorHandler(handler: (error: StorageError) => void): void {
  errorHandler = handler;
}

function notifyError(error: StorageError): void {
  if (error.type === 'quota_exceeded') {
    console.warn(`[Storage] ${error.type} for key "${error.key}": ${error.message}`);
  } else {
    console.error(`[Storage] ${error.type} for key "${error.key}": ${error.message}`);
  }
  errorHandler?.(error);
}

/**
 * Safely write a value to localStorage
 * Handles QuotaExceededError with a user-friendly fallback
 * @returns true on success, false on failure
 */
export function safeSetItem(key: string, value: string): boolean {
  try {
    localStorage.setItem(key, value);
    return true;
  } catch (error: any) {
    if (isQuotaExceeded(error)) {
      notifyError({
        type: 'quota_exceeded',
        key,
        message: 'localStorage 空间不足，数据可能未保存。请清理回收站或删除旧的分享链接以释放空间。',
      });
    } else {
      notifyError({
        type: 'write_error',
        key,
        message: error?.message ?? '写入失败',
      });
    }
    return false;
  }
}

/**
 * Safely read a value from localStorage
 * @returns The stored string, or null if not found or on error
 */
export function safeGetItem(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch (error: any) {
    notifyError({
      type: 'read_error',
      key,
      message: error?.message ?? '读取失败',
    });
    return null;
  }
}

/**
 * Safely parse JSON from localStorage
 * @returns Parsed value, or null on parse error
 */
export function safeGetJSON<T>(key: string): T | null {
  const raw = safeGetItem(key);
  if (raw === null) return null;
  try {
    return JSON.parse(raw) as T;
  } catch (error: any) {
    notifyError({
      type: 'parse_error',
      key,
      message: `JSON 解析失败: ${error?.message ?? '未知错误'}`,
    });
    return null;
  }
}

/**
 * Safely serialize and write JSON to localStorage
 * @returns true on success, false on failure
 */
export function safeSetJSON(key: string, value: unknown): boolean {
  try {
    return safeSetItem(key, JSON.stringify(value));
  } catch (error: any) {
    notifyError({
      type: 'write_error',
      key,
      message: `JSON 序列化失败: ${error?.message ?? '未知错误'}`,
    });
    return false;
  }
}

/**
 * Check if an error is a QuotaExceededError
 */
function isQuotaExceeded(error: any): boolean {
  return (
    error?.name === 'QuotaExceededError' ||
    error?.name === 'NS_ERROR_DOM_QUOTA_REACHED' ||
    error?.code === 22 ||
    error?.code === 1014
  );
}

/**
 * Get approximate localStorage usage in bytes
 */
export function getStorageUsage(): number {
  let total = 0;
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key) {
      total += key.length + (localStorage.getItem(key)?.length ?? 0);
    }
  }
  return total * 2; // UTF-16 encoding: 2 bytes per char
}
