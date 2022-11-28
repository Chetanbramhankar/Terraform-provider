package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type CocoapodsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl"`
}

func ResourceArtifactoryRemoteCocoapodsRepository() *schema.Resource {
	const packageType = "cocoapods"

	var cocoapodsRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, VcsRemoteRepoSchema, map[string]*schema.Schema{
		"pods_specs_repo_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://github.com/CocoaPods/Specs",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote CocoaPods Specs repositories. Default value is "https://github.com/CocoaPods/Specs".`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackCocoapodsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			PodsSpecsRepoUrl:           d.GetString("pods_specs_repo_url", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(cocoapodsRemoteSchema, packer.Default(cocoapodsRemoteSchema), unpackCocoapodsRemoteRepo, func() interface{} {
		repoLayout, _ := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		return &CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
