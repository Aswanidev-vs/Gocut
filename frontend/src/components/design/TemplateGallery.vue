<script setup>
import { ref, computed } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { Play, Plus } from 'lucide-vue-next'

const emit = defineEmits(['insert'])
const designStore = useDesignStore()
const search = ref('')

const templates = [
  {
    name: 'Fade In Title',
    desc: 'Text that fades in with opacity',
    cat: 'Text Intro',
    create: () => {
      const t = designStore.addNode('text', { x: 100, y: 150, label: 'Title' })
      designStore.addNode('output', { x: 500, y: 150, label: 'Output' })
      if (t) {
        designStore.addKeyframe(t.id, 'opacity', 0, 0, 'easeOut')
        designStore.addKeyframe(t.id, 'opacity', 1, 1, 'linear')
      }
    }
  },
  {
    name: 'Slide In from Left',
    desc: 'Text or rectangle slides in from offscreen',
    cat: 'Motion',
    create: () => {
      const t = designStore.addNode('text', { x: 100, y: 200, label: 'Slide Text' })
      designStore.addNode('output', { x: 500, y: 200 })
      if (t) {
        designStore.addKeyframe(t.id, 'x', 0, -400, 'easeOut')
        designStore.addKeyframe(t.id, 'x', 1.5, 0, 'easeInOut')
        designStore.addKeyframe(t.id, 'opacity', 0, 0, 'linear')
        designStore.addKeyframe(t.id, 'opacity', 0.3, 1, 'easeOut')
      }
    }
  },
  {
    name: 'Bounce Pop',
    desc: 'Scale bounce effect',
    cat: 'Motion',
    create: () => {
      const t = designStore.addNode('text', { x: 100, y: 250, label: 'Bounce', params: { fontSize: 64 } })
      designStore.addNode('output', { x: 500, y: 250 })
      if (t) {
        designStore.addKeyframe(t.id, 'scaleX', 0, 0.3, 'easeIn')
        designStore.addKeyframe(t.id, 'scaleY', 0, 0.3, 'easeIn')
        designStore.addKeyframe(t.id, 'scaleX', 0.8, 1.3, 'easeOut')
        designStore.addKeyframe(t.id, 'scaleY', 0.8, 1.3, 'easeOut')
        designStore.addKeyframe(t.id, 'scaleX', 1.2, 1, 'bounce')
        designStore.addKeyframe(t.id, 'scaleY', 1.2, 1, 'bounce')
      }
    }
  },
  {
    name: 'Neon Glow Pulse',
    desc: 'Glow intensity pulsing over time',
    cat: 'Effects',
    create: () => {
      const t = designStore.addNode('text', { x: 100, y: 300, label: 'NEON' })
      const g = designStore.addNode('glow', { x: 350, y: 300, label: 'Glow' })
      const o = designStore.addNode('output', { x: 600, y: 300 })
      if (t && g && o) {
        designStore.addConnection(t.id, 'out', g.id, 'in')
        designStore.addConnection(g.id, 'out', o.id, 'in')
        designStore.addKeyframe(g.id, 'intensity', 0, 0.2, 'easeInOut')
        designStore.addKeyframe(g.id, 'intensity', 1, 2.0, 'easeInOut')
        designStore.addKeyframe(g.id, 'intensity', 2, 0.2, 'easeInOut')
      }
    }
  },
  {
    name: 'Chroma Key Demo',
    desc: 'Green screen keying setup',
    cat: 'Effects',
    create: () => {
      const src = designStore.addNode('media', { x: 100, y: 350, label: 'Video' })
      const ck = designStore.addNode('chromaKey', { x: 350, y: 350 })
      const out = designStore.addNode('output', { x: 600, y: 350 })
      if (src && ck) designStore.addConnection(src.id, 'out', ck.id, 'in')
      if (ck && out) designStore.addConnection(ck.id, 'out', out.id, 'in')
    }
  },
  {
    name: 'Quick Composite',
    desc: 'Two sources merged with blend',
    cat: 'Composite',
    create: () => {
      const a = designStore.addNode('text', { x: 100, y: 400, label: 'Layer A' })
      const b = designStore.addNode('rectangle', { x: 100, y: 470, label: 'Layer B' })
      const m = designStore.addNode('merge', { x: 400, y: 435 })
      const o = designStore.addNode('output', { x: 650, y: 435 })
      if (a && b && m && o) {
        designStore.addConnection(a.id, 'out', m.id, 'bg')
        designStore.addConnection(b.id, 'out', m.id, 'fg')
        designStore.addConnection(m.id, 'out', o.id, 'in')
      }
    }
  },
]

function applyTemplate(tpl) {
  tpl.create()
  emit('insert')
}

const filteredTemplates = computed(() => {
  if (!search.value) return templates
  const q = search.value.toLowerCase()
  return templates.filter(t => t.name.toLowerCase().includes(q) || t.desc.toLowerCase().includes(q) || t.cat.toLowerCase().includes(q))
})
</script>

<template>
  <div class="p-2 space-y-2">
    <input
      v-model="search"
      type="text"
      placeholder="Search templates…"
      class="w-full bg-bg border border-border rounded px-2.5 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent placeholder:text-text-secondary/60"
    />
    <div v-for="(tpl, i) in filteredTemplates" :key="i" class="space-y-1">
      <div class="text-[10px] text-text-secondary uppercase tracking-wider px-1">{{ tpl.cat }}</div>
      <button
        class="w-full rounded-lg border border-border/60 bg-panel/60 p-2.5 text-left transition hover:border-accent/40 hover:bg-accent/5"
        @click="applyTemplate(tpl)"
      >
        <div class="flex items-center justify-between">
          <div class="text-[12px] text-text-primary font-medium">{{ tpl.name }}</div>
          <Plus :size="11" class="text-accent" />
        </div>
        <div class="text-[10px] text-text-secondary mt-0.5">{{ tpl.desc }}</div>
      </button>
    </div>
  </div>
</template>