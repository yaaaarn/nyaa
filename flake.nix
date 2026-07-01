{
  description = "nyaa";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        chafa = pkgs.chafa.overrideAttrs (finalAttrs: {
          buildInputs = finalAttrs.buildInputs ++ [ pkgs.libwebp ];
        });
      in
      {
        packages.default = pkgs.callPackage ./package.nix { inherit chafa; };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gcc
            cobra-cli
            lowdown
            xdg-utils
            chafa
          ];
        };
      }
    );
}
