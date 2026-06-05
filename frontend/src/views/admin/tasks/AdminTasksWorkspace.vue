<template>
  <section class="admin-tasks-page">
    <div class="tasks-shell">
      <header class="tasks-hero">
        <div>
          <p class="hero-kicker">Admin Console</p>
          <h1>后台任务</h1>
          <p>集中查看队列任务、执行状态、失败原因和处理节点记录。</p>
        </div>
        <div class="hero-orbit" aria-hidden="true">
          <span></span>
          <strong>{{ total }}</strong>
        </div>
      </header>

      <TaskSummaryCards :metrics="metrics" :queue-cards="queueCards" :loading="statsLoading" />

      <TaskToolbar
        v-model:queue-key="filters.queueKey"
        v-model:status="filters.status"
        v-model:node-id="filters.nodeId"
        :loading="loading"
        :nodes="nodes"
        :active-filter-text="activeFilterText"
        :last-updated-at="lastUpdatedAt"
        @refresh="refresh"
        @apply="applyFilters"
        @reset="clearCurrentView"
      />

      <TaskRecordsTable
        v-model:selected-ids="selectedJobIds"
        :jobs="jobs"
        :loading="loading"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @detail="openDetail"
        @delete="removeJob"
        @page-change="changePage"
        @page-size-change="changePageSize"
      />

      <TaskDetailDrawer v-model="detailVisible" :job="selectedJob" />
    </div>
  </section>
</template>

<script setup lang="ts">
import TaskDetailDrawer from './components/TaskDetailDrawer.vue';
import TaskRecordsTable from './components/TaskRecordsTable.vue';
import TaskSummaryCards from './components/TaskSummaryCards.vue';
import TaskToolbar from './components/TaskToolbar.vue';
import { useAdminTasksWorkspace } from './useAdminTasksWorkspace';

const {
  activeFilterText,
  applyFilters,
  changePage,
  changePageSize,
  clearCurrentView,
  detailVisible,
  filters,
  jobs,
  lastUpdatedAt,
  loading,
  metrics,
  nodes,
  openDetail,
  page,
  pageSize,
  queueCards,
  refresh,
  removeJob,
  selectedJob,
  selectedJobIds,
  statsLoading,
  total,
} = useAdminTasksWorkspace();
</script>

<style scoped>
.admin-tasks-page {
  min-height: calc(100vh - 96px);
  color: #213245;
}

.tasks-shell {
  position: relative;
  overflow: hidden;
  display: grid;
  gap: 22px;
  padding: 28px;
  border: 1px solid rgba(255, 255, 255, 0.88);
  border-radius: 30px;
  background:
    radial-gradient(circle at 8% 0%, rgba(138, 216, 255, 0.34), transparent 28%),
    radial-gradient(circle at 92% 4%, rgba(255, 186, 204, 0.24), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.84), rgba(248, 252, 255, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 26px 60px rgba(90, 145, 190, 0.14);
  backdrop-filter: blur(18px);
}

.tasks-shell::after {
  content: '';
  position: absolute;
  right: 8%;
  bottom: -150px;
  width: 420px;
  height: 260px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 184, 205, 0.22), transparent 68%);
  filter: blur(24px);
  pointer-events: none;
}

.tasks-hero {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.hero-kicker,
.tasks-hero h1,
.tasks-hero p {
  margin: 0;
}

.hero-kicker {
  color: #708195;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.tasks-hero h1 {
  margin-top: 14px;
  color: #1f2d3d;
  font-size: clamp(2.8rem, 5vw, 4.8rem);
  line-height: 1;
  font-weight: 880;
}

.tasks-hero p:not(.hero-kicker) {
  margin-top: 18px;
  color: #68798b;
  font-size: 17px;
  line-height: 1.8;
}

.hero-orbit {
  display: grid;
  place-items: center;
  width: 96px;
  height: 96px;
  border-radius: 30px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(229, 246, 255, 0.64));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 34px rgba(86, 166, 214, 0.14);
}

.hero-orbit span {
  width: 34px;
  height: 8px;
  border-radius: 999px;
  background: linear-gradient(90deg, #7dd3fc, #fda4af);
}

.hero-orbit strong {
  margin-top: 8px;
  color: #1d425f;
  font-size: 28px;
}

@media (max-width: 900px) {
  .tasks-shell {
    padding: 18px;
    border-radius: 24px;
  }

  .tasks-hero {
    flex-direction: column;
  }
}
</style>
