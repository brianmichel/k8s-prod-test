# Kubernetes learning lab

This is a small, real-ish Kubernetes playground for learning the control-plane, workload, and GitOps delivery concepts that show up in production. The application is intentionally boring: a Go HTTP server with health endpoints, configuration, a secret, pod identity, structured logs, and graceful shutdown.

The local lab uses Docker plus [kind](https://kind.sigs.k8s.io/). `mise` owns the versions of the CLI tools so the workflow is repeatable.

## Start here

```sh
mise run install
mise run doctor
mise run deploy
mise run argocd-install
mise run rollouts-install
```

In another terminal, expose the service:

```sh
mise run port-forward
```

Then visit [http://127.0.0.1:8080](http://127.0.0.1:8080), or query it with:

```sh
curl -s http://127.0.0.1:8080 | jq
```

You should see the ConfigMap message, the pod name/namespace/IP, and `secretConfigured: true`.

## Workflow

| Command | Purpose |
| --- | --- |
| `mise run install` | Install the pinned Go/Kubernetes tooling |
| `mise run doctor` | Check Docker and every CLI used by the lab |
| `mise run deploy` | Create kind, build/load the image, apply manifests, wait for rollout |
| `mise run argocd-install` | Install the pinned Argo CD control plane |
| `mise run argocd-application` | Register the GitOps Application resource |
| `mise run argocd-bluegreen-application` | Register the blue-green GitOps Application |
| `mise run argocd-canary-application` | Register the canary GitOps Application |
| `mise run argocd-get` | Inspect Application health and sync state |
| `mise run argocd-ui` | Reach the Argo CD UI at `https://127.0.0.1:8081` |
| `mise run argocd-password` | Print the initial local Argo CD admin password |
| `mise run rollouts-install` | Install the Argo Rollouts controller and CRDs |
| `mise run flightdeck-deploy` | Build and deploy the Flightdeck delivery dashboard |
| `mise run flightdeck-ui` | Reach Flightdeck at `http://127.0.0.1:8090` |
| `mise run bluegreen-apply` | Apply the blue-green rollout example |
| `mise run canary-apply` | Apply the canary rollout example |
| `mise run status` | Inspect workloads, service, ConfigMap, and Secret metadata |
| `mise run logs` | Follow application logs |
| `mise run port-forward` | Reach the ClusterIP service from your laptop |
| `mise run shell` | Try an interactive command inside a running container |
| `mise run test` | Run Go tests and render the Kustomize overlay |
| `mise run clean` | Delete only the lab namespace |
| `mise run cluster-down` | Delete the whole kind cluster |

Read [docs/concepts.md](docs/concepts.md) before changing anything, then work through [docs/exercises.md](docs/exercises.md). For the delivery path, continue with [docs/argocd.md](docs/argocd.md) and [docs/progressive-delivery.md](docs/progressive-delivery.md).

## Repository map

```text
cmd/playground-api/     application and unit test
infra/kind.yaml         two-node local cluster definition
k8s/base/               reusable Kubernetes resources
k8s/overlays/local/     local namespace and environment overlay
gitops/argocd/           Argo CD Application declarations
k8s/progressive/         blue-green and canary Rollout examples
k8s/optional/            PDB and HPA examples that are not applied by default
docs/                   concept notes, exercises, and troubleshooting
.mise.toml              tool pins and the lab command surface
```

## Safety and cleanup

This lab is local-only. The Secret is demo data and is committed as `stringData` so you can see how the manifest works; do not copy that pattern into a real repository. `mise run clean` removes the namespace, while `mise run cluster-down` removes the disposable kind cluster.

## Deployments and GitOps

There are now two deployment paths to compare:

1. `mise run apply` is an operator-driven deployment: your terminal sends the manifests to the API server.
2. Argo CD is a controller-driven deployment: an `Application` declares a Git repository and path, then Argo CD continuously reconciles that source into the cluster.

The `Application` currently contains a placeholder Git URL because this checkout has no remote repository. Replace `repoURL` in [gitops/argocd/playground-api.yaml](gitops/argocd/playground-api.yaml) with a reachable repository before registering it. See [docs/argocd.md](docs/argocd.md) for the full flow.

## Flightdeck dashboard

Flightdeck is a deliberately thin dashboard over the resources already in the cluster. It displays Argo CD Applications, Argo Rollouts, Deployments, and pod readiness without introducing a new environment or promotion model. It can request an Argo CD sync and promote or abort a Rollout using its service account RBAC.

```sh
mise run flightdeck-deploy
mise run flightdeck-ui
```

Open [http://127.0.0.1:8090](http://127.0.0.1:8090). The dashboard reads all namespaces so it can correlate the `argocd` control plane with workloads in `learning-lab`; its write access is limited to patching Applications and Rollouts. The frontend is a Vue 3 application in `frontend/`; daisyUI provides its shared controls and custom Flightdeck theme, while Vue Flow handles viewport and edge rendering and Dagre provides automatic Kubernetes ownership layout. The production container builds and embeds the frontend into the Go server.

Progressive delivery is separate from the baseline Deployment. Install Argo Rollouts, then choose one experiment:

```sh
mise run rollouts-install
mise run bluegreen-apply
# or: mise run canary-apply
```

See [docs/progressive-delivery.md](docs/progressive-delivery.md) for promotion, abort, preview traffic, analysis, and the limits of local canary traffic splitting.

## Source conversation review

The supplied ChatGPT share links redirected to the logged-out ChatGPT home in the available browser session, so the original conversation was not readable and is not being represented as reviewed here. The lab is therefore a Kubernetes-first foundation based on this checkout’s name and can be aligned to the conversation once its contents are pasted or made accessible.
