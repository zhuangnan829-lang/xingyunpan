/**
 * Virtual scrolling composable for large lists
 * Requirements: 6.1-6.3 - Optimize search performance with virtual scrolling
 */

import { ref, computed, onMounted, onUnmounted, type Ref } from 'vue';

export interface VirtualScrollOptions {
  itemHeight: number;
  bufferSize?: number;
  containerHeight?: number;
}

export function useVirtualScroll<T>(
  items: Ref<T[]>,
  options: VirtualScrollOptions
) {
  const { itemHeight, bufferSize = 5, containerHeight = 600 } = options;

  const scrollTop = ref(0);
  const containerRef = ref<HTMLElement | null>(null);

  // Calculate visible range
  const visibleRange = computed(() => {
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const startIndex = Math.floor(scrollTop.value / itemHeight);
    const endIndex = startIndex + visibleCount;

    // Add buffer to reduce flickering
    const bufferedStart = Math.max(0, startIndex - bufferSize);
    const bufferedEnd = Math.min(items.value.length, endIndex + bufferSize);

    return {
      start: bufferedStart,
      end: bufferedEnd,
      visibleCount
    };
  });

  // Get visible items
  const visibleItems = computed(() => {
    const { start, end } = visibleRange.value;
    return items.value.slice(start, end).map((item, index) => ({
      item,
      index: start + index
    }));
  });

  // Calculate total height
  const totalHeight = computed(() => {
    return items.value.length * itemHeight;
  });

  // Calculate offset for visible items
  const offsetY = computed(() => {
    return visibleRange.value.start * itemHeight;
  });

  // Handle scroll event
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement;
    scrollTop.value = target.scrollTop;
  };

  // Scroll to specific index
  const scrollToIndex = (index: number) => {
    if (containerRef.value) {
      const targetScrollTop = index * itemHeight;
      containerRef.value.scrollTop = targetScrollTop;
      scrollTop.value = targetScrollTop;
    }
  };

  // Scroll to top
  const scrollToTop = () => {
    scrollToIndex(0);
  };

  return {
    containerRef,
    visibleItems,
    totalHeight,
    offsetY,
    handleScroll,
    scrollToIndex,
    scrollToTop,
    visibleRange
  };
}
