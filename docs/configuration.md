# Configuration

Here is an example of configuration you can have.

```yaml
# Namespace where kubernetes-tagger is installed
# namespace: "kube-system"

# Kubernetes configuration file
# kubeconfig: "~/.kube/config"

# Server listener address
# address: :8085

# Log level
# loglevel: info

# Log format
# logformat: json

# Kubernetes provider
# provider: aws

# AWS configuration
aws:
  # Region
  region: eu-central-1

# Rules to add / delete tags
rules:
  # Rule definition add value hardcoded
  - tag: tag-hardcoded
    value: hardcoded-value
    action: add
  # Rule definition add value from query (only a value exists for your query)
  - tag: tag-query
    query: persistentvolume.phase
    action: add
  # Rule definition with condition
  - tag: tag-condition
    query: persistentvolume.name
    action: add
    when:
      - condition: persistentvolume.phase
        value: Bound
        operator: Equal
  # Rule definition delete tag
  - tag: tag-to-be-deleted
    action: delete
```
