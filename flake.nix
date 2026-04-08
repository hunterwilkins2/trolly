{
  description = "Trolly Grocery List";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  };

  outputs = { self, nixpkgs }: let 
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    fs = nixpkgs.lib.fileset;
    goFiles = fs.unions [ ./cmd ./internal (fs.fileFilter (file: nixpkgs.lib.hasSuffix ".go" file.name) ./components) ];
    staticFiles = fs.unions [ ./static/css/dist/output.css ./static/img ./static/js ];
    source = fs.unions [ 
      ./go.mod
      ./go.sum
      goFiles
      staticFiles
    ];
  in {
    nixosModules.trolly = { config, lib, pkgs, ... }: import ./trolly.nix { 
      inherit config lib pkgs;
      trolly = self.packages.${system}.default; 
      migration = self.packages.${system}.migration; 
    };

    packages.${system} = {
      default = pkgs.buildGoModule (finalArgs:  {
        name = "trolly";
        version = "2.0.0";
        vendorHash = "sha256-WGisCMPhXDsXp+hmrn6dRRf8OqlfFIDjEl63otNHong=";
        src = fs.toSource { 
          root = ./.;
          fileset = source;
        };
        goSum = ./go.sum;
        postInstall = ''
          mv $out/bin/web $out/bin/trolly
          cp -r static $out/bin
        '';
      });

      migration = pkgs.writeShellApplication {
        name = "trolly-migrate";
        runtimeInputs = let 
          go-migrate-pg = pkgs.go-migrate.overrideAttrs(oldAttrs: {
            tags = ["mysql"];
          }); 
        in [ go-migrate-pg ]; 
        text = ''
          migrate -path=${./migrations} -database="mysql://$DB_USER@$DB_HOST/trolly" up
        '';
      };
    };
  }; 
}
