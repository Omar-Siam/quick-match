provider "aws" {
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  endpoints {
    dynamodb = "http://localhost:4566"
    es = "http://localhost:4566"
  }
}

resource "aws_dynamodb_table" "user_table" {
  name           = "quickmatch_users"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }

  tags = {
    Name = "QuickMatchUsers"
  }
}

resource "aws_dynamodb_table" "swipes_table" {
  name             = "quickmatch_swipes"
  billing_mode = "PAY_PER_REQUEST"
  hash_key         = "UserID"
  range_key        = "SwipedUserID"

  attribute {
    name = "UserID"
    type = "S"
  }

  attribute {
    name = "SwipedUserID"
    type = "S"
  }

  global_secondary_index {
    name               = "SwipedUserIndex"
    hash_key           = "SwipedUserID"
    range_key          = "UserID"
    projection_type    = "ALL"
  }

  tags = {
    Name = "QuickMatchSwipes"
  }
}

resource "aws_elasticsearch_domain" "discover_domain" {
  domain_name           = "quickmatch-discover"
  elasticsearch_version = "7.9"

  cluster_config {
    instance_type = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 1
  }

  tags = {
    Domain = "QuickMatchDiscover"
  }
}




