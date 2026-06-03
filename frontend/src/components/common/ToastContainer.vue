<script setup>
import { useUiStore } from '../../stores/uiStore'
import { X, Info, CheckCircle2, AlertCircle } from 'lucide-vue-next'

const uiStore = useUiStore()

const iconMap = {
  info: Info,
  success: CheckCircle2,
  error: AlertCircle,
}
</script>

<template>
  <div class="fixed top-3 right-3 z-50 flex flex-col gap-2 pointer-events-none">
    <TransitionGroup name="toast">
      <div
        v-for="toast in uiStore.toasts"
        :key="toast.id"
        class="bg-panel border border-border rounded px-3 py-2 text-xs text-text-primary shadow-lg flex items-center gap-2 min-w-[220px] pointer-events-auto"
        :class="{
          'border-accent/30': toast.type === 'success',
          'border-red-400/30': toast.type === 'error',
        }"
      >
        <component
          :is="iconMap[toast.type] || Info"
          :size="14"
          :class="{
            'text-accent': toast.type === 'success',
            'text-red-400': toast.type === 'error',
          }"
        />
        <span class="flex-1">{{ toast.message }}</span>
        <button class="p-0.5 rounded hover:bg-border text-text-secondary" @click="uiStore.removeToast(toast.id)">
          <X :size="10" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(30px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
