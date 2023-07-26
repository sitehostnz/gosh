# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [v0.3.x] - 2032-03-15
### Fixes
- correct the ssh key update params


## [v0.3.x] - 2032-03-13
### Added
- database list/get/update/remove for `/cloud/stack/image`
- cloud stack image list_all `/cloud/image`
- `/ssh/key/` add/update 

## [v0.3.x] - 2032-03-07
### Added
- helper for handling boolish results
- image list and get endpoints for `/cloud/image` endpoint

## [v0.3.x] - 2032-03-03
### Fixes
- Fix a crash when unmarshalling when the `/server` returns a different type for the server disk size depending on the type of server.

### Added
- Add some helpers for filtering the DNS list and getting new records since there is no get record end point, and the add api does not add the new record id 

## [v0.3.1] - 2023-02-28
### Added
- Added `/cloud/server`, `/cloud/stack`, `/cloud/stack/environment` endpoints.
- Added `/dns` endpoint.
- Added `/ssh` endpoint.
- Added `/api/get_info` endpoint.

## [v0.3.0] - 2022-12-08
### Added
- GitHub PR actions to run go vet, go lint and go mod tidy.

### Updated
- Updated project layout and structure.
- Moved and refactored server and job api endpoints.
- Updated our code to conform with Golang linter.
- Updated our README to link to our Golang style and our license.
- Upgraded golang v1.19.3.
- Upgraded golangci-lint v1.50.1.

## [v0.2.2] - 2022-05-20
### Added
- Added support for SSH keys when provisioning a server.

## [v0.2.1] - 2022-05-16
### Added
- Added label update function. 