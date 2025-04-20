rpm.build({
  name: "techaro-repo-keys",
  description: "Public keys for techaro.lol RPM packages",
  homepage: "https://techaro.lol",
  license: "MIT",
  goarch: "all",

  build: (out) => {
    yeet.run(`mkdir`, `-p`, `${out}/etc/pki/rpm-gpg/techaro.lol-keys`);
    file.install("./techaro-pkgs.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-pkgs.asc`);
    file.install("./techaro-root.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-root.asc`);
  },
});

deb.build({
  name: "techaro-repo-keys",
  description: "Public keys for techaro.lol RPM packages",
  homepage: "https://techaro.lol",
  license: "MIT",
  goarch: "all",

  build: (out) => {
    yeet.run(`mkdir`, `-p`, `${out}/usr/share/keyrings`);
    file.install("./techaro-pkgs.pub.asc", `${out}/usr/share/keyrings/techaro-pkgs.asc`);
    file.install("./techaro-root.pub.asc", `${out}/usr/share/keyrings/techaro-root.asc`);
  },
});