<script setup lang="ts">
import { Handle, Position, type NodeProps } from '@vue-flow/core'
import type { TopologyResource } from '../topology'
import KubeIcon from './KubeIcon.vue'

defineProps<NodeProps<TopologyResource>>()
</script>

<template>
  <Handle type="target" :position="Position.Left" />
  <div class="resource-icon"><KubeIcon :kind="data.kind" /></div>
  <div class="resource-node-copy"><small>{{ data.kind }}</small><strong :title="data.name">{{ data.name }}</strong><span><i />{{ data.status }}</span></div>
  <Handle type="source" :position="Position.Right" />
</template>
