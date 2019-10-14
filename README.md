# Ordered Chaos Monkey Operator
This k8 operator is created for demonstrating purpose using the [operator-sdk](https://github.com/operator-framework/operator-sdk). 
It introduces a new Custom Ressource (Definition) "PodReplica". 
It works like a regular k8 deployment just with a predefined busybox image which sleeps.

### Run locally (instead of Build & Publish)
Register CRD:

`kubectl create -f deploy/crds/app_v1alpha1_podreplica_crd.yaml`

Set operator name as environment variable

`export OPERATOR_NAME=podreplica`

Start:

`operator-sdk up local --namespace=default`

Start with debugging:

`operator-sdk up local --namespace=default --enable-delve`


### Create Custom Resource
After operator is running you can create a podreplica, for example with:

`kubectl create -f deploy/crds/app_v1alpha1_podreplica_cr.yaml`

### Build & Publish Operator (on dockerhub)
```
# On Linux
operator-sdk build <user>/podreplica:v0.0.1
sed -i 's|REPLACE_IMAGE|<user>/podreplica:v0.0.1|g' deploy/operator.yaml
docker push <user>/podreplica:v0.0.1

# On OSX
sed -i "" 's|REPLACE_IMAGE|<user>/podreplica-operator:v0.0.1|g' deploy/operator.yaml
docker push <user>/podreplica-operator:v0.0.1
```

### Create PodReplica Operator
Register CRD:
`kubectl create -f deploy/crds/app_v1alpha1_podreplica_crd.yaml`

Create RBAC and Operator:
```
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml
```

Afterwards you can follow [Create Custom Resource](#create-custom-resource)
