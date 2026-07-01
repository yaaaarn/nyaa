# 🐈 nyaa

a simple cli client for [nyaa.si](https://nyaa.si/). search, view, and download torrents from the terminal.

## prerequisites

- `lowdown`
- `chafa`

## install

### go

```
go install github.com/yaaaarn/nyaa@latest
```

or to run temporarily

```
go run github.com/yaaaarn/nyaa@latest
```

### nix flake

add to your `flake.nix` inputs:

```nix
nyaa = {
  url = "github:yaaaarn/nyaa";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

then add `nyaa.packages.${system}.default` to your `environment.systemPackages` or home-manager packages.

## dev

```bash
# enter the dev shell (if using nix)
nix develop

# install dependencies
go mod tidy

# build the binary
go build .
```

## license

mit

