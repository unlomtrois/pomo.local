{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
    packages = with pkgs; [ go_1_26 libgcc gnumake libnotify ];

    shellHook = ''
      echo "Dev environment loaded!"
      gcc --version | head -n 1
      make --version | head -n 1
      go version
    '';
}
