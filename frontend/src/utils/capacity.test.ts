import { describe, expect, it } from 'vitest';
import { capacityToBytes } from './capacity';

describe('capacityToBytes', () => {
  it('converts 2 GB to bytes', () => {
    expect(capacityToBytes(2, 'GB')).toBe(2147483648);
  });

  it('converts 2048 MB to the same bytes as 2 GB', () => {
    expect(capacityToBytes(2048, 'MB')).toBe(2147483648);
  });

  it('saves unlimited capacity as zero', () => {
    expect(capacityToBytes(2, 'GB', true)).toBe(0);
  });
});
