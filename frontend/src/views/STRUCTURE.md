# Frontend View Structure

## auth
- `auth/`
  - `Login.vue`
  - `Register.vue`
  - `ForgotPassword.vue`

## drive
- `drive/my-files`
- `drive/images`
- `drive/videos`
- `drive/music`
- `drive/documents`
- `drive/shared-with-me`
- `drive/my-shares`
- `drive/transfer-tasks`
- `drive/background-jobs`
- `drive/offline-downloads`
- `drive/storage-overview`

## admin
- `admin/dashboard`
- `admin/settings`
- `admin/file-system`
- `admin/storage-policy`
- `admin/nodes`
- `admin/user-groups`
- `admin/users`
- `admin/files`
- `admin/file-blobs`
- `admin/shares`
- `admin/background-jobs`
- `admin/orders`
- `admin/events`
- `admin/abuse-reports`
- `admin/oauth-apps`

## screenshot mapping
- 图 1 左侧主导航:
  - 我的文件 -> `drive/my-files`
  - 图片 -> `drive/images`
  - 视频 -> `drive/videos`
  - 音乐 -> `drive/music`
  - 文档 -> `drive/documents`
  - 与我共享 -> `drive/shared-with-me`
  - 回收站 -> 可继续复用现有 `RecycleBin.vue`，后续也可迁入 `drive` 目录
  - 我的分享 -> `drive/my-shares`
  - 连接与挂载 -> `drive/transfer-tasks`
  - 后台任务 -> `drive/background-jobs`
  - 离线下载 -> `drive/offline-downloads`
  - 管理面板 -> `admin/dashboard`
- 图 2 管理面板:
  - 面板首页 -> `admin/dashboard`
  - 参数设置 -> `admin/settings`
  - 文件系统 -> `admin/file-system`
  - 存储策略 -> `admin/storage-policy`
  - 节点 -> `admin/nodes`
  - 用户组 -> `admin/user-groups`
  - 用户 -> `admin/users`
  - 文件 -> `admin/files`
  - 文件 Blob -> `admin/file-blobs`
  - 分享 -> `admin/shares`
  - 后台任务 -> `admin/background-jobs`
  - 订单 -> `admin/orders`
  - 事件 -> `admin/events`
  - 滥用举报 -> `admin/abuse-reports`
  - OAuth 应用 -> `admin/oauth-apps`
