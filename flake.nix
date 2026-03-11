{
  description = "Agent Orchestrator - AI Agent Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Platform-specific dependencies
        linuxDeps = with pkgs; lib.optionals stdenv.isLinux [
          gtk3
          webkitgtk_4_1
          pkg-config
          stdenv.cc.cc.lib
        ];
        
        # Common build inputs
        buildInputs = with pkgs; [
          # Go
          go_1_22
          gopls
          gotools
          
          # Node.js
          nodejs_20
          nodePackages.npm
          
          # Build tools
          gnumake
          git
        ] ++ linuxDeps;
        
      in
      {
        devShells.default = pkgs.mkShell {
          inherit buildInputs;
          
          # Library path for Linux GUI
          LD_LIBRARY_PATH = pkgs.lib.optionalString pkgs.stdenv.isLinux
            "${pkgs.stdenv.cc.cc.lib}/lib";
          
          shellHook = ''
            echo ""
            echo "╔══════════════════════════════════════════╗"
            echo "║     Agent Orchestrator Dev Environment   ║"
            echo "╚══════════════════════════════════════════╝"
            echo ""
            echo "Go:        $(go version | cut -d' ' -f3)"
            echo "Node:      $(node --version)"
            echo "NPM:       $(npm --version)"
            echo ""
            echo "Commands:"
            echo "  make dev      - Start development server"
            echo "  make build    - Build production binary"
            echo "  make test     - Run tests"
            echo "  make lint     - Run linters"
            echo ""
          '';
        };
        
        # For nix build
        packages.default = pkgs.buildGoModule {
          pname = "agent-orch";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # Will be set after first build
          
          inherit buildInputs;
        };
      }
    );
}
