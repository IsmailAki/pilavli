package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Users struct {
	gorm.Model
	UserID string
}

var db *gorm.DB

func main() {

	db, err := gorm.Open(sqlite.Open("pilavli.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Users{})

	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!bilgilendir" {
		fmt.Println(db.Where("user_id = ?", m.Author.ID).First(&Users{}).RowsAffected)
		if db.Where("user_id = ?", m.Author.ID).First(&Users{}).RowsAffected == 0 {
			channel, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				fmt.Println("error creating channel:", err)
				return
			}
			_, err = s.ChannelMessageSend(channel.ID, "Bildirimler basariyla acildi. Gelecek yayinlarda buradan bildirim alacaksin.")
			if err != nil {
				fmt.Println(err)
				s.ChannelMessageSendReply(m.ChannelID, "DM'in kapali oldugu icin bildirimler acilamadi. DM'i acip tekrar !bilgilendir yazabilirsin.", &discordgo.MessageReference{MessageID: m.Message.ID, ChannelID: m.ChannelID, GuildID: m.GuildID})

				return
			} else {
				db.Create(&Users{UserID: m.Author.ID})
				s.MessageReactionAdd(m.ChannelID, m.Message.ID, "👍")
				return
			}
		} else {
			s.ChannelMessageSendReply(m.ChannelID, "Bildirimler daha onceden acilmis iptal etmek icin !bilgilendirme yazabilirsin.", &discordgo.MessageReference{MessageID: m.Message.ID, ChannelID: m.ChannelID, GuildID: m.GuildID})
			return
		}
	}

	if m.Content == "!bilgilendirme" {
		fmt.Println(db.Where("user_id = ?", m.Author.ID).First(&Users{}).RowsAffected)
		if db.Where("user_id = ?", m.Author.ID).First(&Users{}).RowsAffected > 0 {
			db.Where("user_id = ?", m.Author.ID).Delete(&Users{})
			s.MessageReactionAdd(m.ChannelID, m.Message.ID, "👍")
			return
		} else {
			s.ChannelMessageSendReply(m.ChannelID, "Acik bildirim yok bildirimleri acmak icin !bilgilendir yazabilirsin.", &discordgo.MessageReference{MessageID: m.Message.ID, ChannelID: m.ChannelID, GuildID: m.GuildID})
			return
		}
	}

	if strings.Split(m.Content, " ")[0] == "!bilgiver" && m.Author.ID == "268062426312736770" {
		var users []Users
		db.Find(&users)
		for _, d := range users {
			channel, err := s.UserChannelCreate(d.UserID)
			if err != nil {
				fmt.Println("error creating channel:", err)
				continue
			}
			_, err = s.ChannelMessageSend(channel.ID, strings.Join(strings.Split(m.Content, " ")[1:], " "))
			if err != nil {
				fmt.Println("error sending message:", err)
				continue
			}
		}
	}
}
