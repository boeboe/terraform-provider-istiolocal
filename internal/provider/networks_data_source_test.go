package provider

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Struct to store parsed docker network info
type dockerNetwork struct {
	ID      string
	Name    string
	Subnet  string
	Gateway string
}

// Helper function to retrieve cmd docker network information
func getDockerNetworks(t *testing.T) []dockerNetwork {
	var dockerNetworks []dockerNetwork

	// First list all docker network shortened IDs
	lsCmd := exec.Command("docker", "network", "ls", "--filter", "driver=bridge", "-q")
	var lsOut bytes.Buffer
	lsCmd.Stdout = &lsOut
	err := lsCmd.Run()
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
	lsOutput := lsOut.String()
	shortIDs := strings.Split(lsOutput, "\n")

	// Then we fetch de details of the docker networks
	inspectArgs := append([]string{
		"network", "inspect", "--format={{.ID}} {{.Name}} {{range .IPAM.Config}}{{.Subnet}} {{.Gateway}}{{end}}"},
		shortIDs[:len(shortIDs)-1]...,
	)
	inspectCmd := exec.Command("docker", inspectArgs...)
	var inspectOut bytes.Buffer
	inspectCmd.Stdout = &inspectOut
	err = inspectCmd.Run()
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
	inspectOutput := inspectOut.String()
	inspectLines := strings.Split(inspectOutput, "\n")
	for _, line := range inspectLines[:len(inspectLines)-1] {
		lineElements := strings.Split(line, " ")
		dockerNetworks = append(dockerNetworks, dockerNetwork{
			ID:      lineElements[0],
			Name:    lineElements[1],
			Subnet:  lineElements[2],
			Gateway: lineElements[3]},
		)
	}

	return dockerNetworks
}

func TestAccNetworksDataSource(t *testing.T) {
	dockerNetworks := getDockerNetworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "istiolocal_networks" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of networks returned
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.#", strconv.FormatInt(int64(len(dockerNetworks)), 10)),
					// Verify the first network to ensure all attributes are set
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.id", dockerNetworks[0].ID),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.name", dockerNetworks[0].Name),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.subnet", dockerNetworks[0].Subnet),
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "networks.0.gateway", dockerNetworks[0].Gateway),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.istiolocal_networks.test", "id", "placeholder"),
				),
			},
		},
	})
}
