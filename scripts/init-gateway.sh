kubectl delete statefulset etcd
kubectl delete pvc data-etcd-0

# (2) Chạy 1 pod riêng để restore
kubectl run etcd-restore --rm -it --restart=Never \
  --image=bitnami/etcd:3.5.21 \
  --overrides='
{
  "spec": {
    "volumes": [{
      "name": "backup",
      "hostPath": { "path": "/path/to/your/etcd-backup.db-dir" }
    }],
    "containers": [{
      "name": "restore",
      "image": "bitnami/etcd:3.5.21",
      "command": ["etcdctl", "snapshot", "restore", "/backup/etcd-backup.db", "--data-dir=/restored"]
    }]
  }
}'

