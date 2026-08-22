<script setup>
// Sticker inspector — edits the selected clip's StickerProps.
//
// Mounting note: this is an isolated, self-contained component. The RightPanel
// agent owns RightPanel.vue and should render `<StickerInspector />` when the
// selected clip has `stickerProps` (or add a dedicated "Sticker" tab). It reads
// everything it needs from the timeline store, so no props are required.
//
// Binding model: each control writes through `timelineStore.updateClip`, which
// performs a shallow `Object.assign` on the clip — so we must spread the
// existing `stickerProps` when mutating a single field (same pattern as
// `setTextProp` in RightPanel.vue). This keeps the existing store action as the
// single source of truth and avoids introducing new store mutations.

import { computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { RotateCcw, Move, Maximize, RotateCw, Droplets } from 'lucide-vue-next'

const timelineStore = useTimelineStore()

const clip = computed(() => timelineStore.selectedClips[0] || null)
const stickerProps = computed(() => clip.value?.stickerProps || null)

// Defaults mirror the values used when a sticker clip is created in
// StickerPanel `addSticker` / PreviewPlayer fallback (width/height act as a
// scale factor in the live preview, not literal px).
const DEFAULTS = {
  x: 0,
  y: 0,
  width: 0.2,
  height: 0.2,
  rotation: 0,
  opacity: 1,
  flipH: false,
  flipV: false,
}

function getProp(key) {
  const sp = stickerProps.value
  if (!sp) return DEFAULTS[key]
  return sp[key] !== undefined ? sp[key] : DEFAULTS[key]
}

function setProp(key, value) {
  const c = clip.value
  if (!c || !c.stickerProps) return
  timelineStore.updateClip(c.id, {
    stickerProps: { ...c.stickerProps, [key]: value },
  })
}

// Number helpers (X / Y are px offsets, Width / Height are scale factors).
function setNumber(key, raw) {
  const n = parseFloat(raw)
  setProp(key, isNaN(n) ? DEFAULTS[key] : n)
}

// Opacity is stored 0..1 but shown as a percentage.
const opacityPct = computed(() => Math.round(getProp('opacity') * 100))
function setOpacityPct(raw) {
  const pct = parseFloat(raw)
  const clamped = Math.max(0, Math.min(100, isNaN(pct) ? 100 : pct))
  setProp('opacity', clamped / 100)
}

function toggleFlip(key) {
  setProp(key, !getProp(key))
}

function resetSticker() {
  const c = clip.value
  if (!c || !c.stickerProps) return
  // Merge transform defaults over the existing props so identity fields
  // (svg / name / imagePath) set by addSticker are preserved — the live
  // preview renders from `svg`, so dropping it would hide the sticker.
  timelineStore.updateClip(c.id, {
    stickerProps: { ...c.stickerProps, ...DEFAULTS },
  })
}
</script>

<template>
  <div v-if="stickerProps" class="flex flex-col gap-4">
    <!-- Position -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider flex items-center gap-1">
          <Move :size="11" /> Position
        </h4>
        <button class="p-0.5 rounded text-text-secondary hover:text-accent" @click="resetSticker" title="Reset sticker">
          <RotateCcw :size="10" />
        </button>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="text-[10px] text-text-secondary block mb-1">X (px)</label>
          <input
            type="number"
            :value="getProp('x')"
            @input="(e) => setNumber('x', e.target.value)"
            class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
          />
        </div>
        <div>
          <label class="text-[10px] text-text-secondary block mb-1">Y (px)</label>
          <input
            type="number"
            :value="getProp('y')"
            @input="(e) => setNumber('y', e.target.value)"
            class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
          />
        </div>
      </div>
    </div>

    <hr class="border-border" />

    <!-- Size (scale) -->
    <div>
      <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider flex items-center gap-1 mb-2">
        <Maximize :size="11" /> Size
      </h4>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="text-[10px] text-text-secondary block mb-1">Width</label>
          <input
            type="number" step="0.05" min="0.05"
            :value="getProp('width')"
            @input="(e) => setNumber('width', e.target.value)"
            class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
          />
        </div>
        <div>
          <label class="text-[10px] text-text-secondary block mb-1">Height</label>
          <input
            type="number" step="0.05" min="0.05"
            :value="getProp('height')"
            @input="(e) => setNumber('height', e.target.value)"
            class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
          />
        </div>
      </div>
    </div>

    <hr class="border-border" />

    <!-- Rotation -->
    <div>
      <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider flex items-center gap-1 mb-2">
        <RotateCw :size="11" /> Rotation
      </h4>
      <div class="flex items-center gap-2">
        <input
          type="range" min="-180" max="180" step="1"
          :value="getProp('rotation')"
          @input="(e) => setProp('rotation', parseFloat(e.target.value))"
          class="flex-1 accent-accent"
        />
        <input
          type="number"
          :value="getProp('rotation')"
          @input="(e) => setNumber('rotation', e.target.value)"
          class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
        />
      </div>
    </div>

    <hr class="border-border" />

    <!-- Opacity -->
    <div>
      <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider flex items-center gap-1 mb-2">
        <Droplets :size="11" /> Opacity
      </h4>
      <div class="flex items-center gap-2">
        <input
          type="range" min="0" max="100" step="1"
          :value="opacityPct"
          @input="(e) => setOpacityPct(e.target.value)"
          class="flex-1 accent-accent"
        />
        <input
          type="number" min="0" max="100"
          :value="opacityPct"
          @input="(e) => setOpacityPct(e.target.value)"
          class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono"
        />
      </div>
    </div>

    <hr class="border-border" />

    <!-- Flip -->
    <div>
      <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Flip</h4>
      <div class="flex items-center gap-2">
        <button
          class="px-2 py-1 rounded text-xs border transition-colors"
          :class="getProp('flipH') ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'"
          @click="toggleFlip('flipH')"
        >
          Flip H
        </button>
        <button
          class="px-2 py-1 rounded text-xs border transition-colors"
          :class="getProp('flipV') ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'"
          @click="toggleFlip('flipV')"
        >
          Flip V
        </button>
      </div>
    </div>

    <hr class="border-border" />

    <!-- Reset -->
    <button
      class="w-full py-1.5 rounded text-xs font-medium border border-border text-text-secondary hover:text-accent hover:border-accent/50 transition-colors flex items-center justify-center gap-1.5"
      @click="resetSticker"
    >
      <RotateCcw :size="12" /> Reset Sticker
    </button>
  </div>
</template>
