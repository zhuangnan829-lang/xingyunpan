<template>
  <section class="file-system-page">
    <header class="hero-card">
      <div class="hero-copy">
        <p class="eyebrow">File System Console</p>
        <h1>文件系统参数设置</h1>
        <p class="hero-text">
          统一管理参数设置、全文搜索、文件图标、文件浏览应用和自定义属性。当前页面直接连接真实后端接口，保存后会立即持久化。
        </p>
        <div class="hero-tags">
          <span class="hero-tag">{{ activeTabMeta.label }}</span>
          <span class="hero-tag hero-tag-soft">{{ busy ? '处理中' : '已连接真实配置' }}</span>
        </div>
      </div>

      <div class="hero-metrics">
        <article class="metric-card accent-blue">
          <span>当前页签</span>
          <strong>{{ activeTabMeta.label }}</strong>
          <small>支持独立保存与恢复</small>
        </article>
        <article class="metric-card accent-gold">
          <span>图标规则</span>
          <strong>{{ iconRules.length }}</strong>
          <small>扩展名、MIME 与默认图标</small>
        </article>
        <article class="metric-card accent-green">
          <span>浏览应用</span>
          <strong>{{ browserApps.length }}</strong>
          <small>刷新后仍会保留</small>
        </article>
        <article class="metric-card accent-cyan">
          <span>自定义属性</span>
          <strong>{{ customProperties.length }}</strong>
          <small>文件详情抽屉字段</small>
        </article>
      </div>
    </header>

    <section
      v-if="connectionAlert.visible"
      class="connection-alert"
      :class="connectionAlert.kind === 'success' ? 'is-success' : 'is-error'"
    >
      <div class="connection-alert-icon">
        <svg v-if="connectionAlert.kind === 'success'" viewBox="0 0 20 20" aria-hidden="true">
          <path d="M4.5 10.5 8 14l7.5-8" />
        </svg>
        <svg v-else viewBox="0 0 20 20" aria-hidden="true">
          <path d="M10 6v4.5" />
          <circle cx="10" cy="13.5" r="0.9" fill="currentColor" stroke="none" />
          <path d="M10 3.8 17 16.2H3L10 3.8Z" />
        </svg>
      </div>

      <div class="connection-alert-copy">
        <strong>{{ connectionAlert.message }}</strong>
        <span>{{ connectionAlert.detail }}</span>
      </div>

      <div class="connection-alert-actions">
        <button class="ghost-button connection-alert-button" type="button" :disabled="loading" @click="reloadAll">
          {{ loading ? '检测中...' : '重新检测连接' }}
        </button>
        <button class="ghost-button connection-alert-button" type="button" @click="connectionAlert.visible = false">关闭提示</button>
      </div>
    </section>

    <section class="panel-card">
      <div class="tabs-row">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="tab-button"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          <el-icon class="tab-icon"><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </button>
      </div>

      <div class="panel-body">
        <template v-if="activeTab === 'parameters'">
          <section class="section-header settings-header">
            <div class="settings-heading">
              <span class="settings-kicker">File System Settings</span>
              <h2>参数设置</h2>
              <p>管理文件系统的编辑、回收、列表、缓存、传输与安全策略。</p>
            </div>
          </section>

          <section class="overview-grid settings-metrics-grid">
            <article class="overview-card settings-metric-card">
              <span>在线编辑上限</span>
              <strong>{{ fileSystemForm.online_editor_size }} {{ fileSystemForm.online_editor_unit }}</strong>
              <small>超出阈值后关闭在线编辑入口</small>
            </article>
            <article class="overview-card settings-metric-card">
              <span>分页模式</span>
              <strong>{{ fileSystemForm.list_pagination_mode }}</strong>
              <small>最大每页 {{ fileSystemForm.max_page_size }} 条</small>
            </article>
            <article class="overview-card settings-metric-card">
              <span>传输并行数</span>
              <strong>{{ fileSystemForm.transfer_parallelism }}</strong>
              <small>分片重试上限 {{ fileSystemForm.max_chunk_retry }} 次</small>
            </article>
            <article class="overview-card settings-metric-card">
              <span>主密钥存储</span>
              <strong>{{ fileSystemForm.master_key_storage }}</strong>
              <small>加密状态 {{ fileSystemForm.show_encryption_status ? '显示' : '隐藏' }}</small>
            </article>
          </section>

          <section class="group-grid settings-stack">
            <article class="group-card">
              <div class="group-head settings-section-head">
                <h3>编辑与回收</h3>
                <p>控制在线编辑、回收站和 Blob 清理节奏。</p>
              </div>
              <div class="form-grid settings-form-list two-column">
                <label class="field-card">
                  <span>在线编辑最大文件</span>
                  <div class="split-row">
                    <input v-model.number="fileSystemForm.online_editor_size" type="number" class="field-input" />
                    <select v-model="fileSystemForm.online_editor_unit" class="field-select">
                      <option v-for="unit in units" :key="unit" :value="unit">{{ unit }}</option>
                    </select>
                  </div>
                </label>
                <label class="field-card">
                  <span>回收站扫描间隔</span>
                  <input v-model="fileSystemForm.recycle_scan_interval" type="text" class="field-input" />
                </label>
                <label class="field-card">
                  <span>Blob 回收间隔</span>
                  <input v-model="fileSystemForm.blob_recycle_interval" type="text" class="field-input" />
                </label>
                <label class="field-card">
                  <span>静态缓存有效期（秒）</span>
                  <input v-model.number="fileSystemForm.static_cache_ttl" type="number" class="field-input" />
                </label>
              </div>
            </article>

            <article class="group-card">
              <div class="group-head settings-section-head">
                <h3>列表与地图</h3>
                <p>优化目录浏览、递归搜索和地图展示体验。</p>
              </div>
              <div class="form-grid settings-form-list two-column">
                <label class="field-card">
                  <span>分页模式</span>
                  <select v-model="fileSystemForm.list_pagination_mode" class="field-select">
                    <option value="cursor">cursor</option>
                    <option value="offset">offset</option>
                    <option value="hybrid">hybrid</option>
                  </select>
                </label>
                <label class="field-card">
                  <span>最大单页数量</span>
                  <input v-model.number="fileSystemForm.max_page_size" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>最大批量操作数</span>
                  <input v-model.number="fileSystemForm.max_batch_action_size" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>最大递归搜索深度</span>
                  <input v-model.number="fileSystemForm.max_recursive_search" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>地图提供方</span>
                  <select v-model="fileSystemForm.map_provider" class="field-select">
                    <option value="google-leaflet">google-leaflet</option>
                    <option value="osm-leaflet">osm-leaflet</option>
                    <option value="osm-mapbox">osm-mapbox</option>
                  </select>
                </label>
                <label class="field-card">
                  <span>目录统计缓存（秒）</span>
                  <input v-model.number="fileSystemForm.directory_stat_ttl" type="number" class="field-input" />
                </label>
              </div>
            </article>

          </section>

          <section class="text-panels settings-text-stack">
            <label class="field-card span-2">
              <span>MIME 映射</span>
              <textarea v-model="fileSystemForm.mime_map" class="field-textarea" rows="5"></textarea>
            </label>
            <label class="field-card">
              <span>图片分类查询</span>
              <textarea v-model="fileSystemForm.image_query" class="field-textarea" rows="5"></textarea>
            </label>
            <label class="field-card">
              <span>视频分类查询</span>
              <textarea v-model="fileSystemForm.video_query" class="field-textarea" rows="5"></textarea>
            </label>
            <label class="field-card">
              <span>音频分类查询</span>
              <textarea v-model="fileSystemForm.audio_query" class="field-textarea" rows="5"></textarea>
            </label>
            <label class="field-card">
              <span>文档分类查询</span>
              <textarea v-model="fileSystemForm.document_query" class="field-textarea" rows="5"></textarea>
            </label>
          </section>

          <section class="group-grid settings-stack">
            <article class="group-card">
              <div class="group-head settings-section-head">
                <h3>文件加密</h3>
                <p>控制主加密密钥的存储方式，以及是否向用户展示文件加密状态。</p>
              </div>
              <div class="form-grid settings-form-list two-column">
                <label class="field-card">
                  <span>主加密密钥存储方式</span>
                  <select v-model="fileSystemForm.master_key_storage" class="field-select">
                    <option value="database">数据库</option>
                    <option value="file">文件</option>
                    <option value="env">环境变量</option>
                  </select>
                </label>
                <label class="toggle-card">
                  <input v-model="fileSystemForm.show_encryption_status" type="checkbox" />
                  <div>
                    <strong>显示加密状态</strong>
                    <small>开启后，文件详情和文件夹详情中会展示当前加密状态。</small>
                  </div>
                </label>
              </div>
            </article>

            <article class="group-card">
              <div class="group-head settings-section-head">
                <h3>文件事件推送</h3>
                <p>控制离线补偿与事件合并节奏，影响客户端同步与刷新体验。</p>
              </div>
              <div class="form-grid settings-form-list two-column">
                <label class="toggle-card">
                  <input v-model="fileSystemForm.enable_event_push" type="checkbox" />
                  <div>
                    <strong>启用文件事件推送</strong>
                    <small>开启后，同步客户端可订阅文件变化并及时处理。</small>
                  </div>
                </label>
                <label class="field-card">
                  <span>离线有效期（秒）</span>
                  <input v-model.number="fileSystemForm.offline_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>防抖延时（秒）</span>
                  <input v-model.number="fileSystemForm.debounce_delay" type="number" class="field-input" />
                </label>
              </div>
            </article>
          </section>

          <section class="group-grid settings-stack">
            <article class="group-card">
              <div class="group-head settings-section-head">
                <h3>高级设置</h3>
                <p>上传、下载、签名、目录统计、重试与并行传输等底层参数。</p>
              </div>
              <div class="form-grid settings-form-list two-column">
                <label class="field-card">
                  <span>服务端打包下载会话有效期（秒）</span>
                  <input v-model.number="fileSystemForm.server_side_download_session_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>上传会话有效期（秒）</span>
                  <input v-model.number="fileSystemForm.upload_session_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>从机 API 签名有效期（秒）</span>
                  <input v-model.number="fileSystemForm.slave_api_sign_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>目录统计信息有效期（秒）</span>
                  <input v-model.number="fileSystemForm.directory_stat_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>分片错误最大重试</span>
                  <input v-model.number="fileSystemForm.max_chunk_retry" type="number" class="field-input" />
                </label>
                <label class="toggle-card">
                  <input v-model="fileSystemForm.cache_chunks_for_retry" type="checkbox" />
                  <div>
                    <strong>缓存流式分片文件用于重试</strong>
                    <small>开启后，失败重试可复用已缓存的分片文件。</small>
                  </div>
                </label>
                <label class="field-card">
                  <span>中转最大并行传输</span>
                  <input v-model.number="fileSystemForm.transfer_parallelism" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>OAuth 存储策略凭证刷新间隔</span>
                  <input v-model="fileSystemForm.oauth_refresh_interval" type="text" class="field-input" />
                </label>
                <label class="field-card">
                  <span>WOPI 会话有效期（秒）</span>
                  <input v-model.number="fileSystemForm.wopi_session_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>文件 Blob 临时 URL 有效期（秒）</span>
                  <input v-model.number="fileSystemForm.blob_signed_url_ttl" type="number" class="field-input" />
                </label>
                <label class="field-card">
                  <span>文件 Blob 临时 URL 复用窗口（秒）</span>
                  <input v-model.number="fileSystemForm.blob_signed_url_reuse_ttl" type="number" class="field-input" />
                </label>
              </div>
            </article>

            <article class="group-card">
              <div class="group-head settings-section-head settings-section-head-inline">
                <h3>Blob URL 缓存</h3>
                <p>后端会按你设置的有效期与复用窗口复用预签名 URL，减少重复生成。</p>
              </div>
              <div class="settings-side-body">
                <div class="settings-side-metrics clean-side-metrics">
                  <div class="settings-side-metric">
                    <span>URL TTL</span>
                    <strong>{{ fileSystemForm.blob_signed_url_ttl }}s</strong>
                  </div>
                  <div class="settings-side-metric">
                    <span>Reuse Window</span>
                    <strong>{{ fileSystemForm.blob_signed_url_reuse_ttl }}s</strong>
                  </div>
                </div>
                <div class="settings-side-copy clean-side-copy">
                  <p>适合用于预览、下载和外部资源中转等高频场景，减少重复生成预签名地址，让缓存与有效期更清晰可控。</p>
                </div>
              </div>
              <div class="inline-actions inline-actions-left settings-side-actions">
                <button class="ghost-button" type="button" :disabled="clearingBlobUrlCache" @click="handleClearBlobUrlCache">
                  {{ clearingBlobUrlCache ? '正在清理缓存...' : '清除 Blob URL 缓存' }}
                </button>
              </div>
            </article>
          </section>
        </template>

        <template v-else-if="activeTab === 'search'">
          <section class="section-header search-header">
            <div class="search-heading">
              <h2>全文搜索</h2>
              <p>连接 Meilisearch / Tika 搜索服务，支持真实重建索引并查看任务结果。</p>
            </div>
          </section>

          <section class="overview-grid search-metrics-grid">
            <article class="overview-card search-metric-card search-metric-card-blue">
              <span>全文搜索</span>
              <strong>{{ searchForm.enabled ? '已开启' : '已关闭' }}</strong>
              <small>AI 搜索 {{ searchForm.ai_search ? '开启' : '关闭' }}</small>
            </article>
            <article class="overview-card search-metric-card search-metric-card-violet">
              <span>已索引文件数</span>
              <strong>{{ jobMetrics.indexedFiles }}</strong>
              <small>最近任务真实上报</small>
            </article>
            <article class="overview-card search-metric-card search-metric-card-emerald">
              <span>分块数</span>
              <strong>{{ jobMetrics.chunkCount }}</strong>
              <small>按切块配置生成</small>
            </article>
            <article class="overview-card search-metric-card search-metric-card-gold">
              <span>文档数</span>
              <strong>{{ jobMetrics.documentCount }}</strong>
              <small>进入搜索引擎的索引文档</small>
            </article>
          </section>

          <section class="group-grid single search-config-layout">
            <article class="group-card search-config-card">
              <div class="group-head search-card-head">
                <h3>搜索服务配置</h3>
                <p>开启后会向真实服务写入配置，并用于后续重建索引任务。</p>
              </div>
              <div class="form-grid search-form-grid">
                <label class="toggle-card search-toggle-card">
                  <input v-model="searchForm.enabled" type="checkbox" />
                  <div class="search-toggle-copy">
                    <strong>启用全文搜索</strong>
                    <small>关闭后前台仍可展示搜索入口，但不再使用全文索引。</small>
                  </div>
                </label>
                <label class="toggle-card search-toggle-card">
                  <input v-model="searchForm.ai_search" type="checkbox" />
                  <div class="search-toggle-copy">
                    <strong>AI 搜索增强</strong>
                    <small>保留语义搜索增强开关，方便后续联动更强体验。</small>
                  </div>
                </label>
                <label class="field-card search-field-card">
                  <span>Meilisearch 地址</span>
                  <input v-model="searchForm.meili_endpoint" type="text" class="field-input" />
                </label>
                <label class="field-card search-field-card">
                  <span>Meilisearch API Key</span>
                  <input v-model="searchForm.api_key" type="text" class="field-input" />
                </label>
                <label class="field-card search-field-card">
                  <span>Tika 地址</span>
                  <input v-model="searchForm.tika_endpoint" type="text" class="field-input" />
                </label>
                <label class="field-card search-field-card">
                  <span>结果页大小</span>
                  <input v-model.number="searchForm.result_page_size" type="number" class="field-input" />
                </label>
                <label class="field-card search-field-card">
                  <span>最大索引文件大小</span>
                  <div class="split-row">
                    <input v-model.number="searchForm.max_file_size" type="number" class="field-input" />
                    <select v-model="searchForm.max_file_size_unit" class="field-select">
                      <option v-for="unit in units" :key="unit" :value="unit">{{ unit }}</option>
                    </select>
                  </div>
                </label>
                <label class="field-card search-field-card">
                  <span>切块大小</span>
                  <div class="split-row">
                    <input v-model.number="searchForm.chunk_size" type="number" class="field-input" />
                    <select v-model="searchForm.chunk_unit" class="field-select">
                      <option value="B">B</option>
                      <option value="KB">KB</option>
                      <option value="MB">MB</option>
                    </select>
                  </div>
                </label>
              </div>
            </article>
          </section>

          <section class="job-layout search-job-layout">
            <article class="job-hero search-job-hero">
              <div class="job-copy">
                <span class="job-pill">{{ latestJobStatusLabel }}</span>
                <h3>真实重建索引任务</h3>
                <p>{{ searchJobSummary }}</p>
              </div>
              <button class="primary-button search-cta-button" type="button" :disabled="rebuilding" @click="handleRebuildIndex">
                {{ rebuilding ? '正在提交重建任务...' : '重建索引' }}
              </button>
            </article>

            <article class="job-stats-card search-job-stats-card">
              <div class="job-stats-grid search-job-stats-grid">
                <div>
                  <span>任务编号</span>
                  <strong>{{ latestJob?.id ?? '--' }}</strong>
                </div>
                <div>
                  <span>队列</span>
                  <strong>{{ latestJob?.queue_key ?? '--' }}</strong>
                </div>
                <div>
                  <span>已索引文件数</span>
                  <strong>{{ jobMetrics.indexedFiles }}</strong>
                </div>
                <div>
                  <span>分块数</span>
                  <strong>{{ jobMetrics.chunkCount }}</strong>
                </div>
                <div>
                  <span>文档数</span>
                  <strong>{{ jobMetrics.documentCount }}</strong>
                </div>
                <div>
                  <span>跳过原因</span>
                  <strong>{{ jobMetrics.skipReasonCount }}</strong>
                </div>
              </div>

              <div v-if="jobMetrics.skipReasons.length" class="reason-list search-reason-list">
                <span class="reason-title">最近任务跳过原因</span>
                <div class="reason-chips">
                  <span v-for="reason in jobMetrics.skipReasons" :key="reason" class="reason-chip">{{ reason }}</span>
                </div>
              </div>
            </article>
          </section>

          <section class="text-panels search-text-panels">
            <label class="field-card search-note-card">
              <span>索引扩展名</span>
              <textarea v-model="searchForm.extensions" class="field-textarea" rows="4"></textarea>
            </label>
            <label class="field-card search-note-card">
              <span>索引备注</span>
              <textarea v-model="searchForm.index_notes" class="field-textarea" rows="4"></textarea>
            </label>
          </section>
        </template>

        <template v-else-if="activeTab === 'icons'">
          <section class="section-header icon-section-header">
            <div>
              <h2>文件图标</h2>
              <p>支持可视化维护图标规则，并同步保存到真实的 `file_icon_rules` 持久化字段。</p>
            </div>
            <button class="secondary-button" type="button" @click="addIconRule">添加图标规则</button>
          </section>

          <!-- <section class="overview-grid icon-overview-grid">
            <article class="overview-card icon-overview-card icon-overview-card-blue">
              <span>瀹告彃鎯庨悽銊潐閸?</span>
              <strong>{{ iconRules.length }}</strong>
              <small>瑜版挸澧犻弬鍥︽閸ョ偓鐖ｉ崠褰掑帳鐟欏嫬鍨幀缁樻殶</small>
            </article>
            <article class="overview-card icon-overview-card icon-overview-card-violet">
              <span>妫版粏澹婇崫浣哄閸?</span>
              <strong>{{ iconRules.filter((rule) => String(rule.tint || '').trim()).length }}</strong>
              <small>瀹告煡鍘ょ純顔垮瑜扳晜鐖ｇ拋鎵畱閸ョ偓鐖ｇ憴鍕灟</small>
            </article>
            <article class="overview-card icon-overview-card icon-overview-card-gold">
              <span>Emoji 配置</span>
              <strong>{{ fileSystemForm.emoji_options ? '已填写' : '未配置' }}</strong>
              <small>列表 Emoji 显示策略状态</small>
            </article>
          </section>

          </section> -->

          <section class="overview-grid icon-overview-grid">
            <article class="overview-card icon-overview-card icon-overview-card-blue">
              <span>图标规则数</span>
              <strong>{{ iconRules.length }}</strong>
              <small>当前文件图标匹配规则总数</small>
            </article>
            <article class="overview-card icon-overview-card icon-overview-card-violet">
              <span>已配色规则</span>
              <strong>{{ iconRules.filter((rule) => String(rule.tint || '').trim()).length }}</strong>
              <small>已经加入品牌色标记的图标规则</small>
            </article>
            <article class="overview-card icon-overview-card icon-overview-card-gold">
              <span>Emoji 配置</span>
              <strong>{{ fileSystemForm.emoji_options ? '已填写' : '未配置' }}</strong>
              <small>列表 Emoji 显示策略状态</small>
            </article>
          </section>

          <section class="icon-table-card">
            <div class="icon-table-head">
              <span>预览</span>
              <span>规则名称</span>
              <span>图标 / Emoji</span>
              <span>扩展名或 MIME</span>
              <span>颜色标记</span>
              <span>操作</span>
            </div>

            <div class="icon-rule-list">
            <article v-for="rule in iconRules" :key="rule.id" class="icon-table-row" :style="{ '--icon-rule-tint': rule.tint || '#2563eb' }">
              <div class="editor-head icon-rule-head">
                <div class="preview-icon large">{{ rule.icon || '📁' }}</div>
                <div class="editor-title icon-rule-title">
                  <strong>{{ rule.label || '未命名规则' }}</strong>
                  <span>{{ rule.match || '未填写匹配条件' }}</span>
                </div>
                <button class="danger-text" type="button" @click="removeIconRule(rule.id)">删除</button>
              </div>

              <div class="form-grid icon-rule-form">
                <label class="field-card">
                  <span>规则名称</span>
                  <input v-model="rule.label" type="text" class="field-input" @input="syncIconRulesToJson" />
                </label>
                <label class="field-card">
                  <span>图标 / Emoji</span>
                  <input v-model="rule.icon" type="text" class="field-input" @input="syncIconRulesToJson" />
                </label>
                <label class="field-card">
                  <span>匹配扩展名或 MIME</span>
                  <input v-model="rule.match" type="text" class="field-input" @input="syncIconRulesToJson" />
                </label>
                <label class="field-card">
                  <span>颜色标记</span>
                  <input v-model="rule.tint" type="text" class="field-input" placeholder="#2563eb" @input="syncIconRulesToJson" />
                </label>
              </div>
            </article>

              <article v-if="!iconRules.length" class="empty-card icon-empty-card">
                <strong>暂未配置图标规则</strong>
                <span>点击右上角“添加图标规则”开始配置。</span>
              </article>
            </div>
          </section>

          <section class="text-panels icon-json-stack">
            <section class="emoji-config-card">
              <div class="emoji-config-head">
                <div>
                  <h3>Emoji 选项</h3>
                  <p>Emoji 分类可参与文件图标匹配，match 支持扩展名、完整 MIME、image/* 这类 MIME 通配符。</p>
                </div>
                <div class="emoji-toolbar">
                  <button class="ghost-button" type="button" @click="loadEmojiOptionsFromJson">从 JSON 载入</button>
                  <button class="ghost-button" type="button" @click="resetEmojiCategoriesToPreset">恢复预置</button>
                  <button class="secondary-button" type="button" @click="addEmojiCategory">添加分类</button>
                </div>
              </div>

              <div class="emoji-meta-grid">
                <label class="field-card">
                  <span>启用 Emoji 显示</span>
                  <input v-model="emojiOptionsEditor.enabled" class="switch-input" type="checkbox" @change="syncEmojiOptionsToJson" />
                  <small>控制文件列表是否启用 Emoji 辅助识别。</small>
                </label>
                <label class="field-card">
                  <span>列表中展示</span>
                  <input v-model="emojiOptionsEditor.showInList" class="switch-input" type="checkbox" @change="syncEmojiOptionsToJson" />
                  <small>开启后会在列表视图直接显示对应 Emoji。</small>
                </label>
                <label class="field-card">
                  <span>未知类型兜底</span>
                  <input v-model="emojiOptionsEditor.fallbackUnknown" class="switch-input" type="checkbox" @change="syncEmojiOptionsToJson" />
                  <small>当没有命中分类时，使用默认未知图标兜底。</small>
                </label>
                <label class="field-card">
                  <span>文件夹 Emoji</span>
                  <input v-model="emojiOptionsEditor.folderEmoji" type="text" class="field-input" @input="syncEmojiOptionsToJson" />
                  <small>文件夹默认展示图标，建议使用高识别度符号。</small>
                </label>
                <label class="field-card">
                  <span>未知文件 Emoji</span>
                  <input v-model="emojiOptionsEditor.unknownEmoji" type="text" class="field-input" @input="syncEmojiOptionsToJson" />
                  <small>未命中任意规则时使用的默认图标。</small>
                </label>
                <article class="emoji-config-tip">
                  <strong>{{ emojiOptionsEditor.categories.length }}</strong>
                  <span>当前已配置 {{ emojiOptionsEditor.categories.length }} 个 Emoji 分类；填写 match 后会参与文件列表图标匹配。</span>
                </article>
              </div>

              <div class="emoji-option-table">
                <div class="emoji-option-head">
                  <span>分类名称 / 匹配</span>
                  <span>图标</span>
                  <span>候选 Emoji</span>
                  <span>操作</span>
                </div>

                <div class="emoji-option-body">
                  <article
                    v-for="(category, index) in emojiOptionsEditor.categories"
                    :key="category.id"
                    class="emoji-option-row"
                  >
                    <div class="emoji-rule-column">
                      <label>
                        <span>分类名称</span>
                        <input
                          v-model="category.label"
                          type="text"
                          class="emoji-category-input"
                          :placeholder="`分类 ${index + 1}`"
                          @input="syncEmojiOptionsToJson"
                        />
                      </label>
                      <label>
                        <span>匹配扩展名或 MIME</span>
                        <input
                          v-model="category.match"
                          type="text"
                          class="emoji-category-input"
                          placeholder="jpg,png,image/*"
                          @input="syncEmojiOptionsToJson"
                        />
                      </label>
                      <small>留空时仅作为候选 Emoji 集合，不参与文件类型匹配。</small>
                    </div>

                    <div class="emoji-category-column">
                      <label class="emoji-category-badge">
                        <span class="emoji-category-preview">{{ category.icon || '🙂' }}</span>
                        <input v-model="category.icon" type="text" class="emoji-category-input" @input="syncEmojiOptionsToJson" />
                      </label>
                      <small>命中分类时显示该图标。</small>
                    </div>

                    <div class="emoji-cloud-column">
                      <textarea
                        v-model="category.emojisText"
                        class="emoji-cloud-input"
                        rows="4"
                        placeholder="使用英文逗号分隔多个 Emoji"
                        @input="syncEmojiOptionsToJson"
                      ></textarea>
                      <small>可作为备注或候选 emoji 集合保存，不会单独匹配扩展名。</small>
                    </div>

                    <div class="emoji-row-actions">
                      <button class="emoji-action-button" type="button" @click="moveEmojiCategory(category.id, -1)">上移</button>
                      <button class="emoji-action-button" type="button" @click="moveEmojiCategory(category.id, 1)">下移</button>
                      <button class="emoji-action-button danger" type="button" @click="removeEmojiCategory(category.id)">删除</button>
                    </div>
                  </article>

                  <article v-if="!emojiOptionsEditor.categories.length" class="empty-card icon-empty-card">
                    <strong>暂未配置 Emoji 分类</strong>
                    <span>点击“恢复预置”或“添加分类”，马上搭出更完整的文件图标体验。</span>
                  </article>
                </div>
              </div>
            </section>

            <label class="field-card span-2 icon-json-card">
              <span>文件图标规则 JSON</span>
              <textarea v-model="fileSystemForm.file_icon_rules" class="field-textarea giant" rows="14"></textarea>
              <div class="inline-actions inline-actions-left">
                <button class="ghost-button" type="button" @click="loadIconRulesFromJson">从 JSON 载入可视化规则</button>
              </div>
            </label>
            <label class="field-card span-2 icon-json-card icon-emoji-card">
              <span>Emoji 显示配置 JSON</span>
              <textarea v-model="fileSystemForm.emoji_options" class="field-textarea" rows="6"></textarea>
              <div class="inline-actions inline-actions-left">
                <button class="ghost-button" type="button" @click="loadEmojiOptionsFromJson">从 JSON 刷新上方界面</button>
              </div>
            </label>
          </section>
        </template>

        <template v-else-if="activeTab === 'apps'">
          <section class="app-toolbar">
            <button class="app-add-button" type="button" @click="addBrowserApp">
              <svg viewBox="0 0 20 20" aria-hidden="true">
                <path d="M10 4v12" />
                <path d="M4 10h12" />
              </svg>
              <span>添加应用</span>
            </button>
          </section>

          <section class="app-studio-card app-table-card">
            <div class="app-group-head">
              <div class="app-group-copy">
                <strong>
                  <span class="app-group-folder">
                    <svg viewBox="0 0 20 20" aria-hidden="true">
                      <path d="M2.5 6.5A2.5 2.5 0 0 1 5 4h3.1c.7 0 1.3.29 1.77.8l.76.82c.28.3.67.48 1.09.48H15A2.5 2.5 0 0 1 17.5 8.6v5.9A2.5 2.5 0 0 1 15 17H5a2.5 2.5 0 0 1-2.5-2.5v-8Z" />
                    </svg>
                  </span>
                  <span>应用分组 #1</span>
                </strong>
                <span>单个分组内汇总展示全部内置应用，右侧滚动即可浏览完整品牌图标列表。</span>
              </div>
              <div class="app-group-actions">
                <span class="app-group-count">{{ browserApps.length }} 个应用</span>
                <button class="app-group-toggle" type="button" @click="appsGroupCollapsed = !appsGroupCollapsed">
                  <svg viewBox="0 0 20 20" aria-hidden="true" :class="{ 'is-collapsed': appsGroupCollapsed }">
                    <path d="M5.5 7.5L10 12l4.5-4.5" />
                  </svg>
                </button>
              </div>
            </div>

            <div v-show="!appsGroupCollapsed" class="app-table-shell">
              <section v-if="activeBrowserApp" class="app-inline-editor-card">
                <div class="app-inline-editor-head">
                  <div
                    class="app-brand app-inline-editor-brand"
                    :class="[`app-brand-${resolveAppBrandMeta(activeBrowserApp).key}`]"
                    :style="{ background: resolveAppBrandMeta(activeBrowserApp).background || activeBrowserApp.background || defaultAppBackground(activeBrowserApp.id) }"
                  >
                    <div v-if="resolveAppBrandMeta(activeBrowserApp).svg" class="app-brand-svg" v-html="resolveAppBrandMeta(activeBrowserApp).svg"></div>
                    <template v-else>{{ resolveAppBrandMeta(activeBrowserApp).glyph }}</template>
                  </div>
                  <div class="app-inline-editor-copy">
                    <strong>{{ activeBrowserApp.name || '未命名应用' }}</strong>
                    <span>当前编辑区保持在同一页面里，表格仍然维持参考图那种紧凑的一行一条记录。</span>
                  </div>
                  <button class="app-inline-editor-close" type="button" @click="activeBrowserAppId = ''">
                    <svg viewBox="0 0 20 20" aria-hidden="true">
                      <path d="M5.5 5.5l9 9" />
                      <path d="M14.5 5.5l-9 9" />
                    </svg>
                  </button>
                </div>

                <div class="app-inline-editor-grid">
                  <label class="app-inline-field">
                    <span>品牌图标</span>
                    <input v-model="activeBrowserApp.icon" type="text" class="field-input" @input="syncBrowserAppsToJson" />
                  </label>
                  <label class="app-inline-field">
                    <span>应用名称</span>
                    <input v-model="activeBrowserApp.name" type="text" class="field-input" @input="syncBrowserAppsToJson" />
                  </label>
                  <label class="app-inline-field">
                    <span>打开动作</span>
                    <input v-model="activeBrowserApp.action" type="text" class="field-input" placeholder="/viewer/pdf?path={path}" @input="syncBrowserAppsToJson" />
                  </label>
                  <label class="app-inline-field">
                    <span>按钮背景</span>
                    <input v-model="activeBrowserApp.background" type="text" class="field-input" placeholder="linear-gradient(...)" @input="syncBrowserAppsToJson" />
                  </label>
                  <label class="app-inline-field app-inline-field-wide">
                    <span>说明文案</span>
                    <input v-model="activeBrowserApp.description" type="text" class="field-input" placeholder="应用说明" @input="syncBrowserAppsToJson" />
                  </label>
                </div>
              </section>

              <div class="app-table-head">
                <span>图标</span>
                <span>类型</span>
                <span>名称</span>
                <span>扩展名列表</span>
                <span>平台</span>
                <span>新建文件映射</span>
                <span>启用</span>
                <span>操作</span>
              </div>

              <div class="app-table-body">
                <article
                  v-for="app in browserApps"
                  :key="app.id"
                  class="app-table-row"
                  :class="{ 'is-active': activeBrowserAppId === app.id }"
                >
                  <div class="app-table-icon">
                    <div
                      class="app-brand app-table-brand"
                      :class="[`app-brand-${resolveAppBrandMeta(app).key}`]"
                      :style="{ background: resolveAppBrandMeta(app).background || app.background || defaultAppBackground(app.id) }"
                    >
                      <div v-if="resolveAppBrandMeta(app).svg" class="app-brand-svg" v-html="resolveAppBrandMeta(app).svg"></div>
                      <template v-else>{{ resolveAppBrandMeta(app).glyph }}</template>
                    </div>
                  </div>

                  <div class="app-type-cell">
                    <span class="app-type-badge" :class="resolveAppTypeMeta(app).tone">
                      {{ resolveAppTypeMeta(app).label }}
                    </span>
                  </div>

                  <div class="app-name-cell">
                    <strong>{{ app.name || '未命名应用' }}</strong>
                  </div>

                  <div class="app-extensions-cell">
                    <input
                      v-model="app.extensions"
                      type="text"
                      class="field-input app-extensions-input"
                      placeholder="pdf,docx,txt"
                      @input="syncBrowserAppsToJson"
                    />
                  </div>

                  <div class="app-platform-cell">
                    <span class="app-platform-pill">全平台</span>
                  </div>

                  <div class="app-mapping-cell">
                    <span class="app-mapping-pill">{{ resolveAppMappingLabel(app) }}</span>
                  </div>

                  <div class="app-enable-cell">
                    <label class="app-toggle-readonly">
                      <input type="checkbox" checked disabled />
                      <span>启用</span>
                    </label>
                  </div>

                  <div class="app-actions-cell">
                    <div class="app-row-buttons">
                      <button class="app-row-button app-row-button-line" type="button" aria-label="编辑应用" @click="startEditBrowserApp(app.id)">
                        <svg viewBox="0 0 20 20" aria-hidden="true">
                          <path d="M4 13.5V16h2.5L14.7 7.8l-2.5-2.5L4 13.5z" />
                          <path d="M10.9 4.6l2.5 2.5" />
                        </svg>
                      </button>
                      <button
                        v-if="resolveAppTypeMeta(app).label === '自定义'"
                        class="app-row-button app-row-button-line danger"
                        type="button"
                        aria-label="删除应用"
                        @click="removeBrowserApp(app.id)"
                      >
                        <svg viewBox="0 0 20 20" aria-hidden="true">
                          <path d="M5.5 5.5l9 9" />
                          <path d="M14.5 5.5l-9 9" />
                        </svg>
                      </button>
                      <button class="app-row-button app-row-button-line" type="button" aria-label="上移应用" @click="moveBrowserApp(app.id, -1)">
                        <svg viewBox="0 0 20 20" aria-hidden="true">
                          <path d="M10 15V5" />
                          <path d="M5.5 9.5L10 5l4.5 4.5" />
                        </svg>
                      </button>
                      <button class="app-row-button app-row-button-line" type="button" aria-label="下移应用" @click="moveBrowserApp(app.id, 1)">
                        <svg viewBox="0 0 20 20" aria-hidden="true">
                          <path d="M10 5v10" />
                          <path d="M5.5 10.5L10 15l4.5-4.5" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </article>

                <article v-if="!browserApps.length" class="empty-card app-empty-card">
                  <strong>暂未配置浏览应用</strong>
                  <span>点击右上角“添加应用”，马上搭一个更像应用市场的文件打开生态。</span>
                </article>
              </div>
            </div>
          </section>

          <section class="text-panels app-json-panels">
            <label class="field-card span-2 app-json-card">
              <span>浏览应用 JSON</span>
              <textarea v-model="fileSystemForm.browser_apps" class="field-textarea giant" rows="14"></textarea>
              <div class="inline-actions inline-actions-left">
                <button class="ghost-button" type="button" @click="loadBrowserAppsFromJson">从 JSON 载入可视化应用</button>
              </div>
            </label>
          </section>
        </template>

        <template v-else>
          <section class="custom-props-shell">
            <section class="custom-props-header">
              <div class="custom-props-copy">
                <span class="custom-props-kicker">Custom Properties Studio</span>
                <h2>自定义属性</h2>
                <p>把文件系统属性管理做成更像成熟后台的体验：上方简洁工具条、中部表格列表、右侧同页弹层编辑，视觉更接近你给的参考图。</p>
              </div>

              <div class="custom-props-metrics">
                <article class="custom-props-metric">
                  <span>属性总数</span>
                  <strong>{{ customProperties.length }}</strong>
                </article>
                <article class="custom-props-metric">
                  <span>必填字段</span>
                  <strong>{{ customProperties.filter((item) => item.required).length }}</strong>
                </article>
                <article class="custom-props-metric">
                  <span>字段类型</span>
                  <strong>{{ Array.from(new Set(customProperties.map((item) => item.type))).length }}</strong>
                </article>
              </div>
            </section>

            <section class="custom-props-toolbar">
              <button class="custom-props-add-button" type="button" @click="addCustomProperty">
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M10 4v12" />
                  <path d="M4 10h12" />
                </svg>
                <span>添加</span>
              </button>
            </section>

            <section class="custom-props-table-card">
              <div class="custom-props-table-head">
                <span>名称</span>
                <span>类型</span>
                <span>默认值</span>
                <span>操作</span>
              </div>

              <div class="custom-props-table-body">
                <article
                  v-for="property in customProperties"
                  :key="property.id"
                  class="custom-props-table-row"
                  :class="{ 'is-active': activeCustomPropertyId === property.id }"
                >
                  <div class="custom-props-name-cell">
                    <div class="custom-prop-icon" :class="`is-${property.type}`">
                      {{ resolveCustomPropertyMeta(property).glyph }}
                    </div>
                    <div class="custom-prop-name-copy">
                      <strong>{{ property.name || '未命名属性' }}</strong>
                      <span>{{ property.scope }}</span>
                    </div>
                  </div>

                  <div class="custom-props-type-cell">
                    <span class="custom-props-type-badge" :class="`is-${property.type}`">
                      {{ resolveCustomPropertyMeta(property).label }}
                    </span>
                  </div>

                  <div class="custom-props-default-cell">
                    <template v-if="property.type === 'rating'">
                      <div class="custom-rating-preview">
                        <button
                          v-for="star in 5"
                          :key="star"
                          type="button"
                          class="custom-rating-star"
                          :class="{ 'is-filled': star <= getCustomPropertyPreviewRating(property) }"
                          @click="setCustomPropertyPreviewRating(property.id, star)"
                        >
                          ★
                        </button>
                      </div>
                    </template>
                    <template v-else-if="property.type === 'switch'">
                      <span class="custom-switch-preview" :class="{ 'is-on': getCustomPropertyPreviewSwitch(property) }">
                        {{ getCustomPropertyPreviewSwitch(property) ? '开启' : '关闭' }}
                      </span>
                    </template>
                    <template v-else-if="property.type === 'date'">
                      <span class="custom-date-preview">{{ getCustomPropertyPreviewDate(property) }}</span>
                    </template>
                    <template v-else-if="property.type === 'tags' || property.type === 'multi_select'">
                      <div class="custom-options-preview" :class="{ 'is-multi': property.type === 'multi_select' }">
                        <span v-for="option in previewCustomPropertyOptions(property)" :key="option">{{ option }}</span>
                      </div>
                    </template>
                    <template v-else-if="property.type === 'textarea'">
                      <span class="custom-default-text textarea">{{ getCustomPropertyPreviewText(property) }}</span>
                    </template>
                    <template v-else>
                      <span class="custom-default-text">{{ getCustomPropertyPreviewText(property) }}</span>
                    </template>
                  </div>

                  <div class="custom-props-actions-cell">
                    <button class="custom-action-button" type="button" aria-label="编辑属性" @click="startEditCustomProperty(property.id)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M4 13.5V16h2.5L14.7 7.8l-2.5-2.5L4 13.5z" />
                        <path d="M10.9 4.6l2.5 2.5" />
                      </svg>
                    </button>
                    <button class="custom-action-button danger" type="button" aria-label="删除属性" @click="removeCustomProperty(property.id)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M5.5 5.5l9 9" />
                        <path d="M14.5 5.5l-9 9" />
                      </svg>
                    </button>
                    <button class="custom-action-button" type="button" aria-label="上移属性" @click="moveCustomProperty(property.id, -1)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M10 15V5" />
                        <path d="M5.5 9.5L10 5l4.5 4.5" />
                      </svg>
                    </button>
                    <button class="custom-action-button" type="button" aria-label="下移属性" @click="moveCustomProperty(property.id, 1)">
                      <svg viewBox="0 0 20 20" aria-hidden="true">
                        <path d="M10 5v10" />
                        <path d="M5.5 10.5L10 15l4.5-4.5" />
                      </svg>
                    </button>
                  </div>
                </article>

                <article v-if="!customProperties.length" class="empty-card custom-props-empty-card">
                  <strong>暂未配置自定义属性</strong>
                  <span>点击左上角“添加”，马上搭出更完整的文件详情字段体系。</span>
                </article>
              </div>
            </section>

            <section v-if="activeCustomProperty" class="custom-props-modal-layer">
              <article class="custom-props-modal-card">
                <div class="custom-props-modal-head">
                  <div>
                    <strong>编辑属性</strong>
                    <span>保持同页管理体验，同时做出更接近参考图的弹层结构和节奏。</span>
                  </div>
                  <button class="custom-props-modal-close" type="button" @click="activeCustomPropertyId = ''">
                    <svg viewBox="0 0 20 20" aria-hidden="true">
                      <path d="M5.5 5.5l9 9" />
                      <path d="M14.5 5.5l-9 9" />
                    </svg>
                  </button>
                </div>

                <div class="custom-props-modal-form">
                  <label class="custom-props-modal-field">
                    <span>标识</span>
                    <input :value="activeCustomProperty.id" type="text" class="field-input" readonly />
                    <small>属性唯一标识由系统维护，当前用于前端编辑状态与真实 JSON 同步。</small>
                  </label>

                  <label class="custom-props-modal-field">
                    <span>名称</span>
                    <input v-model="activeCustomProperty.name" type="text" class="field-input" @input="syncCustomPropertiesToJson" />
                    <small>展示名称会直接出现在文件详情与属性表格里。</small>
                  </label>

                  <label class="custom-props-modal-field">
                    <span>字段类型</span>
                    <select v-model="activeCustomProperty.type" class="field-select" @change="syncCustomPropertiesToJson">
                      <option value="text">text</option>
                      <option value="textarea">textarea</option>
                      <option value="switch">switch</option>
                      <option value="date">date</option>
                      <option value="tags">tags</option>
                      <option value="multi_select">multi_select</option>
                      <option value="rating">rating</option>
                    </select>
                    <small>支持文本、开关、日期、标签、多选和评分等常见业务字段。</small>
                  </label>

                  <label class="custom-props-modal-field">
                    <span>作用范围</span>
                    <select v-model="activeCustomProperty.scope" class="field-select" @change="syncCustomPropertiesToJson">
                      <option value="文件">文件</option>
                      <option value="文件夹">文件夹</option>
                      <option value="文件 / 文件夹">文件 / 文件夹</option>
                    </select>
                    <small>控制字段在哪类对象上显示，避免文件与文件夹信息混杂。</small>
                  </label>

                  <label class="custom-props-modal-field custom-props-modal-switch">
                    <span>必填设置</span>
                    <label class="custom-props-required-toggle">
                      <input v-model="activeCustomProperty.required" type="checkbox" @change="syncCustomPropertiesToJson" />
                      <strong>{{ activeCustomProperty.required ? '已设为必填' : '当前为选填' }}</strong>
                    </label>
                    <small>开启后可用于高价值资料的归档规范控制。</small>
                  </label>

                  <label class="custom-props-modal-field custom-props-modal-field-wide">
                    <span>描述</span>
                    <input v-model="activeCustomProperty.description" type="text" class="field-input" placeholder="点击编辑..." @input="syncCustomPropertiesToJson" />
                    <small>建议填写用户可理解的提示语，提升填写转化率与准确度。</small>
                  </label>

                  <label class="custom-props-modal-field custom-props-modal-field-wide">
                    <span>标签 / 选项</span>
                    <input
                      v-model="activeCustomProperty.optionsText"
                      type="text"
                      class="field-input"
                      placeholder="合同,重要,待归档"
                      @input="syncCustomPropertiesToJson"
                    />
                    <small>适用于标签、多选等离散型属性，使用英文逗号分隔。</small>
                  </label>

                  <div class="custom-props-modal-preview">
                    <span>默认值预览</span>
                    <div class="custom-props-modal-preview-box">
                      <template v-if="activeCustomProperty.type === 'rating'">
                        <div class="custom-rating-preview large">
                          <button
                            v-for="star in 5"
                            :key="star"
                            type="button"
                            class="custom-rating-star"
                            :class="{ 'is-filled': star <= getCustomPropertyPreviewRating(activeCustomProperty) }"
                            @click="setCustomPropertyPreviewRating(activeCustomProperty.id, star)"
                          >
                            ★
                          </button>
                        </div>
                      </template>
                      <template v-else-if="activeCustomProperty.type === 'switch'">
                        <button
                          type="button"
                          class="custom-switch-preview large interactive"
                          :class="{ 'is-on': getCustomPropertyPreviewSwitch(activeCustomProperty) }"
                          @click="toggleCustomPropertyPreviewSwitch(activeCustomProperty.id)"
                        >
                          {{ getCustomPropertyPreviewSwitch(activeCustomProperty) ? '默认开启' : '默认关闭' }}
                        </button>
                      </template>
                      <template v-else-if="activeCustomProperty.type === 'date'">
                        <span class="custom-date-preview large">{{ getCustomPropertyPreviewDate(activeCustomProperty) }}</span>
                      </template>
                      <template v-else-if="activeCustomProperty.type === 'tags' || activeCustomProperty.type === 'multi_select'">
                        <div class="custom-options-preview large" :class="{ 'is-multi': activeCustomProperty.type === 'multi_select' }">
                          <span v-for="option in previewCustomPropertyOptions(activeCustomProperty)" :key="option">{{ option }}</span>
                        </div>
                      </template>
                      <template v-else-if="activeCustomProperty.type === 'textarea'">
                        <span class="custom-default-text large textarea">{{ getCustomPropertyPreviewText(activeCustomProperty) }}</span>
                      </template>
                      <template v-else>
                        <span class="custom-default-text large">{{ getCustomPropertyPreviewText(activeCustomProperty) }}</span>
                      </template>
                    </div>
                  </div>
                </div>

                <div class="custom-props-modal-actions">
                  <button class="ghost-button" type="button" @click="activeCustomPropertyId = ''">取消</button>
                  <button class="primary-button" type="button" @click="activeCustomPropertyId = ''">确定</button>
                </div>
              </article>
            </section>

            <label class="field-card span-2 custom-props-json-card">
              <span>自定义属性 JSON</span>
              <textarea v-model="fileSystemForm.custom_properties" class="field-textarea giant" rows="14"></textarea>
              <div class="inline-actions">
                <button class="ghost-button" type="button" @click="loadCustomPropertiesFromJson">从 JSON 载入可视化属性</button>
              </div>
            </label>
          </section>
        </template>
      </div>
    </section>

    <footer class="action-bar">
      <div class="action-copy">
        <strong>{{ actionTitle }}</strong>
        <span>{{ actionDescription }}</span>
      </div>

      <div class="action-group">
        <button class="ghost-button" type="button" :disabled="busy" @click="reloadAll">重新加载</button>
        <button class="ghost-button" type="button" :disabled="busy" @click="resetCurrentTab">恢复当前页签</button>
        <button class="primary-button" type="button" :disabled="busy" @click="saveCurrentTab">
          {{ busy ? '处理中...' : '保存当前页签' }}
        </button>
      </div>
    </footer>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElIcon, ElMessage } from 'element-plus';
import { Connection, Search, Setting, Star, Tickets } from '@element-plus/icons-vue';
import { useRoute, useRouter } from 'vue-router';

import {
  clearFileSystemBlobUrlCache,
  getFileSystemBrowserApps,
  getFileSystemSettings,
  updateFileSystemIconSettings,
  updateFileSystemSettings,
  type BrowserAppGroupPayload,
  type FileSystemSettingsPayload,
} from '@/api/file-system-settings';
import {
  getFullTextSearchSettings,
  rebuildFullTextSearchIndex,
  updateFullTextSearchSettings,
  type FullTextSearchSettingsPayload,
} from '@/api/full-text-search-settings';
import { getQueueJobs, type QueueJobItemPayload } from '@/api/queue-settings';
import { useFileStore } from '@/stores/file';

type TabKey = 'parameters' | 'search' | 'icons' | 'apps' | 'custom';
type ParsedRecord = Record<string, unknown>;

type IconRuleEditor = {
  id: string;
  label: string;
  icon: string;
  match: string;
  tint: string;
};

type EmojiCategoryEditor = {
  id: string;
  label: string;
  match: string;
  icon: string;
  emojisText: string;
};

type EmojiOptionsEditor = {
  enabled: boolean;
  showInList: boolean;
  fallbackUnknown: boolean;
  folderEmoji: string;
  unknownEmoji: string;
  categories: EmojiCategoryEditor[];
};

type BrowserAppEditor = {
  id: string;
  name: string;
  icon: string;
  extensions: string;
  action: string;
  description: string;
  background: string;
  type: string;
  platform: string;
  createMapping: string;
  enabled: boolean;
  openInNewWindow: boolean;
  maxSize: number;
  maxSizeUnit: string;
};

type CustomPropertyEditor = {
  id: string;
  name: string;
  type: string;
  description: string;
  scope: string;
  required: boolean;
  optionsText: string;
};

type CustomPropertySeed = {
  name: string;
  type: string;
  description: string;
  scope: string;
  required: boolean;
  optionsText: string;
};

type EmojiCategoryPreset = {
  label?: string;
  match?: string;
  icon?: string;
  emojis?: string;
};

const route = useRoute();
const router = useRouter();
const fileStore = useFileStore();

const tabs: { key: TabKey; label: string; icon: unknown }[] = [
  { key: 'parameters', label: '参数设置', icon: Setting },
  { key: 'search', label: '全文搜索', icon: Search },
  { key: 'icons', label: '文件图标', icon: Star },
  { key: 'apps', label: '文件浏览应用', icon: Connection },
  { key: 'custom', label: '自定义属性', icon: Tickets },
];

const units = ['B', 'KB', 'MB', 'GB', 'TB'] as const;

const defaultEmojiCategoryPresets = [
  {
    label: '图片文件',
    match: 'jpg,jpeg,png,gif,webp,svg,image/*',
    icon: '😀',
    emojis:
      '😀,😃,😄,😁,😆,😅,🤣,😂,🙂,🙃,🫠,😉,😊,😇,🥰,😍,🤩,😘,😗,😚,😙,🥲,😋,😛,😜,🤪,😝,🤑,🤗,🤭,🫢,🫣,🤫,🤔,🫡,🤐,🤨,😐,😑,😶,😶‍🌫️,😏,😒,🙄,😬,😮‍💨,🤥,😌,😔,😪,🤤,😴,😷,🤒,🤕,🤢,🤮,🤧,🥵,🥶,🥴,😵,😵‍💫,🤯,🤠,🥳,🥸,😎,🤓,🧐,😕,🫤,😟,🙁,😮,😯,😲,😳,🥺,🥹,😦,😧,😨,😰,😥,😢,😭,😱,😖,😣,😞,😓,😩,😫,🥱,😤,😡,😠,🤬,😈,👿,💀,☠️,💩,🤡,👹,👺,👻,👽,👾,🤖,😺,😸,😹,😻,😼,😽,🙀,😿,😾,🙈,🙉,🙊,💋,💌,💘,💝,💖,💗,💓,💞,💕,💟,💔,❤️‍🔥,❤️‍🩹,❤️,🧡,💛,💚,💙,💜,🤎,🖤,🤍,💯,💢,💥,💫,💦,💨,🕳️,💣,💬,👁️‍🗨️,🗨️,🗯️,💭,💤',
  },
  {
    label: '音频文件',
    match: 'mp3,flac,wav,aac,ogg,audio/*',
    icon: '👋',
    emojis:
      '👋,🤚,🖐️,✋,🖖,🫱,🫲,🫳,🫴,👌,🤌,🤏,✌️,🤞,🫰,🤟,🤘,🤙,👈,👉,👆,🖕,👇,☝️,🫵,👍,👎,✊,👊,🤛,🤜,👏,🙌,🫶,👐,🤲,🤝,🙏,✍️,💅,🤳,💪,🦾,🦿,🦵,🦶,👂,🦻,👃,🧠,🫀,🫁,🦷,🦴,👀,👁️,👅,👄,🫦,👶,🧒,👦,👧,🧑,👱,👨,🧔,🧔‍♂️,🧔‍♀️,👨‍🦰,👨‍🦱,👨‍🦳,👨‍🦲,👩,👩‍🦰,🧑‍🦰,👩‍🦱,🧑‍🦱,👩‍🦳,🧑‍🦳,👩‍🦲,🧑‍🦲,👱‍♀️,👱‍♂️,🧓,👴,👵,🙍,🙍‍♂️,🙍‍♀️,🙎,🙎‍♂️,🙎‍♀️,🙅,🙅‍♂️,🙅‍♀️,🙆,🙆‍♂️,🙆‍♀️,💁,💁‍♂️,💁‍♀️,🙋,🙋‍♂️,🙋‍♀️,🧏,🧏‍♂️,🧏‍♀️,🙇,🙇‍♂️,🙇‍♀️,🤦,🤦‍♂️,🤦‍♀️,🤷,🤷‍♂️,🤷‍♀️,🧑‍⚕️,👨‍⚕️,👩‍⚕️,🧑‍🎓,👨‍🎓,👩‍🎓,🧑‍🏫,👨‍🏫,👩‍🏫,🧑‍⚖️,👨‍⚖️,👩‍⚖️,🧑‍🌾,👨‍🌾,👩‍🌾,🧑‍🍳,👨‍🍳,👩‍🍳,🧑‍🔧,👨‍🔧,👩‍🔧,🧑‍🏭,👨‍🏭,👩‍🏭,🧑‍💼,👨‍💼,👩‍💼,🧑‍🔬,👨‍🔬,👩‍🔬,🧑‍💻,👨‍💻,👩‍💻,🧑‍🎤,👨‍🎤,👩‍🎤,🧑‍🎨,👨‍🎨,👩‍🎨,🧑‍✈️,👨‍✈️,👩‍✈️,🧑‍🚀,👨‍🚀,👩‍🚀,🧑‍🚒,👨‍🚒,👩‍🚒,👮,👮‍♂️,👮‍♀️,🕵️,🕵️‍♂️,🕵️‍♀️,💂,💂‍♂️,💂‍♀️,🥷,👷,👷‍♂️,👷‍♀️,🫅,🤴,👸,👳,👳‍♂️,👳‍♀️,👲,🧕,🤵,🤵‍♂️,🤵‍♀️,👰,👰‍♂️,👰‍♀️,🤰,🫃,🫄,🤱,👩‍🍼,👨‍🍼,🧑‍🍼,👼,🎅,🤶,🧑‍🎄,🦸,🦸‍♂️,🦸‍♀️,🦹,🦹‍♂️,🦹‍♀️,🧙,🧙‍♂️,🧙‍♀️,🧚,🧚‍♂️,🧚‍♀️,🧛,🧛‍♂️,🧛‍♀️,🧜,🧜‍♂️,🧜‍♀️,🧝,🧝‍♂️,🧝‍♀️,🧞,🧞‍♂️,🧞‍♀️,🧟,🧟‍♂️,🧟‍♀️,🧌,💆,💆‍♂️,💆‍♀️,💇,💇‍♂️,💇‍♀️,🚶,🚶‍♂️,🚶‍♀️,🧍,🧍‍♂️,🧍‍♀️,🧎,🧎‍♂️,🧎‍♀️,🧑‍🦯,👨‍🦯,👩‍🦯,🧑‍🦼,👨‍🦼,👩‍🦼,🧑‍🦽,👨‍🦽,👩‍🦽,🏃,🏃‍♂️,🏃‍♀️,💃,🕺,🕴️,👯,👯‍♂️,👯‍♀️,🧖,🧖‍♂️,🧖‍♀️,🧗,🧗‍♂️,🧗‍♀️,🤺,🏇,⛷️,🏂,🏌️,🏌️‍♂️,🏌️‍♀️,🏄,🏄‍♂️,🏄‍♀️,🚣,🚣‍♂️,🚣‍♀️,🏊,🏊‍♂️,🏊‍♀️,⛹️,⛹️‍♂️,⛹️‍♀️,🏋️,🏋️‍♂️,🏋️‍♀️,🚴,🚴‍♂️,🚴‍♀️,🚵,🚵‍♂️,🚵‍♀️,🤸,🤸‍♂️,🤸‍♀️,🤼,🤼‍♂️,🤼‍♀️,🤽,🤽‍♂️,🤽‍♀️,🤾,🤾‍♂️,🤾‍♀️,🤹,🤹‍♂️,🤹‍♀️,🧘,🧘‍♂️,🧘‍♀️,🛀,🛌,🧑‍🤝‍🧑,👭,👫,👬,💏,👩‍❤️‍💋‍👨,👨‍❤️‍💋‍👨,👩‍❤️‍💋‍👩,💑,👩‍❤️‍👨,👨‍❤️‍👨,👩‍❤️‍👩,👪,👨‍👩‍👦,👨‍👩‍👧,👨‍👩‍👧‍👦,👨‍👩‍👦‍👦,👨‍👩‍👧‍👧,👨‍👨‍👦,👨‍👨‍👧,👨‍👨‍👧‍👦,👨‍👨‍👦‍👦,👨‍👨‍👧‍👧,👩‍👩‍👦,👩‍👩‍👧,👩‍👩‍👧‍👦,👩‍👩‍👦‍👦,👩‍👩‍👧‍👧,👨‍👦,👨‍👦‍👦,👨‍👧,👨‍👧‍👦,👨‍👧‍👧,👩‍👦,👩‍👦‍👦,👩‍👧,👩‍👧‍👦,👩‍👧‍👧,🗣️,👤,👥,🫂,👣,🦰,🦱,🦳,🦲',
  },
  {
    label: '视频文件',
    match: 'mp4,mkv,mov,webm,video/*',
    icon: '🐵',
    emojis:
      '🐵,🐒,🦍,🦧,🐶,🐕,🦮,🐕‍🦺,🐩,🐺,🦊,🦝,🐱,🐈,🐈‍⬛,🦁,🐯,🐅,🐆,🐴,🐎,🦄,🦓,🦌,🦬,🐮,🐂,🐃,🐄,🐷,🐖,🐗,🐽,🐏,🐑,🐐,🐪,🐫,🦙,🦒,🐘,🦣,🦏,🦛,🐭,🐁,🐀,🐹,🐰,🐇,🐿️,🦫,🦔,🦇,🐻,🐻‍❄️,🐨,🐼,🦥,🦦,🦨,🦘,🦡,🐾,🦃,🐔,🐓,🐣,🐤,🐥,🐦,🐧,🕊️,🦅,🦆,🦢,🦉,🦤,🪶,🦩,🦚,🦜,🐸,🐊,🐢,🦎,🐍,🐲,🐉,🦕,🦖,🐳,🐋,🐬,🦭,🐟,🐠,🐡,🦈,🐙,🐚,🪸,🐌,🦋,🐛,🐜,🐝,🪲,🐞,🦗,🪳,🕷️,🕸️,🦂,🦟,🪰,🪱,🦠,💐,🌸,💮,🪷,🏵️,🌹,🥀,🌺,🌻,🌼,🌷,🌱,🪴,🌲,🌳,🌴,🌵,🌾,🌿,☘️,🍀,🍁,🍂,🍃,🪹,🪺',
  },
  {
    label: '文档文件',
    match: 'pdf,doc,docx,xls,xlsx,ppt,pptx',
    icon: '🍇',
    emojis:
      '🍇,🍈,🍉,🍊,🍋,🍌,🍍,🥭,🍎,🍏,🍐,🍑,🍒,🍓,🫐,🥝,🍅,🫒,🥥,🥑,🍆,🥔,🥕,🌽,🌶️,🫑,🥒,🥬,🥦,🧄,🧅,🍄,🥜,🫘,🌰,🍞,🥐,🥖,🫓,🥨,🥯,🥞,🧇,🧀,🍖,🍗,🥩,🥓,🍔,🍟,🍕,🌭,🥪,🌮,🌯,🫔,🥙,🧆,🥚,🍳,🥘,🍲,🫕,🥣,🥗,🍿,🧈,🧂,🥫,🍱,🍘,🍙,🍚,🍛,🍜,🍝,🍠,🍢,🍣,🍤,🍥,🥮,🍡,🥟,🥠,🥡,🦀,🦞,🦐,🦑,🦪,🍦,🍧,🍨,🍩,🍪,🎂,🍰,🧁,🥧,🍫,🍬,🍭,🍮,🍯,🍼,🥛,☕,🫖,🍵,🍶,🍾,🍷,🍸,🍹,🍺,🍻,🥂,🥃,🫗,🥤,🧋,🧃,🧉,🧊,🥢,🍽️,🍴,🥄,🔪,🫙,🏺',
  },
  {
    icon: '🌍',
    emojis:
      '🌍,🌎,🌏,🌐,🗺️,🗾,🧭,🏔️,⛰️,🌋,🗻,🏕️,🏖️,🏜️,🏝️,🏞️,🏟️,🏛️,🏗️,🧱,🪨,🪵,🛖,🏘️,🏚️,🏠,🏡,🏢,🏣,🏤,🏥,🏦,🏨,🏩,🏪,🏫,🏬,🏭,🏯,🏰,💒,🗼,🗽,⛪,🕌,🛕,🕍,⛩️,🕋,⛲,⛺,🌁,🌃,🏙️,🌄,🌅,🌆,🌇,🌉,♨️,🎠,🛝,🎡,🎢,💈,🎪,🚂,🚃,🚄,🚅,🚆,🚇,🚈,🚉,🚊,🚝,🚞,🚋,🚌,🚍,🚎,🚐,🚑,🚒,🚓,🚔,🚕,🚖,🚗,🚘,🚙,🛻,🚚,🚛,🚜,🏎️,🏍️,🛵,🦽,🦼,🛺,🚲,🛴,🛹,🛼,🚏,🛣️,🛤️,🛢️,⛽,🛞,🚨,🚥,🚦,🛑,🚧,⚓,🛟,⛵,🛶,🚤,🛳️,⛴️,🛥️,🚢,✈️,🛩️,🛫,🛬,🪂,💺,🚁,🚟,🚠,🚡,🛰️,🚀,🛸,🛎️,🧳,⌛,⏳,⌚,⏰,⏱️,⏲️,🕰️,🕛,🕧,🕐,🕜,🕑,🕝,🕒,🕞,🕓,🕟,🕔,🕠,🕕,🕡,🕖,🕢,🕗,🕣,🕘,🕤,🕙,🕥,🕚,🕦,🌑,🌒,🌓,🌔,🌕,🌖,🌗,🌘,🌙,🌚,🌛,🌜,🌡️,☀️,🌝,🌞,🪐,⭐,🌟,🌠,🌌,☁️,⛅,⛈️,🌤️,🌥️,🌦️,🌧️,🌨️,🌩️,🌪️,🌫️,🌬️,🌀,🌈,🌂,☂️,☔,⛱️,⚡,❄️,☃️,⛄,☄️,🔥,💧,🌊',
  },
  {
    icon: '🎃',
    emojis:
      '🎃,🎄,🎆,🎇,🧨,✨,🎈,🎉,🎊,🎋,🎍,🎎,🎏,🎐,🎑,🧧,🎀,🎁,🎗️,🎟️,🎫,🎖️,🏆,🏅,🥇,🥈,🥉,⚽,⚾,🥎,🏀,🏐,🏈,🏉,🎾,🥏,🎳,🏏,🏑,🏒,🥍,🏓,🏸,🥊,🥋,🥅,⛳,⛸️,🎣,🤿,🎽,🎿,🛷,🥌,🎯,🪀,🪁,🎱,🔮,🪄,🧿,🪬,🎮,🕹️,🎰,🎲,🧩,🧸,🪅,🪩,🪆,♠️,♥️,♦️,♣️,♟️,🃏,🀄,🎴,🎭,🖼️,🎨,🧵,🪡,🧶,🪢',
  },
  {
    icon: '👓',
    emojis:
      '👓,🕶️,🥽,🥼,🦺,👔,👕,👖,🧣,🧤,🧥,🧦,👗,👘,🥻,🩱,🩲,🩳,👙,👚,👛,👜,👝,🛍️,🎒,🩴,👞,👟,🥾,🥿,👠,👡,🩰,👢,👑,👒,🎩,🎓,🧢,🪖,⛑️,📿,💄,💍,💎,🔇,🔈,🔉,🔊,📢,📣,📯,🔔,🔕,🎼,🎵,🎶,🎙️,🎚️,🎛️,🎤,🎧,📻,🎷,🪗,🎸,🎹,🎺,🎻,🪕,🥁,🪘,📱,📲,☎️,📞,📟,📠,🔋,🪫,🔌,💻,🖥️,🖨️,⌨️,🖱️,🖲️,💽,💾,💿,📀,🧮,🎥,🎞️,📽️,🎬,📺,📷,📸,📹,📼,🔍,🔎,🕯️,💡,🔦,🏮,🪔,📔,📕,📖,📗,📘,📙,📚,📓,📒,📃,📜,📄,📰,🗞️,📑,🔖,🏷️,💰,🪙,💴,💵,💶,💷,💸,💳,🧾,💹,✉️,📧,📨,📩,📤,📥,📦,📫,📪,📬,📭,📮,🗳️,✏️,✒️,🖋️,🖊️,🖌️,🖍️,📝,💼,📁,📂,🗂️,📅,📆,🗒️,🗓️,📇,📈,📉,📊,📋,📌,📍,📎,🖇️,📏,📐,✂️,🗃️,🗄️,🗑️,🔒,🔓,🔏,🔐,🔑,🗝️,🔨,🪓,⛏️,⚒️,🛠️,🗡️,⚔️,🔫,🪃,🏹,🛡️,🪚,🔧,🪛,🔩,⚙️,🗜️,⚖️,🦯,🔗,⛓️,🪝,🧰,🧲,🪜,⚗️,🧪,🧫,🧬,🔬,🔭,📡,💉,🩸,💊,🩹,🩼,🩺,🩻,🚪,🛗,🪞,🪟,🛏️,🛋️,🪑,🚽,🪠,🚿,🛁,🪤,🪒,🧴,🧷,🧹,🧺,🧻,🪣,🧼,🫧,🪥,🧽,🧯,🛒,🚬,⚰️,🪦,⚱️,🗿,🪧,🪪',
  },
  {
    icon: '🅰️',
    emojis:
      '🏧,🚮,🚰,♿,🚹,🚺,🚻,🚼,🚾,🛂,🛃,🛄,🛅,⚠️,🚸,⛔,🚫,🚳,🚭,🚯,🚱,🚷,📵,🔞,☢️,☣️,⬆️,↗️,➡️,↘️,⬇️,↙️,⬅️,↖️,↕️,↔️,↩️,↪️,⤴️,⤵️,🔃,🔄,🔙,🔚,🔛,🔜,🔝,🛐,⚛️,🕉️,✡️,☸️,☯️,✝️,☦️,☪️,☮️,🕎,🔯,♈,♉,♊,♋,♌,♍,♎,♏,♐,♑,♒,♓,⛎,🔀,🔁,🔂,▶️,⏩,⏭️,⏯️,◀️,⏪,⏮️,🔼,⏫,🔽,⏬,⏸️,⏹️,⏺️,⏏️,🎦,🔅,🔆,📶,📳,📴,♀️,♂️,⚧️,✖️,➕,➖,➗,🟰,♾️,‼️,⁉️,❓,❔,❕,❗,〰️,💱,💲,⚕️,♻️,⚜️,🔱,📛,🔰,⭕,✅,☑️,✔️,❌,❎,➰,➿,〽️,✳️,✴️,❇️,©️,®️,™️,#️⃣,*️⃣,0️⃣,1️⃣,2️⃣,3️⃣,4️⃣,5️⃣,6️⃣,7️⃣,8️⃣,9️⃣,🔟,🔠,🔡,🔢,🔣,🔤,🅰️,🆎,🅱️,🆑,🆒,🆓,ℹ️,🆔,Ⓜ️,🆕,🆖,🅾️,🆗,🅿️,🆘,🆙,🆚,🈁,🈂️,🈷️,🈶,🈯,🉐,🈹,🈚,🈲,🉑,🈸,🈴,🈳,㊗️,㊙️,🈺,🈵,🔴,🟠,🟡,🟢,🔵,🟣,🟤,⚫,⚪,🟥,🟧,🟨,🟩,🟦,🟪,🟫,⬛,⬜,◼️,◻️,◾,◽,▪️,▫️,🔶,🔷,🔸,🔹,🔺,🔻,💠,🔘,🔳,🔲',
  },
  {
    icon: '🏁',
    emojis:
      '🏁,🚩,🎌,🏴,🏳️,🏳️‍🌈,🏳️‍⚧️,🏴‍☠️,🇦🇨,🇦🇩,🇦🇪,🇦🇫,🇦🇬,🇦🇮,🇦🇱,🇦🇲,🇦🇴,🇦🇶,🇦🇷,🇦🇸,🇦🇹,🇦🇺,🇦🇼,🇦🇽,🇦🇿,🇧🇦,🇧🇧,🇧🇩,🇧🇪,🇧🇫,🇧🇬,🇧🇭,🇧🇮,🇧🇯,🇧🇱,🇧🇲,🇧🇳,🇧🇴,🇧🇶,🇧🇷,🇧🇸,🇧🇹,🇧🇻,🇧🇼,🇧🇾,🇧🇿,🇨🇦,🇨🇨,🇨🇩,🇨🇫,🇨🇬,🇨🇭,🇨🇮,🇨🇰,🇨🇱,🇨🇲,🇨🇳,🇨🇴,🇨🇵,🇨🇷,🇨🇺,🇨🇻,🇨🇼,🇨🇽,🇨🇾,🇨🇿,🇩🇪,🇩🇬,🇩🇯,🇩🇰,🇩🇲,🇩🇴,🇩🇿,🇪🇦,🇪🇨,🇪🇪,🇪🇬,🇪🇭,🇪🇷,🇪🇸,🇪🇹,🇪🇺,🇫🇮,🇫🇯,🇫🇰,🇫🇲,🇫🇴,🇫🇷,🇬🇦,🇬🇧,🇬🇩,🇬🇪,🇬🇫,🇬🇬,🇬🇭,🇬🇮,🇬🇱,🇬🇲,🇬🇳,🇬🇵,🇬🇶,🇬🇷,🇬🇸,🇬🇹,🇬🇺,🇬🇼,🇬🇾,🇭🇰,🇭🇲,🇭🇳,🇭🇷,🇭🇹,🇭🇺,🇮🇨,🇮🇩,🇮🇪,🇮🇱,🇮🇲,🇮🇳,🇮🇴,🇮🇶,🇮🇷,🇮🇸,🇮🇹,🇯🇪,🇯🇲,🇯🇴,🇯🇵,🇰🇪,🇰🇬,🇰🇭,🇰🇮,🇰🇲,🇰🇳,🇰🇵,🇰🇷,🇰🇼,🇰🇾,🇰🇿,🇱🇦,🇱🇧,🇱🇨,🇱🇮,🇱🇰,🇱🇷,🇱🇸,🇱🇹,🇱🇺,🇱🇻,🇱🇾,🇲🇦,🇲🇨,🇲🇩,🇲🇪,🇲🇫,🇲🇬,🇲🇭,🇲🇰,🇲🇱,🇲🇲,🇲🇳,🇲🇴,🇲🇵,🇲🇶,🇲🇷,🇲🇸,🇲🇹,🇲🇺,🇲🇻,🇲🇼,🇲🇽,🇲🇾,🇲🇿,🇳🇦,🇳🇨,🇳🇪,🇳🇫,🇳🇬,🇳🇮,🇳🇱,🇳🇴,🇳🇵,🇳🇷,🇳🇺,🇳🇿,🇴🇲,🇵🇦,🇵🇪,🇵🇫,🇵🇬,🇵🇭,🇵🇰,🇵🇱,🇵🇲,🇵🇳,🇵🇷,🇵🇸,🇵🇹,🇵🇼,🇵🇾,🇶🇦,🇷🇪,🇷🇴,🇷🇸,🇷🇺,🇷🇼,🇸🇦,🇸🇧,🇸🇨,🇸🇩,🇸🇪,🇸🇬,🇸🇭,🇸🇮,🇸🇯,🇸🇰,🇸🇱,🇸🇲,🇸🇳,🇸🇴,🇸🇷,🇸🇸,🇸🇹,🇸🇻,🇸🇽,🇸🇾,🇸🇿,🇹🇦,🇹🇨,🇹🇩,🇹🇫,🇹🇬,🇹🇭,🇹🇯,🇹🇰,🇹🇱,🇹🇲,🇹🇳,🇹🇴,🇹🇷,🇹🇹,🇹🇻,🇹🇼,🇹🇿,🇺🇦,🇺🇬,🇺🇲,🇺🇳,🇺🇸,🇺🇾,🇺🇿,🇻🇦,🇻🇨,🇻🇪,🇻🇬,🇻🇮,🇻🇳,🇻🇺,🇼🇫,🇼🇸,🇽🇰,🇾🇪,🇾🇹,🇿🇦,🇿🇲,🇿🇼,🏴',
  },
] as const;

function cloneEmojiCategoryPreset(
  preset: EmojiCategoryPreset,
  index = 0,
): EmojiCategoryEditor {
  return {
    id: uid(`emoji-category-${index}`),
    label: typeof preset.label === 'string' && preset.label.trim() ? preset.label.trim() : `分类 ${index + 1}`,
    match: typeof preset.match === 'string' ? preset.match.trim() : '',
    icon: typeof preset.icon === 'string' && preset.icon.trim() ? preset.icon.trim() : '🙂',
    emojisText: normalizeEmojiListText(typeof preset.emojis === 'string' ? preset.emojis : ''),
  };
}

function defaultEmojiOptionsEditor(): EmojiOptionsEditor {
  return {
    enabled: true,
    showInList: true,
    fallbackUnknown: true,
    folderEmoji: '📁',
    unknownEmoji: '🗂️',
    categories: defaultEmojiCategoryPresets.map((preset, index) => cloneEmojiCategoryPreset(preset, index)),
  };
}

function normalizeEmojiListText(text: string) {
  return text
    .replace(/[，、；;]/g, ',')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .join(', ');
}

function emojiListTextFromUnknown(value: unknown) {
  if (Array.isArray(value)) {
    return value.filter((entry): entry is string => typeof entry === 'string' && entry.trim().length > 0).join(', ');
  }
  return typeof value === 'string' ? value : '';
}

function defaultEmojiOptionsPayload() {
  const defaults = defaultEmojiOptionsEditor();
  return {
    enabled: defaults.enabled,
    showInList: defaults.showInList,
    fallbackUnknown: defaults.fallbackUnknown,
    folderEmoji: defaults.folderEmoji,
    unknownEmoji: defaults.unknownEmoji,
    categories: defaults.categories.map((category) => ({
      icon: category.icon,
      label: category.label,
      match: category.match,
      emojis: normalizeEmojiListText(category.emojisText)
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
    })),
  };
}

function defaultCustomPropertiesPayload(): CustomPropertySeed[] {
  return [
    {
      name: '描述',
      type: 'text',
      description: '用于补充文件说明、摘要或业务备注信息。',
      scope: '文件',
      required: false,
      optionsText: '',
    },
    {
      name: '评级',
      type: 'rating',
      description: '用于标记文件的重要程度与优先级。',
      scope: '文件',
      required: false,
      optionsText: '0',
    },
  ];
}

function createCustomPropertyEditor(seed: Partial<CustomPropertySeed>, index: number): CustomPropertyEditor {
  const propertyType = typeof seed.type === 'string' && seed.type.trim() ? seed.type.trim() : 'text';
  return {
    id: uid(`property-${index}`),
    name: seed.name || `属性 ${index + 1}`,
    type: propertyType,
    description: seed.description || '',
    scope: seed.scope || '文件 / 文件夹',
    required: typeof seed.required === 'boolean' ? seed.required : false,
    optionsText: seed.optionsText || '',
  };
}

function ensureCustomPropertySeeds(properties: CustomPropertyEditor[]) {
  const next = [...properties];
  const hasDescription = next.some((item) => item.name === '描述' || item.type === 'text');
  const hasRating = next.some((item) => item.type === 'rating' || item.name === '评级' || item.name === '评分');

  if (!hasDescription) {
    next.unshift(createCustomPropertyEditor(defaultCustomPropertiesPayload()[0], next.length));
  }

  if (!hasRating) {
    next.push(createCustomPropertyEditor(defaultCustomPropertiesPayload()[1], next.length));
  }

  return next;
}

function defaultFileSystemForm(): FileSystemSettingsPayload {
  return {
    online_editor_size: 50,
    online_editor_unit: 'MB',
    recycle_scan_interval: '@every 33m',
    blob_recycle_interval: '@every 15m',
    static_cache_ttl: 86400,
    list_pagination_mode: 'cursor',
    max_page_size: 2000,
    max_batch_action_size: 3000,
    max_recursive_search: 65535,
    map_provider: 'osm-leaflet',
    mime_map: '',
    image_query: '',
    video_query: '',
    audio_query: '',
    document_query: '',
    file_icon_rules: '[]',
    emoji_options: serializeJson(defaultEmojiOptionsPayload()),
    browser_apps: '[]',
    custom_properties: serializeJson(defaultCustomPropertiesPayload()),
    master_key_storage: 'database',
    show_encryption_status: true,
    enable_event_push: true,
    offline_ttl: 1209600,
    debounce_delay: 5,
    server_side_download_session_ttl: 600,
    upload_session_ttl: 86400,
    slave_api_sign_ttl: 60,
    directory_stat_ttl: 300,
    max_chunk_retry: 5,
    cache_chunks_for_retry: true,
    transfer_parallelism: 4,
    oauth_refresh_interval: '@every 230h',
    wopi_session_ttl: 36000,
    blob_signed_url_ttl: 3600,
    blob_signed_url_reuse_ttl: 600,
  };
}

function defaultSearchForm(): FullTextSearchSettingsPayload {
  return {
    enabled: true,
    meili_endpoint: 'http://localhost:7700',
    api_key: '',
    result_page_size: 5,
    ai_search: false,
    tika_endpoint: 'http://localhost:9998',
    extensions: 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,md',
    max_file_size: 25,
    max_file_size_unit: 'MB',
    chunk_size: 2000,
    chunk_unit: 'B',
    index_notes: '',
  };
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value));
}

function safeParse(input: string, fallback: unknown): unknown {
  try {
    return JSON.parse(input);
  } catch {
    return fallback;
  }
}

function asArray(input: string): ParsedRecord[] {
  const parsed = safeParse(input, []);
  return Array.isArray(parsed) ? parsed.filter((item): item is ParsedRecord => !!item && typeof item === 'object') : [];
}

function createBrowserAppEditor(seed: Partial<BrowserAppEditor>, index: number): BrowserAppEditor {
  const id = String(seed.id || uid(`app-${index}`));
  return {
    id,
    name: seed.name || `应用 ${index + 1}`,
    icon: seed.icon || '🧩',
    extensions: seed.extensions || '',
    action: seed.action || '',
    description: seed.description || '',
    background: seed.background || '',
    type: seed.type || '内置应用',
    platform: seed.platform || 'all',
    createMapping: seed.createMapping || '无',
    enabled: typeof seed.enabled === 'boolean' ? seed.enabled : true,
    openInNewWindow: typeof seed.openInNewWindow === 'boolean' ? seed.openInNewWindow : false,
    maxSize: typeof seed.maxSize === 'number' && seed.maxSize > 0 ? seed.maxSize : 100,
    maxSizeUnit: seed.maxSizeUnit || 'MB',
  };
}

function mapBrowserAppsFromGroups(groups: BrowserAppGroupPayload[]): BrowserAppEditor[] {
  return groups.flatMap((group, groupIndex) =>
    Array.isArray(group.items)
      ? group.items.map((item, itemIndex) =>
          createBrowserAppEditor(
            {
              id: String(item.id || `${group.id || groupIndex + 1}-${itemIndex + 1}`),
              name: item.name,
              icon: item.icon,
              extensions: item.extensions,
              action: '',
              description: group.description || '',
              background: item.accent || '',
              type: item.type || '内置应用',
              platform: String(item.platform || 'all'),
              createMapping: item.create_mapping || '无',
              enabled: item.enabled ?? true,
              openInNewWindow: item.open_in_new_window ?? false,
              maxSize: item.max_size ?? 100,
              maxSizeUnit: item.max_size_unit || 'MB',
            },
            itemIndex,
          ),
        )
      : [],
  );
}

function pickString(record: ParsedRecord, keys: string[], fallback = '') {
  for (const key of keys) {
    const value = record[key];
    if (typeof value === 'string' && value.trim()) {
      return value.trim();
    }
  }
  return fallback;
}

function pickBoolean(record: ParsedRecord, keys: string[], fallback = false) {
  for (const key of keys) {
    const value = record[key];
    if (typeof value === 'boolean') {
      return value;
    }
  }
  return fallback;
}

function pickStringArray(record: ParsedRecord, keys: string[]) {
  for (const key of keys) {
    const value = record[key];
    if (Array.isArray(value)) {
      return value.filter((item): item is string => typeof item === 'string' && item.trim().length > 0);
    }
  }
  return [] as string[];
}

function serializeJson(value: unknown) {
  return JSON.stringify(value, null, 2);
}

function uid(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function defaultAppBackground(seed: string) {
  const palettes = [
    'linear-gradient(135deg, #2563eb, #38bdf8)',
    'linear-gradient(135deg, #f97316, #facc15)',
    'linear-gradient(135deg, #16a34a, #4ade80)',
    'linear-gradient(135deg, #7c3aed, #c084fc)',
    'linear-gradient(135deg, #0f766e, #2dd4bf)',
  ];
  const index = seed.length % palettes.length;
  return palettes[index];
}

function normalizeAppName(input: string) {
  return String(input || '').trim().toLowerCase();
}

function resolveAppBrandMeta(app: BrowserAppEditor) {
  const name = normalizeAppName(app.name);

  if (name.includes('artplayer')) {
    return {
      key: 'artplayer',
      glyph: '▶',
      background: 'linear-gradient(135deg, #eef8ff, #f6f4ff)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="6" y="6" width="36" height="36" rx="10" fill="#0f172a"/>
          <path d="M17 13h5.8l12.6 11L22.8 35H17l12.5-11L17 13Z" fill="#38bdf8"/>
          <path d="M20.6 16.2h4.2L33 24l-8.2 7.8h-4.2l8.1-7.8-8.1-7.8Z" fill="#818cf8"/>
        </svg>`,
    };
  }
  if (name.includes('markdown')) {
    return {
      key: 'markdown',
      glyph: 'M↓',
      background: 'linear-gradient(135deg, #f8fafc, #eef2f7)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="6" y="10" width="36" height="28" rx="7" fill="#1f2937"/>
          <rect x="8.5" y="12.5" width="31" height="23" rx="5" fill="none" stroke="#9ca3af" stroke-width="1.2"/>
          <path d="M13 31V18l5.2 6.1L23.4 18v13" fill="none" stroke="#fff" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M28 24h7" stroke="#fff" stroke-width="3.2" stroke-linecap="round"/>
          <path d="M31.5 19v11" stroke="#fff" stroke-width="3.2" stroke-linecap="round"/>
          <path d="M28.5 27.5l3 3 3-3" fill="none" stroke="#fff" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>`,
    };
  }
  if (name.includes('draw.io') || name.includes('drawio')) {
    return {
      key: 'drawio',
      glyph: '◫',
      background: 'linear-gradient(135deg, #fff7ed, #fff1f2)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="9" fill="#f97316"/>
          <rect x="13" y="14" width="8.5" height="8.5" rx="2.4" fill="#fff"/>
          <rect x="26.5" y="14" width="8.5" height="8.5" rx="2.4" fill="#fde68a"/>
          <rect x="19.75" y="26.2" width="8.5" height="8.5" rx="2.4" fill="#fff7ed"/>
          <path d="M21.5 18.3h5M17.3 22.7l6.6 5m6.8-5l-6.6 5" stroke="#fff" stroke-width="2.6" stroke-linecap="round"/>
        </svg>`,
    };
  }
  if (name.includes('图片')) {
    return {
      key: 'image',
      glyph: '◩',
      background: 'linear-gradient(135deg, #fff5f5, #fef2f2)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="9" fill="#2563eb"/>
          <rect x="11" y="11" width="26" height="26" rx="6" fill="#eff6ff"/>
          <circle cx="30.5" cy="18" r="3.4" fill="#f59e0b"/>
          <path d="M14.5 31.5l6.8-8 5.7 4.7 4.8-6.2 6.2 9.5H14.5Z" fill="#2563eb"/>
          <path d="M14 32h20.2l-4.7-7.1-4.3 5.1-4.8-4.1L14 32Z" fill="#22c55e"/>
        </svg>`,
    };
  }
  if (name.includes('monaco')) {
    return {
      key: 'monaco',
      glyph: 'M',
      background: 'linear-gradient(135deg, #eff6ff, #f0f9ff)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <path d="M11 10.5l8.3 3.8v19.4L11 37.5V10.5Z" fill="#007acc"/>
          <path d="M37 10.5l-8.3 3.8v19.4L37 37.5V10.5Z" fill="#1f9cf0"/>
          <path d="M19.3 14.3l9.4-2.8v25l-9.4-2.8V14.3Z" fill="#005fb8"/>
          <path d="M19.3 18.2l5.3 4.1-5.3 4.1 5.3 4.1" fill="none" stroke="#fff" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>`,
    };
  }
  if (name.includes('photopea')) {
    return {
      key: 'photopea',
      glyph: 'P',
      background: 'linear-gradient(135deg, #ecfeff, #f0fdfa)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="9" fill="#143c43"/>
          <path d="M26.8 13.7c-7 0-12.5 5.5-12.5 12.4 0 6.2 4.7 11.1 10.9 11.1 5.5 0 9.3-3 9.3-8.1 0-5.1-3.5-8.2-8-8.2-3.1 0-5.6 1.4-7.2 3.8.1-4.5 2.8-7.2 7.6-7.2 2.1 0 3.9.5 5.4 1.4l1.8-3.6c-2.3-1.1-5-1.6-7.3-1.6Zm-1 10c2.5 0 4.5 1.7 4.5 4.1 0 2.6-2 4.2-4.6 4.2s-4.6-1.8-4.6-4.3c0-2.3 2-4 4.7-4Z" fill="#32c997"/>
        </svg>`,
    };
  }
  if (name.includes('excalidraw')) {
    return {
      key: 'excalidraw',
      glyph: '✕',
      background: 'linear-gradient(135deg, #f5f3ff, #eef2ff)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="11" fill="#f5f3ff"/>
          <path d="M15 19.5c5.1-6.1 12.9-8.7 19.2-6.2" fill="none" stroke="#4f46e5" stroke-width="2.4" stroke-linecap="round"/>
          <path d="M14.5 29.5c5.8 5.7 15.2 6.7 20.5 2.2" fill="none" stroke="#8b5cf6" stroke-width="2.4" stroke-linecap="round"/>
          <path d="M18.5 16.5l11 14" stroke="#312e81" stroke-width="2.8" stroke-linecap="round"/>
          <path d="M31 17.5l-4.3 4.6" stroke="#312e81" stroke-width="2.8" stroke-linecap="round"/>
        </svg>`,
    };
  }
  if (name.includes('压缩')) {
    return {
      key: 'archive',
      glyph: 'ZIP',
      background: 'linear-gradient(135deg, #fff7ed, #fffbeb)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 8h16l8 8v20a4 4 0 0 1-4 4H12a4 4 0 0 1-4-4V12a4 4 0 0 1 4-4Z" fill="#f59e0b"/>
          <path d="M28 8v8h8" fill="#fdba74"/>
          <rect x="20.4" y="12" width="7.2" height="4" rx="1.2" fill="#78350f"/>
          <path d="M22.4 17.5h3.2M22.4 22h3.2M22.4 26.5h3.2M20.5 31.2h7" stroke="#fff" stroke-width="2.2" stroke-linecap="round"/>
        </svg>`,
    };
  }
  if (name.includes('音频')) {
    return {
      key: 'audio',
      glyph: '♪',
      background: 'linear-gradient(135deg, #f5f3ff, #eef2ff)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="9" fill="#6d28d9"/>
          <circle cx="18" cy="31.5" r="4.8" fill="#c4b5fd"/>
          <circle cx="31.5" cy="27.8" r="4.8" fill="#ddd6fe"/>
          <path d="M22.8 15.2v14.4a5.2 5.2 0 0 1-2.2 4.3" fill="none" stroke="#fff" stroke-width="2.8" stroke-linecap="round"/>
          <path d="M22.8 16.7 33 14v13.8" fill="none" stroke="#fff" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>`,
    };
  }
  if (name.includes('epub')) {
    return {
      key: 'epub',
      glyph: 'e',
      background: 'linear-gradient(135deg, #f7fee7, #ecfccb)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <path d="M13 10h17a5 5 0 0 1 5 5v21a2 2 0 0 1-2.8 1.8c-1.8-.8-3.7-1.2-5.7-1.2H13V10Z" fill="#65a30d"/>
          <path d="M13 10h14a5 5 0 0 1 5 5v20.3c-1.7-.6-3.4-.9-5.2-.9H13V10Z" fill="#84cc16"/>
          <path d="M18 18.5h9M18 24h9M18 29.5h7" stroke="#fff" stroke-width="2.6" stroke-linecap="round"/>
        </svg>`,
    };
  }
  if (name.includes('google docs')) {
    return {
      key: 'gdocs',
      glyph: 'G',
      background: 'linear-gradient(135deg, #eef6ff, #ecfdf5)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <path d="M14 8h14l8 8v18a6 6 0 0 1-6 6H14a4 4 0 0 1-4-4V12a4 4 0 0 1 4-4Z" fill="#1a73e8"/>
          <path d="M28 8v8h8" fill="#8ab4f8"/>
          <path d="M17 20.5h14M17 25.5h14M17 30.5h10" stroke="#fff" stroke-width="2.8" stroke-linecap="round"/>
        </svg>`,
    };
  }
  if (name.includes('microsoft office')) {
    return {
      key: 'office',
      glyph: 'O',
      background: 'linear-gradient(135deg, #f5f3ff, #eff6ff)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <rect x="7" y="7" width="34" height="34" rx="9" fill="#ea580c"/>
          <path d="M20 13.5 30.8 17v14L20 34.5l-7-4.6V18.1l7-4.6Z" fill="#fff" opacity="0.2"/>
          <path d="M22 14.8 31 18v12l-9 3.2-5.5-3.8V18.6l5.5-3.8Z" fill="#fff" opacity="0.96"/>
          <path d="M24.7 18.4 29 19.9v8.2l-4.3 1.5-2.7-1.8v-7.6l2.7-1.8Z" fill="#ea580c"/>
        </svg>`,
    };
  }
  if (name.includes('pdf')) {
    return {
      key: 'pdf',
      glyph: 'PDF',
      background: 'linear-gradient(135deg, #fff7f7, #fff7ed)',
      svg: `
        <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 8h15.5l8.5 8.5V36a4 4 0 0 1-4 4H12a4 4 0 0 1-4-4V12a4 4 0 0 1 4-4Z" fill="#dc2626"/>
          <path d="M27.5 8v8.5H36" fill="#fca5a5"/>
          <path d="M16 30.5h16" stroke="#fff" stroke-width="2.4" stroke-linecap="round"/>
          <path d="M16 18.5h5.1c2 0 3.3 1.3 3.3 3.1s-1.4 3.1-3.3 3.1H18.8v4.8H16v-11Zm10.6 0H31c3 0 5 2.2 5 5.5s-2 5.5-5 5.5h-4.4v-11Zm2.8 2.3v6.4H31c1.5 0 2.3-1.2 2.3-3.2S32.5 20.8 31 20.8h-1.6ZM16 20.8v1.7h4.4c.7 0 1.3-.3 1.3-.9s-.6-.8-1.3-.8H16Zm12.5-2.3H36v2.3h-4.8v2.1h4.2v2.3h-4.2v4.3h-2.7v-11Z" fill="#fff"/>
        </svg>`,
    };
  }

  return {
    key: 'generic',
    glyph: String(app.icon || '🧩').slice(0, 3),
    background: app.background || defaultAppBackground(app.id),
  };
}

function resolveAppTypeMeta(app: BrowserAppEditor) {
  const name = normalizeAppName(app.name);
  if (name.includes('google docs') || name.includes('microsoft office')) {
    return { label: '自定义', tone: 'is-custom' };
  }
  return { label: '内置应用', tone: 'is-built-in' };
}

function resolveAppMappingLabel(app: BrowserAppEditor) {
  const name = normalizeAppName(app.name);
  if (name.includes('markdown')) {
    return '1 个';
  }
  if (name.includes('draw.io') || name.includes('drawio')) {
    return '2 个';
  }
  if (name.includes('monaco') || name.includes('excalidraw')) {
    return '1 个';
  }
  return '无';
}

const activeTab = ref<TabKey>('parameters');
const loading = ref(false);
const saving = ref(false);
const rebuilding = ref(false);
const clearingBlobUrlCache = ref(false);
const backendEndpoint = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const latestJob = ref<QueueJobItemPayload | null>(null);
const searchJobSummary = ref('尚未读取到重建索引任务状态');
const connectionAlert = reactive({
  visible: false,
  kind: 'success' as 'success' | 'error',
  message: '',
  detail: '',
});

const fileSystemForm = reactive<FileSystemSettingsPayload>(defaultFileSystemForm());
const searchForm = reactive<FullTextSearchSettingsPayload>(defaultSearchForm());
const originalFileSystem = ref<FileSystemSettingsPayload>(defaultFileSystemForm());
const originalSearch = ref<FullTextSearchSettingsPayload>(defaultSearchForm());

const iconRules = ref<IconRuleEditor[]>([]);
const emojiOptionsEditor = reactive<EmojiOptionsEditor>(defaultEmojiOptionsEditor());
const browserApps = ref<BrowserAppEditor[]>([]);
const customProperties = ref<CustomPropertyEditor[]>([]);
const appsGroupCollapsed = ref(false);
const activeBrowserAppId = ref('');
const activeCustomPropertyId = ref('');
const customPropertyPreviewRatings = reactive<Record<string, number>>({});
const customPropertyPreviewSwitches = reactive<Record<string, boolean>>({});

const activeTabMeta = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0]);
const busy = computed(() => loading.value || saving.value || rebuilding.value);
const actionTitle = computed(() => `${activeTabMeta.value.label}支持真实保存`);
const actionDescription = computed(() => '底部操作仅影响当前页签，不会误改其他模块。');
const activeBrowserApp = computed(() => browserApps.value.find((item) => item.id === activeBrowserAppId.value) ?? null);
const activeCustomProperty = computed(() => customProperties.value.find((item) => item.id === activeCustomPropertyId.value) ?? null);

function showConnectionAlert(kind: 'success' | 'error', message: string, detail: string) {
  connectionAlert.visible = true;
  connectionAlert.kind = kind;
  connectionAlert.message = message;
  connectionAlert.detail = detail;
}

function syncActiveBrowserAppSelection(preferredId?: string) {
  const candidate = preferredId && browserApps.value.some((item) => item.id === preferredId) ? preferredId : '';
  if (candidate) {
    activeBrowserAppId.value = candidate;
    return;
  }
  if (!browserApps.value.length) {
    activeBrowserAppId.value = '';
    return;
  }
  if (!browserApps.value.some((item) => item.id === activeBrowserAppId.value)) {
    activeBrowserAppId.value = '';
  }
}

function startEditBrowserApp(id: string) {
  activeBrowserAppId.value = id;
}

function startEditCustomProperty(id: string) {
  activeCustomPropertyId.value = id;
}

function resolveCustomPropertyMeta(property: CustomPropertyEditor) {
  switch (property.type) {
    case 'textarea':
      return { label: '多行文本', glyph: '≡' };
    case 'switch':
      return { label: '开关', glyph: '◐' };
    case 'date':
      return { label: '日期', glyph: '◫' };
    case 'tags':
      return { label: '标签', glyph: '#' };
    case 'multi_select':
      return { label: '多选', glyph: '☰' };
    case 'rating':
      return { label: '评分', glyph: '★' };
    default:
      return { label: '文本', glyph: 'T' };
  }
}

function previewCustomPropertyOptions(property: CustomPropertyEditor) {
  const options = property.optionsText
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 4);

  if (options.length) {
    return options;
  }

  if (property.type === 'tags' || property.type === 'multi_select') {
    return ['示例一', '示例二'];
  }

  return [];
}

function getCustomPropertyPreviewText(property: CustomPropertyEditor) {
  if (property.description.trim()) {
    return property.description.trim();
  }
  if (property.type === 'textarea') {
    return '点击编辑更长的说明内容，适合简介、摘要和备注信息。';
  }
  return '点击编辑...';
}

function getCustomPropertyPreviewDate(property: CustomPropertyEditor) {
  const source = property.optionsText.trim();
  const matched = source.match(/\d{4}[-/]\d{1,2}[-/]\d{1,2}/);
  return matched?.[0] || '2026-04-29';
}

function getCustomPropertyPreviewRating(property: CustomPropertyEditor) {
  if (typeof customPropertyPreviewRatings[property.id] === 'number') {
    return customPropertyPreviewRatings[property.id];
  }
  const source = property.optionsText.trim();
  const parsed = Number.parseInt(source, 10);
  if (Number.isFinite(parsed) && parsed >= 0) {
    return Math.min(5, Math.max(0, parsed));
  }
  return 0;
}

function setCustomPropertyPreviewRating(id: string, rating: number) {
  customPropertyPreviewRatings[id] = Math.min(5, Math.max(1, rating));
}

function getCustomPropertyPreviewSwitch(property: CustomPropertyEditor) {
  if (typeof customPropertyPreviewSwitches[property.id] === 'boolean') {
    return customPropertyPreviewSwitches[property.id];
  }
  const normalized = `${property.description} ${property.optionsText}`.toLowerCase();
  return normalized.includes('on') || normalized.includes('true') || normalized.includes('开启');
}

function toggleCustomPropertyPreviewSwitch(id: string) {
  customPropertyPreviewSwitches[id] = !customPropertyPreviewSwitches[id];
}

function mapIconRulesFromJson(json: string): IconRuleEditor[] {
  return asArray(json).map((item, index) => ({
    id: uid(`icon-${index}`),
    label: pickString(item, ['label', 'name', 'title'], `规则 ${index + 1}`),
    icon: pickString(item, ['emoji', 'icon', 'symbol'], '📁'),
    match: pickString(item, ['match', 'extension', 'ext', 'mime'], ''),
    tint: pickString(item, ['tint', 'color', 'accent'], ''),
  }));
}

function mapBrowserAppsFromJson(json: string): BrowserAppEditor[] {
  const records = asArray(json);
  if (!records.length) {
    return [];
  }

  const maybeGroups = records.filter((item) => Array.isArray(item.items));
  if (maybeGroups.length) {
    return mapBrowserAppsFromGroups(
      maybeGroups.map((group, index) => ({
        id: Number(group.id || index + 1),
        name: pickString(group, ['name', 'label', 'title'], `应用分组 #${index + 1}`),
        description: pickString(group, ['description', 'subtitle', 'hint'], ''),
        items: Array.isArray(group.items)
          ? group.items.filter((item): item is ParsedRecord => !!item && typeof item === 'object').map((item, itemIndex) => ({
              id: Number(item.id || itemIndex + 1),
              icon: pickString(item, ['icon', 'emoji'], '🧩'),
              icon_url: pickString(item, ['icon_url', 'iconUrl'], ''),
              accent: pickString(item, ['accent', 'background', 'brand', 'gradient'], ''),
              type: pickString(item, ['type'], '内置应用'),
              name: pickString(item, ['name', 'label', 'title'], `应用 ${itemIndex + 1}`),
              extensions: pickString(item, ['extensions', 'match', 'matcher', 'mime'], ''),
              platform: pickString(item, ['platform'], 'all'),
              create_mapping: pickString(item, ['create_mapping', 'createMapping'], '无'),
              enabled: pickBoolean(item, ['enabled'], true),
              open_in_new_window: pickBoolean(item, ['open_in_new_window', 'openInNewWindow'], false),
              max_size: Number(item.max_size || item.maxSize || 100),
              max_size_unit: pickString(item, ['max_size_unit', 'maxSizeUnit'], 'MB'),
            }))
          : [],
      })),
    );
  }

  return records.map((item, index) =>
    createBrowserAppEditor(
      {
        id: uid(`app-${index}`),
        name: pickString(item, ['name', 'label', 'title'], `应用 ${index + 1}`),
        icon: pickString(item, ['icon', 'emoji'], '🧩'),
        extensions: pickString(item, ['extensions', 'match', 'matcher', 'mime'], ''),
        action: pickString(item, ['url', 'action', 'open', 'target'], ''),
        description: pickString(item, ['description', 'subtitle', 'hint'], ''),
        background: pickString(item, ['background', 'brand', 'gradient'], ''),
        type: pickString(item, ['type'], '内置应用'),
        platform: pickString(item, ['platform'], 'all'),
        createMapping: pickString(item, ['create_mapping', 'createMapping'], '无'),
        enabled: pickBoolean(item, ['enabled'], true),
        openInNewWindow: pickBoolean(item, ['open_in_new_window', 'openInNewWindow'], false),
        maxSize: Number(item.max_size || item.maxSize || 100),
        maxSizeUnit: pickString(item, ['max_size_unit', 'maxSizeUnit'], 'MB'),
      },
      index,
    ),
  );
}

function mapCustomPropertiesFromJson(json: string): CustomPropertyEditor[] {
  const records = asArray(json);
  if (!records.length) {
    return ensureCustomPropertySeeds(defaultCustomPropertiesPayload().map((item, index) => createCustomPropertyEditor(item, index)));
  }

  return ensureCustomPropertySeeds(
    records.map((item, index) =>
      createCustomPropertyEditor(
        {
          name: pickString(item, ['name', 'label', 'title'], `属性 ${index + 1}`),
          type: pickString(item, ['type', 'field_type', 'kind'], 'text'),
          description: pickString(item, ['description', 'placeholder', 'hint'], ''),
          scope: pickString(item, ['scope', 'target', 'applies_to'], '文件 / 文件夹'),
          required: pickBoolean(item, ['required', 'is_required']),
          optionsText: pickStringArray(item, ['options', 'tags', 'choices']).join(','),
        },
        index,
      ),
    ),
  );
}

function parseEmojiOptionsFromJson(json: string): EmojiOptionsEditor {
  const defaults = defaultEmojiOptionsEditor();
  const parsed = safeParse(json, {});
  const record = typeof parsed === 'object' && parsed ? (parsed as ParsedRecord) : {};
  const rawCategories = Array.isArray(record.categories)
    ? record.categories.filter((item): item is ParsedRecord => !!item && typeof item === 'object')
    : [];

  const categories = rawCategories.length
    ? rawCategories.map((item, index) => {
        const emojis = emojiListTextFromUnknown(item.emojis) || pickString(item, ['emoji', 'list'], '');

        return {
          id: uid(`emoji-category-${index}`),
          label: pickString(item, ['label', 'name', 'title'], `分类 ${index + 1}`),
          match: pickString(item, ['match', 'matches', 'extensions', 'mime'], ''),
          icon: pickString(item, ['icon', 'category'], defaults.categories[index]?.icon || '🙂'),
          emojisText: normalizeEmojiListText(emojis),
        };
      })
    : defaults.categories.map((item, index) => ({
        ...item,
        id: uid(`emoji-category-${index}`),
      }));

  return {
    enabled: pickBoolean(record, ['enabled'], defaults.enabled),
    showInList: pickBoolean(record, ['showInList', 'show_in_list'], defaults.showInList),
    fallbackUnknown: pickBoolean(record, ['fallbackUnknown', 'fallback_unknown'], defaults.fallbackUnknown),
    folderEmoji: pickString(record, ['folderEmoji', 'folder_emoji'], defaults.folderEmoji),
    unknownEmoji: pickString(record, ['unknownEmoji', 'unknown_emoji'], defaults.unknownEmoji),
    categories,
  };
}

function syncIconRulesToJson() {
  fileSystemForm.file_icon_rules = serializeJson(
    iconRules.value.map((rule) => ({
      label: rule.label,
      icon: rule.icon,
      match: rule.match,
      tint: rule.tint,
    })),
  );
}

function syncBrowserAppsToJson() {
  fileSystemForm.browser_apps = serializeJson([
    {
      id: 1,
      name: '应用分组 #1',
      description: '单个分组内汇总展示全部内置应用，右侧滚动即可浏览完整品牌图标列表。',
      items: browserApps.value.map((app, index) => ({
        id: Number.parseInt(app.id, 10) || index + 1,
        icon: app.icon,
        icon_url: '',
        accent: app.background || defaultAppBackground(app.id),
        type: app.type || '内置应用',
        name: app.name,
        extensions: app.extensions,
        platform: app.platform || 'all',
        create_mapping: app.createMapping || '无',
        enabled: app.enabled !== false,
        open_in_new_window: app.openInNewWindow === true,
        max_size: app.maxSize || 100,
        max_size_unit: app.maxSizeUnit || 'MB',
      })),
    },
  ]);
}

function syncCustomPropertiesToJson() {
  fileSystemForm.custom_properties = serializeJson(
    customProperties.value.map((property) => ({
      name: property.name,
      type: property.type,
      description: property.description,
      scope: property.scope,
      required: property.required,
      options: property.optionsText
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
    })),
  );
}

function syncEmojiOptionsToJson() {
  fileSystemForm.emoji_options = serializeJson({
    enabled: emojiOptionsEditor.enabled,
    showInList: emojiOptionsEditor.showInList,
    fallbackUnknown: emojiOptionsEditor.fallbackUnknown,
    folderEmoji: emojiOptionsEditor.folderEmoji || '📁',
    unknownEmoji: emojiOptionsEditor.unknownEmoji || '🗂️',
    categories: emojiOptionsEditor.categories.map((category) => ({
      icon: (category.icon || '🙂').trim(),
      label: (category.label || '').trim(),
      match: (category.match || '').trim(),
      emojis: normalizeEmojiListText(category.emojisText)
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
    })),
  });
}

function loadIconRulesFromJson() {
  iconRules.value = mapIconRulesFromJson(fileSystemForm.file_icon_rules);
  ElMessage.success('文件图标规则已从 JSON 载入');
}

function loadEmojiOptionsFromJson() {
  const parsed = parseEmojiOptionsFromJson(fileSystemForm.emoji_options);
  Object.assign(emojiOptionsEditor, {
    enabled: parsed.enabled,
    showInList: parsed.showInList,
    fallbackUnknown: parsed.fallbackUnknown,
    folderEmoji: parsed.folderEmoji,
    unknownEmoji: parsed.unknownEmoji,
    categories: parsed.categories,
  });
  syncEmojiOptionsToJson();
  ElMessage.success('Emoji 配置已从 JSON 载入');
}

function loadBrowserAppsFromJson() {
  browserApps.value = mapBrowserAppsFromJson(fileSystemForm.browser_apps);
  syncActiveBrowserAppSelection();
  ElMessage.success('文件浏览应用已从 JSON 载入');
}

function loadCustomPropertiesFromJson() {
  customProperties.value = mapCustomPropertiesFromJson(fileSystemForm.custom_properties);
  syncCustomPropertiesToJson();
  ElMessage.success('自定义属性已从 JSON 载入');
}

function addIconRule() {
  iconRules.value.unshift({
    id: uid('icon'),
    label: '新图标规则',
    icon: '🗂',
    match: '',
    tint: '#2563eb',
  });
  syncIconRulesToJson();
}

function removeIconRule(id: string) {
  iconRules.value = iconRules.value.filter((item) => item.id !== id);
  syncIconRulesToJson();
}

function addEmojiCategory() {
  emojiOptionsEditor.categories.unshift({
    id: uid('emoji-category'),
    label: '新分类',
    match: '',
    icon: '✨',
    emojisText: '',
  });
  syncEmojiOptionsToJson();
}

function removeEmojiCategory(id: string) {
  emojiOptionsEditor.categories = emojiOptionsEditor.categories.filter((item) => item.id !== id);
  syncEmojiOptionsToJson();
}

function moveEmojiCategory(id: string, direction: -1 | 1) {
  const index = emojiOptionsEditor.categories.findIndex((item) => item.id === id);
  const nextIndex = index + direction;
  if (index < 0 || nextIndex < 0 || nextIndex >= emojiOptionsEditor.categories.length) {
    return;
  }
  const [current] = emojiOptionsEditor.categories.splice(index, 1);
  emojiOptionsEditor.categories.splice(nextIndex, 0, current);
  syncEmojiOptionsToJson();
}

function resetEmojiCategoriesToPreset() {
  const defaults = defaultEmojiOptionsEditor();
  Object.assign(emojiOptionsEditor, defaults);
  syncEmojiOptionsToJson();
  ElMessage.success('Emoji 分类已恢复为预置方案');
}

function addBrowserApp() {
  const id = uid('app');
  browserApps.value.unshift({
    id,
    name: '新应用',
    icon: '🚀',
    extensions: '',
    action: '',
    description: '',
    background: defaultAppBackground(id),
    type: '自定义',
    platform: 'all',
    createMapping: '无',
    enabled: true,
    openInNewWindow: false,
    maxSize: 100,
    maxSizeUnit: 'MB',
  });
  activeBrowserAppId.value = id;
  syncBrowserAppsToJson();
}

function removeBrowserApp(id: string) {
  browserApps.value = browserApps.value.filter((item) => item.id !== id);
  syncActiveBrowserAppSelection();
  syncBrowserAppsToJson();
}

function moveBrowserApp(id: string, direction: -1 | 1) {
  const index = browserApps.value.findIndex((item) => item.id === id);
  const nextIndex = index + direction;
  if (index < 0 || nextIndex < 0 || nextIndex >= browserApps.value.length) {
    return;
  }
  const [current] = browserApps.value.splice(index, 1);
  browserApps.value.splice(nextIndex, 0, current);
  syncBrowserAppsToJson();
}

function addCustomProperty() {
  const nextIndex = customProperties.value.length;
  const nextSeed =
    !customProperties.value.some((item) => item.type === 'rating')
      ? {
          name: '评级',
          type: 'rating',
          description: '用于标记文件的重要程度与优先级。',
          scope: '文件',
          required: false,
          optionsText: '0',
        }
      : {
          name: '新属性',
          type: 'text',
          description: '',
          scope: '文件 / 文件夹',
          required: false,
          optionsText: '',
        };
  const nextProperty = createCustomPropertyEditor(nextSeed, nextIndex);
  customProperties.value.unshift(nextProperty);
  const id = nextProperty.id;
  activeCustomPropertyId.value = id;
  syncCustomPropertiesToJson();
}

function removeCustomProperty(id: string) {
  customProperties.value = customProperties.value.filter((item) => item.id !== id);
  if (activeCustomPropertyId.value === id) {
    activeCustomPropertyId.value = customProperties.value[0]?.id || '';
  }
  syncCustomPropertiesToJson();
}

function moveCustomProperty(id: string, direction: -1 | 1) {
  const index = customProperties.value.findIndex((item) => item.id === id);
  const nextIndex = index + direction;
  if (index < 0 || nextIndex < 0 || nextIndex >= customProperties.value.length) {
    return;
  }
  const [current] = customProperties.value.splice(index, 1);
  customProperties.value.splice(nextIndex, 0, current);
  syncCustomPropertiesToJson();
}

function applyFileSystemForm(data: FileSystemSettingsPayload) {
  Object.assign(fileSystemForm, clone(data));
  originalFileSystem.value = clone(data);
  iconRules.value = mapIconRulesFromJson(data.file_icon_rules);
  Object.assign(emojiOptionsEditor, parseEmojiOptionsFromJson(data.emoji_options));
  browserApps.value = mapBrowserAppsFromJson(data.browser_apps);
  syncActiveBrowserAppSelection();
  customProperties.value = mapCustomPropertiesFromJson(data.custom_properties);
  if (!customProperties.value.some((item) => item.id === activeCustomPropertyId.value)) {
    activeCustomPropertyId.value = '';
  }
}

async function hydrateBrowserAppsFromService() {
  const groups = await getFileSystemBrowserApps();
  const normalizedApps = mapBrowserAppsFromGroups(groups);
  if (!normalizedApps.length) {
    return;
  }
  browserApps.value = normalizedApps;
  syncActiveBrowserAppSelection();
  syncBrowserAppsToJson();
}

function applySearchForm(data: FullTextSearchSettingsPayload) {
  Object.assign(searchForm, clone(data));
  originalSearch.value = clone(data);
}

function parseJobResultPayload() {
  const result = latestJob.value?.result?.trim();
  if (!result) {
    return {};
  }
  const parsed = safeParse(result, {});
  return typeof parsed === 'object' && parsed ? (parsed as ParsedRecord) : {};
}

function parseJobPayload() {
  const payload = latestJob.value?.payload?.trim();
  if (!payload) {
    return {};
  }
  const parsed = safeParse(payload, {});
  return typeof parsed === 'object' && parsed ? (parsed as ParsedRecord) : {};
}

const latestJobStatusLabel = computed(() => latestJob.value?.status || '暂无任务');

const jobMetrics = computed(() => {
  const result = parseJobResultPayload();
  const payload = parseJobPayload();

  const indexedFiles = Number(result.indexed_files ?? result.indexedFileCount ?? result.file_count ?? payload.indexed_files ?? 0);
  const chunkCount = Number(result.chunk_count ?? result.chunks ?? result.indexed_chunks ?? payload.chunk_count ?? 0);
  const documentCount = Number(result.document_count ?? result.documents ?? result.meili_documents ?? payload.document_count ?? 0);
  const reasons = [
    ...pickStringArray(result, ['skip_reasons', 'skipped_reasons', 'skipReasons']),
    ...pickStringArray(payload, ['skip_reasons', 'skipped_reasons', 'skipReasons']),
  ];

  return {
    indexedFiles: Number.isFinite(indexedFiles) ? indexedFiles : 0,
    chunkCount: Number.isFinite(chunkCount) ? chunkCount : 0,
    documentCount: Number.isFinite(documentCount) ? documentCount : 0,
    skipReasons: Array.from(new Set(reasons)),
    skipReasonCount: Array.from(new Set(reasons)).length,
  };
});

async function loadSearchJobs() {
  try {
    const data = await getQueueJobs({ page: 1, page_size: 10 });
    const searchRelated = data.list?.find((item) => {
      const haystack = `${item.job_type} ${item.resource_type} ${item.resource_id}`.toLowerCase();
      return haystack.includes('search') || haystack.includes('index');
    });
    latestJob.value = searchRelated ?? data.list?.[0] ?? null;

    if (!latestJob.value) {
      searchJobSummary.value = '暂未发现队列任务记录';
      return;
    }

    searchJobSummary.value = `最近任务 #${latestJob.value.id}，状态 ${latestJob.value.status}，类型 ${latestJob.value.job_type}，队列 ${latestJob.value.queue_key}`;
  } catch {
    latestJob.value = null;
    searchJobSummary.value = '任务状态暂时读取失败';
  }
}

async function reloadAll() {
  loading.value = true;
  try {
    const [fileSystem, search] = await Promise.all([getFileSystemSettings(), getFullTextSearchSettings()]);
    applyFileSystemForm(fileSystem);
    await hydrateBrowserAppsFromService();
    applySearchForm(search);
    await loadSearchJobs();
    showConnectionAlert('success', '后端连接正常', `已成功连接 ${backendEndpoint}，当前页签数据已刷新。`);
  } catch (error) {
    const message = error instanceof Error ? error.message : '加载文件系统设置失败';
    showConnectionAlert('error', '后端服务当前不可用', `${message} 当前检测地址：${backendEndpoint}`);
    ElMessage.error(message);
  } finally {
    loading.value = false;
  }
}

async function saveFileSystem() {
  const data = await updateFileSystemSettings(clone(fileSystemForm));
  applyFileSystemForm(data);
  await hydrateBrowserAppsFromService();
  showConnectionAlert('success', '后端连接正常', `配置已成功保存到 ${backendEndpoint}。`);
}

async function saveFileSystemIcons() {
  const data = await updateFileSystemIconSettings({
    file_icon_rules: fileSystemForm.file_icon_rules,
    emoji_options: fileSystemForm.emoji_options,
  });
  fileSystemForm.file_icon_rules = data.file_icon_rules;
  fileSystemForm.emoji_options = data.emoji_options;
  originalFileSystem.value = {
    ...originalFileSystem.value,
    file_icon_rules: data.file_icon_rules,
    emoji_options: data.emoji_options,
  };
  iconRules.value = mapIconRulesFromJson(data.file_icon_rules);
  Object.assign(emojiOptionsEditor, parseEmojiOptionsFromJson(data.emoji_options));
  await refreshFileListAfterIconSettingsSave();
  showConnectionAlert('success', '后端连接正常', `文件图标配置已成功保存到 ${backendEndpoint}。`);
}

async function refreshFileListAfterIconSettingsSave() {
  try {
    await fileStore.refresh();
  } catch (error) {
    console.warn('Failed to refresh file list after icon settings save:', error);
  }
}

async function saveSearch() {
  const data = await updateFullTextSearchSettings(clone(searchForm));
  applySearchForm(data);
}

async function saveCurrentTab() {
  saving.value = true;
  try {
    if (activeTab.value === 'search') {
      await saveSearch();
    } else if (activeTab.value === 'icons') {
      await saveFileSystemIcons();
    } else {
      await saveFileSystem();
    }
    ElMessage.success(`${activeTabMeta.value.label}已保存`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : `${activeTabMeta.value.label}\u4fdd\u5b58\u5931\u8d25`);
  } finally {
    saving.value = false;
  }
}

function resetCurrentTab() {
  if (activeTab.value === 'search') {
    applySearchForm(originalSearch.value);
  } else {
    applyFileSystemForm(originalFileSystem.value);
  }
  ElMessage.success(`${activeTabMeta.value.label}已恢复到最近一次加载状态`);
}

async function handleRebuildIndex() {
  rebuilding.value = true;
  try {
    const result = await rebuildFullTextSearchIndex();
    ElMessage.success(`已提交重建索引任务 #${result.job_id}`);
    await loadSearchJobs();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重建索引失败');
  } finally {
    rebuilding.value = false;
  }
}

async function handleClearBlobUrlCache() {
  clearingBlobUrlCache.value = true;
  try {
    await clearFileSystemBlobUrlCache();
    ElMessage.success('Blob URL 缓存已清除');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '清除 Blob URL 缓存失败');
  } finally {
    clearingBlobUrlCache.value = false;
  }
}

async function switchTab(tab: TabKey) {
  activeTab.value = tab;
  await router.replace({
    query: {
      ...route.query,
      tab,
    },
  });
}

onMounted(async () => {
  const queryTab = route.query.tab;
  if (typeof queryTab === 'string' && tabs.some((tab) => tab.key === queryTab)) {
    activeTab.value = queryTab as TabKey;
  }
  await reloadAll();
});
</script>

<style scoped>
.file-system-page {
  display: grid;
  gap: 20px;
  min-height: calc(100vh - 96px);
  color: #172033;
}

.hero-card,
.panel-card,
.action-bar,
.overview-card,
.group-card,
.field-card,
.toggle-card,
.metric-card,
.job-hero,
.job-stats-card,
.editor-card,
.empty-card {
  border: 1px solid rgba(220, 231, 245, 0.9);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 20px 44px rgba(79, 102, 145, 0.12);
}

.hero-card {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.95fr);
  gap: 20px;
  padding: 30px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top left, rgba(56, 189, 248, 0.18), transparent 36%),
    radial-gradient(circle at bottom right, rgba(250, 204, 21, 0.16), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.96));
}

.eyebrow {
  margin: 0 0 12px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-copy h1 {
  margin: 0;
  font-size: 42px;
  line-height: 1.05;
}

.hero-text {
  margin: 16px 0 0;
  max-width: 760px;
  color: #52607a;
  font-size: 15px;
  line-height: 1.85;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 18px;
}

.hero-tag {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, #1d4ed8, #0ea5e9);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}

.hero-tag-soft {
  border: 1px solid rgba(147, 197, 253, 0.9);
  background: rgba(255, 255, 255, 0.82);
  color: #2563eb;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.metric-card,
.overview-card,
.editor-card,
.empty-card,
.job-stats-grid div {
  display: grid;
  gap: 8px;
  padding: 18px;
  border-radius: 20px;
}

.metric-card strong,
.overview-card strong,
.job-stats-grid strong {
  font-size: 22px;
}

.accent-blue {
  background: linear-gradient(180deg, rgba(219, 234, 254, 0.88), rgba(255, 255, 255, 0.94));
}

.accent-gold {
  background: linear-gradient(180deg, rgba(254, 240, 138, 0.72), rgba(255, 255, 255, 0.94));
}

.accent-green {
  background: linear-gradient(180deg, rgba(187, 247, 208, 0.72), rgba(255, 255, 255, 0.94));
}

.accent-cyan {
  background: linear-gradient(180deg, rgba(165, 243, 252, 0.72), rgba(255, 255, 255, 0.94));
}

.connection-alert {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  border: 1px solid rgba(220, 231, 245, 0.9);
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.98));
  box-shadow: 0 18px 36px rgba(79, 102, 145, 0.12);
}

.connection-alert.is-error {
  border-color: rgba(248, 113, 113, 0.22);
  background: linear-gradient(135deg, rgba(254, 242, 242, 0.98), rgba(255, 255, 255, 0.96));
}

.connection-alert.is-success {
  border-color: rgba(34, 197, 94, 0.22);
  background: linear-gradient(135deg, rgba(236, 253, 245, 0.98), rgba(255, 255, 255, 0.96));
}

.connection-alert-icon {
  display: grid;
  place-items: center;
  width: 42px;
  height: 42px;
  border-radius: 14px;
  color: #fff;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.14);
}

.connection-alert.is-error .connection-alert-icon {
  background: linear-gradient(135deg, #ef4444, #fb7185);
}

.connection-alert.is-success .connection-alert-icon {
  background: linear-gradient(135deg, #16a34a, #34d399);
}

.connection-alert-icon svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.connection-alert-copy {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.connection-alert-copy strong {
  color: #132238;
  font-size: 15px;
  font-weight: 800;
}

.connection-alert-copy span {
  color: #6b7c93;
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;
}

.connection-alert-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.connection-alert-button {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 12px;
}

.panel-card {
  overflow: hidden;
  border-radius: 28px;
}

.tabs-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 18px 18px 0;
}

.tab-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 46px;
  padding: 0 18px;
  border: 1px solid transparent;
  border-radius: 16px;
  background: linear-gradient(180deg, #f8fbff, #edf4fb);
  color: #3b4b63;
  font-size: 14px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.tab-button:hover,
.primary-button:hover,
.ghost-button:hover,
.secondary-button:hover {
  transform: translateY(-1px);
}

.tab-button.active {
  border-color: rgba(59, 130, 246, 0.25);
  background: linear-gradient(135deg, #1d4ed8, #38bdf8);
  color: #fff;
  box-shadow: 0 12px 26px rgba(37, 99, 235, 0.18);
}

.tab-icon {
  font-size: 16px;
}

.panel-body {
  display: grid;
  gap: 18px;
  padding: 22px 18px 24px;
  background: #f8fafc;
}

.section-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 18px;
}

.section-header h2,
.group-head h3,
.job-copy h3 {
  margin: 0;
  color: #132238;
}

.section-header h2 {
  font-size: 30px;
}

.section-header p,
.group-head p,
.job-copy p,
.field-card span,
.metric-card span,
.metric-card small,
.overview-card span,
.overview-card small,
.editor-title span,
.action-copy span,
.reason-title {
  color: #64748b;
}

.settings-header {
  padding: 10px 10px 4px;
}

.settings-header h2 {
  font-size: 32px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.settings-header p {
  max-width: 820px;
  color: #7c8aa5;
  line-height: 1.8;
}

.settings-heading {
  display: grid;
  gap: 10px;
}

.settings-kicker {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border: 1px solid rgba(96, 165, 250, 0.22);
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(239, 246, 255, 0.95));
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.settings-header h2 {
  font-size: 0;
}

.settings-header h2::before {
  content: '参数设置';
  font-size: 34px;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: #0f172a;
}

.settings-header p {
  font-size: 0;
}

.settings-header p::before {
  content: '统一管理编辑、回收、列表、缓存、传输与安全策略。';
  font-size: 14px;
  line-height: 1.9;
  color: #7c8aa5;
}

.settings-metrics-grid {
  gap: 14px;
  margin-bottom: 2px;
}

.settings-metric-card {
  padding: 20px 22px;
  border: 1px solid #e6edf6;
  border-radius: 22px;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.12), transparent 34%),
    linear-gradient(180deg, #ffffff, #f8fbff);
  box-shadow: 0 14px 32px rgba(148, 163, 184, 0.13);
}

.settings-metric-card {
  position: relative;
  overflow: hidden;
}

.settings-metric-card::after {
  content: '';
  position: absolute;
  inset: auto -20px -36px auto;
  width: 120px;
  height: 120px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.settings-stack {
  grid-template-columns: 1fr;
  gap: 20px;
  max-width: 1240px;
}

.settings-stack .group-card {
  gap: 0;
  padding: 0;
  overflow: hidden;
  border: 1px solid #e7edf5;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(252, 253, 255, 0.99));
  box-shadow:
    0 18px 38px rgba(148, 163, 184, 0.14),
    0 1px 0 rgba(255, 255, 255, 0.65) inset;
}

.settings-section-head {
  padding: 24px 26px 18px;
  border-bottom: 1px solid #eef3f8;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.06), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
}

.settings-section-head h3 {
  font-size: 0;
}

.settings-section-head p {
  margin-top: 6px;
  color: #7b879c;
  line-height: 1.75;
  font-size: 0;
}

.settings-section-head-inline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.settings-form-list {
  gap: 0;
  padding: 0 26px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(249, 251, 255, 0.9));
}

.settings-form-list.two-column {
  grid-template-columns: 1fr;
}

.settings-form-list.single-column {
  grid-template-columns: 1fr;
}

.settings-text-stack {
  grid-template-columns: 1fr;
  gap: 18px;
  max-width: 1240px;
}

.settings-text-stack > .field-card {
  padding: 0;
  overflow: hidden;
  border: 1px solid #e7edf5;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.14);
}

.settings-text-stack > .field-card,
.settings-text-stack .nested-list {
  grid-column: 1 / -1;
}

.settings-text-stack > .field-card > span,
.settings-text-stack .nested-list .field-card > span {
  font-size: 15px;
  font-weight: 700;
  color: #172033;
}

.settings-text-stack > .field-card {
  gap: 12px;
  padding: 24px 26px;
}

.nested-list {
  gap: 0;
  padding: 0;
}

.nested-list .field-card {
  margin: 0;
}

.settings-stack .field-card,
.settings-stack .toggle-card,
.settings-text-stack .field-card {
  gap: 10px;
  margin: 0;
  padding: 20px 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.settings-stack .field-card {
  display: grid;
  grid-template-columns: minmax(220px, 260px) minmax(0, 1fr);
  column-gap: 28px;
  row-gap: 10px;
  align-items: start;
}

.settings-stack .field-card > span {
  grid-column: 1;
  grid-row: 1;
  padding-top: 10px;
}

.settings-stack .field-card > .field-input,
.settings-stack .field-card > .field-select,
.settings-stack .field-card > .field-textarea,
.settings-stack .field-card > .split-row {
  grid-column: 2;
  grid-row: 1;
}

.settings-stack .field-card::after {
  grid-column: 2;
  grid-row: 2;
  margin-top: -2px;
}

.settings-stack .toggle-card {
  grid-template-columns: minmax(220px, 260px) minmax(0, 1fr);
  column-gap: 28px;
  row-gap: 8px;
  align-items: start;
}

.settings-stack .toggle-card > input[type='checkbox'] {
  grid-column: 2;
  grid-row: 1;
  justify-self: start;
  margin-top: 12px;
}

.settings-stack .toggle-card > div {
  grid-column: 2;
  grid-row: 1;
  display: grid;
  gap: 6px;
  padding-left: 34px;
}

.settings-stack .toggle-card::before {
  content: '';
  grid-column: 1;
  grid-row: 1;
}

.settings-stack .field-card:not(.span-2),
.settings-stack .toggle-card,
.nested-list .field-card,
.settings-text-stack > .field-card {
  border-bottom: 1px solid #eef3f8;
}

.settings-stack .form-grid > :nth-last-child(-n + 2),
.settings-text-stack .nested-list > :nth-last-child(-n + 2) {
  border-bottom: 0;
}

.settings-text-stack > .field-card:last-child,
.settings-stack .group-card > .inline-actions:last-child {
  border-bottom: 0;
}

.settings-stack .field-card span,
.settings-stack .toggle-card strong,
.settings-text-stack .field-card span {
  color: #172033;
  font-size: 0;
  font-weight: 700;
}

.settings-stack .toggle-card small,
.settings-text-stack .field-card small {
  color: #8a97ad;
  font-size: 0;
  line-height: 1.7;
}

.settings-stack .field-input,
.settings-stack .field-select,
.settings-stack .field-textarea,
.settings-text-stack .field-input,
.settings-text-stack .field-select,
.settings-text-stack .field-textarea {
  min-height: 44px;
  padding: 12px 14px;
  border: 1px solid #d7e2ee;
  border-radius: 14px;
  background: #fbfdff;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.settings-stack .field-input:focus,
.settings-stack .field-select:focus,
.settings-stack .field-textarea:focus,
.settings-text-stack .field-input:focus,
.settings-text-stack .field-select:focus,
.settings-text-stack .field-textarea:focus {
  border-color: #60a5fa;
  box-shadow:
    0 0 0 4px rgba(96, 165, 250, 0.18),
    0 10px 24px rgba(59, 130, 246, 0.08);
}

.settings-stack .split-row {
  grid-template-columns: minmax(0, 1fr) 132px;
}

.settings-stack .inline-actions {
  padding: 22px 26px 26px;
  justify-content: flex-start;
}

.settings-stack .ghost-button {
  min-height: 48px;
  padding: 0 20px;
  border-color: #d5e3f5;
  border-radius: 16px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.1);
}

.settings-form-list > .field-card,
.settings-form-list > .toggle-card,
.settings-text-stack > .field-card {
  position: relative;
}

.settings-form-list > .field-card::after,
.settings-text-stack > .field-card::after {
  display: block;
  margin-top: 2px;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.75;
}

.settings-form-list > .field-card > span::before,
.settings-text-stack > .field-card > span::before,
.settings-form-list > .toggle-card strong::before {
  font-size: 15px;
  font-weight: 700;
  color: #172033;
}

.settings-form-list > .toggle-card small::before,
.settings-text-stack > .field-card small::before {
  font-size: 12px;
  line-height: 1.75;
  color: #94a3b8;
}

.settings-stack .toggle-card {
  grid-template-columns: 22px minmax(0, 1fr);
  align-items: start;
}

.settings-stack .toggle-card input[type='checkbox'] {
  width: 18px;
  height: 18px;
  margin-top: 4px;
  accent-color: #2563eb;
}

.settings-text-stack {
  gap: 20px;
}

.settings-text-stack > .field-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 252, 0.96));
}

.settings-side-body {
  gap: 20px;
  padding: 22px 26px 14px;
}

.settings-side-copy p {
  font-size: 0;
}

.settings-side-copy p::before {
  content: '适合高频预览、下载与外链场景，减少重复生成签名结果。';
  font-size: 13px;
  line-height: 1.85;
  color: #7c8aa5;
}

.settings-side-metrics {
  grid-template-columns: repeat(2, minmax(0, 220px));
}

.panel-body > section.settings-stack {
  grid-template-columns: 1fr;
}

.panel-body > section.settings-stack:last-of-type {
  grid-template-columns: 1fr;
}

.panel-body > section.settings-stack:last-of-type > .group-card:last-child {
  align-content: start;
}

.settings-stack .group-card,
.settings-text-stack > .field-card {
  max-width: 1240px;
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .settings-section-head h3::before {
  content: '编辑与回收';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .settings-section-head p::before {
  content: '控制在线编辑、回收站扫描和 Blob 回收节奏。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .settings-section-head h3::before {
  content: '列表与地图';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .settings-section-head p::before {
  content: '优化目录浏览、分页加载、批量操作和地图展示。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .settings-section-head h3::before {
  content: '文件加密';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .settings-section-head p::before {
  content: '配置主密钥存储方式，并决定是否展示加密状态。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .settings-section-head h3::before {
  content: '文件事件推送';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .settings-section-head p::before {
  content: '调节离线补偿、事件合并与多端同步体验。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .settings-section-head h3::before {
  content: '高级设置';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .settings-section-head p::before {
  content: '上传、下载、签名、统计、重试与并行传输等底层参数。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(2) .settings-section-head h3::before {
  content: 'Blob URL 缓存';
  font-size: 20px;
  font-weight: 800;
  color: #132238;
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(2) .settings-section-head p::before {
  content: '将有效期与复用窗口集中展示，让缓存策略更直观。';
  font-size: 14px;
  color: #7b879c;
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(1) > span::before {
  content: '在线编辑最大文件';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(1)::after {
  content: '限制浏览器内直接编辑的文件大小，避免拖慢前端体验。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(2) > span::before {
  content: '回收站扫描间隔';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(2)::after {
  content: '后台按该周期巡检回收站并触发清理任务。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(3) > span::before {
  content: 'Blob 回收间隔';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(3)::after {
  content: '控制临时 Blob 数据的回收频率，减少对象存储残留。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(4) > span::before {
  content: '静态缓存有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(1) .field-card:nth-child(4)::after {
  content: '影响静态资源缓存时长，兼顾刷新速度与命中率。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(1) > span::before {
  content: '分页模式';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(1)::after {
  content: '建议大目录优先使用 cursor，以提升连续翻页性能。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(2) > span::before {
  content: '最大单页数量';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(2)::after {
  content: '限制每次读取的列表数量，避免目录一次返回过多数据。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(3) > span::before {
  content: '最大批量操作数';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(3)::after {
  content: '用于删除、移动、复制等批量动作的单次处理上限。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(4) > span::before {
  content: '最大递归搜索深度';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(4)::after {
  content: '限制深层目录检索深度，平衡搜索能力与后端负载。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(5) > span::before {
  content: '地图提供方';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(5)::after {
  content: '影响照片地图视图底图来源与交互能力。';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(6) > span::before {
  content: '目录统计缓存（秒）';
}

.panel-body > section.settings-stack:nth-of-type(3) > .group-card:nth-child(2) .field-card:nth-child(6)::after {
  content: '缓存文件夹大小与数量统计，减少重复聚合查询。';
}

.panel-body > section.settings-text-stack .field-card:nth-child(1) > span::before {
  content: 'MIME 映射';
}

.panel-body > section.settings-text-stack .field-card:nth-child(1)::after {
  content: '维护扩展名与 MIME 的映射关系，供预览与识别逻辑复用。';
}

.panel-body > section.settings-text-stack .field-card:nth-child(2) > span::before {
  content: '图片分类查询';
}

.panel-body > section.settings-text-stack .field-card:nth-child(2)::after {
  content: '用于识别图片类型文件，决定图库等界面的内容聚合。';
}

.panel-body > section.settings-text-stack .field-card:nth-child(3) > span::before {
  content: '视频分类查询';
}

.panel-body > section.settings-text-stack .field-card:nth-child(3)::after {
  content: '用于视频类资源的分类归档与筛选展示。';
}

.panel-body > section.settings-text-stack .field-card:nth-child(4) > span::before {
  content: '音频分类查询';
}

.panel-body > section.settings-text-stack .field-card:nth-child(4)::after {
  content: '用于识别音频资源，支撑音乐与语音文件聚合。';
}

.panel-body > section.settings-text-stack .field-card:nth-child(5) > span::before {
  content: '文档分类查询';
}

.panel-body > section.settings-text-stack .field-card:nth-child(5)::after {
  content: '用于 Office、PDF、文本等文档类文件的检索分组。';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .field-card > span::before {
  content: '主加密密钥存储方式';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .field-card::after {
  content: '切换密钥持久化位置时，需要与部署和备份策略保持一致。';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .toggle-card strong::before {
  content: '显示加密状态';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(1) .toggle-card small::before {
  content: '开启后，文件与文件夹详情会展示当前加密状态。';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .toggle-card strong::before {
  content: '启用文件事件推送';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .toggle-card small::before {
  content: '开启后，客户端可以接收文件变更事件，提升刷新与同步实时性。';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .field-card:nth-child(2) > span::before {
  content: '离线有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .field-card:nth-child(2)::after {
  content: '定义离线缓存与补偿事件可继续生效的时间窗口。';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .field-card:nth-child(3) > span::before {
  content: '防抖延时（秒）';
}

.panel-body > section.settings-stack:nth-of-type(5) > .group-card:nth-child(2) .field-card:nth-child(3)::after {
  content: '短时间内合并重复事件，减少前端收到的无效刷新。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(1) > span::before {
  content: '服务端打包下载会话有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(1)::after {
  content: '控制服务端打包下载任务会话保留时间。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(2) > span::before {
  content: '上传会话有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(2)::after {
  content: '断点上传会话在服务端的保留时间。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(3) > span::before {
  content: '从机 API 签名有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(3)::after {
  content: '限制从机接口签名可用时间，降低重放风险。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(4) > span::before {
  content: '目录统计信息有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(4)::after {
  content: '影响目录大小与文件数统计的更新频率。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(5) > span::before {
  content: '分片错误最大重试';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(5)::after {
  content: '上传失败后单个分片允许自动重试的最大次数。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .toggle-card strong::before {
  content: '缓存流式分片用于重试';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .toggle-card small::before {
  content: '开启后，失败重试可复用已缓存分片，减少重复读取与上传时间。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(7) > span::before {
  content: '中转最大并行传输';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(7)::after {
  content: '并行度越高速度越快，但也会增加客户端与服务端压力。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(8) > span::before {
  content: 'OAuth 存储策略凭证刷新间隔';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(8)::after {
  content: '用于周期性刷新第三方存储凭证，避免服务中断。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(9) > span::before {
  content: 'WOPI 会话有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(9)::after {
  content: '控制在线文档协同会话的授权持续时间。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(10) > span::before {
  content: '文件 Blob 临时 URL 有效期（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(10)::after {
  content: '决定预签名 URL 的有效时间，影响预览与下载链接可用期。';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(11) > span::before {
  content: '文件 Blob 临时 URL 复用窗口（秒）';
}

.panel-body > section.settings-stack:nth-of-type(6) > .group-card:nth-child(1) .field-card:nth-child(11)::after {
  content: '在复用窗口内复用已有 URL，减少重复签名带来的后端开销。';
}

.panel-body > section.settings-stack:last-of-type {
  grid-template-columns: minmax(0, 1.8fr) minmax(320px, 0.92fr);
  align-items: stretch;
}

.panel-body > section.settings-stack:last-of-type > .group-card:last-child {
  min-height: 100%;
}

.panel-body > section.settings-stack:last-of-type .settings-section-head-inline {
  display: grid;
  gap: 10px;
  align-items: start;
  justify-content: start;
}

.settings-side-body {
  display: grid;
  gap: 18px;
  padding: 22px 26px 8px;
}

.settings-side-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.settings-side-metric {
  display: grid;
  gap: 6px;
  padding: 16px 16px 14px;
  border: 1px solid #e7edf7;
  border-radius: 18px;
  background: linear-gradient(180deg, #ffffff, #f6f9ff);
  box-shadow: 0 10px 22px rgba(148, 163, 184, 0.08);
}

.settings-side-metric span {
  color: #7a8aa3;
  font-size: 12px;
  font-weight: 700;
}

.settings-side-metric strong {
  color: #132238;
  font-size: 22px;
  line-height: 1;
}

.settings-side-copy {
  padding: 2px 2px 0;
}

.settings-side-copy p {
  margin: 0;
  color: #7c8aa5;
  font-size: 13px;
  line-height: 1.85;
}

.settings-side-actions {
  margin-top: auto;
  padding-top: 8px;
}

.search-header {
  padding: 8px 6px 4px;
}

.search-heading {
  display: grid;
  gap: 12px;
}

.search-heading::before {
  content: 'Full Text Search Center';
  display: inline-flex;
  width: fit-content;
  min-height: 30px;
  align-items: center;
  padding: 0 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(168, 85, 247, 0.1));
  border: 1px solid rgba(99, 102, 241, 0.18);
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.search-header h2 {
  font-size: 0;
}

.search-header h2::before {
  content: '全文搜索';
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: #0f172a;
}

.search-header p {
  max-width: 860px;
  font-size: 0;
}

.search-header p::before {
  content: '连接 Meilisearch 与 Tika，把搜索能力、内容提取和重建任务集中展示。';
  font-size: 14px;
  line-height: 1.9;
  color: #74839d;
}

.search-metrics-grid {
  gap: 16px;
}

.search-metric-card {
  position: relative;
  overflow: hidden;
  min-height: 152px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 24px;
  box-shadow: 0 18px 40px rgba(148, 163, 184, 0.12);
}

.search-metric-card::after {
  content: '';
  position: absolute;
  right: -18px;
  top: -24px;
  width: 130px;
  height: 130px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.65), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.search-metric-card strong {
  font-size: 30px;
  letter-spacing: -0.03em;
}

.search-metric-card-blue {
  background: linear-gradient(145deg, #eef6ff, #ffffff 58%, #f0f9ff);
}

.search-metric-card-violet {
  background: linear-gradient(145deg, #f5f3ff, #ffffff 58%, #eef2ff);
}

.search-metric-card-emerald {
  background: linear-gradient(145deg, #ecfdf5, #ffffff 58%, #f0fdfa);
}

.search-metric-card-gold {
  background: linear-gradient(145deg, #fff7ed, #ffffff 58%, #fefce8);
}

.search-config-layout {
  gap: 20px;
}

.search-config-card {
  overflow: hidden;
  border: 1px solid #e8eef7;
  border-radius: 26px;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.06), transparent 22%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 252, 0.96));
  box-shadow:
    0 22px 46px rgba(148, 163, 184, 0.12),
    0 1px 0 rgba(255, 255, 255, 0.8) inset;
}

.search-card-head {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 16px;
  padding: 26px 28px 20px;
  border-bottom: 1px solid #eef3f8;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(246, 249, 253, 0.95));
}

.search-card-head h3 {
  margin: 0;
  color: #132238;
  font-size: 24px;
  font-weight: 800;
}

.search-card-head p {
  margin: 8px 0 0;
  color: #7b879c;
  line-height: 1.8;
}

.search-card-head::after {
  content: '高可用搜索架构';
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  padding: 0 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(14, 165, 233, 0.12));
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  white-space: nowrap;
}

.search-form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
  padding: 8px 28px 10px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(249, 251, 255, 0.92));
}

.search-toggle-card,
.search-field-card {
  margin: 0;
  padding: 24px 0;
  border: 0;
  border-bottom: 1px solid #eef3f8;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.search-form-grid > :last-child {
  border-bottom: 0;
}

.search-toggle-card {
  grid-template-columns: minmax(0, 1fr) 56px;
  align-items: center;
  gap: 0 20px;
  padding: 24px 0;
}

.search-toggle-card input[type='checkbox'] {
  appearance: none;
  -webkit-appearance: none;
  position: relative;
  width: 56px;
  height: 32px;
  margin: 0;
  border: 1px solid #d9e3ef;
  border-radius: 999px;
  background: linear-gradient(180deg, #eef2f7, #e2e8f0);
  box-shadow:
    inset 0 1px 2px rgba(15, 23, 42, 0.08),
    0 6px 16px rgba(148, 163, 184, 0.12);
  cursor: pointer;
  justify-self: end;
  transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.search-toggle-card input[type='checkbox']::before {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow:
    0 4px 10px rgba(15, 23, 42, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  transition: transform 0.2s ease;
}

.search-toggle-card input[type='checkbox']:checked {
  border-color: rgba(37, 99, 235, 0.22);
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  box-shadow:
    inset 0 1px 1px rgba(255, 255, 255, 0.18),
    0 10px 20px rgba(37, 99, 235, 0.22);
}

.search-toggle-card input[type='checkbox']:checked::before {
  transform: translateX(24px);
}

.search-toggle-card input[type='checkbox']:focus {
  outline: none;
  box-shadow:
    0 0 0 4px rgba(96, 165, 250, 0.16),
    0 10px 20px rgba(37, 99, 235, 0.16);
}

.search-toggle-copy {
  display: grid;
  gap: 6px;
  align-content: center;
  min-width: 0;
}

.search-toggle-card strong {
  color: #122033;
  font-size: 18px;
  line-height: 1.2;
}

.search-toggle-card small {
  color: #8a97ad;
  font-size: 13px;
  line-height: 1.75;
  max-width: 760px;
}

.search-field-card {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  align-items: start;
}

.search-field-card span {
  color: #132238;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.35;
}

.search-field-card .field-input,
.search-field-card .field-select,
.search-field-card .field-textarea {
  width: 100%;
  min-height: 50px;
  border: 1px solid #d8e3ef;
  border-radius: 16px;
  background: #fbfdff;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.search-field-card .field-input:focus,
.search-field-card .field-select:focus,
.search-field-card .field-textarea:focus {
  box-shadow:
    0 0 0 4px rgba(96, 165, 250, 0.16),
    0 12px 24px rgba(59, 130, 246, 0.08);
}

.search-field-card::after {
  display: block;
  margin-top: 2px;
  color: #93a0b5;
  font-size: 12px;
  line-height: 1.8;
}

.search-field-card:nth-child(3)::after {
  content: '搜索引擎服务地址，用于写入配置并驱动真实索引构建。';
}

.search-field-card:nth-child(4)::after {
  content: '服务端访问 Meilisearch 的授权密钥，建议使用受限 API Key。';
}

.search-field-card:nth-child(5)::after {
  content: 'Tika 提取服务端点，负责从真实文件中抽取可检索文本。';
}

.search-field-card:nth-child(6)::after {
  content: '控制每次搜索返回的结果规模，平衡信息密度与响应速度。';
}

.search-field-card:nth-child(7)::after {
  content: '限制可进入索引流程的文件体积，避免超大文件拖慢搜索集群。';
}

.search-field-card:nth-child(8)::after {
  content: '切块越细搜索精度越高，切块越大索引体积越小。';
}

.search-job-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 18px;
}

.search-job-hero {
  padding: 26px 28px;
  border: 1px solid rgba(59, 130, 246, 0.12);
  border-radius: 28px;
  background:
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.16), transparent 30%),
    radial-gradient(circle at bottom right, rgba(16, 185, 129, 0.14), transparent 28%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.99), rgba(241, 247, 255, 0.98));
  box-shadow: 0 20px 42px rgba(148, 163, 184, 0.12);
}

.search-job-hero .job-copy h3 {
  font-size: 28px;
}

.search-job-hero .job-copy p {
  max-width: 780px;
  line-height: 1.85;
}

.search-job-hero .job-pill {
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(14, 165, 233, 0.14));
}

.search-cta-button {
  min-width: 188px;
  min-height: 54px;
  border-radius: 18px;
  background: linear-gradient(135deg, #2563eb, #0ea5e9 48%, #10b981);
  box-shadow: 0 18px 32px rgba(37, 99, 235, 0.24);
}

.search-job-stats-card {
  border: 1px solid #e7edf5;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.12);
}

.search-job-stats-grid div {
  border: 1px solid #e8eef7;
  background: linear-gradient(180deg, #ffffff, #f6f9ff);
}

.search-reason-list {
  padding-top: 4px;
}

.search-text-panels {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.search-note-card {
  border: 1px solid #e7edf5;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 252, 0.96));
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.12);
}

.search-note-card span {
  color: #132238;
  font-size: 15px;
  font-weight: 700;
}

.search-note-card small {
  color: #8b97aa;
  font-size: 12px;
  line-height: 1.8;
}

.search-note-card .field-textarea {
  width: 100%;
  min-height: 132px;
  border-radius: 16px;
  background: #fbfdff;
}

.overview-grid,
.group-grid,
.form-grid,
.toggle-grid,
.text-panels,
.editor-grid,
.job-stats-grid {
  display: grid;
  gap: 16px;
}

.overview-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.group-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.group-grid.single {
  grid-template-columns: 1fr;
}

.group-card,
.job-hero,
.job-stats-card {
  display: grid;
  gap: 16px;
  padding: 20px;
  border-radius: 22px;
}

.form-grid,
.text-panels {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.toggle-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.field-card,
.toggle-card {
  display: grid;
  gap: 10px;
  padding: 18px;
  border-radius: 18px;
}

.field-card.span-2 {
  grid-column: 1 / -1;
}

.checkbox-field {
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
}

.field-input,
.field-select,
.field-textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d9e3f0;
  border-radius: 14px;
  background: #fff;
  color: #172033;
  font-size: 14px;
}

.field-input:focus,
.field-select:focus,
.field-textarea:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.14);
}

.field-textarea {
  resize: vertical;
  min-height: 110px;
}

.field-textarea.giant {
  min-height: 280px;
  font-family: Consolas, Monaco, monospace;
}

.split-row {
  display: grid;
  grid-template-columns: 1fr 120px;
  gap: 10px;
}

.toggle-card {
  grid-template-columns: 20px minmax(0, 1fr);
  align-items: start;
}

.toggle-card strong {
  color: #132238;
}

.job-hero {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  background:
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.12), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(243, 248, 255, 0.98));
}

.job-pill,
.property-badge {
  display: inline-flex;
  width: fit-content;
  min-height: 30px;
  align-items: center;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(29, 78, 216, 0.1);
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 800;
}

.job-stats-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.reason-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.reason-chip {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 12px;
  border-radius: 999px;
  background: #fff7ed;
  color: #c2410c;
  font-size: 13px;
  font-weight: 700;
}

.icon-section-header {
  align-items: end;
  padding: 10px 4px 2px;
}

.icon-section-header > div {
  display: grid;
  gap: 10px;
}

.icon-section-header h2 {
  margin: 0;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: #0f172a;
}

.icon-section-header p {
  max-width: 880px;
  margin: 0;
  color: #74839d;
  line-height: 1.85;
}

.icon-section-header::before {
  content: 'ICON RULE STUDIO';
  display: inline-flex;
  width: fit-content;
  min-height: 30px;
  align-items: center;
  padding: 0 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(16, 185, 129, 0.12));
  border: 1px solid rgba(59, 130, 246, 0.18);
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.icon-section-header .secondary-button {
  min-width: 196px;
  min-height: 52px;
  border-radius: 18px;
  border-color: rgba(37, 99, 235, 0.18);
  background: linear-gradient(135deg, #2563eb, #0ea5e9 52%, #34d399);
  color: #fff;
  box-shadow: 0 18px 34px rgba(37, 99, 235, 0.22);
}

.icon-section-header .secondary-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 22px 38px rgba(37, 99, 235, 0.24);
}

.icon-rule-list {
  grid-template-columns: 1fr;
  gap: 20px;
}

.icon-overview-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.icon-overview-card {
  position: relative;
  overflow: hidden;
  min-height: 144px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 24px;
  box-shadow: 0 18px 40px rgba(148, 163, 184, 0.12);
}

.icon-overview-card::after {
  content: '';
  position: absolute;
  right: -18px;
  top: -24px;
  width: 120px;
  height: 120px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0));
}

.icon-overview-card strong {
  font-size: 28px;
  letter-spacing: -0.03em;
}

.icon-overview-card-blue {
  background: linear-gradient(145deg, #eef6ff, #ffffff 58%, #f0f9ff);
}

.icon-overview-card-violet {
  background: linear-gradient(145deg, #f5f3ff, #ffffff 58%, #eef2ff);
}

.icon-overview-card-gold {
  background: linear-gradient(145deg, #fff7ed, #ffffff 58%, #fefce8);
}

.icon-rule-card {
  gap: 0;
  overflow: hidden;
  border: 1px solid #e6edf8;
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 12%, transparent), transparent 20%),
    radial-gradient(circle at bottom left, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 8%, transparent), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(247, 250, 255, 0.96));
  box-shadow:
    0 20px 42px rgba(148, 163, 184, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.icon-rule-head {
  position: relative;
  padding: 24px 26px 20px;
  border-bottom: 1px solid #edf2f8;
  background:
    radial-gradient(circle at left top, rgba(14, 165, 233, 0.12), transparent 22%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(244, 248, 255, 0.92));
}

.icon-rule-head::after {
  content: '';
  position: absolute;
  right: 18px;
  top: 16px;
  width: 88px;
  height: 88px;
  border-radius: 999px;
  background: radial-gradient(circle, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 18%, rgba(255, 255, 255, 0.75)), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.icon-rule-head .preview-icon.large {
  width: 72px;
  height: 72px;
  border-radius: 22px;
  font-size: 34px;
  color: #fff;
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 82%, white), var(--icon-rule-tint, #2563eb) 58%, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 70%, #0f172a));
  box-shadow:
    0 16px 30px rgba(37, 99, 235, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.icon-rule-title strong {
  color: #102038;
  font-size: 22px;
  line-height: 1.2;
}

.icon-rule-title span {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(215, 226, 241, 0.95);
  color: #7d8aa1;
  line-height: 1.7;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.icon-rule-title::after {
  content: '可视化图标规则';
  display: inline-flex;
  width: fit-content;
  align-items: center;
  min-height: 28px;
  margin-top: 6px;
  padding: 0 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--icon-rule-tint, #2563eb) 12%, white);
  color: color-mix(in srgb, var(--icon-rule-tint, #2563eb) 82%, #1e293b);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.03em;
}

.icon-rule-head .danger-text {
  position: relative;
  z-index: 1;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 999px;
  border: 1px solid rgba(239, 68, 68, 0.16);
  background: linear-gradient(180deg, #fff5f5, #fff1f2);
  color: #dc2626;
  box-shadow: 0 8px 16px rgba(239, 68, 68, 0.08);
}

.icon-rule-head .danger-text:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 18px rgba(239, 68, 68, 0.12);
}

.icon-rule-form {
  grid-template-columns: 1fr;
  gap: 0;
  padding: 6px 26px 10px;
}

.icon-rule-form > .field-card {
  display: grid;
  grid-template-columns: minmax(200px, 240px) minmax(0, 1fr);
  gap: 18px 22px;
  align-items: start;
  padding: 22px 0;
  border: 0;
  border-bottom: 1px solid #edf2f8;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.icon-rule-form > .field-card:hover {
  background: linear-gradient(90deg, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 4%, white), rgba(255, 255, 255, 0));
}

.icon-rule-form > .field-card:last-child {
  border-bottom: 0;
}

.icon-rule-form > .field-card > span {
  padding-top: 12px;
  color: #122033;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.4;
}

.icon-rule-form > .field-card > .field-input {
  min-height: 52px;
  border-radius: 18px;
  border-color: #d8e3ef;
  background: #fbfdff;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.icon-rule-form > .field-card > .field-input:focus {
  border-color: color-mix(in srgb, var(--icon-rule-tint, #2563eb) 55%, #93c5fd);
  box-shadow:
    0 0 0 4px color-mix(in srgb, var(--icon-rule-tint, #2563eb) 16%, transparent),
    0 14px 26px color-mix(in srgb, var(--icon-rule-tint, #2563eb) 10%, transparent);
}

.icon-rule-form > .field-card::after {
  grid-column: 2;
  margin-top: 2px;
  color: #8f9db2;
  font-size: 12px;
  line-height: 1.8;
}

.icon-rule-form > .field-card:nth-child(1)::after {
  content: '用于后台快速识别这条图标规则，建议按文件类型或业务场景命名。';
}

.icon-rule-form > .field-card:nth-child(2)::after {
  content: '支持 Emoji 或短图标字符，让文件列表更有辨识度。';
}

.icon-rule-form > .field-card:nth-child(3)::after {
  content: '使用逗号分隔多个扩展名或 MIME，如 pdf,docx,image/png。';
}

.icon-rule-form > .field-card:nth-child(4)::after {
  content: '推荐使用品牌色或类型色，统一文件系统的视觉记忆点。';
}

.icon-empty-card {
  border-radius: 28px;
  border: 1px dashed #cfe0f6;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(243, 248, 255, 0.94));
  box-shadow: 0 18px 34px rgba(148, 163, 184, 0.1);
}

.app-toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.app-add-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 52px;
  padding: 0 20px;
  border: 1px solid #edf2f7;
  border-radius: 18px;
  background: linear-gradient(180deg, #fff, #f8fafc);
  color: #55657d;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 10px 24px rgba(148, 163, 184, 0.1);
}

.app-add-button svg {
  width: 20px;
  height: 20px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
}

.app-studio-card {
  overflow: hidden;
  border: 1px solid #e4edf7;
  border-radius: 30px;
  background:
    radial-gradient(circle at top right, rgba(245, 158, 11, 0.08), transparent 20%),
    radial-gradient(circle at bottom left, rgba(20, 184, 166, 0.08), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(247, 250, 255, 0.96));
  box-shadow:
    0 24px 46px rgba(148, 163, 184, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.85);
}

.app-studio-head {
  display: grid;
  grid-template-columns: 190px minmax(240px, 1fr) minmax(260px, 1.1fr) minmax(220px, 1fr);
  gap: 20px;
  padding: 18px 24px;
  border-bottom: 1px solid #ebf1f8;
  background: linear-gradient(180deg, #fbfdff, #f3f8ff);
  color: #71829a;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.app-editor-grid {
  display: grid;
}

.app-editor-card {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 0 24px;
  padding: 24px;
  border-bottom: 1px solid #edf2f8;
}

.app-editor-card:last-child {
  border-bottom: 0;
}

.app-editor-hero {
  display: grid;
  align-content: start;
  gap: 14px;
}

.app-brand-large {
  width: 88px;
  height: 88px;
  border-radius: 28px;
  font-size: 34px;
  box-shadow:
    0 18px 36px rgba(15, 23, 42, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.28);
}

.app-hero-copy {
  display: grid;
  gap: 8px;
}

.app-hero-copy strong {
  color: #102038;
  font-size: 22px;
  line-height: 1.2;
}

.app-hero-copy span {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  min-height: 34px;
  align-items: center;
  padding: 0 12px;
  border-radius: 999px;
  border: 1px solid rgba(217, 228, 241, 0.95);
  background: rgba(255, 255, 255, 0.84);
  color: #73839c;
  line-height: 1.7;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.app-hero-copy small {
  color: #8291a7;
  line-height: 1.85;
}

.app-delete-button {
  width: fit-content;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 999px;
  border: 1px solid rgba(239, 68, 68, 0.16);
  background: linear-gradient(180deg, #fff5f5, #fff1f2);
  color: #dc2626;
  box-shadow: 0 8px 16px rgba(239, 68, 68, 0.08);
}

.app-field-stack {
  display: grid;
  gap: 0;
}

.app-field-row {
  display: grid;
  grid-template-columns: minmax(200px, 220px) minmax(0, 1fr);
  gap: 18px 22px;
  align-items: start;
  padding: 20px 0;
  border-bottom: 1px solid #edf2f8;
}

.app-field-row:last-child {
  border-bottom: 0;
  padding-bottom: 2px;
}

.app-field-row > span {
  padding-top: 12px;
  color: #122033;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.4;
}

.app-field-main {
  display: grid;
  gap: 8px;
}

.app-field-main .field-input {
  min-height: 52px;
  border-radius: 18px;
  border-color: #d8e3ef;
  background: #fbfdff;
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.app-field-main .field-input:focus {
  border-color: #14b8a6;
  box-shadow:
    0 0 0 4px rgba(20, 184, 166, 0.14),
    0 14px 26px rgba(20, 184, 166, 0.1);
}

.app-field-main small {
  color: #8d9cb0;
  font-size: 12px;
  line-height: 1.8;
}

.app-empty-card {
  margin: 24px;
}

.app-json-panels {
  grid-template-columns: 1fr;
}

.app-json-card {
  border-radius: 28px;
  border: 1px solid #dbe7f6;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(245, 249, 255, 0.95));
  box-shadow: 0 18px 34px rgba(148, 163, 184, 0.1);
}

.app-table-card {
  padding: 0;
}

.app-group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 18px 14px 16px;
  border-bottom: 1px solid #e8eef7;
  background: linear-gradient(180deg, #fbfdff, #f3f7fc);
  box-shadow: inset 0 -1px 0 rgba(255, 255, 255, 0.9);
}

.app-group-copy {
  display: grid;
  gap: 4px;
}

.app-group-head strong {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #102038;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.2;
}

.app-group-folder {
  display: inline-grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 8px;
  background: linear-gradient(180deg, #ecf3ff, #dbeafe);
  color: #2563eb;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.app-group-folder svg {
  width: 15px;
  height: 15px;
  fill: currentColor;
}

.app-group-head span {
  color: #7d8aa1;
  font-size: 12px;
  line-height: 1.6;
}

.app-group-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-group-count {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid #dce6f2;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.96);
  color: #607089;
  font-size: 11px;
  font-weight: 700;
}

.app-group-toggle {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 1px solid #d7e0ec;
  border-radius: 8px;
  background: linear-gradient(180deg, #fff, #f6f9fd);
  color: #66758b;
}

.app-group-toggle svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2.1;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: transform 0.2s ease;
}

.app-group-toggle svg.is-collapsed {
  transform: rotate(-90deg);
}

.app-table-shell {
  position: relative;
  overflow: auto;
  max-height: 720px;
  padding-right: 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.9) rgba(241, 245, 249, 0.95);
}

.app-table-shell::-webkit-scrollbar {
  width: 5px;
  height: 8px;
}

.app-table-shell::-webkit-scrollbar-track {
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(241, 245, 249, 0.68), rgba(226, 232, 240, 0.72));
}

.app-table-shell::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(148, 163, 184, 0.92), rgba(100, 116, 139, 0.96));
}

.app-table-shell::-webkit-scrollbar-corner {
  background: transparent;
}

.app-inline-editor-card {
  display: grid;
  gap: 14px;
  padding: 16px 22px;
  border-bottom: 1px solid #edf2f8;
  background: linear-gradient(180deg, #fbfdff, #f7fbff);
}

.app-inline-editor-head {
  display: flex;
  align-items: center;
  gap: 16px;
}

.app-inline-editor-brand {
  width: 56px;
  height: 56px;
  border-radius: 18px;
}

.app-inline-editor-copy {
  display: grid;
  gap: 4px;
}

.app-inline-editor-copy strong {
  color: #102038;
  font-size: 18px;
}

.app-inline-editor-copy span {
  color: #7d8aa1;
  font-size: 13px;
  line-height: 1.7;
}

.app-inline-editor-close {
  margin-left: auto;
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border: 1px solid #dce6f2;
  border-radius: 12px;
  background: #fff;
  color: #6b7c93;
}

.app-inline-editor-close svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.9;
  stroke-linecap: round;
}

.app-inline-editor-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
}

.app-inline-field {
  display: grid;
  gap: 8px;
}

.app-inline-field > span {
  color: #42556f;
  font-size: 13px;
  font-weight: 700;
}

.app-inline-field-wide {
  grid-column: span 2;
}

.app-table-head,
.app-table-row {
  display: grid;
  grid-template-columns: 58px 88px 228px minmax(320px, 1fr) 82px 108px 70px 98px;
  gap: 8px;
  align-items: center;
  min-width: 1060px;
}

.app-table-head {
  position: sticky;
  top: 0;
  z-index: 4;
  padding: 10px 16px;
  border-bottom: 1px solid #e8eef7;
  background: linear-gradient(180deg, rgba(251, 253, 255, 0.99), rgba(245, 249, 254, 0.99));
  color: #607089;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.app-table-body {
  display: grid;
}

.app-table-row {
  padding: 7px 16px;
  border-bottom: 1px solid #edf2f8;
  background: #fff;
  min-height: 58px;
}

.app-table-row.is-active {
  background: linear-gradient(180deg, #f5f9ff, #ffffff);
}

.app-table-row:last-child {
  border-bottom: 0;
}

.app-table-icon,
.app-type-cell,
.app-platform-cell,
.app-mapping-cell,
.app-enable-cell {
  display: flex;
  align-items: center;
}

.app-table-brand {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  font-size: 16px;
  font-weight: 800;
  box-shadow:
    0 5px 12px rgba(15, 23, 42, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.24);
}

.app-brand-svg {
  display: grid;
  place-items: center;
  width: 100%;
  height: 100%;
}

.app-brand-svg :deep(svg),
.app-brand-svg svg {
  width: 100%;
  height: 100%;
  display: block;
}

.app-brand-artplayer,
.app-brand-markdown,
.app-brand-drawio,
.app-brand-image,
.app-brand-monaco,
.app-brand-photopea,
.app-brand-excalidraw,
.app-brand-archive,
.app-brand-audio,
.app-brand-epub,
.app-brand-gdocs,
.app-brand-office,
.app-brand-pdf {
  letter-spacing: -0.04em;
}

.app-brand-drawio,
.app-brand-image,
.app-brand-excalidraw {
  font-size: 24px;
}

.app-brand-markdown,
.app-brand-pdf,
.app-brand-archive {
  font-size: 15px;
}

.app-type-badge,
.app-platform-pill,
.app-mapping-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 8px;
  border-radius: 999px;
  border: 1px solid #dbe4f1;
  background: #fff;
  color: #32455f;
  font-size: 11px;
  font-weight: 700;
}

.app-type-badge.is-built-in {
  background: linear-gradient(180deg, #eff6ff, #ffffff);
  color: #2563eb;
}

.app-type-badge.is-custom {
  background: linear-gradient(180deg, #f5f3ff, #ffffff);
  color: #7c3aed;
}

.app-name-cell,
.app-actions-cell {
  display: grid;
  gap: 4px;
}

.app-name-cell strong {
  color: #132238;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.45;
}

.app-name-cell small {
  color: #8c9ab0;
  font-size: 12px;
  line-height: 1.55;
}

.app-inline-input,
.app-mini-input,
.app-inline-field .field-input {
  min-height: 42px;
  border-radius: 14px;
  border-color: #d8e3ef;
  background: #fbfdff;
}

.app-extensions-cell {
  display: flex;
  align-items: center;
}

.app-extensions-input {
  min-height: 32px;
  padding: 0 10px;
  border-radius: 10px;
  border-color: #d8e3ef;
  background: #fbfdff;
  font-size: 11px;
  line-height: 1.4;
}

.app-toggle-readonly {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #3d4f68;
  font-size: 11px;
  font-weight: 700;
}

.app-toggle-readonly input {
  width: 15px;
  height: 15px;
  accent-color: #2f7de1;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.app-row-buttons {
  display: flex;
  gap: 4px;
  align-items: center;
  justify-content: flex-start;
}

.app-row-button {
  min-width: 26px;
  min-height: 26px;
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: #7b8798;
  font-weight: 800;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.app-row-button-line {
  display: grid;
  place-items: center;
  padding: 0;
}

.app-row-button:hover {
  border-color: #dce5f1;
  background: #f8fbff;
  color: #1f3b63;
  transform: translateY(-1px);
}

.app-row-button-line svg {
  width: 14px;
  height: 14px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.app-row-button.danger {
  color: #dc2626;
}

.emoji-config-card {
  display: grid;
  gap: 20px;
  padding: 24px;
  border: 1px solid #e6edf8;
  border-radius: 28px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(247, 250, 255, 0.96));
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.12);
}

.emoji-config-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.emoji-config-head h3 {
  margin: 0;
  color: #102038;
  font-size: 22px;
  line-height: 1.2;
}

.emoji-config-head p {
  margin: 8px 0 0;
  color: #7b8aa2;
  line-height: 1.7;
}

.emoji-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.emoji-meta-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.emoji-meta-grid .field-card {
  min-height: 148px;
}

.emoji-meta-grid .field-card small {
  color: #90a0b7;
  line-height: 1.7;
}

.emoji-config-tip {
  display: grid;
  align-content: space-between;
  gap: 10px;
  min-height: 148px;
  padding: 20px 22px;
  border: 1px solid #dbe7f6;
  border-radius: 22px;
  background: linear-gradient(135deg, #eff6ff, #ffffff 58%, #eef2ff);
}

.emoji-config-tip strong {
  color: #1d4ed8;
  font-size: 36px;
  line-height: 1;
}

.emoji-config-tip span {
  color: #57708f;
  line-height: 1.8;
}

.emoji-option-table {
  overflow: hidden;
  border: 1px solid #e3ebf7;
  border-radius: 24px;
  background: #fff;
}

.emoji-option-head,
.emoji-option-row {
  display: grid;
  grid-template-columns: minmax(220px, 0.95fr) 180px minmax(260px, 1.2fr) 150px;
  gap: 18px;
}

.emoji-option-head {
  padding: 16px 20px;
  background: linear-gradient(180deg, #f8fbff, #f2f7ff);
  border-bottom: 1px solid #e8eef8;
  color: #6a7a92;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.05em;
}

.emoji-option-body {
  display: grid;
}

.emoji-option-row {
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #edf2f8;
}

.emoji-option-row:last-child {
  border-bottom: 0;
}

.emoji-rule-column,
.emoji-category-column,
.emoji-cloud-column {
  display: grid;
  gap: 10px;
}

.emoji-rule-column label {
  display: grid;
  gap: 8px;
}

.emoji-rule-column label span {
  color: #64748b;
  font-size: 12px;
  font-weight: 800;
}

.emoji-rule-column small,
.emoji-category-column small,
.emoji-cloud-column small {
  color: #8ea0b7;
  line-height: 1.7;
}

.emoji-category-badge {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  min-height: 88px;
  padding: 12px;
  border: 1px solid #dde7f3;
  border-radius: 20px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
}

.emoji-category-preview {
  display: grid;
  place-items: center;
  width: 72px;
  height: 60px;
  border-radius: 18px;
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  box-shadow: 0 14px 24px rgba(37, 99, 235, 0.18);
  color: #fff;
  font-size: 28px;
}

.emoji-category-input {
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid #d7e2ef;
  border-radius: 14px;
  background: #fbfdff;
  color: #0f172a;
  font-size: 15px;
  font-weight: 600;
}

.emoji-cloud-input {
  width: 100%;
  min-height: 108px;
  padding: 14px 16px;
  border: 1px solid #d7e2ef;
  border-radius: 20px;
  background: linear-gradient(180deg, #fbfdff, #ffffff);
  color: #0f172a;
  font-size: 18px;
  line-height: 1.85;
  resize: vertical;
}

.emoji-category-input:focus,
.emoji-cloud-input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.16);
}

.emoji-row-actions {
  display: grid;
  gap: 10px;
}

.emoji-action-button {
  min-height: 40px;
  border: 1px solid #d8e3ef;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  color: #37506f;
  font-weight: 700;
}

.emoji-action-button.danger {
  border-color: rgba(239, 68, 68, 0.18);
  background: linear-gradient(180deg, #fff5f5, #fff1f2);
  color: #dc2626;
}

@media (max-width: 1180px) {
  .emoji-option-head {
    display: none;
  }

  .emoji-option-row {
    grid-template-columns: 1fr;
  }

  .emoji-row-actions {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.icon-json-stack {
  grid-template-columns: 1fr;
  gap: 18px;
}

.icon-json-card {
  border: 1px solid #e7edf5;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 252, 0.96));
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.12);
}

.icon-json-card > span {
  color: #132238;
  font-size: 15px;
  font-weight: 800;
}

.icon-json-card .field-textarea {
  border-radius: 18px;
  background: #fbfdff;
}

.icon-table-card {
  overflow: hidden;
  border: 1px solid #e6edf7;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 16px 30px rgba(148, 163, 184, 0.08);
}

.icon-table-head {
  display: grid;
  grid-template-columns: 72px 180px 140px minmax(280px, 1fr) 160px 74px;
  gap: 12px;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e8eef7;
  color: #132238;
  font-size: 14px;
  font-weight: 800;
}

.icon-rule-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
}

.icon-table-row {
  display: grid;
  grid-template-columns: 72px 180px 140px minmax(280px, 1fr) 160px 74px;
  gap: 12px;
  align-items: center;
  padding: 10px 20px;
  border-top: 1px solid #edf2f8;
  background: #fff;
}

.icon-table-row:first-child {
  border-top: 0;
}

.icon-table-row .editor-head,
.icon-table-row .icon-rule-form {
  display: contents;
}

.icon-table-row .editor-title {
  display: none;
}

.icon-table-row .preview-icon.large {
  grid-column: 1;
  width: 40px;
  height: 40px;
  margin-left: 4px;
  border-radius: 12px;
  font-size: 18px;
  color: #fff;
  background: linear-gradient(135deg, color-mix(in srgb, var(--icon-rule-tint, #2563eb) 82%, white), var(--icon-rule-tint, #2563eb));
  box-shadow: none;
}

.icon-table-row .danger-text {
  grid-column: 6;
  justify-self: center;
}

.icon-table-row .field-card {
  display: block;
  padding: 0;
  border: 0;
  background: transparent;
  box-shadow: none;
}

.icon-table-row .field-card > span,
.icon-table-row .field-card::after {
  display: none;
}

.icon-table-row .field-card:nth-child(1) {
  grid-column: 2;
}

.icon-table-row .field-card:nth-child(2) {
  grid-column: 3;
}

.icon-table-row .field-card:nth-child(3) {
  grid-column: 4;
}

.icon-table-row .field-card:nth-child(4) {
  grid-column: 5;
}

.icon-table-row .field-input {
  min-height: 42px;
  padding: 10px 14px;
  border-radius: 16px;
  border-color: #d8e2ef;
  background: #fff;
  box-shadow: none;
}

.icon-table-row .field-input:focus {
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.12);
}

.icon-row-delete {
  min-height: 34px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(239, 68, 68, 0.12);
  background: #fff5f5;
}

.custom-props-shell {
  display: grid;
  gap: 18px;
}

.custom-props-header {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) 360px;
  gap: 18px;
  padding: 8px 2px 2px;
}

.custom-props-copy {
  display: grid;
  gap: 10px;
}

.custom-props-kicker {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border: 1px solid rgba(59, 130, 246, 0.16);
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.92));
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.custom-props-copy h2 {
  margin: 0;
  color: #0f172a;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: -0.04em;
}

.custom-props-copy p {
  margin: 0;
  max-width: 860px;
  color: #73839c;
  line-height: 1.85;
}

.custom-props-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.custom-props-metric {
  display: grid;
  gap: 6px;
  padding: 18px 18px 16px;
  border: 1px solid #e3ebf5;
  border-radius: 20px;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.12), transparent 36%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 255, 0.96));
  box-shadow: 0 16px 32px rgba(148, 163, 184, 0.12);
}

.custom-props-metric span {
  color: #7a8aa2;
  font-size: 12px;
  font-weight: 700;
}

.custom-props-metric strong {
  color: #102038;
  font-size: 26px;
  line-height: 1;
}

.custom-props-toolbar {
  display: flex;
  align-items: center;
}

.custom-props-add-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 52px;
  padding: 0 20px;
  border: 1px solid rgba(29, 78, 216, 0.12);
  border-radius: 18px;
  background: linear-gradient(135deg, #ffffff, #eff6ff 52%, #dbeafe);
  color: #1e3a8a;
  font-size: 15px;
  font-weight: 800;
  box-shadow:
    0 18px 34px rgba(59, 130, 246, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.custom-props-add-button svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
}

.custom-props-table-card {
  overflow: hidden;
  border: 1px solid #e4ebf6;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 255, 0.97));
  box-shadow: 0 20px 40px rgba(148, 163, 184, 0.12);
}

.custom-props-table-head,
.custom-props-table-row {
  display: grid;
  grid-template-columns: minmax(260px, 1.2fr) 180px minmax(260px, 1fr) 190px;
  gap: 18px;
  align-items: center;
}

.custom-props-table-head {
  padding: 16px 20px;
  border-bottom: 1px solid #e8eef7;
  background: linear-gradient(180deg, #ffffff, #f6f9fd);
  color: #132238;
  font-size: 14px;
  font-weight: 800;
}

.custom-props-table-body {
  display: grid;
}

.custom-props-table-row {
  padding: 14px 20px;
  border-top: 1px solid #edf2f8;
  background: #fff;
}

.custom-props-table-row:first-child {
  border-top: 0;
}

.custom-props-table-row.is-active {
  background: linear-gradient(180deg, #f8fbff, #ffffff);
}

.custom-props-name-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}

.custom-prop-icon {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 12px;
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  box-shadow: 0 10px 20px rgba(59, 130, 246, 0.18);
}

.custom-prop-icon.is-text,
.custom-props-type-badge.is-text {
  background: linear-gradient(135deg, #64748b, #94a3b8);
}

.custom-prop-icon.is-textarea,
.custom-props-type-badge.is-textarea {
  background: linear-gradient(135deg, #0f766e, #14b8a6);
}

.custom-prop-icon.is-switch,
.custom-props-type-badge.is-switch {
  background: linear-gradient(135deg, #2563eb, #38bdf8);
}

.custom-prop-icon.is-date,
.custom-props-type-badge.is-date {
  background: linear-gradient(135deg, #7c3aed, #a78bfa);
}

.custom-prop-icon.is-tags,
.custom-props-type-badge.is-tags {
  background: linear-gradient(135deg, #ea580c, #fb923c);
}

.custom-prop-icon.is-multi_select,
.custom-props-type-badge.is-multi_select {
  background: linear-gradient(135deg, #0891b2, #22d3ee);
}

.custom-prop-icon.is-rating,
.custom-props-type-badge.is-rating {
  background: linear-gradient(135deg, #f59e0b, #facc15);
}

.custom-prop-name-copy {
  display: grid;
  gap: 4px;
}

.custom-prop-name-copy strong {
  color: #132238;
  font-size: 15px;
  font-weight: 700;
}

.custom-prop-name-copy span {
  color: #8b99ad;
  font-size: 12px;
}

.custom-props-type-cell,
.custom-props-default-cell,
.custom-props-actions-cell {
  display: flex;
  align-items: center;
}

.custom-props-type-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.24);
}

.custom-default-text {
  color: #8a97ad;
  font-size: 13px;
  line-height: 1.7;
}

.custom-default-text.large {
  color: #6b7c93;
  font-size: 15px;
}

.custom-rating-preview {
  display: inline-flex;
  gap: 4px;
  align-items: center;
  line-height: 1;
}

.custom-rating-preview.large {
  gap: 8px;
}

.custom-rating-star {
  padding: 0;
  border: 0;
  background: transparent;
  color: #c5ceda;
  font-size: 24px;
  line-height: 1;
  transition:
    color 0.2s ease,
    transform 0.2s ease,
    text-shadow 0.2s ease;
}

.custom-rating-preview.large .custom-rating-star {
  font-size: 38px;
}

.custom-rating-star.is-filled {
  color: #f5b014;
  text-shadow: 0 6px 14px rgba(245, 176, 20, 0.22);
}

.custom-rating-star:hover {
  transform: translateY(-1px) scale(1.04);
}

.custom-switch-preview {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid #dce5f1;
  border-radius: 999px;
  background: linear-gradient(180deg, #f8fbff, #ffffff);
  color: #5d708c;
  font-size: 12px;
  font-weight: 700;
}

.custom-switch-preview.is-on {
  border-color: rgba(37, 99, 235, 0.18);
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  color: #fff;
  box-shadow: 0 10px 18px rgba(37, 99, 235, 0.18);
}

.custom-switch-preview.large {
  min-height: 42px;
  padding: 0 18px;
  font-size: 14px;
}

.custom-switch-preview.interactive {
  cursor: pointer;
}

.custom-date-preview {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid #dbe4f1;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  color: #4f6380;
  font-size: 12px;
  font-weight: 700;
}

.custom-date-preview.large {
  min-height: 44px;
  padding: 0 16px;
  font-size: 14px;
}

.custom-options-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.custom-options-preview span {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 12px;
  border: 1px solid #dbe4f1;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  color: #4f6380;
  font-size: 12px;
  font-weight: 700;
}

.custom-options-preview.is-multi span {
  border-style: dashed;
}

.custom-options-preview.large span {
  min-height: 38px;
  padding: 0 14px;
  font-size: 13px;
}

.custom-default-text.textarea {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.custom-props-actions-cell {
  gap: 6px;
  justify-content: flex-start;
}

.custom-action-button {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: #75859c;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.custom-action-button:hover {
  border-color: #dbe4f1;
  background: #f8fbff;
  color: #1f3b63;
  transform: translateY(-1px);
}

.custom-action-button.danger {
  color: #dc2626;
}

.custom-action-button svg,
.custom-props-modal-close svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.custom-props-empty-card {
  margin: 22px;
}

.custom-props-modal-layer {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: grid;
  place-items: center;
  padding: 28px;
  background: rgba(15, 23, 42, 0.26);
  backdrop-filter: blur(7px) saturate(1.02);
}

.custom-props-modal-card {
  width: min(100%, 860px);
  max-height: calc(100vh - 56px);
  display: grid;
  gap: 22px;
  padding: 30px 32px 26px;
  border: 1px solid #e4ebf5;
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.08), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.995), rgba(248, 250, 255, 0.985));
  box-shadow:
    0 30px 60px rgba(15, 23, 42, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  overflow: auto;
}

.custom-props-modal-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.custom-props-modal-head strong {
  display: block;
  color: #111827;
  font-size: 24px;
  font-weight: 800;
}

.custom-props-modal-head span {
  display: block;
  margin-top: 8px;
  color: #7b8aa2;
  line-height: 1.8;
}

.custom-props-modal-close {
  display: grid;
  place-items: center;
  width: 42px;
  height: 42px;
  border: 1px solid #dbe4f1;
  border-radius: 14px;
  background: #fff;
  color: #73839c;
}

.custom-props-modal-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
}

.custom-props-modal-field {
  display: grid;
  gap: 8px;
}

.custom-props-modal-field > span,
.custom-props-modal-preview > span {
  color: #172033;
  font-size: 15px;
  font-weight: 800;
}

.custom-props-modal-field small {
  color: #8b99ad;
  font-size: 12px;
  line-height: 1.75;
}

.custom-props-modal-field .field-input,
.custom-props-modal-field .field-select {
  min-height: 54px;
  border-radius: 18px;
  border-color: #d8e3ef;
  background: linear-gradient(180deg, #ffffff, #fbfdff);
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.custom-props-modal-field .field-input:focus,
.custom-props-modal-field .field-select:focus {
  box-shadow:
    0 0 0 4px rgba(96, 165, 250, 0.14),
    0 14px 28px rgba(96, 165, 250, 0.1);
}

.custom-props-modal-field-wide,
.custom-props-modal-preview {
  grid-column: 1 / -1;
}

.custom-props-required-toggle {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 54px;
  padding: 0 16px;
  border: 1px solid #d8e3ef;
  border-radius: 18px;
  background: linear-gradient(180deg, #ffffff, #fbfdff);
  color: #1f334d;
}

.custom-props-required-toggle input {
  width: 18px;
  height: 18px;
  accent-color: #2563eb;
}

.custom-props-required-toggle strong {
  font-size: 14px;
}

.custom-props-modal-preview {
  display: grid;
  gap: 10px;
}

.custom-props-modal-preview-box {
  min-height: 92px;
  display: flex;
  align-items: center;
  padding: 18px 20px;
  border: 1px solid #dbe4f1;
  border-radius: 20px;
  background: linear-gradient(180deg, #ffffff, #f8fbff);
}

.custom-props-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.custom-props-json-card {
  border: 1px solid #e7edf5;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(248, 250, 252, 0.96));
  box-shadow: 0 18px 38px rgba(148, 163, 184, 0.12);
}

.editor-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.editor-card {
  gap: 16px;
}

.editor-head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 14px;
  align-items: center;
}

.editor-title {
  display: grid;
  gap: 4px;
}

.preview-icon,
.app-brand {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  font-size: 26px;
  color: #fff;
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.24);
}

.preview-icon.large {
  width: 64px;
  height: 64px;
  font-size: 30px;
}

.danger-text {
  border: 0;
  background: transparent;
  color: #dc2626;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
}

.inline-actions-left {
  justify-content: flex-start;
}

.action-bar {
  position: sticky;
  bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 20px;
  border-radius: 22px;
  backdrop-filter: blur(14px);
}

.action-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.primary-button,
.ghost-button,
.secondary-button {
  min-height: 46px;
  padding: 0 18px;
  border-radius: 15px;
  font-size: 14px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.primary-button {
  border: 1px solid rgba(37, 99, 235, 0.2);
  background: linear-gradient(135deg, #2563eb, #0ea5e9 52%, #22c55e);
  color: #fff;
  box-shadow: 0 14px 26px rgba(37, 99, 235, 0.24);
}

.ghost-button {
  border: 1px solid #d6e3f3;
  background: #fff;
  color: #28415f;
}

.secondary-button {
  border: 1px solid rgba(37, 99, 235, 0.18);
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #1d4ed8;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.08);
}

.primary-button:disabled,
.ghost-button:disabled,
.secondary-button:disabled {
  cursor: not-allowed;
  opacity: 0.65;
  transform: none;
  box-shadow: none;
}

/* Aurora glass refresh: shared by all five file-system tabs. */
.file-system-page {
  gap: 24px;
  padding-bottom: 12px;
}

.hero-card,
.panel-card,
.connection-alert,
.action-bar,
.overview-card,
.metric-card,
.group-card,
.field-card,
.toggle-card,
.editor-card,
.empty-card,
.icon-table-card,
.app-studio-card,
.custom-props-table-card,
.custom-props-modal-card,
.custom-props-json-card,
.search-config-card,
.job-hero,
.job-stats-card {
  border-color: rgba(210, 228, 247, 0.78);
  background:
    radial-gradient(circle at 12% 8%, rgba(125, 211, 252, 0.12), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.84), rgba(249, 252, 255, 0.66));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 22px 50px rgba(79, 124, 176, 0.12);
  backdrop-filter: blur(20px);
}

.hero-card {
  min-height: 414px;
  padding: 40px 38px;
  border-radius: 30px;
  background:
    radial-gradient(circle at 7% 10%, rgba(116, 220, 255, 0.28), transparent 28%),
    radial-gradient(circle at 92% 74%, rgba(255, 232, 139, 0.46), transparent 30%),
    radial-gradient(circle at 72% 24%, rgba(255, 199, 220, 0.24), transparent 26%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.78), rgba(241, 249, 255, 0.64));
}

.hero-copy h1 {
  color: #111d33;
  font-size: 46px;
  font-weight: 900;
}

.hero-text {
  max-width: 760px;
  color: #60718c;
  font-size: 16px;
}

.eyebrow,
.settings-kicker,
.search-heading::before {
  color: #256dff;
  letter-spacing: 0.2em;
  text-shadow: 0 0 18px rgba(74, 144, 255, 0.18);
}

.hero-tag,
.tab-button.active,
.primary-button {
  background: linear-gradient(135deg, #1f72ff 0%, #16a7ef 58%, #16c8ba 100%);
  box-shadow:
    0 16px 34px rgba(37, 118, 235, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.hero-tag-soft {
  background: rgba(255, 255, 255, 0.62);
  color: #256dff;
}

.hero-metrics,
.overview-grid,
.app-overview-grid,
.custom-props-metrics {
  gap: 18px;
}

.metric-card,
.overview-card,
.settings-side-metric {
  min-height: 160px;
  border-radius: 22px;
  padding: 22px 24px;
  overflow: hidden;
}

.metric-card strong,
.overview-card strong,
.settings-side-metric strong {
  color: #122039;
  font-size: 28px;
  font-weight: 900;
}

.metric-card span,
.overview-card span,
.settings-side-metric span,
.metric-card small,
.overview-card small {
  color: #72839d;
}

.accent-blue,
.search-metric-card-blue,
.icon-overview-card-blue {
  background: linear-gradient(160deg, rgba(225, 240, 255, 0.94), rgba(255, 255, 255, 0.74));
}

.accent-gold,
.search-metric-card-gold,
.icon-overview-card-gold {
  background: linear-gradient(160deg, rgba(255, 246, 165, 0.78), rgba(255, 255, 255, 0.76));
}

.accent-green,
.search-metric-card-emerald {
  background: linear-gradient(160deg, rgba(201, 250, 218, 0.78), rgba(255, 255, 255, 0.76));
}

.accent-cyan,
.search-metric-card-violet,
.icon-overview-card-violet {
  background: linear-gradient(160deg, rgba(187, 245, 250, 0.82), rgba(255, 255, 255, 0.76));
}

.connection-alert {
  min-height: 96px;
  padding: 20px 22px;
  border-radius: 24px;
}

.connection-alert.is-success {
  border-color: rgba(89, 222, 156, 0.38);
  background:
    radial-gradient(circle at 8% 30%, rgba(103, 232, 181, 0.2), transparent 28%),
    linear-gradient(135deg, rgba(240, 255, 249, 0.92), rgba(255, 255, 255, 0.78));
}

.panel-card {
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(246, 251, 255, 0.66));
}

.tabs-row {
  gap: 12px;
  padding: 20px 22px 0;
}

.tab-button {
  min-height: 56px;
  padding: 0 24px;
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(237, 246, 255, 0.74));
  color: #33445d;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 14px 28px rgba(97, 141, 190, 0.08);
}

.tab-button:focus {
  outline: none;
}

.tab-button:focus-visible {
  box-shadow:
    0 0 0 4px rgba(99, 179, 255, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 14px 28px rgba(97, 141, 190, 0.1);
}

.panel-body {
  gap: 22px;
  padding: 28px 24px 32px;
  background:
    radial-gradient(circle at 8% 0%, rgba(125, 211, 252, 0.12), transparent 30%),
    linear-gradient(180deg, rgba(247, 251, 255, 0.78), rgba(244, 249, 255, 0.62));
}

.section-header h2,
.settings-header h2::before,
.search-header h2::before,
.icon-section-header h2,
.app-section-header h2,
.custom-props-title h2 {
  color: #111d33;
  font-weight: 900;
}

.settings-stack .group-card,
.settings-text-stack > .field-card,
.search-config-card,
.icon-table-card,
.app-studio-card,
.custom-props-table-card,
.custom-props-json-card {
  border-radius: 26px;
}

.settings-section-head,
.search-card-head,
.icon-table-head,
.app-table-head,
.custom-props-table-head,
.app-group-head,
.custom-props-modal-head {
  border-bottom-color: rgba(222, 234, 247, 0.82);
  background:
    radial-gradient(circle at 8% 0%, rgba(116, 220, 255, 0.12), transparent 28%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.76), rgba(247, 251, 255, 0.62));
}

.settings-stack .field-card,
.settings-stack .toggle-card,
.nested-list .field-card,
.settings-text-stack > .field-card,
.app-field-row,
.custom-props-modal-field,
.app-table-row,
.custom-props-table-row {
  border-bottom-color: rgba(228, 238, 248, 0.92);
}

.field-input,
.field-select,
.field-textarea,
.settings-stack .field-input,
.settings-stack .field-select,
.settings-stack .field-textarea,
.settings-text-stack .field-input,
.settings-text-stack .field-select,
.settings-text-stack .field-textarea,
.search-field-card .field-input,
.app-field-main .field-input,
.app-inline-field .field-input,
.custom-props-modal-field .field-input {
  min-height: 54px;
  border-color: rgba(203, 222, 242, 0.94);
  border-radius: 17px;
  background: rgba(251, 253, 255, 0.78);
  color: #22314a;
  box-shadow:
    inset 0 1px 2px rgba(148, 163, 184, 0.08),
    0 10px 22px rgba(111, 151, 197, 0.05);
}

.field-input:focus,
.field-select:focus,
.field-textarea:focus,
.settings-stack .field-input:focus,
.settings-stack .field-select:focus,
.settings-stack .field-textarea:focus,
.settings-text-stack .field-input:focus,
.settings-text-stack .field-select:focus,
.settings-text-stack .field-textarea:focus,
.search-field-card .field-input:focus,
.app-field-main .field-input:focus,
.app-inline-field .field-input:focus,
.custom-props-modal-field .field-input:focus {
  border-color: #74b7ff;
  box-shadow:
    0 0 0 5px rgba(90, 169, 255, 0.2),
    0 16px 34px rgba(83, 149, 225, 0.12);
}

.search-toggle-card input[type='checkbox'] {
  background: linear-gradient(180deg, #f7fbff, #dfeaf5);
  box-shadow:
    inset 0 1px 2px rgba(118, 144, 174, 0.18),
    0 10px 20px rgba(91, 133, 176, 0.12);
}

.search-toggle-card input[type='checkbox']:checked {
  background: linear-gradient(135deg, #2f7df5, #22c5de);
}

.icon-table-row,
.app-table-row,
.custom-props-table-row {
  background: rgba(255, 255, 255, 0.56);
}

.icon-table-row:hover,
.app-table-row:hover,
.custom-props-table-row:hover,
.app-table-row.is-active,
.custom-props-table-row.is-active {
  background:
    linear-gradient(90deg, rgba(232, 247, 255, 0.84), rgba(255, 238, 243, 0.62));
  box-shadow: inset 4px 0 0 rgba(75, 156, 255, 0.58);
}

.preview-icon,
.app-brand,
.custom-prop-icon {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.36),
    0 14px 28px rgba(61, 132, 217, 0.18);
}

.action-bar {
  bottom: 18px;
  min-height: 102px;
  padding: 20px 24px;
  border-radius: 24px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(247, 251, 255, 0.76));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 24px 56px rgba(72, 116, 170, 0.18);
  z-index: 20;
}

.action-copy strong {
  color: #122039;
  font-size: 18px;
}

.primary-button,
.ghost-button,
.secondary-button {
  min-height: 56px;
  padding: 0 24px;
  border-radius: 18px;
}

.ghost-button,
.secondary-button {
  background: rgba(255, 255, 255, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 12px 24px rgba(104, 143, 188, 0.1);
}

@media (max-width: 1280px) {
  .hero-card,
  .overview-grid,
  .group-grid,
  .editor-grid,
  .job-stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .hero-card {
    grid-template-columns: 1fr;
  }

  .settings-stack .field-card,
  .settings-stack .toggle-card {
    grid-template-columns: minmax(180px, 220px) minmax(0, 1fr);
  }

  .icon-table-head,
  .icon-table-row {
    grid-template-columns: 64px 150px 110px minmax(180px, 1fr) 130px 68px;
    gap: 10px;
    padding-left: 14px;
    padding-right: 14px;
  }

  .app-studio-head {
    display: none;
  }

  .app-editor-card {
    grid-template-columns: 1fr;
    gap: 22px;
  }

  .app-table-head,
  .app-table-row {
    min-width: 1360px;
  }

  .app-inline-editor-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .custom-props-header {
    grid-template-columns: 1fr;
  }

  .custom-props-table-head,
  .custom-props-table-row {
    min-width: 920px;
  }

  .custom-props-modal-form {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .form-grid,
  .toggle-grid,
  .text-panels,
  .hero-metrics,
  .overview-grid,
  .group-grid,
  .editor-grid,
  .job-stats-grid {
    grid-template-columns: 1fr;
  }

  .section-header,
  .job-hero,
  .action-bar {
    display: grid;
    grid-template-columns: 1fr;
  }

  .connection-alert {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .connection-alert-actions {
    width: 100%;
    justify-content: stretch;
  }

  .action-group {
    width: 100%;
  }

  .settings-stack .field-card,
  .settings-stack .toggle-card {
    grid-template-columns: 1fr;
  }

  .app-overview-grid {
    grid-template-columns: 1fr;
  }

  .app-section-header .secondary-button {
    width: 100%;
  }

  .app-field-row {
    grid-template-columns: 1fr;
  }

  .app-field-row > span {
    padding-top: 0;
  }

  .app-group-head {
    display: grid;
    grid-template-columns: 1fr;
  }

  .app-inline-editor-head,
  .app-inline-editor-grid {
    grid-template-columns: 1fr;
  }

  .app-inline-editor-head {
    display: grid;
  }

  .app-inline-field-wide {
    grid-column: span 1;
  }

  .custom-props-metrics {
    grid-template-columns: 1fr;
  }

  .custom-props-table-head {
    display: none;
  }

  .custom-props-table-row {
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .custom-props-default-cell,
  .custom-props-type-cell,
  .custom-props-actions-cell {
    justify-content: flex-start;
  }

  .custom-props-modal-card {
    padding: 22px 18px 20px;
    border-radius: 22px;
  }

  .custom-props-modal-layer {
    padding: 14px;
  }

  .icon-overview-grid {
    grid-template-columns: 1fr;
  }

  .icon-table-head {
    display: none;
  }

  .icon-table-row {
    grid-template-columns: 52px 1fr;
    gap: 10px;
    padding: 14px;
  }

  .icon-table-row .preview-icon.large {
    grid-column: 1;
    grid-row: 1 / span 5;
    margin-left: 0;
  }

  .icon-table-row .field-card:nth-child(1),
  .icon-table-row .field-card:nth-child(2),
  .icon-table-row .field-card:nth-child(3),
  .icon-table-row .field-card:nth-child(4),
  .icon-table-row .danger-text {
    grid-column: 2;
  }

  .icon-table-row .danger-text {
    justify-self: start;
  }

  .icon-rule-form > .field-card {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .icon-rule-form > .field-card > span,
  .icon-rule-form > .field-card::after {
    grid-column: 1;
  }

  .icon-rule-form > .field-card > span {
    padding-top: 0;
  }

  .settings-stack .field-card > span,
  .settings-stack .field-card > .field-input,
  .settings-stack .field-card > .field-select,
  .settings-stack .field-card > .field-textarea,
  .settings-stack .field-card > .split-row,
  .settings-stack .field-card::after,
  .settings-stack .toggle-card > input[type='checkbox'],
  .settings-stack .toggle-card > div,
  .settings-stack .toggle-card::before {
    grid-column: 1;
    grid-row: auto;
  }

  .settings-stack .field-card > span {
    padding-top: 0;
  }

  .settings-stack .toggle-card > input[type='checkbox'] {
    margin-top: 2px;
  }

  .settings-stack .toggle-card > div {
    padding-left: 30px;
    margin-top: -28px;
  }

  .settings-side-metrics {
    grid-template-columns: 1fr 1fr;
  }

  .primary-button,
  .ghost-button,
  .secondary-button {
    width: 100%;
  }

  .settings-section-head-inline {
    display: grid;
    grid-template-columns: 1fr;
  }

  .panel-body > section.settings-stack:last-of-type {
    grid-template-columns: 1fr;
  }

  .settings-side-metrics {
    grid-template-columns: 1fr;
  }

  .settings-form-list.two-column,
  .nested-list {
    grid-template-columns: 1fr;
  }

  .settings-stack .form-grid > * ,
  .settings-text-stack .nested-list > * {
    border-bottom: 1px solid #eef3f8;
  }

  .settings-stack .form-grid > :last-child,
  .settings-text-stack .nested-list > :last-child {
    border-bottom: 0;
  }
}
</style>
