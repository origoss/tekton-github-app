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
          src = ./.;
          vendorHash = "sha256-429PH9/M/NhNtj3a5Udkx+UawidQZzRG1tQzGfh674o=";
          # vendorHash = pkgs.lib.fakeHash;
        };
        tekton-github-app-app = {
          type = "app";
          program = "${tekton-github-app}/bin/tekton-github-app";
        };
      in {
        devShells.default = pkgs.mkShell { packages = with pkgs; [ go ]; };
        packages.default = tekton-github-app;
        packages.tekton-github-app = tekton-github-app;
        apps.default = tekton-github-app-app;
        apps.tekton-github-app = tekton-github-app-app;
      });
}
