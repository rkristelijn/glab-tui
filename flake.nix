{
  description = "glab-tui - A k9s-inspired TUI for GitLab";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            delve
            git
          ];

          shellHook = ''
            echo "🚀 glab-tui development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go run .           - Run the TUI"
            echo "  go build           - Build binary"
            echo "  go mod tidy        - Clean up dependencies"
            echo "  go test ./...      - Run tests"
            echo ""
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "glab-tui";
          version = "0.1.0";
          src = ./.;
          
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Will need to update this
          
          meta = with pkgs.lib; {
            description = "A k9s-inspired Terminal User Interface for GitLab CI/CD pipelines";
            homepage = "https://github.com/rkristelijn/glab-tui";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
      });
}
