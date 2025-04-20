rpm.build({
  name: "techaro-repos-stable",
  description: "Repo definitions for stable Techaro packages",
  homepage: "https://techaro.lol",
  license: "MIT",
  goarch: "all",
  version: "1.0.0",

  build: (out) => {
    file.install("./techaro-pkgs.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-pkgs.asc`);
    file.install("./techaro-root.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-root.asc`);
    file.install("./techaro-stable.repo", `${out}/etc/yum.repos.d/techaro-stable.repo`);
  },
});

rpm.build({
  name: "techaro-repos-unstable",
  description: "Repo definitions for unstable Techaro packages",
  homepage: "https://techaro.lol",
  license: "MIT",
  goarch: "all",
  version: "1.0.0",

  build: (out) => {
    file.install("./techaro-pkgs.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-pkgs.asc`);
    file.install("./techaro-root.pub.asc", `${out}/etc/pki/rpm-gpg/techaro.lol-keys/techaro-root.asc`);
    file.install("./techaro-stable.repo", `${out}/etc/yum.repos.d/techaro-stable.repo`);
  },
});