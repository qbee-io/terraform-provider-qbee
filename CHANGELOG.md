# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- qbee_tag_filedistribution resources can now be imported
- Support for the qbee_node_filedistribution resource

### Changed

- The qbee_filemanager_directory resource now has a single required attribute, 'path', replacing 
  'parent' and 'name' and the computed 'path' attributes.

## [0.2.0] - 2023-08-29

### Added

- qbee_filemanager_directory resources can now be imported

## [0.1.0] - 2023-08-29

### Added

- Support for the qbee_filemanager_file resource
- Support for the qbee_filemanager_directory resource
- Support for the qbee_filedistribution resource
- Support for the qbee_grouptree_group resource
