{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule rec {
    pname = "pomo-local";
    version = "0.25.0";
    src = pkgs.fetchFromGitHub {
        owner = "unlomtrois";
        repo = "pomo.local";
        tag = "v${version}";
        hash = "sha256-ZdfV19gvMB70RqcezrgQbnzNwn3wjw0+1cDYTif8AN0=";
    };
    subPackages = [ "cmd/pomo" ];

    vendorHash = "sha256-8kIP7fxIoYq+09EJIM1TmkO9O3zY04SVyDrNMgdBhEI=";

    ldflags=["-X main.version=${version}"];

    buildInputs = [ pkgs.libnotify ];

    meta = with pkgs.lib; {
        description = "Simple pomodoro cli";
        homepage = "https://github.com/unlomtrois/pomo.local";
        license = licenses.mit;
        platforms = platforms.linux;
        mainProgram = "pomo";
        maintainers = with maintainers; [ unlomtrois ];
    };
}
