# Artifactory Remote Repository Resource

Provides an Artifactory remote `docker` repository resource. This provides docker specific fields and is the only way to
get them


## Example Usage
Includes only new and relevant fields
```hcl
# Create a new Artifactory remote docker repository called my-remote-docker
resource "artifactory_remote_docker_repository" "my-remote-docker" {
  key                            = "my-remote-docker"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/hub.docker.io/**", "**/bintray.jfrog.io/**"]
  enable_token_authentication    = true
  url                            = "https://hub.docker.io/"
  block_pushing_schema1          = true
}
```

## Argument Reference

Arguments have a one to one mapping with
the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are
supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `block_pushing_schema1` - (Optional) When set, Artifactory will block the pulling of Docker images with manifest v2
  schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1
  that exist in the cache.
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
  * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
  * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
  * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
  * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'