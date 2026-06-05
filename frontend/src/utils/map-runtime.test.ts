import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  applyLeafletRuntime,
  createMapboxRuntime,
  loadMapRuntimeConfig,
  normalizeMapRuntimeConfig,
  type LeafletLike,
  type MapboxLike,
} from './map-runtime';
import { getFileSystemMapProvider, type FileSystemMapRuntimeConfig } from '@/api/file-system-settings';

vi.mock('@/api/file-system-settings', () => ({
  getFileSystemMapProvider: vi.fn(),
}));

const mockedGetMapProvider = vi.mocked(getFileSystemMapProvider);

function leafletConfig(provider: 'google-leaflet' | 'osm-leaflet', tileUrl: string): FileSystemMapRuntimeConfig {
  return {
    provider,
    engine: 'leaflet',
    tile_url: tileUrl,
    attribution: provider,
    requires_token: false,
  };
}

describe('map runtime provider consumption', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    delete (window as unknown as { __MAPBOX_ACCESS_TOKEN__?: string }).__MAPBOX_ACCESS_TOKEN__;
  });

  it('loads fresh runtime config for the next page entry', async () => {
    mockedGetMapProvider
      .mockResolvedValueOnce(leafletConfig('osm-leaflet', 'https://osm.test/{z}/{x}/{y}.png'))
      .mockResolvedValueOnce(leafletConfig('google-leaflet', 'https://google.test/vt?x={x}&y={y}&z={z}'));

    await expect(loadMapRuntimeConfig()).resolves.toMatchObject({ provider: 'osm-leaflet' });
    await expect(loadMapRuntimeConfig()).resolves.toMatchObject({ provider: 'google-leaflet' });
    expect(mockedGetMapProvider).toHaveBeenCalledTimes(2);
  });

  it('uses OSM tile_url in the Leaflet page adapter', () => {
    const tileLayer = vi.fn(() => ({ addTo: vi.fn() }));
    const leaflet = { tileLayer } satisfies LeafletLike;

    const result = applyLeafletRuntime('map', leaflet, leafletConfig('osm-leaflet', 'https://osm.test/{z}/{x}/{y}.png'));

    expect(result).toMatchObject({ engine: 'leaflet', tileUrl: 'https://osm.test/{z}/{x}/{y}.png' });
    expect(tileLayer).toHaveBeenCalledWith('https://osm.test/{z}/{x}/{y}.png', { attribution: 'osm-leaflet' });
  });

  it('uses Google tile_url in the Leaflet page adapter', () => {
    const tileLayer = vi.fn(() => ({ addTo: vi.fn() }));
    const leaflet = { tileLayer } satisfies LeafletLike;

    const result = applyLeafletRuntime('map', leaflet, leafletConfig('google-leaflet', 'https://google.test/vt?x={x}&y={y}&z={z}'));

    expect(result).toMatchObject({ engine: 'leaflet', tileUrl: 'https://google.test/vt?x={x}&y={y}&z={z}' });
    expect(tileLayer).toHaveBeenCalledWith('https://google.test/vt?x={x}&y={y}&z={z}', { attribution: 'google-leaflet' });
  });

  it('uses Mapbox style_url and exposes missing-token state', () => {
    const Map = vi.fn();
    const mapbox = { Map } as unknown as MapboxLike;
    const config = normalizeMapRuntimeConfig({
      provider: 'osm-mapbox',
      engine: 'mapbox',
      style_url: 'mapbox://styles/example/custom',
      attribution: 'mapbox',
      requires_token: true,
    });

    const result = createMapboxRuntime('container', mapbox, config);

    expect(result).toMatchObject({
      engine: 'mapbox',
      styleUrl: 'mapbox://styles/example/custom',
      tokenMissing: true,
    });
    expect(Map).toHaveBeenCalledWith({
      container: 'container',
      style: 'mapbox://styles/example/custom',
      accessToken: undefined,
    });
  });
});
