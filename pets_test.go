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
	"log"
	"testing"
)

func TestPetsRetreival(t *testing.T) {
	pets := NewPets()
	if count := pets.GetPetsCount(); count == 0 {
		t.Error("Cannot get any pets from Taipei Open Data, count:", count)
	}
}

func TestGetPet(t *testing.T) {
	pets := NewPets()
	pet := pets.GetNextPet()
	if pet == nil {
		t.Error("Cannot get pet..")
	}

	log.Println(pet)
}

func TestGetCat(t *testing.T) {
	pets := NewPets()
	pet := pets.GetNextCat()
	if pet == nil {
		t.Error("Cannot get pet..")
		return
	}

	if pet.PetType() != Cat {
		t.Error("Get cat error")
		return
	}
	log.Println("Get cat:", pet)
}

func TestGetDog(t *testing.T) {
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
