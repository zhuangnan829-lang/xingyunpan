import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus, { ElMessage } from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import { setupGlobalErrorHandler } from './utils/error-logger'
import { registerStorageErrorHandler } from './utils/storage'
import './styles/global.css'

// Setup global error handler
setupGlobalErrorHandler()

// Register storage error handler to show user-facing notifications
registerStorageErrorHandler((error) => {
  if (error.type === 'quota_exceeded') {
    ElMessage.warning({
      message: error.message,
      duration: 5000,
      showClose: true,
    });
  }
})

// Performance monitoring in development
if (import.meta.env.DEV) {
  import('./utils/performance').then(({ getWebVitals, monitorBundleLoading }) => {
    // Monitor Web Vitals
    window.addEventListener('load', () => {
      setTimeout(() => {
        getWebVitals()
        monitorBundleLoading()
      }, 0)
    })
  })
}

const app = createApp(App)
const pinia = createPinia()

// Register Element Plus icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(ElementPlus)

app.mount('#app')
