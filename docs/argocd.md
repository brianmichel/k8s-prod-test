# Argo CD and GitOps

Argo CD is the delivery controller in this lab. It watches an `Application` resource, reads the declared Git revision and path, renders the manifests, and reconciles the live cluster toward that desired state.

## The deployment boundary

```text
Git commit
   |
   v
Argo CD Application
   |
   v
Argo CD repo-server -> Kustomize -> Kubernetes API
                                      |
                                      v
                              Deployment -> Pods
```

The existing `mise run apply` command skips the Git and Argo CD steps. It is useful for learning Kubernetes directly, but it is not GitOps. `argocd-application.yaml` is the declarative handoff between the repository and the controller.

## Install Argo CD locally

The task uses the pinned Argo CD `v3.4.2` release and server-side apply:

```sh
mise run argocd-install
```

The upstream installation is intentionally not vendored into this repository. The task uses the versioned official manifest URL so the installed control plane is explicit and repeatable. Server-side apply is important for Argo CD CRDs because some are too large for client-side apply annotations.

Check the components:

```sh
mise run argocd-status
```

## Access the UI

In one terminal:

```sh
mise run argocd-ui
```

Open [https://127.0.0.1:8081](https://127.0.0.1:8081) and accept the local self-signed certificate warning. In another terminal, get the initial password:

```sh
mise run argocd-password
```

Log in as `admin`. This is disposable local-cluster access; change or delete the initial credential before using an environment that matters.

## Register the Application

Argo CD cannot read this workspace’s local filesystem directly. It needs a Git repository reachable from the Argo CD repo-server. Push this repository to a Git host, then replace the placeholder `repoURL` in [gitops/argocd/playground-api.yaml](../gitops/argocd/playground-api.yaml):

```yaml
source:
  repoURL: https://github.com/your-user/k8s-prod-test.git
  targetRevision: main
  path: k8s/overlays/local
```

Then register it:

```sh
mise run argocd-application
mise run argocd-get
```

The destination uses `https://kubernetes.default.svc`, which means the Application targets the same cluster where Argo CD is running. The namespace must already exist because the Application uses `CreateNamespace=false`.

## Sync behavior

The Application enables:

- `automated`: reconcile without clicking Sync for each Git change.
- `prune`: remove objects that disappear from Git.
- `selfHeal`: restore live objects changed out-of-band.
- `retry`: retry transient render or API failures.

Try the contrast:

```sh
kubectl --context kind-learning-lab -n learning-lab scale deployment/playground-api --replicas=1
mise run argocd-get
```

With self-healing active, Argo CD should restore the declared two replicas. Then edit `APP_MESSAGE` in Git, commit, and push. Argo CD should render the new Kustomize output and roll the Deployment.

## Core mode versus server mode

The `argocd-get` and `argocd-sync` tasks use `--core`, which lets the CLI talk to Kubernetes directly for a local exercise without saving an Argo CD login session. The UI and `argocd-password` tasks expose the normal server-based workflow so you can learn both surfaces.

## Production differences to keep in mind

This lab uses the default Argo CD project and a public-style Git URL for simplicity. Real teams should add repository credentials, Projects with least privilege, signed or protected branches, separate environment overlays, image promotion policy, RBAC, TLS, backups, and resource limits for Argo CD itself.

