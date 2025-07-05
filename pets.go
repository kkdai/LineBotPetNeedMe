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

	for {
		q := p.GetNextPet()
		if q == nil {
			break
		}
		if q.PetType() == Dog {
			return q
		}
	}
	return nil
}

//GetNextCat :
func (p *Pets) GetNextCat() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	for {
		q := p.GetNextPet()
		if q == nil {
			break
		}
		if q.PetType() == Cat {
			return q
		}
	}
	return nil
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
		pt.ID = v.AnimalID
		pt.Name = v.AnimalSubid
		pt.ImageName = v.AlbumFile
		pt.HairType = v.AnimalColour
		pt.Type = v.AnimalBodytype
		pt.Variety = v.AnimalKind
		pt.Resettlement = v.ShelterName + "(" + v.ShelterAddress + ")"
		pt.Phone = v.ShelterTel
		pt.Sex = v.AnimalSex
		pt.Note = v.AnimalRemark
		pt.Age = v.AnimalAge
		p.allPets = append(p.allPets, pt)
	}
}

//SearchPets :
func (p *Pets) SearchPets(criteria *SearchCriteria) []*Pet {
	var result []*Pet
	for i := range p.allPets {
		pet := &p.allPets[i]
		match := true

		if criteria.Kind != "" && pet.Variety != criteria.Kind {
			match = false
		}
		if criteria.Sex != "" && pet.Sex != criteria.Sex {
			match = false
		}
		if criteria.BodyType != "" && pet.Type != criteria.BodyType {
			match = false
		}
		if criteria.Age != "" && pet.Age != criteria.Age {
			match = false
		}

		if match {
			result = append(result, pet)
		}
	}
	return result
}

//GetPet :
func (p *Pets) GetPet(id int) *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	for _, pet := range p.allPets {
		if pet.ID == id {
			return &pet
		}
	}
	return nil
}
