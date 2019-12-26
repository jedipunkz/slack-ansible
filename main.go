package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

var (
	botId   string
	botName string
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".slack-ansible")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	}
}

func main() {
	token := viper.GetString("token")

	bot := NewBot(token)

	go bot.rtm.ManageConnection()

	for {
		select {
		case msg := <-bot.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				botId = ev.Info.User.ID
				botName = ev.Info.User.Name

			case *slack.MessageEvent:
				user := ev.User
				text := ev.Text
				channel := ev.Channel
				// shellArray := strings.Fields(text)
				// shell := shellArray[:]
				shell := strings.Replace(text, "<@"+botId+">", "", 1)

				if ev.Type == "message" && strings.HasPrefix(text, "<@"+botId+">") {
					bot.handleResponse(user, text, channel, shell)
				}
			}
		}
	}
}
