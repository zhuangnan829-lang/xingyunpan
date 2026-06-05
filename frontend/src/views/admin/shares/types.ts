import type { AdminSharePayload } from '@/api/admin-shares';

export type ShareStatusFilter = 'all' | 'active' | 'expired' | 'protected' | 'unavailable' | 'download_limit_reached';

export interface ShareDisplayRecord extends AdminSharePayload {
  index: number;
  sourceLabel: string;
  ownerName: string;
  statusText: string;
  expiryText: string;
  createdText: string;
  score: number;
  isExpired: boolean;
  isUnavailable: boolean;
  isProtected: boolean;
}

export interface ShareFilters {
  keyword: string;
  status: ShareStatusFilter;
  minDownloads: number | null;
  expiringOnly: boolean;
}

export interface ShareMetric {
  label: string;
  value: string;
  detail: string;
  tone: string;
}
