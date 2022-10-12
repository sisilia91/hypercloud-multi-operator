/*
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

package hyperAuth

import (
	"net/http"
)

func IsOK(check int) bool {
	SuccessStatusList := map[int]bool{
		http.StatusOK:             true,
		http.StatusCreated:        true,
		http.StatusNoContent:      true,
		http.StatusPartialContent: true,
		http.StatusConflict:       true,
		// http.StatusContinue:       true,
	}
	_, ok := SuccessStatusList[check]
	return ok
}

// func GetHyperAuthTLSCertificate(hyperauthHttpsSecret *coreV1.Secret) string {
// 	hyperauthTlsCert := strings.TrimSpace(string(hyperauthHttpsSecret.Data["tls.crt"]))
// 	hyperauthTlsCert = strings.ReplaceAll(hyperauthTlsCert, "\n", "\\n")
// 	hyperauthTlsCert = strings.ReplaceAll(hyperauthTlsCert, "\t", "\\t")
// 	return hyperauthTlsCert
// }
