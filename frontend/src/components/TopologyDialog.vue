<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { MarkerType, VueFlow, useVueFlow, type Edge, type Node } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import ResourceNode from './ResourceNode.vue'
import KubeIcon from './KubeIcon.vue'
import { buildTopology, type TopologyResource } from '../topology'
import { healthOf, text, type Overview, type ResourceModel } from '../domain'

const props=defineProps<{model:ResourceModel;overview:Overview}>()
const emit=defineEmits<{close:[];action:[action:'sync'|'promote'|'abort',model:ResourceModel]}>()
const {fitView}=useVueFlow()
const selected=ref<TopologyResource|null>(null)
const topology=computed(()=>buildTopology(props.model,props.overview))
const nodes=computed<Node<TopologyResource>[]>(()=>topology.value.nodes.map(node=>({id:node.id,type:'resource',position:node.position,data:node,class:`kind-${kindClass(node.kind)}`,selectable:true,draggable:true})))
const edges=computed<Edge[]>(()=>topology.value.edges.map(edge=>({id:edge.id,source:edge.source,target:edge.target,type:'smoothstep',markerEnd:MarkerType.ArrowClosed,style:{stroke:'#8793a3',strokeWidth:1.5}})))
const activeResource=computed(()=>selected.value??topology.value.nodes[0]??null)
const labels=computed(()=>Object.entries(activeResource.value?.labels??{}).slice(0,10))
watch(topology,async()=>{selected.value=null;await nextTick();await fitView({padding:.2,duration:250})},{immediate:true})
function selectNode(event:{node:Node<TopologyResource>}):void{if(event.node.data)selected.value=event.node.data}
function requestAction(action:'sync'|'promote'|'abort'):void{emit('action',action,props.model)}
function kindClass(kind:string):string{return kind.replace(/([a-z])([A-Z])/g,'$1-$2').toLowerCase()}
</script>

<template>
  <div class="topology-overlay" @click.self="emit('close')">
    <section class="topology-window" role="dialog" aria-modal="true" :aria-label="`${model.name} topology`">
      <header class="topology-titlebar"><div><span>{{ model.kind }}</span><h2>{{ model.name }}</h2><p>{{ model.namespace }} · {{ model.status }} · {{ model.ready }}</p></div><div><span class="live"><i />Live topology</span><button class="btn btn-ghost btn-sm btn-square close" aria-label="Close topology" @click="emit('close')">×</button></div></header>
      <div class="topology-tools"><div class="legend kind-legend"><span v-for="kind in ['Application','Deployment','Rollout','ReplicaSet','Pod','Service','AnalysisRun']" :key="kind" :class="`kind-${kindClass(kind)}`"><KubeIcon :kind="kind" />{{ kind==='AnalysisRun'?'Analysis':kind }}</span></div><button class="btn btn-ghost btn-xs" @click="fitView({padding:.2,duration:250})">Fit to view</button></div>
      <div class="topology-layout">
        <div class="flow-wrap">
          <VueFlow :nodes="nodes" :edges="edges" :nodes-connectable="false" :elements-selectable="true" :min-zoom="0.35" :max-zoom="1.75" fit-view-on-init @node-click="selectNode">
            <template #node-resource="nodeProps"><ResourceNode v-bind="nodeProps" /></template>
            <Background pattern-color="#ccd3dc" :gap="20" :size="1" />
            <Controls :show-interactive="false" position="bottom-left" />
          </VueFlow>
        </div>
        <aside class="resource-inspector">
          <header><h3>Resource details</h3></header>
          <template v-if="activeResource">
            <div class="inspector-scroll">
              <div class="inspector-identity" :class="`kind-${kindClass(activeResource.kind)}`"><div class="inspector-icon"><KubeIcon :kind="activeResource.kind" /></div><div><small>{{ activeResource.kind }}</small><strong>{{ activeResource.name }}</strong><span :class="['health',healthOf(activeResource.status)]"><i />{{ activeResource.status }}</span></div></div>
              <section><h4>Properties</h4><dl><div><dt>Namespace</dt><dd>{{ activeResource.namespace }}</dd></div><div><dt>Created</dt><dd>{{ activeResource.item ? new Date(text(activeResource.item,'metadata.creationTimestamp')).toLocaleString() : '—' }}</dd></div><div v-if="activeResource.item"><dt>UID</dt><dd class="mono">{{ text(activeResource.item,'metadata.uid','—') }}</dd></div></dl></section>
              <section><h4>Labels <span>{{ labels.length }}</span></h4><div class="labels"><code v-for="([key,value]) in labels" :key="key"><b>{{ key }}</b><span>{{ value }}</span></code><p v-if="labels.length===0">No labels</p></div></section>
            </div>
            <footer v-if="activeResource.kind==='Application'||activeResource.kind==='Rollout'">
              <button v-if="activeResource.kind==='Application'" class="btn btn-primary btn-sm primary" @click="requestAction('sync')">Sync</button>
              <template v-else><button class="btn btn-primary btn-sm primary" :disabled="!model.paused" @click="requestAction('promote')">Promote</button><button class="btn btn-outline btn-error btn-sm danger" @click="requestAction('abort')">Abort</button></template>
            </footer>
          </template>
        </aside>
      </div>
    </section>
  </div>
</template>
