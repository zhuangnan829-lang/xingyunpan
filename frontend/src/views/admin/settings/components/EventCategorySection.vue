<template>
  <section class="event-category">
    <div class="category-head">
      <div>
        <h3>{{ category.title }}</h3>
        <p>{{ category.description }}</p>
      </div>
      <span class="category-count">{{ selectedCount }} / {{ category.items.length }}</span>
    </div>

    <label class="master-toggle">
      <input
        :checked="allChecked"
        :disabled="disabled"
        type="checkbox"
        @change="emit('toggle-category', category.key, !allChecked)"
      />
      <span class="checkmark" />
      <span class="toggle-copy">
        <strong>启用/禁用所有事件</strong>
        <small>启用或禁用此类别中的所有事件。</small>
      </span>
    </label>

    <div class="event-grid">
      <label v-for="item in category.items" :key="item.key" class="event-option">
        <input
          :checked="model[item.key]"
          :disabled="disabled"
          type="checkbox"
          @change="emit('toggle-event', item.key, !model[item.key])"
        />
        <span class="checkmark" />
        <span>{{ item.label }}</span>
      </label>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { EventCategory } from './event-settings.data';

defineProps<{
  category: EventCategory;
  model: Record<string, boolean>;
  allChecked: boolean;
  selectedCount: number;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  'toggle-category': [categoryKey: string, checked: boolean];
  'toggle-event': [eventKey: string, checked: boolean];
}>();
</script>

<style scoped>
.event-category {
  display: grid;
  gap: 22px;
  padding: 26px 0 28px;
  border-bottom: 1px solid rgba(207, 222, 232, 0.72);
}

.event-category:first-child {
  padding-top: 0;
}

.event-category:last-child {
  border-bottom: none;
  padding-bottom: 4px;
}

.category-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 16px;
  align-items: start;
}

h3,
p,
strong,
small,
span {
  margin: 0;
}

h3 {
  color: #142339;
  font-size: 21px;
  font-weight: 780;
  line-height: 1.3;
  letter-spacing: 0;
}

p,
small {
  color: #66768a;
  line-height: 1.7;
}

p {
  margin-top: 6px;
  font-size: 14px;
}

.category-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(224, 244, 255, 0.84), rgba(255, 232, 236, 0.78));
  color: #2288d6;
  font-size: 12px;
  font-weight: 760;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.master-toggle,
.event-option {
  position: relative;
  display: inline-grid;
  grid-template-columns: 20px minmax(0, 1fr);
  align-items: start;
  gap: 14px;
  color: #1f2d3d;
  cursor: pointer;
}

.master-toggle {
  display: grid;
  width: 100%;
  max-width: 620px;
  padding-left: 38px;
}

.toggle-copy {
  display: grid;
  gap: 7px;
  min-width: 0;
}

.toggle-copy strong {
  color: #142339;
  font-size: 16px;
  font-weight: 760;
  white-space: normal;
}

.event-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(160px, 1fr));
  gap: 24px 52px;
  padding-left: 38px;
}

.event-option {
  min-height: 28px;
  font-size: 15px;
  font-weight: 650;
}

input[type='checkbox'] {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.checkmark {
  position: relative;
  width: 18px;
  height: 18px;
  margin-top: 2px;
  border: 2px solid rgba(99, 113, 128, 0.72);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.76);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 8px 16px rgba(99, 155, 203, 0.08);
  transition:
    background 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.checkmark::after {
  content: '';
  position: absolute;
  left: 4px;
  top: 0;
  width: 5px;
  height: 10px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  opacity: 0;
  transform: rotate(45deg) scale(0.7);
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

input[type='checkbox']:checked + .checkmark {
  border-color: rgba(36, 137, 223, 0.96);
  background: linear-gradient(135deg, #43b5f0 0%, #4fd7dc 62%, #f4a8b3 100%);
  box-shadow:
    0 10px 20px rgba(66, 181, 230, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.58);
}

input[type='checkbox']:checked + .checkmark::after {
  opacity: 1;
  transform: rotate(45deg) scale(1);
}

input[type='checkbox']:disabled + .checkmark,
input[type='checkbox']:disabled ~ span {
  cursor: not-allowed;
  opacity: 0.58;
}

.event-option:hover .checkmark,
.master-toggle:hover .checkmark {
  border-color: rgba(36, 137, 223, 0.86);
  box-shadow:
    0 10px 20px rgba(66, 181, 230, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.84);
}

@media (max-width: 1180px) {
  .event-grid {
    grid-template-columns: repeat(3, minmax(150px, 1fr));
    gap: 22px 34px;
  }
}

@media (max-width: 780px) {
  .category-head {
    grid-template-columns: 1fr;
  }

  .event-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    padding-left: 0;
  }

  .master-toggle {
    padding-left: 0;
  }
}

@media (max-width: 520px) {
  .event-grid {
    grid-template-columns: 1fr;
  }
}
</style>
