package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	_ "github.com/joho/godotenv/autoload"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func getEmoji() string {
	emojis := []string{"👍", "👎", "😢", "🎉", "😡", "🙏"}

	// Seed the random number generator
	// Create a local random generator with a time-based seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomEmoji := emojis[r.Intn(len(emojis))]

	fmt.Println("Random emoji:", randomEmoji)

	return randomEmoji
}

func botOrFail(name string, opts []bot.Option) *bot.Bot {
	b, err := bot.New(os.Getenv(name), opts...)
	if err != nil {
		failOnError(err, "failed to connect to telegram api")
	}

	return b
}

func main() {
	log.Println("starting telebot")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	msgs := make(chan models.Update)

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		log.Printf("g received: ChatID: %d, MSG: %s", update.Message.Chat.ID, update.Message.Text)
		msgs <- *update
	}

	masterBot := botOrFail("TELEGRAM_BOT_TOKEN", []bot.Option{
		bot.WithDefaultHandler(handler),
	})

	reactionBots := []bot.Bot{
		// *masterBot,
		*botOrFail("TELEGRAM_BOT_TOKEN_2", []bot.Option{}),
		*botOrFail("TELEGRAM_BOT_TOKEN_3", []bot.Option{}),
	}

	go func() {
		for update := range msgs {
			for i := range reactionBots {
				b := &reactionBots[i]
				isBig := rand.Intn(2) == 1

				ok, error := b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
					Reaction: []models.ReactionType{{
						Type: models.ReactionTypeTypeEmoji,
						ReactionTypeEmoji: &models.ReactionTypeEmoji{
							Emoji: getEmoji(),
						},
					}},
					IsBig: &isBig,
				})
				log.Printf("result: %v, MSG: %v", ok, error)
			}
		}
	}()

	log.Printf("Listening for messages")
	masterBot.Start(ctx)
}
