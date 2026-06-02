# TF AWS Module Primitive - CloudWatch Query Definition

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![License: CC BY-NC-ND 4.0](https://img.shields.io/badge/License-CC_BY--NC--ND_4.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-nd/4.0/)

## Overview

This Terraform module creates an [AWS CloudWatch Logs query definition](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_query_definition) for saved Insights queries.

## Pre-Commit Hooks

[.pre-commit-config.yaml](.pre-commit-config.yaml) defines pre-commit hooks for Terraform, Go, and common linting. The `commitlint` hook enforces conventional commit format. The `detect-secrets-hook` prevents new secrets from being introduced. See [pre-commit](https://pre-commit.com/#install) for installation. Install the commit-msg hook manually:

```
pre-commit install --hook-type commit-msg
```

## Usage

See [examples/complete](examples/complete) for a full working example.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.10 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 5.100, < 7.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_query_definition.query_definition](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_query_definition) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_log_group_names"></a> [log\_group\_names](#input\_log\_group\_names) | List of log group names associated with this query definition. | `list(string)` | `null` | no |
| <a name="input_name"></a> [name](#input\_name) | Name of the CloudWatch Logs saved query. | `string` | n/a | yes |
| <a name="input_query_string"></a> [query\_string](#input\_query\_string) | CloudWatch Logs Insights query string to save. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_id"></a> [id](#output\_id) | The query definition ID. |
| <a name="output_log_group_names"></a> [log\_group\_names](#output\_log\_group\_names) | Log group names associated with the query definition. |
| <a name="output_name"></a> [name](#output\_name) | The name of the query definition. |
| <a name="output_query_string"></a> [query\_string](#output\_query\_string) | The saved query string. |
<!-- END_TF_DOCS -->
