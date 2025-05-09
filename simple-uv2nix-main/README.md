# Simple `uv2nix` devshell template for python development in NixOS ❄️🐍

I had trouble finding a simple template for NixOS to setup up a python development environment with nothing sophisticated but dependency and python version management. This guide explains how to set up the environment for a desired Python version. 

---

## Steps to Customize the Environment 🚀

### 1. Set the Python Version 🐍

1. Open the `.python-version` file.
2. Update its contents to the desired Python version (e.g., `3.10`).

### 2. Update `pyproject.toml` ✏️

1. Open the `pyproject.toml` file.
2. Locate the `[project]` section.
3. Update the `requires-python` field to match the desired Python version range. For example:
   ```toml
   requires-python = ">=3.11,<3.12"
   ```

### 3. Regenerate the `uv.lock` File 🔄

1. Run the following command in your terminal to regenerate the `uv.lock` file (nix will complain if it doesn't exist): 
   ```bash
   nix run nixpkgs#uv lock
   ```

### 4. Enter the Development Environment 🌟

1. Use `nix develop` to enter the development environment:
2. Your environment should now be configured and ready to use with the selected Python version.

---

## Usage 🧐

Use `uv run python` to run the configured python interpreter. To manage dependencies use `uv add <package>` and `uv remove <package>`. If `pyproject.toml` is modified directly without using `uv` commands, run `uv lock` to regenerate the `uv.lock` file.

