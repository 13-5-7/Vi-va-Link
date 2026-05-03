<template>
  <span
    :class="badgeClass"
    class="px-2 py-0.5 rounded text-xs font-bold inline-block"
  >{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ status: { type: String, required: true } })

const statusMap = {
  accepted:   { label: '受付済', cls: 'bg-blue-100 text-blue-800' },
  loaded:     { label: '積載済', cls: 'bg-orange-100 text-orange-700' },
  in_transit: { label: '輸送中', cls: 'bg-purple-100 text-purple-800' },
  delivered:  { label: '配達済', cls: 'bg-green-100 text-green-800' },
  cancelled:  { label: 'キャンセル', cls: 'bg-gray-200 text-gray-600' },
}

const current = computed(() => statusMap[props.status] || { label: props.status, cls: 'bg-gray-100 text-gray-600' })
const label = computed(() => current.value.label)
const badgeClass = computed(() => current.value.cls)
</script>
