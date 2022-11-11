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
controller-xq44p           1/1     Running   0          2m37s
webhook-6fc6f4dc54-6kq6g   1/1     Running   0          10m
```
