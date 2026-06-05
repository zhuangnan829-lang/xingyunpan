export interface OAuthApp {
  id: string;
  name: string;
  description: string;
  appName: string;
  iconPath: string;
  clientId: string;
  clientSecret: string;
  redirectUris: string[];
  scopes: string[];
  isSystem: boolean;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
  tokenTtl: string;
  refreshTokenTtlSeconds: number;
  permissions: OAuthPermission[];
}

export interface OAuthAppDraft {
  name: string;
  description: string;
  redirectUri: string;
  scopes: string[];
}

export interface OAuthMetric {
  label: string;
  value: string;
  detail: string;
  tone: 'sky' | 'mint' | 'coral' | 'violet';
}

export interface OAuthPermission {
  id: string;
  title: string;
  description: string;
  scope: string;
  enabled: boolean;
  required?: boolean;
  mode?: 'read' | 'write';
  modeSwitch?: boolean;
  icon: 'document' | 'offline' | 'user' | 'lock' | 'folder' | 'share' | 'task' | 'finance' | 'dav' | 'admin';
}
