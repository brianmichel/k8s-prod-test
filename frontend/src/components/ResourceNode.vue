<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position, type NodeProps } from '@vue-flow/core'
import type { TopologyResource } from '../topology'
import KubeIcon from './KubeIcon.vue'

const props=defineProps<NodeProps<TopologyResource>>()
const preferredLabels=['app.kubernetes.io/name','app.kubernetes.io/component','app','rollouts-pod-template-hash','pod-template-hash']
const tags=computed(()=>Object.entries(props.data.labels).filter((entry):entry is [string,string]=>typeof entry[1]==='string').sort(([left],[right])=>labelRank(left)-labelRank(right)).slice(0,2))
function labelRank(key:string):number{const index=preferredLabels.indexOf(key);return index===-1?preferredLabels.length:index}
</script>

<template>
  <Handle type="target" :position="Position.Left" />
  <div class="resource-icon"><KubeIcon :kind="data.kind" /></div>
  <div class="resource-node-copy"><small>{{ data.kind }}</small><strong :title="data.name">{{ data.name }}</strong><span class="node-status"><i />{{ data.status }}</span><div v-if="tags.length" class="node-tags"><code v-for="([key,value]) in tags" :key="key" :title="`${key}=${value}`">{{ key.split('/').at(-1) }}={{ value }}</code></div></div>
  <Handle type="source" :position="Position.Right" />
</template>
