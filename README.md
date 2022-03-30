# operator-elk-extra
This operator provide features that not yet exist on ECK

## Features

### License

The ECK doesn't allow to manage multiple license from same operator. So, you need to deploy one operator per license. Moreoever, it force you to use `enterprise` license.

This operator fix this behaviour. You can now use `gold, platinium or enterprise` license. And you can manage all your license from one operator.

> You can manage license from cluster managed by ECK or others.

1. You need to create `Secret` with license key on Elasticsearch namespace.
2. You need to create `License` that reference secret previously created on same namespace.
3. If Elasticsearch cluster is not managed by ECK, you need to create `Secret` with key `adresses, username, password` and reference it on License object.

__Secret sample__:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: elasticsearch-license
  namespace: elk
type: Opaque
data:
  license: YOUR_LICENSE_CONTEND_BASE64
```

__License sample when Elasticsearch is deployed by ECK__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: License
metadata:
  name: license
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  secretName: elasticsearch-license
```

> The ECK name in this sample is `cluster-sample`

__License sample when Elasticsearch is not deployed by ECK__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: License
metadata:
  name: license
  namespace: elk
spec:
  elasticsearchRef:
    addresses:
      - https://elasticsearch.domain.com
    secretName: elasticsearch-credentials
  secretName: elasticsearch-license

---
piVersion: v1
kind: Secret
metadata:
  name: elasticsearchcredentials
  namespace: elk
type: Opaque
data:
  user: YOUR_USER_BASE64
  password: YOUR_PASSWORD_BASE64
```

If you use ECK to deploy ECK and you should to use gold or platinium license, you need to add this on your setting:

elasticsearch.yaml:
```yaml
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: cluster-sample
  namespace: elk
spec:
  nodeSets:
    - config:
        xpack:
          license:
            upload:
              types:
                - gold
                - platinium
```

> If you should display license management on kibana

kibana.yaml:
```yaml
apiVersion: kibana.k8s.elastic.co/v1
kind: Kibana
metadata:
  name: cluster-sample
  namespace: elk
spec:
  config:
    xpack.license_management.ui.enabled: true
```