export type EventItem = {
  key: string;
  label: string;
};

export type EventCategory = {
  key: string;
  title: string;
  description: string;
  items: EventItem[];
};

export const eventCategories: EventCategory[] = [
  {
    key: 'system',
    title: '系统事件',
    description: '与系统操作和状态相关的事件。',
    items: [
      { key: 'system.server.started', label: '服务器启动' },
    ],
  },
  {
    key: 'user',
    title: '用户事件',
    description: '与用户账号、认证和配置文件更改相关的事件。',
    items: [
      { key: 'user.registered', label: '用户注册' },
      { key: 'user.activated', label: '用户激活' },
      { key: 'user.login', label: '用户登录' },
      { key: 'user.login.failed', label: '登录失败' },
      { key: 'user.token.refreshed', label: '令牌刷新' },
      { key: 'user.status.changed', label: '用户状态更改' },
      { key: 'user.quota.exceeded', label: '超出配额通知' },
      { key: 'user.nickname.changed', label: '更改昵称' },
      { key: 'user.avatar.changed', label: '更改头像' },
      { key: 'user.password.changed', label: '更改密码' },
      { key: 'user.2fa.enabled', label: '启用 2FA' },
      { key: 'user.2fa.disabled', label: '禁用 2FA' },
      { key: 'user.passkey.added', label: '添加通行密钥' },
      { key: 'user.passkey.removed', label: '移除通行密钥' },
      { key: 'user.external.linked', label: '链接外部账号' },
      { key: 'user.external.unlinked', label: '取消链接外部账号' },
      { key: 'user.abuse.reported', label: '举报滥用' },
      { key: 'oauth.app.authorized', label: 'OAuth 应用授权' },
      { key: 'oauth.token.exchanged', label: 'OAuth 令牌交换' },
      { key: 'oauth.app.revoked', label: 'OAuth 应用授权撤销' },
    ],
  },
  {
    key: 'file',
    title: '文件事件',
    description: '与文件操作相关的事件，如上传、下载和修改。',
    items: [
      { key: 'file.created', label: '文件创建' },
      { key: 'file.imported', label: '外部文件导入' },
      { key: 'file.renamed', label: '文件重命名' },
      { key: 'file.permission.changed', label: '权限更改' },
      { key: 'file.uploaded', label: '文件上传或更新' },
      { key: 'file.downloaded', label: '文件下载' },
      { key: 'file.source.copied', label: '复制来源' },
      { key: 'file.copied', label: '复制到' },
      { key: 'file.moved', label: '移动到' },
      { key: 'file.deleted', label: '文件删除' },
      { key: 'file.trashed', label: '移动到回收站' },
      { key: 'file.metadata.updated', label: '元数据更新' },
      { key: 'file.direct-link.created', label: '获取直链' },
      { key: 'file.direct-link.deleted', label: '删除直链' },
      { key: 'file.view-settings.changed', label: '更改视图设置' },
    ],
  },
  {
    key: 'share',
    title: '分享事件',
    description: '与文件分享和链接访问相关的事件。',
    items: [
      { key: 'share.created', label: '分享创建' },
      { key: 'share.link.viewed', label: '分享链接查看' },
      { key: 'share.updated', label: '分享编辑' },
      { key: 'share.deleted', label: '分享删除' },
    ],
  },
  {
    key: 'version',
    title: '版本事件',
    description: '与文件版本管理相关的事件。',
    items: [
      { key: 'version.current.set', label: '设置当前版本' },
      { key: 'version.deleted', label: '删除版本' },
    ],
  },
  {
    key: 'media',
    title: '媒体事件',
    description: '与媒体文件处理相关的事件，如缩略图生成。',
    items: [
      { key: 'media.thumbnail.generated', label: '缩略图生成' },
      { key: 'media.live-photo.uploaded', label: '上传 Live Photo' },
    ],
  },
  {
    key: 'file-system',
    title: '文件系统事件',
    description: '与文件系统操作相关的事件，如挂载和归档处理。',
    items: [
      { key: 'filesystem.mounted', label: '挂载' },
      { key: 'filesystem.policy.transferred', label: '转移存储策略' },
      { key: 'filesystem.archive.created', label: '创建归档' },
      { key: 'filesystem.archive.extracted', label: '解压归档' },
    ],
  },
  {
    key: 'webdav',
    title: 'WebDAV 事件',
    description: '与 WebDAV 账号管理和访问相关的事件。',
    items: [
      { key: 'webdav.login.failed', label: 'WebDAV 登录失败' },
      { key: 'webdav.account.created', label: 'WebDAV 账号创建' },
      { key: 'webdav.account.updated', label: 'WebDAV 账号更新' },
      { key: 'webdav.account.deleted', label: 'WebDAV 账号删除' },
    ],
  },
  {
    key: 'payment',
    title: '支付事件',
    description: '与支付交易和处理相关的事件。',
    items: [
      { key: 'payment.created', label: '支付创建' },
      { key: 'payment.points.changed', label: '积分更改' },
      { key: 'payment.completed', label: '支付完成' },
      { key: 'payment.order.fulfilled', label: '履行订单' },
      { key: 'payment.order.failed', label: '履行订单失败' },
      { key: 'payment.storage.expanded', label: '存储扩容' },
      { key: 'payment.group.changed', label: '用户组更改' },
      { key: 'payment.subscription.canceled', label: '取消订阅' },
      { key: 'payment.gift-code.redeemed', label: '兑换礼品码' },
    ],
  },
  {
    key: 'email',
    title: 'Email 事件',
    description: '与邮件发送和通知相关的事件。',
    items: [
      { key: 'email.sent', label: '邮件发送' },
    ],
  },
];

export function createDefaultEventState(): Record<string, boolean> {
  return Object.fromEntries(eventCategories.flatMap((category) => category.items.map((item) => [item.key, false])));
}
