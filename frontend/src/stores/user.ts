/**
 * User store - Pinia state management for user authentication and profile
 */

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type {
  UserProfile,
  UserPreferences,
  LoginRequest,
  RegisterRequest,
  ChangePasswordRequest,
  UpdateProfileRequest,
} from '@/types/user';
import * as userAPI from '@/api/user';
import { saveToken, getToken, removeToken } from '@/utils/auth';

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref<string | null>(getToken());
  const profile = ref<UserProfile | null>(null);
  const preferences = ref<UserPreferences | null>(null);
  const isAuthenticated = ref<boolean>(!!token.value);

  // Getters
  /**
   * Calculate storage usage percentage
   * @returns Storage usage percentage (0-100)
   */
  const storageUsagePercent = computed(() => {
    if (!profile.value || profile.value.capacity === 0) {
      return 0;
    }
    return Math.floor((profile.value.used_size / profile.value.capacity) * 100);
  });

  /**
   * Check if current user is admin
   * @returns True if user has admin role
   */
  const isAdmin = computed(() => {
    return profile.value?.role === 'admin';
  });

  // Actions
  /**
   * User login
   * @param credentials - Login credentials (username/email and password)
   */
  async function login(credentials: LoginRequest): Promise<void> {
    const response = await userAPI.login(credentials);
    
    // Save token
    token.value = response.access_token;
    saveToken(response.access_token);
    
    // Save user profile
    profile.value = response.user;
    isAuthenticated.value = true;
  }

  /**
   * User registration
   * @param data - Registration data (username, email, password)
   */
  async function register(data: RegisterRequest): Promise<void> {
    await userAPI.register(data);
    // Note: After registration, user needs to login separately
  }

  /**
   * User logout
   * Clears token and profile from state and localStorage
   */
  function logout(): void {
    token.value = null;
    profile.value = null;
    preferences.value = null;
    isAuthenticated.value = false;
    removeToken();
  }

  /**
   * Fetch user profile
   * Retrieves current user's profile information from API
   */
  async function fetchProfile(): Promise<void> {
    const userProfile = await userAPI.getProfile();
    profile.value = userProfile;
  }

  async function updateProfile(data: UpdateProfileRequest): Promise<void> {
    const userProfile = await userAPI.updateProfile(data);
    profile.value = userProfile;
  }

  async function uploadAvatar(file: File): Promise<void> {
    const userProfile = await userAPI.uploadAvatar(file);
    profile.value = userProfile;
  }

  async function fetchPreferences(): Promise<void> {
    preferences.value = await userAPI.getPreferences();
  }

  async function updatePreferences(data: UserPreferences): Promise<void> {
    preferences.value = await userAPI.updatePreferences(data);
  }

  /**
   * Change user password
   * @param data - Password change data (old_password, new_password)
   */
  async function changePassword(data: ChangePasswordRequest): Promise<void> {
    await userAPI.changePassword(data);
    // After password change, logout user for security
    logout();
  }

  return {
    // State
    token,
    profile,
    preferences,
    isAuthenticated,
    
    // Getters
    storageUsagePercent,
    isAdmin,
    
    // Actions
    login,
    register,
    logout,
    fetchProfile,
    updateProfile,
    uploadAvatar,
    fetchPreferences,
    updatePreferences,
    changePassword,
  };
});
