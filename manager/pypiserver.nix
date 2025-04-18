{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    pkgs.python311
    pkgs.python311Packages.pypiserver
    pkgs.python311Packages.setuptools
  ];

  build-system = [
    pkgs.python311Packages.setuptools
    pkgs.python311Packages.setuptools-scm
  ];

  
  dependencies = [
    pkgs.python311Packages.attrs
    pkgs.python311Packages.py
    pkgs.python311Packages.setuptools
    pkgs.python311Packages.six
    pkgs.python311Packages.pluggy
  ];

  shellHook = ''
    echo "Run: pypi-server run -p 8080 ./packages"
  '';
}
