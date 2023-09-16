terraform {
  required_providers {
    istiolocal = {
      source = "hashicorp.com/tetrate/istiolocal"
    }
  }
}

provider "istiolocal" {}

resource "istiolocal_network" "network_one" {
  name = "istio_one"
  subnet = "192.168.200.0/24"
  gateway = "192.168.200.1"
}

resource "istiolocal_network" "network_two" {
  name = "istio_two"
  subnet = "192.168.201.0/24"
  gateway = "192.168.201.1"
}

# resource "istiolocal_network" "network_import" {
# }

output "docker_networks" {
  value = [
    istiolocal_network.network_one,
    istiolocal_network.network_two,
    # istiolocal_network.network_import
  ]
}
