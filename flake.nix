# This is used to create a local development shell with Nix, containing all the
# packages used for development.
#
# To use it, install Nix and run `nix develop`.
#
# You can use it with direnv by creating a file called .envrc.local and adding
# the line, `use flake`.

{
  description = "ndc-sdk";

  inputs = {
    flake-utils.url = github:numtide/flake-utils;
    nixpkgs.url = github:NixOS/nixpkgs/master;
  };

  outputs =
    { self
    , flake-utils
    , nixpkgs
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.default = pkgs.mkShell {
        nativeBuildInputs = [
          pkgs.cargo
          pkgs.cargo-edit
          pkgs.cargo-machete
          pkgs.clippy
          pkgs.rust-analyzer
          pkgs.rustPlatform.rustcSrc
          pkgs.rustc
          pkgs.rustfmt
        ];

        buildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.darwin.apple_sdk.frameworks.Security
          pkgs.libiconv
        ]

        ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
          pkgs.pkg-config
          pkgs.openssl
        ];
      };

      formatter = pkgs.nixpkgs-fmt;
    });
}
