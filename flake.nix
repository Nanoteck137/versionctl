{
  description = "versionctl";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        version = pkgs.lib.strings.fileContents "${self}/version";
        fullVersion = ''${version}-${self.dirtyShortRev or self.shortRev or "dirty"}'';

        app = pkgs.buildGoModule {
          pname = "versionctl";
          version = fullVersion;
          src = ./.;

          # ldflags = [
          #   "-X github.com/nanoteck137/tunebook.Version=${version}"
          #   "-X github.com/nanoteck137/tunebook.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          # ];

          vendorHash = null;
        };
      in
      {
        packages = {
          default = app;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            air
            go
            gopls
            just
          ];
        };
      }
    );
}
