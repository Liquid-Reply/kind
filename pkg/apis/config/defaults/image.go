/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package defaults contains cross-api-version configuration defaults
package defaults

// Image is the default for the Config.Image field, aka the default node image.
const Image = "kindest/node:v1.22.2@sha256:9af3ab3e36fb59890b2fb3a18000930b4792d62d10d2060e0ca701e2c392d487"
