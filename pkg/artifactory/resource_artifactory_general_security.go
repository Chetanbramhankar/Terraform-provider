package artifactory

import (
	"context"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

type GeneralSecurity struct {
	GeneralSettings `yaml:"security" json:"security"`
}

type GeneralSettings struct {
	AnonAccessEnabled bool `yaml:"anonAccessEnabled" json:"anonAccessEnabled"`
}

func resourceArtifactoryGeneralSecurity() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceGeneralSecurityUpdate,
		CreateContext: resourceGeneralSecurityUpdate,
		DeleteContext: resourceGeneralSecurityDelete,
		ReadContext:   resourceGeneralSecurityRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enable_anonymous_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceGeneralSecurityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*resty.Client)

	generalSettings := GeneralSettings{}

	_, err := c.R().SetResult(&generalSettings).Get("artifactory/api/securityconfig")
	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
	}

	s := GeneralSecurity{GeneralSettings: generalSettings}
	packDiag := packGeneralSecurity(&s, d)

	if packDiag != nil {
		return packDiag
	}

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Usage of Undocumented Artifactory API Endpoints",
		Detail:   "The artifactory_general_security resource uses endpoints that are undocumented and may not work with SaaS environments, or may change without notice.",
	}}
}

func resourceGeneralSecurityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpacked := unpackGeneralSecurity(d)
	content, err := yaml.Marshal(&unpacked)

	if err != nil {
		return diag.Errorf("failed to marshal general security settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one general security settings resource, using same id
	d.SetId("security")
	return resourceGeneralSecurityRead(ctx, d, m)
}

func resourceGeneralSecurityDelete(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	var content = `
security:
  anonAccessEnabled: false
`

	err := sendConfigurationPatch([]byte(content), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
	}

	return nil
}

func unpackGeneralSecurity(s *schema.ResourceData) *GeneralSecurity {
	d := &ResourceData{s}
	security := *new(GeneralSecurity)

	settings := GeneralSettings{
		AnonAccessEnabled: d.getBool("enable_anonymous_access", false),
	}

	security.GeneralSettings = settings
	return &security
}

func packGeneralSecurity(s *GeneralSecurity, d *schema.ResourceData) diag.Diagnostics {
	setValue := mkLens(d)

	errors := setValue("enable_anonymous_access", s.GeneralSettings.AnonAccessEnabled)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack general security settings %q", errors)
	}

	return nil
}