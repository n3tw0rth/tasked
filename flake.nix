{
  description = "Manage Google Tasks from your terminal with priorities";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        default = tasked;

        # Signs in with tasked's built-in public OAuth client; users can set
        # TASKED_CLIENT_ID/SECRET at runtime to use their own instead.
        tasked = pkgs.buildGoModule rec {
          pname = "tasked";
          version = "0.1.0";
          src = self;
          vendorHash = "sha256-xQ9UYZJJFM1cwT7PcXzh1lNOoRr4Cx311n9id0F5u4c=";
          env.CGO_ENABLED = 0;
          ldflags = [ "-s" "-w" "-X github.com/n3tw0rth/tasked/cmd.version=${version}" ];
          meta = {
            description = "Manage Google Tasks from your terminal with priorities";
            homepage = "https://github.com/n3tw0rth/tasked";
            license = pkgs.lib.licenses.mit;
            mainProgram = "tasked";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [ go gopls just ];
        };
      });
    };
}
