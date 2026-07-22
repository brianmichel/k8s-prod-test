<script setup lang="ts">
defineProps<{ kind: string }>()
</script>

<template>
  <svg v-if="kind==='Application'" viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="4" width="8" height="7" rx="1"/><rect x="13" y="13" width="8" height="7" rx="1"/><path d="M11 7.5h3a3 3 0 0 1 3 3V13M7 11v2a3 3 0 0 0 3 3h3"/></svg>
  <svg v-else-if="kind==='Deployment'" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 7A7 7 0 0 1 19 10M17 6l2 4-4 1M17.5 17A7 7 0 0 1 5 14M7 18l-2-4 4-1"/><rect x="9" y="9" width="6" height="6" rx="1"/></svg>
  <svg v-else-if="kind==='Rollout'" viewBox="0 0 24 24" aria-hidden="true"><circle cx="5" cy="12" r="2"/><circle cx="19" cy="6" r="2"/><circle cx="19" cy="18" r="2"/><path d="M7 12h3a4 4 0 0 0 4-4V6h3M10 12h2a4 4 0 0 1 4 4v2h1"/></svg>
  <svg v-else-if="kind==='ReplicaSet'" viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 8 4-8 4-8-4 8-4Z"/><path d="m4 11 8 4 8-4M4 15l8 4 8-4"/></svg>
  <svg v-else-if="kind==='Pod'" viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 8 4.5v9L12 21l-8-4.5v-9L12 3Z"/><path d="m4 7.5 8 4.5 8-4.5M12 12v9"/></svg>
  <svg v-else-if="kind==='Service'" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="3"/><circle cx="5" cy="5" r="2"/><circle cx="19" cy="5" r="2"/><circle cx="5" cy="19" r="2"/><circle cx="19" cy="19" r="2"/><path d="m7 7 3 3m4 0 3-3m-7 7-3 3m7-3 3 3"/></svg>
  <svg v-else-if="kind==='AnalysisRun'" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 19V5M4 19h16"/><path d="m7 15 4-4 3 2 5-6"/><circle cx="7" cy="15" r="1"/><circle cx="11" cy="11" r="1"/><circle cx="14" cy="13" r="1"/><circle cx="19" cy="7" r="1"/></svg>
  <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 8 4.5v9L12 21l-8-4.5v-9L12 3Z"/></svg>
</template>
