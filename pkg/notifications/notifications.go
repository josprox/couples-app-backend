package notifications

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var client *messaging.Client

// InitFirebase initializes the Firebase Admin SDK.
// It expects a service account file at the given path.
func InitFirebase(serviceAccountPath string) error {
	ctx := context.Background()
	opt := option.WithServiceAccountFile(serviceAccountPath)
	
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	client, err = app.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}

	return nil
}

// SendPushNotification sends a notification to a specific FCM token.
func SendPushNotification(token, title, body string, data map[string]string) error {
	if client == nil {
		log.Println("Firebase client not initialized, skipping notification")
		return nil
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := client.Send(context.Background(), message)
	return err
}
