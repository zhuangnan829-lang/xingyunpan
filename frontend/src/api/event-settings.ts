import request from './request';
import type { EventCategory } from '@/views/admin/settings/components/event-settings.data';

export interface EventSettingsPayload {
  categories: EventCategory[];
  events: Record<string, boolean>;
}

export function getEventSettings(): Promise<EventSettingsPayload> {
  return request.get<EventSettingsPayload>('/api/v1/admin/event-settings');
}

export function updateEventSettings(data: Pick<EventSettingsPayload, 'events'>): Promise<EventSettingsPayload> {
  return request.put<EventSettingsPayload>('/api/v1/admin/event-settings', data);
}

export function resetEventSettings(): Promise<EventSettingsPayload> {
  return request.post<EventSettingsPayload>('/api/v1/admin/event-settings/reset');
}

export function toggleAllEventSettings(enabled: boolean): Promise<EventSettingsPayload> {
  return request.post<EventSettingsPayload>('/api/v1/admin/event-settings/toggle-all', { enabled });
}

export function toggleEventCategory(categoryKey: string, enabled: boolean): Promise<EventSettingsPayload> {
  return request.post<EventSettingsPayload>(`/api/v1/admin/event-settings/categories/${encodeURIComponent(categoryKey)}`, {
    enabled,
  });
}

export function toggleEventSetting(eventKey: string, enabled: boolean): Promise<EventSettingsPayload> {
  return request.patch<EventSettingsPayload>(`/api/v1/admin/event-settings/events/${encodeURIComponent(eventKey)}`, {
    enabled,
  });
}
