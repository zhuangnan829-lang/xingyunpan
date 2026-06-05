// Collaboration API module for file sharing with permission management

import request from './request';

/**
 * Permission level type
 */
export type PermissionLevel = 'view' | 'download' | 'edit';

/**
 * Request interface for adding a collaborator
 */
export interface AddCollaboratorRequest {
  file_id: string;
  username: string;
  permission: PermissionLevel;
}

/**
 * Interface for collaborator information
 */
export interface Collaborator {
  user_id: string;
  username: string;
  permission: PermissionLevel;
  added_at: string;
}

/**
 * Request interface for updating collaborator permission
 */
export interface UpdatePermissionRequest {
  permission: PermissionLevel;
}

/**
 * Interface for collaboration file (files shared with user)
 */
export interface CollaborationFile {
  file_id: string;
  file_name: string;
  file_type: string;
  file_size: number;
  owner_name: string;
  permission: PermissionLevel;
  shared_at: string;
}

/**
 * Interface for permission check result
 */
export interface PermissionCheck {
  can_view: boolean;
  can_download: boolean;
  can_edit: boolean;
  is_owner: boolean;
}

/**
 * Add a collaborator to a file
 * @param req - Collaborator addition request
 * @returns Added collaborator information
 */
export async function addCollaborator(req: AddCollaboratorRequest): Promise<Collaborator> {
  return request.post('/api/v1/collaborations', req);
}

/**
 * Get collaborators for a file
 * @param fileId - File ID
 * @returns List of collaborators
 */
export async function getCollaborators(fileId: string): Promise<Collaborator[]> {
  return request.get(`/api/v1/files/${fileId}/collaborators`);
}

/**
 * Update collaborator permission
 * @param fileId - File ID
 * @param userId - User ID of the collaborator
 * @param req - Permission update request
 */
export async function updateCollaboratorPermission(
  fileId: string,
  userId: string,
  req: UpdatePermissionRequest
): Promise<void> {
  return request.put(`/api/v1/files/${fileId}/collaborators/${userId}`, req);
}

/**
 * Remove a collaborator from a file
 * @param fileId - File ID
 * @param userId - User ID of the collaborator to remove
 */
export async function removeCollaborator(fileId: string, userId: string): Promise<void> {
  return request.delete(`/api/v1/files/${fileId}/collaborators/${userId}`);
}

/**
 * Get files shared with current user (collaborations)
 * @returns List of collaboration files
 */
export async function getMyCollaborations(): Promise<CollaborationFile[]> {
  return request.get('/api/v1/collaborations/me');
}

/**
 * Check current user's permission for a file
 * @param fileId - File ID
 * @returns Permission check result
 */
export async function checkFilePermission(fileId: string): Promise<PermissionCheck> {
  return request.get(`/api/v1/files/${fileId}/permissions`);
}
