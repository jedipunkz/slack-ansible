package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/go-homedir"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

var (
	bucketname string = "samplebucket"
	filename   string = "sample.txt"
	outputfile string = "/tmp/slack-ansible"
)

const (
	S3Endpoint = "http://localhost:4572"
	timeLayout = "2006-01-02"
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
	api := slack.New(token)
	os.Exit(run(api))
}

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		msg := <-rtm.IncomingEvents
		log.Printf("MSG: %#v\n", msg.Data)

		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			log.Printf("Start up!")

		case *slack.MessageEvent:
			if strings.HasPrefix(ev.Text, "!echo") {
				out, err := exec.Command("sh", "-c", ev.Text[1:]).Output()

				r := bytes.NewReader(out)

				// get now time
				t := time.Now()
				fmt.Printf("%s", t.Format(timeLayout))

				// get random string
				s := random()

				// upload object to s3 bucket
				url := S3PutObject(bucketname, t.Format(timeLayout)+"-"+s+"/"+filename, r)
				// url := S3PutObject(bucketname, filename, r)
				if err != nil {
					fmt.Printf("Fatal error : %s \n", err)
				}

				rtm.SendMessage(rtm.NewOutgoingMessage(url+"\n"+"```"+string(out)+"```", ev.Channel))
			}
		}
	}
}

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

func random() string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, 36)
}
