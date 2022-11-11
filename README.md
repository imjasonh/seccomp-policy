# Seccomp Profile Distribution Controller

[![GoDoc](https://godoc.org/github.com/imjasonh/seccomp-profile?status.svg)](https://godoc.org/github.com/imjasonh/seccomp-profile)

This is a work in progress.

This provides a CRD and controller components to distribute seccomp profiles to nodes, automating the process described in Kubernetes docs: https://kubernetes.io/docs/tutorials/security/seccomp

Seccomp profiles allow users to audit or limit system calls ("syscalls") used by containers.

## Try it out

If you don't have a cluster, you can create one with KinD:

```
kind create cluster --config=kind.yaml
```

This creates a cluster with 2 worker nodes, to demonstrate that profiles are distributed to _every_ node.

I've also tested this on a GKE cluster.

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

...and a Pod that uses the `violation` policy; this should fail:

```
$ kubectl create -f pods/violation-pod.yaml
pod/violation-pod-jsmkp created
$ kubectl get pod violation-pod-jsmkp
NAME                  READY   STATUS       RESTARTS   AGE
violation-pod-jsmkp   0/1     StartError   0          4s
```

## Future Work

Container images could distribute their seccomp profiles in their metadata.
If they did, the webhook component could extract these profiles from incoming images, and create `SeccompPolicy` resources, and mutate `PodSpec`s to use those policies.

An image build tool could determine the seccomp profile based on source analysis, or hand-curated overrides, and distribute those profiles with the image.
