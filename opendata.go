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

// 「政府資料開放平臺: 動物認領養」API存取: https://data.gov.tw/dataset/85903
const (
	OpenDataURL string = "https://data.moa.gov.tw/Service/OpenData/TransService.aspx?UnitId=QcbUEzN6E6DL"
)

type TaiwanPets []TaiwanPet

type TaiwanPet struct {
	AnimalID            int         `json:"animal_id"`
	AnimalSubid         string      `json:"animal_subid"`
	AnimalAreaPkid      int         `json:"animal_area_pkid"`
	AnimalShelterPkid   int         `json:"animal_shelter_pkid"`
	AnimalPlace         string      `json:"animal_place"`
	AnimalKind          string      `json:"animal_kind"`
	AnimalSex           string      `json:"animal_sex"`
	AnimalBodytype      string      `json:"animal_bodytype"`
	AnimalColour        string      `json:"animal_colour"`
	AnimalAge           string      `json:"animal_age"`
	AnimalSterilization string      `json:"animal_sterilization"`
	AnimalBacterin      string      `json:"animal_bacterin"`
	AnimalFoundplace    string      `json:"animal_foundplace"`
	AnimalTitle         string      `json:"animal_title"`
	AnimalStatus        string      `json:"animal_status"`
	AnimalRemark        string      `json:"animal_remark"`
	AnimalCaption       string      `json:"animal_caption"`
	AnimalOpendate      string      `json:"animal_opendate"`
	AnimalCloseddate    string      `json:"animal_closeddate"`
	AnimalUpdate        string      `json:"animal_update"`
	AnimalCreatetime    string      `json:"animal_createtime"`
	ShelterName         string      `json:"shelter_name"`
	AlbumName           interface{} `json:"album_name"`
	AlbumFile           string      `json:"album_file"`
	AlbumBase64         interface{} `json:"album_base64"`
	AlbumUpdate         interface{} `json:"album_update"`
	CDate               string      `json:"cDate"`
	ShelterAddress      string      `json:"shelter_address"`
	ShelterTel          string      `json:"shelter_tel"`
}
