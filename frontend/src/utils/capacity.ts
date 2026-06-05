export type CapacityUnit = 'MB' | 'GB' | 'TB';

const CAPACITY_MULTIPLIERS: Record<CapacityUnit, number> = {
  MB: 1024 ** 2,
  GB: 1024 ** 3,
  TB: 1024 ** 4,
};

export function capacityToBytes(value: number, unit: CapacityUnit, unlimited = false): number {
  if (unlimited || !Number.isFinite(value) || value <= 0) {
    return 0;
  }

  return Math.round(value * CAPACITY_MULTIPLIERS[unit]);
}

export function bytesToCapacity(bytes: number): { value: number; unit: CapacityUnit; unlimited: boolean } {
  if (!bytes || bytes <= 0) {
    return { value: 50, unit: 'GB', unlimited: true };
  }

  if (bytes % CAPACITY_MULTIPLIERS.TB === 0) {
    return { value: Math.max(1, bytes / CAPACITY_MULTIPLIERS.TB), unit: 'TB', unlimited: false };
  }

  if (bytes % CAPACITY_MULTIPLIERS.GB === 0) {
    return { value: Math.max(1, bytes / CAPACITY_MULTIPLIERS.GB), unit: 'GB', unlimited: false };
  }

  return { value: Math.max(1, Math.round(bytes / CAPACITY_MULTIPLIERS.MB)), unit: 'MB', unlimited: false };
}
