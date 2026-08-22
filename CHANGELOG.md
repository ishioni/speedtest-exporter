# Changelog

## [0.2.5](https://github.com/ishioni/speedtest-exporter/compare/0.2.4...0.2.5) (2026-08-22)


### Features

* **container:** update image grafana/grafana (13.1.3 → 13.2.0) ([#45](https://github.com/ishioni/speedtest-exporter/issues/45)) ([cfc5221](https://github.com/ishioni/speedtest-exporter/commit/cfc52214a9a733f3974d1024c864ebd7c7130503))
* **container:** update image prom/prometheus (v3.13.2 → v3.14.0) ([#44](https://github.com/ishioni/speedtest-exporter/issues/44)) ([c18a3ca](https://github.com/ishioni/speedtest-exporter/commit/c18a3ca4f24d54a9081c1cc8215b3fccac9c2134))
* **mise:** update tool go (1.26.6 → 1.27.0) ([#50](https://github.com/ishioni/speedtest-exporter/issues/50)) ([a122218](https://github.com/ishioni/speedtest-exporter/commit/a12221859f9d1d563f34f90fa8a400d03de5ac98))


### Bug Fixes

* **ci:** add node to mise ([ba6bae7](https://github.com/ishioni/speedtest-exporter/commit/ba6bae7d05a7711628cf7e742bbd69576a79cc36))
* **renovate:** bump Go module minor versions ([#51](https://github.com/ishioni/speedtest-exporter/issues/51)) ([c4d967a](https://github.com/ishioni/speedtest-exporter/commit/c4d967ae2771dc8e73f0ae20d84d39f66a9c9629))


### Miscellaneous Chores

* **ci:** update release-please config ([1f8adeb](https://github.com/ishioni/speedtest-exporter/commit/1f8adebcc000451cca9e2a0e1cd9a44e4af71525))
* **github-action:** update action docker/github-builder (v1.16.0 → v1.17.0) ([#49](https://github.com/ishioni/speedtest-exporter/issues/49)) ([be3695a](https://github.com/ishioni/speedtest-exporter/commit/be3695a841fe97c0263581422e32ba4fa919159a))
* **mise:** update mise tools ([#52](https://github.com/ishioni/speedtest-exporter/issues/52)) ([219d5bd](https://github.com/ishioni/speedtest-exporter/commit/219d5bda42e8b1acc3b2cf0832039ebc052d4eaf))
* **mise:** update tool oxfmt (0.63.0 → 0.64.0) ([#46](https://github.com/ishioni/speedtest-exporter/issues/46)) ([f20a316](https://github.com/ishioni/speedtest-exporter/commit/f20a316bee2342ef998ed5c234c89c32e06b75a6))
* **mise:** update tool yq (4.53.3 → 4.53.4) ([#48](https://github.com/ishioni/speedtest-exporter/issues/48)) ([de18c15](https://github.com/ishioni/speedtest-exporter/commit/de18c150c54c169db74dcc55b2600632589de364))
* run go fix with Go 1.27 ([#53](https://github.com/ishioni/speedtest-exporter/issues/53)) ([a739cbb](https://github.com/ishioni/speedtest-exporter/commit/a739cbba2fed6240b4b1376ea827229ac1833163))

## [0.2.4](https://github.com/ishioni/speedtest-exporter/compare/0.2.3...0.2.4) (2026-08-14)


### Bug Fixes

* **container:** update image grafana/grafana (13.1.1 → 13.1.2) ([#33](https://github.com/ishioni/speedtest-exporter/issues/33)) ([381755a](https://github.com/ishioni/speedtest-exporter/commit/381755a0ac9f0a69d9c117e36d40fe110458e05d))
* **container:** update image grafana/grafana (13.1.2 → 13.1.3) ([#35](https://github.com/ishioni/speedtest-exporter/issues/35)) ([b56bf02](https://github.com/ishioni/speedtest-exporter/commit/b56bf0294c73b863f08a9dd84ac2d3b999a5a964))
* **docker:** source Go version from build argument ([55abb20](https://github.com/ishioni/speedtest-exporter/commit/55abb2052e9e07e9dad11686beb807b991600b1d))
* **go:** update module go (1.26.5 → 1.26.6) ([#40](https://github.com/ishioni/speedtest-exporter/issues/40)) ([ee2c5fc](https://github.com/ishioni/speedtest-exporter/commit/ee2c5fc00b65ac7fe86a857cdd400b84573b7407))
* **renovate:** group Go toolchain updates ([cda561a](https://github.com/ishioni/speedtest-exporter/commit/cda561afdc09bfa3777e83269470487e26b48076))

## [0.2.3](https://github.com/ishioni/speedtest-exporter/compare/0.2.2...0.2.3) (2026-07-31)


### Bug Fixes

* accept speedtest retry logs ([#29](https://github.com/ishioni/speedtest-exporter/issues/29)) ([8929a94](https://github.com/ishioni/speedtest-exporter/commit/8929a943170a8eeeffe6065619a0120c49264f8c))
* **container:** update image prom/prometheus (v3.13.1 → v3.13.2) ([#27](https://github.com/ishioni/speedtest-exporter/issues/27)) ([679da31](https://github.com/ishioni/speedtest-exporter/commit/679da31338da6f9c21a85ed4e0e845f61b790e59))
* **deps:** update module github.com/prometheus/client_golang (v1.24.0 → v1.24.1) ([#23](https://github.com/ishioni/speedtest-exporter/issues/23)) ([4e875f6](https://github.com/ishioni/speedtest-exporter/commit/4e875f64d7cc0031755ca5128e5e07c29829f19b))

## [0.2.2](https://github.com/ishioni/speedtest-exporter/compare/0.2.1...0.2.2) (2026-07-23)


### Features

* **deps:** update module github.com/prometheus/client_golang (v1.23.2 → v1.24.0) ([#18](https://github.com/ishioni/speedtest-exporter/issues/18)) ([a6c8804](https://github.com/ishioni/speedtest-exporter/commit/a6c8804dcf7bd3139df328dfa0e4be09887c62b1))


### Bug Fixes

* **container:** update image grafana/grafana (13.1.0 → 13.1.1) ([#10](https://github.com/ishioni/speedtest-exporter/issues/10)) ([dfd67ae](https://github.com/ishioni/speedtest-exporter/commit/dfd67ae97aae6a763d074b26e63fe871d60e01e3))
* omit invalid speedtest measurements ([#22](https://github.com/ishioni/speedtest-exporter/issues/22)) ([0e471f8](https://github.com/ishioni/speedtest-exporter/commit/0e471f834ab278ba4a256f0ff393905c022db339))

## [0.2.1](https://github.com/ishioni/speedtest-exporter/compare/0.2.0...0.2.1) (2026-07-18)


### Bug Fixes

* **helm:** saner default values ([d53d4f4](https://github.com/ishioni/speedtest-exporter/commit/d53d4f4b65752b8dfa9ce205e5268860f2f31838))

## [0.2.0](https://github.com/ishioni/speedtest-exporter/compare/0.1.0...0.2.0) (2026-07-18)


### ⚠ BREAKING CHANGES

* **container:** Update image grafana/grafana (12.0.2 → 13.1.0) ([#12](https://github.com/ishioni/speedtest-exporter/issues/12))

### Features

* **container:** update image curlimages/curl (8.17.0 → 8.21.0) ([#9](https://github.com/ishioni/speedtest-exporter/issues/9)) ([1463355](https://github.com/ishioni/speedtest-exporter/commit/14633559284a7063f60ba2a9b616eb8523e689e4))
* **container:** Update image grafana/grafana (12.0.2 → 13.1.0) ([#12](https://github.com/ishioni/speedtest-exporter/issues/12)) ([72507bb](https://github.com/ishioni/speedtest-exporter/commit/72507bb385a5ed1011c4f23a4081ee0b05020371))
* **container:** update image prom/prometheus (v3.4.1 → v3.13.1) ([#11](https://github.com/ishioni/speedtest-exporter/issues/11)) ([52f1203](https://github.com/ishioni/speedtest-exporter/commit/52f12038e205f427c2e15c6b313f502b977a74c3))

## 0.1.0 (2026-07-18)


### Features

* rewrite speedtest exporter in Go ([#1](https://github.com/ishioni/speedtest-exporter/issues/1)) ([008da55](https://github.com/ishioni/speedtest-exporter/commit/008da55b395a006f63262d5be4c624612145f597))


### Bug Fixes

* bootstrap release please at 0.1.0 ([#3](https://github.com/ishioni/speedtest-exporter/issues/3)) ([5098b8d](https://github.com/ishioni/speedtest-exporter/commit/5098b8de4523384b1c6d24a1ec6693db49e3fd60))
* ignore release please generated files ([#5](https://github.com/ishioni/speedtest-exporter/issues/5)) ([99fb101](https://github.com/ishioni/speedtest-exporter/commit/99fb101670a2f2d9d62549ac423f224d0ed13c8c))

## Changelog
