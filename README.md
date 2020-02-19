# kt ![](https://github.com/knight42/kt/workflows/Cross%20Platform%20Build/badge.svg)

kt is short for Kubernetes Tail. It behaves like `kubect logs -f` and
its usage is similar to `kubectl get`.

# Table of Contents

* [0. Features](#0-features)
* [1. Usage](#1-usage)
    * [1.1 Install bash/zsh completion](#11-install-bashzsh-completion)
    * [1.2 Filter pods by name or regexp](#12-filter-pods-by-name-or-regexp)
    * [1.3 Filter pods by labels](#13-filter-pods-by-labels)
    * [1.4 Tails pods belong to a higher level object](#14-tails-pods-belong-to-a-higher-level-object)
* [2. Installtion](#2-installtion)

# 0. Features

* Tail all containers in a pod by default.
* Automatically tail new pods, discard deleted pods and retry if the pod
switches to running phase from pending phase.
* Recover from containers restart.
* Auto completion.
* Colorized output.

# 1. Usage

#### 1.1 Install bash/zsh completion

> NOTE: If you install `kt` using homebrew, the completion is already installed.

Load the completion code into current shell
```
## bash
$ source <(kt --completion bash)
## zsh
$ source <(kt --completion zsh)
```

#### 1.2 Filter pods by name or regexp

```
$ kt foo
```
or

```
$ kt 'foo-\w+'
```

#### 1.3 Filter pods by labels

```
$ kt -n prod -lapp=foo
```

#### 1.4 Tail pods belong to a higher level object

Currently only the following resources are supported:
* Service
* Deployment
* StatefulSet
* DaemonSet
* HPA
* Job
* ReplicaSet
* ReplicationController
* Cronjob(partially supported. You must specify labels in the pod template.)

```
$ kt hpa foo

# You could limit which containers are tailed using regexp
$ kt -c 'sidecar-\w' svc foo

$ kt -n test --tail 30 deploy foo

$ kt --timestamps sts foo

$ kt --context prod ds foo

$ kt --cluster dev job foo
```

# 2. Installtion

Using Homebrew:
```
$ brew tap knight42/tap
$ brew install knight42/tap/kt
```

Or download from the [release page](https://github.com/knight42/kt/releases).
