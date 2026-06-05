package service

type EventSettingItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type EventSettingCategory struct {
	Key         string             `json:"key"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Items       []EventSettingItem `json:"items"`
}

func defaultEventCategories() []EventSettingCategory {
	return []EventSettingCategory{
		{Key: "system", Title: "系统事件", Description: "与系统操作和状态相关的事件。", Items: []EventSettingItem{
			{Key: "system.server.started", Label: "服务器启动"},
		}},
		{Key: "user", Title: "用户事件", Description: "与用户账号、认证和配置文件更改相关的事件。", Items: []EventSettingItem{
			{Key: "user.registered", Label: "用户注册"},
			{Key: "user.activated", Label: "用户激活"},
			{Key: "user.login", Label: "用户登录"},
			{Key: "user.login.failed", Label: "登录失败"},
			{Key: "user.token.refreshed", Label: "令牌刷新"},
			{Key: "user.status.changed", Label: "用户状态更改"},
			{Key: "user.quota.exceeded", Label: "超出配额通知"},
			{Key: "user.nickname.changed", Label: "更改昵称"},
			{Key: "user.avatar.changed", Label: "更改头像"},
			{Key: "user.password.changed", Label: "更改密码"},
			{Key: "user.2fa.enabled", Label: "启用 2FA"},
			{Key: "user.2fa.disabled", Label: "禁用 2FA"},
			{Key: "user.passkey.added", Label: "添加通行密钥"},
			{Key: "user.passkey.removed", Label: "移除通行密钥"},
			{Key: "user.external.linked", Label: "链接外部账号"},
			{Key: "user.external.unlinked", Label: "取消链接外部账号"},
			{Key: "user.abuse.reported", Label: "举报滥用"},
			{Key: "oauth.app.authorized", Label: "OAuth 应用授权"},
			{Key: "oauth.token.exchanged", Label: "OAuth 令牌交换"},
			{Key: "oauth.app.revoked", Label: "OAuth 应用授权撤销"},
		}},
		{Key: "file", Title: "文件事件", Description: "与文件操作相关的事件，如上传、下载和修改。", Items: []EventSettingItem{
			{Key: "file.created", Label: "文件创建"},
			{Key: "file.imported", Label: "外部文件导入"},
			{Key: "file.renamed", Label: "文件重命名"},
			{Key: "file.permission.changed", Label: "权限更改"},
			{Key: "file.uploaded", Label: "文件上传或更新"},
			{Key: "file.downloaded", Label: "文件下载"},
			{Key: "file.source.copied", Label: "复制来源"},
			{Key: "file.copied", Label: "复制到"},
			{Key: "file.moved", Label: "移动到"},
			{Key: "file.deleted", Label: "文件删除"},
			{Key: "file.trashed", Label: "移动到回收站"},
			{Key: "file.metadata.updated", Label: "元数据更新"},
			{Key: "file.direct-link.created", Label: "获取直链"},
			{Key: "file.direct-link.deleted", Label: "删除直链"},
			{Key: "file.view-settings.changed", Label: "更改视图设置"},
		}},
		{Key: "share", Title: "分享事件", Description: "与文件分享和链接访问相关的事件。", Items: []EventSettingItem{
			{Key: "share.created", Label: "分享创建"},
			{Key: "share.link.viewed", Label: "分享链接查看"},
			{Key: "share.updated", Label: "分享编辑"},
			{Key: "share.deleted", Label: "分享删除"},
		}},
		{Key: "version", Title: "版本事件", Description: "与文件版本管理相关的事件。", Items: []EventSettingItem{
			{Key: "version.current.set", Label: "设置当前版本"},
			{Key: "version.deleted", Label: "删除版本"},
		}},
		{Key: "media", Title: "媒体事件", Description: "与媒体文件处理相关的事件，如缩略图生成。", Items: []EventSettingItem{
			{Key: "media.thumbnail.generated", Label: "缩略图生成"},
			{Key: "media.live-photo.uploaded", Label: "上传 Live Photo"},
		}},
		{Key: "file-system", Title: "文件系统事件", Description: "与文件系统操作相关的事件，如挂载和归档处理。", Items: []EventSettingItem{
			{Key: "filesystem.mounted", Label: "挂载"},
			{Key: "filesystem.policy.transferred", Label: "转移存储策略"},
			{Key: "filesystem.archive.created", Label: "创建归档"},
			{Key: "filesystem.archive.extracted", Label: "解压归档"},
		}},
		{Key: "webdav", Title: "WebDAV 事件", Description: "与 WebDAV 账号管理和访问相关的事件。", Items: []EventSettingItem{
			{Key: "webdav.login.failed", Label: "WebDAV 登录失败"},
			{Key: "webdav.account.created", Label: "WebDAV 账号创建"},
			{Key: "webdav.account.updated", Label: "WebDAV 账号更新"},
			{Key: "webdav.account.deleted", Label: "WebDAV 账号删除"},
		}},
		{Key: "payment", Title: "支付事件", Description: "与支付交易和处理相关的事件。", Items: []EventSettingItem{
			{Key: "payment.created", Label: "支付创建"},
			{Key: "payment.points.changed", Label: "积分更改"},
			{Key: "payment.completed", Label: "支付完成"},
			{Key: "payment.order.fulfilled", Label: "履行订单"},
			{Key: "payment.order.failed", Label: "履行订单失败"},
			{Key: "payment.storage.expanded", Label: "存储扩容"},
			{Key: "payment.group.changed", Label: "用户组更改"},
			{Key: "payment.subscription.canceled", Label: "取消订阅"},
			{Key: "payment.gift-code.redeemed", Label: "兑换礼品码"},
		}},
		{Key: "email", Title: "Email 事件", Description: "与邮件发送和通知相关的事件。", Items: []EventSettingItem{
			{Key: "email.sent", Label: "邮件发送"},
		}},
	}
}
