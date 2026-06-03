<script setup>
import { ref, computed } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { X, FolderOpen, Sparkles } from 'lucide-vue-next'

const projectStore = useProjectStore()
const uiStore = useUiStore()

const projectName = ref('')
const aspectRatio = ref('16:9')
const resolution = ref('1080p')
const fps = ref(30)

const resolutionPresets = {
  '4K':    { width: 3840, height: 2160, label: '4K (3840×2160)' },
  '1080p': { width: 1920, height: 1080, label: '1080p (1920×1080)' },
  '720p':  { width: 1280, height: 720,  label: '720p (1280×720)' },
  '480p':  { width: 854,  height: 480,  label: '480p (854×480)' },
}

const aspectOptions = ['16:9', '9:16', '1:1', '4:3']
const fpsOptions = [24, 25, 30, 60]

const currentRes = computed(() => resolutionPresets[resolution.value] || resolutionPresets['1080p'])
const currentAspect = computed(() => aspectRatio.value)

function close() {
  uiStore.isNewProjectDialogOpen = false
}

async function create() {
  try {
    const res = currentRes.value
    await projectStore.createProject({
      name: projectName.value.trim() || 'Untitled',
      aspectRatio: currentAspect.value,
      resolution: { width: res.width, height: res.height },
      fps: fps.value,
    })
    uiStore.isNewProjectDialogOpen = false
    uiStore.addToast('Project created', 'success', 1800)
  } catch (e) {
    uiStore.addToast('Could not create project: ' + (e?.message || e), 'error')
  }
}

function openExisting() {
  // close this dialog and trigger the open flow on App.vue
  uiStore.isNewProjectDialogOpen = false
  uiStore.openProjectDialog = true
}
</script>

<template>
  <div
    v-if="uiStore.isNewProjectDialogOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-bg/70 backdrop-blur-sm"
    @click.self="close"
  >
    <div class="w-[480px] bg-panel border border-border rounded-md shadow-2xl">
      <div class="flex items-center justify-between px-4 py-3 border-b border-border">
        <h3 class="text-sm font-semibold text-text-primary">New Project</h3>
        <button class="p-1 rounded hover:bg-border text-text-secondary" @click="close">
          <X :size="14" />
        </button>
      </div>

      <div class="px-4 py-4 flex flex-col gap-3">
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Project Name</label>
          <input
            v-model="projectName"
            placeholder="My Awesome Edit"
            class="w-full bg-bg border border-border rounded px-2.5 py-2 text-sm text-text-primary outline-none focus:border-accent"
            @keyup.enter="create"
            autofocus
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Aspect Ratio</label>
            <div class="grid grid-cols-4 gap-1">
              <button
                v-for="a in aspectOptions"
                :key="a"
                class="px-2 py-1.5 rounded text-xs border transition-colors"
                :class="aspectRatio === a
                  ? 'border-accent bg-accent/10 text-accent'
                  : 'border-border text-text-secondary hover:text-text-primary hover:border-text-secondary'"
                @click="aspectRatio = a"
              >
                {{ a }}
              </button>
            </div>
          </div>

          <div>
            <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">FPS</label>
            <div class="grid grid-cols-4 gap-1">
              <button
                v-for="f in fpsOptions"
                :key="f"
                class="px-2 py-1.5 rounded text-xs border transition-colors"
                :class="fps === f
                  ? 'border-accent bg-accent/10 text-accent'
                  : 'border-border text-text-secondary hover:text-text-primary hover:border-text-secondary'"
                @click="fps = f"
              >
                {{ f }}
              </button>
            </div>
          </div>
        </div>

        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Resolution</label>
          <div class="grid grid-cols-2 gap-1.5">
            <button
              v-for="(r, key) in resolutionPresets"
              :key="key"
              class="px-2.5 py-1.5 rounded text-xs border text-left transition-colors"
              :class="resolution === key
                ? 'border-accent bg-accent/10 text-accent'
                : 'border-border text-text-secondary hover:text-text-primary hover:border-text-secondary'"
              @click="resolution = key"
            >
              <div class="font-medium">{{ key }}</div>
              <div class="text-[10px] opacity-70">{{ r.width }} × {{ r.height }}</div>
            </button>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-between gap-2 px-4 py-3 border-t border-border bg-bg/30">
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
          @click="openExisting"
        >
          <FolderOpen :size="14" /> Open Existing
        </button>
        <div class="flex gap-2">
          <button
            class="px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
            @click="close"
          >
            Cancel
          </button>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors"
            @click="create"
          >
            <Sparkles :size="14" /> Create
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
