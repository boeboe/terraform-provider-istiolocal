terraform {
  required_providers {
    istiolocal = {
      source = "hashicorp.com/tetrate/istiolocal"
    }
  }
}

provider "istiolocal" {}

data "istiolocal_networks" "networks" {}

output "docker_networks" {
  value = data.istiolocal_networks.networks
}