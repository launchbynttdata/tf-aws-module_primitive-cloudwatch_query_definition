// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

output "region" {
  description = "The AWS region where resources are deployed."
  value       = data.aws_region.current.region
}

output "log_group_name" {
  description = "The example log group name."
  value       = aws_cloudwatch_log_group.example.name
}

output "kms_key_arn" {
  description = "The KMS key ARN used for log group encryption."
  value       = aws_kms_key.logs.arn
}

output "id" {
  description = "The query definition ID."
  value       = module.query_definition.id
}

output "name" {
  description = "The query definition name."
  value       = module.query_definition.name
}

output "query_string" {
  description = "The saved query string."
  value       = module.query_definition.query_string
}

output "log_group_names" {
  description = "Log group names associated with the query definition."
  value       = module.query_definition.log_group_names
}
