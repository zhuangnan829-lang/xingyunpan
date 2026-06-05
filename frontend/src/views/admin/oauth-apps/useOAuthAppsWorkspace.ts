import { computed, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { createOAuthApp, deleteOAuthApp, listOAuthApps, updateOAuthAppStatus } from '@/api/admin-oauth-apps';
import { oauthAppsSeed, scopeOptions } from './data';
import type { OAuthApp, OAuthAppDraft, OAuthMetric } from './types';

const createDraft = (): OAuthAppDraft => ({
  name: '',
  description: '',
  redirectUri: '',
  scopes: ['profile', 'email', 'openid'],
});

export function useOAuthAppsWorkspace() {
  const apps = ref<OAuthApp[]>([]);
  const loading = ref(false);
  const keyword = ref('');
  const status = ref<'all' | 'enabled' | 'disabled'>('all');
  const page = ref(1);
  const pageSize = ref(11);
  const createVisible = ref(false);
  const draft = reactive<OAuthAppDraft>(createDraft());

  const filteredApps = computed(() => {
    const query = keyword.value.trim().toLowerCase();
    return apps.value.filter((app) => {
      const matchKeyword =
        !query ||
        app.name.toLowerCase().includes(query) ||
        app.clientId.toLowerCase().includes(query) ||
        app.scopes.some((scope) => scope.toLowerCase().includes(query));
      const matchStatus =
        status.value === 'all' ||
        (status.value === 'enabled' && app.enabled) ||
        (status.value === 'disabled' && !app.enabled);
      return matchKeyword && matchStatus;
    });
  });

  const pagedApps = computed(() => {
    const start = (page.value - 1) * pageSize.value;
    return filteredApps.value.slice(start, start + pageSize.value);
  });

  const metrics = computed<OAuthMetric[]>(() => {
    const enabled = apps.value.filter((app) => app.enabled).length;
    const system = apps.value.filter((app) => app.isSystem).length;
    const scopes = new Set(apps.value.flatMap((app) => app.scopes)).size;
    return [
      { label: '应用总数', value: String(apps.value.length), detail: `${enabled} 个已启用`, tone: 'sky' },
      { label: '系统应用', value: String(system), detail: '官方客户端与内置服务', tone: 'mint' },
      { label: '授权范围', value: String(scopes), detail: '当前已使用 scope', tone: 'violet' },
      { label: '回调地址', value: String(apps.value.reduce((sum, app) => sum + app.redirectUris.length, 0)), detail: '已登记跳转入口', tone: 'coral' },
    ];
  });

  const resetDraft = () => {
    Object.assign(draft, createDraft());
  };

  const loadApps = async () => {
    loading.value = true;
    try {
      const data = await listOAuthApps({ page: 1, page_size: 100, keyword: keyword.value, status: status.value });
      apps.value = mergeSystemFallbackApps(data.list);
      page.value = data.page || 1;
    } catch (error) {
      apps.value = mergeSystemFallbackApps([]);
      ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用列表加载失败');
    } finally {
      loading.value = false;
    }
  };

  const mergeSystemFallbackApps = (items: OAuthApp[]) => {
    const byId = new Map(items.map((item) => [item.id, item]));
    for (const systemApp of oauthAppsSeed) {
      if (!byId.has(systemApp.id)) {
        byId.set(systemApp.id, cloneApp(systemApp));
      }
    }
    return Array.from(byId.values()).sort((a, b) => Number(b.isSystem) - Number(a.isSystem));
  };

  const cloneApp = (app: OAuthApp): OAuthApp => ({
    ...app,
    redirectUris: [...app.redirectUris],
    scopes: [...app.scopes],
    permissions: app.permissions.map((permission) => ({ ...permission })),
  });

  const refresh = async () => {
    await loadApps();
    ElMessage.success('OAuth 应用列表已刷新');
  };

  const openCreate = () => {
    resetDraft();
    createVisible.value = true;
  };

  const createApp = async () => {
    if (!draft.name.trim()) {
      ElMessage.warning('请填写应用名称');
      return;
    }
    if (!draft.redirectUri.trim()) {
      ElMessage.warning('请填写回调地址');
      return;
    }

    const now = new Date().toISOString();
    try {
      const created = await createOAuthApp({
        id: `custom-${Date.now()}`,
        name: draft.name.trim(),
        description: draft.description.trim() || '自定义 OAuth 应用。',
        appName: draft.name.trim(),
        iconPath: '/static/img/xingyunpan_oauth.svg',
        clientId: `xyp_custom_${Math.random().toString(16).slice(2, 10)}`,
        clientSecret: '',
        redirectUris: [draft.redirectUri.trim()],
        scopes: [...draft.scopes],
        isSystem: false,
        enabled: true,
        createdAt: now,
        updatedAt: now,
        tokenTtl: '7 天',
        refreshTokenTtlSeconds: 604800,
        permissions: [],
      });
      apps.value.unshift(created);
      createVisible.value = false;
      page.value = 1;
      ElMessage.success('OAuth 应用已创建');
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用创建失败');
    }
  };

  const toggleApp = async (app: OAuthApp) => {
    try {
      const updated = await updateOAuthAppStatus(app.id, !app.enabled);
      Object.assign(app, updated);
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用状态更新失败');
    }
  };

  const deleteApp = async (app: OAuthApp) => {
    if (app.isSystem) {
      ElMessage.warning('系统应用不能删除');
      return;
    }
    try {
      await deleteOAuthApp(app.id);
      apps.value = apps.value.filter((item) => item.id !== app.id);
      ElMessage.success('OAuth 应用已删除');
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : 'OAuth 应用删除失败');
    }
  };

  const changePage = (value: number) => {
    page.value = value;
  };

  const changePageSize = (value: number) => {
    pageSize.value = value;
    page.value = 1;
  };

  return {
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
  };
}
