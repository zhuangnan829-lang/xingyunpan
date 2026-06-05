import { getMyCollaborations, type CollaborationFile, type PermissionLevel } from './collaboration';
import request from './request';

export interface SharedWithMeItem {
  id: string;
  file_id: string;
  file_name: string;
  file_type: string;
  file_size: number;
  owner_name: string;
  permission: PermissionLevel;
  shared_at: string;
}

type SharedWithMeSource = CollaborationFile & {
  content_type?: string;
  is_folder?: boolean;
};

export async function listSharedWithMe(): Promise<SharedWithMeItem[]> {
  try {
    const items = await request.get<SharedWithMeSource[]>('/api/v1/shared-with-me');
    return (items || []).map(normalizeSharedItem);
  } catch {
    const items = await getMyCollaborations();
    return items.map(normalizeSharedItem);
  }
}

function normalizeSharedItem(item: SharedWithMeSource): SharedWithMeItem {
  return {
    id: String(item.file_id),
    file_id: String(item.file_id),
    file_name: item.file_name || 'Untitled file',
    file_type: item.content_type || item.file_type || '',
    file_size: item.file_size || 0,
    owner_name: item.owner_name || 'Unknown user',
    permission: item.permission,
    shared_at: item.shared_at,
  };
}
