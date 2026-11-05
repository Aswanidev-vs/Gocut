<template>
  <button
    class="inline-flex items-center justify-center gap-1.5 rounded px-3 py-1.5 text-[12px] font-medium transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--text-primary)] disabled:opacity-50 disabled:cursor-not-allowed"
    :class="variantClasses"
    v-bind="$attrs"
  >
    <slot />
  </button>
</template>

<script setup>
defineOptions({ inheritAttrs: false })
const props = defineProps({
  variant: { type: String, default: 'secondary' },
})
const variantClasses = computed(() => {
  switch (props.variant) {
    case 'primary':
      return 'bg-[var(--accent)] text-[var(--surface-app)] hover:opacity-90'
    case 'ghost':
      return 'bg-transparent text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]'
    default:
      return 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--surface-panel)]'
  }
})
</script>