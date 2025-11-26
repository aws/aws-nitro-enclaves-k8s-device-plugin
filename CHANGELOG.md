# Changelog

All notable changes to the AWS Nitro Enclaves Kubernetes Device Plugin will be documented in this file.

## [v0.4.0] - 11/26/2025

### Changed
- Upgraded base image from Amazon Linux 2 to Amazon Linux 2023 for extended support
- Updated Go toolchain from 1.22.7 to 1.24.0 with toolchain 1.24.9
- Updated Helm chart version to 0.4.0
- Updated app version to 0.4.0

### Dependencies
- Upgraded k8s.io/kubelet from v0.25.3 to v0.30.14 (LTS)
- Upgraded google.golang.org/grpc from v1.64.1 to v1.75.0
- Upgraded golang.org/x/net from v0.33.0 to v0.47.0
- Upgraded golang.org/x/sys from v0.30.0 to v0.38.0
- Upgraded golang.org/x/text from v0.22.0 to v0.31.0
- Upgraded google.golang.org/protobuf from v1.33.0 to v1.36.10
- Upgraded github.com/fsnotify/fsnotify from v1.6.0 to v1.8.0
- Upgraded github.com/golang/glog from v1.2.4 to v1.2.5

### Fixed
- Fixed typo in Dockerfile comment: "minumum" → "minimum"

## [v0.3.1] - 07/18/2025

### Added
- Support for specifying DaemonSet tolerations via Helm values ([aebfd0c](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/aebfd0c))

### Changed
- Renamed scripts/pipeline.sh to scripts/release.sh for better clarity ([eae0a9e](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/eae0a9e))
- Removed `v` prefix from docker image tag for consistency ([ca16d51](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/ca16d51))
- Updated Helm chart version to 0.3.1
- Updated app version to 0.3.1

## [v0.3] - 04/24/2025

### Added
- vCPUs advertisement for enclaves ([31739f1](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/31739f1))
- Config package for improved configuration management ([819d133](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/819d133))
- Helm chart support ([294f96d](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/294f96d))
- Helm README documentation ([a3848fc](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/a3848fc))
- GitHub workflow for CI/CD ([dba1171](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/dba1171))
- Pipeline orchestration script ([adc13a0](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/adc13a0))
- Helm related scripts for deployment ([24fe4fd](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/24fe4fd))

### Changed
- Refactored device-plugin monitor to avoid code duplication ([fefd765](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/fefd765))
- Refactored device-plugin project structure ([acc9f00](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/acc9f00))
- Extended common.sh functionality ([7410223](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/7410223))
- Added _docker suffix for docker build scripts ([c3bea54](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/c3bea54))
- Added plugin config options to Helm chart ([8244d12](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/8244d12))

### Dependencies
- Bumped golang/glog to v1.2.4 ([a0a4cb7](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/a0a4cb7))

## [v0.2] - 01/29/2025

### Added
- Support for 4 enclaves per instance ([53661fd](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/53661fd))

### Fixed
- Build process issues ([d1595e1](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/d1595e1))

### Changed
- Updated go build process

## [v0.1] - 02/18/2023

### Added
- First version of `aws-nitro-enclave-k8s-device-plugin`
- Initial implementation of Kubernetes device plugin for AWS Nitro Enclaves

### Security
- Bump golang.org/x/net from 0.1.0 to 0.7.0 ([93dbf9e](https://github.com/aws/aws-nitro-enclaves-k8s-device-plugin/commit/93dbf9e))
