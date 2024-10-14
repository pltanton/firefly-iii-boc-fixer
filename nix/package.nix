{
  lib,
  buildGoModule,
}:
buildGoModule {
  pname = "firefly-iii-boc-fixer";
  version = "0";
  vendorHash = "sha256-5R4oNEO+S0Q6Bl/6H6RKrBIG+eFnHlDGiuOD73UQG58=";

  meta = with lib; {
    homepage = "https://github.com/pltanton/firefly-iii-boc-fixer";
    description = "Firefly webhook listener to re-format BoC ugly transactions";
    license = licenses.gpl3Only;
    platforms = platforms.linux;
    mainProgram = "firefly-iii-boc-fixer";
  };

  src = lib.cleanSourceWith {
    filter = name: type: let
      baseName = baseNameOf (toString name);
    in
      !(lib.hasSuffix ".nix" baseName);
    src = lib.cleanSource ../.;
  };
}
