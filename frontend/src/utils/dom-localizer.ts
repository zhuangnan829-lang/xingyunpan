import { watch } from 'vue';
import { currentLanguage } from './language';

type AttrName = 'placeholder' | 'title' | 'aria-label';

const textOriginals = new WeakMap<Text, string>();
const attrOriginals = new WeakMap<Element, Map<AttrName, string>>();

const hasHan = /[\u3400-\u9fff]/;
const hanRun = /[\u3400-\u9fff]+/g;

const exactDictionary: Record<string, string> = {
  // Navigation and common shell
  '星云盘': 'Xingyunpan',
  '个人工作区': 'Personal workspace',
  '管理控制台': 'Admin console',
  '我的文件': 'My Files',
  '图片': 'Images',
  '视频': 'Videos',
  '音乐': 'Music',
  '文档': 'Documents',
  '与我共享': 'Shared with me',
  '回收站': 'Trash',
  '我的分享': 'My Shares',
  '连接与挂载': 'Connections',
  '后台任务': 'Background jobs',
  '离线下载': 'Offline downloads',
  '管理面板': 'Admin panel',
  '面板首页': 'Dashboard',
  '参数设置': 'Parameter settings',
  '文件系统参数设置': 'File system parameter settings',
  '全文搜索': 'Full-text search',
  '全文搜索中心': 'Full text search center',
  '文件系统': 'File system',
  '存储策略': 'Storage policy',
  '节点': 'Nodes',
  '用户组': 'User groups',
  '用户': 'Users',
  '文件': 'Files',
  '文件 Blob': 'File blobs',
  '文件 blobs': 'File blobs',
  '分享': 'Shares',
  'OAuth 应用': 'OAuth apps',
  '返回主页': 'Back home',
  '登录': 'Sign in',
  '注册': 'Sign up',
  '找回密码': 'Reset password',
  '分享访问': 'Share access',
  '分享详情': 'Share details',
  '个人资料': 'Profile',
  '协作': 'Collaboration',
  '存储概览': 'Storage overview',

  // Common actions and states
  '新建': 'New',
  '上传文件': 'Upload files',
  '新建文件夹': 'New folder',
  '创建新账号': 'Create account',
  '搜索': 'Search',
  '清除搜索': 'Clear search',
  '刷新': 'Refresh',
  '重新加载': 'Reload',
  '重试': 'Retry',
  '删除': 'Delete',
  '批量删除': 'Batch delete',
  '清空': 'Empty',
  '取消': 'Cancel',
  '确定': 'Confirm',
  '确认': 'Confirm',
  '保存': 'Save',
  '保存更改': 'Save changes',
  '保存中...': 'Saving...',
  '打开': 'Open',
  '下载': 'Download',
  '复制': 'Copy',
  '复制链接': 'Copy link',
  '复制直链': 'Copy direct link',
  '重命名': 'Rename',
  '移动到': 'Move to',
  '复制到': 'Copy to',
  '版本历史': 'Version history',
  '协作管理': 'Collaboration',
  '筛选': 'Filter',
  '重置': 'Reset',
  '应用筛选': 'Apply filters',
  '取消选择': 'Cancel selection',
  '批量管理': 'Batch manage',
  '恢复默认': 'Restore defaults',
  '恢复当前页': 'Restore current tab',
  '保存当前页': 'Save current tab',
  '恢复当前页签': 'Restore current tab',
  '保存当前页签': 'Save current tab',
  '底部操作仅影响当前页签，不会误改其他模块。': 'Bottom actions affect only the current tab and will not change other modules.',
  '支持真实保存': 'supports live saving',
  '当前页签': 'Current tab',
  '支持独立保存与恢复': 'Supports independent save and restore',
  '图标规则': 'Icon rules',
  '扩展名、MIME 与默认图标': 'Extensions, MIME types, and default icons',
  '浏览应用': 'Browser apps',
  '刷新后仍会保留': 'Kept after refresh',
  '自定义属性': 'Custom properties',
  '文件详情抽屉字段': 'File detail drawer fields',
  '关闭提示': 'Dismiss',
  '重新检测连接': 'Retest connection',
  '关闭预览': 'Close preview',
  '更多': 'More',
  '播放': 'Play',
  '还原': 'Restore',
  '永久删除': 'Delete permanently',
  '清空回收站': 'Empty trash',
  '全选': 'Select all',
  '取消全选': 'Deselect all',
  '全选当前页': 'Select current page',
  '全部': 'All',
  '全部状态': 'All statuses',
  '等待中': 'Pending',
  '校验中': 'Checking',
  '上传中': 'Uploading',
  '已完成': 'Completed',
  '已取消': 'Cancelled',
  '失败': 'Failed',
  '成功': 'Succeeded',
  '处理中': 'Processing',
  '挂起': 'Pending',
  '已提交': 'Submitted',
  '已过期': 'Expired',
  '有效': 'Active',
  '未知文件': 'Unknown file',
  '未知类型': 'Unknown type',
  '未知时间': 'Unknown time',
  '未设置': 'Not set',
  '已开启': 'Enabled',
  '已关闭': 'Disabled',
  '已生效': 'Active',
  '配置生效': 'Config active',
  '后台连接正常': 'Backend connected',
  '后端连接正常': 'Backend connected',
  '已成功连接': 'Connected successfully',
  '当前页签数据已刷新。': 'Current tab data has been refreshed.',
  '配置已成功保存到': 'Config was saved to',
  '后端服务当前不可用': 'Backend service is currently unavailable',
  '当前检测地址': 'Current check URL',
  '检测中...': 'Checking...',

  // View controls
  '视图': 'View',
  '布局': 'Layout',
  '网格': 'Grid',
  '列表': 'List',
  '画廊': 'Gallery',
  '专辑': 'Album',
  '卡片': 'Cards',
  '卡片视图': 'Card view',
  '列表视图': 'List view',
  '显示方式': 'Display mode',
  '排序': 'Sort',
  '排序方式': 'Sort by',
  '最近更新': 'Recently updated',
  '最近分享': 'Recently shared',
  '最新创建': 'Newest first',
  '最新优先': 'Newest first',
  '从近到远': 'Newest first',
  '从远到近': 'Oldest first',
  '名称': 'Name',
  '名称 A-Z': 'Name A-Z',
  '名称排序': 'Sort by name',
  '大小': 'Size',
  '大小排序': 'Sort by size',
  '大小从大到小': 'Largest first',
  '过期时间最近': 'Expiring soon',
  '删除时间从近到远': 'Deleted newest first',
  '删除时间从远到近': 'Deleted oldest first',
  '列设置': 'Column settings',
  '卡片大小': 'Card size',
  '图片尺寸': 'Image size',
  '分页大小': 'Page size',

  // Tables and metadata
  '操作': 'Actions',
  '状态': 'Status',
  '类型': 'Type',
  '关键词': 'Keyword',
  '文件名': 'File name',
  '文件名称': 'File name',
  '分享链接': 'Share link',
  '创建时间': 'Created at',
  '修改时间': 'Updated at',
  '有效期': 'Expiration',
  '下载次数': 'Downloads',
  '过期时间': 'Expires at',
  '原始位置': 'Original location',
  '占用空间': 'Space used',
  '已选择': 'Selected',
  '个对象': 'objects',
  '对象': 'objects',
  '项': 'items',
  '页': 'page',
  '分片': 'chunks',
  '分片并行数': 'chunk concurrency',
  '路并行': 'parallel routes',
  '预分配': 'preallocation',
  '速度': 'Speed',
  '剩余时间': 'Time left',
  '任务 ID': 'Task ID',

  // Admin settings
  '当前面板': 'Current panel',
  '可配置模块': 'Configurable modules',
  '站点信息': 'Site info',
  '会话安全': 'Session security',
  '验证码': 'Captcha',
  '媒体处理': 'Media processing',
  '邮件': 'Mail',
  '队列': 'Queue',
  '外观': 'Appearance',
  '事件': 'Events',
  '服务器': 'Server',
  '统一管理站点信息、会话安全与服务参数': 'Manage site info, session security, and service parameters in one place',
  '品牌预览': 'Brand preview',
  '基础信息': 'Basic info',
  '品牌文案': 'Brand copy',
  '站点名称': 'Site name',
  '站点标语': 'Site tagline',
  '站点描述': 'Site description',
  '使用条款链接': 'Terms link',
  '隐私政策链接': 'Privacy policy link',
  '域名设置': 'Domain settings',
  '访问入口': 'Access endpoints',
  '新增备选域名': 'Add alternate domain',
  '主站 URL': 'Primary URL',
  '主站地址': 'Primary site URL',
  '备选域名': 'Alternate domains',
  '品牌资源': 'Brand assets',
  '图标与 Logo': 'Icons and logo',
  '浅色 Logo': 'Light logo',
  '深色 Logo': 'Dark logo',
  '192px 图标': '192px icon',
  '自定义注入代码': 'Custom injection code',
  '客户端引导': 'Client guidance',
  '移动端与桌面端': 'Mobile and desktop',
  '移动端引导': 'Mobile guidance',
  '桌面端引导': 'Desktop guidance',
  '移动端反馈 URL': 'Mobile feedback URL',
  '桌面端社区 URL': 'Desktop community URL',
  '注册与登录': 'Registration and sign-in',
  '允许新用户注册': 'Allow new user registration',
  '邮件激活': 'Email activation',
  '邮件发信设置': 'Mail delivery settings',
  '使用通行密钥登录': 'Allow passkey sign-in',
  '默认用户组': 'Default user group',
  '头像': 'Avatar',
  '点击条目即可修改': 'Click an item to edit it',
  '头像存储路径': 'Avatar storage path',
  '头像文件大小限制': 'Avatar file size limit',
  '图像尺寸 (px)': 'Image size (px)',
  'Gravatar 服务器': 'Gravatar server',
  '服务配置': 'Service config',
  'Captcha供应商': 'Captcha provider',
  'CaptchaType': 'Captcha type',
  '图形Captcha': 'Image captcha',
  '难度等级': 'Difficulty',
  '平衡': 'Balanced',
  '站点 Key': 'Site key',
  '服务端 Secret': 'Server secret',
  'Failed阈值': 'Failure threshold',
  '冷却时间（秒）': 'Cooldown (seconds)',
  '白名单路径': 'Whitelist paths',
  '启用Captcha': 'Enable captcha',
  'Sign up启用Captcha': 'Enable captcha for sign-up',
  'Reset password启用Captcha': 'Enable captcha for password reset',
  '媒体资产引擎': 'Media asset engine',
  '图像视觉引擎': 'Image vision engine',
  '流媒体直觉': 'Streaming preview',
  '质量优先': 'Quality first',
  '兼容性优先': 'Compatibility first',
  '歌曲封面提取': 'Cover extraction',
  '图像预览引擎': 'Image preview engine',
  '流媒体抽帧引擎': 'Streaming frame extraction engine',
  '图像预览': 'Image preview',
  '抽帧预览': 'Frame preview',
  '元数据提取': 'Metadata extraction',
  '文档审阅能力': 'Document review capability',
  '秒开感知': 'Instant preview',
  '画面细节': 'Visual detail',
  '播放稳定': 'Playback stability',
  '抽帧预览起始点': 'Frame preview start',
  '邮件设置': 'Mail settings',
  '管理发件身份、SMTP 投递以及账号激活和密码Reset两类可编辑Mail模板。编辑器已改为更接近代码编辑器的标签pageLayout。':
    'Manage sender identity, SMTP delivery, account activation, and password reset mail templates. The editor now uses a tabbed layout closer to a code editor.',
  'Mail投递': 'Mail delivery',
  'SMTP 设置': 'SMTP settings',
  '发件人名': 'Sender name',
  '发件人邮箱': 'Sender email',
  'SMTP Server': 'SMTP server',
  'SMTP 端口': 'SMTP port',
  'SMTP 认证': 'SMTP authentication',
  'SMTP Users名': 'SMTP username',
  'SMTP 密码': 'SMTP password',
  '辅助面板': 'Side panel',
  '投递优化建议': 'Delivery optimization suggestions',
  '模板数量': 'Template count',
  '右侧保留发送与模板维护要点，方便你在配置 SMTP 和编辑 HTML 模板时同步检查。':
    'The right panel keeps delivery and template maintenance notes so SMTP configuration and HTML template editing can be checked together.',
  '标签式Language切换': 'Tabbed language switching',
  'Language切换位于编辑器顶部，更接近代码编辑器的标签page体验。':
    'Language switching is at the top of the editor, closer to code-editor tabs.',
  '大尺寸代码编辑区': 'Large code editor',
  '参数设置支持真实Save底部Actions仅影响Current tab，不会误改其他模块。':
    'Parameter settings support real saving. Bottom actions affect only the current tab and will not change other modules.',
  '文件系统设置': 'File system settings',
  '统一管理编辑、回收、列表、缓存、传输与安全策略。':
    'Manage editing, trash, lists, cache, transfer, and security policies in one place.',
  '统一管理参数设置、全文搜索、文件图标、文件浏览应用和自定义属性。当前页面直接连接真实后端接口，保存后会立即持久化。':
    'Manage parameters, full-text search, file icons, browser apps, and custom properties. This page connects directly to live backend APIs and persists changes immediately after saving.',
  '连接 Meilisearch / Tika 搜索服务，支持真实重建索引并查看任务结果。':
    'Connect Meilisearch and Tika search services, rebuild indexes, and inspect real task results.',
  '已索引文件数': 'Indexed files',
  '最近任务真实上报': 'Reported by the latest task',
  '分块数': 'Chunks',
  '按切块配置生成': 'Generated by chunk settings',
  '文档数': 'Documents',
  '进入搜索引擎的索引文档': 'Indexed documents sent to the search engine',
  '搜索服务配置': 'Search service config',
  '开启后会向真实服务写入配置，并用于后续重建索引任务。':
    'When enabled, config is written to the live service and used for later index rebuild jobs.',
  '启用全文搜索': 'Enable full-text search',
  '关闭后前台仍可展示搜索入口，但不再使用全文索引。':
    'When disabled, the frontend can still show search entry points but will not use full-text indexes.',
  'AI 搜索增强': 'AI search enhancement',
  '保留语义搜索增强开关，方便后续联动更强体验。':
    'Keeps the semantic search enhancement toggle for future integrations.',
  'Meilisearch 地址': 'Meilisearch URL',
  'Meilisearch API Key': 'Meilisearch API key',
  'Tika 地址': 'Tika URL',
  '最大索引文件大小': 'Maximum indexed file size',
  '真实重建索引任务': 'Live index rebuild job',
  '正在提交重建任务...': 'Submitting rebuild job...',
  '重建索引': 'Rebuild index',
  '索引扩展名': 'Indexed extensions',
  '索引备注': 'Index notes',
  '尚未读取到重建索引任务状态': 'No index rebuild job status has been read yet',
  '搜索引擎服务地址，用于写入配置并驱动真实索引构建。':
    'Search engine service URL used to write config and drive live index builds.',
  '服务端访问 Meilisearch 的授权密钥，建议使用受限 API Key。':
    'Authorization key used by the server to access Meilisearch. A restricted API key is recommended.',
  'Tika 提取服务端点，负责从真实文件中抽取可检索文本。':
    'Tika extraction endpoint used to pull searchable text from real files.',
  '控制每次搜索返回的结果规模，平衡信息密度与响应速度。':
    'Controls search result size to balance information density and response speed.',
  '限制可进入索引流程的文件体积，避免超大文件拖慢搜索集群。':
    'Limits file size entering the indexing flow to avoid slowing the search cluster.',
  '切块越细搜索精度越高，切块越大索引体积越小。':
    'Smaller chunks improve precision; larger chunks reduce index size.',
  '编辑与回收': 'Editing and trash',
  '控制在线编辑、回收站扫描和 Blob 回收节奏。': 'Control online editing, trash scanning, and blob cleanup cadence.',
  '在线编辑最大文件': 'Maximum online edit size',
  'online编辑上限': 'Online edit limit',
  '超出阈值后关闭online编辑入口': 'Disable online editing after the threshold is exceeded',
  '分page模式': 'Pagination mode',
  '最大每page 2000 rows': 'Maximum 2000 rows per page',
  '传输并行数': 'Transfer concurrency',
  '分片Retry上限 5 次': 'Chunk retry limit: 5 times',
  'Master key存储': 'Master key storage',
  'Encryption status 显示': 'Encryption status display',
  '后台按该周期巡检回收站并触发清理任务。': 'The backend checks trash and triggers cleanup jobs on this schedule.',
  '策略编辑器': 'Policy editor',
  '编辑 默认Storage policy': 'Edit default storage policy',
  '最近变更记录': 'Recent changes',
  '最近命中记录': 'Recent hits',
  '修复历史数据': 'Repair historical data',
  '创建、Edit、Delete都会记录Actions者、影响范围和配置快照。':
    'Create, edit, and delete actions record the actor, scope, and config snapshot.',
  '上传、分片、Download、预览和ShareDownload实际命中的策略。':
    'Policies actually hit by uploads, chunks, downloads, previews, and share downloads.',
  '查看差异': 'View diff',
  '回滚到此版本': 'Roll back to this version',
  '已有Files': 'Existing files',
  '新上传': 'New uploads',
  '影响Users': 'Affected users',
  'User groups策略': 'User group policy',
  '历史 Blob 元数据保持不变': 'Historical blob metadata stays unchanged',
  '按User groups绑定实时统计': 'Realtime stats by user group binding',
  '共享文件名、分享人或权限': 'shared file name, owner, or permission',
  '外部客户端接入': 'External client access',
  'WebDAV 账号管理': 'WebDAV account management',
  'WebDAV 账号': 'WebDAV account',
  '没有连接账号': 'No connected accounts',
  '没有找到别人的Share': 'No shares from others found',
  '与我Share': 'Shared with me',

  // Dashboard and admin pages
  '今日上传': 'Uploads today',
  '平均延迟': 'Average latency',
  '活跃用户': 'Active users',
  '文件对象': 'File objects',
  '当前在线用户': 'Online users',
  '节点健康': 'Node health',
  '今日流量趋势': 'Traffic trend today',
  '峰值吞吐': 'Peak throughput',
  '核心指标': 'Core metrics',
  '控制台状态': 'Console status',
  '刷新概览': 'Refresh overview',
  '系统设置': 'System settings',
  '近 5 分钟活跃': 'Active in the last 5 minutes',
  '暂无延迟': 'No latency',
  '运行中': 'Running',
  '存储空间': 'Storage',
  '入站': 'Inbound',
  '出站': 'Outbound',
  '文件总量': 'Total files',
  '当前筛选范围': 'Current filter scope',
  '已选文件': 'Selected files',
  '支持批量管理': 'Batch management supported',
  '本页体积': 'This page size',
  '文件大小合计': 'Total file size',
  '当前筛选下暂无文件': 'No files match the current filters',
  'Blob 总数': 'Total blobs',
  '资产体积': 'Asset size',
  '引用计数': 'Reference count',
  '加密 / 孤儿': 'Encrypted / orphaned',
  '创建者 ID': 'Creator ID',
  '引用状态': 'Reference status',
  '加密状态': 'Encryption status',
  '巡检 Blob': 'Inspect blobs',
  '批量删除孤儿': 'Delete orphaned blobs',
  '重置筛选': 'Reset filters',
  '暂无产物': 'No artifacts',
  '暂无来源': 'No sources',

  // Empty states and messages
  '暂无分享记录': 'No share records',
  '未找到匹配的分享记录': 'No matching share records',
  '暂无任务数据': 'No task data',
  '没有记录': 'No records',
  '回收站为空': 'Trash is empty',
  '删除的文件将会在这里保留 30 天': 'Deleted files are kept here for 30 days',
  '正在加载回收站...': 'Loading trash...',
  '正在加载音乐...': 'Loading music...',
  '这里还没有音乐': 'No music yet',
  '这个音乐暂时无法预览': 'This music file cannot be previewed yet',
};

const phraseDictionary: Record<string, string> = {
  '按文件名搜索': 'Search by file name',
  '筛选音乐名称': 'Filter music names',
  '筛选任务名称、状态或文件类型': 'Filter task name, status, or file type',
  '筛选账号Name、folders、权限或Status': 'Filter account name, folders, permissions, or status',
  '筛选账号名称、目录、权限或状态': 'Filter account name, folder, permission, or status',
  'Filter账号Name、 folders、 权限或Status': 'Filter account name, folders, permissions, or status',
  'FilterShareFile name、 Share人或权限': 'Filter shared file name, owner, or permission',
  '筛选分享文件名称、分享人或权限': 'Filter shared file name, owner, or permission',
  '访问别人给你的Share链接后，可以把它Save为快捷方式，后续会在这里集中管理。':
    'After opening a share link from someone else, you can save it as a shortcut and manage it here later.',
  '访问别人给你的分享链接后，可以把它保存为快捷方式，后续会在这里集中管理。':
    'After opening a share link from someone else, you can save it as a shortcut and manage it here later.',
  '创建 WebDAV 账号后，可以在这里管理挂载folders、 权限、 代理与客户端访问地址。':
    'After creating a WebDAV account, manage mount folders, permissions, proxies, and client access URLs here.',
  '创建 WebDAV 账号后，可以在这里管理挂载目录、权限、代理与客户端访问地址。':
    'After creating a WebDAV account, manage mount folders, permissions, proxies, and client access URLs here.',
  '控制是否展示移动客户端入口、下载说明或帮助链接。':
    'Controls whether mobile client entry points, download notes, or help links are shown.',
  '控制是否展示桌面客户端安装说明或社区入口。':
    'Controls whether desktop install notes or community links are shown.',
  '当前没有备选域名，可按需新增。': 'No alternate domains yet. Add one when needed.',
  '用户注册后的初始用户组。': 'Initial user group for newly registered users.',
  '关闭后，无法再通过前台注册新的用户。': 'When disabled, new users cannot sign up from the public site.',
  '开启后，新用户注册需要点击邮件中的激活链接才能完成。需确认邮件发信设置是否正确，否则激活邮件无法送达。':
    'When enabled, new users must click the activation link in email. Make sure mail delivery settings are correct.',
  '是否允许用户使用绑定的硬件认证设备登录，比如：人脸、指纹或 USB 密钥；站点必须启用 HTTPS 才能使用。':
    'Allow users to sign in with hardware authenticators such as face, fingerprint, or USB keys. HTTPS must be enabled.',
  '全格式图像超清预览。让 RAW 原片、封面图与复杂设计素材，在点开之前就已足够动人。':
    'Ultra-clear previews for all image formats. RAW photos, covers, and design assets look good before opening.',
  'Videos秒开与智能预览。精彩不需要先Download，情绪也不必等待缓冲。':
    'Instant video startup and smart previews. Highlights do not need to download first, and playback does not wait on buffering.',
  '保留摄影作品的质感': 'Preserve the texture of photographic work',
  '让歌单不再只是Files名': 'Make playlists more than file names',
  'Large依然清晰、有层次': 'Large files remain clear and layered',
  '显示在收件箱中的品牌Name。': 'Brand name shown in the recipient inbox.',
  '这里已经修正为真实发件邮箱。': 'This has been corrected to the real sender email.',
  'QQ 邮箱默认 SMTP Server为 smtp.qq.com。': 'QQ Mail uses smtp.qq.com as the default SMTP server.',
  '常用端口为 587。': 'The common port is 587.',
  '实际发件邮箱：': 'Actual sender email:',
  '限制浏览器内直接编辑的文件大小，避免拖慢前端体验。':
    'Limits the file size for direct browser editing to avoid slowing down the frontend.',
  '修复历史数据': 'Repair historical data',
};

const tokenDictionary: Record<string, string> = {
  账号: 'account',
  名称: 'name',
  目录: 'folder',
  文件夹: 'folder',
  权限: 'permission',
  状态: 'status',
  分享人: 'owner',
  别人: 'others',
  链接: 'link',
  快捷方式: 'shortcut',
  后续: 'later',
  集中: 'centralized',
  管理: 'manage',
  创建: 'create',
  新账号: 'new account',
  客户端: 'client',
  接入: 'access',
  外部: 'external',
  挂载: 'mount',
  代理: 'proxy',
  地址: 'address',
  媒体: 'media',
  资产: 'assets',
  引擎: 'engine',
  配置: 'config',
  供应商: 'provider',
  难度: 'difficulty',
  等级: 'level',
  白名单: 'whitelist',
  路径: 'path',
  阈值: 'threshold',
  冷却: 'cooldown',
  时间: 'time',
  秒: 'seconds',
  启用: 'enable',
  图形: 'image',
  重置: 'reset',
  密码: 'password',
  流程: 'flow',
  提高: 'improve',
  账户: 'account',
  恢复: 'recovery',
  安全性: 'security',
  减少: 'reduce',
  恶意: 'malicious',
  垃圾账户: 'spam accounts',
  图像: 'image',
  视觉: 'vision',
  流媒体: 'streaming media',
  审阅: 'review',
  能力: 'capability',
  封面: 'cover',
  提取: 'extraction',
  质量: 'quality',
  兼容性: 'compatibility',
  优先: 'priority',
  平衡: 'balanced',
  预览: 'preview',
  元数据: 'metadata',
  原片: 'original',
  还原: 'restore',
  摄影: 'photography',
  作品: 'work',
  质感: 'texture',
  歌单: 'playlist',
  清晰: 'clear',
  层次: 'depth',
  邮箱: 'email',
  发件人: 'sender',
  投递: 'delivery',
  模板: 'template',
  数量: 'count',
  建议: 'suggestions',
  辅助: 'side',
  面板: 'panel',
  编辑器: 'editor',
  标签: 'tabs',
  切换: 'switch',
  代码: 'code',
  区域: 'area',
  真实: 'real',
  底部: 'bottom',
  仅影响: 'only affects',
  当前: 'current',
  不会: 'will not',
  误改: 'accidentally change',
  其他: 'other',
  模块: 'modules',
  编辑: 'edit',
  回收: 'recycle',
  缓存: 'cache',
  传输: 'transfer',
  安全: 'security',
  策略: 'policy',
  控制: 'control',
  在线: 'online',
  扫描: 'scan',
  节奏: 'cadence',
  最大: 'maximum',
  上限: 'limit',
  超出: 'exceeds',
  后: 'after',
  关闭: 'disable',
  入口: 'entry',
  历史: 'history',
  数据: 'data',
  保持: 'keep',
  不变: 'unchanged',
  绑定: 'binding',
  实时: 'realtime',
  统计: 'stats',
  变更: 'changes',
  记录: 'records',
  命中: 'hits',
  查看: 'view',
  差异: 'diff',
  回滚: 'rollback',
  版本: 'version',
  影响: 'affected',
  已有: 'existing',
  新上传: 'new uploads',
  最近: 'recent',
  默认: 'default',
  存储: 'storage',
  新建: 'new',
  没有: 'no',
  找到: 'found',
  连接: 'connection',
  文件: 'file',
  系统: 'system',
  参数: 'parameters',
};

const textPatterns: Array<[RegExp, (match: RegExpMatchArray) => string]> = [
  [/^已选择\s+(\d+)\s+个离线任务/, (match) => `${match[1]} offline tasks selected`],
  [/^已选择\s+(\d+)\s+个对象$/, (match) => `${match[1]} objects selected`],
  [/^已选择\s+(\d+)\s+个文件$/, (match) => `${match[1]} files selected`],
  [/^已选择\s+(\d+)\s+首/, (match) => `${match[1]} tracks selected`],
  [/^(\d+)\s+项$/, (match) => `${match[1]} items`],
  [/^(\d+)\s+天后$/, (match) => `${match[1]} days later`],
  [/^剩余\s+(\d+)\s+天$/, (match) => `${match[1]} days left`],
  [/^第\s*(\d+)\s*页\s*\/\s*共\s*(\d+)\s*页$/, (match) => `Page ${match[1]} of ${match[2]}`],
  [/^每页\s*(\d+)\s*条$/, (match) => `${match[1]} per page`],
  [/^(\d+)\s*\/\s*(\d+)\s*在线$/, (match) => `${match[1]} / ${match[2]} online`],
  [/^(\d+)\s*运行中$/, (match) => `${match[1]} running`],
  [/^备选 URL\s*(\d+)$/, (match) => `Alternate URL ${match[1]}`],
  [/^最近轮转\s*(.+)$/, (match) => `Last rotated ${translateText(match[1]) ?? match[1]}`],
  [/^正在保存(.+)$/, (match) => `Saving ${translateText(match[1]) ?? match[1]}`],
  [/^正在加载(.+)$/, (match) => `Loading ${translateText(match[1]) ?? match[1]}`],
  [/^(.+)支持真实保存$/, (match) => `${translateText(match[1]) ?? match[1]} supports live saving`],
  [/^(.+)有未保存修改$/, (match) => `${translateText(match[1]) ?? match[1]} has unsaved changes`],
  [/^(.+)已与当前配置同步$/, (match) => `${translateText(match[1]) ?? match[1]} is synced with the current config`],
  [/^“(.+)”暂未接入(.+)逻辑$/, (match) => `${translateText(match[1]) ?? match[1]} does not support ${translateText(match[2]) ?? match[2]} yet`],
  [/^当前底部操作条将作用于“(.+)”面板。$/, (match) => `The bottom action bar applies to the ${translateText(match[1]) ?? match[1]} panel.`],
  [/^编辑队列设置\s*-\s*(.+)$/, (match) => `Edit queue settings - ${translateText(match[1]) ?? match[1]}`],
  [/^确定要删除选中的\s+(\d+)\s+条分享记录吗？$/, (match) => `Delete ${match[1]} selected share records?`],
  [/^确定永久删除选中的\s+(\d+)\s+个文件吗？$/, (match) => `Permanently delete ${match[1]} selected files?`],
  [/^已删除\s+(\d+)\s+条分享记录$/, (match) => `Deleted ${match[1]} share records`],
  [/^已永久删除\s+(\d+)\s+个文件$/, (match) => `Permanently deleted ${match[1]} files`],
  [/^已还原\s+(\d+)\s+个文件$/, (match) => `Restored ${match[1]} files`],
  [/^确定删除\s+(\d+)\s+首音乐吗？此操作不可撤销。$/, (match) => `Delete ${match[1]} tracks? This action cannot be undone.`],
  [/^已添加\s+(\d+)\s+个音乐到上传队列$/, (match) => `Added ${match[1]} tracks to the upload queue`],
];

function normalizeText(value: string) {
  return value.replace(/\s+/g, ' ').trim();
}

function applyFullReplacements(value: string) {
  let replaced = value;
  const entries = Object.entries({ ...phraseDictionary, ...exactDictionary })
    .filter(([source]) => source.length >= 2 && replaced.includes(source))
    .sort((a, b) => b[0].length - a[0].length);
  for (const [source, target] of entries) {
    replaced = replaced.split(source).join(target);
  }
  return replaced;
}

function translateHanRun(value: string) {
  const exact = exactDictionary[value] || phraseDictionary[value] || tokenDictionary[value];
  if (exact) return exact;

  let translated = value;
  const entries = Object.entries(tokenDictionary)
    .filter(([source]) => source.length >= 2 && translated.includes(source))
    .sort((a, b) => b[0].length - a[0].length);
  for (const [source, target] of entries) {
    translated = translated.split(source).join(` ${target} `);
  }

  if (!hasHan.test(translated)) return translated.replace(/\s+/g, ' ').trim();
  return '';
}

function cleanupEnglish(value: string) {
  return value
    .replace(/\s*、\s*/g, ', ')
    .replace(/\s*，\s*/g, ', ')
    .replace(/\s*。\s*/g, '. ')
    .replace(/\s*；\s*/g, '; ')
    .replace(/\s*：\s*/g, ': ')
    .replace(/\s*（\s*/g, ' (')
    .replace(/\s*）\s*/g, ') ')
    .replace(/\s*？\s*/g, '? ')
    .replace(/\s*“\s*/g, '"')
    .replace(/\s*”\s*/g, '"')
    .replace(/\s+/g, ' ')
    .replace(/\s+([,.;:?!])/g, '$1')
    .replace(/([([])\s+/g, '$1')
    .trim();
}

function translateText(value: string): string | null {
  const trimmed = normalizeText(value);
  if (!trimmed || !hasHan.test(trimmed)) return null;

  const exact = exactDictionary[trimmed] || phraseDictionary[trimmed];
  if (exact) return value.replace(value.trim(), exact);

  for (const [pattern, formatter] of textPatterns) {
    const match = trimmed.match(pattern);
    if (match) return value.replace(value.trim(), formatter(match));
  }

  let replaced = applyFullReplacements(trimmed);
  if (hasHan.test(replaced)) {
    replaced = replaced.replace(hanRun, (source) => translateHanRun(source));
  }
  replaced = cleanupEnglish(replaced);

  return replaced && replaced !== trimmed ? value.replace(value.trim(), replaced) : null;
}

function localizeTextNode(node: Text, english: boolean) {
  if (!english) {
    const original = textOriginals.get(node);
    if (original !== undefined) node.nodeValue = original;
    return;
  }
  const current = node.nodeValue || '';
  if (!textOriginals.has(node)) textOriginals.set(node, current);
  const translated = translateText(textOriginals.get(node) || current);
  if (translated && node.nodeValue !== translated) node.nodeValue = translated;
}

function localizeAttributes(element: Element, english: boolean) {
  const originals = attrOriginals.get(element) || new Map<AttrName, string>();
  (['placeholder', 'title', 'aria-label'] as AttrName[]).forEach((attr) => {
    const value = element.getAttribute(attr);
    if (value === null) return;
    if (!originals.has(attr)) originals.set(attr, value);
    if (english) {
      const translated = translateText(originals.get(attr) || value);
      if (translated && value !== translated) element.setAttribute(attr, translated);
    } else {
      element.setAttribute(attr, originals.get(attr) || value);
    }
  });
  if (originals.size) attrOriginals.set(element, originals);
}

function shouldSkip(node: Node): boolean {
  const parent = node.parentElement;
  return Boolean(parent?.closest('script, style, code, pre, textarea, [contenteditable="true"]'));
}

export function applyDomLanguage(root: ParentNode = document) {
  const english = currentLanguage.value === 'en-US';
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT, {
    acceptNode(node) {
      return shouldSkip(node) ? NodeFilter.FILTER_REJECT : NodeFilter.FILTER_ACCEPT;
    },
  });

  let node: Node | null = walker.currentNode;
  while (node) {
    if (node.nodeType === Node.TEXT_NODE) {
      localizeTextNode(node as Text, english);
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      localizeAttributes(node as Element, english);
    }
    node = walker.nextNode();
  }
}

export function installDomLocalizer() {
  let scheduled = false;
  const schedule = () => {
    if (scheduled) return;
    scheduled = true;
    window.requestAnimationFrame(() => {
      scheduled = false;
      applyDomLanguage(document);
      window.requestAnimationFrame(() => applyDomLanguage(document));
    });
  };

  watch(currentLanguage, schedule, { immediate: true });

  const observer = new MutationObserver(schedule);
  observer.observe(document.body, {
    childList: true,
    subtree: true,
    characterData: true,
    attributes: true,
    attributeFilter: ['placeholder', 'title', 'aria-label'],
  });

  window.addEventListener('DOMContentLoaded', schedule);
  window.addEventListener('load', schedule);
}
