package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/Arman92/go-tdlib"
)

var (
	dataDir string
	logfile string
	delay   time.Duration

	apiID   = "187786"
	apiHash = "e782045df67ba48e441ccb105da8fc85"
)

func newConfig() tdlib.Config {
	return tdlib.Config{
		APIID:               apiID,
		APIHash:             apiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
		UseMessageDatabase:  true,
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseTestDataCenter:   false,
		DatabaseDirectory:   filepath.Join(dataDir, "db"),
		FileDirectory:       filepath.Join(dataDir, "files"),
		IgnoreFileNames:     false,
	}
}

func main() {
	flag.StringVar(&dataDir, "dir", "tdlib", "tdlib data directory")
	flag.StringVar(&logfile, "logfile", "", "tdlib logfile")
	flag.DurationVar(&delay, "delay", time.Second*5, "time to receive updates")
	flag.Parse()

	tdlib.SetLogVerbosityLevel(1)

	if logfile != "" {
		tdlib.SetFilePath(logfile)
	}

	client := tdlib.NewClient(newConfig())
	defer client.Close()

	if err := authorize(client); err != nil {
		log.Println(fmt.Errorf("failed to authorize: %w", err))
		return
	}

	currentUser, err := client.GetMe()
	if err != nil {
		log.Println(fmt.Errorf("failed to get current user information: %w", err))
		return
	}

	log.Printf("Logged in as %s (%d)\n", currentUser.FirstName, currentUser.ID)

	log.Printf("Waiting %s to get updates\n", delay)
	time.Sleep(delay)

	chats, err := getChats(client)
	if err != nil {
		log.Println(fmt.Errorf("failed to get chat list: %w", err))
		return
	}

	var supergroups []*tdlib.Chat

	for _, chat := range chats {
		if chat.Type.GetChatTypeEnum() != tdlib.ChatTypeSupergroupType {
			continue
		}

		supergroup, ok := chat.Type.(*tdlib.ChatTypeSupergroup)
		if !ok {
			log.Println(errors.New("failed to cast chat type"))
			return
		}

		if supergroup.IsChannel {
			continue
		}

		supergroups = append(supergroups, chat)
	}

	if len(supergroups) == 0 {
		log.Println("You don't seem to participate in any supergroup")
		return
	}

	for _, chat := range supergroups {
		log.Printf("Trying to delete all your messages in %s (%d)\n", chat.Title, chat.ID)

		messageIDs, err := getDeletableMesages(client, currentUser.ID, chat.ID)
		if err != nil {
			log.Println(fmt.Errorf("failed to get deletable messages: %w", err))
			return
		}

		if len(messageIDs) == 0 {
			log.Println("No deletable messages were found")
			continue
		}

		log.Printf("Messages to be deleted: %d\n", len(messageIDs))

		if _, err = client.DeleteMessages(chat.ID, messageIDs, true); err != nil {
			log.Println(fmt.Errorf("failed to delete messages: %w", err))
			return
		}
	}

	log.Println("Done. Congratulations, you coward.")
}
