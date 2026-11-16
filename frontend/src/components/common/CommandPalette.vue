<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[100] flex items-start justify-center pt-[12vh]">
      <div class="absolute inset-0 bg-surface-well/70 backdrop-blur-sm" @click="close" />
      <div class="relative w-full max-w-xl bg-surface-panel border border-line rounded-lg shadow-2xl overflow-hidden">
        <div class="flex items-center gap-2 px-3 py-2 border-b border-line">
          <input
            ref="inputRef"
            v-model="query"
            type="text"
            placeholder="Type a command…"
            class="flex-1 bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
            @keydown.escape="close"
            @keydown.enter="runSelected"
            @keydown.down.prevent="move(1)"
            @keydown.up.prevent="move(-1)"
          />
          <kbd class="text-[10px] text-text-muted bg-surface-raised border border-line px-1.5 py-0.5 rounded font-jetbrains-mono">Esc</kbd>
        </div>
        <div class="max-h-80 overflow-y-auto py-1">
          <button
            v-for="(cmd, i) in filtered"
            :key="cmd.id"
            class="w-full text-left px-3 py-2 text-sm flex items-center justify-between transition-colors"
            :class="i === selected ? 'bg-accent/10 text-accent' : 'text-text-primary hover:bg-surface-raised'"
            @click="run(cmd)"
            @mouseenter="selected = i"
          >
            <span>{{ cmd.label }}</span>
            <span v-if="cmd.hint" class="text-[10px] text-text-muted font-jetbrains-mono">{{ cmd.hint }}</span>
          </button>
          <div v-if="filtered.length === 0" class="px-3 py-6 text-center text-sm text-text-muted">
            No matching commands
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useUiStore } from '../../stores/uiStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'

const props = defineProps({ open: { type: Boolean, default: false } })
const emit = defineEmits(['close'])

const uiStore = useUiStore()
const timelineStore = useTimelineStore()
const projectStore = useProjectStore()

const query = ref('')
const selected = ref(0)
const inputRef = ref(null)

const commands = [
  { id: 'new-project', label: 'New Project', hint: 'Ctrl+N', run: () => uiStore.isNewProjectDialogOpen = true },
  { id: 'open-project', label: 'Open Project', hint: 'Ctrl+O', run: async () => { /* handled by TopBar */ } },
  { id: 'save', label: 'Save Project', hint: 'Ctrl+S', run: () => projectStore.saveProject() },
  { id: 'export', label: 'Export', hint: 'Ctrl+E', run: () => uiStore.isExportDialogOpen = true },
  { id: 'settings', label: 'Settings', run: () => uiStore.isSettingsDialogOpen = true },
  { id: 'ws-edit', label: 'Workspace: Edit', run: () => uiStore.setActiveWorkspace('edit') },
  { id: 'ws-design', label: 'Workspace: Design', run: () => uiStore.setActiveWorkspace('design') },
  { id: 'ws-color', label: 'Workspace: Color', run: () => uiStore.setActiveWorkspace('color') },
  { id: 'ws-audio', label: 'Workspace: Audio', run: () => uiStore.setActiveWorkspace('audio') },
  { id: 'tool-select', label: 'Tool: Select (V)', run: () => timelineStore.setActiveTool('select') },
  { id: 'tool-razor', label: 'Tool: Razor (C)', run: () => timelineStore.setActiveTool('razor') },
  { id: 'snap-toggle', label: 'Toggle Snap (S)', run: () => { timelineStore.snapEnabled = !timelineStore.snapEnabled } },
]

const filtered = computed(() => {
  const q = query.value.toLowerCase().trim()
  if (!q) return commands
  return commands.filter(c => c.label.toLowerCase().includes(q))
})

watch(() => props.open, (v) => {
  if (v) {
    query.value = ''
    selected.value = 0
    nextTick(() => inputRef.value?.focus())
  }
})

watch(filtered, () => { selected.value = 0 })

function move(delta) {
  selected.value = Math.max(0, Math.min(filtered.value.length - 1, selected.value + delta))
}

function run(cmd) {
  try { cmd.run() } catch (e) { /* ignore */ }
  close()
}

function runSelected() {
  const cmd = filtered.value[selected.value]
  if (cmd) run(cmd)
}

function close() {
  emit('close')
}
</script>
