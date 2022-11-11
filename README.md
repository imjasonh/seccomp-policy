# Seccomp Profile Distribution Controller

[![GoDoc](https://godoc.org/github.com/imjasonh/seccomp-profile?status.svg)](https://godoc.org/github.com/imjasonh/seccomp-profile)

This is a work in progress.

## Testing

```
kind create cluster --config=kind.yaml
```

This will create a KinD cluster with built-in seccomp profiles as described here: https://kubernetes.io/docs/tutorials/security/seccomp

Then install the components:

```
KO_DOCKER_REPO=kind.local ko apply -f config/
```

(On Apple Silicon this also needs `--platform=linux/arm64`)

Check that the components are up:

```
$ kubectl get pods -n seccomp-profile
NAME                       READY   STATUS    RESTARTS   AGE
controller-pg4fg           1/1     Running   0          11m
controller-ssshb           1/1     Running   0          11m
webhook-65f995489c-chvll   1/1     Running   0          22m
```

Create a `SeccompProfile` resource:

```
$ kubectl apply -f profiles/audit.yaml
seccompprofile.seccomp.imjasonh.dev/2aff5d4800d3c60f930b6f10188bc2ca4d366359246bec86de6dc0d6ed61d818 unchanged
```

(This one corresponds to the [`audit.json`](./profiles/audit.json) policy)

This will result in the profile being synced to the necessary location on all nodes.
