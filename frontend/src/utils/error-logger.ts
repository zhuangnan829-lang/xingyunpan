/**
 * 错误日志记录工具
 * 用于统一记录和追踪应用中的错误
 */

export interface ErrorLog {
  timestamp: string;
  level: 'error' | 'warn' | 'info';
  message: string;
  stack?: string;
  context?: Record<string, any>;
}

class ErrorLogger {
  private logs: ErrorLog[] = [];
  private maxLogs: number = 100; // 最多保存 100 条日志

  /**
   * 记录错误
   */
  error(message: string, error?: Error, context?: Record<string, any>): void {
    const log: ErrorLog = {
      timestamp: new Date().toISOString(),
      level: 'error',
      message,
      stack: error?.stack,
      context,
    };

    this.addLog(log);
    
    // 在开发环境下输出到控制台
    if (import.meta.env.DEV) {
      console.error('[Error]', message, error, context);
    }
  }

  /**
   * 记录警告
   */
  warn(message: string, context?: Record<string, any>): void {
    const log: ErrorLog = {
      timestamp: new Date().toISOString(),
      level: 'warn',
      message,
      context,
    };

    this.addLog(log);
    
    if (import.meta.env.DEV) {
      console.warn('[Warning]', message, context);
    }
  }

  /**
   * 记录信息
   */
  info(message: string, context?: Record<string, any>): void {
    const log: ErrorLog = {
      timestamp: new Date().toISOString(),
      level: 'info',
      message,
      context,
    };

    this.addLog(log);
    
    if (import.meta.env.DEV) {
      console.info('[Info]', message, context);
    }
  }

  /**
   * 添加日志到队列
   */
  private addLog(log: ErrorLog): void {
    this.logs.push(log);
    
    // 保持日志数量在限制内
    if (this.logs.length > this.maxLogs) {
      this.logs.shift();
    }
  }

  /**
   * 获取所有日志
   */
  getLogs(): ErrorLog[] {
    return [...this.logs];
  }

  /**
   * 获取指定级别的日志
   */
  getLogsByLevel(level: 'error' | 'warn' | 'info'): ErrorLog[] {
    return this.logs.filter(log => log.level === level);
  }

  /**
   * 清除所有日志
   */
  clearLogs(): void {
    this.logs = [];
  }

  /**
   * 导出日志为 JSON 字符串
   */
  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }
}

// 导出单例实例
export const errorLogger = new ErrorLogger();

// 全局错误处理器
export function setupGlobalErrorHandler(): void {
  // 捕获未处理的 Promise 错误
  window.addEventListener('unhandledrejection', (event) => {
    errorLogger.error(
      'Unhandled Promise Rejection',
      event.reason instanceof Error ? event.reason : new Error(String(event.reason)),
      { type: 'unhandledrejection' }
    );
  });

  // 捕获全局错误
  window.addEventListener('error', (event) => {
    errorLogger.error(
      'Global Error',
      event.error || new Error(event.message),
      {
        type: 'error',
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      }
    );
  });
}
