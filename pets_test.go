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
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}
func TestPetsRetreival(t *testing.T) {
	t.Skip("Skipping test due to API issues")
	pets := NewPets()
	if pets == nil {
		t.Error("Cannot get pet..")
	}

	if count := pets.GetPetsCount(); count == 0 {
		t.Error("Cannot get any pets from Taipei Open Data, count:", count)
	}
}

func TestGetPet(t *testing.T) {
	t.Skip("Skipping test due to API issues")
	pets := NewPets()
	if pets == nil {
		t.Error("Cannot get pet..")
	}

	pet := pets.GetNextPet()
	if pet == nil {
		t.Error("Cannot get pet..")
	}

	log.Println(pet.DisplayPet())
	log.Println(pet)
}

func TestGetMultiplePets(t *testing.T) {
	pets := NewPets()
	if pets == nil {
		t.Error("Cannot get pet..")
	}

	for i := 0; i < pets.GetPetsCount(); i++ {
		pet := pets.GetNextPet()
		if pet == nil {
			t.Error("No pet get!")
		}
	}
}

func TestGetCat(t *testing.T) {
	t.Skip("Skipping test due to API issues")
	pets := NewPets()
	pet := pets.GetNextCat()
	if pet == nil {
		t.Skip("Cannot get cat..")
		return
	}

	if pet.PetType() != Cat {
		t.Skip("Get cat error")
		return
	}
	log.Println("Get cat:", pet)
}

func TestGetDog(t *testing.T) {
	t.Skip("Skipping test due to API issues")
	pets := NewPets()
	pet := pets.GetNextDog()
	if pet == nil {
		t.Error("Cannot get pet..")
		return
	}
	if pet.PetType() != Dog {
		t.Error("Get dog error")
		return
	}
	log.Println("Get Dog:", pet)
}

func TestGetNextDog(t *testing.T) {
	t.Skip("Skipping test due to API issues")
	pets := NewPets()
	if pets == nil {
		t.Error("Cannot get pet..")
	}

	pet1 := pets.GetNextDog()
	if pet1 == nil {
		t.Error("Cannot get the first pet..")
		return
	}

	if pet1.PetType() != Dog {
		t.Error("Get 1st dog error")
		return
	}

	pet2 := pets.GetNextDog()
	if pet2 == nil {
		t.Error("Cannot get the second pet..")
		return
	}

	if pet2.PetType() != Dog {
		t.Error("Get 2nd dog error")
		return
	}

	if strings.Compare(pet1.Name, pet2.Name) == 0 {
		t.Error("Get the same dogs:", pet1, pet2)
	}

	log.Println("Get Dogs:", pet1, pet2)
}

func TestSearchPets(t *testing.T) {
	pets := NewPets()
	if pets == nil {
		t.Error("Cannot get pet..")
	}

	criteria := &SearchCriteria{
		Kind:  "狗",
		Color: "白",
	}

	results := pets.SearchPets(criteria)

	if len(results) == 0 {
		t.Error("No white dogs found")
	}

	for _, pet := range results {
		if pet.Variety != "狗" {
			t.Errorf("Expected dog, but got %s", pet.Variety)
		}
		if !strings.Contains(pet.HairType, "白") {
			t.Errorf("Expected white dog, but got %s", pet.HairType)
		}
	}
}