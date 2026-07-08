{
  description = "md-tui — Sony NetMD MiniDisc TUI dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          # go_1_26 matches the `go 1.26.1` directive in go.mod.
          # pkg-config + libusb1 satisfy gousb's `#cgo pkg-config: libusb-1.0`.
          # ffmpeg is required at runtime for non-WAV upload conversion.
          packages = with pkgs; [
            go_1_26
            pkg-config
            libusb1
            ffmpeg
          ];

          shellHook = ''
            echo "md-tui dev shell: $(go version)"
            if ! command -v atracdenc >/dev/null; then
              echo "Note: LP2 uploads need 'atracdenc' in PATH — not in nixpkgs; install via your atracdenc overlay."
            fi
          '';
        };
      });
    };
}
