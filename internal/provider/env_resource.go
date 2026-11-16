package provider

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	envNameRe     = regexp.MustCompile(`^[A-Z_][A-Z0-9_]{0,127}$`)
	projectNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	appNameRe     = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)
	envRe         = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	fromRe        = regexp.MustCompile(`^[a-z][a-z0-9_-]*(\.[a-z0-9][a-z0-9_-]*)+$`)
)

// Generate kinds, as the compiler writes them.
var generateKinds = []string{"32-bytes-base64url", "uuid", "password"}

type envResource struct{}

func newEnvResource() resource.Resource { return &envResource{} }

type envModel struct {
	ID       types.String `tfsdk:"id"`
	Project  types.String `tfsdk:"project"`
	Env      types.String `tfsdk:"env"`
	Name     types.String `tfsdk:"name"`
	Apps     types.List   `tfsdk:"apps"`
	Generate types.String `tfsdk:"generate"`
	Ask      *askModel    `tfsdk:"ask"`
	From     types.String `tfsdk:"from"`
}

type askModel struct {
	Question types.String `tfsdk:"question"`
	Help     types.String `tfsdk:"help"`
	Pattern  types.String `tfsdk:"pattern"`
	Required types.Bool   `tfsdk:"required"`
}

func (r *envResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

// A value is generated, asked for or resolved: exactly one.
func (r *envResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(path.MatchRoot("generate"), path.MatchRoot("ask"), path.MatchRoot("from")),
	}
}

func (r *envResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "One value grippd keeps for an environment: generated once, asked of the team, or " +
			"resolved from something grippd mints. The value itself is never in state or the plan.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "`<project>/<env>/<name>`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project": schema.StringAttribute{
				Required:      true,
				Description:   "The grippd project.",
				Validators:    []validator.String{stringvalidator.RegexMatches(projectNameRe, "must be a grippd project name")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"env": schema.StringAttribute{
				Required:      true,
				Description:   "The environment, such as `dev`, `staging` or `production`.",
				Validators:    []validator.String{stringvalidator.RegexMatches(envRe, "must be an environment name")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "The variable's name, UPPER_SNAKE_CASE.",
				Validators:    []validator.String{stringvalidator.RegexMatches(envNameRe, "names are UPPER_SNAKE_CASE")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"apps": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The apps that get the value.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(appNameRe, "must be an app name")),
				},
			},
			"generate": schema.StringAttribute{
				Optional:      true,
				Description:   "Made once for the environment: `32-bytes-base64url`, `uuid` or `password`. Changing it makes a new value.",
				Validators:    []validator.String{stringvalidator.OneOf(generateKinds...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"from": schema.StringAttribute{
				Optional:    true,
				Description: "Resolved by grippd when the app deploys, such as `database.main.url` or `files.media.access-key`.",
				Validators:  []validator.String{stringvalidator.RegexMatches(fromRe, "must be a dotted grippd reference")},
			},
			"ask": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Given by the team on the deploy form or with `grippd env set`.",
				Attributes: map[string]schema.Attribute{
					"question": schema.StringAttribute{Optional: true, Description: "What the deploy form asks."},
					"help":     schema.StringAttribute{Optional: true, Description: "Where to find the value."},
					"pattern":  schema.StringAttribute{Optional: true, Description: "A pattern the value must match."},
					"required": schema.BoolAttribute{
						Optional: true, Computed: true, Default: booldefault.StaticBool(true),
						Description: "Deploys wait until a required value is set. Defaults to true.",
					},
				},
			},
		},
	}
}

// project, env and name all force a new resource, so the id never changes in place.
func envID(m *envModel) types.String {
	return types.StringValue(strings.Join([]string{m.Project.ValueString(), m.Env.ValueString(), m.Name.ValueString()}, "/"))
}

// The resource is a declaration: grippd reads it from the compiled config and state, and
// creates, rotates and deletes the value itself. So every operation only records the plan.
func (r *envResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	m.ID = envID(&m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *envResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *envResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	m.ID = envID(&m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *envResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {}
