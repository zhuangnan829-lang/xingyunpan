import type { AdminUserPayload } from '@/api/admin-users';
import type { StoragePolicyPayload } from '@/api/storage-policy';
import type { BlobRecord } from './types';

function fallbackPolicy(): StoragePolicyPayload {
  return {
    id: 1,
    name: '默认存储策略',
    type: 'local',
    groups: [],
    blob_path: '/cloudreve/data/uploads/{uid}/{path}',
    blob_name_pattern: '{uid}_{randomkey8}_{originname}',
    max_file_size: 1024,
    max_file_size_unit: 'MB',
    extension_mode: 'allow',
    extensions: '',
    name_rule_mode: 'allow',
    name_regex: '',
    chunk_size: 20,
    chunk_size_unit: 'MB',
    pre_allocate: false,
    parallel_chunk_count: 3,
    enable_cdn: false,
    download_cdn: '',
    enable_encryption: false,
    encryption_key_id: '',
  };
}

function fallbackUser(): AdminUserPayload {
  return {
    id: 3518974413,
    username: '3518974413',
    email: 'demo@example.com',
    role: 'admin',
    enabled: true,
    user_group_id: 1,
    user_group_name: '默认用户组',
    capacity: 1024 * 1024 * 1024 * 20,
    used_size: 1024 * 1024 * 1024 * 6,
    created_at: '2026-05-19T19:10:00+08:00',
    updated_at: '2026-05-19T19:10:00+08:00',
  };
}

export function createMockBlobRecords(users: AdminUserPayload[], policies: StoragePolicyPayload[]): BlobRecord[] {
  const creator = users[0] ?? fallbackUser();
  const secondUser = users[1] ?? {
    ...fallbackUser(),
    id: 3518974414,
    username: 'xingyunpan-admin',
    email: 'admin@example.com',
  };

  const defaultPolicy = policies[0] ?? fallbackPolicy();
  const archivePolicy = policies[1] ?? {
    ...fallbackPolicy(),
    id: 2,
    name: '归档存储策略',
  };

  return [
    {
      id: 3,
      kind: 'live-photo',
      kindLabel: '实时照片',
      source: '/cloudreve/data/uploads/2/live/2026/05/18/realtime-photo-cover.mov',
      sizeBytes: 4.5 * 1024 * 1024,
      referenceCount: 2,
      createdAt: '2026-05-18T15:21:10+08:00',
      creatorId: secondUser.id,
      creatorName: secondUser.username,
      creatorBadge: 'A',
      storagePolicyId: Number(archivePolicy.id ?? 2),
      storagePolicyName: archivePolicy.name,
      storagePolicySubtitle: '归档存储策略',
      encrypted: true,
      uploadSessionId: 'live_20260518_152110',
      linkedFiles: [
        {
          id: 18,
          name: '实时封面.heic',
          extension: 'H',
          sizeBytes: 2.1 * 1024 * 1024,
          ownerId: secondUser.id,
          ownerName: secondUser.username,
          createdAt: '2026-05-18T15:21:03+08:00',
        },
        {
          id: 19,
          name: '实时封面.mov',
          extension: 'M',
          sizeBytes: 2.4 * 1024 * 1024,
          ownerId: secondUser.id,
          ownerName: secondUser.username,
          createdAt: '2026-05-18T15:21:10+08:00',
        },
      ],
    },
    {
      id: 2,
      kind: 'thumbnail',
      kindLabel: '缩略图',
      source: '/cloudreve/data/uploads/1/1_tox9MCuM-ui-preview-thumb.webp',
      sizeBytes: 21 * 1024,
      referenceCount: 1,
      createdAt: '2026-05-19T19:18:01+08:00',
      creatorId: creator.id,
      creatorName: creator.username,
      creatorBadge: '3',
      storagePolicyId: Number(defaultPolicy.id ?? 1),
      storagePolicyName: defaultPolicy.name,
      storagePolicySubtitle: '默认存储策略',
      encrypted: false,
      linkedFiles: [
        {
          id: 2,
          name: '星云盘首页视觉方案.fig',
          extension: 'W',
          sizeBytes: 20 * 1024,
          ownerId: creator.id,
          ownerName: creator.username,
          createdAt: '2026-05-19T19:17:54+08:00',
        },
      ],
    },
    {
      id: 1,
      kind: 'version',
      kindLabel: '版本',
      source: '/cloudreve/data/uploads/1/1_tox9MCuM-ui-design-v12.fig',
      sizeBytes: 20 * 1024,
      referenceCount: 1,
      createdAt: '2026-05-19T19:17:54+08:00',
      creatorId: creator.id,
      creatorName: creator.username,
      creatorBadge: '3',
      storagePolicyId: Number(defaultPolicy.id ?? 1),
      storagePolicyName: defaultPolicy.name,
      storagePolicySubtitle: '默认存储策略',
      encrypted: false,
      uploadSessionId: 'upload_20260519_191754',
      linkedFiles: [
        {
          id: 1,
          name: '界面设计迭代稿.fig',
          extension: 'W',
          sizeBytes: 20 * 1024,
          ownerId: creator.id,
          ownerName: creator.username,
          createdAt: '2026-05-19T19:17:54+08:00',
        },
      ],
    },
  ];
}
