package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryFederatedGenericRepository(repoType string) *schema.Resource {
	localRepoSchema := local.GetGenericRepoSchema(repoType)

	var federatedSchema = util.MergeMaps(localRepoSchema, memberSchema, repository.RepoLayoutRefSchema(rclass, repoType))

	type FederatedRepositoryParams struct {
		local.RepositoryBaseParams
		Members []Member `hcl:"member" json:"members"`
	}

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := FederatedRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(rclass, data, repoType),
			Members:              unpackMembers(data),
		}
		return repo, repo.Id(), nil
	}

	var packGenericMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*FederatedRepositoryParams).Members
		return packMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packGenericMembers,
	)

	constructor := func() (interface{}, error) {
		return &FederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(repoType),
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(federatedSchema, pkr, unpackFederatedRepository, constructor)
}
