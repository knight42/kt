# kt ![](https://github.com/knight42/kt/workflows/CI/badge.svg)

kt is short for Kubernetes Tail. It behaves like `kubectl logs -f` and
its usage is similar to `kubectl get`.

# Table of Contents

* [0. Features](#0-features)
* [1. Usage](#1-usage)
    * [1.1 Install bash/zsh completion](#11-install-bashzsh-completion)
    * [1.2 Filter pods by name or regexp](#12-filter-pods-by-name-or-regexp)
    * [1.3 Filter pods by labels](#13-filter-pods-by-labels)
    * [1.4 Tail pods belong to a higher level object](#14-tails-pods-belong-to-a-higher-level-object)
    * [1.5 Filter logs by query](#15-filter-logs-by-query)
    * [1.6 Prefix mode](#16-prefix-mode)
* [2. Installation](#2-installation)

# 0. Features

* Tail all containers in a pod by default.
* Automatically tail new pods, discard deleted pods and retry if the pod
switches to running phase from pending phase.
* Recover from containers restart.
* Filter logs by query DSL with `and`, `or`, parentheses, and quoted strings.
* Auto-hide pod/container prefix when tailing a single container.
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
* CronJob

```
$ kt hpa foo

# You could limit which containers are tailed using regexp
$ kt -c 'sidecar-\w' svc foo

$ kt -n test --tail 30 deploy foo

$ kt --timestamps sts foo

$ kt --context prod ds foo

$ kt --cluster dev job foo
```

#### 1.5 Filter logs by query

Use `-q/--query` to filter log lines with a boolean DSL:

```
# Match lines containing both keywords
$ kt deploy foo -q 'error and timeout'

# Match lines containing either keyword
$ kt deploy foo -q 'error or warning'

# Use parentheses to group expressions (and binds tighter than or)
$ kt deploy foo -q '(error or warn) and timeout'

# Use quotes for keywords with spaces
$ kt deploy foo -q '"error code" and 500'
```

#### 1.6 Prefix mode

The `--prefix` flag controls pod/container prefix display:

* `auto` (default): hide prefix when tailing a single pod with a single container
* `always`: always show the prefix
* `off`: never show the prefix

```
$ kt deploy foo --prefix=always
$ kt deploy foo --prefix=off
```

# 2. Installation

Using Homebrew:
```
$ brew tap knight42/tap
$ brew install knight42/tap/kt
```

Or download from the [release page](https://github.com/knight42/kt/releases).
