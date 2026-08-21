# Complete Example

This example creates a customer-managed KMS key, a CloudWatch log group encrypted with that key, and a saved Logs Insights query definition scoped to that log group.

## Usage

```hcl
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

module "resource_names" {
  source  = "terraform.registry.launch.nttdata.com/module_library/resource_name/launch"
  version = "~> 2.0"

  for_each = var.resource_names_map

  logical_product_family  = var.logical_product_family
  logical_product_service = var.logical_product_service
  class_env               = var.class_env
  instance_env            = var.instance_env
  instance_resource       = var.instance_resource
  cloud_resource_type     = each.value.name
  maximum_length          = each.value.max_length

  region = join("", split("-", data.aws_region.current.region))
}

data "aws_iam_policy_document" "logs_kms" {
  statement {
    sid    = "EnableIAMUserPermissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }

  statement {
    sid    = "AllowCloudWatchLogs"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["logs.${data.aws_region.current.region}.amazonaws.com"]
    }
    actions = [
      "kms:Encrypt*",
      "kms:Decrypt*",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:Describe*",
      "kms:CreateGrant"
    ]
    resources = ["*"]
    condition {
      test     = "ArnLike"
      variable = "kms:EncryptionContext:aws:logs:arn"
      values   = ["arn:aws:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:*"]
    }
  }
}

resource "aws_kms_key" "logs" {
  description             = "KMS key for CloudWatch Logs encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.logs_kms.json

  tags = merge(var.tags, { Name = module.resource_names["kms_key"].standard })
}

resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/example/${module.resource_names["log_group"].standard}"
  retention_in_days = 1
  kms_key_id        = aws_kms_key.logs.arn

  tags = var.tags
}

module "query_definition" {
  source = "../.."

  name            = module.resource_names["query_definition"].minimal_random_suffix
  query_string    = var.query_string
  log_group_names = coalesce(var.log_group_names, [aws_cloudwatch_log_group.example.name])

  depends_on = [aws_cloudwatch_log_group.example]
}
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.10 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 5.100, < 7.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 6.61.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_resource_names"></a> [resource\_names](#module\_resource\_names) | terraform.registry.launch.nttdata.com/module_library/resource_name/launch | ~> 2.0 |
| <a name="module_query_definition"></a> [query\_definition](#module\_query\_definition) | ../.. | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_log_group.example](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_kms_key.logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_iam_policy_document.logs_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_resource_names_map"></a> [resource\_names\_map](#input\_resource\_names\_map) | Map of resource names for the resource naming module. | <pre>map(object({<br/>    name       = string<br/>    max_length = number<br/>  }))</pre> | n/a | yes |
| <a name="input_logical_product_family"></a> [logical\_product\_family](#input\_logical\_product\_family) | Logical product family for resource naming. | `string` | n/a | yes |
| <a name="input_logical_product_service"></a> [logical\_product\_service](#input\_logical\_product\_service) | Logical product service for resource naming. | `string` | n/a | yes |
| <a name="input_class_env"></a> [class\_env](#input\_class\_env) | Class environment for resource naming. | `string` | n/a | yes |
| <a name="input_instance_env"></a> [instance\_env](#input\_instance\_env) | Instance environment number for resource naming. | `number` | n/a | yes |
| <a name="input_instance_resource"></a> [instance\_resource](#input\_instance\_resource) | Instance resource number for resource naming. | `number` | n/a | yes |
| <a name="input_query_string"></a> [query\_string](#input\_query\_string) | CloudWatch Logs Insights query string. | `string` | n/a | yes |
| <a name="input_log_group_names"></a> [log\_group\_names](#input\_log\_group\_names) | Log group names associated with the query definition. Defaults to the example log group when null. | `list(string)` | `null` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to supporting resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_region"></a> [region](#output\_region) | The AWS region where resources are deployed. |
| <a name="output_log_group_name"></a> [log\_group\_name](#output\_log\_group\_name) | The example log group name. |
| <a name="output_kms_key_arn"></a> [kms\_key\_arn](#output\_kms\_key\_arn) | The KMS key ARN used for log group encryption. |
| <a name="output_id"></a> [id](#output\_id) | The query definition ID. |
| <a name="output_name"></a> [name](#output\_name) | The query definition name. |
| <a name="output_query_string"></a> [query\_string](#output\_query\_string) | The saved query string. |
| <a name="output_log_group_names"></a> [log\_group\_names](#output\_log\_group\_names) | Log group names associated with the query definition. |
<!-- END_TF_DOCS -->
