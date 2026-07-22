export type Health = 'healthy' | 'progressing' | 'degraded'
export type ResourceType = 'application' | 'rollout' | 'deployment'
export type JsonObject = Readonly<Record<string, unknown>>

export interface Overview {
  readonly generatedAt: string
  readonly applications: readonly JsonObject[]
  readonly rollouts: readonly JsonObject[]
  readonly deployments: readonly JsonObject[]
  readonly replicaSets: readonly JsonObject[]
  readonly services: readonly JsonObject[]
  readonly analysisRuns: readonly JsonObject[]
  readonly pods: readonly JsonObject[]
  readonly warnings: readonly string[]
}

export interface ResourceModel {
  readonly id: string
  readonly type: ResourceType
  readonly kind: string
  readonly name: string
  readonly namespace: string
  readonly created: string
  readonly status: string
  readonly health: Health
  readonly ready: string
  readonly percent: number
  readonly revision: string
  readonly image: string
  readonly target: string
  readonly step: number
  readonly steps: number
  readonly paused: boolean
}

/** Reads a nested value from an untrusted Kubernetes JSON object. */
export function read(value: unknown, path: string): unknown {
  let current = value
  for (const part of path.split('.')) {
    if (!isObject(current)) return undefined
    current = current[part]
  }
  return current
}

/** Returns a string projection or its fallback. */
export function text(value: unknown, path: string, fallback = ''): string {
  const result = read(value, path)
  return typeof result === 'string' ? result : fallback
}

/** Returns a finite numeric projection or its fallback. */
export function number(value: unknown, path: string, fallback = 0): number {
  const result = read(value, path)
  return typeof result === 'number' && Number.isFinite(result) ? result : fallback
}

/** Returns an object projection without trusting the transport shape. */
export function object(value: unknown, path: string): JsonObject {
  const result = read(value, path)
  return isObject(result) ? result : {}
}

/** Returns an object-array projection without trusting the transport shape. */
export function objects(value: unknown, path: string): readonly JsonObject[] {
  const result = read(value, path)
  return Array.isArray(result) ? result.filter(isObject) : []
}

/** Parses the overview API boundary into a stable UI contract. */
export function parseOverview(value: unknown): Overview {
  if (!isObject(value)) throw new Error('The cluster response is not an object')
  const generatedAt = text(value, 'generatedAt')
  if (!generatedAt) throw new Error('The cluster response has no generation timestamp')
  const warningsValue = read(value, 'warnings')
  return {
    generatedAt,
    applications: objects(value, 'applications'),
    rollouts: objects(value, 'rollouts'),
    deployments: objects(value, 'deployments'),
    replicaSets: objects(value, 'replicaSets'),
    services: objects(value, 'services'),
    analysisRuns: objects(value, 'analysisRuns'),
    pods: objects(value, 'pods'),
    warnings: Array.isArray(warningsValue) ? warningsValue.filter((item): item is string => typeof item === 'string') : [],
  }
}

/** Classifies controller status for consistent visual treatment. */
export function healthOf(value: string): Health {
  const status = value.toLowerCase()
  if (['healthy', 'synced', 'completed', 'running', 'succeeded'].some((part) => status.includes(part))) return 'healthy'
  if (['degraded', 'error', 'failed', 'missing', 'unavailable'].some((part) => status.includes(part))) return 'degraded'
  return 'progressing'
}

/** Determines whether an unknown value is a JSON-style object. */
export function isObject(value: unknown): value is JsonObject {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
