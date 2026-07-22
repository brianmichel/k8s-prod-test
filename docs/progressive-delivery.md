# Progressive delivery

The baseline application uses a Kubernetes `Deployment` with `RollingUpdate`. The progressive examples use Argo Rollouts, a separate controller and CRD set that manages ReplicaSets with blue-green, canary, analysis, pause, promotion, and abort behavior.

## Install the controller

```sh
mise run rollouts-install
kubectl --context kind-learning-lab -n argo-rollouts get pods
kubectl argo rollouts version
```

The lab pins Argo Rollouts `1.9.1` in `.mise.toml` and applies its versioned controller manifest. The `kubectl-argo-rollouts` plugin is installed by mise as the inspection and promotion CLI.

The progressive examples can also be managed by Argo CD after you replace the placeholder repository URL in their Application manifests:

```sh
mise run argocd-bluegreen-application
# or
mise run argocd-canary-application
```

## Blue-green

Blue-green keeps a stable stack serving the active Service while a new stack is created behind a preview Service. This lab also runs a pre-promotion smoke check and pauses before the active Service switches:

```sh
mise run bluegreen-apply
mise run bluegreen-status
```

In another terminal, inspect the preview stack:

```sh
mise run bluegreen-preview
curl -s http://127.0.0.1:8082 | jq
```

To trigger a new version, edit `APP_VERSION` in [k8s/progressive/bluegreen/rollout.yaml](../k8s/progressive/bluegreen/rollout.yaml) from `bluegreen-v1` to `bluegreen-v2`, then apply again:

```sh
kubectl --context kind-learning-lab apply --kustomize k8s/progressive/bluegreen
mise run bluegreen-status
```

The expected sequence is:

1. The new ReplicaSet starts behind the preview Service.
2. The `/readyz` smoke AnalysisRun must pass.
3. The Rollout pauses while active traffic remains on the old ReplicaSet.
4. `mise run bluegreen-promote` switches the active Service.
5. The old ReplicaSet scales down after the delay.

Abort instead of promote when the preview is bad:

```sh
mise run bluegreen-abort
```

## Canary

The canary example uses four replicas and declarative steps:

```text
25% -> pause forever -> 50% -> pause 15s -> 100%
```

Run it with:

```sh
mise run canary-apply
mise run canary-status
```

Change `APP_VERSION` from `canary-v1` to `canary-v2`, apply the Kustomize directory, and watch the pause:

```sh
kubectl --context kind-learning-lab apply --kustomize k8s/progressive/canary
mise run canary-status
```

Advance one step at a time with `mise run canary-promote`, or roll it back with `mise run canary-abort`.

This local example does not have an ingress controller or service mesh traffic router. Its `setWeight` is therefore approximated by the number of stable and canary Pods behind one Service, not a precise percentage of HTTP requests. Precise traffic percentages require traffic management integration.

## Optional reliability policies

The optional overlay demonstrates two adjacent production controls:

```sh
mise run optional-apply
kubectl --context kind-learning-lab -n learning-lab get pdb,hpa
```

- PodDisruptionBudget: protects availability during voluntary disruptions.
- HorizontalPodAutoscaler: changes desired replicas from CPU metrics.

The kind lab does not install Metrics Server by default, so the HPA may show `<unknown>` until a metrics provider exists. That is an intentional exercise boundary, not a broken manifest.

## How the pieces fit

```text
Git commit
   |
   v
Argo CD Application
   |
   v
Rollout CRD --------------------> Argo Rollouts controller
   |                                         |
   v                                         v
desired Pod template              ReplicaSets + Services
                                             |
                                             v
                                    pause / analyze / promote
```

Argo CD answers “which manifests should be live?” Argo Rollouts answers “how should a changed Pod template become live?” Keeping those responsibilities separate is the useful production mental model.
