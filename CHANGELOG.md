# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2020-09-21

### Fixed
- Renamed config parameter "insecure" to "insecure-skip-tls-verify" for clarity

### Enhancements
- Added trusted-ca-file to config generation
- Added timeout duration to config generation

## [0.1.1] - 2020-09-20

### Fixed
- Read $HOME directory properly
- Use more generic utf8 arrow 

## [0.1.0] - 2020-09-20

### Initial Commit
- osc (operate sensu cluster) is a utility to help Sensu operators quickly switch between clusters
in their environment.