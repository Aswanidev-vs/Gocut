<script setup>
import { Sparkles, Wand2, Box, Play, ArrowRight, Lightbulb } from 'lucide-vue-next'

const emit = defineEmits(['start-template', 'start-blank', 'start-effect'])

const cards = [
  {
    id: 'template',
    icon: Sparkles,
    title: 'Start from a Template',
    desc: 'Pre-built animations & text effects. Just pick one and customize — no node wiring needed.',
    badge: 'Easiest',
    accent: 'from-pink-500/20 to-pink-500/5 border-pink-500/30',
    iconColor: 'text-pink-400',
    action: 'start-template',
  },
  {
    id: 'effect',
    icon: Wand2,
    title: 'Add an Effect',
    desc: 'Apply glow, blur, shadow, or color grading with simple sliders. Great for quick enhancements.',
    badge: 'Quick',
    accent: 'from-amber-500/20 to-amber-500/5 border-amber-500/30',
    iconColor: 'text-amber-400',
    action: 'start-effect',
  },
  {
    id: 'blank',
    icon: Box,
    title: 'Build with Nodes',
    desc: 'Full node-based compositing for advanced users. Connect sources, transforms, and effects manually.',
    badge: 'Advanced',
    accent: 'from-cyan-500/20 to-cyan-500/5 border-cyan-500/30',
    iconColor: 'text-cyan-400',
    action: 'start-blank',
  },
]
</script>

<template>
  <div class="flex-1 flex items-center justify-center p-6 overflow-y-auto">
    <div class="max-w-2xl w-full">
      <!-- Welcome header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-accent/10 border border-accent/20 text-accent text-[11px] font-medium mb-4">
          <Sparkles :size="12" />
          Design Workspace
        </div>
        <h2 class="text-2xl font-bold text-text-primary mb-2">What do you want to create?</h2>
        <p class="text-sm text-text-secondary max-w-md mx-auto">
          Pick an option below to get started. You can always switch between simple and advanced modes later.
        </p>
      </div>

      <!-- Option cards -->
      <div class="grid gap-3">
        <button
          v-for="card in cards"
          :key="card.id"
          class="group relative flex items-center gap-4 p-4 rounded-xl border bg-gradient-to-br transition-all text-left hover:scale-[1.01] hover:shadow-lg"
          :class="[card.accent, 'hover:border-accent/50']"
          @click="emit(card.action)"
        >
          <!-- Icon -->
          <div class="w-12 h-12 rounded-lg bg-bg/60 border border-border/60 flex items-center justify-center flex-shrink-0">
            <component :is="card.icon" :size="22" :class="card.iconColor" />
          </div>

          <!-- Text -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-0.5">
              <span class="text-sm font-bold text-text-primary">{{ card.title }}</span>
              <span class="text-[9px] uppercase font-bold px-1.5 py-0.5 rounded bg-bg/80 text-text-secondary border border-border/60">{{ card.badge }}</span>
            </div>
            <p class="text-[11px] text-text-secondary leading-relaxed">{{ card.desc }}</p>
          </div>

          <!-- Arrow -->
          <ArrowRight :size="16" class="text-text-secondary/40 group-hover:text-accent group-hover:translate-x-0.5 transition-all flex-shrink-0" />
        </button>
      </div>

      <!-- Tip -->
      <div class="mt-6 flex items-start gap-2 p-3 rounded-lg bg-bg/40 border border-border/40">
        <Lightbulb :size="14" class="text-amber-400 flex-shrink-0 mt-0.5" />
        <p class="text-[11px] text-text-secondary leading-relaxed">
          <strong class="text-text-primary">Tip:</strong> The Design workspace is separate from the Edit timeline.
          Changes here create reusable compositions you can apply to your clips.
          Switch back to <strong class="text-text-primary">Edit</strong> at the bottom bar anytime to see your timeline.
        </p>
      </div>
    </div>
  </div>
</template>