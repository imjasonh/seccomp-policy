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

Create some `SeccompProfile` resources:

```
$ kubectl apply -f profiles/
seccompprofile.seccomp.imjasonh.dev/audit unchanged
seccompprofile.seccomp.imjasonh.dev/fine-grained unchanged
seccompprofile.seccomp.imjasonh.dev/violation unchanged
```

Then create a Pod that uses the `audit` policy:

```
$ kubectl create -f pods/audit-pod.yaml
pod/audit-pod-wczbg created
$ kubectl get pod audit-pod-wczbg
NAME              READY   STATUS      RESTARTS   AGE
audit-pod-wczbg   0/1     Completed   0          9s
```

...and a Pod that uses the `violation` policy:

```
$ kubectl create -f pods/violation-pod.yaml
pod/violation-pod-jsmkp created
$ kubectl get pod violation-pod-jsmkp
NAME                  READY   STATUS       RESTARTS   AGE
violation-pod-jsmkp   0/1     StartError   0          4s
```
