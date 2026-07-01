{
  lib,
  buildGoModule,
  makeBinaryWrapper,
  lowdown,
  xdg-utils,
  chafa 
}:
buildGoModule {
  pname = "nyaa";
  version = "unstable";

  src = ./.;

  proxyVendor = true;
  vendorHash = "sha256-S9YKUvRgmeI/S797ZSp30RKJZGM6Dwys64MfOy6Pgpg=";

  nativeBuildInputs = [ makeBinaryWrapper ];

  postInstall = ''
    wrapProgram $out/bin/nyaa \
      --prefix PATH : ${lib.makeBinPath [ lowdown xdg-utils chafa ]}
  ''; 

  meta = with lib; {
    description = "a simple nyaa.si client";
    homepage = "https://github.com/yaaaarn/nyaa";
    license = licenses.mit;
    mainProgram = "nyaa";
  };
}
