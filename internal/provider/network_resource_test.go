package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "istiolocal_network" "test" {
	name = "testcrud"
	subnet = "192.168.202.0/24"
	gateway = "192.168.202.1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify network
					resource.TestCheckResourceAttr("istiolocal_network.test", "name", "testcrud"),
					resource.TestCheckResourceAttr("istiolocal_network.test", "subnet", "192.168.202.0/24"),
					resource.TestCheckResourceAttr("istiolocal_network.test", "gateway", "192.168.202.1"),
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("istiolocal_network.test", "id"),
					resource.TestCheckResourceAttrSet("istiolocal_network.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "istiolocal_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "istiolocal_network" "test" {
	name = "testcrudbis"
	subnet = "192.168.203.0/24"
	gateway = "192.168.203.1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify network has been updated
					resource.TestCheckResourceAttr("istiolocal_network.test", "name", "testcrudbis"),
					resource.TestCheckResourceAttr("istiolocal_network.test", "subnet", "192.168.203.0/24"),
					resource.TestCheckResourceAttr("istiolocal_network.test", "gateway", "192.168.203.1"),
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("istiolocal_network.test", "id"),
					resource.TestCheckResourceAttrSet("istiolocal_network.test", "created"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
