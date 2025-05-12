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
  };

  outputs = { self, nixpkgs, uv2nix, pyproject-nix, pyproject-build-systems, ... }:
    let
      inherit (nixpkgs) lib;
      system = "x86_64-linux";  # Specify the system explicitly

      workspace = uv2nix.lib.workspace.loadWorkspace { workspaceRoot = ./.; };

      overlay = workspace.mkPyprojectOverlay {
        sourcePreference = "wheel";
      };

      pyprojectOverrides = final: prev: {
        artiq = prev.artiq.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
        sipyco = prev.sipyco.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
        pythonparser = prev.pythonparser.overrideAttrs (old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
        });
	pyqt6 = prev.pyqt6.overrideAttrs (old: {
	  dontWrapQtApps = true;

	  #nativeBuildInputs = (old.nativeBuildInputs or []) ++ [ 
	  nativeBuildInputs = (old.nativeBuildInputs or []) ++ [ 
	    pkgs.autoPatchelfHook
	    #pkgs.qt6.wrapQtAppsHook
	  ];
	  propagatedBuildInputs = (old.propagatedBuildInputs or []) ++ (with pkgs; [
            qt6.qtbase
            qt6.qtsvg
            qt6.qtdeclarative
            qt6.qtsensors
	    qt6.qttools
	    qt6.qtmultimedia
	    kdePackages.qtspeech
	    kdePackages.qtserialport
	    kdePackages.qtwebchannel
	    kdePackages.qtscxml
	    kdePackages.qtwebengine
	    kdePackages.qtconnectivity
	    kdePackages.qtremoteobjects
	    kdePackages.full
	   # qt6.full
	    #qt6.webchannel
	   # qt6.webchannel
    	  ]);
  	});
 # pyqt6 = prev.pyqt6.overrideAttrs (old: {
  #      buildInputs = (old.buildInputs or [ ]) ++ [ final.setuptools ];
   #     });
      };

      pkgs = nixpkgs.legacyPackages.${system};
      python = pkgs.python311;

      pythonSet = (pkgs.callPackage pyproject-nix.build.packages {
        inherit python;
      }).overrideScope (
        lib.composeManyExtensions [
          pyproject-build-systems.overlays.default
          overlay
          pyprojectOverrides
        ]
      );

    in
    {
      packages.${system}.default = pythonSet.mkVirtualEnv "artiq-env" workspace.deps.default;

      devShells.${system} = {
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
    	   } // lib.optionalAttrs pkgs.stdenv.isLinux {
#	      LD_LIBRARY_PATH = lib.makeLibraryPath (with pkgs; [
#                pythonManylinuxPackages.manylinux1
#              ]);
	      #1
	      #LD_LIBRARY_PATH = "${pkgs.libGL}/lib";
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
