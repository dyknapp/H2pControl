{
  description = "ARTIQ uv2nix environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    pyproject-nix = {
      url = "github:pyproject-nix/pyproject.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    uv2nix = {
      url = "github:pyproject-nix/uv2nix";
      inputs.pyproject-nix.follows = "pyproject-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    pyproject-build-systems = {
      url = "github:pyproject-nix/build-system-pkgs";
      inputs.pyproject-nix.follows = "pyproject-nix";
      inputs.uv2nix.follows = "uv2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };


    uv2nix_hammer_overrides.url = "github:TyberiusPrime/uv2nix_hammer_overrides";
    uv2nix_hammer_overrides.inputs.nixpkgs.follows = "nixpkgs";

  };


  outputs = { self, nixpkgs, uv2nix, pyproject-nix, pyproject-build-systems, uv2nix_hammer_overrides, ... }:
    let
      inherit (nixpkgs) lib;
      system = "x86_64-linux";  # Specify the system explicitly

      workspace = uv2nix.lib.workspace.loadWorkspace { workspaceRoot = ./.; };

      overlay = workspace.mkPyprojectOverlay {
        sourcePreference = "wheel";
      };

      pkgs = nixpkgs.legacyPackages.${system};
      python = pkgs.python311;

      pyprojectOverrides = final: prev: {
        artiq = prev.artiq.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
        artiq-comtools = prev.artiq-comtools.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
        sipyco = prev.sipyco.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
        pythonparser = prev.pythonparser.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
      };
#      pythonSet = (pkgs.callPackage pyproject-nix.build.packages {
#        inherit python;
#      }).overrideScope (
#        lib.composeManyExtensions [
#          pyproject-build-systems.overlays.default
#          overlay
#          pyprojectOverrides
#        ]
#      );
#
     pythonSet =
        # Use base package set from pyproject.nix builders
        (pkgs.callPackage pyproject-nix.build.packages {
          inherit python;
        }).overrideScope
          (
            lib.composeManyExtensions [
              pyproject-build-systems.overlays.default
              overlay
              pyprojectOverrides
              (import ./pyproject-overrides.nix pkgs)
            ]
          );



    in
    {
      packages.${system}.default = pythonSet.mkVirtualEnv "artiq-env" workspace.deps.default;

      devShells.${system} = {

        # It is of course perfectly OK to keep using an impure virtualenv workflow and only use uv2nix to build packages.
        # This devShell simply adds Python and undoes the dependency leakage done by Nixpkgs Python infrastructure.
        impure = pkgs.mkShell {
          packages = [
            python
            pkgs.uv
          ];
          shellHook =
            let
              libraries = [
                pkgs.libGL
                pkgs.stdenv.cc.cc.lib
                pkgs.glib
                pkgs.zlib
                "/run/opgengl-driver"
                pkgs.libxkbcommon
                pkgs.fontconfig
                pkgs.xorg.libX11
                pkgs.freetype
                pkgs.dbus

                pkgs.xorg.libxcb
                pkgs.xorg.xcbutil
                pkgs.xorg.xcbutilcursor
                pkgs.xorg.xcbutilerrors
                pkgs.xorg.xcbutilimage
                pkgs.xorg.xcbutilkeysyms
                pkgs.xorg.xcbutilrenderutil
                pkgs.xorg.xcbutilwm

                pkgs.zstd
              ];
            in
            ''
              # fixes libstdc++ issues and libgl.so issues
              export LD_LIBRARY_PATH="${pkgs.lib.makeLibraryPath libraries}"
              # https://github.com/NixOS/nixpkgs/issues/80147#issuecomment-784857897
              # pyqt6 config, starting at artiq 9
              # export QT_PLUGIN_PATH="${pkgs.qt6.qtbase}/${pkgs.qt6.qtbase.qtPluginPrefix}"

              # PYQT5 FOR ARTIQ 8
              export QT_PLUGIN_PATH="${pkgs.qt5.qtbase}/${pkgs.qt5.qtbase.qtPluginPrefix}"
              # export QT_DEBUG_PLUGINS=1

              # uv2nix environment settings
              unset PYTHONPATH
              export UV_PYTHON_DOWNLOADS=never

              echo ""
              echo "    uv venv"
              echo "    source .venv/bin/activate"
              echo "    uv sync --all-extras"
              echo ""

            '';
        };

        # Pure development shell using uv2nix
        default = 
          let
            editableOverlay = workspace.mkEditablePyprojectOverlay {
              root = "$REPO_ROOT";
            };

            editablePythonSet = pythonSet.overrideScope editableOverlay;
            virtualenv = editablePythonSet.mkVirtualEnv "artiq-dev-env" workspace.deps.all;
          in
          pkgs.mkShell {
            packages = [
              virtualenv
              pkgs.uv
            ];

            env = {
              UV_NO_SYNC = "1";
              UV_PYTHON = python.interpreter;
              UV_PYTHON_DOWNLOADS = "never";
            };

            shellHook = ''
              unset PYTHONPATH
              export REPO_ROOT=$(git rev-parse --show-toplevel)

              echo "========================================="
              echo "Debug information:"
              echo "Current directory: $(pwd)"
              #echo "REPO_ROOT: $REPO_ROOT"
              #echo "PYTHONPATH: $PYTHONPATH"
              #echo "PATH: $PATH"
              echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
              echo "========================================="

              echo "----------------------------------------"
              echo "🚀 ARTIQ development environment activated"
              echo "Python: $(python --version)"
              echo "ARTIQ commands should be available in PATH:"
              which artiq_client || echo "❌ artiq_client not found"
              which artiq_run || echo "❌ artiq_run not found"
              echo "----------------------------------------"
            '';
          };
      };
    };
}
