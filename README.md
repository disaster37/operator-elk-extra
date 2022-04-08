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
There are no way to disable license reconciler on ECK.

To get more info about license manegement, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/licensing-apis.html)


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

#### Paramateres

- **secretName** (string): the secret that contain license as JSON string on key `license`


#### Extra configuration

If you use ECK to deploy ECK and you should to use gold or platinium license, you need to add this on your setting:

__elasticsearch.yaml__:
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

__kibana.yaml__:
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

> You need also use ECK operator fork, because of the default operator only work with basic or enterprise license.

### ILM policy

You can manage ILM policy. ILM allow you to set life cycle of your index.

To get more info about ILM, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-lifecycle-management-api.html)


__Sample__:
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

#### Paramaters

- **policy** (JSON string): The ILM policy as JSON string


### SLM policy

You can manage SLM policy. SLM allow you to set life cycle of your snapshot.

To get more info about SLM, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management-api.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchSLM
metadata:
  name: policy-log
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  name: '<daily-snap-{now/d}>'
  schedule: '0 30 1 * * ?'
  repository: 'my_repository'
  config:
    expand_wildcards: 'all'
    ignore_unavailable: false
    include_global_state: false
    indices:
      - 'data-*'
      - 'important'
    partial: false
  retention:
    expire_after: '30d'
    max_count: 10
    min_count: 5
```

#### Paramaters

- **name** (string): Name automatically assigned to each snapshot created by the policy
- **schedule** (string): Periodic or absolute schedule at which the policy creates snapshots
- **repository** (string): Repository used to store snapshots created by this policy
- **config** (object): Configuration for each snapshot created by the policy
- **retention** (object): Retention rules used to retain and delete snapshots created by the policy

**Config object**:
- **expand_wildcards** (string): Determines how wildcard patterns in the indices parameter match data streams and indices
- **ignore_unavailable** (bool):  If false, the snapshot fails if any data stream or index in indices is missing or closed
- **include_global_state** (bool): If true, include the cluster state in the snapshot
- **indices** (list of string): list of data streams and indices to include in the snapshot
- **partial** (bool): If false, the entire snapshot will fail if one or more indices included in the snapshot do not have all primary shards available
- **feature_states** (string): Feature states to include in the snapshot
- **metadata** (JSON string): attaches arbitrary metadata to the snapshot, such as a record of who took the snapshot, why it was taken, or any other useful data

***Retention object***:
- **expire_after**: Time period after which a snapshot is considered expired and eligible for deletion
- **max_count**: Maximum number of snapshots to retain, even if the snapshots have not yet expired
- **min_count**: Minimum number of snapshots to retain, even if the snapshots have expired

### Snapshot repository

It permit to manage snapshot repository

To get more info about snapshot repository, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshots-register-repository.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchSnapshotRepository
metadata:
  name: backup
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  type: 'fs'
  settings: |
    {
        "location": "/backup"
    }
```

#### Paramaters

- **type** (string): It's a repository type
- **settings** (JSON string): It's config for repository


### User

This resource permit to manage internal user in Elasticsearch.

To get more info about user, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html)

__Secret sample__:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: users
  namespace: elk
type: Opaque
data:
  admin: YOUR_PASSWORD_CONTEND_BASE64
```

__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: User
metadata:
  name: admin
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  enabled: true
  secret:
    name: users
    key: admin
  roles:
    - 'superuser'
```

#### Paramaters

- **enabled** (bool): Specifies whether the user is enabled
- **email** (string): The email of the user
- **full_name** (string): The full name of the user
- **metadata** (string): Arbitrary metadata that you want to associate with the user
- **secret** (object): Secret that store user password
- **password_hash** (string): A hash of the user’s password. It must be generated with bcrypt
- **roles** (list of string): A set of roles the user has

**Secret object**:
- **name** (string): the secret name
- **key** (string): the secret key that store effective password


### Role

This resource permit to manage role in Elasticsearch.

To get more info about role, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchRole
metadata:
  name: workers
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  cluster:
    - 'all'
  indices:
    - names:
        - 'index1'
        - 'index2'
      privileges:
        - 'all'
      field_security: |
        {
            "grant" : [ "title", "body" ]
        }
      query: '{"match": {"title": "foo"}}'
  applications:
    - application: 'myapp'
      privileges:
        - 'admin'
        - 'read'
      resources: 
        - '*'
  run_as:
    - 'other_user'
```

#### Paramaters

- **cluster** (list of string): A list of cluster privileges
- **indices** (list of object): A list of indices permissions entries
- **applications** (list of object): A list of application privilege entries
- **run_as** (list of string): A list of users that the owners of this role can impersonate
- **global** (JSON string): An object defining global privileges
- **metadata** (JSON string): Optional meta-data
- **transient_metadata** (JSON string):

**Indice object**:
- **names** (list of string): A list of indices (or index name patterns) to which the permissions in this entry apply
- **privileges** (list of strings): The index level privileges that the owners of the role have on the specified indices
- **field_security** (JSON string): The document fields that the owners of the role have read access to
- **query** (string): A search query that defines the documents the owners of the role have read access to

**Application object**:
- **application** (string): The name of the application to which this entry applies
- **privileges** (list of string): A list of strings, where each element is the name of an application privilege or action
- **resources** (list of string): A list resources to which the privileges are applied


### Role mapping

This resource permit to manage role mapping in Elasticsearch.

To get more info about role, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchRole
metadata:
  name: workers
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  enabled: true
  roles:
    - 'user'
  rules: |
    {
    "field" : { "username" : "*" }
    }
```

#### Paramaters

- **enabled** (bool): Mappings that have enabled set to false are ignored when role mapping is performed
- **roles** (list of string): A list of role names that are granted to the users that match the role mapping rules
- **rules** (JSON string): The rules that determine which users should be matched by the mapping
- **metadata** (JSON string): Additional metadata that helps define which roles are assigned to each user


### Component template

This resource permit to manage component template in Elasticsearch.

To get more info about component template, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchComponentTemplate
metadata:
  name: my-company
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  settings: |
    {
      "number_of_shards": 1
    }
  mappings: |
    {
      "_source": {
        "enabled": false
      },
      "properties": {
        "host_name": {
          "type": "keyword"
        },
        "created_at": {
          "type": "date",
          "format": "EEE MMM dd HH:mm:ss Z yyyy"
        }
      }
    }
```

#### Paramaters

- **settings** (JSON string): Configuration options for the index
- **mappings** (JSON string): Mapping for fields in the index
- **aliases** (JSON string): Aliases to add


### Index template

This resource permit to manage index template in Elasticsearch.

To get more info about index template, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchIndexTemplate
metadata:
  name: my-template
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  index_patterns:
    - 'te*'
    - 'bar*'
  composed_of:
    - 'component_template1'
    - 'runtime_component_template'    
  priority: 500
  version: 3
  template:
    settings: |
        {
            "number_of_shards": 1
        }        
    mappings: |
        {
            "_source": {
                "enabled": true
            },
            "properties": {
                "host_name": {
                    "type": "keyword"
                },
                "created_at": {
                    "type": "date",
                    "format": "EEE MMM dd HH:mm:ss Z yyyy"
                }
            }
        }
```

#### Paramaters

- **index_patterns** (list of string): list of index pattern to apply template
- **composed_of** (list of string): list of component template
- **priority** (number): the priority to apply template
- **version** (number): 
- **template** (object): the template specification
- **_meta** (JSON string): optional meta data
- **allow_auto_create** (bool):

**Template object**
- **settings** (JSON string): Configuration options for the index
- **mappings** (JSON string): Mapping for fields in the index
- **aliases** (JSON string): Aliases to add


### Watcher

This resource permit to manage watcher in Elasticsearch.

> This feature is not include on basic license

To get more info about index template, read the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html)


__Sample__:
```yaml
apiVersion: elk.k8s.webcenter.fr/v1alpha1
kind: ElasticsearchWatcher
metadata:
  name: my-watcher
  namespace: elk
spec:
  elasticsearchRef:
    name: cluster-sample
  trigger: |
    {
        "schedule" : { "cron" : "0 0/1 * * * ?" }
    }
  input: |
    {
    "search" : {
      "request" : {
        "indices" : [
          "logstash*"
        ],
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "response": 404
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-5m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
  condition: |
    {
        "compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
    }
  actions: |
    {
        "email_admin" : {
            "email" : {
                "to" : "admin@domain.host.com",
                "subject" : "404 recently encountered"
            }
        }
    }
```

#### Paramaters

- **trigger** (JSON string): The trigger that defines when the watch should run
- **input** (JSON string): The input that defines the input that loads the data for the watch
- **condition** (JSON string): The condition that defines if the actions should be run
- **actions** (JSON string): The list of actions that will be run if the condition matches
- **transform** (JSON string): 
- **throttle_period** (string): The minimum time between actions being run, the default for this is 5 seconds
- **throttle_period_in_millis** (number):  Minimum time in milliseconds between actions being run
- **metadata** (JSON string): Metadata json that will be copied into the history entries
