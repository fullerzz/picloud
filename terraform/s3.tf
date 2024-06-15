resource "aws_s3_bucket" "picloud-bucket" {
  bucket_prefix = "picloud-bucket"

  tags = {
    name        = "picloud-bucket"
    environment = var.env
  }
}

resource "aws_s3_bucket_ownership_controls" "picloud-bucket" {
  bucket = aws_s3_bucket.picloud-bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "picloud-bucket" {
  depends_on = [aws_s3_bucket_ownership_controls.picloud-bucket]

  bucket = aws_s3_bucket.picloud-bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "picloud-bucket-logs" {
  bucket = "${aws_s3_bucket.picloud-bucket.id}-logs"
}

resource "aws_s3_bucket_ownership_controls" "picloud-bucket-logs" {
  bucket = aws_s3_bucket.picloud-bucket-logs.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "picloud-bucket-logs-acl" {
  depends_on = [aws_s3_bucket_ownership_controls.picloud-bucket-logs]
  bucket     = aws_s3_bucket.picloud-bucket-logs.id
  acl        = "log-delivery-write"
}

resource "aws_s3_bucket_logging" "picloud-bucket-logging" {
  bucket = aws_s3_bucket.picloud-bucket.id

  target_bucket = aws_s3_bucket.picloud-bucket-logs.id
  target_prefix = "log/"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "picloud-bucket" {
  bucket = aws_s3_bucket.picloud-bucket.id
  rule {
    bucket_key_enabled = true
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "picloud-bucket" {
  bucket = aws_s3_bucket.picloud-bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
