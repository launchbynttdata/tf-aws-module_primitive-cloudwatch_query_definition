resource_names_map = {
  "kms_key" = {
    name       = "kmskey1"
    max_length = 64
  }
  "log_group" = {
    name       = "queryloggroup1"
    max_length = 64
  }
  "query_definition" = {
    name       = "querydef1"
    max_length = 255
  }
}

logical_product_family  = "launch"
logical_product_service = "cloudwatch"
class_env               = "dev"
instance_env            = 2
instance_resource       = 2

query_string = <<-EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 25
EOF
