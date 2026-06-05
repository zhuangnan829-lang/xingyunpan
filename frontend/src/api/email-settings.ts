import request from './request';

export interface EmailSettingsPayload {
  enabled: boolean;
  provider: string;
  host: string;
  port: number;
  username: string;
  password: string;
  from_name: string;
  from_address: string;
  reply_to: string;
  force_ssl: boolean;
  connection_timeout: number;
  code_ttl_seconds: number;
  send_interval_seconds: number;
}

export interface EmailTemplateLanguagePayload {
  code: string;
  label: string;
  subject: string;
  content: string;
}

export interface EmailTemplatePayload {
  template_key: string;
  name: string;
  description: string;
  status: string;
  status_tone: string;
  pro: boolean;
  languages: EmailTemplateLanguagePayload[];
}

export function getEmailSettings(): Promise<EmailSettingsPayload> {
  return request.get<EmailSettingsPayload>('/api/v1/admin/email-settings');
}

export function updateEmailSettings(data: EmailSettingsPayload): Promise<EmailSettingsPayload> {
  return request.put<EmailSettingsPayload>('/api/v1/admin/email-settings', data);
}

export function sendTestEmail(toEmail: string): Promise<{ to_email: string }> {
  return request.post<{ to_email: string }>('/api/v1/admin/email-settings/test', {
    to_email: toEmail,
  });
}

export function getEmailTemplates(): Promise<EmailTemplatePayload[]> {
  return request.get<EmailTemplatePayload[]>('/api/v1/admin/email-templates');
}

export function updateEmailTemplate(templateKey: string, data: EmailTemplatePayload): Promise<EmailTemplatePayload> {
  return request.put<EmailTemplatePayload>(`/api/v1/admin/email-templates/${templateKey}`, data);
}
