package main

import (
	"log"
	"os"
	"os/signal"
	"role-invites-bot/internal/database"
	"role-invites-bot/internal/handlers"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("TOKEN")

	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	sess.AddHandler(handlers.CommandHandler)
	handlers.EventHandler(sess)

	database.Init()

	log.Println("Logged as " + sess.State.User.Username + "#" + sess.State.User.Discriminator)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
