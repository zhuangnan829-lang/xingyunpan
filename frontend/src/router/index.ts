import { createRouter, createWebHistory } from 'vue-router';
import { isAuthenticated } from '@/utils/auth';

const PlaceholderPage = () => import('@/views/console/PlaceholderPage.vue');
const DriveImagesPage = () => import('@/views/drive/images/index.vue');
const DriveVideosPage = () => import('@/views/drive/videos/index.vue');
const DriveMusicPage = () => import('@/views/drive/music/index.vue');
const DriveDocumentsPage = () => import('@/views/drive/documents/index.vue');
const DriveSharedWithMePage = () => import('@/views/drive/shared-with-me/index.vue');
const DriveMySharesPage = () => import('@/views/drive/my-shares/index.vue');
const DriveTransferTasksPage = () => import('@/views/drive/transfer-tasks/index.vue');
const DriveOfflineDownloadsPage = () => import('@/views/drive/offline-downloads/index.vue');
const DriveBackgroundJobsPage = () => import('@/views/drive/background-jobs/index.vue');
const AdminSettingsPage = () => import('@/views/admin/settings/index.vue');
const AdminFileSystemPage = () => import('@/views/admin/file-system/index.vue');
const AdminStoragePolicyPage = () => import('@/views/admin/storage-policy/index.vue');
const AdminNodesPage = () => import('@/views/admin/nodes/index.vue');
const AdminUserGroupsPage = () => import('@/views/admin/user-groups/index.vue');
const AdminUsersPage = () => import('@/views/admin/users/index.vue');
const AdminFilesPage = () => import('@/views/admin/files/index.vue');
const AdminFileBlobsPage = () => import('@/views/admin/file-blobs/index.vue');
const AdminSharesPage = () => import('@/views/admin/shares/index.vue');
const AdminTasksPage = () => import('@/views/admin/tasks/index.vue');
const AdminOAuthAppsPage = () => import('@/views/admin/oauth-apps/index.vue');
const AdminOAuthAppDetailPage = () => import('@/views/admin/oauth-apps/OAuthAppDetailPage.vue');

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { requiresAuth: false, title: '登录' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/Register.vue'),
      meta: { requiresAuth: false, title: '注册' },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/auth/ForgotPassword.vue'),
      meta: { requiresAuth: false, title: '找回密码' },
    },
    {
      path: '/',
      redirect: '/drive/my-files',
    },
    {
      path: '/drive/my-files',
      name: 'drive-my-files',
      component: () => import('@/views/Home.vue'),
      meta: { requiresAuth: true, title: '我的文件' },
    },
    {
      path: '/drive/images',
      name: 'drive-images',
      component: DriveImagesPage,
      meta: { requiresAuth: true, title: '图片' },
    },
    {
      path: '/drive/videos',
      name: 'drive-videos',
      component: DriveVideosPage,
      meta: { requiresAuth: true, title: '视频' },
    },
    {
      path: '/drive/music',
      name: 'drive-music',
      component: DriveMusicPage,
      props: {
        section: 'Drive',
        title: '音乐',
        description: '这里承接音乐分类视图，后续可以接入播放列表、音频预览和歌单管理。',
      },
      meta: { requiresAuth: true, title: '音乐' },
    },
    {
      path: '/drive/documents',
      name: 'drive-documents',
      component: DriveDocumentsPage,
      props: {
        section: 'Drive',
        title: '文档',
        description: '这里承接文档分类视图，后续可以接入最近编辑、标签筛选和列表排序。',
      },
      meta: { requiresAuth: true, title: '文档' },
    },
    {
      path: '/drive/shared-with-me',
      name: 'drive-shared-with-me',
      component: DriveSharedWithMePage,
      props: {
        section: 'Drive',
        title: '与我共享',
        description: '这里承接共享给我的文件，后续可以接入权限说明、协作者和访问记录。',
      },
      meta: { requiresAuth: true, title: '与我共享' },
    },
    {
      path: '/drive/my-shares',
      name: 'drive-my-shares',
      component: DriveMySharesPage,
      props: {
        section: 'Drive',
        title: '我的分享',
        description: '这里承接我的分享管理，后续可以接入分享状态、访问次数和失效时间。',
      },
      meta: { requiresAuth: true, title: '我的分享' },
    },
    {
      path: '/drive/transfer-tasks',
      name: 'drive-transfer-tasks',
      alias: '/connect',
      component: DriveTransferTasksPage,
      props: {
        section: 'Drive',
        title: '连接与挂载',
        description: '这里承接 WebDAV 账号、客户端连接和挂载配置。',
      },
      meta: { requiresAuth: true, title: '连接与挂载' },
    },
    {
      path: '/drive/background-jobs',
      name: 'drive-background-jobs',
      component: DriveBackgroundJobsPage,
      props: {
        section: 'Drive',
        title: '后台任务',
        description: '这里承接后台任务视图，后续可以接入上传、转码、同步和清理任务状态。',
      },
      meta: { requiresAuth: true, title: '后台任务' },
    },
    {
      path: '/drive/offline-downloads',
      name: 'drive-offline-downloads',
      component: DriveOfflineDownloadsPage,
      props: {
        section: 'Drive',
        title: '离线下载',
        description: '这里管理离线下载任务、进度信息、保存位置和失败重试。',
      },
      meta: { requiresAuth: true, title: '离线下载' },
    },
    {
      path: '/drive/storage-overview',
      name: 'drive-storage-overview',
      component: PlaceholderPage,
      props: {
        section: 'Drive',
        title: '存储概览',
        description: '这里承接存储概览，后续可以接入容量趋势、热点目录和空间统计。',
      },
      meta: { requiresAuth: true, title: '存储概览' },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/Profile.vue'),
      meta: { requiresAuth: true, title: '个人资料' },
    },
    {
      path: '/shares',
      name: 'shares',
      component: () => import('@/views/Shares.vue'),
      meta: { requiresAuth: true, title: '分享' },
    },
    {
      path: '/s/:shareToken',
      name: 'share-access',
      component: () => import('@/views/SharePage.vue'),
      meta: { requiresAuth: false, title: '分享访问' },
    },
    {
      path: '/share/:shareId',
      name: 'share',
      component: () => import('@/views/SharePage.vue'),
      meta: { requiresAuth: false, title: '分享详情' },
    },
    {
      path: '/recycle',
      name: 'recycle',
      component: () => import('@/views/RecycleBin.vue'),
      meta: { requiresAuth: true, title: '回收站' },
    },
    {
      path: '/collaborations',
      name: 'collaborations',
      component: () => import('@/views/Collaborations.vue'),
      meta: { requiresAuth: true, title: '协作' },
    },
    {
      path: '/admin/dashboard',
      name: 'admin-dashboard',
      component: () => import('@/views/admin/Dashboard.vue'),
      meta: { requiresAuth: true, title: '面板首页' },
    },
    {
      path: '/admin/settings',
      name: 'admin-settings',
      component: AdminSettingsPage,
      props: {
        section: 'Admin',
        title: '参数设置',
        description: '这里承接系统参数配置、站点设置和基础开关。',
      },
      meta: { requiresAuth: true, title: '参数设置' },
    },
    {
      path: '/admin/file-system',
      name: 'admin-file-system',
      component: AdminFileSystemPage,
      meta: { requiresAuth: true, title: '文件系统' },
    },
    {
      path: '/admin/storage-policy',
      name: 'admin-storage-policy',
      component: AdminStoragePolicyPage,
      meta: { requiresAuth: true, title: '存储策略' },
    },
    {
      path: '/admin/storage-policy/:policyId',
      name: 'admin-storage-policy-edit',
      component: AdminStoragePolicyPage,
      meta: { requiresAuth: true, title: '存储策略' },
    },
    {
      path: '/admin/nodes',
      name: 'admin-nodes',
      component: AdminNodesPage,
      props: {
        section: 'Admin',
        title: '节点',
        description: '这里承接节点管理，后续可以接入健康状态、心跳、负载和可用区信息。',
      },
      meta: { requiresAuth: true, title: '节点' },
    },
    {
      path: '/admin/user-groups',
      name: 'admin-user-groups',
      component: AdminUserGroupsPage,
      props: {
        section: 'Admin',
        title: '用户组',
        description: '这里承接用户组与策略配置，后续可以接入权限、配额和角色模板。',
      },
      meta: { requiresAuth: true, title: '用户组' },
    },
    {
      path: '/admin/nodes/:nodeId',
      name: 'admin-nodes-edit',
      component: AdminNodesPage,
      meta: { requiresAuth: true, title: '鑺傜偣' },
    },
    {
      path: '/admin/users',
      name: 'admin-users',
      component: AdminUsersPage,
      props: {
        section: 'Admin',
        title: '用户',
        description: '这里承接用户管理，后续可以接入用户列表、状态筛选和详情抽屉。',
      },
      meta: { requiresAuth: true, title: '用户' },
    },
    {
      path: '/admin/files',
      name: 'admin-files',
      component: AdminFilesPage,
      props: {
        section: 'Admin',
        title: '文件',
        description: '这里承接文件管理，后续可以接入审计、搜索和批量处理。',
      },
      meta: { requiresAuth: true, title: '文件' },
    },
    {
      path: '/admin/file',
      redirect: '/admin/files',
    },
    {
      path: '/admin/blobs',
      name: 'admin-blobs',
      component: AdminFileBlobsPage,
      props: {
        section: 'Admin',
        title: '文件 Blob',
        description: '这里承接 Blob 管理，后续可以接入对象明细、去重情况和存储位置。',
      },
      meta: { requiresAuth: true, title: '文件 Blob' },
    },
    {
      path: '/admin/blob',
      redirect: '/admin/blobs',
    },
    {
      path: '/admin/shares',
      name: 'admin-shares',
      component: AdminSharesPage,
      meta: { requiresAuth: true, title: '分享' },
    },
    {
      path: '/admin/tasks',
      name: 'admin-tasks',
      component: AdminTasksPage,
      props: {
        section: 'Admin',
        title: '后台任务',
        description: '这里承接后台任务总览，后续可以接入队列、执行日志和任务详情。',
      },
      meta: { requiresAuth: true, title: '后台任务' },
    },
    {
      path: '/admin/oauth',
      name: 'admin-oauth',
      component: AdminOAuthAppsPage,
      props: {
        section: 'Admin',
        title: 'OAuth 应用',
        description: '这里承接 OAuth 应用管理，后续可以接入应用列表、密钥和授权范围。',
      },
      meta: { requiresAuth: true, title: 'OAuth 应用' },
    },
    {
      path: '/admin/oauth/:appId',
      name: 'admin-oauth-detail',
      component: AdminOAuthAppDetailPage,
      meta: { requiresAuth: true, title: 'OAuth 应用配置' },
    },
  ],
});

router.beforeEach((to, _from, next) => {
  const authenticated = isAuthenticated();
  const requiresAuth = to.meta.requiresAuth !== false;
  const allowAuthenticatedPublicAccess = to.name === 'share-access' || to.name === 'share';

  if (requiresAuth && !authenticated) {
    next('/login');
    return;
  }

  if (!requiresAuth && authenticated && !allowAuthenticatedPublicAccess) {
    next('/drive/my-files');
    return;
  }

  document.title = `${String(to.meta.title || '星云盘')} - 星云盘`;
  next();
});

export default router;
