{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        tekton-github-app = pkgs.buildGoModule {
          pname = "tekton-github-app";
          version = "0.0.1";
          src = pkgs.lib.sourceFilesBySuffices ./. [ ".go" ".mod" ];
          subPackages = [ "cmd/tekton-github-app" "cmd/ghapp-client" ];
          vendorHash = "sha256-429PH9/M/NhNtj3a5Udkx+UawidQZzRG1tQzGfh674o=";
          # vendorHash = pkgs.lib.fakeHash;
        };
        tekton-github-app-app = {
          type = "app";
          program = "${tekton-github-app}/bin/tekton-github-app";
        };
        tekton-github-client-app = {
          type = "app";
          program = "${tekton-github-app}/bin/ghapp-client";
        };
        client-image = pkgs.dockerTools.buildLayeredImage {
          name = "ghapp-client";
          tag = "latest";
          contents = [ pkgs.cacert tekton-github-app ];
          config = {
            EntryPoint = [ "${tekton-github-app}/bin/ghapp-client"];
          };
        };
        upload-client-image = pkgs.writeShellApplication {
          name = "upload-client-image";
          runtimeInputs = [ pkgs.skopeo ];
          text = ''
            echo "Uploading client image to ghcr"
            skopeo login --username skopeo --password "$GHCR_PAT" ghcr.io
            skopeo copy docker-archive:${client-image} docker://ghcr.io/origoss/tekton-github-app:latest --insecure-policy
          '';
        };
      in {
        devShells.default = pkgs.mkShell { packages = [ pkgs.go upload-client-image ]; };
        packages.default = tekton-github-app;
        packages.tekton-github-app = tekton-github-app;
        packages.client-image = client-image;
        apps.default = tekton-github-app-app;
        apps.tekton-github-app = tekton-github-app-app;
        apps.client = tekton-github-client-app;
      });
}
