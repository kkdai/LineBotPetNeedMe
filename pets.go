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
	"encoding/json"
	"fmt"
	"log"
)

//Pets :All pet related API
type Pets struct {
	allPets    []Pet
	queryIndex int
}

//NewPets :
func NewPets() *Pets {
	p := new(Pets)
	p.getPets()
	return p
}

//GetNextPet :
func (p *Pets) GetNextPet() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	retPet := &p.allPets[p.getNextIndex()]
	return retPet
}

//GetNextDog :
func (p *Pets) GetNextDog() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	var retPet *Pet
	for _, p := range p.allPets {
		if p.PetType() == Dog {
			retPet = &p
			break
		}
	}
	return retPet
}

//GetNextCat :
func (p *Pets) GetNextCat() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	var retPet *Pet
	for _, p := range p.allPets {
		if p.PetType() == Cat {
			retPet = &p
			break
		}
	}
	return retPet
}

//GetPetsCount :
func (p *Pets) GetPetsCount() int {
	return len(p.allPets)
}

func (p *Pets) getPets() {
	c := NewClient(OpenDataURL)
	body, err := c.GetHttpRes()
	if err != nil {
		return
	}

	// log.Println("ret:", string(body))
	var results TaiwanPets
	err = json.Unmarshal(body, &results)

	if err != nil {
		//error
		log.Fatal(err)
	}
	log.Println("All pets is :", len(results))
	p.LoadPets(results)
}

func (p *Pets) getNextIndex() int {
	if p.queryIndex >= len(p.allPets) {
		p.queryIndex = 0
	}

	retInt := p.queryIndex
	p.queryIndex++
	return retInt
}

func (p *Pets) LoadPets(pets TaiwanPets) {
	//Mapping
	for _, v := range pets {
		pt := Pet{}
		pt.Name = v.AnimalSubid
		pt.ImageName = v.AlbumFile
		pt.Type = fmt.Sprintf("%s (%s)", v.AnimalKind, v.AnimalColour)
		pt.Resettlement = v.ShelterName + "(" + v.ShelterAddress + ")"
		pt.Phone = v.ShelterTel
		pt.Sex = v.AnimalSex
		pt.Note = v.AnimalRemark
		p.allPets = append(p.allPets, pt)
	}
}
