{
  description = "Go Dev Shell";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      encrypt-env = pkgs.writeShellApplication {
        name = "encrypt-env";
        runtimeInputs = [ pkgs.sops ];
        text = ''
          sops encrypt .env > encrypt.env
        '';
      };

      decrypt-env = pkgs.writeShellApplication {
        name = "decrypt-env";
        runtimeInputs = [ pkgs.sops ];
        text = ''
          sops decrypt encrypt.env > .env
        '';
      };
    in {
      devShell = pkgs.mkShell {
        buildInputs = with pkgs; [
          just
          opencode

          go
          air

          golangci-lint
          gotools
          gopls

          templ
          nodejs

          sqlc
          goose

          sops
          age
        ] ++ [ encrypt-env decrypt-env ];
      };
    });
}
