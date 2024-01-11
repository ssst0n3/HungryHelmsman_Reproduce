# HungryHelmsman_Reproduce

Reproduce for the CTF challenge [37C3 Potluck CTF-Qualifier-2023/Misc/Hungry Helmsman](https://ctftime.org/task/27382)

## prepare environment

```
$ kind create cluster
$ kubectl apply -f challenge.yaml
$ cat << EOF > kubeconfig-ctf-player.yaml
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