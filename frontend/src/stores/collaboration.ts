// Collaboration Store for managing file sharing with permission management
import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import {
  addCollaborator as apiAddCollaborator,
  getCollaborators as apiGetCollaborators,
  updateCollaboratorPermission as apiUpdateCollaboratorPermission,
  removeCollaborator as apiRemoveCollaborator,
  getMyCollaborations as apiGetMyCollaborations,
  checkFilePermission as apiCheckFilePermission,
  type AddCollaboratorRequest,
  type Collaborator,
  type CollaborationFile,
  type PermissionCheck,
  type PermissionLevel
} from '@/api/collaboration';
import {
  validateAddCollaborator,
  validateUpdatePermission,
  validateRemoveCollaborator,
  handleAddCollaboratorError,
  handleUpdatePermissionError,
  handleRemoveCollaboratorError,
  handleCollaborationError,
  getPermissionDescription as getPermDesc
} from '@/utils/collaboration-validation';

export const useCollaborationStore = defineStore('collaboration', () => {
  // State
  const collaborators: Ref<Map<string, Collaborator[]>> = ref(new Map());
  const myCollaborations: Ref<CollaborationFile[]> = ref([]);
  const permissions: Ref<Map<string, PermissionCheck>> = ref(new Map());
  const isLoading: Ref<boolean> = ref(false);

  /**
   * Add a collaborator to a file
   * @param fileId - File ID
   * @param username - Username of the collaborator
   * @param permission - Permission level (view/download/edit)
   */
  async function addCollaborator(
    fileId: string,
    username: string,
    permission: PermissionLevel
  ): Promise<void> {
    isLoading.value = true;
    try {
      // Validate add collaborator parameters
      validateAddCollaborator(fileId, username, permission);

      const request: AddCollaboratorRequest = {
        file_id: fileId,
        username,
        permission
      };

      const collaborator = await apiAddCollaborator(request);

      // Update local state
      const fileCollaborators = collaborators.value.get(fileId) || [];
      collaborators.value.set(fileId, [...fileCollaborators, collaborator]);
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleAddCollaboratorError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Load collaborators for a file
   * @param fileId - File ID
   */
  async function loadCollaborators(fileId: string): Promise<void> {
    isLoading.value = true;
    try {
      const fileCollaborators = await apiGetCollaborators(fileId);
      collaborators.value.set(fileId, fileCollaborators);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Update collaborator permission
   * @param fileId - File ID
   * @param userId - User ID of the collaborator
   * @param permission - New permission level
   */
  async function updatePermission(
    fileId: string,
    userId: string,
    permission: PermissionLevel
  ): Promise<void> {
    isLoading.value = true;
    try {
      // Validate update permission parameters
      validateUpdatePermission(fileId, userId, permission);

      await apiUpdateCollaboratorPermission(fileId, userId, { permission });

      // Update local state
      const fileCollaborators = collaborators.value.get(fileId);
      if (fileCollaborators) {
        const collaborator = fileCollaborators.find(c => c.user_id === userId);
        if (collaborator) {
          collaborator.permission = permission;
        }
      }

      // Clear cached permission check to force reload
      permissions.value.delete(fileId);
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleUpdatePermissionError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Remove a collaborator from a file
   * @param fileId - File ID
   * @param userId - User ID of the collaborator to remove
   */
  async function removeCollaborator(fileId: string, userId: string): Promise<void> {
    isLoading.value = true;
    try {
      // Validate remove collaborator parameters
      validateRemoveCollaborator(fileId, userId);

      await apiRemoveCollaborator(fileId, userId);

      // Update local state
      const fileCollaborators = collaborators.value.get(fileId);
      if (fileCollaborators) {
        const updatedCollaborators = fileCollaborators.filter(c => c.user_id !== userId);
        collaborators.value.set(fileId, updatedCollaborators);
      }
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleRemoveCollaboratorError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Load files shared with current user (collaborations)
   */
  async function loadMyCollaborations(): Promise<void> {
    isLoading.value = true;
    try {
      myCollaborations.value = await apiGetMyCollaborations();
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Check current user's permission for a file
   * @param fileId - File ID
   * @returns Permission check result
   */
  async function checkPermission(fileId: string): Promise<PermissionCheck> {
    // Return cached permission if available
    const cached = permissions.value.get(fileId);
    if (cached) {
      return cached;
    }

    isLoading.value = true;
    try {
      const permission = await apiCheckFilePermission(fileId);
      permissions.value.set(fileId, permission);
      return permission;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Check if current user has a specific permission for a file
   * @param fileId - File ID
   * @param action - Action to check (view/download/edit)
   * @returns True if user has permission
   */
  function hasPermission(fileId: string, action: 'view' | 'download' | 'edit'): boolean {
    const permission = permissions.value.get(fileId);
    if (!permission) {
      return false;
    }

    switch (action) {
      case 'view':
        return permission.can_view;
      case 'download':
        return permission.can_download;
      case 'edit':
        return permission.can_edit;
      default:
        return false;
    }
  }

  /**
   * Check if current user is the owner of a file
   * @param fileId - File ID
   * @returns True if user is owner
   */
  function isOwner(fileId: string): boolean {
    const permission = permissions.value.get(fileId);
    return permission?.is_owner || false;
  }

  /**
   * Get collaborators for a file from local state
   * @param fileId - File ID
   * @returns Array of collaborators or empty array
   */
  function getCollaborators(fileId: string): Collaborator[] {
    return collaborators.value.get(fileId) || [];
  }

  /**
   * Get collaborator count for a file
   * @param fileId - File ID
   * @returns Number of collaborators
   */
  function getCollaboratorCount(fileId: string): number {
    const fileCollaborators = collaborators.value.get(fileId);
    return fileCollaborators ? fileCollaborators.length : 0;
  }

  /**
   * Clear collaborators for a file from local state
   * @param fileId - File ID
   */
  function clearCollaborators(fileId: string): void {
    collaborators.value.delete(fileId);
    permissions.value.delete(fileId);
  }

  /**
   * Clear all collaboration data from local state
   */
  function clearAllCollaborations(): void {
    collaborators.value.clear();
    myCollaborations.value = [];
    permissions.value.clear();
  }

  /**
   * Get permission level description
   * @param permission - Permission level
   * @returns Human-readable description
   */
  function getPermissionDescription(permission: PermissionLevel): string {
    return getPermDesc(permission);
  }

  return {
    // State
    collaborators,
    myCollaborations,
    permissions,
    isLoading,

    // Actions
    addCollaborator,
    loadCollaborators,
    updatePermission,
    removeCollaborator,
    loadMyCollaborations,
    checkPermission,

    // Helpers
    hasPermission,
    isOwner,
    getCollaborators,
    getCollaboratorCount,
    clearCollaborators,
    clearAllCollaborations,
    getPermissionDescription
  };
});
