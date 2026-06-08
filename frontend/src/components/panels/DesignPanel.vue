<script setup>
import { ref } from 'vue'
import { useUiStore } from '../../stores/uiStore'
import {
  Layers, Sparkles, Wand2, Move3D, Type, Palette, TimerReset,
} from 'lucide-vue-next'

const uiStore = useUiStore()
const selectedMode = ref('motion')

const kits = [
  { id: 'motion', label: 'Motion', hint: 'Keyframe-ready motion presets', accent: 'from-pink-500/15 to-fuchsia-500/5' },
  { id: 'text', label: 'Text FX', hint: 'Glow, shimmer, and type animation', accent: 'from-amber-400/15 to-orange-400/5' },
  { id: 'composite', label: 'Composite', hint: 'Blend, mask, and layer stack tools', accent: 'from-cyan-400/15 to-sky-500/5' },
]

const presets = {
  motion: [
    'Slide In',
    'Zoom Pulse',
    'Orbit Drift',
    'Bounce Pop',
  ],
  text: [
    'Neon Glow',
    'Shimmer Title',
    'Typewriter',
    'Soft Drop',
  ],
  composite: [
    'Light Leak',
    'Duotone Wash',
    'Soft Mask',
    'Blend Stack',
  ],
}

function applyPreset(name) {
  uiStore.addToast(`Applied ${name} preset`, 'success', 1400)
}
</script>

<template>
  <div class="p-3 space-y-3 text-text-primary">
    <div class="rounded-xl border border-border bg-bg/80 p-3 shadow-sm shadow-black/10">
      <div class="flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-text-secondary">
        <Sparkles :size="12" class="text-accent" />
        Design studio
      </div>
      <p class="mt-2 text-[12px] text-text-secondary">
        A lightweight creative workspace for animation, text effects, and composite layering.
      </p>
      <div class="mt-3 flex flex-wrap gap-2">
        <button
          v-for="kit in kits"
          :key="kit.id"
          class="rounded-lg border px-2.5 py-1.5 text-[11px] transition-colors"
          :class="selectedMode === kit.id
            ? 'border-accent/60 bg-accent/10 text-accent'
            : 'border-border text-text-secondary hover:border-accent/40 hover:text-text-primary'"
          @click="selectedMode = kit.id"
        >
          {{ kit.label }}
        </button>
      </div>
    </div>

    <div class="grid gap-2">
      <button
        v-for="preset in presets[selectedMode]"
        :key="preset"
        class="rounded-xl border border-border bg-panel/90 p-3 text-left transition hover:border-accent/40 hover:bg-accent/5"
        @click="applyPreset(preset)"
      >
        <div class="flex items-center gap-2 text-[12px] font-medium text-text-primary">
          <component
            :is="selectedMode === 'motion' ? Move3D : selectedMode === 'text' ? Type : Layers"
            :size="13"
            class="text-accent"
          />
          {{ preset }}
        </div>
        <div class="mt-1 text-[10px] text-text-secondary">
          {{ selectedMode === 'motion' ? 'Smooth keyframe motion with easing' : selectedMode === 'text' ? 'Styling, glow, and character timing' : 'Blend modes and masks for layered compositing' }}
        </div>
      </button>
    </div>

    <div class="rounded-xl border border-border bg-panel/90 p-3">
      <div class="flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-text-secondary">
        <Palette :size="12" class="text-accent" />
        Creative stack
      </div>
      <div class="mt-3 grid gap-2 text-[11px] text-text-secondary">
        <div class="flex items-center justify-between rounded-lg bg-bg/70 px-2.5 py-2">
          <span>Animation curves</span>
          <strong class="text-text-primary">Ready</strong>
        </div>
        <div class="flex items-center justify-between rounded-lg bg-bg/70 px-2.5 py-2">
          <span>Preset library</span>
          <strong class="text-text-primary">8 packs</strong>
        </div>
        <div class="flex items-center justify-between rounded-lg bg-bg/70 px-2.5 py-2">
          <span>Live preview</span>
          <strong class="text-text-primary">On</strong>
        </div>
      </div>
    </div>

    <button class="w-full rounded-xl border border-dashed border-accent/40 bg-accent/5 px-3 py-2 text-[11px] text-accent transition hover:bg-accent/10">
      <span class="flex items-center justify-center gap-2">
        <Wand2 :size="12" />
        Create custom animation pass
      </span>
    </button>
  </div>
</template>
