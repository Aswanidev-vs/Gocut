<script setup>
// The Design rail: Moves (recipes), Looks and the full node list, driven from
// the recipe library so the language users see stays human.
import { computed, ref } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useUiStore } from '../../stores/uiStore'
import {
  RECIPES, RECIPE_GROUPS, LOOK_RECIPES, applyRecipe, applyLook,
} from '../../lib/designRecipes'
import { Search, Layers, ChevronDown, ChevronRight, Crosshair } from 'lucide-vue-next'
import NodeLibrary from './NodeLibrary.vue'
import TrackingPanel from '../tracking/TrackingPanel.vue'

const emit = defineEmits(['applied'])

const designStore = useDesignStore()
const uiStore = useUiStore()

const search = ref('')
const showAllNodes = ref(false)
const showTracking = ref(false)
const openGroups = ref(new Set(RECIPE_GROUPS))

const query = computed(() => search.value.trim().toLowerCase())

function matches(item) {
  if (!query.value) return true
  return item.name.toLowerCase().includes(query.value)
    || item.blurb.toLowerCase().includes(query.value)
}

const groups = computed(() => RECIPE_GROUPS
  .map(name => ({ name, items: RECIPES.filter(r => r.group === name && matches(r)) }))
  .filter(g => g.items.length > 0))

const looks = computed(() => LOOK_RECIPES.filter(matches))

function toggleGroup(name) {
  const next = new Set(openGroups.value)
  if (next.has(name)) next.delete(name)
  else next.add(name)
  openGroups.value = next
}

// A look with a node selected splices itself into the chain, so say which node
// it landed on — otherwise it looks like nothing happened.
function runLook(look) {
  const target = designStore.selectedNode
  const usable = target && target.type !== 'output' && target.type !== 'mediaOut'
  applyLook(designStore, look)
  uiStore.addToast(usable ? `${look.name} → ${target.label}` : `${look.name} added`, 'success', 1800)
  emit('applied')
}

function runRecipe(recipe) {
  applyRecipe(designStore, recipe)
  uiStore.addToast(`${recipe.name} added`, 'success', 1600)
  emit('applied')
}
</script>

<template>
  <div class="flex flex-col h-full bg-cr-panel min-h-0">
    <!-- Search -->
    <div class="p-2 border-b border-cr-line">
      <div class="relative">
        <Search :size="11" class="absolute left-2 top-1/2 -translate-y-1/2 text-ink-faint pointer-events-none" />
        <input
          v-model="search"
          type="text"
          placeholder="Search moves…"
          class="w-full bg-cr-void border border-cr-line pl-7 pr-2 py-1.5 text-[11px] text-ink outline-none focus:border-accent placeholder:text-ink-faint"
        />
      </div>
    </div>

    <div class="flex-1 overflow-y-auto">
      <!-- Moves -->
      <div class="px-2 py-2">
        <div class="flex items-center justify-between px-0.5 pb-1.5">
          <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">Moves</span>
          <span class="text-[9px] font-jetbrains-mono tabular-nums text-ink-faint">{{ RECIPES.length }}</span>
        </div>

        <div v-for="group in groups" :key="group.name" class="mb-1">
          <button
            class="w-full flex items-center gap-1 px-0.5 py-1 text-[10px] uppercase tracking-[0.14em] text-ink-dim hover:text-ink transition-colors"
            @click="toggleGroup(group.name)"
          >
            <ChevronDown v-if="openGroups.has(group.name)" :size="10" />
            <ChevronRight v-else :size="10" />
            <span class="flex-1 text-left">{{ group.name }}</span>
            <span class="font-jetbrains-mono tabular-nums text-[9px] text-ink-faint">{{ group.items.length }}</span>
          </button>

          <div v-show="openGroups.has(group.name)" class="space-y-px">
            <button
              v-for="recipe in group.items"
              :key="recipe.id"
              class="w-full text-left px-2 py-1.5 border-l-2 border-transparent hover:border-accent hover:bg-white/5 transition-colors group"
              @click="runRecipe(recipe)"
            >
              <div class="text-[11px] text-ink group-hover:text-accent transition-colors">{{ recipe.name }}</div>
              <div class="text-[9px] text-ink-faint leading-snug">{{ recipe.blurb }}</div>
            </button>
          </div>
        </div>

        <div v-if="groups.length === 0" class="text-[10px] text-ink-faint px-0.5 py-2">
          No move matches “{{ search }}”.
        </div>
      </div>

      <!-- Looks -->
      <div class="px-2 py-2 border-t border-cr-line">
        <div class="flex items-center justify-between px-0.5 pb-1.5">
          <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">Looks</span>
          <span class="text-[9px] font-jetbrains-mono tabular-nums text-ink-faint">{{ LOOK_RECIPES.length }}</span>
        </div>
        <div class="text-[9px] text-ink-faint px-0.5 pb-1.5 leading-snug">
          {{ designStore.selectedNode
            ? `Applies after “${designStore.selectedNode.label}”.`
            : 'Select a node first to grade its output, or click to add a demo.' }}
        </div>
        <div class="space-y-px">
          <button
            v-for="look in looks"
            :key="look.id"
            class="w-full text-left px-2 py-1.5 border-l-2 border-transparent hover:border-signal hover:bg-white/5 transition-colors group"
            @click="runLook(look)"
          >
            <div class="text-[11px] text-ink group-hover:text-signal transition-colors">{{ look.name }}</div>
            <div class="text-[9px] text-ink-faint leading-snug">{{ look.blurb }}</div>
          </button>
        </div>
      </div>

      <!-- Full node list, tucked away so it never leads the interface -->
      <div class="border-t border-cr-line">
        <button
          class="w-full flex items-center gap-1.5 px-2.5 py-2 text-[9px] uppercase tracking-[0.14em] text-ink-dim hover:text-ink transition-colors"
          @click="showAllNodes = !showAllNodes"
        >
          <Layers :size="11" />
          <span class="flex-1 text-left">All nodes</span>
          <ChevronDown v-if="showAllNodes" :size="11" />
          <ChevronRight v-else :size="11" />
        </button>
        <div v-show="showAllNodes" class="border-t border-cr-line-soft">
          <NodeLibrary />
        </div>
      </div>

      <!-- Motion tracking lives with the other tools but stays out of the way -->
      <div class="border-t border-cr-line">
        <button
          class="w-full flex items-center gap-1.5 px-2.5 py-2 text-[9px] uppercase tracking-[0.14em] text-ink-dim hover:text-ink transition-colors"
          @click="showTracking = !showTracking"
        >
          <Crosshair :size="11" />
          <span class="flex-1 text-left">Motion tracking</span>
          <ChevronDown v-if="showTracking" :size="11" />
          <ChevronRight v-else :size="11" />
        </button>
        <div v-show="showTracking" class="border-t border-cr-line-soft">
          <TrackingPanel />
        </div>
      </div>
    </div>
  </div>
</template>
