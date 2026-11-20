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
    # AWS Copilot は brew formula copilot (link しない)。GitHub の copilot と名前が衝突するため別名で出す
    (writeShellScriptBin "aws-copilot" ''
      exec /opt/homebrew/opt/copilot/bin/copilot "$@"
    '')
  ];
}
