<template>
  <div id="app">
    <router-view v-slot="{ Component, route }">
      <!-- Use layout for authenticated routes -->
      <AppLayout v-if="route.meta.requiresAuth">
        <component :is="Component" />
      </AppLayout>
      <!-- No layout for auth pages -->
      <component v-else :is="Component" />
    </router-view>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import AppLayout from '@/components/AppLayout/index.vue';
import { useShareStore } from '@/stores/share';
import { useSearchStore } from '@/stores/search';
import { useRecycleStore } from '@/stores/recycle';
import { installDomLocalizer } from '@/utils/dom-localizer';

// Initialize all Phase 4 stores on app startup
// Each store loads from localStorage and recycle store cleans expired items
onMounted(() => {
  installDomLocalizer();
  useShareStore();
  useSearchStore();
  const recycleStore = useRecycleStore();
  // Keep recycle-bin startup work from blocking auth pages if the store shape changes.
  if (typeof recycleStore.cleanExpiredItems === 'function') {
    recycleStore.cleanExpiredItems();
  }
});
</script>

