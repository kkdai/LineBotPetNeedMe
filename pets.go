package main

import (
	"encoding/json"
	"log"
)

//Pets :All pet related API
type Pets struct {
	allPets []Pet
}

//NewPets :
func NewPets() *Pets {
	p := new(Pets)
	p.getPets()
	return p
}

//GetPetByIndex :
func (p *Pets) GetPetByIndex(index int) *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}

	if index > len(p.allPets) || index < 0 {
		index = 0
	}

	return &p.allPets[index]
}

//GetFirstDog :
func (p *Pets) GetFirstDog() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
	}
	return &p.allPets[0]
}

//GetFirstCat :
func (p *Pets) GetFirstCat() *Pet {
	if len(p.allPets) == 0 {
		p.getPets()
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
	var result TaipeiPets
	err = json.Unmarshal(body, &result)

	if err != nil {
		//error
		log.Fatal(err)
	}
	log.Println("All pets is :", len(result.Result.Results))
	// for _, v := range result.Result.Results {
	// 	p.allPets = append(p.allPets, v)
	// }
	p.allPets = result.Result.Results
}
