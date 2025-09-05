provider "altr" {
}


resource "altr_sidecar" "test_sidecar" {
  name  = "test-sidecar"
  description     = "test description"
  hostname     = "localhost"
  unsupported_query_bypass = false
  public_key_1 = "-----BEGIN PUBLIC KEY-----\nMIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQB2VO3t0vpIEKFHbC8c3Ljf\n8b4dSthL7LHXsYtqEvnHWPVbDDJXyp8nVLEdcH6dM7lzOZncfhHHVzREgYgV8LRV\nrMEX4TgkmNFhHk/HasY56BwEU6Vy75L5i34xwkIIGFVNyO/DCJHqFKvCVfTv4tOP\n1Qk7T205Wa53f7lINKQp21BMncl1PjjN1TIqL4IKCZ4AWQB0tybZ6rojODKABzSU\n3ljF4GXkvJUPjfXEfylcarSjukQRurJEhd8vcxZlaVvMWXV2vOy85K/YS2s/BIMa\n9WC6qELXwaMAXVN2er7/U9e12emp/QYpZxjNpTemQCjSa+hFKb1CSK8Ezriypt43\nAgMBAAE=\n-----END PUBLIC KEY-----"
}

resource "altr_sidecar_listener" "test_sidecar_listener" {
    sidecar_id = altr_sidecar.test_sidecar.id
    port = 8080
    database_type = "Oracle"
    advertised_version = "19.0.0"
}

data "altr_sidecar" "test_sidecar" {
  id = altr_sidecar.test_sidecar.id
}

data "altr_sidecar_listener" "test_sidecar_listener" {
  sidecar_id = altr_sidecar.test_sidecar.id
  port = altr_sidecar_listener.test_sidecar_listener.port
}