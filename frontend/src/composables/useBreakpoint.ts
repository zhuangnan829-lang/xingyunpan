import { ref, onMounted, onUnmounted } from 'vue';

// Breakpoint constants
export const BREAKPOINTS = {
  mobile: 768,
  tablet: 1024,
  desktop: 1280,
} as const;

export type DeviceType = 'mobile' | 'tablet' | 'desktop';

export function useBreakpoint() {
  const deviceType = ref<DeviceType>('desktop');
  const windowWidth = ref(window.innerWidth);

  const updateDeviceType = () => {
    windowWidth.value = window.innerWidth;
    
    if (windowWidth.value < BREAKPOINTS.mobile) {
      deviceType.value = 'mobile';
    } else if (windowWidth.value < BREAKPOINTS.tablet) {
      deviceType.value = 'tablet';
    } else {
      deviceType.value = 'desktop';
    }
  };

  const isMobile = () => deviceType.value === 'mobile';
  const isTablet = () => deviceType.value === 'tablet';
  const isDesktop = () => deviceType.value === 'desktop';

  onMounted(() => {
    updateDeviceType();
    window.addEventListener('resize', updateDeviceType);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', updateDeviceType);
  });

  return {
    deviceType,
    windowWidth,
    isMobile,
    isTablet,
    isDesktop,
  };
}
