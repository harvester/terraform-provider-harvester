resource "harvester_backup_target" "cloudflare-r2" {
  name = "cloudflare-r2"
  type = "s3"
  endpoint_url = "https://324293h49ugfwgwr.r2.cloudflarestorage.com"
  access_key = "a99454f829b7a73049f042360b125493"
  secret_access_key = "8316e1a2d71e7e63af82373140d1c3cec86346c454f60207024cd7030a186fe5"
  bucket_name = "vm-backups"
  bucket_region = "auto"
  virtual_hosted = false
}
