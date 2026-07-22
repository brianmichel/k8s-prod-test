# Troubleshooting

## `mise run doctor` fails

Install the pinned tools first:

```sh
mise run install
```

Docker Desktop must be running because kind creates Kubernetes nodes as containers. `kubectl`, `kind`, `helm`, `jq`, and `yq` are expected to come from mise.

## `kind create cluster` says Docker is unavailable

Run `docker info`. If it fails, start Docker Desktop and retry `mise run cluster-up`.

## `ImagePullBackOff` or `ErrImageNeverPull`

The Deployment intentionally uses `imagePullPolicy: Never` for a local image. Rebuild and load it:

```sh
mise run image-build
mise run image-load
kubectl --context kind-learning-lab -n learning-lab rollout restart deployment/playground-api
```

## Pods are Running but not Ready

Inspect the probe and events:

```sh
kubectl --context kind-learning-lab -n learning-lab describe pod -l app=playground-api
mise run logs
```

The default readiness delay is three seconds. A longer delay is expected if you changed `READY_AFTER_SECONDS`.

## The port-forward is already in use

Stop the old `mise run port-forward` process, or use a different local port:

```sh
kubectl --context kind-learning-lab -n learning-lab port-forward service/playground-api 18080:8080
```

## The conversation source is not visible

The shared URLs currently redirect to the logged-out ChatGPT home in this environment. Paste the conversation text or make the share available in a signed-in browser session before treating its specific concepts as reviewed.

