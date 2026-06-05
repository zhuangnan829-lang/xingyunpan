export type BlobKind = 'version' | 'thumbnail' | 'live-photo' | 'file' | 'orphan';

export interface BlobLinkedFile {
  id: number;
  name: string;
  extension: string;
  sizeBytes: number;
  ownerId: number;
  ownerName: string;
  createdAt: string;
}

export interface BlobLinkedVersion {
  id: number;
  fileId: number;
  fileName: string;
  versionNumber: number;
  sizeBytes: number;
  ownerId: number;
  ownerName: string;
  createdAt: string;
}

export interface BlobReferenceSource {
  type: string;
  id: string;
  name: string;
}

export interface BlobRecord {
  id: number;
  kind: BlobKind;
  kindLabel: string;
  source: string;
  hash?: string;
  contentType?: string;
  sizeBytes: number;
  referenceCount: number;
  storedReferenceCount?: number;
  createdAt: string;
  updatedAt?: string;
  creatorId: number;
  creatorName: string;
  creatorBadge: string;
  storagePolicyId: number;
  storagePolicyName: string;
  storagePolicySubtitle: string;
  encrypted: boolean;
  locked?: boolean;
  lockedReason?: string;
  canDelete?: boolean;
  deleteBlockedReasons?: string[];
  missingOnStorage?: boolean;
  healthStatus?: string;
  uploadSessionId?: string;
  linkedFiles: BlobLinkedFile[];
  linkedVersions?: BlobLinkedVersion[];
  referenceSources?: BlobReferenceSource[];
}

export interface BlobFilterState {
  ownerId: string;
  kind: 'all' | BlobKind;
  storagePolicyId: number | 'all';
  keyword: string;
  minSize: string;
  maxSize: string;
  refCountMin: string;
  refCountMax: string;
  encrypted: 'all' | 'true' | 'false';
  createdFrom: string;
  createdTo: string;
}
