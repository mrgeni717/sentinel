# Generated once per apply, baked into the EC2 instance's user data
# script. Same trade-off noted as in bank-demo: fine for a short-lived
# demo, not how a real production deployment would handle secrets
# (Secrets Manager / SSM Parameter Store would be the real approach).

resource "random_password" "ingest_key" {
  length  = 32
  special = false
}
