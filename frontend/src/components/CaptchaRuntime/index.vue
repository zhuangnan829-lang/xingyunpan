<template>
  <div v-if="config?.required" class="captcha-runtime">
    <div v-if="config.provider === 'image'" class="captcha-image-row">
      <el-input
        v-model.trim="answer"
        class="captcha-input"
        size="large"
        maxlength="8"
        placeholder="请输入图形验证码"
        clearable
        @input="emitPayload"
      />
      <button class="captcha-image-button" type="button" @click="reloadChallenge">
        <img v-if="challenge?.image_data_url" :src="challenge.image_data_url" alt="图形验证码" />
        <span>刷新</span>
      </button>
    </div>

    <div v-else-if="config.provider === 'slider'" class="captcha-slider">
      <input
        v-model.number="sliderValue"
        class="slider-input"
        type="range"
        min="0"
        :max="challenge?.width || 280"
        @input="recordSlider"
      />
      <span class="slider-label">拖动完成验证</span>
    </div>

    <div v-else class="captcha-third-party">
      <div
        ref="thirdPartyRef"
        class="third-party-box"
        :data-provider="config.provider"
        :data-sitekey="config.site_key"
      >
        <span>{{ config.provider === 'turnstile' ? 'Turnstile' : 'reCAPTCHA' }}</span>
        <small>{{ config.site_key }}</small>
      </div>
      <el-input
        v-model.trim="token"
        size="large"
        placeholder="第三方验证码 token"
        clearable
        @input="emitPayload"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue';
import {
  createCaptchaChallenge,
  getCaptchaConfig,
  type CaptchaChallenge,
  type CaptchaPayload,
  type CaptchaPublicConfig,
  type CaptchaScene,
} from '@/api/captcha';

const props = defineProps<{
  scene: CaptchaScene;
  identity?: string;
  path?: string;
}>();

const emit = defineEmits<{
  (event: 'update:payload', value: CaptchaPayload): void;
  (event: 'required-change', value: boolean): void;
}>();

const config = ref<CaptchaPublicConfig | null>(null);
const challenge = ref<CaptchaChallenge | null>(null);
const answer = ref('');
const token = ref('');
const sliderValue = ref(0);
const sliderTrack = ref<number[]>([]);
const thirdPartyRef = ref<HTMLElement | null>(null);

async function loadConfig() {
  config.value = await getCaptchaConfig(props.scene, props.identity || '', props.path || '');
  emit('required-change', !!config.value.required);
  if (config.value.required && (config.value.provider === 'image' || config.value.provider === 'slider')) {
    await reloadChallenge();
  } else {
    challenge.value = null;
  }
  if (config.value.required && (config.value.provider === 'turnstile' || config.value.provider === 'recaptcha')) {
    await nextTick();
    void renderThirdParty();
  }
  emitPayload();
}

async function reloadChallenge() {
  if (!config.value?.required) return;
  challenge.value = await createCaptchaChallenge(props.scene, props.identity || '', props.path || '');
  answer.value = '';
  sliderValue.value = 0;
  sliderTrack.value = [];
  emitPayload();
}

function recordSlider() {
  sliderTrack.value.push(Number(sliderValue.value));
  answer.value = String(sliderValue.value);
  emitPayload();
}

function emitPayload() {
  if (!config.value?.required) {
    emit('update:payload', {});
    return;
  }
  if (config.value.provider === 'turnstile' || config.value.provider === 'recaptcha') {
    emit('update:payload', { captcha_token: token.value });
    return;
  }
  emit('update:payload', {
    captcha_id: challenge.value?.captcha_id,
    captcha_answer: answer.value,
    slider_track: config.value.provider === 'slider' ? sliderTrack.value : undefined,
  });
}

async function renderThirdParty() {
  if (!config.value || !thirdPartyRef.value || !config.value.site_key) return;
  try {
    if (config.value.provider === 'turnstile') {
      await loadScript('https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit');
      const turnstile = (window as any).turnstile;
      if (turnstile?.render) {
        thirdPartyRef.value.innerHTML = '';
        turnstile.render(thirdPartyRef.value, {
          sitekey: config.value.site_key,
          callback: (value: string) => {
            token.value = value;
            emitPayload();
          },
        });
      }
      return;
    }
    await loadScript('https://www.google.com/recaptcha/api.js?render=explicit');
    const grecaptcha = (window as any).grecaptcha;
    if (grecaptcha?.render) {
      thirdPartyRef.value.innerHTML = '';
      grecaptcha.render(thirdPartyRef.value, {
        sitekey: config.value.site_key,
        callback: (value: string) => {
          token.value = value;
          emitPayload();
        },
      });
    }
  } catch {
    // The manual token input below remains available when third-party scripts are blocked.
  }
}

function loadScript(src: string): Promise<void> {
  const existing = document.querySelector<HTMLScriptElement>(`script[src="${src}"]`);
  if (existing) return Promise.resolve();
  return new Promise((resolve, reject) => {
    const script = document.createElement('script');
    script.src = src;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`failed to load ${src}`));
    document.head.appendChild(script);
  });
}

watch(() => props.identity, () => {
  void loadConfig();
});

onMounted(() => {
  void loadConfig();
});

defineExpose({ reload: reloadChallenge });
</script>

<style scoped>
.captcha-runtime {
  margin-bottom: 20px;
}

.captcha-image-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 158px;
  gap: 12px;
}

.captcha-image-button,
.third-party-box {
  min-height: 56px;
  border: 1px solid rgba(203, 213, 225, 0.84);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255,255,255,.96), rgba(246,250,255,.94));
}

.captcha-image-button {
  display: grid;
  place-items: center;
  padding: 4px;
  cursor: pointer;
}

.captcha-image-button img {
  max-width: 148px;
  height: 54px;
}

.captcha-image-button span {
  color: #8290aa;
  font-size: 12px;
}

.captcha-slider {
  display: grid;
  gap: 8px;
}

.slider-input {
  width: 100%;
  accent-color: #2962ea;
}

.slider-label,
.third-party-box small {
  color: #8290aa;
  font-size: 12px;
}

.third-party-box {
  display: grid;
  gap: 4px;
  align-content: center;
  padding: 10px 14px;
  margin-bottom: 10px;
}

.third-party-box span {
  color: #173f9b;
  font-weight: 700;
}

@media (max-width: 768px) {
  .captcha-image-row {
    grid-template-columns: 1fr;
  }
}
</style>
