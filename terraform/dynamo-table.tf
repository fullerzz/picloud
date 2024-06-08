resource "aws_dynamodb_table" "picloud-file-metadata-table" {
  name         = "picloud-file-metadata-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "file_name"
  range_key    = "file_sha256"

  attribute {
    name = "file_name"
    type = "S"
  }

  attribute {
    name = "file_sha256"
    type = "S"
  }

  attribute {
    name = "file_extension"
    type = "S"
  }

  attribute {
    name = "upload_timestamp"
    type = "N"
  }


  global_secondary_index {
    name            = "FileExtensionIndex"
    hash_key        = "file_extension"
    range_key       = "file_sha256"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "UploadTimestampSortIndex"
    hash_key        = "file_name"
    range_key       = "upload_timestamp"
    projection_type = "ALL"
  }
}
