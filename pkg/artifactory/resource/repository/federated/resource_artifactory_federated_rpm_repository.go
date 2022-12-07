package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

type RpmFederatedRepositoryParams struct {
	local.RpmLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
}

func ResourceArtifactoryFederatedRpmRepository() *schema.Resource {
	packageType := "rpm"

	rpmFederatedSchema := util.MergeMaps(
		local.RpmLocalSchema,
		memberSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedRpmRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.UnpackLocalRpmRepository(data, rclass),
			Members:                  unpackMembers(data),
		}
		return repo, repo.Id(), nil
	}

	var packRpmMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*RpmFederatedRepositoryParams).Members
		return packMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packRpmMembers,
	)

	constructor := func() (interface{}, error) {
		return &RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.RpmLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
				RootDepth:               0,
				CalculateYumMetadata:    false,
				EnableFileListsIndexing: false,
				GroupFileNames:          "",
			},
		}, nil
	}

	return repository.MkResourceSchema(rpmFederatedSchema, pkr, unpackFederatedRpmRepository, constructor)
}
