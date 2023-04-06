package federated

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/util"
)

const rclass = "federated"
const RepositoriesEndpoint = "artifactory/api/repositories/{key}"

var PackageTypesLikeGeneric = []string{
	"bower",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"helm",
	"npm",
	"opkg",
	"puppet",
	"pypi",
	"swift",
	"vagrant",
}

type Member struct {
	Url     string `hcl:"url" json:"url"`
	Enabled bool   `hcl:"enabled" json:"enabled"`
}

var MemberSchemaGenerator = func(isRequired bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cleanup_on_delete": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Delete all federated members on `terraform destroy` if set to `true`. Caution: it will delete all the repositories in the federation on other Artifactory instances.",
		},
		"member": {
			Type:     schema.TypeSet,
			Required: isRequired,
			Optional: !isRequired,
			Description: "The list of Federated members. If a Federated member receives a request that does not include the repository URL, it will " +
				"automatically be added with the combination of the configured base URL and `key` field value. " +
				"Note that each of the federated members will need to have a base URL set. Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)" +
				" to set up Federated repositories correctly.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "Full URL to ending with the repositoryName",
						ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
					},
					"enabled": {
						Type:     schema.TypeBool,
						Required: true,
						Description: "Represents the active state of the federated member. It is supported to " +
							"change the enabled status of my own member. The config will be updated on the other " +
							"federated members automatically.",
					},
				},
			},
		},
	}
}

var memberSchema = MemberSchemaGenerator(true)

func unpackMembers(data *schema.ResourceData) []Member {
	d := &util.ResourceData{ResourceData: data}
	var members []Member

	if v, ok := d.GetOkExists("member"); ok {
		federatedMembers := v.(*schema.Set).List()
		if len(federatedMembers) == 0 {
			return members
		}

		for _, federatedMember := range federatedMembers {
			id := federatedMember.(map[string]interface{})

			member := Member{
				Url:     id["url"].(string),
				Enabled: id["enabled"].(bool),
			}
			members = append(members, member)
		}
	}
	return members
}

func PackMembers(members []Member, d *schema.ResourceData) error {
	setValue := util.MkLens(d)

	var federatedMembers []interface{}

	for _, member := range members {
		federatedMember := map[string]interface{}{
			"url":     member.Url,
			"enabled": member.Enabled,
		}

		federatedMembers = append(federatedMembers, federatedMember)
	}

	errors := setValue("member", federatedMembers)
	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed saving members to state %q", errors)
	}

	return nil
}
func deleteRepo(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// For federated repositories we delete all the federated members, if the flag `cleanup_on_delete` is set to `true`
	s := &util.ResourceData{ResourceData: d}
	if s.GetBool("cleanup_on_delete", false) {
		var membersUrl []string
		if v, ok := d.GetOk("member"); ok {
			federatedMembers := v.(*schema.Set).List()
			if len(federatedMembers) == 0 {
				return nil
			}
			for _, federatedMember := range federatedMembers {
				id := federatedMember.(map[string]interface{})
				memberUrl := id["url"].(string)
				membersUrl = append(membersUrl, memberUrl)
			}
			for _, memberUrl := range membersUrl {
				repoName := memberUrl[strings.LastIndex(memberUrl, "/")+1:]
				idx := strings.LastIndex(memberUrl, "/artifactory")
				if idx != -1 {
					memberUrl = memberUrl[:idx]
				}
				resp, err := m.(util.ProvderMetadata).Client.SetBaseURL(memberUrl).R().
					AddRetryCondition(client.RetryOnMergeError).
					SetPathParam("key", repoName).
					Delete(RepositoriesEndpoint)
				if err != nil && (resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound)) {
					return diag.FromErr(err)
				}
			}
		}
	}

	resp, err := m.(util.ProvderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetPathParam("key", d.Id()).
		Delete(RepositoriesEndpoint)

	if err != nil && (resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound)) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

func mkResourceSchema(skeema map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor repository.Constructor) *schema.Resource {
	var reader = repository.MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: repository.MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: repository.MkRepoUpdate(unpack, reader),
		DeleteContext: deleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        skeema,
		SchemaVersion: 2,
		CustomizeDiff: repository.ProjectEnvironmentsDiff,
	}
}
