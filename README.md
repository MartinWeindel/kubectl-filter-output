# kubectl-filter-output

Plugin for kubectl to filter uninteresting metadata fields for humans from JSON or YAML output.

## Installation

### Compile From Source

These instructions assume you have go installed and a `$GOPATH` set. 
Just run

```bash
make install
```

The binary is installed in `$GOPATH/bin`. If this directory is on your `$PATH`, you are done.
Otherwise you have to copy the binary in a directory of your `$PATH`.

## Usage

The plugin can be used whenever `kubectl` is used with *YAML* or *JSON* output.

```
kubectl fo [--filter <expr1>,<expr2>] ...
```

Command line examples:
```bash
kubectl fo get svc kubernetes -o yaml
kubectl fo get pod -o json
kubectl fo --filter metadata,-metadata.label,status get svc kubernetes -oyaml
```

If no `--filter` is specified, only the default filters are applied.
These fields are always removed from the output:

- `metadata.managedFields`

- `metadata.selfLink`

- `metadata.annotations.'kubectl.kubernetes.io/last-applied-configuration'`

### `--filter` / `-f` option

With the `--filter <filter-expression>` or `-f <filter-expression>` option additional fields can be filtered.
This is simple alternative for using the advanced `-o jsonpath` output.

The filter expression is a comma separated list of simple JSON path field expressions.
Each item can be prefixed with `+` or `-` optionally to keep or remove the fields respectively.
Keeping (`+`) is the default if not specified.

Filter expression examples:

1. `--filter metadata.name,metadata.namespace` only keeps `metadata.name` and `metadata.namespace`.
It is the same as `--filter +metadata.name,+metadata.namespace`

2. `--filter -status` drops the status field of on object

3. `--filter metadata,-metadata.labels,-metadata.annotations` keeps the metadata field but drops labels and annotations

## Examples

### Before
```yaml
$ kubectl get svc kubernetes -oyaml
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2021-01-07T10:33:32Z"
  labels:
    component: apiserver
    provider: kubernetes
  managedFields:
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:labels:
            .: {}
            f:component: {}
            f:provider: {}
        f:spec:
          f:clusterIP: {}
          f:ports:
            .: {}
            k:{"port":443,"protocol":"TCP"}:
              .: {}
              f:name: {}
              f:port: {}
              f:protocol: {}
              f:targetPort: {}
          f:sessionAffinity: {}
          f:type: {}
      manager: kube-apiserver
      operation: Update
      time: "2021-01-07T10:33:32Z"
  name: kubernetes
  namespace: default
  resourceVersion: "154"
  selfLink: /api/v1/namespaces/default/services/kubernetes
  uid: e25e7b7f-1d86-4f3d-aabd-9ab373dfa958
spec:
  clusterIP: 100.64.0.1
  ports:
    - name: https
      port: 443
      protocol: TCP
      targetPort: 443
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```

### With default filters

```yaml
$ kubectl fo get svc kubernetes -oyaml
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2021-01-07T10:33:32Z"
  labels:
    component: apiserver
    provider: kubernetes
  name: kubernetes
  namespace: default
  resourceVersion: "154"
  uid: e25e7b7f-1d86-4f3d-aabd-9ab373dfa958
spec:
  clusterIP: 100.64.0.1
  ports:
    - name: https
      port: 443
      protocol: TCP
      targetPort: 443
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```

### With additional filters

```yaml
$ kubectl fo --filter metadata,-metadata.label,status get svc kubernetes -oyaml
metadata:
  creationTimestamp: "2021-01-07T10:33:32Z"
  labels:
    component: apiserver
    provider: kubernetes
  name: kubernetes
  namespace: default
  resourceVersion: "154"
  uid: e25e7b7f-1d86-4f3d-aabd-9ab373dfa958
status:
  loadBalancer: {}
```

Note that the filters always work on the items of a `List`

```yaml
$ kubectl fo --filter metadata,-metadata.labels,-metadata.annotations,spec get node -oyaml
apiVersion: v1
items:
- metadata:
    creationTimestamp: "2021-01-07T10:35:34Z"
    name: my-node-1234
    resourceVersion: "716271"
    uid: 4f424ebb-8e2c-42b4-9e80-c7f15bbc4fb9
  spec:
    podCIDR: 100.96.0.0/24
    podCIDRs:
    - 100.96.0.0/24
    providerID: vsphere://42070914-70ac-7214-0886-69104a2dc946
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```
