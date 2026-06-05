/**
 * 格式化工具函数
 * 提供文件大小和时间戳的格式化功能
 */

/**
 * 格式化文件大小
 * @param bytes 文件大小（字节）
 * @returns 格式化后的文件大小字符串
 * 
 * Requirements: 26.1-26.4
 * - < 1KB: 显示为 Bytes
 * - 1KB - 1MB: 显示为 KB (2位小数)
 * - 1MB - 1GB: 显示为 MB (2位小数)
 * - > 1GB: 显示为 GB (2位小数)
 */
export function formatFileSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} Bytes`;
  }
  
  if (bytes < 1024 * 1024) {
    const kb = bytes / 1024;
    return `${kb.toFixed(2)} KB`;
  }
  
  if (bytes < 1024 * 1024 * 1024) {
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(2)} MB`;
  }
  
  const gb = bytes / (1024 * 1024 * 1024);
  return `${gb.toFixed(2)} GB`;
}

/**
 * 格式化时间戳
 * @param timestamp ISO 8601 时间戳字符串或 Date 对象
 * @returns 格式化后的时间字符串
 * 
 * Requirements: 27.1-27.4
 * - 今天: "HH:mm"
 * - 昨天: "昨天 HH:mm"
 * - 今年: "MM-DD HH:mm"
 * - 往年: "YYYY-MM-DD"
 */
export function formatTimestamp(timestamp: string | Date): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
  const now = new Date();
  
  // 获取今天的开始时间（00:00:00）
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  
  // 获取昨天的开始时间
  const yesterdayStart = new Date(todayStart);
  yesterdayStart.setDate(yesterdayStart.getDate() - 1);
  
  // 获取今年的开始时间
  const thisYearStart = new Date(now.getFullYear(), 0, 1);
  
  // 格式化时间部分 HH:mm
  const formatTime = (d: Date): string => {
    const hours = d.getHours().toString().padStart(2, '0');
    const minutes = d.getMinutes().toString().padStart(2, '0');
    return `${hours}:${minutes}`;
  };
  
  // 格式化日期部分 MM-DD
  const formatMonthDay = (d: Date): string => {
    const month = (d.getMonth() + 1).toString().padStart(2, '0');
    const day = d.getDate().toString().padStart(2, '0');
    return `${month}-${day}`;
  };
  
  // 格式化完整日期 YYYY-MM-DD
  const formatFullDate = (d: Date): string => {
    const year = d.getFullYear();
    const month = (d.getMonth() + 1).toString().padStart(2, '0');
    const day = d.getDate().toString().padStart(2, '0');
    return `${year}-${month}-${day}`;
  };
  
  // 判断是今天
  if (date >= todayStart) {
    return formatTime(date);
  }
  
  // 判断是昨天
  if (date >= yesterdayStart && date < todayStart) {
    return `昨天 ${formatTime(date)}`;
  }
  
  // 判断是今年
  if (date >= thisYearStart) {
    return `${formatMonthDay(date)} ${formatTime(date)}`;
  }
  
  // 往年
  return formatFullDate(date);
}
