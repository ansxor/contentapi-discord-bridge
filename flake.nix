{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    pyproject-nix.url = "github:pyproject-nix/pyproject.nix";
  };

  outputs = { pyproject-nix, flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = (import (nixpkgs) { inherit system; });
        project = pyproject-nix.lib.project.loadPyproject {
          projectRoot = ./.;
        };
        python = pkgs.python312.withPackages (ps: with ps; [
          nextcord
          sqlalchemy
          aiosqlite
        ]);
        pythonApp = python.pkgs.buildPythonPackage (project.renderers.buildPythonPackage { inherit python; });
      in
      {
        packages.default = pythonApp;
        packages.docker = pkgs.dockerTools.buildImage {
          name = "contentapi-discord-bridge";
          tag = "latest";
          contents = [
            pythonApp
          ];
          config = {
            Cmd = [ "${pythonApp}/bin/cli" ];
          };
        };
        devShell = pkgs.mkShell {
          buildInputs=[
            pkgs.python312
            pkgs.python312Packages.aiosqlite
            pkgs.python312Packages.sqlalchemy
            pkgs.python312Packages.nextcord
            (pkgs.poetry.override { python3 = pkgs.python312; })
          ];
        };
      }
    );
}
