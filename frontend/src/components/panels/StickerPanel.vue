<script setup>
import { ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { OpenFilePicker } from '../../lib/wails'
import { Smile, Plus, Upload, X } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

// SVG stickers rendered inline so we don't need to ship PNG assets.
// Open-licensed / public-domain style glyphs drawn as simple paths.
const stickers = [
  { name: 'Heart',    svg: '<path d="M50 85 L20 55 a18 18 0 0 1 30 -10 a18 18 0 0 1 30 10 Z" fill="#EC4899" />' },
  { name: 'Star',     svg: '<polygon points="50,15 61,40 88,40 66,57 75,84 50,67 25,84 34,57 12,40 39,40" fill="#F59E0B" />' },
  { name: 'Thumb',    svg: '<path d="M30 50 h15 v30 h-15 z M45 80 v-30 l10 -25 a8 8 0 0 1 8 8 v15 h15 a8 8 0 0 1 8 10 l-5 22 a8 8 0 0 1 -8 6 h-28 z" fill="#10B981" />' },
  { name: 'Sparkle',  svg: '<path d="M50 10 L55 40 L85 45 L55 50 L50 80 L45 50 L15 45 L45 40 Z" fill="#00D4FF" />' },
  { name: 'Check',    svg: '<path d="M20 50 L40 70 L80 30" stroke="#10B981" stroke-width="12" fill="none" stroke-linecap="round" stroke-linejoin="round" />' },
  { name: 'X',        svg: '<path d="M25 25 L75 75 M75 25 L25 75" stroke="#EF4444" stroke-width="12" stroke-linecap="round" />' },
  { name: 'Arrow',    svg: '<path d="M10 50 L70 50 L60 35 M70 50 L60 65" stroke="#F59E0B" stroke-width="8" fill="none" stroke-linecap="round" stroke-linejoin="round" />' },
  { name: 'Circle',   svg: '<circle cx="50" cy="50" r="35" fill="#8B5CF6" />' },
  { name: 'Triangle', svg: '<polygon points="50,15 88,80 12,80" fill="#F59E0B" />' },
  { name: 'Square',   svg: '<rect x="20" y="20" width="60" height="60" fill="#3B82F6" />' },
  { name: 'Lightning',svg: '<polygon points="55,10 25,55 45,55 35,90 75,40 55,40 65,10" fill="#FACC15" />' },
  { name: 'Crown',    svg: '<path d="M15 70 L25 30 L40 55 L50 20 L60 55 L75 30 L85 70 Z" fill="#F59E0B" />' },
  { name: 'Fire',     svg: '<path d="M50 90 c-20 0 -25 -15 -10 -30 c-2 -15 5 -25 10 -30 c5 5 12 15 10 30 c15 15 10 30 -10 30 z" fill="#EF4444" />' },
  { name: 'Eye',      svg: '<ellipse cx="50" cy="50" rx="35" ry="20" fill="white" stroke="#888" stroke-width="3" /><circle cx="50" cy="50" r="10" fill="#0F0F0F" />' },
  { name: 'Bubble',   svg: '<path d="M15 20 h70 a8 8 0 0 1 8 8 v30 a8 8 0 0 1 -8 8 h-50 l-15 12 v-12 a8 8 0 0 1 -8 -8 v-30 a8 8 0 0 1 8 -8 z" fill="#00D4FF" />' },
  { name: '100',      svg: '<text x="50" y="62" text-anchor="middle" font-size="38" font-weight="900" fill="#FACC15">100</text>' },
  { name: 'Moon',     svg: '<path d="M62 18 a38 38 0 1 0 20 64 a30 30 0 1 1 -20 -64 z" fill="#FBBF24" />' },
  { name: 'Sun',      svg: '<circle cx="50" cy="50" r="18" fill="#F59E0B"/><g stroke="#F59E0B" stroke-width="6" stroke-linecap="round"><line x1="50" y1="8" x2="50" y2="22"/><line x1="50" y1="78" x2="50" y2="92"/><line x1="8" y1="50" x2="22" y2="50"/><line x1="78" y1="50" x2="92" y2="50"/><line x1="20" y1="20" x2="30" y2="30"/><line x1="70" y1="70" x2="80" y2="80"/><line x1="20" y1="80" x2="30" y2="70"/><line x1="70" y1="30" x2="80" y2="20"/></g>' },
  { name: 'Drop',     svg: '<path d="M50 15 C50 15 25 45 25 62 a25 25 0 0 0 50 0 C75 45 50 15 50 15 Z" fill="#38BDF8" />' },
  { name: 'Cloud',    svg: '<g fill="#94A3B8"><circle cx="38" cy="55" r="15"/><circle cx="62" cy="50" r="19"/><circle cx="70" cy="62" r="13"/><rect x="28" y="56" width="50" height="16" rx="8"/></g>' },
  { name: 'Flower',   svg: '<g fill="#F472B6"><circle cx="50" cy="30" r="12"/><circle cx="31" cy="45" r="12"/><circle cx="69" cy="45" r="12"/><circle cx="39" cy="68" r="12"/><circle cx="61" cy="68" r="12"/></g><circle cx="50" cy="50" r="12" fill="#FACC15"/>' },
  { name: 'Smile',    svg: '<circle cx="50" cy="50" r="38" fill="#FACC15"/><circle cx="38" cy="42" r="5" fill="#3F3F46"/><circle cx="62" cy="42" r="5" fill="#3F3F46"/><path d="M35 60 q15 15 30 0" stroke="#3F3F46" stroke-width="5" fill="none" stroke-linecap="round"/>' },
  { name: 'Diamond',  svg: '<polygon points="50,12 82,50 50,88 18,50" fill="#22D3EE" />' },
  { name: 'Hexagon',  svg: '<polygon points="50,12 85,32 85,68 50,88 15,68 15,32" fill="#A78BFA" />' },
  { name: 'Question', svg: '<text x="50" y="68" text-anchor="middle" font-size="52" font-weight="900" fill="#38BDF8">?</text>' },
  { name: 'Bang',     svg: '<text x="50" y="70" text-anchor="middle" font-size="56" font-weight="900" fill="#F87171">!</text>' },
  { name: 'Music',    svg: '<g fill="#34D399"><rect x="40" y="24" width="7" height="44" rx="2"/><path d="M40 24 L70 18 L70 28 L47 33 Z"/><circle cx="38" cy="68" r="9"/><circle cx="68" cy="62" r="9"/></g>' },
  { name: 'Leaf',     svg: '<path d="M25 75 C25 35 60 20 80 20 C80 60 45 75 25 75 Z" fill="#22C55E"/><path d="M30 70 C45 55 65 35 78 25" stroke="#15803D" stroke-width="3" fill="none"/>' },
  { name: 'Bell',     svg: '<path d="M50 20 a18 18 0 0 1 18 18 v18 l8 10 h-52 l8 -10 v-18 a18 18 0 0 1 18 -18 Z" fill="#FB923C"/><circle cx="50" cy="80" r="6" fill="#FB923C"/>' },
  { name: 'Camera',   svg: '<rect x="22" y="35" width="56" height="38" rx="6" fill="#64748B"/><rect x="38" y="28" width="24" height="10" rx="3" fill="#64748B"/><circle cx="50" cy="54" r="13" fill="#E2E8F0"/><circle cx="50" cy="54" r="7" fill="#64748B"/>' },
  { name: 'Snow',     svg: '<g stroke="#7DD3FC" stroke-width="5" stroke-linecap="round"><line x1="50" y1="15" x2="50" y2="85"/><line x1="22" y1="32" x2="78" y2="68"/><line x1="78" y1="32" x2="22" y2="68"/></g>' },
  { name: 'Target',   svg: '<g fill="none" stroke="#F43F5E" stroke-width="6"><circle cx="50" cy="50" r="35"/><circle cx="50" cy="50" r="20"/></g><circle cx="50" cy="50" r="7" fill="#F43F5E"/>' },
  { name: 'Ghost',    svg: '<path d="M25 80 V45 a25 25 0 0 1 50 0 V80 l-8 -8 l-8 8 l-8 -8 l-8 8 l-8 -8 Z" fill="#E5E7EB"/><circle cx="40" cy="45" r="5" fill="#1F2937"/><circle cx="60" cy="45" r="5" fill="#1F2937"/>' },
]

// Custom stickers imported by the user
const customStickers = ref([])

function addSticker(s) {
  const t = timelineStore.getTrackByType('sticker') || timelineStore.addTrack('sticker')
  const clip = timelineStore.addClip({
    assetId: '',
    trackId: t.id,
    trackType: 'sticker',
    startTime: timelineStore.currentTime,
    duration: 3,
    stickerProps: {
      x: 0, y: 0, width: 0.2, height: 0.2, rotation: 0, opacity: 1,
      flipH: false, flipV: false, svg: s.svg, name: s.name,
      imagePath: s.imagePath || null,
    },
  })
  timelineStore.selectClip(clip.id)
  uiStore.addToast(`Added ${s.name} sticker`, 'success', 1500)
}

async function importSticker() {
  try {
    const paths = await OpenFilePicker([
      { name: 'Images', extensions: ['png', 'jpg', 'jpeg', 'gif', 'webp'] }
    ])

    if (!Array.isArray(paths) || paths.length === 0) return

    for (const path of paths) {
      const name = path.split(/[/\\]/).pop() || 'Custom Sticker'
      customStickers.value.push({
        name: name.replace(/\.[^/.]+$/, ''),
        imagePath: path,
        svg: null,
        isCustom: true,
      })
    }

    uiStore.addToast(`Imported ${paths.length} sticker(s)`, 'success', 1500)
  } catch (err) {
    uiStore.addToast('Import failed: ' + (err?.message || err), 'error')
  }
}

function removeCustomSticker(index) {
  customStickers.value.splice(index, 1)
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <!-- Import button -->
    <button
      class="flex items-center justify-center gap-1.5 w-full py-2 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors"
      @click="importSticker"
    >
      <Upload :size="12" /> Import Sticker
    </button>

    <!-- Custom stickers section -->
    <div v-if="customStickers.length > 0" class="flex flex-col gap-1.5">
      <div class="text-[10px] text-text-secondary uppercase tracking-wider px-1">Custom</div>
      <div class="grid grid-cols-3 gap-1.5">
        <div
          v-for="(s, i) in customStickers"
          :key="'custom-' + i"
          class="aspect-square rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors flex items-center justify-center group relative"
        >
          <img
            :src="'file:///' + s.imagePath"
            :alt="s.name"
            class="w-10 h-10 object-contain"
            @error="(e) => e.target.style.display = 'none'"
          />
          <button
            class="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-bg border border-border flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-500/20 hover:border-red-500/50"
            @click.stop="removeCustomSticker(i)"
          >
            <X :size="8" class="text-text-secondary" />
          </button>
          <button
            class="absolute inset-0"
            @click="addSticker(s)"
            :title="s.name"
          />
        </div>
      </div>
    </div>

    <!-- Built-in stickers -->
    <div class="flex flex-col gap-1.5">
      <div class="text-[10px] text-text-secondary uppercase tracking-wider px-1">Built-in</div>
      <div class="grid grid-cols-3 gap-1.5">
        <button
          v-for="(s, i) in stickers"
          :key="i"
          class="aspect-square rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors flex items-center justify-center group"
          @click="addSticker(s)"
          :title="s.name"
        >
          <svg viewBox="0 0 100 100" class="w-10 h-10" v-html="s.svg" />
        </button>
      </div>
    </div>
  </div>
</template>
