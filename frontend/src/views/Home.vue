<template>
  <div class="workspace-page">
    <section class="workspace-card">
      <header class="workspace-toolbar">
        <div class="toolbar-copy">
          <strong>{{ currentFolderTitle }}</strong>
          <span>{{ folderSummary }}</span>
        </div>

        <div class="toolbar-actions">
          <button class="toolbar-button" type="button" :disabled="!canGoBack || fileStore.loading" @click="navigateToParent">
            <Back class="button-icon" />
            <span>{{ t('backToParent') }}</span>
          </button>
          <button class="toolbar-button accent" type="button" :disabled="fileStore.loading" @click="refreshFiles">
            <RefreshRight class="button-icon" />
            <span>{{ t('reload') }}</span>
          </button>
          <div class="view-control">
            <button class="toolbar-button" type="button" @click="viewPanelVisible = !viewPanelVisible">
              <component :is="viewModeIcon" class="button-icon" />
              <span>{{ t('view') }}</span>
            </button>

            <div v-if="viewPanelVisible" class="view-panel">
              <p class="view-panel-title">{{ t('layout') }}</p>
              <div class="view-segment">
                <button
                  v-for="option in viewOptions"
                  :key="option.mode"
                  class="view-option"
                  :class="{ active: activeViewMode === option.mode }"
                  type="button"
                  @pointerdown.stop
                  @click.stop.prevent="setViewMode(option.mode)"
                >
                  <component :is="option.icon" class="button-icon" />
                  <span>{{ option.label }}</span>
                </button>
              </div>

              <template v-if="activeViewMode === 'table'">
                <p class="view-panel-title">{{ t('columnSettings') }}</p>
                <button class="panel-wide-button" type="button" @pointerdown.stop @click.stop.prevent="columnDialogVisible = true">
                  <Setting class="button-icon" />
                  <span>{{ t('columnSettings') }}</span>
                </button>
              </template>

              <template v-if="activeViewMode === 'grid'">
                <p class="view-panel-title">{{ t('thumbnails') }}</p>
                <div class="view-segment two">
                  <button class="view-option" :class="{ active: showThumbnails }" type="button" @pointerdown.stop @click.stop.prevent="showThumbnails = true">
                    <Picture class="button-icon" />
                    <span>{{ t('enable') }}</span>
                  </button>
                  <button class="view-option" :class="{ active: !showThumbnails }" type="button" @pointerdown.stop @click.stop.prevent="showThumbnails = false">
                    <Close class="button-icon" />
                    <span>{{ t('disable') }}</span>
                  </button>
                </div>
                <label class="slider-field">
                  <span>{{ t('cardSize') }}</span>
                  <input v-model.number="gridItemSize" type="range" min="150" max="320" step="10" />
                  <small><span>150</span><span>320</span></small>
                </label>
              </template>

              <template v-if="activeViewMode === 'gallery'">
                <label class="slider-field">
                  <span>{{ t('imageSize') }}</span>
                  <input v-model.number="galleryImageSize" type="range" min="50" max="500" step="10" />
                  <small><span>50</span><span>500</span></small>
                </label>
              </template>

              <label class="slider-field">
                <span>{{ t('pageSize') }}</span>
                <input v-model.number="pageSize" type="range" min="50" :max="maxPageSize" step="50" />
                <small><span>50</span><span>{{ maxPageSize }}</span></small>
              </label>
            </div>
          </div>
        </div>
      </header>

      <div v-if="selectedItems.length > 0" class="selection-toolbar">
        <button class="selection-icon-button" type="button" :title="t('cancelSelection')" @click="clearSelection">
          <Close class="button-icon" />
        </button>
        <span class="selection-count">{{ selectedItems.length }} {{ t('selectedObjects') }}</span>
        <div class="selection-actions">
          <button class="selection-icon-button" type="button" :title="t('download')" :disabled="selectedFiles.length === 0" @click="downloadSelected">
            <Download class="button-icon" />
          </button>
          <button class="selection-icon-button" type="button" :title="t('copyTo')" @click="openBatchCopyDialog">
            <CopyDocument class="button-icon" />
          </button>
          <button class="selection-icon-button" type="button" :title="t('moveTo')" @click="openBatchMoveDialog">
            <FolderOpened class="button-icon" />
          </button>
          <button class="selection-icon-button danger" type="button" :title="t('delete')" @click="deleteSelected">
            <Delete class="button-icon" />
          </button>
        </div>
      </div>

      <div class="workspace-breadcrumb">
        <button class="crumb-button root-crumb" type="button" @click="navigateToRoot">{{ t('allFiles') }}</button>
        <template v-for="item in fileStore.breadcrumb" :key="item.id">
          <span class="crumb-separator">/</span>
          <button class="crumb-button" type="button" @click="navigateToFolder(item.id)">{{ item.name }}</button>
        </template>
      </div>

      <div class="workspace-list-shell" @contextmenu="showWorkspaceContextMenu">
        <FileList
          :view-mode="activeViewMode"
          :show-thumbnails="showThumbnails"
          :item-size="gridItemSize"
          :image-size="galleryImageSize"
          :page-size="pageSize"
          :visible-columns="visibleColumns"
          :selected-keys="selectedKeys"
          :highlight-keyword="fileStore.searchKeyword"
          @refresh="refreshFiles"
          @folder-click="handleFolderOpen"
          @folder-preview="handleFolderPrimaryClick"
          @file-click="handleFilePrimaryClick"
          @file-open="openFilePreview"
          @rename="handleActionRename"
          @delete="handleDelete"
          @move="openMoveDialog"
          @copy="openCopyDialog"
          @download="handleDownload"
          @share="openShareForFile"
          @version-history="openVersionHistory"
          @collaborate="openCollaborate"
          @toggle-select="toggleSelection"
          @select-one="selectSingle"
          @blank-click="clearSelection"
        />

      </div>
    </section>

    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="workspace-context-menu"
        :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
        role="menu"
        @click.stop
        @contextmenu.prevent
      >
        <div class="context-menu-section">
          <button class="context-menu-item" type="button" role="menuitem" @click="triggerWorkspaceFileUpload">
            <Upload class="context-menu-icon" />
            <span>{{ t('uploadFiles') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="triggerWorkspaceFolderUpload">
            <FolderAdd class="context-menu-icon" />
            <span>{{ t('uploadFolder') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="uploadFromClipboard">
            <CopyDocument class="context-menu-icon" />
            <span>{{ t('uploadFromClipboard') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="openOfflineDownload">
            <Download class="context-menu-icon" />
            <span>{{ t('offlineDownloads') }}</span>
          </button>
        </div>

        <div class="context-menu-section">
          <button class="context-menu-item" type="button" role="menuitem" @click="openWorkspaceCreateFolderDialog">
            <FolderAdd class="context-menu-icon" />
            <span>{{ t('createFolder') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('file')">
            <DocumentAdd class="context-menu-icon" />
            <span>{{ t('createFile') }}</span>
          </button>
        </div>

        <div class="context-menu-section">
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('markdown')">
            <span class="context-app-icon markdown">MD</span>
            <span>Markdown (.md)</span>
          </button>
          <button
            class="context-menu-item has-submenu"
            :class="{ active: drawioSubmenuVisible }"
            type="button"
            role="menuitem"
            @mouseenter="drawioSubmenuVisible = true"
            @focus="drawioSubmenuVisible = true"
          >
            <span class="context-app-icon drawio">IO</span>
            <span>draw.io</span>
            <ArrowRight class="context-submenu-arrow" />
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('text')">
            <Tickets class="context-app-icon text" />
            <span>{{ t('textFile') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('excalidraw')">
            <EditPen class="context-app-icon excalidraw" />
            <span>Excalidraw (.excalidraw)</span>
          </button>
        </div>

        <div class="context-menu-section">
          <button class="context-menu-item" type="button" role="menuitem" @click="refreshFromContextMenu">
            <RefreshRight class="context-menu-icon" />
            <span>{{ t('reload') }}</span>
          </button>
        </div>

        <div v-if="drawioSubmenuVisible" class="workspace-context-submenu" @mouseleave="drawioSubmenuVisible = false">
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('drawio')">
            <span>{{ t('diagramFile') }}</span>
          </button>
          <button class="context-menu-item" type="button" role="menuitem" @click="createWorkspaceFile('dwb')">
            <span>{{ t('whiteboardFile') }}</span>
          </button>
        </div>
      </div>
    </Teleport>

    <input ref="workspaceFileInputRef" class="hidden-file-input" type="file" multiple @change="handleWorkspaceFileSelect" />
    <input
      ref="workspaceFolderInputRef"
      class="hidden-file-input"
      type="file"
      multiple
      webkitdirectory
      directory
      @change="handleWorkspaceFileSelect"
    />

    <CreateFolderDialog v-model:visible="createFolderVisible" @confirm="handleCreateFolder" />

    <el-drawer
      v-model="detailDrawerVisible"
      size="560px"
      :with-header="false"
      destroy-on-close
      class="file-detail-drawer"
    >
      <div v-if="selectedEntry" class="detail-shell">
        <section class="detail-hero">
          <div class="detail-hero-badge">{{ selectedEntry.isFolder ? t('folderDetails') : fileTypeLabel }}</div>
          <h2>{{ selectedEntry.name }}</h2>
          <p>{{ selectedEntry.isFolder ? folderPathLabel : selectedFileMime }}</p>

          <div class="detail-stat-grid">
            <article class="detail-stat-card">
              <span>{{ selectedEntry.isFolder ? t('type') : t('fileSize') }}</span>
              <strong>{{ selectedEntry.isFolder ? t('folder') : formattedSelectedFileSize }}</strong>
            </article>
            <article class="detail-stat-card">
              <span>{{ t('updatedAt') }}</span>
              <strong>{{ formattedSelectedTime }}</strong>
            </article>
            <article class="detail-stat-card">
              <span>{{ selectedEntry.isFolder ? t('parentFolder') : t('physicalFileId') }}</span>
              <strong>{{ selectedEntry.isFolder ? folderParentLabel : selectedFile?.physical_file_id || '-' }}</strong>
            </article>
          </div>

          <div class="detail-primary-actions">
            <button class="hero-button primary" type="button" @click="openRenameDialog">{{ t('rename') }}</button>
            <button
              v-if="selectedFile"
              class="hero-button"
              type="button"
              @click="handleDownload(selectedFile)"
            >
              {{ t('download') }}
            </button>
            <button
              v-if="selectedFile"
              class="hero-button"
              type="button"
              @click="openShareDialog"
            >
              {{ t('share') }}
            </button>
            <button
              v-if="selectedFile"
              class="hero-button accent-button"
              type="button"
              @click="openMatchedApp"
            >
              {{ t('open') }} app
            </button>
            <button
              v-if="selectedFolder"
              class="hero-button accent-button"
              type="button"
              @click="openFolderFromDrawer"
            >
              {{ t('open') + t('folder') }}
            </button>
          </div>
        </section>

        <template v-if="selectedFile">
          <section class="detail-panel">
            <div class="panel-head">
              <div>
                <p class="panel-kicker">Built-in app</p>
                <h3>File browser apps</h3>
              </div>
              <span v-if="selectedFile.browser_app?.name" class="panel-pill success">Matched</span>
            </div>

            <div v-if="selectedFile.browser_app" class="browser-app-card">
              <div class="browser-app-mark" :style="{ '--app-accent': selectedFile.browser_app.accent || '#2563eb' }">
                {{ selectedFile.browser_app.icon || 'APP' }}
              </div>
              <div class="browser-app-copy">
                <strong>{{ selectedFile.browser_app.name || 'Built-in app' }}</strong>
                <span>{{ selectedFile.browser_app.type || 'File browser apps' }}</span>
                <small>
                  {{ selectedFile.browser_app.open_in_new_window ? 'Open in new window' : 'Open in current window' }}
                  路 {{ selectedFile.browser_app.max_size || '-' }} {{ selectedFile.browser_app.max_size_unit || 'MB' }}
                </small>
              </div>
            </div>
            <div v-else class="panel-empty">No browser app rule matches this file yet.</div>
          </section>

          <section v-if="selectedFile.encryption_status?.visible" class="detail-panel">
            <div class="panel-head">
              <div>
                <p class="panel-kicker">Encryption</p>
                <h3>Encryption status</h3>
              </div>
              <span class="panel-pill" :class="{ success: selectedFile.encryption_status.encrypted }">
                {{ selectedFile.encryption_status.encrypted ? 'Encrypted' : 'Not encrypted' }}
              </span>
            </div>

            <div class="encryption-card" :class="{ active: selectedFile.encryption_status.encrypted }">
              <strong>{{ selectedFile.encryption_status.encrypted ? 'Encrypted by storage policy' : 'Standard storage policy' }}</strong>
              <span>
                {{ selectedFile.encryption_status.storage_policy_name || 'No storage policy bound' }}
                <template v-if="selectedFile.encryption_status.key_id"> / {{ selectedFile.encryption_status.key_id }}</template>
              </span>
            </div>
          </section>

          <section class="detail-panel">
            <div class="panel-head">
              <div>
                <p class="panel-kicker">Custom properties</p>
                <h3>File business fields</h3>
              </div>
              <span class="panel-pill">{{ customPropertyState.definitions.length }} {{ t('itemsUnit') }}</span>
            </div>

            <div v-if="customPropertyState.loading" class="panel-empty">Loading property settings...</div>
            <div v-else-if="customPropertyState.definitions.length === 0" class="panel-empty">
              No custom properties have been configured yet. Add fields in Parameter settings -> File system -> Custom properties.
            </div>
            <div v-else class="detail-form">
              <article v-for="item in customPropertyState.definitions" :key="item.key" class="detail-field-card">
                <div class="detail-field-head">
                  <div>
                    <strong>{{ item.name }}</strong>
                    <span>{{ item.key }}</span>
                  </div>
                  <em class="detail-field-type">{{ propertyTypeLabel(item.type) }}</em>
                </div>

                <template v-if="item.type === 'rating'">
                  <div class="rating-row">
                    <button
                      v-for="star in item.maxValue || 5"
                      :key="star"
                      type="button"
                      class="rating-star"
                      :class="{ active: star <= Number(customPropertyState.values[item.key] || 0) }"
                      @click="customPropertyState.values[item.key] = String(star)"
                    >
                      ★
                    </button>
                  </div>
                </template>

                <template v-else-if="item.type === 'switch'">
                  <label class="switch-row">
                    <input
                      :checked="customPropertyState.values[item.key] === 'true'"
                      type="checkbox"
                      @change="setSwitchValue(item.key, ($event.target as HTMLInputElement).checked)"
                    />
                    <span>{{ customPropertyState.values[item.key] === 'true' ? 'Enabled' : 'Disabled' }}</span>
                  </label>
                </template>

                <template v-else-if="item.type === 'date'">
                  <input
                    v-model="customPropertyState.values[item.key]"
                    type="date"
                    class="detail-input"
                  />
                </template>

                <template v-else-if="item.type === 'multi_select'">
                  <div class="option-chip-wrap">
                    <button
                      v-for="option in item.options || []"
                      :key="option"
                      type="button"
                      class="option-chip"
                      :class="{ active: selectedArrayValues(item.key).includes(option) }"
                      @click="toggleArrayOption(item.key, option)"
                    >
                      {{ option }}
                    </button>
                  </div>
                  <small class="detail-field-help">Multiple values can be selected and saved to the current file.</small>
                </template>

                <template v-else-if="item.type === 'tags'">
                  <div class="option-chip-wrap" v-if="(item.options || []).length">
                    <button
                      v-for="option in item.options || []"
                      :key="option"
                      type="button"
                      class="option-chip"
                      :class="{ active: selectedArrayValues(item.key).includes(option) }"
                      @click="toggleArrayOption(item.key, option)"
                    >
                      {{ option }}
                    </button>
                  </div>
                  <input
                    :value="selectedArrayValues(item.key).join(', ')"
                    type="text"
                    class="detail-input"
                    :placeholder="item.options?.length ? 'You can also type tags directly, separated by commas' : 'Enter tags, separated by commas'"
                    @input="setArrayInput(item.key, ($event.target as HTMLInputElement).value)"
                  />
                </template>

                <template v-else>
                  <textarea
                    v-model="customPropertyState.values[item.key]"
                    class="detail-textarea"
                    :placeholder="item.defaultValue || 'Please enter content...'"
                    rows="3"
                  ></textarea>
                  <small class="detail-field-help">{{ textFieldRuleLabel(item) }}</small>
                </template>
              </article>
            </div>

            <div class="detail-actions">
              <span class="detail-status">
                {{ customPropertyState.lastModified ? 'Last saved: ' + customPropertyState.lastModified : 'No file properties saved yet' }}
              </span>
              <button
                class="hero-button primary save-button"
                type="button"
                :disabled="customPropertyState.loading || customPropertyState.saving || !selectedFile"
                @click="saveFileCustomProperties"
              >
                <span>{{ customPropertyState.saving ? 'Saving...' : 'Save properties' }}</span>
              </button>
            </div>
          </section>
        </template>

        <template v-else-if="selectedFolder">
          <section class="detail-panel">
            <div class="panel-head">
              <div>
                <p class="panel-kicker">Folder information</p>
                <h3>Current directory overview</h3>
              </div>
              <span class="panel-pill">{{ t('folder') }}</span>
            </div>

            <div class="folder-info-grid">
              <article class="folder-info-card">
                <span>{{ t('folderPath') }}</span>
                <strong>{{ folderPathLabel }}</strong>
              </article>
              <article class="folder-info-card">
                <span>{{ t('folderCountInLevel') }}</span>
                <strong>{{ selectedFolder.directory_stats?.folder_count ?? fileStore.folders.length }}</strong>
              </article>
              <article class="folder-info-card">
                <span>{{ t('fileCountInLevel') }}</span>
                <strong>{{ selectedFolder.directory_stats?.file_count ?? fileStore.files.length }}</strong>
              </article>
              <article class="folder-info-card">
                <span>{{ t('childCount') }}</span>
                <strong>{{ selectedFolder.directory_stats?.child_count ?? (fileStore.folders.length + fileStore.files.length) }}</strong>
              </article>
              <article class="folder-info-card">
                <span>{{ t('totalFileSize') }}</span>
                <strong>{{ formattedFolderTotalSize }}</strong>
              </article>
              <article class="folder-info-card">
                <span>{{ t('statsCache') }}</span>
                <strong>{{ folderCacheStatusLabel }}</strong>
              </article>
            </div>
          </section>
        </template>
      </div>
    </el-drawer>

    <RenameDialog
      v-model:visible="renameDialogVisible"
      :item="selectedEntry ? { ...selectedEntry } : null"
      :is-folder="Boolean(selectedFolder)"
      @confirm="confirmRename"
    />

    <ShareDialog
      v-model:visible="shareDialogVisible"
      :file-ids="selectedFile ? [String(selectedFile.id)] : []"
    />

    <FolderSelectDialog
      v-model:visible="folderSelectVisible"
      :title="folderSelectTitle"
      :current-folder-id="fileStore.currentFolderId"
      :exclude-folder-id="folderSelectExcludeId"
      @confirm="confirmFolderOperation"
    />

    <el-dialog
      v-model="previewDialogVisible"
      width="min(1080px, 92vw)"
      top="5vh"
      class="file-preview-dialog"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>File preview</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <button class="dialog-close-button" type="button" @click="previewDialogVisible = false">
            <Close class="button-icon" />
          </button>
        </div>
      </template>

      <div class="preview-shell">
        <div v-if="previewState.loading" class="preview-empty">Loading preview...</div>
        <div v-else-if="previewState.error" class="preview-empty">{{ previewState.error }}</div>
        <div v-else-if="previewState.kind === 'docx'" class="docx-preview" v-html="previewState.html"></div>
        <pre v-else-if="previewState.kind === 'text'" class="text-preview">{{ previewState.text }}</pre>
        <img v-else-if="previewState.kind === 'image'" class="media-preview image-preview" :src="previewState.url" alt="" />
        <video v-else-if="previewState.kind === 'video'" class="media-preview" :src="previewState.url" controls autoplay></video>
        <audio v-else-if="previewState.kind === 'audio'" class="audio-preview" :src="previewState.url" controls autoplay></audio>
        <iframe v-else-if="previewState.kind === 'office'" class="office-preview" :src="previewState.url" title="Office file preview"></iframe>
        <div v-else-if="previewState.kind === 'pdf'" class="pdf-preview">
          <img v-for="(page, index) in previewState.pdfPages" :key="index" :src="page" :alt="'PDF page ' + (index + 1)" />
        </div>
        <div v-else class="preview-empty">This file type cannot be previewed in the app yet. Download it to view.</div>
      </div>
    </el-dialog>

    <el-dialog v-model="columnDialogVisible" width="560px" :show-close="false" class="column-dialog">
      <template #header>
        <div class="dialog-heading">
          <strong>鍒楄缃</strong>
          <button class="dialog-close-button" type="button" @click="columnDialogVisible = false">
            <Close class="button-icon" />
          </button>
        </div>
      </template>

      <div class="column-table">
        <div class="column-row head">
          <span>{{ t('columns') }}</span>
          <span>{{ t('show') }}</span>
        </div>
        <label v-for="column in columnOptions" :key="column.key" class="column-row">
          <span>{{ column.label }}</span>
          <input v-model="visibleColumns[column.key]" type="checkbox" />
        </label>
      </div>

      <template #footer>
        <button class="toolbar-button accent" type="button" @click="columnDialogVisible = false">纭畾</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { ElDrawer, ElMessage, ElMessageBox } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { t } from '@/utils/language';
import {
  ArrowRight,
  Back,
  Close,
  CopyDocument,
  Delete,
  DocumentAdd,
  Download,
  EditPen,
  FolderAdd,
  FolderOpened,
  Grid,
  List,
  Picture,
  RefreshRight,
  Setting,
  Tickets,
  Upload,
} from '@element-plus/icons-vue';

import FileList from '@/components/FileList/index.vue';
import CreateFolderDialog from '@/components/CreateFolderDialog/index.vue';
import FolderSelectDialog from '@/components/FolderSelectDialog/index.vue';
import RenameDialog from '@/components/RenameDialog/index.vue';
import ShareDialog from '@/components/ShareDialog/index.vue';
import {
  createFile as apiCreateFile,
  downloadFile as apiDownloadFile,
  getFileCustomProperties,
  previewFileAsPdf,
  updateFileCustomProperties,
  type FileCustomPropertyDefinition,
} from '@/api/file';
import { getFileSystemClientSettings } from '@/api/file-system-settings';
import { getToken } from '@/utils/auth';
import type { FileItem } from '@/types/file';
import type { FolderItem } from '@/types/folder';
import { formatFileSize, formatTimestamp } from '@/utils/format';
import { useFileStore, type ViewMode } from '@/stores/file';
import { useUploadStore } from '@/stores/upload';

type EntryDetail =
  | ({ isFolder: false } & FileItem)
  | ({ isFolder: true } & FolderItem);
type WorkspaceFileKind = 'file' | 'markdown' | 'text' | 'drawio' | 'dwb' | 'excalidraw';

const router = useRouter();
const route = useRoute();
const fileStore = useFileStore();
const uploadStore = useUploadStore();
const detailDrawerVisible = ref(false);
const renameDialogVisible = ref(false);
const shareDialogVisible = ref(false);
const previewDialogVisible = ref(false);
const viewPanelVisible = ref(false);
const columnDialogVisible = ref(false);
const createFolderVisible = ref(false);
const contextMenuVisible = ref(false);
const drawioSubmenuVisible = ref(false);
const selectedEntry = ref<EntryDetail | null>(null);
const selectedKeys = ref<string[]>([]);
const folderSelectVisible = ref(false);
const folderSelectMode = ref<'move' | 'copy'>('move');
const folderSelectTargets = ref<EntryDetail[]>([]);
const workspaceFileInputRef = ref<HTMLInputElement | null>(null);
const workspaceFolderInputRef = ref<HTMLInputElement | null>(null);
const fileEventSource = ref<EventSource | null>(null);
const contextMenuPosition = reactive({ x: 0, y: 0 });
const activeViewMode = ref<ViewMode>(fileStore.viewMode);
const showThumbnails = ref(true);
const maxBatchActionSize = ref(3000);
const maxPageSize = ref(2000);
const gridItemSize = ref(216);
const galleryImageSize = ref(180);
const pageSize = ref(2000);
const visibleColumns = reactive({
  name: true,
  size: true,
  updatedAt: true,
});
const customPropertyState = reactive<{
  definitions: FileCustomPropertyDefinition[];
  values: Record<string, string>;
  loading: boolean;
  saving: boolean;
  lastModified: string;
}>({
  definitions: [],
  values: {},
  loading: false,
  saving: false,
  lastModified: '',
});
const previewState = reactive<{
  title: string;
  loading: boolean;
  kind: 'empty' | 'docx' | 'text' | 'image' | 'video' | 'audio' | 'pdf' | 'office' | 'unsupported';
  url: string;
  html: string;
  text: string;
  pdfPages: string[];
  error: string;
}>({
  title: '',
  loading: false,
  kind: 'empty',
  url: '',
  html: '',
  text: '',
  pdfPages: [],
  error: '',
});

const selectedFile = computed(() => (selectedEntry.value && !selectedEntry.value.isFolder ? selectedEntry.value : null));
const selectedFolder = computed(() => (selectedEntry.value && selectedEntry.value.isFolder ? selectedEntry.value : null));
const selectedItems = computed<EntryDetail[]>(() => {
  const folders = fileStore.folders
    .filter((item) => selectedKeys.value.includes(entryKey(item)))
    .map((item) => ({ ...item, isFolder: true as const }));
  const files = fileStore.files
    .filter((item) => selectedKeys.value.includes(entryKey(item)))
    .map((item) => ({ ...item, isFolder: false as const }));
  return [...folders, ...files];
});
const selectedFiles = computed(() => selectedItems.value.filter((item): item is EntryDetail & { isFolder: false } => !item.isFolder));
const selectedFolders = computed(() => selectedItems.value.filter((item): item is EntryDetail & { isFolder: true } => item.isFolder));
const folderSelectTitle = computed(() => (folderSelectMode.value === 'move' ? t('moveTo') : t('copyTo')));
const folderSelectExcludeId = computed(() => {
  if (folderSelectMode.value !== 'move' || folderSelectTargets.value.length !== 1) return null;
  const target = folderSelectTargets.value[0];
  return target?.isFolder ? target.id : null;
});
const currentFolderTitle = computed(() => fileStore.currentFolder?.name || t('myFiles'));
const folderSummary = computed(
  () =>
    `${t('currentFolderSummaryPrefix')} ${fileStore.folders.length} ${t('foldersCountUnit')}, ${fileStore.files.length} ${t(
      'filesCountUnit',
    )}. ${t('folderSummaryHint')}`,
);
const canGoBack = computed(() => fileStore.currentFolderId !== null);
const fileTypeLabel = computed(() => selectedFile.value?.browser_app?.name || t('fileDetails'));
const selectedFileMime = computed(() => selectedFile.value?.content_type || selectedFile.value?.mime_type || 'application/octet-stream');
const formattedSelectedFileSize = computed(() => (selectedFile.value ? formatFileSize(selectedFile.value.size) : '-'));
const formattedSelectedTime = computed(() => (selectedEntry.value ? formatTimestamp(selectedEntry.value.updated_at) : '-'));
const folderParentLabel = computed(() => {
  if (!selectedFolder.value) return '-';
  if (selectedFolder.value.parent_id === null) return t('rootFolder');
  return `Directory #${selectedFolder.value.parent_id}`;
});
const folderPathLabel = computed(() => {
  if (!selectedFolder.value) return '/';
  const base = fileStore.breadcrumb.map((item) => item.name);
  return `/${[...base, selectedFolder.value.name].join('/')}`;
});
const formattedFolderTotalSize = computed(() => {
  if (!selectedFolder.value?.directory_stats) return '-';
  return formatFileSize(selectedFolder.value.directory_stats.total_size || 0);
});
const folderCacheStatusLabel = computed(() => {
  if (!selectedFolder.value?.directory_stats) return t('cacheNotEnabled');
  if (!selectedFolder.value.directory_stats.cache_enabled) return t('cacheDisabled');
  return selectedFolder.value.directory_stats.cached ? t('cacheHit') : t('realtimeStats');
});

const viewOptions = computed<{ mode: ViewMode; label: string; icon: typeof Grid }[]>(() => [
  { mode: 'grid', label: t('grid'), icon: Grid },
  { mode: 'table', label: t('list'), icon: List },
  { mode: 'gallery', label: t('gallery'), icon: Picture },
]);

const columnOptions = computed<{ key: keyof typeof visibleColumns; label: string }[]>(() => [
  { key: 'name', label: t('name') },
  { key: 'size', label: t('size') },
  { key: 'updatedAt', label: t('updatedAt') },
]);

const viewModeIcon = computed(() => viewOptions.value.find((item) => item.mode === activeViewMode.value)?.icon || Grid);

function closeWorkspaceContextMenu() {
  contextMenuVisible.value = false;
  drawioSubmenuVisible.value = false;
}

function showWorkspaceContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.file-item') ||
    target?.closest('.gallery-item') ||
    target?.closest('.el-table__row') ||
    target?.closest('.el-dropdown') ||
    target?.closest('button')
  ) {
    return;
  }

  event.preventDefault();
  clearSelection();
  detailDrawerVisible.value = false;
  viewPanelVisible.value = false;

  contextMenuPosition.x = event.clientX;
  contextMenuPosition.y = event.clientY;
  contextMenuVisible.value = true;
  drawioSubmenuVisible.value = false;
}

function triggerWorkspaceFileUpload() {
  closeWorkspaceContextMenu();
  workspaceFileInputRef.value?.click();
}

function triggerWorkspaceFolderUpload() {
  closeWorkspaceContextMenu();
  workspaceFolderInputRef.value?.click();
}

async function handleWorkspaceFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  const files = Array.from(target.files || []);
  target.value = '';
  if (!files.length) return;

  try {
    for (const file of files) {
      await uploadStore.addTask(file, fileStore.currentFolderId ?? undefined);
    }
    ElMessage.success('Added ' + files.length + ' files to the upload queue.');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to add upload tasks.');
  }
}

async function uploadFromClipboard() {
  closeWorkspaceContextMenu();
  try {
    if (!navigator.clipboard?.read) {
      ElMessage.warning('This browser cannot read files from the clipboard.');
      return;
    }

    const clipboardItems = await navigator.clipboard.read();
    const files: File[] = [];

    for (const item of clipboardItems) {
      for (const type of item.types) {
        const blob = await item.getType(type);
        if (!blob.size) continue;
        const extension = clipboardExtension(type);
        files.push(new File([blob], uniqueWorkspaceName('clipboard-file.' + extension), { type }));
      }
    }

    if (!files.length) {
      ElMessage.warning('No uploadable file content was found in the clipboard.');
      return;
    }

    for (const file of files) {
      await uploadStore.addTask(file, fileStore.currentFolderId ?? undefined);
    }
    ElMessage.success('Added ' + files.length + ' clipboard files to the upload queue.');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to read clipboard.');
  }
}

function clipboardExtension(type: string) {
  if (type.includes('png')) return 'png';
  if (type.includes('jpeg') || type.includes('jpg')) return 'jpg';
  if (type.includes('gif')) return 'gif';
  if (type.includes('html')) return 'html';
  if (type.includes('plain')) return 'txt';
  return 'bin';
}

function uniqueWorkspaceName(baseName: string) {
  const existingNames = new Set([...fileStore.files, ...fileStore.folders].map((item) => item.name));
  if (!existingNames.has(baseName)) return baseName;

  const dotIndex = baseName.lastIndexOf('.');
  const stem = dotIndex > 0 ? baseName.slice(0, dotIndex) : baseName;
  const extension = dotIndex > 0 ? baseName.slice(dotIndex) : '';
  let index = 2;
  let nextName = stem + ' ' + index + extension;
  while (existingNames.has(nextName)) {
    index += 1;
    nextName = stem + ' ' + index + extension;
  }
  return nextName;
}

async function openOfflineDownload() {
  closeWorkspaceContextMenu();
  await router.push('/drive/offline-downloads');
}

function openWorkspaceCreateFolderDialog() {
  closeWorkspaceContextMenu();
  createFolderVisible.value = true;
}

async function handleCreateFolder(name: string) {
  try {
    await fileStore.createFolder(name);
    createFolderVisible.value = false;
    ElMessage.success('Folder created.');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to create folder.');
  }
}

async function createWorkspaceFile(kind: WorkspaceFileKind) {
  closeWorkspaceContextMenu();
  const defaultNames: Record<WorkspaceFileKind, string> = {
    file: '\u65b0\u5efa\u6587\u4ef6.txt',
    markdown: '\u65b0\u5efa Markdown.md',
    text: '\u65b0\u5efa\u6587\u672c.txt',
    drawio: '\u65b0\u5efa\u56fe\u8868.drawio',
    dwb: '\u65b0\u5efa\u767d\u677f.dwb',
    excalidraw: '\u65b0\u5efa Excalidraw.excalidraw',
  };

  try {
    await apiCreateFile(kind, uniqueWorkspaceName(defaultNames[kind]), fileStore.currentFolderId);
    await refreshFiles();
    ElMessage.success('File created.');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to create file.');
  }
}

async function refreshFromContextMenu() {
  closeWorkspaceContextMenu();
  await refreshFiles();
}

function setViewMode(mode: ViewMode) {
  activeViewMode.value = mode;
  fileStore.setViewMode(mode);
}

function entryKey(item: EntryDetail | FileItem | FolderItem) {
  return (isFolderItem(item) ? 'folder' : 'file') + ':' + item.id;
}

function asEntryDetail(item: FileItem | FolderItem): EntryDetail {
  return isFolderItem(item)
    ? { ...item, isFolder: true }
    : { ...item, isFolder: false };
}

function toggleSelection(item: FileItem | FolderItem) {
  const key = entryKey(item);
  selectedKeys.value = selectedKeys.value.includes(key)
    ? selectedKeys.value.filter((itemKey) => itemKey !== key)
    : [...selectedKeys.value, key];
}

function addSelection(item: FileItem | FolderItem) {
  const key = entryKey(item);
  if (selectedKeys.value.includes(key)) return;
  selectedKeys.value = [...selectedKeys.value, key];
}

function selectSingle(item: FileItem | FolderItem) {
  selectedKeys.value = [entryKey(item)];
  selectedEntry.value = asEntryDetail(item);
}

function clearSelection() {
  selectedKeys.value = [];
}

function isSelectedEntry(item: FileItem | FolderItem) {
  return selectedKeys.value.includes(entryKey(item));
}

function handleFilePrimaryClick(file: FileItem) {
  if (isSelectedEntry(file)) {
    openFileDetail(file);
    return;
  }
  addSelection(file);
}

function handleFolderPrimaryClick(folder: FolderItem) {
  if (isSelectedEntry(folder)) {
    openFolderDetail(folder);
    return;
  }
  addSelection(folder);
}

async function openFilePreview(file: FileItem) {
  selectedEntry.value = { ...file, isFolder: false };
  await openMatchedApp();
}

async function refreshFiles() {
  try {
    await fileStore.refresh();
    syncSelectedEntry();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '\u5237\u65b0\u6587\u4ef6\u5217\u8868\u5931\u8d25');
  }
}

async function navigateToRoot() {
  detailDrawerVisible.value = false;
  await fileStore.navigateToFolder(null);
}

async function navigateToFolder(folderId: number) {
  detailDrawerVisible.value = false;
  await fileStore.navigateToFolder(folderId);
}

async function navigateToParent() {
  detailDrawerVisible.value = false;
  await fileStore.navigateToParent();
}

async function applyRoutePreviewQuery() {
  const previewFileValue = Array.isArray(route.query.preview_file) ? route.query.preview_file[0] : route.query.preview_file;
  if (!previewFileValue) return;

  const folderValue = Array.isArray(route.query.folder_id) ? route.query.folder_id[0] : route.query.folder_id;
  const folderId = folderValue ? Number(folderValue) : null;
  const previewFileId = Number(previewFileValue);
  if (!Number.isFinite(previewFileId) || previewFileId <= 0) return;

  if (Number.isFinite(folderId as number)) {
    await fileStore.fetchFiles(folderId as number, pageSize.value);
  }

  const file = fileStore.files.find((item) => item.id === previewFileId);
  if (!file) {
    ElMessage.warning('The downloaded file was not found in this folder.');
    return;
  }

  await openFilePreview(file);
  void router.replace({ name: 'drive-my-files', query: folderValue ? { folder_id: folderValue } : {} });
}

async function handleFolderOpen(folder: FolderItem) {
  await navigateToFolder(folder.id);
}

async function handleDownload(file: FileItem) {
  try {
    await fileStore.downloadFile(file.id);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '\u6587\u4ef6\u4e0b\u8f7d\u5931\u8d25');
  }
}

function resetCustomPropertyState() {
  customPropertyState.definitions = [];
  customPropertyState.values = {};
  customPropertyState.loading = false;
  customPropertyState.saving = false;
  customPropertyState.lastModified = '';
}

async function openFileDetail(file: FileItem) {
  selectedEntry.value = { ...file, isFolder: false };
  detailDrawerVisible.value = true;
  customPropertyState.loading = true;

  try {
    const data = await getFileCustomProperties(file.id);
    customPropertyState.definitions = data.definitions || [];
    customPropertyState.values = { ...(data.values || {}) };
    customPropertyState.lastModified = data.last_modified || '';
  } catch (error) {
    resetCustomPropertyState();
    ElMessage.error(error instanceof Error ? error.message : '\u52a0\u8f7d\u6587\u4ef6\u8be6\u60c5\u5931\u8d25');
  } finally {
    customPropertyState.loading = false;
  }
}

function openFolderDetail(folder: FolderItem) {
  selectedEntry.value = { ...folder, isFolder: true };
  detailDrawerVisible.value = true;
  resetCustomPropertyState();
}

function openRenameDialog() {
  if (!selectedEntry.value) return;
  renameDialogVisible.value = true;
}

function isFolderItem(item: FileItem | FolderItem): item is FolderItem {
  return !('hash' in item);
}

function handleActionRename(item: FileItem | FolderItem) {
  selectedEntry.value = isFolderItem(item)
    ? { ...item, isFolder: true }
    : { ...item, isFolder: false };
  renameDialogVisible.value = true;
}

async function handleDelete(item: FileItem | FolderItem) {
  const isFolder = isFolderItem(item);
  try {
    await ElMessageBox.confirm(
      'Delete "' + item.name + '"? It will be moved to Trash.',
      isFolder ? 'Delete folder' : 'Delete file',
      {
        confirmButtonText: '\u5220\u9664',
        cancelButtonText: '\u53d6\u6d88',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    );

    if (isFolder) {
      await fileStore.deleteFolder(item.id);
    } else {
      await fileStore.deleteFile(item.id);
    }

    if (selectedEntry.value?.id === item.id) {
      detailDrawerVisible.value = false;
      selectedEntry.value = null;
    }
    ElMessage.success('\u5220\u9664\u6210\u529f');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '\u5220\u9664\u5931\u8d25');
  }
}

function openFolderSelect(mode: 'move' | 'copy', targets: EntryDetail[]) {
  if (targets.length === 0) return;
  folderSelectMode.value = mode;
  folderSelectTargets.value = targets;
  folderSelectVisible.value = true;
}

function openMoveDialog(item: FileItem | FolderItem) {
  openFolderSelect('move', [asEntryDetail(item)]);
}

function openCopyDialog(item: FileItem | FolderItem) {
  openFolderSelect('copy', [asEntryDetail(item)]);
}

function openBatchMoveDialog() {
  if (!validateSelectedBatchSize()) return;
  openFolderSelect('move', selectedItems.value);
}

function openBatchCopyDialog() {
  if (!validateSelectedBatchSize()) return;
  openFolderSelect('copy', selectedItems.value);
}

function validateSelectedBatchSize() {
  if (selectedItems.value.length > maxBatchActionSize.value) {
    ElMessage.warning('Batch actions cannot exceed ' + maxBatchActionSize.value + ' items.');
    return false;
  }
  return true;
}

async function confirmFolderOperation(folderId: number | null) {
  const targets = [...folderSelectTargets.value];
  try {
    for (const item of targets) {
      if (folderSelectMode.value === 'move') {
        if (item.isFolder) {
          await fileStore.moveFolder(item.id, folderId);
        } else {
          await fileStore.moveFile(item.id, folderId);
        }
      } else if (item.isFolder) {
        await fileStore.copyFolder(item.id, folderId);
      } else {
        await fileStore.copyFile(item.id, folderId);
      }
    }
    await fileStore.refresh();
    clearSelection();
    syncSelectedEntry();
    ElMessage.success(folderSelectMode.value === 'move' ? '\u79fb\u52a8\u6210\u529f' : '\u590d\u5236\u6210\u529f');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : (folderSelectMode.value === 'move' ? '\u79fb\u52a8\u5931\u8d25' : '\u590d\u5236\u5931\u8d25'));
  } finally {
    folderSelectTargets.value = [];
  }
}

async function downloadSelected() {
  try {
    for (const file of selectedFiles.value) {
      await fileStore.downloadFile(file.id);
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '\u4e0b\u8f7d\u5931\u8d25');
  }
}

async function deleteSelected() {
  if (selectedItems.value.length === 0) return;
  if (!validateSelectedBatchSize()) return;
  try {
    await ElMessageBox.confirm(
      'Delete ' + selectedItems.value.length + ' selected items? They will be moved to Trash.',
      'Delete selected items',
      {
        confirmButtonText: '\u5220\u9664',
        cancelButtonText: '\u53d6\u6d88',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    );

    for (const item of selectedItems.value) {
      if (item.isFolder) {
        await fileStore.deleteFolder(item.id);
      } else {
        await fileStore.deleteFile(item.id);
      }
    }

    detailDrawerVisible.value = false;
    selectedEntry.value = null;
    clearSelection();
    await fileStore.refresh();
    ElMessage.success('\u5220\u9664\u6210\u529f');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '\u5220\u9664\u5931\u8d25');
  }
}

function openShareForFile(file: FileItem) {
  selectedEntry.value = { ...file, isFolder: false };
  shareDialogVisible.value = true;
}

function openVersionHistory(_file: FileItem) {
  ElMessage.info('Version history is not available yet.');
}

function openCollaborate(_file: FileItem) {
  ElMessage.info('Collaboration management is not available yet.');
}

function openShareDialog() {
  if (!selectedFile.value) return;
  shareDialogVisible.value = true;
}

async function confirmRename(name: string) {
  if (!selectedEntry.value) return;

  try {
    if (selectedEntry.value.isFolder) {
      await fileStore.renameFolder(selectedEntry.value.id, name);
    } else {
      await fileStore.renameFile(selectedEntry.value.id, name);
    }
    renameDialogVisible.value = false;
    syncSelectedEntry(name);
    ElMessage.success('Renamed successfully.');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Rename failed.');
  }
}

function fileExtension(fileName: string) {
  const dotIndex = fileName.lastIndexOf('.');
  return dotIndex >= 0 ? fileName.slice(dotIndex + 1).toLowerCase() : '';
}

function isPlainTextFile(file: FileItem) {
  const mime = file.content_type || file.mime_type || '';
  const ext = fileExtension(file.name);
  return (
    mime.startsWith('text/') ||
    ['md', 'markdown', 'txt', 'json', 'xml', 'yaml', 'yml', 'csv', 'log', 'ini', 'css', 'js', 'ts', 'html', 'htm', 'go', 'py', 'java', 'c', 'cpp', 'h', 'sql', 'sh'].includes(ext)
  );
}

function isImageExtension(ext: string) {
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'avif'].includes(ext);
}

function canPreviewWithoutBrowserApp(file: FileItem) {
  const mime = file.content_type || file.mime_type || '';
  const ext = fileExtension(file.name);
  return (
    isImageExtension(ext) ||
    isPlainTextFile(file) ||
    mime === 'application/pdf' ||
    mime.startsWith('image/') ||
    mime.startsWith('video/') ||
    mime.startsWith('audio/') ||
    ['pdf', 'docx', 'pptx', 'ppt'].includes(ext)
  );
}

function syncFileSize(file: FileItem, size: number) {
  if (!Number.isFinite(size) || size <= 0 || file.size === size) return;
  file.size = size;

  const storeFile = fileStore.files.find((item) => item.id === file.id);
  if (storeFile) {
    storeFile.size = size;
  }

  if (selectedEntry.value && !selectedEntry.value.isFolder && selectedEntry.value.id === file.id) {
    selectedEntry.value = { ...selectedEntry.value, size };
  }
}

async function renderPdfPages(blob: Blob) {
  const [pdfjsLib, pdfWorker] = await Promise.all([
    import('pdfjs-dist/legacy/build/pdf.mjs'),
    import('pdfjs-dist/legacy/build/pdf.worker.mjs?url'),
  ]);
  pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker.default;

  const pdf = await pdfjsLib.getDocument({ data: await blob.arrayBuffer() }).promise;
  const pages: string[] = [];
  const maxPages = Math.min(pdf.numPages, 20);

  for (let pageNumber = 1; pageNumber <= maxPages; pageNumber += 1) {
    const page = await pdf.getPage(pageNumber);
    const viewport = page.getViewport({ scale: 1.45 });
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d');
    if (!context) continue;

    canvas.width = Math.ceil(viewport.width);
    canvas.height = Math.ceil(viewport.height);
    await page.render({ canvas, canvasContext: context, viewport }).promise;
    pages.push(canvas.toDataURL('image/png'));
  }

  if (pages.length === 0) {
    throw new Error('PDF \u6ca1\u6709\u53ef\u6e32\u67d3\u7684\u9875\u9762');
  }

  return pages;
}

function resetPreviewState() {
  if (previewState.url) {
    URL.revokeObjectURL(previewState.url);
  }
  previewState.title = '';
  previewState.loading = false;
  previewState.kind = 'empty';
  previewState.url = '';
  previewState.html = '';
  previewState.text = '';
  previewState.pdfPages = [];
  previewState.error = '';
}

function clearPreview() {
  resetPreviewState();
}

async function openMatchedApp() {
  if (!selectedFile.value) return;
  if (!selectedFile.value.browser_app && isImageExtension(fileExtension(selectedFile.value.name))) {
    selectedFile.value.browser_app = { name: '\u56fe\u7247\u9884\u89c8' } as any;
  }
  if (!selectedFile.value.browser_app && !canPreviewWithoutBrowserApp(selectedFile.value)) {
    ElMessage.warning('No preview app is configured for this file yet.');
    return;
  }

  const file = selectedFile.value;
  resetPreviewState();
  previewState.title = file.name;
  previewState.loading = true;
  previewDialogVisible.value = true;

  try {
    const ext = fileExtension(file.name);
    if (ext === 'pptx' || ext === 'ppt') {
      const pdfBlob = await previewFileAsPdf(file.id);
      previewState.kind = 'pdf';
      previewState.pdfPages = await renderPdfPages(pdfBlob);
      return;
    }

    const blob = await apiDownloadFile(file.id);
    syncFileSize(file, blob.size);
    if (blob.size <= 0) {
      previewState.kind = 'unsupported';
      previewState.error = 'The downloaded content is empty. Please upload the original file again.';
      return;
    }

    const mime = blob.type || file.content_type || file.mime_type || '';

    if (ext === 'docx') {
      const mammoth = await import('mammoth');
      const arrayBuffer = await blob.arrayBuffer();
      const result = await mammoth.convertToHtml({ arrayBuffer });
      previewState.kind = 'docx';
      previewState.html = result.value || '<p>\u6587\u6863\u6ca1\u6709\u53ef\u663e\u793a\u7684\u6b63\u6587\u5185\u5bb9\u3002</p>';
    } else if (mime === 'application/pdf' || ext === 'pdf') {
      previewState.kind = 'pdf';
      previewState.pdfPages = await renderPdfPages(blob);
    } else if (mime.startsWith('image/') || isImageExtension(ext)) {
      previewState.kind = 'image';
      previewState.url = URL.createObjectURL(blob);
    } else if (mime.startsWith('video/')) {
      previewState.kind = 'video';
      previewState.url = URL.createObjectURL(blob);
    } else if (mime.startsWith('audio/')) {
      previewState.kind = 'audio';
      previewState.url = URL.createObjectURL(blob);
    } else if (isPlainTextFile(file)) {
      previewState.kind = 'text';
      previewState.text = await blob.text();
    } else {
      previewState.kind = 'unsupported';
      previewState.error = 'This file type cannot be previewed in the app yet. Download it to view.';
    }
  } catch (error) {
    previewState.kind = 'unsupported';
    previewState.error = error instanceof Error ? error.message : '\u52a0\u8f7d\u9884\u89c8\u5931\u8d25';
  } finally {
    previewState.loading = false;
  }
}

async function openFolderFromDrawer() {
  if (!selectedFolder.value) return;
  await navigateToFolder(selectedFolder.value.id);
}

function propertyTypeLabel(type: FileCustomPropertyDefinition['type']) {
  switch (type) {
    case 'rating':
      return '\u8bc4\u5206';
    case 'switch':
      return 'Switch';
    case 'date':
      return '\u65e5\u671f';
    case 'tags':
      return '\u6807\u7b7e';
    case 'multi_select':
      return 'Multi select';
    default:
      return '\u6587\u672c';
  }
}

function textFieldRuleLabel(item: FileCustomPropertyDefinition) {
  const min = typeof item.minLength === 'number' ? item.minLength : null;
  const max = typeof item.maxLength === 'number' ? item.maxLength : null;
  if (min === null && max === null) return '\u957f\u5ea6\u4e0d\u9650';
  return 'Length range: ' + (min ?? 0) + ' - ' + (max ?? '∞');
}

function selectedArrayValues(key: string) {
  const rawValue = customPropertyState.values[key];
  if (!rawValue) return [];

  try {
    const parsed = JSON.parse(rawValue);
    if (Array.isArray(parsed)) {
      return parsed.map((item) => String(item).trim()).filter(Boolean);
    }
  } catch {
    return rawValue
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean);
  }

  return [];
}

function writeArrayValues(key: string, values: string[]) {
  const unique = Array.from(new Set(values.map((item) => item.trim()).filter(Boolean)));
  customPropertyState.values[key] = JSON.stringify(unique);
}

function toggleArrayOption(key: string, option: string) {
  const current = selectedArrayValues(key);
  if (current.includes(option)) {
    writeArrayValues(
      key,
      current.filter((item) => item !== option),
    );
    return;
  }
  writeArrayValues(key, [...current, option]);
}

function setArrayInput(key: string, raw: string) {
  writeArrayValues(key, raw.split(','));
}

function setSwitchValue(key: string, checked: boolean) {
  customPropertyState.values[key] = checked ? 'true' : 'false';
}

async function saveFileCustomProperties() {
  if (!selectedFile.value) return;
  customPropertyState.saving = true;
  try {
    const data = await updateFileCustomProperties(selectedFile.value.id, customPropertyState.values);
    customPropertyState.definitions = data.definitions || [];
    customPropertyState.values = { ...(data.values || {}) };
    customPropertyState.lastModified = data.last_modified || '';
    ElMessage.success('\u6587\u4ef6\u81ea\u5b9a\u4e49\u5c5e\u6027\u5df2\u4fdd\u5b58');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to save file properties.');
  } finally {
    customPropertyState.saving = false;
  }
}

function syncSelectedEntry(nextName?: string) {
  if (!selectedEntry.value) return;

  if (selectedEntry.value.isFolder) {
    const folder = fileStore.folders.find((item) => item.id === selectedEntry.value?.id);
    if (!folder) {
      selectedEntry.value = nextName
        ? { ...selectedEntry.value, name: nextName, isFolder: true }
        : selectedEntry.value;
      return;
    }
    selectedEntry.value = { ...folder, isFolder: true };
    return;
  }

  const file = fileStore.files.find((item) => item.id === selectedEntry.value?.id);
  if (!file) {
    selectedEntry.value = nextName
      ? { ...selectedEntry.value, name: nextName, isFolder: false }
      : selectedEntry.value;
    return;
  }
  selectedEntry.value = { ...file, isFolder: false };
}

onMounted(async () => {
  document.addEventListener('click', closeWorkspaceContextMenu);
  window.addEventListener('resize', closeWorkspaceContextMenu);
  try {
    getFileSystemClientSettings()
      .then((settings) => {
        if (settings.max_batch_action_size > 0) {
          maxBatchActionSize.value = settings.max_batch_action_size;
        }
        if (settings.max_page_size > 0) {
          maxPageSize.value = settings.max_page_size;
          if (pageSize.value > maxPageSize.value) {
            pageSize.value = maxPageSize.value;
          }
        }
        if (settings.enable_event_push) {
          connectFileEvents();
        }
      })
      .catch(() => {});
    const folderValue = Array.isArray(route.query.folder_id) ? route.query.folder_id[0] : route.query.folder_id;
    const folderId = folderValue ? Number(folderValue) : null;
    await fileStore.fetchFiles(Number.isFinite(folderId as number) ? (folderId as number) : null, pageSize.value);
    await applyRoutePreviewQuery();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to initialize file list.');
  }
});

onBeforeUnmount(() => {
  document.removeEventListener('click', closeWorkspaceContextMenu);
  window.removeEventListener('resize', closeWorkspaceContextMenu);
  closeFileEvents();
  resetPreviewState();
});

function connectFileEvents() {
  closeFileEvents();
  const token = getToken();
  if (!token || typeof EventSource === 'undefined') return;
  const url = new URL('/api/v1/file/events', window.location.origin);
  url.searchParams.set('access_token', token);
  fileEventSource.value = new EventSource(url.toString());
  fileEventSource.value.addEventListener('file-change', () => {
    void fileStore.fetchFiles(fileStore.currentFolderId, pageSize.value).then(() => syncSelectedEntry()).catch(() => {});
  });
  fileEventSource.value.onerror = () => {
    closeFileEvents();
  };
}

function closeFileEvents() {
  if (!fileEventSource.value) return;
  fileEventSource.value.close();
  fileEventSource.value = null;
}

watch(pageSize, async (next, previous) => {
  if (next === previous || fileStore.loading) return;
  try {
    await fileStore.fetchFiles(fileStore.currentFolderId, next);
    syncSelectedEntry();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to refresh files.');
  }
});
</script>

<style scoped>
.workspace-page {
  display: grid;
  gap: 18px;
  padding: 0 4px 24px;
  color: #10213f;
}

.workspace-hero,
.workspace-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.84);
  border-radius: 34px;
  background:
    radial-gradient(circle at 94% 12%, rgba(102, 219, 255, 0.26), transparent 28%),
    radial-gradient(circle at 12% 0%, rgba(86, 138, 255, 0.14), transparent 24%),
    radial-gradient(circle at 74% 104%, rgba(255, 197, 220, 0.22), transparent 32%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(247, 252, 255, 0.6));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 28px 70px rgba(93, 133, 180, 0.14),
    0 12px 28px rgba(86, 178, 226, 0.08);
  backdrop-filter: blur(22px);
}

.workspace-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  min-height: 300px;
  padding: 48px 46px 44px;
}

.workspace-hero::before,
.workspace-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.72), transparent 30%),
    linear-gradient(315deg, rgba(255, 255, 255, 0.34), transparent 36%);
  pointer-events: none;
}

.hero-ambient {
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
  filter: blur(12px);
}

.hero-ambient-one {
  right: 48px;
  top: 38px;
  width: 210px;
  height: 210px;
  background: rgba(126, 218, 255, 0.28);
}

.hero-ambient-two {
  right: 290px;
  bottom: -68px;
  width: 280px;
  height: 140px;
  background: rgba(255, 197, 220, 0.2);
}

.hero-copy {
  position: relative;
  z-index: 1;
  max-width: 760px;
}

.hero-kicker,
.panel-kicker {
  margin: 0 0 10px;
  color: #2684e8;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.hero-copy h1 {
  margin: 0;
  color: #0d1b33;
  font-size: clamp(34px, 3.2vw, 52px);
  line-height: 1.06;
  letter-spacing: 0;
  max-width: 830px;
}

.hero-copy p {
  margin: 14px 0 0;
  color: #536987;
  font-size: 16px;
  line-height: 1.8;
}

.hero-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}

.hero-metrics span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.52);
  color: #58708d;
  font-size: 13px;
  font-weight: 720;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88);
}

.hero-metrics strong {
  color: #15345e;
}

.hero-actions,
.toolbar-actions,
.detail-actions,
.detail-primary-actions {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.view-control {
  position: relative;
  z-index: 5;
  isolation: isolate;
}

.view-panel {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  z-index: 20;
  display: grid;
  gap: 14px;
  width: 330px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 24px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.16), transparent 34%),
    rgba(255, 255, 255, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 24px 54px rgba(73, 112, 160, 0.22);
  backdrop-filter: blur(18px);
  pointer-events: auto;
}

.view-panel-title {
  margin: 0;
  color: #657892;
  font-size: 14px;
  font-weight: 820;
}

.view-segment {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid rgba(207, 223, 242, 0.78);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.56);
}

.view-segment.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.view-option {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 50px;
  border: 0;
  border-right: 1px solid rgba(207, 223, 242, 0.78);
  background: transparent;
  color: #657892;
  font-weight: 820;
  cursor: pointer;
  pointer-events: auto;
  user-select: none;
}

.view-option:last-child {
  border-right: 0;
}

.view-option.active {
  color: #1d70da;
  background: linear-gradient(180deg, rgba(232, 245, 255, 0.96), rgba(211, 234, 255, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 10px 22px rgba(64, 147, 226, 0.12);
}

.panel-wide-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 52px;
  width: 100%;
  border: 1px solid rgba(207, 223, 242, 0.78);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.56);
  color: #657892;
  font-weight: 820;
  cursor: pointer;
}

.panel-wide-button:hover {
  color: #1d70da;
  border-color: rgba(142, 212, 255, 0.66);
}

.slider-field {
  display: grid;
  gap: 10px;
  color: #657892;
  font-size: 14px;
  font-weight: 820;
}

.slider-field input {
  width: 100%;
  accent-color: #247be8;
}

.slider-field small {
  display: flex;
  justify-content: space-between;
  color: #7d8da6;
  font-size: 12px;
}

.hero-button,
.toolbar-button,
.crumb-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 56px;
  padding: 0 22px;
  border: 1px solid rgba(255, 255, 255, 0.84);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.62);
  color: #172846;
  font-size: 14px;
  font-weight: 800;
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease;
}

.hero-button:hover,
.toolbar-button:hover,
.crumb-button:hover {
  transform: translateY(-1px);
  border-color: rgba(120, 205, 255, 0.72);
  box-shadow: 0 16px 32px rgba(45, 127, 240, 0.16);
}

.hero-button.primary,
.toolbar-button.accent,
.save-button,
.accent-button {
  border-color: rgba(61, 142, 255, 0.66);
  background: linear-gradient(135deg, #2d70ff 0%, #19aeea 100%);
  color: #fff;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.3),
    0 18px 34px rgba(45, 127, 240, 0.28);
}

.button-icon {
  width: 16px;
  height: 16px;
}

.workspace-card {
  padding: 34px 34px 24px;
  min-height: calc(100vh - 170px);
  overflow: visible;
}

.workspace-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.selection-toolbar {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 18px 0 0;
  padding: 12px 14px;
  border: 1px solid rgba(207, 223, 242, 0.86);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.66);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 14px 28px rgba(82, 128, 178, 0.1);
}

.selection-count {
  min-height: 42px;
  display: inline-flex;
  align-items: center;
  padding: 0 18px;
  border-left: 1px solid rgba(207, 223, 242, 0.86);
  border-right: 1px solid rgba(207, 223, 242, 0.86);
  color: #10213f;
  font-weight: 820;
}

.selection-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selection-icon-button {
  display: grid;
  place-items: center;
  width: 46px;
  height: 42px;
  border: 1px solid rgba(207, 223, 242, 0.86);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  color: #172846;
  cursor: pointer;
}

.selection-icon-button:hover:not(:disabled) {
  border-color: rgba(45, 112, 255, 0.48);
  color: #1d70da;
  background: rgba(235, 246, 255, 0.92);
}

.selection-icon-button:disabled {
  opacity: 0.48;
  cursor: not-allowed;
}

.selection-icon-button.danger:hover {
  border-color: rgba(239, 68, 68, 0.42);
  color: #dc2626;
  background: rgba(254, 242, 242, 0.92);
}

.toolbar-copy {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 6px;
}

.toolbar-kicker {
  color: #2684e8;
  font-size: 12px;
  font-weight: 820;
  letter-spacing: 0.12em;
}

.toolbar-copy strong {
  color: #10213f;
  font-size: 26px;
  line-height: 1.1;
}

.toolbar-copy span,
.detail-status {
  color: #657892;
  font-size: 13px;
}

.toolbar-segment {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px;
  border: 1px solid rgba(207, 223, 242, 0.8);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.52);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.84);
}

.segment-button {
  min-height: 42px;
  padding: 0 16px;
  border: none;
  border-radius: 14px;
  background: transparent;
  color: #657892;
  font-weight: 800;
  cursor: pointer;
}

.segment-button.active {
  color: #1d70da;
  background: linear-gradient(180deg, rgba(232, 245, 255, 0.96), rgba(211, 234, 255, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 10px 22px rgba(64, 147, 226, 0.12);
}

.dialog-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.dialog-heading strong {
  color: #10213f;
  font-size: 22px;
  font-weight: 860;
}

.dialog-close-button {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.62);
  color: #657892;
  cursor: pointer;
}

.column-table {
  overflow: hidden;
  border: 1px solid rgba(219, 234, 249, 0.78);
  border-radius: 18px;
}

.column-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 92px;
  align-items: center;
  min-height: 54px;
  padding: 0 16px;
  border-bottom: 1px solid rgba(219, 234, 249, 0.78);
  color: #24344f;
  font-weight: 760;
}

.column-row:last-child {
  border-bottom: 0;
}

.column-row.head {
  min-height: 46px;
  background: rgba(245, 250, 255, 0.76);
  color: #657892;
  font-size: 13px;
  font-weight: 820;
}

.column-row input {
  width: 18px;
  height: 18px;
  accent-color: #247be8;
}

:deep(.column-dialog .el-dialog) {
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 26px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.14), transparent 34%),
    rgba(255, 255, 255, 0.92);
  box-shadow: 0 28px 68px rgba(73, 112, 160, 0.2);
  backdrop-filter: blur(18px);
}

:deep(.file-preview-dialog .el-dialog) {
  display: flex;
  max-height: calc(100vh - 10vh);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.88);
  border-radius: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.15), transparent 34%),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 30px 80px rgba(73, 112, 160, 0.24);
  backdrop-filter: blur(20px);
}

:deep(.file-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.file-preview-dialog .el-dialog__body) {
  min-height: 0;
  padding: 0 24px 24px;
}

.preview-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.preview-heading p {
  margin: 0 0 6px;
  color: #2684e8;
  font-size: 12px;
  font-weight: 820;
  letter-spacing: 0.12em;
}

.preview-heading strong {
  color: #10213f;
  font-size: 20px;
}

.preview-shell {
  overflow: auto;
  min-height: min(520px, calc(100vh - 210px));
  max-height: calc(100vh - 210px);
  border: 1px solid rgba(219, 234, 249, 0.82);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.78);
}

.preview-empty {
  display: grid;
  min-height: 520px;
  place-items: center;
  padding: 32px;
  color: #657892;
  font-weight: 800;
  text-align: center;
}

.docx-preview {
  max-width: 820px;
  min-height: 520px;
  margin: 0 auto;
  padding: 56px 64px;
  background: #fff;
  color: #111827;
  box-shadow: 0 0 0 1px rgba(226, 232, 240, 0.9);
  line-height: 1.85;
}

.docx-preview :deep(p) {
  margin: 0 0 12px;
}

.docx-preview :deep(img) {
  max-width: 100%;
  height: auto;
}

.text-preview {
  min-height: 520px;
  margin: 0;
  padding: 24px;
  color: #172846;
  font: 500 14px/1.8 Consolas, 'SFMono-Regular', monospace;
  white-space: pre-wrap;
  word-break: break-word;
}

.media-preview {
  display: block;
  width: 100%;
  min-height: 520px;
  border: 0;
  background: #0f172a;
}

.office-preview {
  display: block;
  width: 100%;
  min-height: 620px;
  border: 0;
  background: #fff;
}

.pdf-preview {
  display: grid;
  gap: 22px;
  justify-items: center;
  min-height: 520px;
  padding: 28px;
  background: #0f172a;
}

.pdf-preview img {
  display: block;
  width: min(100%, 920px);
  height: auto;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.28);
}

.image-preview {
  min-height: 0;
  max-height: 72vh;
  object-fit: contain;
  background: rgba(15, 23, 42, 0.92);
}

.audio-preview {
  width: calc(100% - 48px);
  margin: 220px 24px 0;
}

.workspace-breadcrumb {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin: 22px 0 26px;
}

.crumb-button {
  min-height: 42px;
  padding: 0 14px;
  border-radius: 14px;
  font-size: 13px;
}

.root-crumb {
  background: rgba(255, 255, 255, 0.72);
}

.crumb-separator {
  color: #94a3b8;
}

.workspace-list-shell {
  position: relative;
  z-index: 1;
  min-height: 520px;
}

.hidden-file-input {
  position: fixed;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.workspace-context-menu,
.workspace-context-submenu {
  position: fixed;
  z-index: 3000;
  width: 252px;
  padding: 6px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background:
    radial-gradient(circle at 8% 0%, rgba(118, 196, 255, 0.22), transparent 34%),
    radial-gradient(circle at 100% 20%, rgba(255, 196, 221, 0.2), transparent 32%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.9), rgba(246, 252, 255, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    inset 0 -1px 0 rgba(197, 224, 247, 0.5),
    0 18px 42px rgba(63, 114, 174, 0.18),
    0 8px 20px rgba(255, 168, 205, 0.08);
  color: #23272f;
  font-family:
    "Microsoft YaHei",
    "PingFang SC",
    system-ui,
    sans-serif;
  backdrop-filter: blur(22px) saturate(1.25);
}

.workspace-context-menu {
  left: 0;
  top: 0;
}

.workspace-context-submenu {
  position: absolute;
  left: calc(100% + 8px);
  top: 242px;
  width: 172px;
  padding: 6px;
}

.context-menu-section {
  padding: 4px 0;
  border-bottom: 1px solid rgba(206, 221, 238, 0.72);
}

.context-menu-section:last-of-type {
  border-bottom: 0;
}

.context-menu-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 14px;
  align-items: center;
  width: 100%;
  min-height: 34px;
  padding: 0 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #22304a;
  font-size: 14px;
  font-weight: 720;
  line-height: 1.2;
  text-align: left;
  cursor: pointer;
  transition:
    background 0.14s ease,
    box-shadow 0.14s ease,
    transform 0.14s ease;
}

.context-menu-item span {
  min-width: 0;
}

.context-menu-item:hover,
.context-menu-item.active {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.82), rgba(232, 245, 255, 0.68)),
    rgba(255, 255, 255, 0.52);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 8px 18px rgba(70, 136, 210, 0.12);
  transform: translateX(1px);
}

.context-menu-icon,
.context-app-icon {
  width: 20px;
  height: 20px;
  color: #6f7b8d;
}

.context-app-icon {
  display: inline-grid;
  place-items: center;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 900;
  line-height: 1;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.32),
    0 5px 10px rgba(55, 89, 130, 0.12);
}

.context-app-icon.markdown {
  background: #3f3f46;
  color: #fff;
}

.context-app-icon.drawio {
  background: linear-gradient(135deg, #ff9c1a, #f06c00);
  color: #fff;
}

.context-app-icon.text {
  color: #1682c5;
}

.context-app-icon.excalidraw {
  color: #6957d9;
}

.context-submenu-arrow {
  width: 14px;
  height: 14px;
  justify-self: end;
  color: #686c72;
}

.workspace-context-submenu .context-menu-item {
  grid-template-columns: minmax(0, 1fr);
  padding: 0 14px;
}

.detail-shell {
  display: grid;
  gap: 18px;
  padding: 18px;
}

.detail-hero,
.detail-panel {
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 24px;
  background: linear-gradient(180deg, #ffffff, #f8fafc);
  padding: 20px;
}

.detail-hero-badge,
.panel-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.1);
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 800;
}

.panel-pill.success {
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
}

.detail-hero h2,
.panel-head h3 {
  margin: 12px 0 6px;
  color: #0f172a;
}

.detail-hero p,
.browser-app-copy span,
.browser-app-copy small {
  color: #64748b;
}

.detail-stat-grid,
.folder-info-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 18px 0 0;
}

.detail-stat-card,
.folder-info-card {
  padding: 14px 16px;
  border-radius: 18px;
  background: #f8fafc;
  border: 1px solid rgba(226, 232, 240, 0.9);
  display: grid;
  gap: 8px;
}

.detail-stat-card span,
.folder-info-card span,
.detail-field-head span {
  color: #64748b;
  font-size: 12px;
}

.detail-stat-card strong,
.folder-info-card strong {
  color: #0f172a;
  font-size: 16px;
}

.detail-primary-actions {
  margin-top: 18px;
}

.panel-head,
.detail-field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.browser-app-card {
  display: flex;
  gap: 14px;
  margin-top: 16px;
  padding: 16px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(239, 246, 255, 0.9), rgba(248, 250, 252, 0.95));
  border: 1px solid rgba(191, 219, 254, 0.8);
}

.browser-app-mark {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  font-weight: 900;
  color: #fff;
  background: linear-gradient(135deg, var(--app-accent), #0f172a);
}

.browser-app-copy {
  display: grid;
  gap: 6px;
}

.encryption-card {
  display: grid;
  gap: 8px;
  margin-top: 16px;
  padding: 16px;
  border-radius: 20px;
  background: #f8fafc;
  border: 1px solid rgba(226, 232, 240, 0.9);
}

.encryption-card.active {
  background: linear-gradient(135deg, rgba(236, 253, 245, 0.92), rgba(240, 253, 250, 0.96));
  border-color: rgba(110, 231, 183, 0.85);
}

.encryption-card strong {
  color: #0f172a;
}

.encryption-card span {
  color: #64748b;
}

.panel-empty {
  margin-top: 16px;
  padding: 18px;
  border-radius: 18px;
  background: #f8fafc;
  color: #64748b;
}

.detail-form {
  display: grid;
  gap: 14px;
  margin-top: 16px;
}

.detail-field-card {
  padding: 16px;
  border-radius: 20px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: #fff;
}

.detail-field-type {
  color: #2563eb;
  font-size: 12px;
  font-style: normal;
  font-weight: 800;
}

.detail-textarea,
.detail-input {
  width: 100%;
  margin-top: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 14px;
  background: #fff;
  font-size: 14px;
  outline: none;
}

.detail-textarea:focus,
.detail-input:focus {
  border-color: rgba(37, 99, 235, 0.6);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}

.detail-field-help {
  display: inline-block;
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
}

.rating-row,
.option-chip-wrap {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.rating-star,
.option-chip {
  min-width: 42px;
  min-height: 42px;
  padding: 0 14px;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 14px;
  background: #fff;
  cursor: pointer;
  transition: all 0.18s ease;
}

.rating-star.active,
.option-chip.active {
  border-color: rgba(37, 99, 235, 0.4);
  background: rgba(37, 99, 235, 0.1);
  color: #1d4ed8;
}

.switch-row {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
  color: #0f172a;
  font-weight: 700;
}

.detail-actions {
  margin-top: 18px;
  align-items: center;
  justify-content: space-between;
}

@media (max-width: 960px) {
  .workspace-hero,
  .workspace-toolbar,
  .detail-actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .detail-stat-grid,
  .folder-info-grid {
    grid-template-columns: 1fr;
  }
}
</style>


