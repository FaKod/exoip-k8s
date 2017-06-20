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
```
