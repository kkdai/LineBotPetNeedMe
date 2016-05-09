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
	"net/http"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
var PetDB *Pets

func main() {
	strID := os.Getenv("ChannelID")
	numID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		log.Fatal("Wrong environment setting about ChannelID")
	}

	PetDB = NewPets()
	bot, err = linebot.NewClient(numID, os.Getenv("ChannelSecret"), os.Getenv("MID"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	received, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, result := range received.Results {
		content := result.Content()

		//Log detail receive content
		if content != nil {
			log.Println("RECEIVE Msg:", content.IsMessage, " OP:", content.IsOperation, " type:", content.ContentType, " from:", content.From)
		}
		//Add with new friend.
		if content != nil && content.IsOperation {
			out := fmt.Sprintf("您好，感謝你加入成為好友一起幫助流浪動物找到新的家．輸入任何文字後，會隨機得到一個流浪動物，你可以不斷重複輸入文字然後查看目前所有的流浪動物．")
			_, err = bot.SendText([]string{content.From}, out)
			if err != nil {
				log.Println(err)
			}
			log.Println("New friend add, send cue to new friend.")
		}

		if content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText {

			pet := PetDB.GetNextPet()
			out := fmt.Sprintf("您好，目前的動物：名為%s, 所在地為:%s, 敘述: %s 電話為:%s", pet.Name, pet.Resettlement, pet.Note, pet.Phone)

			// text, err := content.TextContent()
			_, err = bot.SendText([]string{content.From}, out)
			if err != nil {
				log.Println(err)
			}

			_, err = bot.SendImage([]string{content.From}, pet.ImageName, pet.ImageName)
			if err != nil {
				log.Println(err)
			}

		}
	}
}
