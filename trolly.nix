{ config, lib, pkgs, trolly, migration, ... }: 
{
  options = {
    services.trolly = {
      enable = lib.mkEnableOption "Enable trolly web service";
      port = lib.mkOption {
        type = lib.types.port;
        description = "Trolly web service port"; 
        default = 4000;
      };
      db = lib.mkOption {
        type = lib.types.submodule {
          options = {
            host = lib.mkOption {
              type = lib.types.str;
              description = "MySQL host URI";
              default = "unix(/var/run/mysqld/mysqld.sock)";
            };
            user = lib.mkOption {
              type = lib.types.str;
              description = "MySQL user";
              default = "trolly";
            };
            migration-user = lib.mkOption {
              type = lib.types.str;
              description = "MySQL user";
              default = "trolly-migrator";
            };
            name = lib.mkOption {
              type = lib.types.str;
              description = "MySQL Trolly database name";
              default = "trolly";
            };
          };
        };
        default = {};
      };
    };
  };

  config = lib.mkIf config.services.trolly.enable {
    users.users.${config.services.trolly.db.user} = {
      isSystemUser = true;
      group = config.services.trolly.db.user;
    };
    users.users.${config.services.trolly.db.migration-user} = {
      isSystemUser = true;
      group = config.services.trolly.db.migration-user;
    };
    users.groups.${config.services.trolly.db.user} = {}; 
    users.groups.${config.services.trolly.db.migration-user} = {}; 

    environment.systemPackages = [
      trolly
      migration
    ];

    services.mysql = {
      enable = true;
      user  = config.services.trolly.db.user;
      group  = config.services.trolly.db.user;
      package = pkgs.mariadb;
      ensureDatabases = [ config.services.trolly.db.name ];
      ensureUsers = [
        {
          name = config.services.trolly.db.user;
          ensurePermissions = {
            "trolly.*" = "SELECT, INSERT, UPDATE, DELETE";
          };
        }
        {
          name = config.services.trolly.db.migration-user;
          ensurePermissions = {
            "trolly.*" = "SELECT, TRIGGER, INSERT, UPDATE, DELETE, INDEX, CREATE, ALTER, DROP";
          };
        }
      ];
    };

    systemd.services.trolly-migrate = {
      description = "Trolly MySQL DB Migration";
      after = [ "mysql.service" ];
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = "${migration}/bin/trolly-migrate";
        Environment = [
          "DB_USER=${config.services.trolly.db.migration-user}"
          "DB_HOST=${config.services.trolly.db.host}"
          "DB_NAME=${config.services.trolly.db.name}"
        ];
        User = config.services.trolly.db.migration-user;
        Group = config.services.trolly.db.migration-user;
        Restart = "no";
        AmbientCapabilities = "";
        CapabilityBoundingSet = "";
        LockPersonality = true;
        MemoryDenyWriteExecute = false;
        BindPaths = "/var/run/mysqld:/var/run/mysqld";
        MountAPIVFS = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        PrivateMounts = true;
        PrivateTmp = true;
        PrivateUsers = true;
        ProtectClock = true;
        ProtectControlGroups = "strict";
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        ProtectSystem = "strict";
        RemoveIPC = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
          "AF_UNIX"
          "AF_NETLINK"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
        UMask = 27;
      };
    };

    systemd.services.trolly = {
      description = "Trolly web service";
      after = [ "trolly-migrate.service" ];
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        Type = "simple";
        ExecStart = "${trolly}/bin/trolly -port ${builtins.toString config.services.trolly.port} -db-host ${config.services.trolly.db.host} -db-user ${config.services.trolly.db.user} -db-name ${config.services.trolly.db.name}";
        User = config.services.trolly.db.user;
        Group = config.services.trolly.db.user;
        Restart = "always";
        AmbientCapabilities = "";
        CapabilityBoundingSet = "";
        LockPersonality = true;
        MemoryDenyWriteExecute = false;
        WorkingDirectory = "${trolly}/bin";
        BindPaths = "/var/run/mysqld:/var/run/mysqld";
        MountAPIVFS = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        PrivateMounts = true;
        PrivateTmp = true;
        PrivateUsers = true;
        ProtectClock = true;
        ProtectControlGroups = "strict";
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        ProtectSystem = "strict";
        RemoveIPC = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
          "AF_UNIX"
          "AF_NETLINK"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
        UMask = 27;
      };
    };
  };
}
