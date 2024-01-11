## install container software

install docker

```
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done
# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```


install kind

```
# For AMD64 / x86_64
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
# For ARM64
[ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-arm64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

install kubectl

```
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

## create cluster

create cluster

```
kind create cluster
```

## create namespace with PSP/PSA

```
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: flag-reciever
  labels:
    pod-security.kubernetes.io/enforce: "restricted"
EOF

kubectl create namespace flag-sender
```

## create user

```
kubectl create serviceaccount ctf-player
cat << EOF > kubeconfig-ctf-player.yaml
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: $(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
  name: ctf-cluster
contexts:
- context:
    cluster: ctf-cluster
    user: ctf-player
  name: ctf-cluster
current-context: ctf-cluster
kind: Config
preferences: {}
users:
- name: ctf-player
  user:
    token: $(kubectl create token ctf-player)
EOF
```

## create pod

```
kubectl create secret generic flag --from-literal=flag=potluck{kubernetes_can_be_a_bit_weird} --namespace=flag-sender
cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flag-sender
  namespace: flag-sender
  labels:
    app: flag-sender
spec:
  replicas: 1
  selector:
    matchLabels:
      app: flag-sender
  template:
    metadata:
      labels:
        app: flag-sender
    spec:
      containers:
      - name: container
        image: busybox
        imagePullPolicy: IfNotPresent
        command:
        - sh
        args:
        - -c
        - while true; do echo \$FLAG | nc 1.1.1.1 80 || continue; echo 'Flag Send'; sleep 10; done
        env:
        - name: FLAG
          valueFrom:
            secretKeyRef:
              name: flag
              key: flag
      restartPolicy: Always
      serviceAccountName: default
      dnsPolicy: ClusterFirst
      securityContext: {}
EOF
```

## create network policy

```
cat << EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: flag-reciever
  namespace: flag-reciever
spec:
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          ns: flag-sender
      podSelector:
        matchLabels:
          app: flag-sender
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF
```

## setup RBAC

give user 

* list namespace
* pods
* svc

```
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ctf-player
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["list"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ctf-player
subjects:
- kind: ServiceAccount
  name: ctf-player
  namespace: default
roleRef:
  kind: ClusterRole
  name: ctf-player
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: flag-reciever
  name: ctf-player
rules:
- apiGroups: [""]
  resources: ["pods.*"]
  verbs: ["create", "delete"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "create", "delete"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["services.*"]
  verbs: ["create", "delete"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "create", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies"]
  verbs: ["list", "get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ctf-player
  namespace: flag-reciever
subjects:
- kind: ServiceAccount
  name: ctf-player
  namespace: default
roleRef:
  kind: Role
  name: ctf-player
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: flag-sender
  name: ctf-player
rules:
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ctf-player
  namespace: flag-sender
subjects:
- kind: ServiceAccount
  name: ctf-player
  namespace: default
roleRef:
  kind: Role
  name: ctf-player
  apiGroup: rbac.authorization.k8s.io
EOF
```

## Clean

```
kubectl get ns --no-headers -o custom-columns=:metadata.name | xargs -n 1 kubectl delete sa ctf-player -n
kubectl get ns --no-headers -o custom-columns=:metadata.name | xargs -n 1 kubectl delete role ctf-player -n
kubectl get ns --no-headers -o custom-columns=:metadata.name | xargs -n 1 kubectl delete rolebinding ctf-player -n
kubectl delete clusterrole ctf-player
kubectl delete clusterrolebinding ctf-player
```

## all-in-one

```
kind create cluster
kubectl apply -f challenge.yaml
cat << EOF > kubeconfig-ctf-player.yaml
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: $(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
  name: ctf-cluster
contexts:
- context:
    cluster: ctf-cluster
    user: ctf-player
  name: ctf-cluster
current-context: ctf-cluster
kind: Config
preferences: {}
users:
- name: ctf-player
  user:
    token: $(kubectl create token ctf-player)
EOF
```