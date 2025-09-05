provider "altr" {
    api_key = var.api_key
    secret  = var.secret
    org_id  = var.org_id
    base_url = var.base_url
}


resource "altr_repo" "test_repo" {
    name = "testrepo"
    hostname = var.oracle_hostname
    port = var.oracle_port
    type = "Oracle" 
    description = ""
}

resource "altr_repo_user" "test_repo_user" {
    repo_name = altr_repo.test_repo.name
    username = "testuser"
    aws_secrets_manager = {
        iam_role = var.iam_role
        secrets_path = var.secrets_path
    }
}

resource "altr_sidecar" "test_repo_sidecar" {
  name  = "test-repo-sidecar"
  description     = "test description"
  hostname     = "localhost"
  unsupported_query_bypass = false
  public_key_1 = "-----BEGIN PUBLIC KEY-----\nMIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQB2VO3t0vpIEKFHbC8c3Ljf\n8b4dSthL7LHXsYtqEvnHWPVbDDJXyp8nVLEdcH6dM7lzOZncfhHHVzREgYgV8LRV\nrMEX4TgkmNFhHk/HasY56BwEU6Vy75L5i34xwkIIGFVNyO/DCJHqFKvCVfTv4tOP\n1Qk7T205Wa53f7lINKQp21BMncl1PjjN1TIqL4IKCZ4AWQB0tybZ6rojODKABzSU\n3ljF4GXkvJUPjfXEfylcarSjukQRurJEhd8vcxZlaVvMWXV2vOy85K/YS2s/BIMa\n9WC6qELXwaMAXVN2er7/U9e12emp/QYpZxjNpTemQCjSa+hFKb1CSK8Ezriypt43\nAgMBAAE=\n-----END PUBLIC KEY-----"
}

resource "altr_sidecar_listener" "test_repo_sidecar_listener" {
    sidecar_id = altr_sidecar.test_repo_sidecar.id
    port = 8080
    database_type = "Oracle"
    advertised_version = "19.0.0"
    depends_on = [ altr_sidecar.test_repo_sidecar ]
}

resource "altr_repo_sidecar_binding" "test_repo_sidecar_binding" {
    repo_name = altr_repo.test_repo.name
    sidecar_id = altr_sidecar.test_repo_sidecar.id
    port = 8080
    depends_on = [ altr_repo.test_repo, altr_sidecar_listener.test_repo_sidecar_listener ]
}

data "altr_repo" "test_repo" {
    name = altr_repo.test_repo.name
}

data "altr_repo_user" "test_repo_user" {
    repo_name = altr_repo.test_repo.name
    username = altr_repo_user.test_repo_user.username
}

data "altr_sidecar" "test_repo_sidecar" {
  id = altr_sidecar.test_repo_sidecar.id
}

data "altr_sidecar_listener" "test_repo_sidecar_listener" {
    sidecar_id = altr_sidecar.test_repo_sidecar.id
    port = altr_sidecar_listener.test_repo_sidecar_listener.port
}

data "altr_repo_sidecar_binding" "test_repo_sidecar_binding" {
    repo_name = altr_repo.test_repo.name
    sidecar_id = altr_sidecar.test_repo_sidecar.id
    port = altr_repo_sidecar_binding.test_repo_sidecar_binding.port
}