{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      app-version = "0.0.1";
      app = pkgs.buildGoModule {
        pname = "tekton-github-app";
        version = app-version;
        src = ./.;
        vendorHash = "sha256-429PH9/M/NhNtj3a5Udkx+UawidQZzRG1tQzGfh674o=";
        # vendorHash = pkgs.lib.fakeHash;
      };
    in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [go];
        };
        packages.default = app;
        apps.default = {
          type = "app";
          program = "${app}/bin/tekton-github-app";
        };
      }
  );
}
