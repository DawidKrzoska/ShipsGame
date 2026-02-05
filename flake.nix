{
  description = "WebShips Node.js dev shell with Playwright";

  inputs = { nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; };

  outputs = { self, nixpkgs }:
    let
      systems =
        [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in {
      devShells = forAllSystems (system:
        let pkgs = import nixpkgs { inherit system; };
        in {
          default = pkgs.mkShell {
            packages = [
              pkgs.nodejs_20
              pkgs.playwright-driver
              pkgs.chromium
            ];

            shellHook = ''
              echo "Node: $(node -v)"
              echo "npm:  $(npm -v)"
              export PLAYWRIGHT_BROWSERS_PATH=${pkgs.playwright-driver.browsers}
              export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
              export PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=${pkgs.chromium}/bin/chromium
            '';
          };
        });
    };
}
