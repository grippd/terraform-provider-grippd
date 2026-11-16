package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Run with TF_ACC=1. TF_ACC_TERRAFORM_PATH points at `tofu` to test against OpenTofu.
var factories = map[string]func() (tfprotov6.ProviderServer, error){
	"grippd": providerserver.NewProtocol6WithError(New("test")()),
}

const generated = `
resource "grippd_env" "jwt" {
  project  = "tns-review"
  env      = "staging"
  name     = "GRIPPD_JWT"
  apps     = ["api"]
  generate = "32-bytes-base64url"
}
`

const asked = `
resource "grippd_env" "sentry" {
  project = "saas"
  env     = "production"
  name    = "SENTRY_DSN"
  apps    = ["api", "web"]
  ask = {
    question = "Sentry DSN for errors"
    help     = "Sentry: Settings → Projects → Client Keys (DSN)"
    required = false
  }
}
`

func TestEnvGenerate(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: generated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grippd_env.jwt", "id", "tns-review/staging/GRIPPD_JWT"),
					resource.TestCheckResourceAttr("grippd_env.jwt", "generate", "32-bytes-base64url"),
					resource.TestCheckNoResourceAttr("grippd_env.jwt", "ask"),
				),
			},
			// Giving the value to another app is an in-place update, not a new value.
			{
				Config: regexp.MustCompile(`\["api"\]`).ReplaceAllString(generated, `["api", "worker"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("grippd_env.jwt", plancheck.ResourceActionUpdate)},
				},
				Check: resource.TestCheckResourceAttr("grippd_env.jwt", "apps.#", "2"),
			},
			// A different kind of generated value is a new value.
			{
				Config: regexp.MustCompile(`32-bytes-base64url`).ReplaceAllString(generated, `uuid`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("grippd_env.jwt", plancheck.ResourceActionReplace)},
				},
			},
		},
	})
}

func TestEnvAsk(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config: asked,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("grippd_env.sentry", "ask.required", "false"),
				resource.TestCheckResourceAttr("grippd_env.sentry", "apps.#", "2"),
			),
		}},
	})
}

func TestEnvAskDefaultsToRequired(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config: `
resource "grippd_env" "key" {
  project = "saas"
  env     = "staging"
  name    = "STRIPE_KEY"
  apps    = ["api"]
  ask     = {}
}`,
			Check: resource.TestCheckResourceAttr("grippd_env.key", "ask.required", "true"),
		}},
	})
}

func TestEnvFrom(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config: `
resource "grippd_env" "media" {
  project = "review"
  env     = "staging"
  name    = "MEDIA_ENDPOINT"
  apps    = ["api"]
  from    = "files.media.endpoint"
}`,
			Check: resource.TestCheckResourceAttr("grippd_env.media", "from", "files.media.endpoint"),
		}},
	})
}

func TestEnvRejects(t *testing.T) {
	cases := map[string]struct{ body, err string }{
		"two kinds":  {`generate = "uuid"` + "\n" + `from = "database.main.url"`, `Invalid Attribute Combination`},
		"no kind":    {``, `Missing Attribute Configuration`},
		"bad kind":   {`generate = "hex"`, `(?s)value must be one of`},
		"bad from":   {`from = "Database"`, `(?s)dotted grippd reference`},
		"lower name": {`generate = "uuid"`, `(?s)UPPER_SNAKE_CASE`},
		"empty apps": {`generate = "uuid"`, `(?s)at least 1`},
	}
	for label, c := range cases {
		t.Run(label, func(t *testing.T) {
			name, apps := `"KEY"`, `["api"]`
			if label == "lower name" {
				name = `"key"`
			}
			if label == "empty apps" {
				apps = `[]`
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: factories,
				Steps: []resource.TestStep{{
					Config: `
resource "grippd_env" "x" {
  project = "p"
  env     = "dev"
  name    = ` + name + `
  apps    = ` + apps + `
  ` + c.body + `
}`,
					ExpectError: regexp.MustCompile(c.err),
				}},
			})
		})
	}
}
