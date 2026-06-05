<template>
  <section class="oauth-page">
    <div class="oauth-shell">
      <header class="oauth-hero">
        <div>
          <p class="hero-kicker">Admin Console</p>
          <h1>OAuth 应用</h1>
          <p>管理星云盘官方客户端、桌面挂载和第三方接入授权，统一维护 Client ID、回调地址与授权范围。</p>
        </div>
        <div class="hero-panel" aria-label="OAuth 状态">
          <span>已启用</span>
          <strong>{{ apps.filter((app) => app.enabled).length }}</strong>
          <small>当前可授权应用</small>
        </div>
      </header>

      <OAuthAppMetrics :metrics="metrics" />

      <OAuthAppToolbar v-model:keyword="keyword" v-model:status="status" @refresh="refresh" @create="openCreate" />

      <section v-loading="loading" class="apps-grid">
        <button class="create-card" type="button" @click="openCreate">
          <Plus />
          <span>新建 OAuth 应用</span>
        </button>

        <OAuthAppCard v-for="app in pagedApps" :key="app.id" :app="app" @open="openApp" @toggle="toggleApp" @delete="deleteApp" />
      </section>

      <footer class="oauth-pagination">
        <el-pagination
          layout="prev, pager, next"
          :total="filteredApps.length"
          :page-size="pageSize"
          :current-page="page"
          @current-change="changePage"
        />
        <select :value="pageSize" @change="changePageSize(Number(($event.target as HTMLSelectElement).value))">
          <option :value="11">每页 11 条</option>
          <option :value="20">每页 20 条</option>
          <option :value="50">每页 50 条</option>
        </select>
      </footer>

      <OAuthAppCreateDialog v-model="createVisible" :draft="draft" :scope-options="scopeOptions" @submit="createApp" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { Plus } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';
import OAuthAppCard from './components/OAuthAppCard.vue';
import OAuthAppCreateDialog from './components/OAuthAppCreateDialog.vue';
import OAuthAppMetrics from './components/OAuthAppMetrics.vue';
import OAuthAppToolbar from './components/OAuthAppToolbar.vue';
import { useOAuthAppsWorkspace } from './useOAuthAppsWorkspace';

const {
  apps,
  changePage,
  changePageSize,
  createApp,
  createVisible,
  deleteApp,
  draft,
  filteredApps,
  keyword,
  loading,
  loadApps,
  metrics,
  openCreate,
  page,
  pagedApps,
  pageSize,
  refresh,
  scopeOptions,
  status,
  toggleApp,
} = useOAuthAppsWorkspace();

const router = useRouter();

const openApp = (app: { id: string }) => {
  router.push(`/admin/oauth/${app.id}`);
};

onMounted(loadApps);
</script>

<style scoped>
.oauth-page {
  min-height: calc(100vh - 96px);
  color: #1d2d3f;
}

.oauth-shell {
  display: grid;
  gap: 18px;
  min-height: calc(100vh - 96px);
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 30px;
  background:
    radial-gradient(circle at 12% 4%, rgba(125, 211, 252, 0.34), transparent 30%),
    radial-gradient(circle at 88% 8%, rgba(252, 188, 202, 0.28), transparent 26%),
    radial-gradient(circle at 24% 96%, rgba(205, 183, 255, 0.18), transparent 30%),
    linear-gradient(135deg, rgba(248, 252, 255, 0.9), rgba(255, 248, 252, 0.82) 52%, rgba(246, 251, 255, 0.92));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.96), 0 28px 68px rgba(91, 145, 186, 0.13);
}

.oauth-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  min-height: 178px;
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 26px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.42));
  box-shadow: 0 22px 54px rgba(81, 120, 154, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(18px);
}

.hero-kicker,
.oauth-hero h1,
.oauth-hero p {
  margin: 0;
}

.hero-kicker {
  color: #6c7c90;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.oauth-hero h1 {
  margin-top: 16px;
  color: #152235;
  font-size: clamp(44px, 5.2vw, 76px);
  line-height: 1;
  font-weight: 880;
}

.oauth-hero p:not(.hero-kicker) {
  max-width: 780px;
  margin-top: 18px;
  color: #66768a;
  font-size: 16px;
  line-height: 1.85;
}

.hero-panel {
  display: grid;
  gap: 8px;
  min-width: 146px;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 22px;
  background: linear-gradient(145deg, rgba(229, 247, 255, 0.82), rgba(255, 255, 255, 0.46));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.hero-panel span,
.hero-panel small {
  color: #6c7e92;
  font-size: 13px;
  font-weight: 800;
}

.hero-panel strong {
  color: #1674bd;
  font-size: 40px;
  line-height: 1;
}

.apps-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.create-card {
  display: grid;
  place-items: center;
  align-content: center;
  gap: 12px;
  min-height: 260px;
  border: 1px dashed rgba(255, 255, 255, 0.88);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.28);
  color: #66778a;
  font-size: 18px;
  font-weight: 820;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.76);
  backdrop-filter: blur(14px);
}

.create-card svg {
  width: 32px;
  height: 32px;
}

.oauth-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  min-height: 52px;
}

.oauth-pagination select {
  min-height: 42px;
  padding: 0 34px 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  outline: 0;
  background: rgba(255, 255, 255, 0.56);
  color: #26394d;
  font: inherit;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

@media (max-width: 1280px) {
  .apps-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 780px) {
  .oauth-shell {
    padding: 16px;
    border-radius: 24px;
  }

  .oauth-hero {
    flex-direction: column;
    padding: 22px;
  }

  .hero-panel {
    width: 100%;
    box-sizing: border-box;
  }

  .apps-grid {
    grid-template-columns: 1fr;
  }
}
</style>
