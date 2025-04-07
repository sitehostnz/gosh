# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [v0.3.x] - 2024-05-14
### Added
- Added support for all endpoints under `/cloud/db/`.
- list/get/update/remove for `/cloud/db`
- list/get/update/remove for `/cloud/db/user`
- list/get/update/remove for `/cloud/db/grants`
- cloud stack image list_all `/cloud/image`
- Added support for the `/cloud/stack/image/list_all.json` endpoint.

### Fixed
- Corrected the SSH key update parameters.

### Fixes
- Updated Go from 1.19 to 1.22.
- version updates
- correct the ssh key update params
- Changed the type of the `CustomImageAccess` API request struct fields from
- add wrappers around api return types to get bools, not ints.

## [v0.3.4] - 2024-03-12
### Added
- Add ability to update SSH Keys.

## [v0.3.3] - 2024-03-12
### Added
- Add ability to create, get and delete SSH Keys.

## [v0.3.2] - 2023-03-22
### Fixed
- Fix a crash when unmarshalling when the `/server` returns a different type for the server disk size depending on the type of server.
- Fix GetRecordWithRecord with default priority.
- Fix ListRecords to remove the final dot in the content value.
- Fix the priority value in the UpdateRecord function.

### Added
- Add some helpers for filtering the DNS list and getting new records since there is no get record end point, and the add api does not add the new record id.
- Add helper for handling boolish results.
- Add image list and get endpoints.

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
