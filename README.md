# GitOps zombies

This simple tool will help you find kubernetes resources which are not managed via GitOps (flux2).

![zombie gopher](assets/logo.png)

## How does it work?

It will discover all apis installed on a cluster and identify resources which are not part of a Kustomization or a HelmRelease.
The app will also acknowledge the following things:

* Ignores resources which are owned by a parent resource (For example pods which are created by a deployment)
* Ignores resources which are considered dynamic (metrics, leases, events, endpoints, ...)
* Filter out resources which are created by the apiserver itself (like default rbacs)
* Filters secrets which are managed by other parties including helm or ServiceAccount tokens
* Checks if the referenced HelmRelease or kustomization exists


## How do I install it?

```
brew install raffis/gitops-zombies
```

## How to use

Exclude all resources which are not labeled `key: value`:
```
gitops-zombies -l "key!=value"
```

## CLI reference
```
Finds all kubernetes resources from all installed apis on a kubernetes cluste and evaluates whether they are managed by a flux kustomization or a helmrelease.

Usage:
  gitops-zombies [flags]

Flags:
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "/home/raffi/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
  -h, --help                           help for gitops-zombies
  -a, --include-all                    Includes resources which are considered dynamic resources
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
  -o, --output string                  Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file, custom-columns, custom-columns-file, wide). See custom columns [https://kubernetes.io/docs/reference/kubectl/overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [https://kubernetes.io/docs/reference/kubectl/jsonpath/].
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -l, --selector string                Label selector (Is used for all apis)
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
  -v, --verbose                        Verbose mode (Logged to stderr)
      --version                        Print version and exit
```