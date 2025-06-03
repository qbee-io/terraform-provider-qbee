resource "qbee_role" "test" {
  name        = "device-manager"
  description = "A role for managing devices"
  policies = [
    {
      permission = "device:read"
      resources  = ["*"]
    },
    {
      permission = "device:manage"
      resources  = ["*"]
    }
  ]
}
