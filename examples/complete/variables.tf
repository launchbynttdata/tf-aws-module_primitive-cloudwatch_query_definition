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

variable "resource_names_map" {
  description = "Map of resource names for the resource naming module."
  type = map(object({
    name       = string
    max_length = number
  }))
}

variable "logical_product_family" {
  description = "Logical product family for resource naming."
  type        = string
}

variable "logical_product_service" {
  description = "Logical product service for resource naming."
  type        = string
}

variable "class_env" {
  description = "Class environment for resource naming."
  type        = string
}

variable "instance_env" {
  description = "Instance environment number for resource naming."
  type        = number
}

variable "instance_resource" {
  description = "Instance resource number for resource naming."
  type        = number
}

variable "query_string" {
  description = "CloudWatch Logs Insights query string."
  type        = string
}

variable "tags" {
  description = "Tags applied to supporting resources."
  type        = map(string)
  default     = {}
}
