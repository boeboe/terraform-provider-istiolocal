package provider

import (
	"context"
	"fmt"
	"log"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tfTypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithConfigure   = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
)

// NewNetworkResource is a helper function to simplify the provider implementation.
func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

// networkResource is the resource implementation.
type networkResource struct {
	dockerClient *client.Client
}

// Metadata returns the resource type name.
func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

// Configure adds the provider configured client to the resource.
func (r *networkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.dockerClient = dockerClient
}

// Schema defines the schema for the resource.
func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"subnet": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"gateway": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
		},
	}
}

// networkResourceModel maps the resource schema data.
type networkResourceModel struct {
	ID      tfTypes.String `tfsdk:"id"`
	Created tfTypes.String `tfsdk:"created"`
	Name    tfTypes.String `tfsdk:"name"`
	Subnet  tfTypes.String `tfsdk:"subnet"`
	Gateway tfTypes.String `tfsdk:"gateway"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan networkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipamConfig := &network.IPAM{
		Config: []network.IPAMConfig{
			{
				Subnet:  plan.Subnet.ValueString(),
				Gateway: plan.Gateway.ValueString(),
			},
		},
		Driver: "default",
	}

	// Network creation options
	createOptions := dockerTypes.NetworkCreate{
		Scope:      "local",
		Driver:     "bridge",
		IPAM:       ipamConfig,
		EnableIPv6: false,
		Internal:   false,
		Attachable: false,
		Ingress:    false,
		Labels:     map[string]string{"purpose": "istiolocal"},
		Options:    map[string]string{"com.docker.network.driver.mtu": "1500"},
	}

	// Create network
	network, err := r.dockerClient.NetworkCreate(ctx, plan.Name.ValueString(), createOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Docker Network",
			err.Error(),
		)
		tflog.Error(ctx, "Unable to Create Docker Network", map[string]any{"success": false})
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = tfTypes.StringValue(network.ID)
	plan.Created = tfTypes.StringValue(time.Now().Format(time.RFC3339Nano))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state networkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed network value from network
	networkDetails, err := r.dockerClient.NetworkInspect(ctx, state.ID.ValueString(), dockerTypes.NetworkInspectOptions{})
	if err != nil {
		log.Fatalf("Failed to inspect Docker network: %v", err)
		resp.Diagnostics.AddError(
			"Failed to inspect Docker network",
			"Could not read Docker network ID "+state.ID.ValueString()+": "+err.Error(),
		)
		tflog.Error(ctx, "Failed to inspect Docker network", map[string]any{"success": false})
		return
	}

	// Overwrite items with refreshed state
	state.ID = tfTypes.StringValue(networkDetails.ID)
	state.Created = tfTypes.StringValue(networkDetails.Created.Format(time.RFC3339Nano))
	state.Name = tfTypes.StringValue(networkDetails.Name)
	state.Subnet = tfTypes.StringValue(networkDetails.IPAM.Config[0].Subnet)
	state.Gateway = tfTypes.StringValue(networkDetails.IPAM.Config[0].Gateway)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan networkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Docker does not allow to modify existing networks. We need pave and nuke instead.
	tflog.Warn(ctx, "Not implementend... updating Docker networks is not supported", map[string]any{"success": false})
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state networkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.dockerClient.NetworkRemove(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Docker network",
			"Could not delete Docker network, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
