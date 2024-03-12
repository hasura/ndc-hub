# This is used to create a local development shell with Nix, containing all the
# packages used for development.
#
# To use it, install Nix and run `nix develop`.
#
# You can use it with direnv by creating a file called .envrc.local and adding
# the line, `use flake`.

{
  description = "ndc-hub";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/master";

    crane = {
      url = "github:ipetkov/crane";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
  };

  outputs =
    { self
    , flake-utils
    , nixpkgs
    , crane
    , rust-overlay
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ rust-overlay.overlays.default ];
      };
      rustToolchain = pkgs.rust-bin.fromRustupToolchainFile ./rust-toolchain.toml;
      craneLib = (crane.mkLib pkgs).overrideToolchain rustToolchain;

      buildArgs = with pkgs; {
        pname = "ndc-sdk";

        src = craneLib.cleanCargoSource (craneLib.path ./.);

        strictDeps = true;

        # build-time inputs
        nativeBuildInputs = [
          openssl.dev # required to build Rust crates that can conduct TLS connections
          pkg-config # required to find OpenSSL
        ];

        # runtime inputs
        buildInputs = [
          openssl # required for TLS connections
          protobuf # required by opentelemetry-proto, a dependency of axum-tracing-opentelemetry
        ] ++ lib.optionals hostPlatform.isDarwin [
          # macOS-specific dependencies
          libiconv
          darwin.apple_sdk.frameworks.CoreFoundation
          darwin.apple_sdk.frameworks.Security
          darwin.apple_sdk.frameworks.SystemConfiguration
        ];
      };
    in
    {
      packages = {
        deps = craneLib.buildDepsOnly buildArgs;
        default = craneLib.buildPackage
          (buildArgs // {
            cargoArtifacts = self.packages.${system}.deps;
            doCheck = false;
          });
      };

      apps = {
        example = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
          exePath = "/bin/ndc_hub_example";
        };
      };

      devShells.default = with pkgs; mkShell {
        inputsFrom = [ self.packages.${system}.default ];

        nativeBuildInputs = [
          rustToolchain
          cargo-edit
          cargo-machete
          cargo-nextest
          cargo-watch

          just
        ];
      };

      formatter = pkgs.nixpkgs-fmt;
    });
}
