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

import (
	"fmt"
	"log"

	"golang.org/x/text/encoding/unicode/utf32"
)

//PetType :
type PetType int

const (
	//Dog :
	Dog PetType = iota
	//Cat :
	Cat
)

//Pet :
type Pet struct {
	ID              string `json:"_id"`
	Name            string `json:"Name"`
	Sex             string `json:"Sex"`
	Type            string `json:"Type"`
	Build           string `json:"Build"`
	Age             string `json:"Age"`
	Variety         string `json:"Variety"`
	Reason          string `json:"Reason"`
	AcceptNum       string `json:"AcceptNum"`
	ChipNum         string `json:"ChipNum"`
	IsSterilization string `json:"IsSterilization"`
	HairType        string `json:"HairType"`
	Note            string `json:"Note"`
	Resettlement    string `json:"Resettlement"`
	Phone           string `json:"Phone"`
	Email           string `json:"Email"`
	ChildreAnlong   string `json:"ChildreAnlong"`
	AnimalAnlong    string `json:"AnimalAnlong"`
	Bodyweight      string `json:"Bodyweight"`
	ImageName       string `json:"ImageName"`
}

//PetType :
func (p *Pet) PetType() PetType {
	var retType PetType
	switch p.Type {
	case "犬":
		retType = Dog
	case "貓":
		retType = Cat
	}

	return retType
}

//DisplayPet : Display single pet on chatbot
func (p *Pet) DisplayPet() string {
	utf32BEIB := utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM)
	dec := utf32BEIB.NewDecoder()
	brownEmoji, err := dec.String("\x00\x10\x00\x84")
	if err != nil {
		log.Print(err)
	}

	return fmt.Sprintf("%s 您好，目前的動物品種是 %s 名為%s, 所在地為:%s, 電話為:%s  敘述為:%s", brownEmoji, p.Type+"("+p.Sex+")", p.Name, p.Resettlement, p.Phone, p.Note)
}
