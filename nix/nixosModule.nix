{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.firefly-iii-boc-fixer;

  defaultSettings = {
    LOG_LEVEL = "INFO";
    FIREFLY_BOC_FIXER_PORT = "3300";
    FIREFLY_BOC_FIXER_HOST = "127.0.0.1";
  };
in {
  options.services.firefly-iii-boc-fixer = {
    enable = lib.mkEnableOption "Sidecar service to fix BoC transaction in Firefly";

    environmentFile = lib.mkOption {
      default = null;
      description = ''
        Environment file (see {manpage}`systemd.exec(5)` "EnvironmentFile="
        section for the syntax) passed to the service. This option can be
        used to safely include secrets in the configuration.
      '';
      example = "/run/secrets/firefly-iii-boc-fixer-envfile";
      type = with lib.types; nullOr path;
    };

    settings = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      description = ''
        Firefly III BoC fixer settings passed as Nix attribute set, they will be merged with
        the defaults. Settings will be passed as environment variables.
      '';
      default = defaultSettings;
      example = {
        FIREFLY_URL = "https://firefly.example.com";
      };
    };
  };
  config = let
    # User-provided settings should be merged with default settings,
    # overwriting where necessary
    mergedConfig = defaultSettings // cfg.settings;
  in
    lib.mkIf cfg.enable {
      systemd.services.firefly-iii-boc-fixer = {
        wantedBy = ["multi-user.target"];
        after = ["network.target"];
        description = "Firefly iii BoC fixer";
        environment = mergedConfig;
        serviceConfig =
          {
            DynamicUser = true;
            ExecStart = "${lib.getExe pkgs.firefly-iii-boc-fixer}";
            Restart = "on-failure";
            RestartSec = "5s";
          }
          // lib.optionalAttrs (cfg.environmentFile != null) {EnvironmentFile = cfg.environmentFile;};
      };
    };
}
