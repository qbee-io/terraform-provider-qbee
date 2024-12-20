---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "qbee_filemanager_file Resource - qbee"
subcategory: ""
description: |-
  The filemanager_file resource allows you to create and manage files in the file manager.
---

# qbee_filemanager_file (Resource)

The filemanager_file resource allows you to create and manage files in the file manager.

## Example Usage

```terraform
# Uploads the file 'local/path/file.txt' to '/root/path/file.txt'.
resource "qbee_filemanager_file" "example" {
  path        = "/root/path/file.txt"
  source      = "/tmp/example.txt"
  file_sha256 = filesha256("local/path/file.txt")
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `file_sha256` (String) The filebase64sha256 of the source file. Required to ensure resource updates if the file changes.
- `path` (String) The full path of the uploaded file.
- `sourcefile` (String) The source file to upload.

## Import

Import is supported using the following syntax:

```shell
# Filemanager file can be imported by specifying the full path as the identifier.
terraform import qbee_filemanager_file.example /full/path/to/file.txt
```
