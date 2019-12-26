package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nlopes/slack"
)

const (
	bucketname = "samplebucket"
	filename   = "aisnble-output.txt"
	S3Endpoint = "http://localhost:4572"
	timeLayout = "2006-01-02"
	// botIcon    = ":robot:"
)

var (
	commands = map[string]string{
		"help":             "Display all of commands.",
		"ansible-playbook": "Exec ansible-playbook command.",
	}
	output string = ""
)

type Bot struct {
	api *slack.Client
	rtm *slack.RTM
}

func NewBot(token string) *Bot {
	bot := new(Bot)
	bot.api = slack.New(token)
	bot.rtm = bot.api.NewRTM()
	return bot
}

func (b *Bot) handleResponse(user, text, channel, shell string) {
	var cmd string

	commandArray := strings.Fields(text)
	if len(commandArray) <= 1 {
		cmd = "help"
	} else {
		cmd = commandArray[1]
	}

	var attachment slack.Attachment
	var err error

	switch cmd {
	case "ansible-playbook":
		output, attachment, err = b.ansiblePlaybook(shell)
	case "help":
		attachment = b.help()
	default:
		attachment = b.help()
	}

	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("Error: %s", err), channel))
		return
	}

	params := slack.PostMessageParameters{
		Username: botName,
		// IconEmoji: botIcon,
	}

	msgOptText := slack.MsgOptionText("", false)
	msgOptParams := slack.MsgOptionPostMessageParameters(params)
	msgOptAttachment := slack.MsgOptionAttachments(attachment)

	_, _, err = b.api.PostMessage(channel, msgOptText, msgOptParams, msgOptAttachment)

	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("Sorry %s is error... %s", cmd, err), channel))
		return
	}

	// b.rtm.SendMessage(b.rtm.NewOutgoingMessage(string(output), channel))
}

func (b *Bot) help() (attachment slack.Attachment) {
	fields := make([]slack.AttachmentField, 0)

	for k, v := range commands {
		fields = append(fields, slack.AttachmentField{
			Title: "@" + botName + " " + k,
			Value: v,
		})
	}

	attachment = slack.Attachment{
		Pretext: "Command List",
		Color:   "#B733FF",
		Fields:  fields,
	}
	return attachment
}

// execute ansible command and upload log to s3
func (f *Bot) ansiblePlaybook(shell string) (output string, attachment slack.Attachment, err error) {
	fields := make([]slack.AttachmentField, 0)

	cmdOutput, _ := exec.Command("sh", "-c", shell).CombinedOutput()

	r := bytes.NewReader(cmdOutput)

	t := time.Now()
	fmt.Printf("%s", t.Format(timeLayout))

	s := random()

	url := S3PutObject(bucketname, t.Format(timeLayout)+"-"+s+"/"+filename, r)

	fields = append(fields, slack.AttachmentField{
		Title: "@" + botName + " Ansible Output",
		Value: url,
	})

	attachment = slack.Attachment{
		Pretext: "Ansible-Playbook Execution output :",
		Color:   "#A9F5F2",
		Fields:  fields,
	}

	return string(cmdOutput), attachment, nil
}

// func of uploading log to s3
func S3PutObject(bucketname, key string, rs io.ReadSeeker) string {
	s := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("foo", "var", ""),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(endpoints.UsWest2RegionID),
		Endpoint:         aws.String(S3Endpoint),
	}))

	c := s3.New(s, &aws.Config{})

	p := s3.PutObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   rs,
	}

	_, err := c.PutObject(&p)
	if err != nil {
		panic(err)
	}

	filepath := S3Endpoint + "/" + bucketname + "/" + key

	return filepath
}

// get random string
func random() string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, 36)
}
