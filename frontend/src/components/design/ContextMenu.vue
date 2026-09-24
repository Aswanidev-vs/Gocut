<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps({
  items: { type: Array, required: true },
  x: { type: Number, required: true },
  y: { type: Number, required: true },
})

const emit = defineEmits(['close', 'action'])
const menuRef = ref(null)
const adjustedPos = ref({ x: props.x, y: props.y })

onMounted(() => {
  nextTick(() => {
    if (menuRef.value) {
      const rect = menuRef.value.getBoundingClientRect()
      const vw = window.innerWidth
      const vh = window.innerHeight
      let x = props.x
      let y = props.y
      if (x + rect.width > vw) x = vw - rect.width - 8
      if (y + rect.height > vh) y = vh - rect.height - 8
      if (x < 0) x = 8
      if (y < 0) y = 8
      adjustedPos.value = { x, y }
    }
  })
  window.addEventListener('mousedown', onClickOutside)
  window.addEventListener('contextmenu', onClickOutside)
})

onUnmounted(() => {
  window.removeEventListener('mousedown', onClickOutside)
  window.removeEventListener('contextmenu', onClickOutside)
})

function onClickOutside(e) {
  if (menuRef.value && !menuRef.value.contains(e.target)) {
    emit('close')
  }
}

function handleAction(item) {
  if (item.disabled) return
  emit('action', item.action)
  emit('close')
}
</script>

<template>
  <div
    ref="menuRef"
    class="fixed z-50 min-w-[180px] bg-cr-panel border border-cr-line rounded-none py-1 overflow-hidden"
    :style="{ left: adjustedPos.x + 'px', top: adjustedPos.y + 'px' }"
    @mousedown.stop
  >
    <template v-for="(item, i) in items" :key="i">
      <div v-if="item.separator" class="h-px bg-cr-line my-1" />
      <button
        v-else
        class="w-full flex items-center gap-2.5 px-3 py-1.5 text-[11px] transition-colors text-left bg-cr-raise"
        :class="item.disabled
          ? 'text-ink-faint cursor-not-allowed'
          : 'text-ink hover:bg-white/5 hover:text-accent'"
        @click="handleAction(item)"
      >
        <component v-if="item.icon" :is="item.icon" :size="12" class="flex-shrink-0" />
        <span class="flex-1">{{ item.label }}</span>
        <span v-if="item.shortcut" class="text-[9px] text-ink-faint font-jetbrains-mono tabular-nums ml-4">{{ item.shortcut }}</span>
      </button>
    </template>
  </div>
</template>
