import { healthOf, number, object, objects, read, text, type JsonObject, type Overview, type ResourceModel } from './domain'

function metadata(item: JsonObject): Pick<ResourceModel, 'name' | 'namespace' | 'created'> {
  return { name: text(item, 'metadata.name', 'unknown'), namespace: text(item, 'metadata.namespace', 'default'), created: text(item, 'metadata.creationTimestamp') }
}

/** Projects raw Kubernetes resources into the table's stable view model. */
export function buildModels(overview: Overview): readonly ResourceModel[] {
  const applications = overview.applications.map((item): ResourceModel => {
    const meta = metadata(item), status = text(item, 'status.health.status', 'Unknown'), sync = text(item, 'status.sync.status', 'Unknown'), history=objects(item,'status.history'), latest=history.at(-1)
    return { id:`application:${meta.namespace}:${meta.name}`, type:'application', kind:'Application', ...meta, status, health:healthOf(status), ready:sync, percent:sync==='Synced'?100:30, revision:text(item,'status.sync.revision',text(item,'spec.source.targetRevision','—')), image:text(item,'spec.source.path','—'), target:text(item,'spec.destination.namespace','—'), step:sync==='Synced'?2:1, steps:3, paused:false, application:meta.name, argo:{repoURL:text(item,'spec.source.repoURL','—'),targetRevision:text(item,'spec.source.targetRevision','—'),project:text(item,'spec.project','default'),operationPhase:text(item,'status.operationState.phase','Unknown'),operationMessage:text(item,'status.operationState.message'),lastDeployedAt:latest?text(latest,'deployedAt'):'',deployedBy:latest?deploymentInitiator(latest):'—',reconciledAt:text(item,'status.reconciledAt'),automated:read(item,'spec.syncPolicy.automated')!==undefined,selfHeal:read(item,'spec.syncPolicy.automated.selfHeal')===true,prune:read(item,'spec.syncPolicy.automated.prune')===true,managedResources:objects(item,'status.resources').length} }
  })
  const rollouts = overview.rollouts.map((item): ResourceModel => {
    const meta=metadata(item), phase=text(item,'status.phase','Unknown'), desired=number(item,'spec.replicas'), ready=number(item,'status.readyReplicas'), steps=objects(item,'spec.strategy.canary.steps'), blueGreen=read(item,'spec.strategy.blueGreen')!==undefined, pauses=objects(item,'status.pauseConditions'), paused=pauses.length>0
    return { id:`rollout:${meta.namespace}:${meta.name}`, type:'rollout', kind:blueGreen?'Blue/green':'Canary', ...meta, status:paused?'Paused':phase, health:paused?'progressing':healthOf(phase), ready:`${ready}/${desired}`, percent:desired?Math.round(ready/desired*100):0, revision:text(item,'status.currentPodHash','—'), image:text(item,'spec.template.spec.containers.0.image','—'), target:blueGreen?'Blue/green':`${steps.length} steps`, step:number(item,'status.currentStepIndex'), steps:Math.max(blueGreen?3:steps.length,3), paused, application:trackedApplication(item), argo:null }
  })
  const deployments = overview.deployments.map((item): ResourceModel => {
    const meta=metadata(item), desired=number(item,'spec.replicas'), ready=number(item,'status.readyReplicas'), available=number(item,'status.availableReplicas'), healthy=desired>0&&available===desired
    return { id:`deployment:${meta.namespace}:${meta.name}`, type:'deployment', kind:'Deployment', ...meta, status:healthy?'Healthy':'Progressing', health:healthy?'healthy':'progressing', ready:`${ready}/${desired}`, percent:desired?Math.round(ready/desired*100):0, revision:`Generation ${textOrNumber(item,'metadata.generation')}`, image:text(item,'spec.template.spec.containers.0.image','—'), target:`${desired} replicas`, step:healthy?2:1, steps:3, paused:false, application:trackedApplication(item), argo:null }
  })
  return [...applications,...rollouts,...deployments]
}

/** Determines pod readiness from Kubernetes conditions. */
export function isPodReady(pod: JsonObject): boolean {
  return objects(pod,'status.conditions').some((condition)=>text(condition,'type')==='Ready'&&text(condition,'status')==='True')
}

function deploymentInitiator(history:JsonObject):string{if(read(history,'initiatedBy.automated')===true)return'automated';return text(history,'initiatedBy.username','unknown')}

function trackedApplication(item: JsonObject): string | null {
  const tracking=object(item,'metadata.annotations')['argocd.argoproj.io/tracking-id']
  if(typeof tracking!=='string'||!tracking)return null
  return tracking.split(':')[0] ?? null
}

function textOrNumber(value: unknown,path: string): string {
  const result=read(value,path)
  return typeof result==='string'||typeof result==='number'?String(result):'—'
}
