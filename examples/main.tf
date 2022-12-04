terraform {
  required_providers {
    twilliate = {
      source  = "twilliate/de/twilliate"
      version = "0.3.2"
    }
    aws = {
      source = "hashicorp/aws"
      version = ">=4.21.0"
    }
  }
}

provider "twilliate" {
  region = "eu-central-1"
}

provider "aws" {
  region = "eu-central-1"
}

#-----------------------------------------------------------------------------------------------------------------------
# S3 BUCKET USED AS AN ORIGIN FOR THE DISTRIBUTION
#-----------------------------------------------------------------------------------------------------------------------

resource "aws_s3_bucket" "origin_bucket" {
  bucket = "twilaw-domain-content"
}

resource "aws_s3_bucket_versioning" "origin_bucket_versioning" {
  bucket = aws_s3_bucket.origin_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "origin_bucket_encryption" {
  bucket = aws_s3_bucket.origin_bucket.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "origin_bucket_access" {
  bucket                  = aws_s3_bucket.origin_bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

module "template_files" {
  source   = "hashicorp/dir/template"
  base_dir = "static"
}

resource "aws_s3_object" "origin_bucket_objects" {
  for_each     = module.template_files.files
  bucket       = aws_s3_bucket.origin_bucket.id
  key          = each.key
  source       = each.value.source_path
  etag         = each.value.digests.md5
  content_type = each.value.content_type
}

#-----------------------------------------------------------------------------------------------------------------------
# ORIGIN ACCESS IDENTITY WITH ACCESS TO THE S3 BUCKET
#-----------------------------------------------------------------------------------------------------------------------

resource "aws_cloudfront_origin_access_identity" "origin_access_identity" {
  comment = "OAI for the dev.twilliate.de impressum"
}

data "aws_iam_policy_document" "origin_bucket_oai_policy" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.origin_bucket.arn}/*"]
    principals {
      type        = "AWS"
      identifiers = [aws_cloudfront_origin_access_identity.origin_access_identity.iam_arn]
    }
  }
}

resource "aws_s3_bucket_policy" "bootstrap_cloudfront_bucket_oai_access" {
  bucket = aws_s3_bucket.origin_bucket.id
  policy = data.aws_iam_policy_document.origin_bucket_oai_policy.json
}


resource "twilliate_cloudfront_origin" "twilaw_cloudfront_origin" {
  distribution_id = "E1WO5WCDX9Q7CD"
  origin_id = "impressum"
  s3_origin_config = {
    origin_access_identity = aws_cloudfront_origin_access_identity.origin_access_identity.id
  }
  origin_domain = aws_s3_bucket.origin_bucket.bucket_regional_domain_name
}

data "aws_cloudfront_cache_policy" "optimized_cache_policy" {
  name = "Managed-CachingOptimized"
}

#resource "twilliate_cloudfront_cache_behaviour" "twilaw_cloudfront_cache_behaviour" {
#  distribution_id = "E1WO5WCDX9Q7CD"
#  origin_id = twilliate_cloudfront_origin.twilaw_cloudfront_origin.origin_id
#  viewer_protocol_policy = "redirect-to-https"
#  path_pattern = "/impressum*"
#  cache_policy_id = data.aws_cloudfront_cache_policy.optimized_cache_policy.id
#}
