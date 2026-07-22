# Exercises

Do these in order. Each one has an observable outcome and a reset path.

## 1. Find the objects

```sh
mise run status
kubectl --context kind-learning-lab -n learning-lab get deploy,rs,pods,svc,cm,secret -o wide
```

Draw the ownership chain: Deployment -> ReplicaSet -> Pods, and separately Service -> selected Pods.

## 2. Watch readiness

Restart the deployment and watch Pods become ready:

```sh
kubectl --context kind-learning-lab -n learning-lab rollout restart deployment/playground-api
kubectl --context kind-learning-lab -n learning-lab get pods --watch
```

While a Pod is warming up, compare `STATUS`, `READY`, and the Service endpoints.

## 3. Prove Service discovery

Run the port-forward, call `/` repeatedly, and observe the Pod name. Then delete one Pod and keep calling the Service. The client should continue working while the ReplicaSet replaces the Pod.

## 4. Change configuration

Edit `k8s/base/configmap.yaml`, change `APP_MESSAGE`, run `mise run apply`, and call the endpoint again. Notice that environment variables are read at process startup, so a changed ConfigMap does not magically mutate an existing process.

To make the new value visible, restart the Deployment:

```sh
kubectl --context kind-learning-lab -n learning-lab rollout restart deployment/playground-api
kubectl --context kind-learning-lab -n learning-lab rollout status deployment/playground-api
```

## 5. Break readiness safely

Change `READY_AFTER_SECONDS` to `20`, apply, and inspect the rollout. The startup probe allows the process to remain alive while readiness stays false. Restore it to `3` when finished.

## 6. Inspect a rolling update

Change `APP_MESSAGE`, apply, and watch:

```sh
kubectl --context kind-learning-lab -n learning-lab get pods --watch
kubectl --context kind-learning-lab -n learning-lab get events --sort-by=.lastTimestamp
```

Explain why `maxUnavailable: 0` affects the order of replacement.

## 7. Experiment with resources

Temporarily set the memory limit to `8Mi` or set an impossible CPU request. Observe the Pod or scheduler events, then restore the manifest and run `mise run deploy`.

## 8. Inspect the rendered object graph

Kustomize is built into `kubectl`:

```sh
kubectl kustomize k8s/overlays/local | yq 'select(.kind != "Secret")'
```

Use `yq` and `jq` as inspection tools, not as required application dependencies.

## 9. Compare imperative deployment with GitOps

`mise run apply` sends manifests from your working tree directly to Kubernetes. Argo CD should instead read the same `k8s/overlays/local` path from Git:

```sh
mise run argocd-install
# Set repoURL in gitops/argocd/playground-api.yaml to a reachable Git repository.
mise run argocd-application
mise run argocd-get
```

Write down which system knows about the change in each flow and where an audit trail exists.

## 10. Observe self-healing

After the Argo CD Application is synced, make an out-of-band change:

```sh
kubectl --context kind-learning-lab -n learning-lab scale deployment/playground-api --replicas=1
kubectl --context kind-learning-lab -n learning-lab get deployment/playground-api --watch
```

Explain why the Deployment returns to two replicas and which controller made that happen.

## 11. Observe prune

Create a harmless extra ConfigMap in the managed overlay, commit and push it, sync it, then remove it from Git and wait for Argo CD to prune it. This is the operational difference between "apply what is present" and "converge to exactly what Git declares."

## 12. Run a blue-green release

```sh
mise run rollouts-install
mise run bluegreen-apply
mise run bluegreen-preview
```

Change the blue-green `APP_VERSION`, apply the directory, inspect the preview response, and promote only after the smoke AnalysisRun passes.

## 13. Run a canary release

```sh
mise run canary-apply
mise run canary-status
```

Change the canary `APP_VERSION`, apply it, then promote through the 25% and 50% pauses. Explain why four replicas cannot produce precise 25% request traffic without a traffic router.

## 14. Compare rollback mechanisms

Compare `kubectl rollout undo deployment/playground-api` with `mise run bluegreen-abort` and `mise run canary-abort`. Identify which object owns the rollout state in each case and where the old ReplicaSet remains available.

## 15. Explore reliability constraints

Apply the optional PDB/HPA overlay, inspect the resources, and explain why the PDB can work without Metrics Server while the HPA cannot make a decision without metrics.

## Stretch goals

- Add an Ingress and compare it with a Service.
- Add a PodDisruptionBudget and reason about voluntary disruption.
- Add a second version of the image and practice rollback.
- Add a NetworkPolicy, then prove the difference between allowed and denied traffic.
- Add a Helm chart only after you understand the plain YAML; compare the two representations.
