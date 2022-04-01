# operator-elk-extra
This operator provide features that not yet exist on ECK

## Features

All operator features can work with cluster managed by ECK but also with all other cluster

If your cluster is managed by ECK, you just need to specify this on `spec` resources:
```yaml
spec:
  elasticsearchRef:
    name: cluster-sample
```

else, you need to specify this:
```yaml
spec:
  elasticsearchRef:
    addresses:
      - https://elasticsearch.domain.com
    secretName: elasticsearch-credentials
```

The secret must be contain `user` and `password` key like this:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: elasticsearchcredentials
  namespace: elk
type: Opaque
data:
  user: YOUR_USER_BASE64
  password: YOUR_PASSWORD_BASE64
```


### License

This feature not working when cluster is deployed by ECK, because off is already managed by it.
There are no way to disable license reconciler on ECK


1. You need to create `Secret` with license key on Elasticsearch namespace.
2. You need to create `License` that reference secret previously created on same namespace.

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
```


## ILM policy

You can manage ILM policy, with ILM resource like this:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchILM
metadata:
  name: policy-log
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  policy: |
    {
        "policy": {
            "phases": {
                "hot": {
                    "min_age": "0ms",
                    "actions": {
                        "rollover": {
                            "max_size": "5gb",
                            "max_age": "7d"
                        },
                        "set_priority" : {
                            "priority": 100
                        }
                    }
                },
                "warm": {
                    "min_age": "0ms",
                    "actions": {
                        "forcemerge": {
                            "max_num_segments": 1
                        },
                        "shrink": {
                            "number_of_shards": 1
                        },
                        "set_priority" : {
                            "priority": 50
                        },
                        "readonly": {}
                    }
                },
                "delete": {
                    "min_age": "0d",
                    "actions": {
                        "delete": {}
                    }
                }
            }
        }
    }
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