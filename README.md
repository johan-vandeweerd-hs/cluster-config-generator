# README

Small test project to play
with [Argocd plugin generators](https://argo-cd.readthedocs.io/en/latest/operator-manual/applicationset/Generators-Plugin/)
and [ApplicationSets](https://argo-cd.readthedocs.io/en/latest/operator-manual/applicationset/). The cluster config
plugin generator can be used to store environment specific parameters in a configmap that can later on be used as Helm
parameters in an ApplicationSet.

***CAUTION*** This project is just a quick hack and not production ready (but could be with some extra work).

## Prerequisites

- [ ] Docker
- [ ] Kind

## Steps

* Create a Kind cluster `kind create cluster --config kind-config.yaml`
* Build and install the `cluster-config-generator` image into the Kind cluster:
  `docker image build -t cluster-config-generator:0.0.1 image/cluster-config-generator && kind load docker-image cluster-config-generator:0.0.1`
* Deploy Argocd:
  `helm dep up ./gitops/argocd && helm upgrade --install -n argocd --create-namespace argocd ./gitops/argocd`
* Deploy the `cluster-config-generator` service: `kubectl apply -n argocd -f ./gitops/cluster-config-generator`
* Create a cluster configuration ConfigMap and role+rolebinding for the `cluster-config-generator` service:
  `kubectl apply -n kube-system -f ./gitops/kube-system`
* Deploy the `apps` ApplicationSet that generates three apps and has the `clusterName` parameter passed as Helm
  parameters: `kubectl apply -n argocd -f ./gitops/app-of-apps`
* Check the generated apps: `kubectl get application -n argocd -oyaml`

## How to access Argocd

The Kind cluster will port forward the Argocd service to localhost on 8443. The admin password can be retrieved with the
following command:
`kubectl get secret -n argocd argocd-initial-admin-secret  -ojson | jq -r '.data["password"]' | base64 -d | clipcopy`.
Go to [the ArgoCD UI](https://localhost:8443) and login with `admin` and the password from the previous command.

## Teardown

To remove the Kind cluster and the Argocd resources run `kind delete cluster`.
