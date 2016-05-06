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

package main

// 「臺北市開放認養動物」API存取
const (
	OpenDataURL string = "http://data.taipei/opendata/datalist/apiAccess?scope=resourceAquire&rid=f4a75ba9-7721-4363-884d-c3820b0b917c"
)


//TaipeiPets :Get from  http://data.taipei/opendata/datalist/apiAccess?scope=resourceAquire&rid=f4a75ba9-7721-4363-884d-c3820b0b917c
type TaipeiPets struct {
	Result struct {
		Offset  int    `json:"offset"`
		Limit   int    `json:"limit"`
		Count   int    `json:"count"`
		Sort    string `json:"sort"`
		Results []Pet  `json:"results"`
	} `json:"result"`
}

