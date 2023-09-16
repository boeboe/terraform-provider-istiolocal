package provider

import (
	"context"
	"fmt"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tfTypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &networksDataSource{}
	_ datasource.DataSourceWithConfigure = &networksDataSource{}
)

// NewNetworksDataSource is a helper function to simplify the provider implementation.
func NewNetworksDataSource() datasource.DataSource {
	return &networksDataSource{}
}

// networksDataSource is the data source implementation.
type networksDataSource struct {
	dockerClient *client.Client
}

// Metadata returns the data source type name.
func (d *networksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

// Configure adds the provider configured client to the data source.
func (d *networksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Ping the Docker daemon
	_, err = dockerClient.Ping(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to connect to Docker daemon",
			err.Error(),
		)
		tflog.Error(ctx, "Failed to connect to Docker daemon", map[string]any{"success": false})
		return
	}

	tflog.Info(ctx, "Configured docker client", map[string]any{"success": true})
	d.dockerClient = dockerClient
}

// Schema defines the schema for the data source.
func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"subnet": schema.StringAttribute{
							Computed: true,
						},
						"gateway": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// networksDataSourceModel maps the data source schema data.
type networksDataSourceModel struct {
	Networks []networksModel `tfsdk:"networks"`
}

// networksModel maps networks schema data.
type networksModel struct {
	ID      tfTypes.String `tfsdk:"id"`
	Name    tfTypes.String `tfsdk:"name"`
	Subnet  tfTypes.String `tfsdk:"subnet"`
	Gateway tfTypes.String `tfsdk:"gateway"`
}

// Read refreshes the Terraform state with the latest data.
func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networksDataSourceModel

	filterArgs := filters.NewArgs()
	filterArgs.Add("driver", "bridge")

	networks, err := d.dockerClient.NetworkList(ctx, dockerTypes.NetworkListOptions{Filters: filterArgs})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Docker Networks",
			err.Error(),
		)
		tflog.Error(ctx, "Unable to Read Docker Networks", map[string]any{"success": false})
		return
	}

	// Map response body to model
	for _, network := range networks {
		networkState := networksModel{
			ID:      tfTypes.StringValue(network.ID),
			Name:    tfTypes.StringValue(network.Name),
			Subnet:  tfTypes.StringValue(network.IPAM.Config[0].Subnet),
			Gateway: tfTypes.StringValue(network.IPAM.Config[0].Gateway),
		}

		state.Networks = append(state.Networks, networkState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
