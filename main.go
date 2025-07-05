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
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var ImgSrv string
var bot *linebot.Client

// PetDB :
var PetDB *Pets
var userFavorites = make(map[string][]*Pet)
var favoritesMutex = &sync.Mutex{}

func main() {
	var err error
	ImgSrv = os.Getenv("IMG_SRV")

	PetDB = NewPets()
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func newPetFlexMessage(pet *Pet) *linebot.FlexMessage {
	// Share button
	shareURI := fmt.Sprintf("line://msg/text/?%s", url.QueryEscape(pet.DisplayPet()))
	shareButton := &linebot.ButtonComponent{
		Style:  linebot.FlexButtonStyleTypeLink,
		Action: linebot.NewURIAction("分享給好友", shareURI),
	}

	// Favorite button
	favoriteButton := &linebot.ButtonComponent{
		Style:  linebot.FlexButtonStyleTypePrimary,
				Action: linebot.NewMessageAction("加入收藏", "favorite "+strconv.Itoa(pet.ID)),
	}

	bubble := &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         pet.ImageName,
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType20to13,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   pet.Name,
					Weight: linebot.FlexTextWeightTypeBold,
					Size:   linebot.FlexTextSizeTypeXl,
				},
				&linebot.BoxComponent{
					Type:   linebot.FlexComponentTypeBox,
					Layout: linebot.FlexBoxLayoutTypeVertical,
					Margin: linebot.FlexComponentMarginTypeLg,
					Contents: []linebot.FlexComponent{
						createDetailRow("種類", pet.Variety),
						createDetailRow("性別", pet.Sex),
						createDetailRow("體型", pet.Type),
						createDetailRow("毛色", pet.HairType),
						createDetailRow("年紀", pet.Age),
						createDetailRow("收容所", pet.Resettlement),
						createDetailRow("聯絡電話", pet.Phone),
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeSm,
			Contents: []linebot.FlexComponent{
				favoriteButton,
				shareButton,
				&linebot.ButtonComponent{
					Style:  linebot.FlexButtonStyleTypeLink,
					Action: linebot.NewURIAction("聯絡我", "tel:"+pet.Phone),
				},
			},
		},
	}

	return linebot.NewFlexMessage("寵物資���", bubble)
}

func createDetailRow(title, value string) *linebot.BoxComponent {
	return &linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeBaseline,
		Spacing: linebot.FlexComponentSpacingTypeSm,
		Contents: []linebot.FlexComponent{
			&linebot.TextComponent{
				Type:  linebot.FlexComponentTypeText,
				Text:  title,
				Color: "#aaaaaa",
				Size:  linebot.FlexTextSizeTypeSm,
				Flex:  linebot.IntPtr(2),
			},
			&linebot.TextComponent{
				Type:  linebot.FlexComponentTypeText,
				Text:  value,
				Wrap:  true,
				Color: "#666666",
				Size:  linebot.FlexTextSizeTypeSm,
				Flex:  linebot.IntPtr(5),
			},
		},
	}
}

func getSecureImageAddress(oriAdd string) string {
	eURL := url.QueryEscape(oriAdd)
	imgGetUrl := fmt.Sprintf("%surl?%s", ImgSrv, eURL)
	log.Println("eURL:", eURL, " url:", imgGetUrl, " ImgApi:", ImgSrv)

	response, err := http.Get(imgGetUrl)
	if err != nil {
		log.Println("Error while downloading:", err)
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return ""
	}

	totalBody, err := io.ReadAll(response.Body)
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
		switch event.Type {
		case linebot.EventTypePostback:
			data := event.Postback.Data
			params, err := url.ParseQuery(data)
			if err != nil {
				log.Printf("Failed to parse postback data: %s", err)
				continue
			}
			action := params.Get("action")
			if action == "favorite" {
				petIDStr := params.Get("petID")
				petID, err := strconv.Atoi(petIDStr)
				if err != nil {
					log.Printf("Invalid pet ID: %s", petIDStr)
					continue
				}

				pet := PetDB.GetPet(petID)
				if pet == nil {
					log.Printf("Pet with ID %d not found", petID)
					continue
				}

				favoritesMutex.Lock()
				userFavorites[event.Source.UserID] = append(userFavorites[event.Source.UserID], pet)
				favoritesMutex.Unlock()

				log.Printf("User %s favorited pet %d", event.Source.UserID, petID)
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已將寵物加入您的收藏！")).Do(); err != nil {
					log.Print(err)
				}
			}

		case linebot.EventTypeUnsend:
			log.Println("Unsend")
			target := ""
			if event.Source.GroupID != "" {
				target = event.Source.GroupID
			} else {
				target = event.Source.RoomID
			}
			if _, err = bot.PushMessage(target, linebot.NewTextMessage("不要害羞地回收訊息，趕快打狗或是貓來看流浪動物。")).Do(); err != nil {
				log.Print(err)
			}

		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				var pet *Pet
				log.Println(message.Text)
				inText := strings.ToLower(message.Text)

				// First, try to parse with Gemini
				criteria, err := ParseSearchCriteriaFromQuery(inText)
				if err != nil {
					log.Printf("Gemini parsing error: %v", err)
				}

				if criteria != nil {
					log.Printf("Gemini parsed criteria: %+v", criteria)
					results := PetDB.SearchPets(criteria)
					if len(results) == 0 {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("很抱歉，目前沒有找到符合條件的寵物。")).Do(); err != nil {
							log.Print(err)
						}
						return
					}

					var bubbles []*linebot.BubbleContainer
					for _, p := range results {
						if len(p.ImageName) > 0 {
							p.ImageName = getSecureImageAddress(p.ImageName)
						}
						bubbles = append(bubbles, newPetFlexMessage(p).Contents.(*linebot.BubbleContainer))
						if len(bubbles) == 10 { // Carousel limit is 10
							break
						}
					}

					carousel := &linebot.CarouselContainer{
						Type:     linebot.FlexContainerTypeCarousel,
						Contents: bubbles,
					}

					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage("為您找到這些寵物", carousel)).Do(); err != nil {
						log.Print(err)
					}
					return
				}

				// Fallback to original logic
				if strings.HasPrefix(inText, "favorite") {
					petIDStr := strings.TrimSpace(strings.TrimPrefix(inText, "favorite"))
					petID, err := strconv.Atoi(petIDStr)
					if err != nil {
						log.Printf("Invalid pet ID: %s", petIDStr)
						return
					}

					pet := PetDB.GetPet(petID)
					if pet == nil {
						log.Printf("Pet with ID %d not found", petID)
						return
					}

					favoritesMutex.Lock()
					userFavorites[event.Source.UserID] = append(userFavorites[event.Source.UserID], pet)
					favoritesMutex.Unlock()

					log.Printf("User %s favorited pet %d", event.Source.UserID, petID)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已將寵物加入您的收藏！")).Do(); err != nil {
						log.Print(err)
					}
					return
				}

				if strings.Contains(inText, "狗") || strings.Contains(inText, "dog") {
					pet = PetDB.GetNextDog()
				} else if strings.Contains(inText, "貓") || strings.Contains(inText, "cat") {
					pet = PetDB.GetNextCat()
				} else if strings.Contains(inText, "收藏") {
					favoritesMutex.Lock()
					favs := userFavorites[event.Source.UserID]
					favoritesMutex.Unlock()

					if len(favs) == 0 {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("您尚未收藏任何寵物。")).Do(); err != nil {
							log.Print(err)
						}
						return
					}

					var bubbles []*linebot.BubbleContainer
					for _, p := range favs {
						bubbles = append(bubbles, newPetFlexMessage(p).Contents.(*linebot.BubbleContainer))
					}

					carousel := &linebot.CarouselContainer{
						Type:     linebot.FlexContainerTypeCarousel,
						Contents: bubbles,
					}

					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage("您的收藏清單", carousel)).Do(); err != nil {
						log.Print(err)
					}
					return
				}

				if pet == nil {
					pet = PetDB.GetNextPet()
				}

				if pet != nil && len(pet.ImageName) > 0 {
					pet.ImageName = getSecureImageAddress(pet.ImageName)
					flexMessage := newPetFlexMessage(pet)
					if _, err := bot.ReplyMessage(event.ReplyToken, flexMessage).Do(); err != nil {
						log.Print(err)
					}
				} else if pet != nil {
					out := pet.DisplayPet()
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(out)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}