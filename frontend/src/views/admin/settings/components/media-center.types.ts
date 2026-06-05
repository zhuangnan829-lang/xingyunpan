import type { Component } from 'vue';

export type EngineStatus = {
  name: string;
  label: string;
  ready: boolean;
  progress: number;
};

export type FeatureCardTone = 'sky' | 'violet' | 'amber' | 'emerald';

export type FeatureCardMetric = {
  label: string;
  value: string;
  width: string;
};

export type FeatureCardBadge = {
  label: string;
  active?: boolean;
};

export type FeatureCardItem = {
  title: string;
  desc: string;
  icon: Component;
};
