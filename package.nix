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

  vendorHash = "sha256-8vQC2VZXcB3TkI87pPGikurwFQLzUICeWzBF52oTRfo=";

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
