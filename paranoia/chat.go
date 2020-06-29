package main

import (
	"fmt"
	"math"

	"github.com/Arman92/go-tdlib"
)

func getChats(client *tdlib.Client) ([]*tdlib.Chat, error) {
	var (
		chats []*tdlib.Chat

		offsetOrder  = int64(math.MaxInt64)
		offsetChatID = int64(0)
	)

	for {
		resp, err := client.GetChats(
			tdlib.JSONInt64(offsetOrder),
			offsetChatID,
			1000,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat IDs: %w", err)
		}

		if len(resp.ChatIDs) == 0 {
			break
		}

		for _, chatID := range resp.ChatIDs {
			chat, err := client.GetChat(chatID)
			if err != nil {
				return nil, fmt.Errorf("failed to get chat %d: %w", chatID, err)
			}

			chats = append(chats, chat)
		}

		lastChat := chats[len(chats)-1]

		offsetOrder = int64(lastChat.Order)
		offsetChatID = lastChat.ID
	}

	return chats, nil
}

func getDeletableMesages(client *tdlib.Client, userID int32, chatID int64) ([]int64, error) {
	var (
		messageIDs    []int64
		fromMessageID int64
		messageCount  int32
	)

	for {
		resp, err := client.SearchChatMessages(chatID, "", userID, fromMessageID, 0, 100, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search chat messages: %w", err)
		}

		for _, msg := range resp.Messages {
			if msg.CanBeDeletedForAllUsers {
				messageIDs = append(messageIDs, msg.ID)
			}

			fromMessageID = msg.ID
		}

		messageCount += int32(len(resp.Messages))

		if messageCount >= resp.TotalCount {
			break
		}
	}

	return messageIDs, nil
}
