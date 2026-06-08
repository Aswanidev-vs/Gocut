<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/uiStore'
import { Video, Music, Type, Smile, Sparkles, ArrowRightLeft, Search, Palette } from 'lucide-vue-next'
import AssetPool from '../panels/AssetPool.vue'
import TextPanel from '../panels/TextPanel.vue'
import StickerPanel from '../panels/StickerPanel.vue'
import TransitionsPanel from '../panels/TransitionsPanel.vue'
import AudioPanel from '../panels/AudioPanel.vue'
import DesignPanel from '../panels/DesignPanel.vue'

const uiStore = useUiStore()

const tabs = [
  { id: 'media',       label: 'Media',       icon: Video },
  { id: 'audio',       label: 'Audio',       icon: Music },
  { id: 'text',        label: 'Text',        icon: Type },
  { id: 'stickers',    label: 'Stickers',    icon: Smile },
  { id: 'fx',          label: 'Effects',     icon: Sparkles },
  // { id: 'design',      label: 'Design',      icon: Palette },
  { id: 'transitions', label: 'Transitions', icon: ArrowRightLeft },
]

const activeComponent = computed(() => {
  switch (uiStore.activePanelTab) {
    case 'media':       return AssetPool
    case 'audio':       return AudioPanel
    case 'text':        return TextPanel
    case 'stickers':    return StickerPanel
    case 'fx':          return TransitionsPanel
    case 'design':      return DesignPanel
    case 'transitions': return TransitionsPanel
    default:            return AssetPool
  }
})
</script>

<template>
  <div class="bg-panel border-r border-border flex flex-col overflow-hidden flex-shrink-0">
    <!-- Tab bar -->
    <div class="flex items-center gap-0.5 p-1.5 border-b border-border overflow-x-auto">
      <button
        v-for="t in tabs"
        :key="t.id"
        class="flex flex-col items-center justify-center gap-0.5 px-1.5 py-1.5 rounded text-[10px] transition-colors min-w-[44px] flex-shrink-0"
        :class="uiStore.activePanelTab === t.id
          ? 'bg-accent/10 text-accent'
          : 'text-text-secondary hover:text-text-primary hover:bg-border/50'"
        :title="t.label"
        @click="uiStore.setActivePanelTab(t.id)"
      >
        <component :is="t.icon" :size="14" />
        <span class="leading-none">{{ t.label }}</span>
      </button>
    </div>

    <!-- Search bar (some panels) -->
    <div v-if="['media', 'audio', 'text', 'stickers', 'design', 'transitions'].includes(uiStore.activePanelTab)"
         class="px-2 pt-2">
      <div class="relative">
        <Search :size="12" class="absolute left-2 top-1/2 -translate-y-1/2 text-text-secondary pointer-events-none" />
        <input
          type="text"
          placeholder="Search…"
          class="w-full bg-bg border border-border rounded pl-7 pr-2 py-1.5 text-xs text-text-primary outline-none focus:border-accent placeholder:text-text-secondary/60"
        />
      </div>
    </div>

    <!-- Active panel content -->
    <div class="flex-1 overflow-y-auto">
      <component :is="activeComponent" />
    </div>
  </div>
</template>
