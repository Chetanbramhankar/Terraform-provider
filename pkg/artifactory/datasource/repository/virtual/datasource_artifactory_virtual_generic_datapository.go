package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryVirtualGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      "virtual",
		}, nil
	}

	genericSchema := util.MergeMaps(BaseVirtualRepoSchema, resource_repository.RepoLayoutRefSchema("virtual", packageType))

	return &schema.Resource{
		Schema:      genericSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}

func DataSourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = util.MergeMaps(
		BaseVirtualRepoSchema,
		retrievalCachePeriodSecondsSchema,
		resource_repository.RepoLayoutRefSchema("virtual", packageType),
	)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      retrievalCachePeriodSecondsSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(repoWithRetrivalCachePeriodSecsVirtualSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}
