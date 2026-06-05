import { getFileSystemMapProvider, type FileSystemMapRuntimeConfig } from '@/api/file-system-settings';

export type LeafletLike = {
  tileLayer: (url: string, options: { attribution: string }) => { addTo: (map: unknown) => unknown };
};

export type MapboxLike = {
  accessToken?: string;
  Map: new (options: { container: unknown; style: string; accessToken?: string }) => unknown;
};

export type MapRuntimeApplyResult =
  | {
      engine: 'leaflet';
      provider: FileSystemMapRuntimeConfig['provider'];
      tileUrl: string;
      attribution: string;
      tokenMissing: false;
    }
  | {
      engine: 'mapbox';
      provider: FileSystemMapRuntimeConfig['provider'];
      styleUrl: string;
      attribution: string;
      tokenMissing: boolean;
    };

export async function loadMapRuntimeConfig(): Promise<FileSystemMapRuntimeConfig> {
  return normalizeMapRuntimeConfig(await getFileSystemMapProvider());
}

export function normalizeMapRuntimeConfig(config: FileSystemMapRuntimeConfig): FileSystemMapRuntimeConfig {
  if (config.engine === 'mapbox') {
    return {
      ...config,
      provider: 'osm-mapbox',
      style_url: config.style_url || 'mapbox://styles/mapbox/streets-v12',
      tile_url: '',
      token_missing: config.requires_token && !getMapboxAccessToken(),
    };
  }

  const provider = config.provider === 'google-leaflet' ? 'google-leaflet' : 'osm-leaflet';
  return {
    ...config,
    provider,
    engine: 'leaflet',
    tile_url:
      config.tile_url ||
      (provider === 'google-leaflet'
        ? 'https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}'
        : 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'),
    style_url: '',
    token_missing: false,
  };
}

export function applyLeafletRuntime(map: unknown, leaflet: LeafletLike, config: FileSystemMapRuntimeConfig): MapRuntimeApplyResult {
  const normalized = normalizeMapRuntimeConfig(config);
  if (normalized.engine !== 'leaflet' || !normalized.tile_url) {
    throw new Error('map runtime config is not a leaflet tile provider');
  }

  leaflet.tileLayer(normalized.tile_url, { attribution: normalized.attribution }).addTo(map);
  return {
    engine: 'leaflet',
    provider: normalized.provider,
    tileUrl: normalized.tile_url,
    attribution: normalized.attribution,
    tokenMissing: false,
  };
}

export function createMapboxRuntime(container: unknown, mapbox: MapboxLike, config: FileSystemMapRuntimeConfig): MapRuntimeApplyResult {
  const normalized = normalizeMapRuntimeConfig(config);
  if (normalized.engine !== 'mapbox' || !normalized.style_url) {
    throw new Error('map runtime config is not a mapbox style provider');
  }

  const accessToken = getMapboxAccessToken();
  if (accessToken) {
    mapbox.accessToken = accessToken;
  }
  new mapbox.Map({
    container,
    style: normalized.style_url,
    accessToken: accessToken || undefined,
  });

  return {
    engine: 'mapbox',
    provider: normalized.provider,
    styleUrl: normalized.style_url,
    attribution: normalized.attribution,
    tokenMissing: normalized.requires_token && !accessToken,
  };
}

export function getMapboxAccessToken(): string {
  const envToken = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
  if (typeof envToken === 'string' && envToken.trim()) {
    return envToken.trim();
  }
  if (typeof window !== 'undefined') {
    const runtimeToken = (window as unknown as { __MAPBOX_ACCESS_TOKEN__?: string }).__MAPBOX_ACCESS_TOKEN__;
    if (typeof runtimeToken === 'string' && runtimeToken.trim()) {
      return runtimeToken.trim();
    }
  }
  return '';
}
