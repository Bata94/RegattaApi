{
  description = "Go Dev Shell";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            just
            go
            air
            golangci-lint

            sqlc
            goose
          ];
          shellHook = ''
            echo "Go version:"
            go version
            go mod tidy
          '';
        };
      });
}

