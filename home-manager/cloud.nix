{ pkgs, ... }:
{
  home.packages = with pkgs; [
    awscli2
    aws-sam-cli
    ssm-session-manager-plugin
    google-cloud-sdk
    terraform
    kubectl
    kubernetes-helm
    kubeseal
    kind
    s3cmd
  ];
}
