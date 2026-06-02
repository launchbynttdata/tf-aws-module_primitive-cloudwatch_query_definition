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

# -----------------------------------------------------------------------------
# Required
# -----------------------------------------------------------------------------

variable "name" {
  description = "Name of the CloudWatch Logs saved query."
  type        = string

  validation {
    condition     = length(var.name) >= 1 && length(var.name) <= 255
    error_message = "Query definition name must be between 1 and 255 characters."
  }
}

variable "query_string" {
  description = "CloudWatch Logs Insights query string to save."
  type        = string

  validation {
    condition     = length(var.query_string) >= 1
    error_message = "Query string must not be empty."
  }
}

# -----------------------------------------------------------------------------
# Optional
# -----------------------------------------------------------------------------

variable "log_group_names" {
  description = "List of log group names associated with this query definition."
  type        = list(string)
  default     = null
}
