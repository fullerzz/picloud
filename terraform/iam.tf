resource "aws_iam_user" "picloud" {
  name = "picloud-${var.env}"
}

resource "aws_iam_access_key" "picloud" {
  user = aws_iam_user.picloud.name
}

data "aws_iam_policy_document" "picloud-s3-policy" {
  statement {
    sid       = "AllowS3Access"
    actions   = ["s3:Describe*", "s3:Get*", "s3:List*", "s3:DeleteObject*", "s3:PutObject*", "s3:RestoreObject", "s3:TagResource"]
    resources = [aws_s3_bucket.picloud-bucket.arn, "${aws_s3_bucket.picloud-bucket.arn}/*"]
  }
  statement {
    sid       = "AllowObjectEncryptDecrypt"
    actions   = ["kms:Decrypt", "kms:Encrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey"]
    resources = ["*"]
  }
}

resource "aws_iam_user_policy" "picloud-s3-policy" {
  name   = "picloud-${var.env}-s3-policy"
  user   = aws_iam_user.picloud.name
  policy = data.aws_iam_policy_document.picloud-s3-policy.json
}
