{
  description = "Open specified browser depends on contextual rules";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flakelight.url = "github:nix-community/flakelight";
  };

  outputs = {flakelight, ...}:
    flakelight ./. {
      devShell.packages = pkgs: with pkgs; [go alejandra dprint];
      formatters = {
        "*.yml" = "dprint fmt";
        "*.md" = "dprint fmt";
        "*.nix" = "alejandra";
      };
    };
}
