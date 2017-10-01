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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

var ImgSrv string = "https://img-cache-server.herokuapp.com/url?"
var bot *linebot.Client

//PetDB :
var PetDB *Pets

func main() {
	var err error
	PetDB = NewPets()
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func getSecureImageAddress(oriAdd string) string {
	eURL := url.QueryEscape(oriAdd)
	imgGetUrl := fmt.Sprintf("%surl?%s", ImgSrv, eURL)
	log.Println("eURL:", eURL, " url:", imgGetUrl, " ImgApi:", ImgSrv)

	response, err := http.Get(imgGetUrl)
	defer response.Body.Close()

	if err != nil {
		log.Println("Error while downloading:", err)
		return ""
	}

	totalBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error while parsing:", err)
		return ""
	}
	log.Println("Got data:", string(totalBody))
	return fmt.Sprintf("%simgs?%s.jpg", ImgSrv, string(totalBody))

}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				var pet *Pet
				log.Println(message.Text)
				inText := strings.ToLower(message.Text)
				if strings.Contains(inText, "狗") || strings.Contains(inText, "dog") {
					pet = PetDB.GetNextDog()
				} else if strings.Contains(inText, "貓") || strings.Contains(inText, "cat") {
					pet = PetDB.GetNextCat()
				}

				if pet == nil {
					pet = PetDB.GetNextPet()
				}

				imgUrl := getSecureImageAddress(pet.ImageName)
				log.Println("img:", imgUrl)
				out := fmt.Sprintf("您好，目前的動物：名為%s, 所在地為:%s, 敘述: %s 電話為:%s ", pet.Name, pet.Resettlement, pet.Note, pet.Phone)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(out), linebot.NewImageMessage(imgUrl, imgUrl)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
