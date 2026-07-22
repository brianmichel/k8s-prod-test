<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from 'vue'
import { MarkerType, VueFlow, type Edge, type Node, type VueFlowStore } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { ArrowTopRightOnSquareIcon, CheckCircleIcon, MinusCircleIcon } from '@heroicons/vue/20/solid'
import ResourceNode from './ResourceNode.vue'
import KubeIcon from './KubeIcon.vue'
import { buildTopology, type TopologyResource } from '../topology'
import { healthOf, objects, read, text, type JsonObject, type Overview, type ResourceActionTarget, type ResourceModel } from '../domain'

const props=defineProps<{model:ResourceModel;overview:Overview}>()
const emit=defineEmits<{close:[];action:[action:'sync'|'promote'|'abort',target:ResourceActionTarget]}>()
const selected=ref<TopologyResource|null>(null)
const flowStore=shallowRef<VueFlowStore|null>(null)
let viewportInitialized=false
const topology=computed(()=>buildTopology(props.model,props.overview))
const nodes=computed<Node<TopologyResource>[]>(()=>topology.value.nodes.map(node=>({id:node.id,type:'resource',position:node.position,data:node,class:`kind-${kindClass(node.kind)}`,selectable:true,draggable:true})))
const edges=computed<Edge[]>(()=>topology.value.edges.map(edge=>({id:edge.id,source:edge.source,target:edge.target,type:'smoothstep',markerEnd:MarkerType.ArrowClosed,style:{stroke:'#8793a3',strokeWidth:1.5}})))
const activeResource=computed(()=>selected.value??topology.value.nodes[0]??null)
const applicationTarget=computed<ResourceActionTarget|null>(()=>props.model.type==='application'?{type:'application',name:props.model.name,namespace:props.model.namespace}:null)
const rolloutResource=computed(()=>topology.value.nodes.find(node=>node.kind==='Rollout')??null)
const rolloutTarget=computed<ResourceActionTarget|null>(()=>rolloutResource.value?{type:'rollout',name:rolloutResource.value.name,namespace:rolloutResource.value.namespace}:null)
const rolloutPaused=computed(()=>rolloutResource.value?.item?objects(rolloutResource.value.item,'status.pauseConditions').length>0:false)
const labels=computed(()=>stringEntries(activeResource.value?.labels??{}))
const annotations=computed(()=>stringEntries(activeResource.value?.annotations??{}).filter(([key])=>key!=='kubectl.kubernetes.io/last-applied-configuration'))
const argoDetails=computed(()=>{const item=activeResource.value?.kind==='Application'?activeResource.value.item:null;if(!item)return null;const history=objects(item,'status.history').slice().reverse(),resources=objects(item,'status.resources'),latest=history[0];return{repoURL:text(item,'spec.source.repoURL','—'),path:text(item,'spec.source.path','—'),targetRevision:text(item,'spec.source.targetRevision','—'),project:text(item,'spec.project','default'),phase:text(item,'status.operationState.phase','Unknown'),message:text(item,'status.operationState.message','No operation message'),reconciledAt:text(item,'status.reconciledAt'),automated:read(item,'spec.syncPolicy.automated')!==undefined,selfHeal:read(item,'spec.syncPolicy.automated.selfHeal')===true,prune:read(item,'spec.syncPolicy.automated.prune')===true,history,resources,latest}})
watch(topology,(current)=>{if(selected.value)selected.value=current.nodes.find(node=>node.id===selected.value?.id)??null})
async function initializeViewport(store:VueFlowStore):Promise<void>{flowStore.value=store;if(viewportInitialized)return;viewportInitialized=true;await nextTick();window.requestAnimationFrame(()=>void store.fitView({padding:.18,duration:0}))}
async function fitGraph():Promise<void>{await flowStore.value?.fitView({padding:.18,duration:250})}
function selectNode(event:{node:Node<TopologyResource>}):void{if(event.node.data)selected.value=event.node.data}
function requestAction(action:'sync'|'promote'|'abort',target:ResourceActionTarget|null):void{if(target)emit('action',action,target)}
function kindClass(kind:string):string{return kind.replace(/([a-z])([A-Z])/g,'$1-$2').toLowerCase()}
function deploymentInitiator(entry:JsonObject):string{if(read(entry,'initiatedBy.automated')===true)return'automated';return text(entry,'initiatedBy.username','unknown')}
function shortRevision(entry:JsonObject):string{const revision=text(entry,'revision','—');return revision.length>10?revision.slice(0,10):revision}
function repositoryURL(repoURL:string):string|null{try{const url=new URL(repoURL);if(url.protocol!=='https:'&&url.protocol!=='http:')return null;return url.toString().replace(/\.git\/?$/,'').replace(/\/$/,'')}catch{return null}}
function repositoryName(repoURL:string):string{return repositoryURL(repoURL)?.split('/').at(-1)??repoURL}
function sourceURL(repoURL:string,revision:string,path:string):string|null{const base=repositoryURL(repoURL);if(!base)return null;const separator=new URL(base).hostname.includes('gitlab')?'/-/tree/':'/tree/';return `${base}${separator}${encodeURIComponent(revision)}/${path.split('/').map(encodeURIComponent).join('/')}`}
function commitURL(repoURL:string,revision:string):string|null{const base=repositoryURL(repoURL);if(!base||!revision)return null;const separator=new URL(base).hostname.includes('gitlab')?'/-/commit/':'/commit/';return `${base}${separator}${encodeURIComponent(revision)}`}
function relativeTime(timestamp:string):string{if(!timestamp)return'unknown';const seconds=Math.max(0,(Date.now()-new Date(timestamp).getTime())/1000);if(seconds<60)return'just now';if(seconds<3600)return`${Math.floor(seconds/60)}m ago`;if(seconds<86400)return`${Math.floor(seconds/3600)}h ago`;return`${Math.floor(seconds/86400)}d ago`}
function stringEntries(record:Readonly<Record<string,unknown>>):readonly [string,string][]{return Object.entries(record).filter((entry):entry is [string,string]=>typeof entry[1]==='string').sort(([left],[right])=>left.localeCompare(right))}
</script>

<template>
  <div class="modal modal-open topology-overlay" @click.self="emit('close')">
    <section class="modal-box topology-window" role="dialog" aria-modal="true" :aria-label="`${model.name} topology`">
      <header class="topology-titlebar"><div><span>{{ model.kind }}</span><h2>{{ model.name }}</h2><p>{{ model.namespace }} · {{ model.status }} · {{ model.ready }}</p></div><div><div class="topology-primary-actions"><button v-if="applicationTarget" class="btn btn-primary btn-sm" @click="requestAction('sync',applicationTarget)">Sync</button><button v-if="rolloutTarget" class="btn btn-primary btn-sm" :disabled="!rolloutPaused" @click="requestAction('promote',rolloutTarget)">Promote</button><button v-if="rolloutTarget" class="btn btn-outline btn-error btn-sm" @click="requestAction('abort',rolloutTarget)">Roll back</button></div><span class="live"><i />Live topology</span><button class="btn btn-ghost btn-sm btn-square close" aria-label="Close topology" @click="emit('close')">×</button></div></header>
      <div class="topology-tools"><div class="legend kind-legend"><span v-for="kind in ['Application','Deployment','Rollout','ReplicaSet','Pod','Service','ConfigMap','Secret','ServiceAccount','Namespace','AnalysisRun']" :key="kind" :class="`kind-${kindClass(kind)}`"><KubeIcon :kind="kind" />{{ kind==='AnalysisRun'?'Analysis':kind }}</span></div><button class="btn btn-ghost btn-xs" @click="fitGraph">Fit to view</button></div>
      <div class="topology-layout">
        <div class="flow-wrap">
          <VueFlow :nodes="nodes" :edges="edges" :nodes-connectable="false" :elements-selectable="true" :min-zoom="0.35" :max-zoom="1.75" @pane-ready="initializeViewport" @node-click="selectNode">
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
              <template v-if="argoDetails"><section class="argo-delivery"><h4>Argo CD delivery <span class="badge badge-ghost badge-xs">{{ argoDetails.project }}</span></h4><div :class="['operation-summary',argoDetails.phase==='Succeeded'?'success':argoDetails.phase==='Failed'?'failure':'pending']"><span :class="['status',argoDetails.phase==='Succeeded'?'status-success':argoDetails.phase==='Failed'?'status-error':'status-warning']" /><div><strong>{{ argoDetails.phase }}</strong><small>{{ argoDetails.message }}</small></div></div><div class="source-links"><div><span>Repository</span><a v-if="repositoryURL(argoDetails.repoURL)" :href="repositoryURL(argoDetails.repoURL)??undefined" target="_blank" rel="noreferrer"><strong>{{ repositoryName(argoDetails.repoURL) }}</strong><ArrowTopRightOnSquareIcon /></a><strong v-else>{{ argoDetails.repoURL }}</strong></div><div><span>Source</span><a v-if="sourceURL(argoDetails.repoURL,argoDetails.targetRevision,argoDetails.path)" :href="sourceURL(argoDetails.repoURL,argoDetails.targetRevision,argoDetails.path)??undefined" target="_blank" rel="noreferrer"><strong>{{ argoDetails.targetRevision }}</strong><small>{{ argoDetails.path }}</small><ArrowTopRightOnSquareIcon /></a></div><div><span>Reconciled</span><strong>{{ argoDetails.reconciledAt ? new Date(argoDetails.reconciledAt).toLocaleString() : '—' }}</strong></div></div><h5>Sync policy</h5><ul class="policy-list"><li><component :is="argoDetails.automated?CheckCircleIcon:MinusCircleIcon" :class="{enabled:argoDetails.automated}"/><span><strong>Automated sync</strong><small>{{ argoDetails.automated?'Enabled':'Disabled' }}</small></span></li><li><component :is="argoDetails.selfHeal?CheckCircleIcon:MinusCircleIcon" :class="{enabled:argoDetails.selfHeal}"/><span><strong>Self-heal</strong><small>{{ argoDetails.selfHeal?'Enabled':'Disabled' }}</small></span></li><li><component :is="argoDetails.prune?CheckCircleIcon:MinusCircleIcon" :class="{enabled:argoDetails.prune}"/><span><strong>Prune</strong><small>{{ argoDetails.prune?'Enabled':'Disabled' }}</small></span></li></ul></section><section><h4>Deployment history <span>{{ argoDetails.history.length }}</span></h4><ul class="list deploy-history"><li v-for="entry in argoDetails.history.slice(0,5)" :key="String(read(entry,'id'))" class="list-row"><a v-if="commitURL(argoDetails.repoURL,text(entry,'revision'))" :href="commitURL(argoDetails.repoURL,text(entry,'revision'))??undefined" target="_blank" rel="noreferrer"><code>{{ shortRevision(entry) }}</code><ArrowTopRightOnSquareIcon /></a><code v-else>{{ shortRevision(entry) }}</code><div><strong>{{ relativeTime(text(entry,'deployedAt')) }}</strong><small>{{ text(entry,'deployedAt') ? new Date(text(entry,'deployedAt')).toLocaleString() : '—' }}</small></div><span class="initiator">{{ deploymentInitiator(entry) }}</span></li></ul></section><section><h4>Managed resources <span>{{ argoDetails.resources.length }}</span></h4><div class="managed-resources"><span v-for="resource in argoDetails.resources" :key="`${text(resource,'kind')}:${text(resource,'name')}`" class="badge badge-ghost badge-sm" :title="text(resource,'name')">{{ text(resource,'kind') }}</span></div></section></template>
              <section><h4>Properties</h4><ul class="list inspector-list"><li class="list-row"><span>Namespace</span><strong>{{ activeResource.namespace }}</strong></li><li class="list-row"><span>Created</span><strong>{{ activeResource.item ? new Date(text(activeResource.item,'metadata.creationTimestamp')).toLocaleString() : '—' }}</strong></li><li v-if="activeResource.item" class="list-row"><span>UID</span><strong class="mono">{{ text(activeResource.item,'metadata.uid','—') }}</strong></li></ul></section>
              <section><h4>Labels <span>{{ labels.length }}</span></h4><ul class="list labels"><li v-for="([key,value]) in labels" :key="key" class="list-row"><code><b :title="key">{{ key }}</b><span :title="value">{{ value }}</span></code></li><li v-if="labels.length===0" class="list-row">No labels</li></ul></section><section><h4>Annotations <span>{{ annotations.length }}</span></h4><ul class="list labels annotations"><li v-for="([key,value]) in annotations" :key="key" class="list-row"><code><b :title="key">{{ key }}</b><span :title="value">{{ value }}</span></code></li><li v-if="annotations.length===0" class="list-row">No annotations</li></ul></section>
            </div>
          </template>
        </aside>
      </div>
    </section>
  </div>
</template>
