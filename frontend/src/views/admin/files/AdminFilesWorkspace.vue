<template>
  <section class="admin-files-page">
    <div class="files-console">
      <header class="console-header">
        <div class="title-block">
          <span class="kicker">Files System Overview</span>
          <h1>文件系统</h1>
          <p>在管理面板中统一查看、筛选、批量维护星云盘的真实文件资产。</p>
        </div>

        <div class="header-metrics" aria-label="文件系统概览">
          <article class="metric-card">
            <span>文件总量</span>
            <strong>{{ total }}</strong>
            <small>当前筛选范围</small>
          </article>
          <article class="metric-card">
            <span>已选文件</span>
            <strong>{{ selectedIds.length }}</strong>
            <small>支持批量管理</small>
          </article>
          <article class="metric-card">
            <span>本页体积</span>
            <strong>{{ formatSize(pageFileSize) }}</strong>
            <small>文件大小合计</small>
          </article>
        </div>
      </header>

      <section class="workspace-card">
        <div class="workspace-toolbar">
          <div class="breadcrumb-line">
            <span>管理面板</span>
            <span>/</span>
            <strong>文件系统</strong>
          </div>

          <div class="toolbar-actions">
            <label class="search-box" aria-label="搜索文件">
              <el-icon><Search /></el-icon>
              <input v-model.trim="filters.keyword" type="search" placeholder="搜索文件名、报告、方案或资产表" @keydown.enter="applyFilters" />
            </label>

            <button class="soft-button primary" type="button" @click="triggerImport">
              <el-icon><Upload /></el-icon>
              <span>导入</span>
            </button>
            <button class="soft-button" type="button" :disabled="loading" @click="fetchFiles">
              <el-icon :class="{ spinning: loading }"><RefreshRight /></el-icon>
              <span>刷新</span>
            </button>

            <div ref="filterAnchorRef" class="filter-anchor">
              <button class="soft-button" type="button" @click="filterVisible = !filterVisible">
                <el-icon><Filter /></el-icon>
                <span>筛选</span>
              </button>

              <div v-if="filterVisible" class="filter-popover">
                <label class="filter-field">
                  <span>关键词</span>
                  <input v-model.trim="filters.keyword" type="text" placeholder="按文件名搜索" />
                </label>
                <label class="filter-field">
                  <span>Owner ID</span>
                  <input
                    :value="filters.ownerId"
                    type="text"
                    placeholder="例如 3518974413"
                    @input="filters.ownerId = ($event.target as HTMLInputElement).value.replace(/[^\d]/g, '')"
                  />
                </label>
                <label class="filter-field">
                  <span>存储策略</span>
                  <select v-model.number="filters.storagePolicyId">
                    <option :value="0">全部策略</option>
                    <option v-for="policy in storagePolicies" :key="policy.id" :value="policy.id">
                      {{ policy.name }}
                    </option>
                  </select>
                </label>

                <div class="filter-actions">
                  <button class="ghost-button" type="button" @click="resetFilters">重置</button>
                  <button class="apply-button" type="button" @click="applyFilters">应用筛选</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="selectedIds.length" class="batch-bar">
          <span>已选择 {{ selectedIds.length }} 个文件</span>
          <div>
            <button class="batch-button" type="button" @click="selectedIds = []">取消选择</button>
            <button class="batch-button danger" type="button">批量管理</button>
          </div>
        </div>

        <section class="file-grid" aria-label="高频文件类型">
          <article v-for="item in showcaseTypes" :key="item.name" class="file-tile">
            <span class="file-icon" :class="item.tone">{{ item.ext }}</span>
            <div>
              <strong>{{ item.name }}</strong>
              <span>{{ item.description }}</span>
            </div>
          </article>
        </section>

        <section class="table-card">
          <div class="table-head">
            <div class="col-check">
              <input :checked="allPageSelected" type="checkbox" @change="toggleSelectAll(($event.target as HTMLInputElement).checked)" />
            </div>
            <button class="col-order sortable" type="button" @click="toggleSort">
              <span>#</span>
              <el-icon :class="{ reverse: sortDirection === 'asc' }"><Sort /></el-icon>
            </button>
            <div class="col-name">文件名称</div>
            <div class="col-size">大小</div>
            <div class="col-occupied">占用空间</div>
            <div class="col-owner">Owner ID</div>
            <div class="col-created">创建时间</div>
            <div class="col-status">状态</div>
            <div class="col-actions">操作</div>
          </div>

          <div v-if="loading" class="state-shell">正在加载文件系统数据...</div>
          <div v-else-if="!pagedFiles.length" class="empty-state">
            <div class="empty-icons">
              <span class="file-icon is-pdf">PDF</span>
              <span class="file-icon is-docx">DOC</span>
              <span class="file-icon is-xlsx">XLS</span>
            </div>
            <strong>当前筛选下暂无文件</strong>
            <span>导入文件或调整筛选条件后，这里会展示真实文件列表。</span>
          </div>

          <article v-for="file in pagedFiles" :key="file.id" class="table-row" :class="{ 'is-selected': selectedIds.includes(file.id) }">
            <div class="col-check">
              <input :checked="selectedIds.includes(file.id)" type="checkbox" @change="toggleSelect(file.id, ($event.target as HTMLInputElement).checked)" />
            </div>
            <div class="col-order">{{ file.rawId }}</div>
            <div class="col-name">
              <div class="file-cell">
                <span
                  class="file-icon small"
                  :class="getFileIconTone(file)"
                  :style="getFileIconStyle(file)"
                  :title="getFileIconTitle(file)"
                >
                  {{ getFileIconLabel(file) }}
                </span>
                <div class="file-copy">
                  <strong>{{ file.fileName }}</strong>
                  <span>{{ file.filePath || file.storagePolicyName }}</span>
                </div>
              </div>
            </div>
            <div class="col-size">{{ formatSize(file.fileSize) }}</div>
            <div class="col-occupied">{{ formatSize(file.occupiedSize) }}</div>
            <div class="col-owner">
              <button class="owner-pill" type="button" @click="quickFilterOwner(file.ownerId)">
                <span class="owner-avatar">{{ String(file.ownerName || file.ownerId).slice(0, 1).toUpperCase() }}</span>
                <span>{{ file.ownerId }}</span>
              </button>
            </div>
            <div class="col-created">{{ formatDate(file.createdAt) }}</div>
            <div class="col-status">
              <span class="status-chip" :class="{ live: !file.uploading, pending: file.uploading }">
                {{ file.uploading ? '上传中' : '已入库' }}
              </span>
            </div>
            <div class="col-actions">
              <button class="action-button" type="button" title="下载" @click="handleAction('download', file)">
                <el-icon><Download /></el-icon>
              </button>
              <button class="action-button" type="button" title="分享" @click="handleAction('share', file)">
                <el-icon><Share /></el-icon>
              </button>
              <button class="action-button" type="button" title="重命名" @click="handleAction('rename', file)">
                <el-icon><EditPen /></el-icon>
              </button>
              <button class="action-button" type="button" title="详情" @click="previewFile(file)">
                <el-icon><View /></el-icon>
              </button>
            </div>
          </article>
        </section>

        <footer class="footer-bar">
          <div class="pager">
            <button class="pager-button" type="button" :disabled="page <= 1" @click="page = page - 1">上一页</button>
            <span class="pager-current">{{ page }} / {{ totalPages }}</span>
            <button class="pager-button" type="button" :disabled="page >= totalPages" @click="page = page + 1">下一页</button>
          </div>

          <select v-model.number="pageSize" class="page-size-select">
            <option :value="10">每页 10 条</option>
            <option :value="20">每页 20 条</option>
            <option :value="50">每页 50 条</option>
          </select>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Download, EditPen, Filter, RefreshRight, Search, Share, Sort, Upload, View } from '@element-plus/icons-vue';
import { useAdminFilesWorkspace } from './useAdminFilesWorkspace';

const showcaseTypes = [
  { ext: 'PDF', name: 'Quarterly_Report.pdf', description: '季度报告与审计归档', tone: 'is-pdf' },
  { ext: 'DOC', name: 'Product_Strategy.docx', description: '产品方案与协作文档', tone: 'is-docx' },
  { ext: 'XLS', name: 'Financial_Model.xlsx', description: '资产表与财务模型', tone: 'is-xlsx' },
];

const {
  filterAnchorRef,
  filterVisible,
  loading,
  page,
  pageSize,
  storagePolicies,
  selectedIds,
  sortDirection,
  filters,
  total,
  totalPages,
  pageFileSize,
  pagedFiles,
  allPageSelected,
  formatSize,
  formatDate,
  getFileIconLabel,
  getFileIconTone,
  getFileIconStyle,
  getFileIconTitle,
  fetchFiles,
  toggleSort,
  toggleSelect,
  toggleSelectAll,
  previewFile,
  quickFilterOwner,
  handleAction,
  applyFilters,
  resetFilters,
  triggerImport,
} = useAdminFilesWorkspace();
</script>

<style scoped>
.admin-files-page {
  position: relative;
  min-height: calc(100vh - 96px);
  overflow: hidden;
  border-radius: 30px;
  background:
    radial-gradient(circle at 8% 10%, rgba(255, 183, 178, 0.42), transparent 28%),
    radial-gradient(circle at 86% 12%, rgba(168, 218, 220, 0.54), transparent 30%),
    linear-gradient(135deg, #fff7fb 0%, #eef9ff 48%, #fffdf7 100%);
}

.files-console {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 22px;
  padding: 26px;
}

.console-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 520px);
  gap: 20px;
  align-items: stretch;
}

.title-block,
.workspace-card,
.metric-card,
.file-tile,
.table-card,
.filter-popover,
.batch-bar {
  border: 1px solid rgba(255, 255, 255, 0.62);
  background: rgba(255, 255, 255, 0.62);
  box-shadow:
    0 22px 48px rgba(91, 118, 155, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.74);
  backdrop-filter: blur(24px) saturate(125%);
}

.title-block {
  display: grid;
  align-content: center;
  gap: 10px;
  min-height: 184px;
  padding: 30px;
  border-radius: 26px;
}

.kicker {
  width: fit-content;
  padding: 7px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.56);
  color: #2f6f8f;
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
}

.title-block h1 {
  margin: 0;
  color: #102238;
  font-size: 42px;
  font-weight: 850;
  letter-spacing: 0;
}

.title-block p {
  max-width: 720px;
  margin: 0;
  color: #65758c;
  font-size: 15px;
  line-height: 1.8;
}

.header-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  align-content: center;
  gap: 9px;
  min-height: 156px;
  padding: 22px;
  border-radius: 22px;
}

.metric-card span,
.metric-card small {
  color: #73839b;
  font-size: 12px;
  font-weight: 700;
}

.metric-card strong {
  color: #12304d;
  font-size: 30px;
  line-height: 1;
}

.workspace-card {
  display: grid;
  gap: 18px;
  padding: 22px;
  border-radius: 28px;
}

.workspace-toolbar {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 16px;
  align-items: center;
}

.breadcrumb-line {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  color: #73839b;
  font-size: 14px;
  white-space: nowrap;
}

.breadcrumb-line strong {
  color: #19334f;
}

.toolbar-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-wrap: wrap;
}

.search-box {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  width: min(100%, 360px);
  min-height: 46px;
  padding: 0 15px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.68);
  color: #5c6f86;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.search-box input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #172033;
  font-size: 14px;
}

.soft-button,
.batch-button,
.apply-button,
.ghost-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  min-height: 46px;
  padding: 0 16px;
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(240, 248, 255, 0.72));
  color: #24415f;
  font-size: 14px;
  font-weight: 800;
  cursor: pointer;
  box-shadow:
    0 12px 24px rgba(86, 116, 150, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.88);
}

.soft-button.primary,
.apply-button {
  border-color: rgba(64, 145, 210, 0.26);
  background: linear-gradient(135deg, #4aa3df, #73c7dc 62%, #ffbbb6);
  color: #fff;
}

.soft-button:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.filter-anchor {
  position: relative;
}

.filter-popover {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  z-index: 20;
  width: min(360px, calc(100vw - 48px));
  display: grid;
  gap: 14px;
  padding: 18px;
  border-radius: 20px;
}

.filter-field {
  display: grid;
  gap: 8px;
}

.filter-field span {
  color: #31445d;
  font-size: 13px;
  font-weight: 800;
}

.filter-field input,
.filter-field select,
.page-size-select {
  min-height: 42px;
  padding: 0 14px;
  border: 1px solid rgba(214, 226, 240, 0.9);
  border-radius: 13px;
  background: rgba(255, 255, 255, 0.84);
  color: #172033;
  font-size: 14px;
  outline: none;
}

.filter-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.ghost-button,
.batch-button {
  min-height: 40px;
  box-shadow: none;
}

.batch-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 14px 12px 18px;
  border-radius: 18px;
  color: #25405c;
  font-size: 14px;
  font-weight: 800;
}

.batch-bar > div {
  display: inline-flex;
  gap: 8px;
}

.batch-button.danger {
  color: #b42318;
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.file-tile {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 96px;
  padding: 18px;
  border-radius: 22px;
}

.file-tile div {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.file-tile strong {
  overflow: hidden;
  color: #132238;
  font-size: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-tile span:last-child {
  color: #73839b;
  font-size: 13px;
}

.file-icon {
  display: inline-grid;
  place-items: center;
  width: 54px;
  height: 54px;
  flex: 0 0 auto;
  border-radius: 16px;
  color: #fff;
  font-size: 13px;
  font-weight: 900;
  box-shadow:
    0 14px 24px rgba(73, 95, 126, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.file-icon.small {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  font-size: 11px;
}

.file-icon.is-pdf {
  background: linear-gradient(135deg, #d92d20, #ff6b6b);
}

.file-icon.is-docx {
  background: linear-gradient(135deg, #1d4ed8, #60a5fa);
}

.file-icon.is-xlsx {
  background: linear-gradient(135deg, #059669, #5eead4);
}

.file-icon.is-image {
  background: linear-gradient(135deg, #7c3aed, #c084fc);
}

.file-icon.is-archive {
  background: linear-gradient(135deg, #d97706, #fbbf24);
}

.file-icon.is-folder {
  background: linear-gradient(135deg, #0284c7, #67e8f9);
}

.file-icon.is-file {
  background: linear-gradient(135deg, #64748b, #94a3b8);
}

.file-icon.is-custom {
  background: linear-gradient(135deg, #2563eb, #38bdf8);
}

.table-card {
  overflow: hidden;
  border-radius: 22px;
}

.table-head,
.table-row {
  display: grid;
  grid-template-columns: 48px 76px minmax(260px, 1.8fr) 110px 120px 160px 190px 108px 176px;
  gap: 12px;
  align-items: center;
  padding: 14px 16px;
}

.table-head {
  color: #31445d;
  font-size: 12px;
  font-weight: 900;
  background: rgba(255, 255, 255, 0.5);
}

.table-row {
  margin: 0 10px 10px;
  border: 1px solid rgba(255, 255, 255, 0.66);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.58);
  color: #42546b;
  font-size: 14px;
  box-shadow: 0 10px 22px rgba(91, 118, 155, 0.08);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.table-row:hover,
.table-row.is-selected {
  border-color: rgba(255, 183, 178, 0.72);
  box-shadow: 0 16px 30px rgba(255, 183, 178, 0.15);
  transform: translateY(-1px);
}

.sortable {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.sortable .reverse {
  transform: rotate(180deg);
}

.file-cell {
  display: flex;
  align-items: center;
  gap: 13px;
  min-width: 0;
}

.file-copy {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.file-copy strong,
.file-copy span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-copy strong {
  color: #102238;
  font-size: 15px;
}

.file-copy span {
  color: #7a8aa2;
  font-size: 12px;
}

.owner-pill {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  max-width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  color: #286f9c;
  font-weight: 800;
  cursor: pointer;
}

.owner-avatar {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  flex: 0 0 auto;
  border-radius: 999px;
  background: linear-gradient(135deg, #73c7dc, #ffbbb6);
  color: #fff;
  font-size: 13px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 11px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
}

.status-chip.live {
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
}

.status-chip.pending {
  background: rgba(245, 158, 11, 0.14);
  color: #b45309;
}

.col-actions {
  display: flex;
  justify-content: flex-end;
  gap: 7px;
}

.action-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 13px;
  background: rgba(255, 255, 255, 0.72);
  color: #53667e;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.state-shell,
.empty-state {
  display: grid;
  place-items: center;
  gap: 10px;
  min-height: 240px;
  color: #7a8aa2;
  text-align: center;
}

.empty-icons {
  display: inline-flex;
  gap: 10px;
}

.empty-state strong {
  color: #19334f;
  font-size: 18px;
}

.footer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.pager {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.pager-button {
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.64);
  color: #304963;
  font-weight: 800;
  cursor: pointer;
}

.pager-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.pager-current {
  color: #52667e;
  font-weight: 800;
}

.spinning {
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1360px) {
  .console-header,
  .workspace-toolbar {
    grid-template-columns: 1fr;
  }

  .toolbar-actions {
    justify-content: flex-start;
  }

  .table-card {
    overflow-x: auto;
  }

  .table-head,
  .table-row {
    min-width: 1320px;
  }
}

@media (max-width: 900px) {
  .files-console {
    padding: 16px;
  }

  .title-block h1 {
    font-size: 34px;
  }

  .header-metrics,
  .file-grid {
    grid-template-columns: 1fr;
  }

  .search-box,
  .soft-button,
  .page-size-select {
    width: 100%;
  }

  .footer-bar,
  .batch-bar {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>
