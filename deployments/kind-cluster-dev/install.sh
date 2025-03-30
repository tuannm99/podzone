# create kind cluster
kind create cluster --config single-cluster.yaml

# certmanager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml

# metallb loadbalancer and IP pools
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.9/config/manifests/metallb-native.yaml
kubectl wait --namespace metallb-system \
  --for=condition=ready pod \
  --selector=app=metallb,component=controller \
  --timeout=90s
kubectl wait --namespace metallb-system \
  --for=condition=ready pod \
  --selector=app=metallb,component=speaker \
  --timeout=90s
kubectl apply -f ./metallb

# longhorn storageclass
kubectl apply -f https://raw.githubusercontent.com/longhorn/longhorn/v1.7.0/deploy/longhorn.yaml

# ingress (we using nginx-ingress)
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace

# rancher --> just dont need for dev
helm repo add rancher-latest https://releases.rancher.com/server-charts/latest
kubectl create namespace cattle-system

helm upgrade --install rancher rancher-latest/rancher \
  --namespace cattle-system \
  --set hostname=rancher.local.com \
  --set replicas=1 \
  --set bootstrapPassword=pwd \
  --set ingress.tls.source=letsEncrypt \
  --set letsEncrypt.email=tesmail.gmail.com \
  --set letsEncrypt.ingress.class=nginx

# todo ------------

#  -----------------

# applying ingress
kubectl apply -f ./ingress-local.yaml
