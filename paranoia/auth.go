package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/Arman92/go-tdlib"
	"golang.org/x/crypto/ssh/terminal"
)

func authorize(client *tdlib.Client) error {
	for {
		currentState, err := client.Authorize()
		if err != nil {
			return fmt.Errorf("failed to authorize: %w", err)
		}

		var input string

		switch v := currentState.GetAuthorizationStateEnum(); v {
		case tdlib.AuthorizationStateWaitPhoneNumberType:
			fmt.Print("Enter phone: ")
			fmt.Scanln(&input)

			if _, err = client.SendPhoneNumber(input); err != nil {
				return fmt.Errorf("failed to send phone number: %w", err)
			}
		case tdlib.AuthorizationStateWaitCodeType:
			fmt.Print("Enter code: ")
			fmt.Scanln(&input)

			if _, err = client.SendAuthCode(input); err != nil {
				return fmt.Errorf("failed to send auth code: %w", err)
			}
		case tdlib.AuthorizationStateWaitPasswordType:
			fmt.Print("Enter password: ")

			passwd, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to scan password: %w", err)
			}

			fmt.Print("\n")

			if _, err = client.SendAuthPassword(string(passwd)); err != nil {
				return fmt.Errorf("failed to send auth password: %w", err)
			}
		case tdlib.AuthorizationStateReadyType:
			return nil
		case tdlib.AuthorizationStateWaitEncryptionKeyType:
			time.Sleep(time.Second)
		default:
			return fmt.Errorf("unexpected authorization state: %s", v)
		}
	}
}
