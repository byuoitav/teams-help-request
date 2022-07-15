terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-586877430255"
    dynamodb_table = "terraform-state-lock-586877430255"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "teams-help.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-cluster-endpoint"
}

provider "kubernetes" {
  host        = data.aws_ssm_parameter.eks_cluster_endpoint.value
  config_path = "~/.kube/config"
}

data "aws_ssm_parameter" "hub_address" {
  name = "/env/hub-address"
}

data "aws_ssm_parameter" "web_hook" {
  name = "/env/teams/web-hook-url"
}

data "aws_ssm_parameter" "av_monitoring" {
  name = "/env/teams/monitoring-url"
}

module "teams-help-request" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "teams-help"
  image          = "docker.pkg.github.com/byuoitav/teams-help-request/teams-help-request"
  image_version  = "v0.0.3"
  container_port = 8080 // doesn't have a port but this is a required field
  repo_url       = "https://github.com/byuoitav/teams-help-request"

  // optional
  image_pull_secret = "github-docker-registry"
  container_env = {
    "GIN_MODE" = "release"
  }
  container_args = [
    "--log-level", "info",
    "--hub-address", "ws://${data.aws_ssm_parameter.hub_address.value}",
    "--webhook-url", data.aws_ssm_parameter.web_hook.value,
    "--monitoring-url", data.aws_ssm_parameter.av_monitoring.value
  ]
  health_check = false
}