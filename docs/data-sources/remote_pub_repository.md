---
subcategory: "Remote Repositories"
---
# Artifactory Remote Pub Repository Data Source

Retrieves a remote Pub repository.

## Example Usage

```hcl
data "artifactory_remote_pub_repository" "my-remote-pub" {
  key = "my-remote-pub"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):