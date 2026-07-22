# Concepts in this lab

The point is to connect each Kubernetes object to a behavior you can observe.

## Request path

```text
laptop
  -> kubectl port-forward
  -> Service (stable virtual address)
  -> Deployment's ReplicaSet
  -> one of two Pods
  -> Go HTTP server
```

The Service selects Pods by the `app=playground-api` label. It does not own the Pods; the Deployment owns the desired replica count and rollout behavior.

## Resource map

| Resource | What to notice | Try |
| --- | --- | --- |
| Namespace | A boundary for names and cleanup | `kubectl get all -A` |
| Deployment | Desired state, replicas, rolling updates | `kubectl describe deployment/playground-api` |
| ReplicaSet | The Deployment's current Pod set | `kubectl get replicasets` |
| Pod | The scheduling and runtime unit | `kubectl describe pods` |
| Service | Stable discovery and load balancing | `kubectl get endpointslices` |
| ConfigMap | Non-sensitive configuration | Change `APP_MESSAGE` and redeploy |
| Secret | Sensitive configuration reference | Inspect metadata, not decoded values |
| ServiceAccount | Workload identity | Note that token mounting is disabled |
| Probes | Liveness, readiness, and startup contracts | Watch the 3-second warm-up |

## Desired state and reconciliation

The YAML says what should exist. The Kubernetes controllers continuously compare that desired state with the observed cluster state and make changes. Delete one Pod and watch the ReplicaSet replace it:

```sh
kubectl --context kind-learning-lab -n learning-lab delete pod -l app=playground-api --wait=false
kubectl --context kind-learning-lab -n learning-lab get pods --watch
```

The new Pod has a different name and IP, but the Service remains the stable entry point.

## Scheduling and resources

The `resources.requests` values influence scheduling; the `limits` values constrain runtime usage. They are deliberately small so you can inspect them without needing a large cluster:

```sh
kubectl --context kind-learning-lab -n learning-lab describe pod -l app=playground-api
kubectl --context kind-learning-lab describe node
```

## Probes are different contracts

- Startup probe: gives a slow-starting process time to boot before liveness takes over.
- Readiness probe: controls whether a Pod receives Service traffic; it is not a process supervisor.
- Liveness probe: tells the kubelet when to restart a stuck process.

The app reports `503` from `/readyz` for three seconds, then becomes ready. During that window the Pod can be Running but not Ready, which is exactly the distinction worth seeing.

## Configuration and identity

The Deployment imports the ConfigMap with `envFrom`, references one Secret key explicitly, and uses the Downward API for Pod metadata. The root endpoint exposes safe diagnostics:

```json
{
  "app": "playground-api",
  "message": "hello from a ConfigMap",
  "podName": "playground-api-...",
  "podNamespace": "learning-lab",
  "podIP": "10.244.0.7",
  "secretConfigured": true
}
```

The actual Secret value is never returned by the app.

## Rollouts and graceful termination

The Deployment uses `maxUnavailable: 0` and `maxSurge: 1`, so a new Pod must become ready before an old Pod is removed. The Go process handles `SIGTERM` and has a 10-second HTTP shutdown window, while the Pod has a 15-second termination grace period.

Watch the rollout after changing the image or manifest:

```sh
kubectl --context kind-learning-lab -n learning-lab rollout history deployment/playground-api
kubectl --context kind-learning-lab -n learning-lab rollout status deployment/playground-api --watch
```

