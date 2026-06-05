/**
 * Performance Monitoring Utilities
 * 
 * Provides utilities for monitoring and measuring application performance
 */

interface PerformanceMetric {
  name: string
  value: number
  timestamp: number
}

class PerformanceMonitor {
  private metrics: PerformanceMetric[] = []
  private marks: Map<string, number> = new Map()

  /**
   * Start measuring a performance metric
   */
  startMeasure(name: string): void {
    this.marks.set(name, performance.now())
  }

  /**
   * End measuring and record the metric
   */
  endMeasure(name: string): number {
    const startTime = this.marks.get(name)
    if (!startTime) {
      console.warn(`No start mark found for: ${name}`)
      return 0
    }

    const duration = performance.now() - startTime
    this.marks.delete(name)

    this.metrics.push({
      name,
      value: duration,
      timestamp: Date.now(),
    })

    return duration
  }

  /**
   * Get all recorded metrics
   */
  getMetrics(): PerformanceMetric[] {
    return [...this.metrics]
  }

  /**
   * Get metrics by name
   */
  getMetricsByName(name: string): PerformanceMetric[] {
    return this.metrics.filter(m => m.name === name)
  }

  /**
   * Get average value for a metric
   */
  getAverageMetric(name: string): number {
    const metrics = this.getMetricsByName(name)
    if (metrics.length === 0) return 0

    const sum = metrics.reduce((acc, m) => acc + m.value, 0)
    return sum / metrics.length
  }

  /**
   * Clear all metrics
   */
  clearMetrics(): void {
    this.metrics = []
    this.marks.clear()
  }

  /**
   * Log performance summary
   */
  logSummary(): void {
    const metricNames = [...new Set(this.metrics.map(m => m.name))]
    
    console.group('📊 Performance Summary')
    metricNames.forEach(name => {
      const avg = this.getAverageMetric(name)
      const metrics = this.getMetricsByName(name)
      const min = Math.min(...metrics.map(m => m.value))
      const max = Math.max(...metrics.map(m => m.value))
      
      console.log(`${name}:`)
      console.log(`  Average: ${avg.toFixed(2)}ms`)
      console.log(`  Min: ${min.toFixed(2)}ms`)
      console.log(`  Max: ${max.toFixed(2)}ms`)
      console.log(`  Count: ${metrics.length}`)
    })
    console.groupEnd()
  }
}

// Singleton instance
export const performanceMonitor = new PerformanceMonitor()

/**
 * Measure the execution time of an async function
 */
export async function measureAsync<T>(
  name: string,
  fn: () => Promise<T>
): Promise<T> {
  performanceMonitor.startMeasure(name)
  try {
    const result = await fn()
    return result
  } finally {
    const duration = performanceMonitor.endMeasure(name)
    if (import.meta.env.DEV) {
      console.log(`⏱️ ${name}: ${duration.toFixed(2)}ms`)
    }
  }
}

/**
 * Measure the execution time of a sync function
 */
export function measureSync<T>(
  name: string,
  fn: () => T
): T {
  performanceMonitor.startMeasure(name)
  try {
    const result = fn()
    return result
  } finally {
    const duration = performanceMonitor.endMeasure(name)
    if (import.meta.env.DEV) {
      console.log(`⏱️ ${name}: ${duration.toFixed(2)}ms`)
    }
  }
}

/**
 * Get Web Vitals metrics
 */
export function getWebVitals(): void {
  if (typeof window === 'undefined') return

  // First Contentful Paint (FCP)
  const paintEntries = performance.getEntriesByType('paint')
  const fcp = paintEntries.find(entry => entry.name === 'first-contentful-paint')
  if (fcp) {
    console.log(`🎨 First Contentful Paint: ${fcp.startTime.toFixed(2)}ms`)
  }

  // Largest Contentful Paint (LCP)
  const observer = new PerformanceObserver((list) => {
    const entries = list.getEntries()
    const lastEntry = entries[entries.length - 1]
    console.log(`🖼️ Largest Contentful Paint: ${lastEntry.startTime.toFixed(2)}ms`)
  })
  
  try {
    observer.observe({ entryTypes: ['largest-contentful-paint'] })
  } catch (e) {
    // LCP not supported
  }

  // Time to Interactive (TTI) approximation
  if (document.readyState === 'complete') {
    const navTiming = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
    if (navTiming) {
      const tti = navTiming.domInteractive
      console.log(`⚡ Time to Interactive: ${tti.toFixed(2)}ms`)
    }
  }
}

/**
 * Monitor bundle loading performance
 */
export function monitorBundleLoading(): void {
  if (typeof window === 'undefined') return

  const resourceEntries = performance.getEntriesByType('resource') as PerformanceResourceTiming[]
  const jsResources = resourceEntries.filter(entry => entry.name.endsWith('.js'))
  const cssResources = resourceEntries.filter(entry => entry.name.endsWith('.css'))

  console.group('📦 Bundle Loading Performance')
  
  let totalJsSize = 0
  let totalJsTime = 0
  
  console.log('JavaScript:')
  jsResources.forEach(resource => {
    const size = resource.transferSize || 0
    const time = resource.duration
    totalJsSize += size
    totalJsTime += time
    
    if (import.meta.env.DEV) {
      console.log(`  ${resource.name.split('/').pop()}: ${(size / 1024).toFixed(2)}KB in ${time.toFixed(2)}ms`)
    }
  })
  
  console.log(`Total JS: ${(totalJsSize / 1024).toFixed(2)}KB in ${totalJsTime.toFixed(2)}ms`)
  
  let totalCssSize = 0
  let totalCssTime = 0
  
  console.log('CSS:')
  cssResources.forEach(resource => {
    const size = resource.transferSize || 0
    const time = resource.duration
    totalCssSize += size
    totalCssTime += time
    
    if (import.meta.env.DEV) {
      console.log(`  ${resource.name.split('/').pop()}: ${(size / 1024).toFixed(2)}KB in ${time.toFixed(2)}ms`)
    }
  })
  
  console.log(`Total CSS: ${(totalCssSize / 1024).toFixed(2)}KB in ${totalCssTime.toFixed(2)}ms`)
  console.groupEnd()
}
