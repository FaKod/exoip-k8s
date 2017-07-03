# exoip: heartbeat monitor for Exoscale Elastic IP Addresses

By using Simple Kubernetes Leader Election

See <https://github.com/kubernetes/contrib/tree/master/election>

can be use with a daemon set

``` yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: leader-elector
spec:
  selector:
    matchLabels:
      run: leader-elector
  template:
    metadata:
      labels:
        run: leader-elector
    spec:
      nodeSelector:
        role: worker
      hostNetwork: true
      containers:
        - image: innoq/exoip-k8s:latest
          name: leader-elector
          env:
          - name: IF_ADDRESS
            value: "123.123.241.69"
          - name: IF_EXOSCALE_PEER_GROUP
            value: "my-security-group"
          - name: IF_EXOSCALE_API_KEY
            value: "EXO..."
          - name: IF_EXOSCALE_API_SECRET
            value: "blaH_bluP"
          - name: EP_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
```

# history

## 0.0.1

- Initial Version

## 0.0.2

- Only add new Nic if it isn't already there (<https://github.com/FaKod/exoip-k8s/pull/1>)

## 0.0.3

- allow to set Endpoint Namespace by ENV
