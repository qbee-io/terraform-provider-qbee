# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Revert the file_sha256 property removal. It is required to trigger updates when the local contents change.
 
## [0.5.1] - 2023-09-11

### Fixed

- Set configuration of file_distribution and software_management to 'enabled'

## [0.5.0] - 2023-09-11

### Changed

- file_sha256 property of qbee_filedistribution_file is now computed and read-only

## [0.4.1] - 2023-09-08

### Fixed

- Support handling of empty configuration responses (no bundle data) from qbee

## [0.4.0] - 2023-09-07

### Added

- Add support for the 'tags' property for qbee_grouptree_group

## [0.3.0] - 2023-09-03

### Added

- Support for the qbee_softwaremanagement resource

### Changed

- qbee_tag_filedistribution and qbee_node_filedistribution have been merged.

## [0.2.0] - 2023-08-31

### Added

- qbee_filemanager_directory resources can now be imported
- qbee_tag_filedistribution resources can now be imported
- Support for the qbee_node_filedistribution resource

### Changed

- The qbee_filemanager_directory resource now has a single required attribute, 'path', replacing 
  'parent' and 'name' and the computed 'path' attributes.
- The qbee_filemanager_file resource now uses a 'path' attribute to replace 'parent' and 'name'.
- The qbee_filemanager_directory now detects resource drift
- The qbee_filemanager_file now detects resource drift
- The qbee_tag_filedistribution now correctly handles unset command and pre_condition attributes

## [0.1.0] - 2023-08-29

### Added

- Support for the qbee_filemanager_file resource
- Support for the qbee_filemanager_directory resource
- Support for the qbee_filedistribution resource
- Support for the qbee_grouptree_group resource
