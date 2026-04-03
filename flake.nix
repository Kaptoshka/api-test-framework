{
  description = "API Testing Framework";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Reporting
            allure

            # Go toolchain
            go_1_26
            gopls
            delve

            # Linting
            golangci-lint
            gotools # goimports
            yamllint
            statix
            deadnix
            nixpkgs-fmt
          ];

          shellHook = ''
            echo "Go API Autotest Framework"
            echo "- Go version: $(go version)"

            export ENV_FILE="$PWD/.env"
            export ALLURE_RESULTS_DIR="$PWD/artifacts/allure-results"
            export LOG_DIR="$PWD/artifacts/logs"
          '';
        };
      }
    );
}
