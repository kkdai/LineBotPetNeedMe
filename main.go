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

// Global variables for services
var (
	ImgSrv   string
	bot      *linebot.Client
	dbClient *db.Client
	PetDB    *Pets
)

// main is the entry point of the application.
func main() {
	var err error
	ctx := context.Background()

	// Initialize services
	if err = initializeFirebase(ctx); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	initializeImgSrv()
	if err = initializeLineBot(); err != nil {
		log.Fatalf("Failed to initialize LINE Bot: %v", err)
	}

	PetDB = NewPets()

	// Setup HTTP server
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// --- Initializers ---

func initializeFirebase(ctx context.Context) error {
	firebaseDBURL := os.Getenv("FIREBASE_DB")
	if firebaseDBURL == "" {
		return fmt.Errorf("FIREBASE_DB environment variable must be set")
	}
	conf := &firebase.Config{DatabaseURL: firebaseDBURL}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return fmt.Errorf("error initializing app: %w", err)
	}
	dbClient, err = app.Database(ctx)
	if err != nil {
		return fmt.Errorf("error getting Database client: %w", err)
	}
	return nil
}

func initializeLineBot() error {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	if err != nil {
		return err
	}
	log.Println("Bot initialized successfully")
	return nil
}

func initializeImgSrv() {
	ImgSrv = os.Getenv("IMG_SRV")
}

// --- HTTP Handlers ---

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		ctx := r.Context()
		if err := dispatchEvent(ctx, event); err != nil {
			log.Printf("Error dispatching event: %v", err)
		}
	}
}

// --- Event Dispatcher ---

func dispatchEvent(ctx context.Context, event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		return handleMessageEvent(ctx, event)
	case linebot.EventTypePostback:
		return handlePostbackEvent(ctx, event)
	case linebot.EventTypeUnsend:
		return handleUnsendEvent(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
		return nil
	}
}

// --- Event Handlers ---

func handleMessageEvent(ctx context.Context, event *linebot.Event) error {
	msg, ok := event.Message.(*linebot.TextMessage)
	if !ok {
		return nil // Not a text message
	}

	inText := strings.ToLower(strings.TrimSpace(msg.Text))
	log.Printf("Received message from %s: %s", event.Source.UserID, inText)

	// 1. Try Gemini AI Search
	criteria, err := ParseSearchCriteriaFromQuery(inText)
	if err != nil {
		log.Printf("Gemini parsing error: %v", err)
	}
	if criteria != nil {
		log.Printf("Gemini parsed criteria: %+v", criteria)
		pets := PetDB.SearchPets(criteria)
		return replyWithPetCarousel(event.ReplyToken, pets, "為您找到這些寵物")
	}

	// 2. Handle Text Commands
	if handled := handleCommand(ctx, event.ReplyToken, event.Source.UserID, inText); handled {
		return nil
	}

	// 3. Default: Get a random pet
	pet := PetDB.GetNextPet()
	return replyWithSinglePet(event.ReplyToken, pet)
}

func handlePostbackEvent(ctx context.Context, event *linebot.Event) error {
	data := event.Postback.Data
	params, err := url.ParseQuery(data)
	if err != nil {
		return fmt.Errorf("failed to parse postback data: %w", err)
	}

	action := params.Get("action")
	if action == "favorite" {
		petIDStr := params.Get("petID")
		return handleAddFavorite(ctx, event.ReplyToken, event.Source.UserID, petIDStr)
	}

	return nil
}

func handleUnsendEvent(event *linebot.Event) error {
	log.Printf("Unsend event from: %+v", event.Source)
	target := event.Source.GroupID
	if target == "" {
		target = event.Source.RoomID
	}
	// Note: Pushing to a user ID from an unsend event is not directly supported.
	// This will only work in groups/rooms where the bot remains.
	if target != "" {
		if _, err := bot.PushMessage(target, linebot.NewTextMessage("不要害羞地回收訊息，趕快打狗或是貓來看流浪動物。")).Do(); err != nil {
			return err
		}
	}
	return nil
}

// --- Command Handler ---

func handleCommand(ctx context.Context, replyToken, userID, text string) bool {
	switch {
	case strings.HasPrefix(text, "favorite"):
		petIDStr := strings.TrimSpace(strings.TrimPrefix(text, "favorite"))
		if err := handleAddFavorite(ctx, replyToken, userID, petIDStr); err != nil {
			log.Printf("Error handling favorite command: %v", err)
		}
		return true
	case text == "狗" || text == "dog":
		return replyWithSinglePet(replyToken, PetDB.GetNextDog()) == nil
	case text == "貓" || text == "cat":
		return replyWithSinglePet(replyToken, PetDB.GetNextCat()) == nil
	case text == "收藏":
		if err := handleShowFavorites(ctx, replyToken, userID); err != nil {
			log.Printf("Error handling show favorites command: %v", err)
		}
		return true
	}
	return false
}

// --- Action Handlers ---

func handleAddFavorite(ctx context.Context, replyToken, userID, petIDStr string) error {
	petID, err := strconv.Atoi(petIDStr)
	if err != nil {
		return fmt.Errorf("invalid pet ID: %s", petIDStr)
	}
	pet := PetDB.GetPet(petID)
	if pet == nil {
		return fmt.Errorf("pet with ID %d not found", petID)
	}

	if err := addFavorite(ctx, userID, pet); err != nil {
		return replyWithError(replyToken, "加入收藏失敗，請稍後再試。")
	}

	_, err = bot.ReplyMessage(replyToken, linebot.NewTextMessage("已將寵物加入您的收藏！")).Do()
	return err
}

func handleShowFavorites(ctx context.Context, replyToken, userID string) error {
	favs, err := getFavorites(ctx, userID)
	if err != nil {
		return replyWithError(replyToken, "抱歉，讀取收藏清單時發生錯誤。")
	}
	return replyWithPetCarousel(replyToken, favs, "您的收藏清單")
}

// --- Reply Helpers ---

func replyWithSinglePet(replyToken string, pet *Pet) error {
	if pet == nil {
		_, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("抱歉，目前沒有找到寵物。")).Do()
		return err
	}
	if len(pet.ImageName) > 0 {
		pet.ImageName = getSecureImageAddress(pet.ImageName)
	}
	flexMessage := newPetFlexMessage(pet)
	_, err := bot.ReplyMessage(replyToken, flexMessage).Do()
	return err
}

func replyWithPetCarousel(replyToken string, pets []*Pet, title string) error {
	if len(pets) == 0 {
		_, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("很抱歉，目前沒有找到符合條件的寵物。")).Do()
		return err
	}

	var bubbles []*linebot.BubbleContainer
	for i, p := range pets {
		if i >= 10 { // Carousel limit is 10
			break
		}
		if len(p.ImageName) > 0 {
			p.ImageName = getSecureImageAddress(p.ImageName)
		}
		bubbles = append(bubbles, newPetFlexMessage(p).Contents.(*linebot.BubbleContainer))
	}

	carousel := &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbles,
	}

	_, err := bot.ReplyMessage(replyToken, linebot.NewFlexMessage(title, carousel)).Do()
	return err
}

func replyWithError(replyToken, message string) error {
	_, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
	return err
}

// --- Firebase Operations ---

func addFavorite(ctx context.Context, userID string, pet *Pet) error {
	ref := dbClient.NewRef("/petneedme/favorites/" + userID)
	petRef := ref.Child(strconv.Itoa(pet.ID)) // Use pet ID as the key to avoid duplicates
	if err := petRef.Set(ctx, pet); err != nil {
		return fmt.Errorf("error adding favorite to Firebase for user %s: %w", userID, err)
	}
	log.Printf("User %s favorited pet %d", userID, pet.ID)
	return nil
}

func getFavorites(ctx context.Context, userID string) ([]*Pet, error) {
	var favorites map[string]Pet
	ref := dbClient.NewRef("/petneedme/favorites/" + userID)
	if err := ref.Get(ctx, &favorites); err != nil {
		return nil, fmt.Errorf("error getting favorites from Firebase for user %s: %w", userID, err)
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

// --- Flex Message Builders ---

func newPetFlexMessage(pet *Pet) *linebot.FlexMessage {
	if pet.ImageName == "" {
		// Use a placeholder image if none is available
		pet.ImageName = "https://petneed.me/static/img/petNeedme_full_color.png"
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
				&linebot.TextComponent{Type: linebot.FlexComponentTypeText, Text: pet.Name, Weight: linebot.FlexTextWeightTypeBold, Size: linebot.FlexTextSizeTypeXl},
				&linebot.BoxComponent{
					Type:     linebot.FlexComponentTypeBox,
					Layout:   linebot.FlexBoxLayoutTypeVertical,
					Margin:   linebot.FlexComponentMarginTypeLg,
					Spacing:  linebot.FlexComponentSpacingTypeSm,
					Contents: createDetailRows(pet),
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

func createDetailRows(pet *Pet) []linebot.FlexComponent {
	return []linebot.FlexComponent{
		createDetailRow("種類", pet.Variety),
		createDetailRow("性別", pet.Sex),
		createDetailRow("體型", pet.Type),
		createDetailRow("毛色", pet.HairType),
		createDetailRow("年紀", pet.Age),
		createDetailRow("收容所", pet.Resettlement),
		createDetailRow("聯絡電話", pet.Phone),
	}
}

func createDetailRow(title, value string) *linebot.BoxComponent {
	if value == "" {
		value = "不詳"
	}
	return &linebot.BoxComponent{
		Type:    linebot.FlexComponentTypeBox,
		Layout:  linebot.FlexBoxLayoutTypeBaseline,
		Spacing: linebot.FlexComponentSpacingTypeSm,
		Contents: []linebot.FlexComponent{
			&linebot.TextComponent{Type: linebot.FlexComponentTypeText, Text: title, Color: "#aaaaaa", Size: linebot.FlexTextSizeTypeSm, Flex: linebot.IntPtr(2)},
			&linebot.TextComponent{Type: linebot.FlexComponentTypeText, Text: value, Wrap: true, Color: "#666666", Size: linebot.FlexTextSizeTypeSm, Flex: linebot.IntPtr(5)},
		},
	}
}

func createFavoriteButton(pet *Pet) *linebot.ButtonComponent {
	return &linebot.ButtonComponent{
		Type:   linebot.FlexComponentTypeButton,
		Style:  linebot.FlexButtonStyleTypePrimary,
		Action: linebot.NewMessageAction("加入收藏", "favorite "+strconv.Itoa(pet.ID)),
	}
}

func createShareButton(pet *Pet) *linebot.ButtonComponent {
	shareText := generateShareText(pet)
	shareURI := fmt.Sprintf("line://msg/text/?%s", url.QueryEscape(shareText))
	return &linebot.ButtonComponent{
		Type:   linebot.FlexComponentTypeButton,
		Style:  linebot.FlexButtonStyleTypeLink,
		Action: linebot.NewURIAction("分享給好友", shareURI),
	}
}

func generateShareText(pet *Pet) string {
	return fmt.Sprintf(
		"我想跟你分享一個可愛的寵物！\n\n"+
			"名字：%s\n"+
			"種類：%s\n"+
			"性別：%s\n"+
			"體型：%s\n"+
			"年紀：%s\n"+
			"收容所：%s\n"+
			"聯絡電話：%s\n\n"+
			"看看牠的照片吧：%s",
		pet.Name, pet.Variety, pet.Sex, pet.Type, pet.Age, pet.Resettlement, pet.Phone, pet.ImageName,
	)
}

// --- Utilities ---

func getSecureImageAddress(oriAdd string) string {
	if ImgSrv == "" || oriAdd == "" {
		return ""
	}
	eURL := url.QueryEscape(oriAdd)
	imgGetURL := fmt.Sprintf("%surl?%s", ImgSrv, eURL)

	response, err := http.Get(imgGetURL)
	if err != nil {
		log.Printf("Error downloading image address from %s: %v", imgGetURL, err)
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("Bad status code from image server: %d", response.StatusCode)
		return ""
	}

	totalBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading image server response: %v", err)
		return ""
	}
	return fmt.Sprintf("%simgs?%s.jpg", ImgSrv, string(totalBody))
}
