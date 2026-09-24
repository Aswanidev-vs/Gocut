<script setup>
// Empty state. The point of this screen is that nobody should have to learn what
// a Merge or a MediaOut is before they can make something move.
import { computed } from 'vue'
import { RECIPES, RECIPE_GROUPS } from '../../lib/designRecipes'
import { ArrowRight, GitBranch } from 'lucide-vue-next'

const emit = defineEmits(['pick', 'blank'])

// A handful of starters spread across the groups, so the first click is always
// a move rather than an empty graph.
const starters = computed(() => {
  const perGroup = Math.ceil(8 / RECIPE_GROUPS.length)
  return RECIPE_GROUPS
    .flatMap(group => RECIPES.filter(r => r.group === group).slice(0, perGroup))
    .slice(0, 8)
})
</script>

<template>
  <div class="flex-1 flex items-center justify-center overflow-y-auto bg-cr-void p-6">
    <div class="w-full max-w-3xl">
      <div class="border-b border-cr-line pb-3 mb-4">
        <div class="text-[10px] uppercase tracking-[0.18em] text-ink-faint">Start here</div>
        <h2 class="text-lg text-ink mt-1">Pick a move. Tune it. Press play.</h2>
        <p class="text-[11px] text-ink-dim mt-1 leading-relaxed">
          Each move builds its own node graph with the animation already keyframed — open Flow when you
          want to see or rewire it.
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-px bg-cr-line border border-cr-line">
        <button
          v-for="recipe in starters"
          :key="recipe.id"
          class="text-left bg-cr-panel hover:bg-white/5 px-3 py-2.5 transition-colors group border-l-2 border-transparent hover:border-accent"
          @click="emit('pick', recipe.id)"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="text-[12px] text-ink group-hover:text-accent transition-colors">{{ recipe.name }}</span>
            <ArrowRight :size="12" class="text-ink-faint group-hover:text-accent transition-colors" />
          </div>
          <div class="text-[10px] text-ink-faint mt-0.5 leading-snug">{{ recipe.blurb }}</div>
          <div class="text-[9px] uppercase tracking-[0.14em] text-ink-faint/70 mt-1.5">{{ recipe.group }}</div>
        </button>
      </div>

      <div class="mt-4 flex items-center justify-between gap-3 border-t border-cr-line pt-3">
        <span class="text-[10px] text-ink-faint">Prefer to wire it yourself?</span>
        <button
          class="flex items-center gap-1.5 px-2.5 py-1 text-[10px] uppercase tracking-[0.14em] border border-cr-line text-ink-dim hover:text-ink hover:border-accent/40 transition-colors"
          @click="emit('blank')"
        >
          <GitBranch :size="11" /> Blank graph
        </button>
      </div>
    </div>
  </div>
</template>