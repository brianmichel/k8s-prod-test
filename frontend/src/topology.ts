import dagre from '@dagrejs/dagre'
import { object, objects, read, text, type JsonObject, type Overview, type ResourceModel } from './domain'

export type NodeCategory = 'controller' | 'network' | 'workload' | 'analysis'
export interface TopologyResource { readonly id:string; readonly kind:string; readonly name:string; readonly namespace:string; readonly category:NodeCategory; readonly status:string; readonly labels:JsonObject; readonly item:JsonObject | null }
export interface TopologyEdge { readonly id:string; readonly source:string; readonly target:string }
export interface PositionedResource extends TopologyResource { readonly position:{readonly x:number;readonly y:number} }
export interface Topology { readonly nodes:readonly PositionedResource[]; readonly edges:readonly TopologyEdge[] }

interface RawResource extends Omit<TopologyResource,'category'|'status'> { readonly owners:readonly JsonObject[] }

function resource(item:JsonObject,kind:string):RawResource { const name=text(item,'metadata.name','unknown'),namespace=text(item,'metadata.namespace','default');return {id:`${kind}:${namespace}:${name}`,kind,name,namespace,item,labels:object(item,'metadata.labels'),owners:objects(item,'metadata.ownerReferences')} }
function ownedBy(child:RawResource,parent:RawResource):boolean{return child.namespace===parent.namespace&&child.owners.some(owner=>text(owner,'name')===parent.name)}
function selectorMatches(labels:JsonObject,selector:JsonObject):boolean{const entries=Object.entries(selector);return entries.length>0&&entries.every(([key,value])=>labels[key]===value)}
function statusOf(item:JsonObject,kind:string):string{if(kind==='Pod')return text(item,'status.phase','Unknown');if(kind==='ReplicaSet')return `${read(item,'status.readyReplicas')??0}/${read(item,'spec.replicas')??0} ready`;if(kind==='Service')return text(item,'spec.type','ClusterIP');if(kind==='AnalysisRun')return text(item,'status.phase','Pending');return text(item,'status.phase',text(item,'status.conditions.0.type','Observed'))}

/** Builds and lays out a Kubernetes ownership graph for one selected controller. */
export function buildTopology(model:ResourceModel,overview:Overview):Topology {
  const deployments=overview.deployments.map(item=>resource(item,'Deployment')),rollouts=overview.rollouts.map(item=>resource(item,'Rollout')),sets=overview.replicaSets.map(item=>resource(item,'ReplicaSet')),pods=overview.pods.map(item=>resource(item,'Pod')),services=overview.services.map(item=>resource(item,'Service')),analyses=overview.analysisRuns.map(item=>resource(item,'AnalysisRun'))
  let roots:RawResource[]=[]
  if(model.type==='deployment')roots=deployments.filter(item=>item.name===model.name&&item.namespace===model.namespace)
  if(model.type==='rollout')roots=rollouts.filter(item=>item.name===model.name&&item.namespace===model.namespace)
  if(model.type==='application')roots=[...deployments,...rollouts].filter(item=>item.namespace===model.target)
  const selectedSets=sets.filter(set=>roots.some(root=>ownedBy(set,root)))
  const selectedPods=pods.filter(pod=>selectedSets.some(set=>ownedBy(pod,set))||roots.some(root=>ownedBy(pod,root)))
  const selectedServices=services.filter(service=>roots.some(root=>service.namespace===root.namespace&&selectorMatches(root.labels,object(service.item,'spec.selector'))))
  const selectedAnalyses=analyses.filter(run=>roots.some(root=>ownedBy(run,root)))
  const nodes:TopologyResource[]=[],edges:TopologyEdge[]=[]
  const add=(item:RawResource,category:NodeCategory,status=statusOf(item.item??{},item.kind)):void=>{if(!nodes.some(node=>node.id===item.id))nodes.push({...item,category,status})}
  if(model.type==='application'){const app:RawResource={id:model.id,kind:'Application',name:model.name,namespace:model.namespace,item:null,labels:{},owners:[]};add(app,'controller',model.status);for(const root of roots){add(root,'controller');edges.push(edge(app,root))}}
  else for(const root of roots)add(root,'controller',model.status)
  for(const service of selectedServices){add(service,'network');for(const root of roots.filter(root=>selectorMatches(root.labels,object(service.item,'spec.selector'))))edges.push(edge(root,service))}
  for(const set of selectedSets){add(set,'workload');const owner=roots.find(root=>ownedBy(set,root));if(owner)edges.push(edge(owner,set))}
  for(const run of selectedAnalyses){add(run,'analysis');const owner=roots.find(root=>ownedBy(run,root));if(owner)edges.push(edge(owner,run))}
  for(const pod of selectedPods){add(pod,'workload');const owner=selectedSets.find(set=>ownedBy(pod,set))??roots.find(root=>ownedBy(pod,root));if(owner)edges.push(edge(owner,pod))}
  return layout(nodes,edges)
}
function edge(source:RawResource,target:RawResource):TopologyEdge{return{id:`${source.id}->${target.id}`,source:source.id,target:target.id}}
function layout(nodes:readonly TopologyResource[],edges:readonly TopologyEdge[]):Topology{const graph=new dagre.graphlib.Graph().setDefaultEdgeLabel(()=>({}));graph.setGraph({rankdir:'LR',nodesep:32,ranksep:90,marginx:30,marginy:30});for(const node of nodes)graph.setNode(node.id,{width:210,height:76});for(const item of edges)graph.setEdge(item.source,item.target);dagre.layout(graph);return{nodes:nodes.map(node=>{const point=graph.node(node.id);return{...node,position:{x:point.x-105,y:point.y-38}}}),edges}}
