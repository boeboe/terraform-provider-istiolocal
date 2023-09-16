package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworksDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "istiolocal_networks" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of networks returned
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.#", "2"),
					// Verify the first network to ensure all attributes are set
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.id", "4a1d9bcdc58e52d75e10636093be6f6d4abf0956a2ad8d7e04cd7587a14b782d"),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.name", "bridge"),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.subnet", "172.17.0.0/16"),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.gateway", "172.17.0.1"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "id", "placeholder"),
				),
			},
		},
	})
}
