<template>
  <section class="mail-settings">
    <div class="mail-shell">
      <section class="mail-hero glass-card">
        <div>
          <p class="section-tag">星云盘邮件中心</p>
          <h2>邮件设置</h2>
          <p class="section-copy">管理发件身份、SMTP 投递以及账号激活和密码重置两类可编辑邮件模板。编辑器已改为更接近代码编辑器的标签页布局。</p>
        </div>
        <div class="mail-stats">
          <article class="stat-pill"><span>状态</span><strong>{{ form.enabled ? 'SMTP 已启用' : 'SMTP 未启用' }}</strong></article>
          <article class="stat-pill"><span>模板数量</span><strong>{{ templates.length }}</strong></article>
        </div>
      </section>

      <section class="mail-layout">
        <div class="form-stack">
          <article class="settings-card glass-card">
            <div class="card-head">
              <div><p class="section-tag">邮件投递</p><h3>SMTP 设置</h3></div>
              <span class="status-chip"><span class="status-dot"></span>已生效</span>
            </div>
            <div class="field-grid">
              <label v-for="field in primaryFields" :key="field.key" class="field-card">
                <span class="field-title"><el-icon><component :is="field.icon" /></el-icon>{{ field.label }}</span>
                <input v-model="form[field.key]" :type="field.type ?? 'text'" :placeholder="field.placeholder" />
                <small>{{ field.hint }}</small>
              </label>
            </div>
            <section class="auth-card">
              <div class="subsection-head">
                <span class="field-title"><el-icon><Lock /></el-icon>SMTP 认证</span>
                <span class="subsection-note">实际发件邮箱：{{ form.senderEmail }}</span>
              </div>
              <div class="field-grid compact">
                <label class="field-card">
                  <span class="field-title"><el-icon><User /></el-icon>SMTP 用户名</span>
                  <input v-model="form.smtpUsername" type="text" placeholder="3518974413@qq.com" />
                  <small>通常与发件邮箱保持一致。</small>
                </label>
                <label class="field-card">
                  <span class="field-title"><el-icon><Key /></el-icon>SMTP 密码</span>
                  <div class="password-shell">
                    <input v-model="form.smtpPassword" :type="showPassword ? 'text' : 'password'" placeholder="SMTP 授权码" />
                    <button type="button" class="icon-ghost" @click="showPassword = !showPassword"><el-icon><component :is="showPassword ? Hide : View" /></el-icon></button>
                  </div>
                  <small>请填写 SMTP 授权码，而不是邮箱登录密码。</small>
                </label>
              </div>
            </section>
            <div class="field-grid compact">
              <label class="field-card">
                <span class="field-title"><el-icon><Promotion /></el-icon>回复邮箱</span>
                <input v-model="form.replyTo" type="email" placeholder="3518974413@qq.com" />
                <small>收件人直接回复时将优先发送到这里。</small>
              </label>
            </div>
          </article>

          <article class="settings-card glass-card advanced-card">
            <button type="button" class="advanced-toggle" @click="advancedOpen = !advancedOpen">
              <div><p class="section-tag">高级选项</p><h3>高级设置</h3></div>
              <el-icon :class="{ rotated: advancedOpen }"><ArrowDown /></el-icon>
            </button>
            <div v-if="advancedOpen" class="advanced-body">
              <div class="toggle-row">
                <div>
                  <span class="field-title"><el-icon><Connection /></el-icon>强制使用 SSL</span>
                  <small>优先使用加密的 SMTP 连接。</small>
                </div>
                <button type="button" class="ssl-switch" :class="{ active: form.forceSSL }" @click="form.forceSSL = !form.forceSSL"><span /></button>
              </div>
              <label class="field-card inline-card">
                <span class="field-title"><el-icon><Timer /></el-icon>连接超时</span>
                <input v-model="form.connectionTimeout" type="number" min="5" placeholder="30" />
                <small>超时时间，单位为秒。</small>
              </label>
            </div>
          </article>

          <div class="form-actions">
            <button type="button" class="outline-action" :disabled="loading || saving" @click="handleSendTest"><el-icon><Message /></el-icon>发送测试邮件</button>
            <button type="button" class="primary-action" :disabled="loading || saving" @click="save">{{ saving ? '保存中...' : '保存设置' }}</button>
          </div>
        </div>

        <aside class="insight-card glass-card">
          <div class="insight-orb"></div>
          <div class="insight-head">
            <div>
              <p class="section-tag">辅助面板</p>
              <h3>投递优化建议</h3>
              <p class="insight-caption">右侧保留发送与模板维护要点，方便你在配置 SMTP 和编辑 HTML 模板时同步检查。</p>
            </div>
            <span class="insight-count">{{ tips.length }} 条建议</span>
          </div>
          <div class="insight-list">
            <article v-for="tip in tips" :key="tip.title" class="insight-item">
              <div class="insight-icon" :class="tip.tone"><el-icon><component :is="tip.icon" /></el-icon></div>
              <div class="insight-copy"><strong>{{ tip.title }}</strong><p>{{ tip.description }}</p></div>
            </article>
          </div>
          <div class="insight-footer">
            <span class="insight-footer-label">当前状态</span>
            <strong>模板已接入</strong>
            <p>账号激活与密码重置都会优先读取 HTML 模板内容发送，不再回落成纯文本验证码。</p>
          </div>
        </aside>
      </section>

      <section class="templates-card glass-card">
        <div class="templates-head">
          <div>
            <p class="section-tag">模板中心</p>
            <h3>可编辑邮件模板</h3>
            <p class="section-copy">这里固定展示账号激活模板和密码重置模板。每个模板都可以展开、按语言切换，并在大编辑器中直接修改。</p>
          </div>
          <div class="templates-badge"><el-icon><EditPen /></el-icon><span>已连接后端</span></div>
        </div>

        <div class="template-toolbar">
          <div class="search-box"><el-icon><Search /></el-icon><input v-model="search" type="text" placeholder="搜索模板" /></div>
          <select v-model="filter">
            <option value="all">全部模板</option>
            <option value="activation">账号激活模板</option>
            <option value="password-reset">密码重置模板</option>
          </select>
        </div>

        <div class="template-list">
          <article v-for="template in filteredTemplates" :key="template.id" class="template-item" :class="{ expanded: expandedTemplateId === template.id }">
            <button type="button" class="template-summary" @click="toggleTemplate(template.id)">
              <div class="template-copy">
                <div class="template-topline"><h4>{{ template.name }}</h4><span class="status-pill" :class="template.statusTone">{{ template.status }}</span></div>
                <p>{{ template.description }}</p>
              </div>
              <div class="template-actions"><span class="edit-chip">编辑模板</span><el-icon class="expand-icon" :class="{ rotated: expandedTemplateId === template.id }"><ArrowDown /></el-icon></div>
            </button>

            <div v-if="expandedTemplateId === template.id" class="template-editor">
              <div class="editor-shell">
                <div class="editor-topbar ide-topbar">
                  <div class="editor-workspace-tabs"><button type="button" class="workspace-tab active-tab single-tab"><span class="workspace-dot violet-dot"></span><span>{{ template.name }}</span></button></div><div class="editor-dots"><span /><span /><span /></div>
                </div>

                <div class="editor-filebar">
                  <div class="editor-file-main">
                    <span class="file-badge">HTML</span>
                    <strong>{{ template.id }}.{{ activeLanguages[template.id] || 'html' }}</strong>
                    <small>Go 模板运行时</small>
                  </div>
                  <div class="editor-meta"><span class="editor-mode">星云盘邮件模板</span><span class="editor-path">/templates/{{ template.id }}</span></div>
                </div>

                <div class="language-row code-tabs">
                  <div class="language-strip">
                    <button v-for="lang in template.languages" :key="lang.code" type="button" class="language-pill" :class="{ active: activeLanguages[template.id] === lang.code }" @click="activeLanguages[template.id] = lang.code">
                      <span class="language-pill-label">{{ lang.label }}</span>
                      <small>{{ lang.code }}</small>
                    </button>
                  </div>
                  <button type="button" class="language-pill ghost-pill add-language-pill" @click="addLanguage(template)"><el-icon><Promotion /></el-icon>添加语言</button>
                </div>

                <div v-if="currentLanguage(template)" class="editor-fields">
                  <div class="editor-info-grid">
                    <label class="field-card inline-card dark-field">
                      <span class="field-title">语言名称</span>
                      <input v-model="currentLanguage(template)!.label" type="text" placeholder="简体中文 / English / Deutsch" />
                      <small>显示在上方语言标签中。</small>
                    </label>

                    <label class="field-card inline-card dark-field">
                      <span class="field-title">语言代码</span>
                      <input v-model="currentLanguage(template)!.code" type="text" placeholder="zh-CN / en-US / de-DE" />
                      <small>用作模板语言标识。</small>
                    </label>
                  </div>
                  <label class="field-card inline-card">
                    <span class="field-title">邮件主题</span>
                    <input v-model="currentLanguage(template)!.subject" type="text" placeholder="邮件主题" />
                    <small>支持 {{ templateVars.siteName }}、{{ templateVars.url }}、{{ templateVars.code }} 等变量。</small>
                  </label>

                  <div class="code-hints">
                    <span class="hint-chip">站点名称：{{ templateVars.siteName }}</span>
                    <span class="hint-chip">站点地址：{{ templateVars.siteUrl }}</span>
                    <span class="hint-chip">站点 Logo：{{ templateVars.logo }}</span>
                    <span class="hint-chip">跳转链接：{{ templateVars.url }}</span>
                    <span class="hint-chip">验证码：{{ templateVars.code }}</span>
                  </div>

                  <label class="editor-card code-editor-card">
                    <span class="editor-title">模板内容</span>
                    <div class="code-editor">
                      <div class="code-gutter"><span v-for="line in lineNumbers(currentLanguage(template)!.content)" :key="`${template.id}-${activeLanguages[template.id]}-${line}`">{{ line }}</span></div>
                      <textarea v-model="currentLanguage(template)!.content" spellcheck="false" />
                    </div>
                  </label>
                </div>

                <div class="editor-footer">
                  <span>当前语言：{{ activeLanguages[template.id] }}</span>
                  <div class="editor-footer-actions">
                    <button type="button" class="mini-action ghost-action" :disabled="template.languages.length <= 1 || !currentLanguage(template)" @click="removeLanguage(template)">删除当前语言</button>
                    <button type="button" class="mini-action" :disabled="loading || saving" @click="saveTemplate(template)">保存模板</button>
                  </div>
                </div>
              </div>
            </div>
          </article>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { ArrowDown, Connection, EditPen, Hide, Key, Lock, Message, Notification, OfficeBuilding, Position, Promotion, Search, Setting, Timer, User, View } from '@element-plus/icons-vue';
import { getEmailSettings, getEmailTemplates, sendTestEmail, updateEmailSettings, updateEmailTemplate, type EmailTemplatePayload } from '@/api/email-settings';
import { emailTemplateLanguagePresets, normalizeTemplateLanguageCode, templateLanguageLabel } from '@/utils/language';

type MailForm = { enabled: boolean; senderName: string; senderEmail: string; smtpHost: string; smtpPort: string; smtpUsername: string; smtpPassword: string; replyTo: string; forceSSL: boolean; connectionTimeout: string };
type PrimaryField = { key: keyof Pick<MailForm, 'senderName' | 'senderEmail' | 'smtpHost' | 'smtpPort'>; label: string; icon: unknown; placeholder: string; hint: string; type?: string };
type TemplateLanguage = { code: string; label: string; subject: string; content: string };
type TemplateItem = { id: string; name: string; description: string; status: string; statusTone: string; pro: boolean; languages: TemplateLanguage[] };

const templateVars = { siteName: '{{ .CommonContext.SiteBasic.Name }}', siteUrl: '{{ .CommonContext.SiteUrl }}', logo: '{{ .CommonContext.Logo.Normal }}', url: '{{ .Url }}', code: '{{ .Code }}' };
const languagePresets = emailTemplateLanguagePresets;
const activationTemplateEnglishHTML = `<html lang=en-US xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:fill></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><td class=t3 style=width:100px><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">Confirm your account</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><td class=t15 style=width:514px><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">Please click the button below to confirm your email address and finish setting up your account. This link is valid for {{ .TTLHours }} hours.</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">Confirm</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">This email is sent automatically.</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                                           </div>`;
const activationTemplateChineseHTML = `<html lang=zh-CN xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:fill></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><td class=t3 style=width:100px><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">激活你的账号</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><td class=t15 style=width:514px><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">请点击下方按钮确认你的电子邮箱并完成账号注册，此链接有效期为 {{ .TTLHours }} 小时。</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">确认激活</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">此邮件由系统自动发送。</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>     `;
const activationTemplateGermanHTML = `<html lang=de-DE xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:fill></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><td class=t3 style=width:100px><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">Bestätigen Sie Ihr Konto</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><td class=t15 style=width:514px><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">Bitte klicken Sie auf die Schaltfläche unten, um Ihre E-Mail-Adresse zu bestätigen und Ihr Konto einzurichten. Dieser Link ist {{ .TTLHours }} Stunden lang gültig.</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">Bestätigen</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">Diese E-Mail wird automatisch vom System gesendet.</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                                           </div>`;
const passwordResetTemplateEnglishHTML = `<html lang=en-US xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"width=600><![endif]--><!--[if !mso]>--><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><!--[if mso]><td class=t3 style=width:42px width=42><![endif]--><!--[if !mso]>--><td class=t3 style=width:100px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">Reset your password</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t15 style=width:514px width=514><![endif]--><!--[if !mso]>--><td class=t15 style=width:514px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">Please click the button below to reset your password. This link is valid for 1 hour.</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><!--[if mso]><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><![endif]--><!--[if !mso]>--><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">Reset</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">This email is sent automatically.</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                              </div>`;
const passwordResetTemplateSimplifiedChineseHTML = `<html lang=zh-CN xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"width=600><![endif]--><!--[if !mso]>--><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><!--[if mso]><td class=t3 style=width:42px width=42><![endif]--><!--[if !mso]>--><td class=t3 style=width:100px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">重设密码</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t15 style=width:514px width=514><![endif]--><!--[if !mso]>--><td class=t15 style=width:514px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">请点击下方按钮重设你的密码，此链接有效期为 1 小时。</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><!--[if mso]><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><![endif]--><!--[if !mso]>--><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">重设密码</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">此邮件由系统自动发送。</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                              </div>`;
const passwordResetTemplateTraditionalChineseHTML = `<html lang=zh-TW xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"width=600><![endif]--><!--[if !mso]>--><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><!--[if mso]><td class=t3 style=width:42px width=42><![endif]--><!--[if !mso]>--><td class=t3 style=width:100px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">重設密碼</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t15 style=width:514px width=514><![endif]--><!--[if !mso]>--><td class=t15 style=width:514px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">請點擊下方按鈕重設你的密碼，此連結有效期為 1 小時。</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><!--[if mso]><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><![endif]--><!--[if !mso]>--><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">重設密碼</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">此郵件由系統自動發送。</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                                  </div>`;
const passwordResetTemplateGermanHTML = `<html lang=de-DE xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"width=600><![endif]--><!--[if !mso]>--><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><!--[if mso]><td class=t3 style=width:42px width=42><![endif]--><!--[if !mso]>--><td class=t3 style=width:100px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">Passwort zurücksetzen</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t15 style=width:514px width=514><![endif]--><!--[if !mso]>--><td class=t15 style=width:514px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">Bitte klicken Sie auf die Schaltfläche unten, um Ihr Passwort zurückzusetzen. Dieser Link ist 1 Stunde lang gültig.</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><!--[if mso]><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><![endif]--><!--[if !mso]>--><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">Passwort zurücksetzen</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">Diese E-Mail wird automatisch vom System gesendet.</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                                  </div>`;
const passwordResetTemplateSpanishHTML = `<html lang=es-ES xmlns=http://www.w3.org/1999/xhtml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml><title></title><meta charset=UTF-8><meta content="text/html; charset=UTF-8"http-equiv=Content-Type><!--[if !mso]>--><meta content="IE=edge"http-equiv=X-UA-Compatible><!--<![endif]--><meta content=""name=x-apple-disable-message-reformatting><meta content="target-densitydpi=device-dpi"name=viewport><meta content=true name=HandheldFriendly><meta content="width=device-width"name=viewport><meta content="telephone=no, date=no, address=no, email=no, url=no"name=format-detection><style>table{border-collapse:separate;table-layout:fixed;mso-table-lspace:0;mso-table-rspace:0}table td{border-collapse:collapse}.ExternalClass{width:100%}.ExternalClass,.ExternalClass div,.ExternalClass font,.ExternalClass p,.ExternalClass span,.ExternalClass td{line-height:100%}a,body,h1,h2,h3,li,p{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}html{-webkit-text-size-adjust:none!important}#innerTable,body{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}#innerTable img+div{display:none;display:none!important}img{Margin:0;padding:0;-ms-interpolation-mode:bicubic}a,h1,h2,h3,p{line-height:inherit;overflow-wrap:normal;white-space:normal;word-break:break-word}a{text-decoration:none}h1,h2,h3,p{min-width:100%!important;width:100%!important;max-width:100%!important;display:inline-block!important;border:0;padding:0;margin:0}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;font-size:inherit!important;font-family:inherit!important;font-weight:inherit!important;line-height:inherit!important}u+#body a{color:inherit;text-decoration:none;font-size:inherit;font-family:inherit;font-weight:inherit;line-height:inherit}a[href^=mailto],a[href^=sms],a[href^=tel]{color:inherit;text-decoration:none}</style><style>@media (min-width:481px){.hd{display:none!important}}</style><style>@media (max-width:480px){.hm{display:none!important}}</style><style>@media (max-width:480px){.t41,.t46{mso-line-height-alt:0!important;line-height:0!important;display:none!important}.t42{padding:40px!important}.t44{border-radius:0!important;width:480px!important}.t15,.t39,.t9{width:398px!important}.t32{text-align:left!important}.t25{display:revert!important}.t27,.t31{vertical-align:top!important;width:auto!important;max-width:100%!important}}</style><!--[if !mso]>--><link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@700&family=Sofia+Sans:wght@700&family=Open+Sans:wght@400;500;600&display=swap"rel=stylesheet><!--<![endif]--><!--[if mso]><xml><o:officedocumentsettings><o:allowpng><o:pixelsperinch>96</o:pixelsperinch></o:officedocumentsettings></xml><![endif]--><body class=t49 id=body style=min-width:100%;Margin:0;padding:0;background-color:#fff><div style=background-color:#fff class=t48><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100%><tr><td class=t47 style=font-size:0;line-height:0;mso-line-height-rule:exactly;background-color:#fff align=center valign=top><!--[if mso]><v:background xmlns:v=urn:schemas-microsoft-com:vml fill=true stroke=false><v:fill color=#FFFFFF></v:background><![endif]--><table cellpadding=0 cellspacing=0 role=presentation align=center border=0 width=100% id=innerTable><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t41>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t45 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"width=600><![endif]--><!--[if !mso]>--><td class=t44 style="background-color:#fff;border:1px solid #ebebeb;overflow:hidden;width:600px;border-radius:12px 12px 12px 12px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t43 style=width:100% width=100%><tr><td class=t42 style="padding:44px 42px 32px 42px"><table cellpadding=0 cellspacing=0 role=presentation style=width:100%!important width=100%><tr><td align=left><table cellpadding=0 cellspacing=0 role=presentation class=t4 style=Margin-right:auto><tr><!--[if mso]><td class=t3 style=width:42px width=42><![endif]--><!--[if !mso]>--><td class=t3 style=width:100px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t2 style=width:100% width=100%><tr><td class=t1><div style=font-size:0><a href="{{ .CommonContext.SiteUrl }}"><img alt=""class=t0 height=100 src="{{ .CommonContext.Logo.Normal }}"style=display:block;border:0;height:auto;width:100%;Margin:0;max-width:100%></a></div></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:22px;line-height:22px;font-size:1px;display:block class=t5>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t10 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t9 style="border-bottom:1px solid #eff1f4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t8 style=width:100% width=100%><tr><td class=t7 style="padding:0 0 18px 0"><h1 class=t6 style="margin:0;Margin:0;font-family:Montserrat,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:28px;font-weight:700;font-style:normal;font-size:24px;text-decoration:none;text-transform:none;letter-spacing:-1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:1px">Restablecer tu contraseña</h1></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:18px;line-height:18px;font-size:1px;display:block class=t11>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t16 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t15 style=width:514px width=514><![endif]--><!--[if !mso]>--><td class=t15 style=width:514px><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t14 style=width:100% width=100%><tr><td class=t13><p class=t12 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:25px;font-weight:400;font-style:normal;font-size:15px;text-decoration:none;text-transform:none;letter-spacing:-.1px;direction:ltr;color:#141414;text-align:left;mso-line-height-rule:exactly;mso-text-raise:3px">Por favor, haz clic en el botón de abajo para restablecer tu contraseña. Este enlace es válido por 1 hora.</table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:24px;line-height:24px;font-size:1px;display:block class=t18>  </div><tr><td align=left><a href="{{ .Url }}"><table cellpadding=0 cellspacing=0 role=presentation class=t22 style=margin-right:auto><tr><!--[if mso]><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><![endif]--><!--[if !mso]>--><td class=t21 style="background-color:#0666eb;overflow:hidden;width:auto;border-radius:40px 40px 40px 40px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t20 style=width:auto><tr><td class=t19 style="line-height:34px;mso-line-height-rule:exactly;mso-text-raise:5px;padding:0 23px 0 23px"><span class=t17 style="display:block;margin:0;Margin:0;font-family:Sofia Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:34px;font-weight:700;font-style:normal;font-size:16px;text-decoration:none;text-transform:none;letter-spacing:-.2px;direction:ltr;color:#fff;mso-line-height-rule:exactly;mso-text-raise:5px">Restablecer</span></table></table></a><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:40px;line-height:40px;font-size:1px;display:block class=t36>  </div><tr><td align=center><table cellpadding=0 cellspacing=0 role=presentation class=t40 style=Margin-left:auto;Margin-right:auto><tr><!--[if mso]><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"width=514><![endif]--><!--[if !mso]>--><td class=t39 style="border-top:1px solid #dfe1e4;width:514px"><!--<![endif]--><table cellpadding=0 cellspacing=0 role=presentation class=t38 style=width:100% width=100%><tr><td class=t37 style="padding:24px 0 0 0"><div style=width:100%;text-align:left class=t35><div style=display:inline-block class=t34><table cellpadding=0 cellspacing=0 role=presentation class=t33 align=left valign=top><tr class=t32><td><td class=t27 valign=top><table cellpadding=0 cellspacing=0 role=presentation class=t26 style=width:auto width=100%><tr><td class=t24 style=background-color:#fff;line-height:20px;mso-line-height-rule:exactly;mso-text-raise:2px><span class=t23 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:600;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#222;mso-line-height-rule:exactly;mso-text-raise:2px">{{ .CommonContext.SiteBasic.Name }}</span> <span class=t28 style="margin:0;Margin:0;font-family:Open Sans,BlinkMacSystemFont,Segoe UI,Helvetica Neue,Arial,sans-serif;line-height:20px;font-weight:500;font-style:normal;font-size:14px;text-decoration:none;direction:ltr;color:#b4becc;mso-line-height-rule:exactly;mso-text-raise:2px;margin-left:8px">Este correo electrónico se envía automáticamente.</span><td class=t25 style=width:20px width=20></table><td></table></div></div></table></table></table></table></table><tr><td><div style=mso-line-height-rule:exactly;mso-line-height-alt:50px;line-height:50px;font-size:1px;display:block class=t46>  </div></table></table></div><div style="display:none;white-space:nowrap;font:15px courier;line-height:0"class=gmail-fix>                                                  </div>`;
const buildActivationFallbackHTML = (lang: string, title: string, body: string, ttlText: string, buttonLabel: string, footerText: string) => `<!doctype html>
<html lang="${lang}">
  <head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width, initial-scale=1.0" /><title>${title}</title></head>
  <body style="margin:0;padding:0;background:#f5f9ff;font-family:'Segoe UI',Arial,sans-serif;">
    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="padding:32px 16px;background:#f5f9ff;">
      <tr><td align="center">
        <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="max-width:600px;background:#fff;border:1px solid #e7edf8;border-radius:18px;">
          <tr><td style="padding:44px 42px 32px;">
            <a href="{{ .CommonContext.SiteUrl }}"><img src="{{ .CommonContext.Logo.Normal }}" alt="{{ .CommonContext.SiteBasic.Name }}" style="display:block;width:100px;height:auto;border:0;" /></a>
            <div style="height:22px"></div><h1 style="margin:0 0 18px;font-size:24px;line-height:28px;color:#141414;">${title}</h1>
            <p style="margin:0 0 10px;font-size:15px;line-height:25px;color:#141414;">${body}</p>
            <p style="margin:0 0 24px;font-size:14px;line-height:22px;color:#64748b;">${ttlText}</p>
            <a href="{{ .Url }}" style="display:inline-block;padding:0 23px;line-height:44px;border-radius:999px;background:#0666eb;color:#fff;text-decoration:none;font-weight:700;">${buttonLabel}</a>
            <div style="height:24px"></div><p style="margin:0;font-size:13px;line-height:22px;color:#64748b;">{{ .Url }}</p>
            <p style="margin:8px 0 0;font-size:14px;line-height:22px;color:#0666eb;">{{ .Code }}</p>
            <div style="height:24px"></div><div style="border-top:1px solid #dfe4ec;padding-top:18px;font-size:13px;line-height:20px;color:#94a3b8;">${footerText}</div>
          </td></tr>
        </table>
      </td></tr>
    </table>
  </body>
</html>`;
const createDefaultForm = (): MailForm => ({ enabled: true, senderName: '星云盘', senderEmail: '3518974413@qq.com', smtpHost: 'smtp.qq.com', smtpPort: '587', smtpUsername: '3518974413@qq.com', smtpPassword: '', replyTo: '3518974413@qq.com', forceSSL: false, connectionTimeout: '30' });
const expandWithPresetLanguages = (languages: TemplateLanguage[]) => { const result = [...languages]; const seed = languages[0]; for (const preset of languagePresets) { if (!result.some((item) => item.code === preset.code)) result.push({ code: preset.code, label: preset.label, subject: seed?.subject || '', content: seed?.content || '' }); } return result; };
const createFallbackTemplates = (): TemplateItem[] => [
  { id: 'activation', name: '账号激活模板', description: '用户注册后发送的账号激活邮件模板。', status: '首选', statusTone: 'info', pro: false, languages: expandWithPresetLanguages([{ code: 'en-US', label: 'English', subject: '{{ .CommonContext.SiteBasic.Name }} Confirm your account', content: buildActivationFallbackHTML('en-US', 'Confirm your account', 'Please click the button below to confirm your email address and finish setting up your account.', 'This link is valid for {{ .TTLHours }} hours.', 'Confirm', 'This email is sent automatically.') }, { code: 'zh-CN', label: '简体中文', subject: '{{ .CommonContext.SiteBasic.Name }} 激活你的账号', content: buildActivationFallbackHTML('zh-CN', '激活你的账号', '请点击下方按钮确认你的电子邮箱并完成账号注册。', '此链接有效期为 {{ .TTLHours }} 小时。', '确认激活', '此邮件由系统自动发送。') }, { code: 'de-DE', label: 'Deutsch', subject: '{{ .CommonContext.SiteBasic.Name }} Bestätigen Sie Ihr Konto', content: buildActivationFallbackHTML('de-DE', 'Bestätigen Sie Ihr Konto', 'Bitte klicken Sie auf die Schaltfläche unten, um Ihre E-Mail-Adresse zu bestätigen und Ihr Konto einzurichten.', 'Dieser Link ist {{ .TTLHours }} Stunden lang gültig.', 'Bestätigen', 'Diese E-Mail wird automatisch vom System gesendet.') }, { code: 'es-ES', label: 'Español', subject: '{{ .CommonContext.SiteBasic.Name }} Confirma tu cuenta', content: buildActivationFallbackHTML('es-ES', 'Confirma tu cuenta', 'Haz clic en el botón para confirmar tu correo y terminar de crear tu cuenta.', 'Este enlace es válido durante {{ .TTLHours }} horas.', 'Confirmar', 'Este correo se envía automáticamente.') }, { code: 'fr-FR', label: 'Français', subject: '{{ .CommonContext.SiteBasic.Name }} Confirmez votre compte', content: buildActivationFallbackHTML('fr-FR', 'Confirmez votre compte', 'Cliquez sur le bouton pour confirmer votre e-mail et terminer la création du compte.', 'Ce lien est valable pendant {{ .TTLHours }} heures.', 'Confirmer', 'Cet e-mail est envoyé automatiquement.') }, { code: 'it-IT', label: 'Italiano', subject: '{{ .CommonContext.SiteBasic.Name }} Conferma il tuo account', content: buildActivationFallbackHTML('it-IT', 'Conferma il tuo account', 'Fai clic sul pulsante per confermare la tua email e completare la registrazione.', 'Questo link è valido per {{ .TTLHours }} ore.', 'Conferma', 'Questa email è inviata automaticamente.') }, { code: 'ja-JP', label: '日本語', subject: '{{ .CommonContext.SiteBasic.Name }} アカウントを有効化してください', content: buildActivationFallbackHTML('ja-JP', 'アカウントを有効化してください', '下のボタンをクリックしてメールアドレスを確認し、登録を完了してください。', 'このリンクは {{ .TTLHours }} 時間有効です。', '確認する', 'このメールは自動送信されています。') }]) },
  { id: 'password-reset', name: '密码重置模板', description: '用户找回密码时发送的密码重置邮件模板。', status: '已启用', statusTone: 'success', pro: false, languages: expandWithPresetLanguages([{ code: 'en-US', label: 'English', subject: '{{ .CommonContext.SiteBasic.Name }} Reset your password', content: passwordResetTemplateEnglishHTML }, { code: 'zh-CN', label: '简体中文', subject: '{{ .CommonContext.SiteBasic.Name }} 重设密码', content: passwordResetTemplateSimplifiedChineseHTML }, { code: 'zh-TW', label: '繁體中文', subject: '{{ .CommonContext.SiteBasic.Name }} 重設密碼', content: passwordResetTemplateTraditionalChineseHTML }, { code: 'de-DE', label: 'Deutsch', subject: '{{ .CommonContext.SiteBasic.Name }} Passwort zurücksetzen', content: passwordResetTemplateGermanHTML }, { code: 'es-ES', label: 'Español', subject: '{{ .CommonContext.SiteBasic.Name }} Restablecer tu contraseña', content: passwordResetTemplateSpanishHTML }]) },
];
const primaryFields: PrimaryField[] = [
  { key: 'senderName', label: '发件人名', icon: Notification, placeholder: '星云盘', hint: '显示在收件箱中的品牌名称。' },
  { key: 'senderEmail', label: '发件人邮箱', icon: Message, type: 'email', placeholder: '3518974413@qq.com', hint: '这里已经修正为真实发件邮箱。' },
  { key: 'smtpHost', label: 'SMTP 服务器', icon: OfficeBuilding, placeholder: 'smtp.qq.com', hint: 'QQ 邮箱默认 SMTP 服务器为 smtp.qq.com。' },
  { key: 'smtpPort', label: 'SMTP 端口', icon: Setting, placeholder: '587', hint: '常用端口为 587。' },
];
const tips = [
  { title: '标签式语言切换', description: '语言切换位于编辑器顶部，更接近代码编辑器的标签页体验。', icon: Position, tone: 'tone-blue' },
  { title: '大尺寸代码编辑区', description: 'HTML 模板编辑器拥有更高的编辑空间，并带有行号区域。', icon: EditPen, tone: 'tone-violet' },
  { title: '模板变量提示', description: '常用模板变量会以标签形式展示，编辑 HTML 更直观。', icon: Connection, tone: 'tone-cyan' },
];

const form = reactive(createDefaultForm());
const templates = ref<TemplateItem[]>([]);
const loading = ref(false);
const saving = ref(false);
const showPassword = ref(false);
const advancedOpen = ref(true);
const search = ref('');
const filter = ref<'all' | 'activation' | 'password-reset'>('all');
const expandedTemplateId = ref('activation');
const activeLanguages = reactive<Record<string, string>>({});
const filteredTemplates = computed(() => templates.value.filter((item) => { const keyword = search.value.trim().toLowerCase(); const matchesKeyword = !keyword || [item.name, item.description].some((value) => String(value || '').toLowerCase().includes(keyword)); const matchesFilter = filter.value === 'all' || item.id === filter.value; return matchesKeyword && matchesFilter; }));
function currentLanguage(template: TemplateItem) { return template.languages.find((item) => item.code === activeLanguages[template.id]); }
function lineNumbers(content: string) { return Array.from({ length: Math.max((content.match(/\n/g)?.length ?? 0) + 1, 12) }, (_, i) => i + 1); }
function applySettings(payload: Awaited<ReturnType<typeof getEmailSettings>>) { Object.assign(form, { enabled: payload.enabled, senderName: payload.from_name, senderEmail: payload.from_address, smtpHost: payload.host || 'smtp.qq.com', smtpPort: String(payload.port || 587), smtpUsername: payload.username, smtpPassword: payload.password, replyTo: payload.reply_to || payload.from_address, forceSSL: payload.force_ssl, connectionTimeout: String(payload.connection_timeout || 30) }); }
function applyTemplates(payload: EmailTemplatePayload[] | undefined) { const normalized = (payload ?? []).map((item) => ({ id: String(item?.template_key ?? '').trim(), name: String(item?.name ?? '').trim() || 'Untitled Template', description: String(item?.description ?? '').trim(), status: String(item?.status ?? '').trim() || 'enabled', statusTone: String(item?.status_tone ?? '').trim() || 'info', pro: Boolean(item?.pro), languages: Array.isArray(item?.languages) ? item.languages.map((lang) => { const code = normalizeTemplateLanguageCode(lang?.code); return { code, label: String(lang?.label ?? '').trim() || templateLanguageLabel(code), subject: String(lang?.subject ?? ''), content: String(lang?.content ?? '') }; }).filter((lang) => lang.code) : [] })).filter((item) => item.id); const ensured = normalized.length ? normalized : createFallbackTemplates(); for (const fallback of createFallbackTemplates()) if (!ensured.some((item) => item.id === fallback.id)) ensured.push(fallback); templates.value = ensured; for (const item of templates.value) { if (!item.languages.length) item.languages = createFallbackTemplates().find((entry) => entry.id === item.id)?.languages ?? []; if (!activeLanguages[item.id] || !item.languages.some((lang) => lang.code === activeLanguages[item.id])) activeLanguages[item.id] = item.languages[0]?.code ?? ''; } }
async function loadData() { loading.value = true; try { const [settings, templateList] = await Promise.all([getEmailSettings(), getEmailTemplates()]); applySettings(settings); applyTemplates(templateList); } catch (error: any) { applyTemplates([]); ElMessage.warning(error?.message || '邮件模板加载失败，已启用本地兜底模板'); } finally { loading.value = false; } }
function toggleTemplate(id: string) { expandedTemplateId.value = expandedTemplateId.value === id ? '' : id; }
async function save() { saving.value = true; try { await updateEmailSettings({ enabled: form.enabled, provider: 'qq', host: form.smtpHost, port: Number(form.smtpPort), username: form.smtpUsername, password: form.smtpPassword, from_name: form.senderName, from_address: form.senderEmail, reply_to: form.replyTo, force_ssl: form.forceSSL, connection_timeout: Number(form.connectionTimeout), code_ttl_seconds: 300, send_interval_seconds: 60 }); ElMessage.success('邮件设置已保存到后端'); } finally { saving.value = false; } }
async function reload() { await loadData(); ElMessage.success('邮件模板已重新加载'); }
async function reset() { Object.assign(form, createDefaultForm()); applyTemplates([]); ElMessage.success('已恢复为推荐默认值'); }
async function handleSendTest() { const target = form.replyTo || form.senderEmail; await sendTestEmail(target); ElMessage.success(`测试邮件已发送到 ${target}`); }
async function saveTemplate(template: TemplateItem) { const languages = template.languages.map((item) => ({ ...item, code: normalizeTemplateLanguageCode(item.code), label: item.label || templateLanguageLabel(item.code) })); await updateEmailTemplate(template.id, { template_key: template.id, name: template.name, description: template.description, status: template.status, status_tone: template.statusTone, pro: template.pro, languages }); template.languages = languages; ElMessage.success(`${template.name} 已保存到后端`); }
function addLanguage(template: TemplateItem) { const used = new Set(template.languages.map((item) => normalizeTemplateLanguageCode(item.code))); const preset = languagePresets.find((item) => !used.has(item.code)); const seed = currentLanguage(template) ?? template.languages[0]; const index = template.languages.length + 1; const item = { code: preset?.code ?? ('custom-' + index), label: preset?.label ?? ('新语言 ' + index), subject: seed?.subject || '', content: seed?.content || '' }; template.languages.push(item); activeLanguages[template.id] = item.code; }
function removeLanguage(template: TemplateItem) { const code = activeLanguages[template.id]; const index = template.languages.findIndex((item) => item.code === code); if (index < 0 || template.languages.length <= 1) return; template.languages.splice(index, 1); activeLanguages[template.id] = template.languages[0]?.code ?? ''; }
onMounted(loadData);
</script>

<style scoped>
*,:before,:after{box-sizing:border-box}.mail-settings,.mail-shell,.field-grid,.mail-stats,.insight-list,.template-list,.editor-fields{display:grid;gap:18px}.form-stack{display:grid;gap:18px;min-width:0}.glass-card{position:relative;border:1px solid rgba(255,255,255,.78);border-radius:28px;background:rgba(255,255,255,.74);box-shadow:0 22px 48px rgba(15,23,42,.08),inset 0 1px 0 rgba(255,255,255,.9);backdrop-filter:blur(18px)}.mail-hero,.settings-card,.templates-card,.insight-card{padding:24px}.mail-hero{display:flex;align-items:flex-start;justify-content:space-between;gap:18px;background:radial-gradient(circle at top right,rgba(37,99,235,.18),transparent 28%),radial-gradient(circle at left bottom,rgba(129,140,248,.12),transparent 22%),rgba(255,255,255,.8)}.mail-layout{display:grid;grid-template-columns:minmax(0,1.32fr) minmax(300px,348px);align-items:start;gap:24px}.section-tag,.subsection-note,.field-card small,.section-copy,.editor-footer span{margin:0;color:#64748b}.section-tag{font-size:12px;font-weight:800;letter-spacing:.16em;text-transform:uppercase}h2,h3,h4,p,strong{margin:0}h2,h3,h4,strong{color:#0f172a}h2{margin-top:8px;font-size:40px;line-height:1}h3{margin-top:6px;font-size:26px}.section-copy{margin-top:12px;max-width:720px;line-height:1.8}.mail-stats{grid-template-columns:repeat(2,minmax(0,1fr));min-width:300px}.stat-pill{padding:18px;border-radius:22px;background:linear-gradient(145deg,rgba(255,255,255,.96),rgba(239,246,255,.76));border:1px solid rgba(219,234,254,.84)}.stat-pill span,.templates-badge span,.status-pill,.edit-chip{font-size:13px;font-weight:700;color:#475569}.stat-pill strong{display:block;margin-top:10px;font-size:26px}.card-head,.subsection-head,.template-summary,.template-topline,.template-actions,.language-row,.advanced-toggle,.toggle-row,.editor-footer,.template-toolbar,.templates-head,.editor-footer-actions,.editor-topbar,.editor-meta{display:flex;align-items:center;justify-content:space-between;gap:14px}.compact-head{margin-bottom:8px}.status-chip,.templates-badge{display:inline-flex;align-items:center;gap:10px;min-height:40px;padding:0 14px;border-radius:999px;background:rgba(239,246,255,.92);color:#1d4ed8;font-weight:700}.status-dot{width:9px;height:9px;border-radius:999px;background:#22c55e;box-shadow:0 0 0 6px rgba(34,197,94,.12)}.field-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.field-grid.compact{grid-template-columns:repeat(2,minmax(0,1fr))}.field-card,.editor-card,.auth-card{padding:18px;border-radius:24px;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(248,250,252,.88));border:1px solid rgba(226,232,240,.92);min-width:0;overflow:hidden}.auth-card{display:grid;gap:16px;background:linear-gradient(180deg,rgba(244,249,255,.92),rgba(255,255,255,.92))}.field-title,.editor-title{display:inline-flex;align-items:center;gap:10px;margin-bottom:12px;color:#0f172a;font-weight:800}.field-title :deep(svg){width:18px;height:18px;color:#1677ff}input,select,textarea{width:100%;border:1px solid rgba(203,213,225,.95);border-radius:18px;background:rgba(255,255,255,.92);color:#0f172a;outline:none;transition:border-color .2s ease,box-shadow .2s ease}input,select{height:54px;padding:0 16px}textarea{padding:16px;font:500 14px/1.7 "Consolas","SFMono-Regular",monospace;resize:vertical}input:focus,select:focus,textarea:focus{border-color:#60a5fa;box-shadow:0 0 0 4px rgba(96,165,250,.18)}.password-shell{display:flex;align-items:center;gap:10px}.password-shell input{flex:1}.icon-ghost{display:inline-flex;align-items:center;justify-content:center;width:54px;height:54px;border:none;border-radius:18px;background:rgba(239,246,255,.92);color:#2563eb;cursor:pointer}.advanced-card{padding:8px 24px 24px}.advanced-toggle{width:100%;padding:12px 0;background:none;border:none;color:inherit;cursor:pointer}.advanced-toggle :deep(svg),.expand-icon,.language-pill,.ssl-switch{transition:all .22s ease}.advanced-toggle .rotated,.expand-icon.rotated{transform:rotate(180deg)}.advanced-body{display:grid;gap:16px}.toggle-row{padding:18px 20px;border-radius:24px;background:linear-gradient(135deg,rgba(239,246,255,.92),rgba(248,250,252,.96));border:1px solid rgba(191,219,254,.88)}.ssl-switch{position:relative;flex:none;width:74px;height:40px;border:none;border-radius:999px;background:#dbe5f3;cursor:pointer}.ssl-switch span{position:absolute;top:4px;left:4px;width:32px;height:32px;border-radius:50%;background:#fff;box-shadow:0 10px 20px rgba(15,23,42,.16)}.ssl-switch.active{background:linear-gradient(135deg,#1677ff 0%,#6d7dff 100%);box-shadow:0 0 0 6px rgba(22,119,255,.12)}.ssl-switch.active span{left:38px}.inline-card{padding-bottom:16px}.form-actions{display:flex;justify-content:flex-end;gap:12px}.outline-action,.primary-action,.mini-action,.language-pill,.template-summary,.advanced-toggle{font:inherit}.outline-action,.primary-action,.mini-action{display:inline-flex;align-items:center;justify-content:center;gap:10px;min-height:52px;padding:0 22px;border-radius:18px;cursor:pointer}.outline-action{border:1px solid rgba(96,165,250,.9);background:#fff;color:#1677ff;box-shadow:0 10px 22px rgba(59,130,246,.08)}.primary-action{border:none;background:linear-gradient(135deg,#1d4ed8 0%,#1677ff 45%,#38bdf8 100%);color:#fff;box-shadow:0 18px 32px rgba(37,99,235,.24)}.outline-action:disabled,.primary-action:disabled,.mini-action:disabled{opacity:.6;cursor:not-allowed}.insight-card{position:sticky;top:20px;align-self:start;display:grid;gap:20px;overflow:hidden;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(244,249,255,.9));border-color:rgba(219,234,254,.94);box-shadow:0 24px 52px rgba(59,130,246,.12),inset 0 1px 0 rgba(255,255,255,.92)}.insight-orb{position:absolute;top:-48px;right:-30px;width:172px;height:172px;border-radius:999px;background:radial-gradient(circle,rgba(56,189,248,.3) 0%,rgba(96,165,250,.2) 34%,transparent 72%);pointer-events:none}.insight-head{position:relative;display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:end;gap:12px 16px;padding-bottom:18px;border-bottom:1px solid rgba(226,232,240,.9)}.insight-caption{margin:0;max-width:none;color:#64748b;line-height:1.8}.insight-count{display:inline-flex;align-items:center;justify-content:center;justify-self:end;align-self:start;min-height:36px;padding:0 14px;border-radius:999px;background:rgba(239,246,255,.98);color:#2563eb;font-size:12px;font-weight:800;white-space:nowrap;box-shadow:inset 0 1px 0 rgba(255,255,255,.9)}.insight-list{display:grid;gap:14px}.insight-item{display:grid;grid-template-columns:auto 1fr;gap:14px;padding:17px 17px 18px;border-radius:22px;background:linear-gradient(180deg,rgba(255,255,255,.98),rgba(248,250,252,.94));border:1px solid rgba(226,232,240,.9);box-shadow:0 14px 28px rgba(148,163,184,.09)}.insight-copy{display:grid;gap:6px}.insight-copy strong{font-size:15px;color:#0f172a}.insight-icon{display:inline-flex;align-items:center;justify-content:center;width:48px;height:48px;border-radius:18px;box-shadow:inset 0 1px 0 rgba(255,255,255,.8)}.insight-icon.tone-blue{background:rgba(37,99,235,.12);color:#2563eb}.insight-icon.tone-violet{background:rgba(139,92,246,.12);color:#7c3aed}.insight-icon.tone-cyan{background:rgba(6,182,212,.12);color:#0891b2}.insight-item p{margin:0;color:#64748b;line-height:1.7}.insight-footer{display:grid;gap:8px;padding:18px 18px 19px;border-radius:22px;background:linear-gradient(135deg,rgba(219,234,254,.56),rgba(255,255,255,.8));border:1px solid rgba(191,219,254,.92)}.insight-footer-label{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#3b82f6}.insight-footer strong{font-size:18px;color:#0f172a}.insight-footer p{margin:0;color:#64748b;line-height:1.75}.templates-card{overflow:hidden}.template-toolbar{margin:8px 0 4px}.search-box{display:flex;align-items:center;gap:10px;flex:1;min-height:54px;padding:0 16px;border:1px solid rgba(203,213,225,.92);border-radius:18px;background:rgba(255,255,255,.92)}.search-box input{height:auto;padding:0;border:none;background:transparent;box-shadow:none}.template-item{border:1px solid rgba(226,232,240,.88);border-radius:26px;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(248,250,252,.9));overflow:hidden;transition:box-shadow .22s ease,border-color .22s ease,transform .22s ease}.template-item.expanded,.template-item:hover{transform:translateY(-2px);border-color:rgba(96,165,250,.72);box-shadow:0 18px 34px rgba(59,130,246,.12)}.template-summary{width:100%;padding:22px 24px;background:none;border:none;cursor:pointer;text-align:left}.template-copy p{margin-top:10px;color:#64748b;line-height:1.7}.status-pill{padding:7px 10px;border-radius:999px;background:rgba(241,245,249,.9)}.status-pill.success{color:#15803d;background:rgba(220,252,231,.9)}.status-pill.info{color:#1d4ed8;background:rgba(219,234,254,.92)}.edit-chip{padding:10px 14px;border-radius:999px;background:rgba(239,246,255,.92);color:#1d4ed8}.template-editor{padding:0 24px 24px;min-width:0;overflow:hidden}.editor-shell{display:grid;gap:16px;padding:16px;border-radius:24px;background:linear-gradient(180deg,#0f172a 0%,#111827 100%);box-shadow:inset 0 1px 0 rgba(255,255,255,.08);min-width:0;overflow:hidden}.editor-topbar{padding:0}.ide-topbar{padding:4px 2px 0;justify-content:space-between}.editor-workspace-tabs{display:flex;gap:10px;flex-wrap:wrap;min-width:0}.workspace-tab{display:inline-flex;align-items:center;gap:10px;min-height:38px;padding:0 14px;border:1px solid rgba(51,65,85,.9);border-radius:14px;background:rgba(8,15,30,.9);color:#94a3b8;cursor:default;max-width:100%}.workspace-tab.active-tab{background:linear-gradient(135deg,rgba(37,99,235,.26),rgba(109,125,255,.24));color:#e2e8f0;border-color:rgba(96,165,250,.6);box-shadow:0 12px 24px rgba(37,99,235,.12)}.single-tab{padding-right:18px}.workspace-dot{width:9px;height:9px;border-radius:999px;flex:none}.blue-dot{background:#38bdf8}.violet-dot{background:#818cf8}.editor-dots{display:flex;gap:8px;flex:none}.editor-dots span{width:10px;height:10px;border-radius:50%;background:#334155}.editor-dots span:nth-child(1){background:#f87171}.editor-dots span:nth-child(2){background:#fbbf24}.editor-dots span:nth-child(3){background:#34d399}.editor-filebar{display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:center;gap:14px;padding:14px 16px;border:1px solid rgba(51,65,85,.9);border-radius:20px;background:linear-gradient(180deg,rgba(9,14,29,.95),rgba(12,19,35,.92));overflow:hidden;min-width:0}.editor-file-main{display:flex;align-items:center;gap:12px;flex-wrap:wrap;min-width:0;overflow:hidden}.file-badge{display:inline-flex;align-items:center;justify-content:center;min-width:50px;height:28px;padding:0 10px;border-radius:999px;background:rgba(37,99,235,.18);color:#7dd3fc;font-size:12px;font-weight:800;letter-spacing:.08em;flex:none}.editor-file-main strong{color:#f8fafc;font-size:15px;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.editor-file-main small{color:#64748b}.editor-meta{margin-left:0;display:flex;align-items:center;justify-content:flex-end;gap:10px;flex-wrap:wrap;min-width:0;max-width:100%;text-align:right}.editor-mode,.editor-path{font-size:12px;color:#94a3b8;word-break:break-all}.language-row{align-items:stretch;gap:12px;flex-wrap:nowrap;min-width:0;overflow:hidden}.language-strip{display:flex;gap:10px;flex:1;min-width:0;overflow-x:auto;overflow-y:hidden;padding:6px;border:1px solid rgba(51,65,85,.8);border-radius:20px;background:rgba(8,15,30,.9)}.language-pill{display:inline-flex;flex-direction:column;align-items:flex-start;justify-content:center;gap:4px;min-width:132px;padding:12px 16px;border:1px solid rgba(71,85,105,.9);border-radius:18px;background:rgba(15,23,42,.9);color:#cbd5e1;cursor:pointer}.language-pill small{font-size:11px;color:#7dd3fc}.language-pill-label{font-weight:700}.language-pill.active{border-color:transparent;background:linear-gradient(135deg,#1677ff 0%,#6d7dff 100%);color:#fff;box-shadow:0 12px 24px rgba(37,99,235,.18)}.language-pill.active small{color:rgba(255,255,255,.82)}.ghost-pill{background:rgba(15,23,42,.72);color:#7dd3fc}.add-language-pill{align-self:stretch;min-width:136px;flex:none;justify-content:center;align-items:center}.editor-info-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px;min-width:0}.dark-field{background:rgba(15,23,42,.72);border-color:rgba(51,65,85,.9)}.dark-field .field-title,.dark-field small{color:#cbd5e1}.dark-field input{background:rgba(15,23,42,.78);border-color:rgba(71,85,105,.9);color:#e2e8f0}.code-hints{display:flex;flex-wrap:wrap;gap:10px}.hint-chip{padding:8px 12px;border:1px solid rgba(51,65,85,.9);border-radius:999px;background:rgba(15,23,42,.72);font-size:12px;color:#cbd5e1}.code-editor-card{padding:0;overflow:hidden;border-color:rgba(51,65,85,.9);background:#0b1220}.editor-title{display:flex;padding:16px 18px;margin:0;border-bottom:1px solid rgba(51,65,85,.9);background:rgba(15,23,42,.96);color:#e2e8f0}.code-editor{display:grid;grid-template-columns:60px minmax(0,1fr);min-height:520px;min-width:0;overflow:hidden}.code-gutter{display:grid;align-content:start;gap:0;padding:18px 12px;background:#08101d;border-right:1px solid rgba(51,65,85,.9);font:500 13px/1.7 Consolas,monospace;color:#64748b;text-align:right}.code-editor textarea{min-height:520px;border:none;border-radius:0;background:#0b1220;color:#dbeafe;box-shadow:none;min-width:0}.ghost-action{background:#fff;color:#475569}@media (max-width:1200px){.mail-layout{grid-template-columns:1fr}.insight-card{position:relative;top:auto;order:-1}.mail-hero{flex-direction:column}.mail-stats,.field-grid,.field-grid.compact{grid-template-columns:1fr}}@media (max-width:720px){.mail-hero,.settings-card,.templates-card,.insight-card{padding:18px}.template-summary,.template-editor{padding-left:18px;padding-right:18px}.template-toolbar,.templates-head,.card-head,.subsection-head,.toggle-row,.editor-footer,.form-actions,.editor-footer-actions,.editor-topbar{flex-direction:column;align-items:stretch}.template-topline,.template-actions,.editor-meta,.language-row{flex-wrap:wrap}.outline-action,.primary-action,select,.code-editor,.add-language-pill{width:100%}.editor-info-grid{grid-template-columns:1fr}.language-strip{width:100%}.code-editor{grid-template-columns:42px minmax(0,1fr)}h2{font-size:34px}}

.mail-settings {
  isolation: isolate;
  position: relative;
}

.mail-settings::before {
  content: '';
  position: absolute;
  inset: -34px -36px auto;
  z-index: -1;
  height: 330px;
  background:
    radial-gradient(circle at 12% 18%, rgba(100, 195, 255, 0.28), transparent 28%),
    radial-gradient(circle at 48% 8%, rgba(146, 127, 255, 0.16), transparent 30%),
    radial-gradient(circle at 82% 24%, rgba(255, 170, 190, 0.22), transparent 32%);
  filter: blur(6px);
  pointer-events: none;
}

.mail-shell {
  gap: 22px;
}

.glass-card {
  border-color: rgba(255, 255, 255, 0.82);
  background:
    linear-gradient(142deg, rgba(255, 255, 255, 0.78), rgba(244, 251, 255, 0.58) 54%, rgba(255, 237, 241, 0.5)),
    rgba(255, 255, 255, 0.62);
  box-shadow:
    0 24px 52px rgba(102, 157, 204, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    inset 0 -1px 0 rgba(133, 180, 224, 0.12);
  backdrop-filter: blur(22px) saturate(1.18);
}

.mail-hero {
  overflow: hidden;
  min-height: 172px;
  border-radius: 30px;
  background:
    radial-gradient(circle at 6% 0%, rgba(255, 255, 255, 0.96), transparent 30%),
    radial-gradient(circle at 74% 10%, rgba(87, 184, 234, 0.24), transparent 32%),
    radial-gradient(circle at 96% 76%, rgba(255, 184, 196, 0.26), transparent 28%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.82), rgba(234, 247, 255, 0.66) 56%, rgba(255, 239, 239, 0.6));
}

.section-tag {
  color: #4c9ee2;
}

h2,
h3,
h4,
strong {
  color: #17334f;
}

.section-copy,
.subsection-note,
.field-card small,
.editor-footer span {
  color: #657b91;
}

.stat-pill,
.field-card,
.auth-card,
.template-item,
.search-box,
select,
input {
  border-color: rgba(255, 255, 255, 0.72);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(248, 253, 255, 0.62));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.82),
    0 12px 24px rgba(127, 172, 213, 0.08);
}

.field-card,
.auth-card {
  border-radius: 22px;
}

.field-title {
  color: #1a3854;
}

.field-title :deep(svg) {
  color: #5ab5df;
}

input:focus,
select:focus,
textarea:focus {
  border-color: rgba(97, 191, 230, 0.92);
  box-shadow:
    0 0 0 4px rgba(130, 217, 241, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.status-chip,
.templates-badge,
.edit-chip {
  border: 1px solid rgba(255, 255, 255, 0.64);
  background: linear-gradient(90deg, rgba(219, 245, 255, 0.88), rgba(255, 218, 226, 0.74));
  color: #2680c5;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.primary-action,
.mini-action:not(.ghost-action) {
  background: linear-gradient(135deg, #58b7ef 0%, #5fd3dc 58%, #f2a5ad 100%);
  box-shadow: 0 16px 30px rgba(86, 181, 227, 0.24);
}

.outline-action,
.ghost-action,
.icon-ghost {
  border: 1px solid rgba(212, 230, 244, 0.92);
  background: rgba(255, 255, 255, 0.68);
  color: #4b92cf;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 10px 22px rgba(107, 162, 210, 0.1);
}

.ssl-switch.active {
  background: linear-gradient(135deg, #62bfeb, #7dd3e6 56%, #f4aab1);
  box-shadow: 0 0 0 6px rgba(116, 211, 235, 0.18);
}

.insight-card {
  background:
    radial-gradient(circle at 96% 0%, rgba(255, 196, 206, 0.34), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.8), rgba(243, 251, 255, 0.62));
}

.insight-orb {
  background: radial-gradient(circle, rgba(106, 214, 245, 0.34) 0%, rgba(255, 184, 196, 0.22) 42%, transparent 72%);
}

.insight-item,
.insight-footer {
  border-color: rgba(255, 255, 255, 0.7);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(248, 253, 255, 0.6));
}

.template-item.expanded,
.template-item:hover {
  border-color: rgba(122, 207, 237, 0.72);
  box-shadow:
    0 18px 36px rgba(82, 174, 225, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.editor-shell {
  border: 1px solid rgba(160, 194, 243, 0.18);
  border-radius: 28px;
  background:
    radial-gradient(circle at 18% 0%, rgba(99, 102, 241, 0.22), transparent 26%),
    radial-gradient(circle at 92% 10%, rgba(91, 205, 232, 0.16), transparent 30%),
    linear-gradient(180deg, #0f172a 0%, #0b1220 100%);
  box-shadow:
    0 28px 54px rgba(15, 23, 42, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.workspace-tab.active-tab,
.language-pill.active {
  background: linear-gradient(135deg, #2f85ff 0%, #6078ff 58%, #76d7f5 100%);
  box-shadow:
    0 14px 30px rgba(53, 132, 255, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.24);
}

.editor-filebar,
.language-strip,
.dark-field,
.hint-chip {
  border-color: rgba(122, 150, 196, 0.28);
  background: rgba(10, 18, 35, 0.68);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.language-pill {
  border-color: rgba(122, 150, 196, 0.3);
  background: rgba(15, 25, 46, 0.72);
}

.file-badge {
  background: rgba(82, 184, 235, 0.18);
  color: #83e2ff;
}

.code-editor-card {
  border-color: rgba(122, 150, 196, 0.26);
  background: #0a1220;
}

.editor-title {
  background: linear-gradient(180deg, rgba(15, 26, 48, 0.96), rgba(11, 18, 32, 0.96));
}

.code-gutter {
  background: #08111f;
}

.code-editor textarea {
  background:
    linear-gradient(180deg, rgba(12, 22, 40, 0.98), rgba(8, 16, 29, 0.98));
  color: #d9edff;
}

.editor-shell {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background:
    radial-gradient(circle at 5% 8%, rgba(105, 202, 255, 0.3), transparent 30%),
    radial-gradient(circle at 88% 4%, rgba(255, 202, 219, 0.22), transparent 34%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.86), rgba(238, 249, 255, 0.7) 54%, rgba(255, 242, 247, 0.62));
  box-shadow:
    0 24px 52px rgba(104, 170, 220, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(22px) saturate(1.18);
}

.workspace-tab {
  border-color: rgba(214, 232, 245, 0.88);
  background: rgba(255, 255, 255, 0.7);
  color: #63788e;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.workspace-tab.active-tab {
  border-color: rgba(255, 255, 255, 0.76);
  background: linear-gradient(135deg, #2f8df7 0%, #56c4ef 64%, #9ce8ff 100%);
  color: #ffffff;
  box-shadow:
    0 14px 28px rgba(58, 158, 231, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.36);
}

.editor-filebar,
.language-strip,
.dark-field,
.hint-chip {
  border-color: rgba(214, 232, 245, 0.88);
  background: rgba(255, 255, 255, 0.62);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 12px 24px rgba(126, 177, 217, 0.08);
}

.file-badge {
  background: rgba(52, 168, 224, 0.14);
  color: #2389cc;
}

.editor-file-main strong,
.dark-field .field-title,
.editor-title {
  color: #183854;
}

.editor-file-main small,
.editor-mode,
.editor-path,
.dark-field small,
.hint-chip {
  color: #6f8398;
}

.language-pill {
  border-color: rgba(214, 232, 245, 0.9);
  background: rgba(255, 255, 255, 0.68);
  color: #17334f;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.language-pill small {
  color: #2f94d4;
}

.language-pill.active {
  border-color: rgba(255, 255, 255, 0.8);
  background: linear-gradient(135deg, #2f8df7 0%, #5fc9ef 68%, #a6ecff 100%);
  color: #ffffff;
}

.language-pill.active small {
  color: rgba(255, 255, 255, 0.84);
}

.ghost-pill {
  background: rgba(255, 255, 255, 0.72);
  color: #2587c8;
}

.dark-field input {
  border-color: rgba(201, 224, 240, 0.92);
  background: rgba(255, 255, 255, 0.74);
  color: #17334f;
}

.code-editor-card {
  border-color: rgba(214, 232, 245, 0.9);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.74), rgba(246, 252, 255, 0.64));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.82),
    0 14px 30px rgba(126, 177, 217, 0.1);
}

.editor-title {
  border-bottom-color: rgba(214, 232, 245, 0.84);
  background: rgba(255, 255, 255, 0.68);
}

.code-gutter {
  background: rgba(232, 244, 252, 0.74);
  border-right-color: rgba(207, 226, 240, 0.86);
  color: #7a91a7;
}

.code-editor textarea {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(246, 252, 255, 0.72));
  color: #17334f;
}

.template-editor {
  border-radius: 32px;
}

.editor-shell {
  isolation: isolate;
  position: relative;
  padding: 20px;
  border-radius: 32px;
  border-color: rgba(255, 255, 255, 0.9);
  background:
    radial-gradient(circle at 8% 18%, rgba(119, 211, 255, 0.34), transparent 34%),
    radial-gradient(circle at 82% 16%, rgba(255, 190, 207, 0.34), transparent 36%),
    radial-gradient(circle at 52% 92%, rgba(223, 238, 255, 0.58), transparent 42%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.94) 0%, rgba(243, 251, 255, 0.82) 52%, rgba(255, 244, 248, 0.78) 100%);
  box-shadow:
    0 28px 58px rgba(116, 173, 218, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.98),
    inset 0 -1px 0 rgba(188, 214, 236, 0.42);
}

.editor-shell::before {
  content: '';
  position: absolute;
  inset: 1px;
  z-index: -1;
  border-radius: inherit;
  background:
    linear-gradient(90deg, rgba(231, 248, 255, 0.48), transparent 42%, rgba(255, 231, 238, 0.52)),
    linear-gradient(180deg, rgba(255, 255, 255, 0.64), rgba(255, 255, 255, 0.2));
  pointer-events: none;
}

.editor-shell::after {
  content: '';
  position: absolute;
  inset: 84px 18px auto;
  z-index: -1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(132, 178, 218, 0.34), rgba(255, 178, 198, 0.32), transparent);
  pointer-events: none;
}

.editor-dots span {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.58),
    0 8px 16px rgba(127, 155, 190, 0.12);
}

.workspace-tab,
.editor-filebar,
.language-strip,
.language-pill,
.dark-field,
.hint-chip,
.code-editor-card {
  border-color: rgba(255, 255, 255, 0.82);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(248, 253, 255, 0.58));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 14px 26px rgba(118, 169, 211, 0.1);
  backdrop-filter: blur(16px) saturate(1.12);
}

.workspace-tab {
  min-height: 42px;
  border-radius: 999px;
  color: #74879a;
}

.workspace-tab.active-tab,
.language-pill.active {
  border-color: rgba(255, 255, 255, 0.86);
  background:
    radial-gradient(circle at 18% 16%, rgba(255, 255, 255, 0.6), transparent 28%),
    linear-gradient(135deg, #2f8df7 0%, #4ebaf0 56%, #86dcff 100%);
  color: #ffffff;
  box-shadow:
    0 16px 30px rgba(58, 158, 231, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.editor-filebar {
  border-radius: 24px;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.84), rgba(241, 250, 255, 0.7) 56%, rgba(255, 244, 248, 0.64));
}

.file-badge {
  background: linear-gradient(135deg, rgba(207, 240, 255, 0.9), rgba(255, 224, 235, 0.72));
  color: #2189cf;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.language-strip {
  padding: 8px;
  border-radius: 24px;
  background:
    linear-gradient(90deg, rgba(237, 250, 255, 0.72), rgba(255, 239, 245, 0.54));
}

.language-pill {
  min-height: 82px;
  border-radius: 22px;
  color: #193853;
}

.ghost-pill {
  color: #2385ca;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(241, 250, 255, 0.62));
}

.dark-field {
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(244, 251, 255, 0.64));
}

.dark-field input {
  border-color: rgba(218, 233, 245, 0.94);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(249, 253, 255, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 10px 20px rgba(118, 169, 211, 0.08);
}

.hint-chip {
  color: #60768d;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.74), rgba(239, 250, 255, 0.62));
}

.code-editor-card {
  border-radius: 28px;
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.84), rgba(243, 251, 255, 0.68) 60%, rgba(255, 245, 248, 0.58));
}

.editor-title {
  color: #183854;
  border-bottom-color: rgba(213, 229, 242, 0.78);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.82), rgba(241, 250, 255, 0.62), rgba(255, 242, 247, 0.52));
}

.code-editor {
  min-height: 500px;
}

.code-gutter {
  background: rgba(235, 247, 254, 0.76);
  color: #8398ac;
}

.code-editor textarea {
  min-height: 500px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(246, 252, 255, 0.78) 58%, rgba(255, 249, 251, 0.74));
  color: #18314c;
}

.editor-footer {
  padding: 12px 2px 0;
  border-top: 1px solid rgba(215, 230, 242, 0.58);
}

.workspace-tab.active-tab,
.language-pill.active {
  border-color: rgba(255, 255, 255, 0.9);
  background:
    radial-gradient(circle at 20% 22%, rgba(130, 112, 238, 0.48), transparent 18%),
    radial-gradient(circle at 82% 18%, rgba(116, 216, 246, 0.58), transparent 32%),
    linear-gradient(135deg, rgba(81, 159, 245, 0.92) 0%, rgba(95, 200, 237, 0.82) 58%, rgba(255, 213, 226, 0.74) 100%);
  color: #ffffff;
  box-shadow:
    0 16px 30px rgba(85, 169, 222, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.46),
    inset 0 -1px 0 rgba(80, 148, 214, 0.16);
}

.file-badge {
  background:
    radial-gradient(circle at 78% 20%, rgba(255, 206, 222, 0.72), transparent 38%),
    linear-gradient(135deg, rgba(214, 244, 255, 0.92), rgba(255, 232, 239, 0.78));
  color: #1987cf;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.78),
    0 10px 18px rgba(116, 174, 214, 0.1);
}
</style>





