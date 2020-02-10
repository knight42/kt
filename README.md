# kt

kt is short for Kubernetes Tail. It behaves like `kubect logs -f` and 
its syntax is similar to `kubectl get`.

# 0. Features

* Tails all containers in a pod by default.
* Automatically tails new pods, discard deleted pods and retries if pod
switches to running phase from pending phase.
* Recover from containers restart.
* Auto completion.
* Colorized output.

# 1. Usage

#### 1.1 Filter pods by name or regexp

```
$ kt foo
```
or

```
$ kt 'foo-\w+'
```

#### 1.2 Filter pods by labels

```
$ kt -n prod -lapp=foo
```

#### 1.3 Tails pods belong to a higher level object

Currently only the following resources are supported:
* Service
* Deployment
* StatefulSet
* DaemonSet
* Job
* ReplicaSet
* ReplicationController
* Cronjob(partially supported. You must specify labels in the pod template.)

TODO:
* HPA

```
# You could limit which containers are tailed using regexp
$ kt --tail 10 -c 'sidecar-\w' svc foo

$ kt -n test --tail 10 deploy foo

$ kt --tail 10 --timestamps sts foo

$ kt --context prod --tail 10 ds foo

$ kt --cluster dev --tail 10 job foo

$ kt --tail 10 rs foo
```


