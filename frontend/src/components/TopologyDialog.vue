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
const labels=computed(()=>stringEntries(activeResource.value?.labels??{}))
const annotations=computed(()=>stringEntries(activeResource.value?.annotations??{}).filter(([key])=>key!=='kubectl.kubernetes.io/last-applied-configuration'))
watch(topology,async()=>{selected.value=null;await nextTick();await fitView({padding:.2,duration:250})},{immediate:true})
function selectNode(event:{node:Node<TopologyResource>}):void{if(event.node.data)selected.value=event.node.data}
function requestAction(action:'sync'|'promote'|'abort'):void{emit('action',action,props.model)}
function kindClass(kind:string):string{return kind.replace(/([a-z])([A-Z])/g,'$1-$2').toLowerCase()}
function stringEntries(record:Readonly<Record<string,unknown>>):readonly [string,string][]{return Object.entries(record).filter((entry):entry is [string,string]=>typeof entry[1]==='string').sort(([left],[right])=>left.localeCompare(right))}
</script>

<template>
  <div class="modal modal-open topology-overlay" @click.self="emit('close')">
    <section class="modal-box topology-window" role="dialog" aria-modal="true" :aria-label="`${model.name} topology`">
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
              <section><h4>Properties</h4><ul class="list inspector-list"><li class="list-row"><span>Namespace</span><strong>{{ activeResource.namespace }}</strong></li><li class="list-row"><span>Created</span><strong>{{ activeResource.item ? new Date(text(activeResource.item,'metadata.creationTimestamp')).toLocaleString() : '—' }}</strong></li><li v-if="activeResource.item" class="list-row"><span>UID</span><strong class="mono">{{ text(activeResource.item,'metadata.uid','—') }}</strong></li></ul></section>
              <section><h4>Labels <span>{{ labels.length }}</span></h4><ul class="list labels"><li v-for="([key,value]) in labels" :key="key" class="list-row"><code><b :title="key">{{ key }}</b><span :title="value">{{ value }}</span></code></li><li v-if="labels.length===0" class="list-row">No labels</li></ul></section><section><h4>Annotations <span>{{ annotations.length }}</span></h4><ul class="list labels annotations"><li v-for="([key,value]) in annotations" :key="key" class="list-row"><code><b :title="key">{{ key }}</b><span :title="value">{{ value }}</span></code></li><li v-if="annotations.length===0" class="list-row">No annotations</li></ul></section>
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
