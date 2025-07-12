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
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var ImgSrv string
var bot *linebot.Client
var dbClient *db.Client

// PetDB :
var PetDB *Pets

func main() {
	var err error
	ctx := context.Background()

	// Initialize Firebase
	firebaseDBURL := os.Getenv("FIREBASE_DB")
	if firebaseDBURL == "" {
		log.Fatal("FIREBASE_DB environment variable must be set")
	}
	conf := &firebase.Config{
		DatabaseURL: firebaseDBURL,
	}
	// ADC will be used for authentication.
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	dbClient, err = app.Database(ctx)
	if err != nil {
		log.Fatalf("error getting Database client: %v\n", err)
	}

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
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeSm,
			Contents: []linebot.FlexComponent{
				createFavoriteButton(pet),
				createShareButton(pet),
			},
		},
	}

	return linebot.NewFlexMessage("寵物資訊", bubble)
}

func createFavoriteButton(pet *Pet) *linebot.ButtonComponent {
	return &linebot.ButtonComponent{
		Style:  linebot.FlexButtonStyleTypePrimary,
		Action: linebot.NewMessageAction("加入收藏", "favorite "+strconv.Itoa(pet.ID)),
	}
}

func createShareButton(pet *Pet) *linebot.ButtonComponent {
	return &linebot.ButtonComponent{
		Style:  linebot.FlexButtonStyleTypeLink,
		Action: linebot.NewPostbackAction("分享給好友", "action=share&petID="+strconv.Itoa(pet.ID), "", "分享給好友", "", ""),
	}
}

func createDetailRow(title, value string) *linebot.BoxComponent {
	return &linebot.BoxComponent{
		Type:    linebot.FlexComponentTypeBox,
		Layout:  linebot.FlexBoxLayoutTypeBaseline,
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

func addFavorite(ctx context.Context, userID string, pet *Pet) error {
	ref := dbClient.NewRef("/petneedme/favorites/" + userID)
	// Use pet ID as the key to avoid duplicates
	petRef := ref.Child(strconv.Itoa(pet.ID))
	if err := petRef.Set(ctx, pet); err != nil {
		log.Printf("Error adding favorite to Firebase for user %s: %v", userID, err)
		return err
	}
	log.Printf("User %s favorited pet %d", userID, pet.ID)
	return nil
}

func getFavorites(ctx context.Context, userID string) ([]*Pet, error) {
	var favorites map[string]Pet
	ref := dbClient.NewRef("/petneedme/favorites/" + userID)
	if err := ref.Get(ctx, &favorites); err != nil {
		log.Printf("Error getting favorites from Firebase for user %s: %v", userID, err)
		return nil, err
	}

	if favorites == nil {
		return []*Pet{}, nil
	}

	favsSlice := make([]*Pet, 0, len(favorites))
	for _, pet := range favorites {
		p := pet // Create a new variable to avoid pointer issues in loops
		favsSlice = append(favsSlice, &p)
	}
	return favsSlice, nil
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
		ctx := r.Context()
		switch event.Type {
		case linebot.EventTypePostback:
			data := event.Postback.Data
			params, err := url.ParseQuery(data)
			if err != nil {
				log.Printf("Failed to parse postback data: %s", err)
				continue
			}
			action := params.Get("action")
			switch action {
			case "favorite":
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

				if err := addFavorite(ctx, event.Source.UserID, pet); err == nil {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已將寵物加入您的收藏！")).Do(); err != nil {
						log.Print(err)
					}
				}
			case "share":
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
				if len(pet.ImageName) > 0 {
					pet.ImageName = getSecureImageAddress(pet.ImageName)
				}

				shareMessage := linebot.NewTextMessage("請長按並轉傳下方的寵物資訊給您的好友")
				flexMessage := newPetFlexMessage(pet)

				if _, err := bot.ReplyMessage(event.ReplyToken, shareMessage, flexMessage).Do(); err != nil {
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

					if err := addFavorite(ctx, event.Source.UserID, pet); err == nil {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已將寵物加入您的收藏！")).Do(); err != nil {
							log.Print(err)
						}
					}
					return
				}

				if strings.Contains(inText, "狗") || strings.Contains(inText, "dog") {
					pet = PetDB.GetNextDog()
				} else if strings.Contains(inText, "貓") || strings.Contains(inText, "cat") {
					pet = PetDB.GetNextCat()
				} else if strings.Contains(inText, "收藏") {
					favs, err := getFavorites(ctx, event.Source.UserID)
					if err != nil {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("抱歉，讀取收藏清單時發生錯誤。")).Do(); err != nil {
							log.Print(err)
						}
						return
					}

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