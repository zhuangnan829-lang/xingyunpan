/**
 * Version management utility functions
 * Provides helper functions for version formatting, comparison, and storage calculations
 */

import type { FileVersion } from '@/api/version'

/**
 * Format version number for display
 * @param versionNumber - The version number (1, 2, 3, etc.)
 * @returns Formatted version string
 */
export function formatVersionNumber(versionNumber: number): string {
  return `v${versionNumber}`
}

/**
 * Compare two versions
 * @param versionA - First version to compare
 * @param versionB - Second version to compare
 * @returns Negative if A < B, 0 if equal, positive if A > B
 */
export function compareVersions(versionA: FileVersion, versionB: FileVersion): number {
  return versionB.version_number - versionA.version_number
}

/**
 * Calculate total storage used by all versions
 * @param versions - Array of file versions
 * @returns Total storage in bytes
 */
export function calculateVersionStorage(versions: FileVersion[]): number {
  return versions.reduce((total, version) => total + version.file_size, 0)
}
