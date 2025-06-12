package external

import (
	"context"
	"ewallet-transaction/constants"
	"ewallet-transaction/external/proto/notification"
	"ewallet-transaction/helpers"
	"fmt"

	"google.golang.org/grpc"
)

func (*External) SendNotification(ctx context.Context, recipient string, templateName string, placeholder map[string]string) error {
	conn, err := grpc.Dial(helpers.GetEnv("NOTIFICATION_GRPC_HOST", ""), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := notification.NewNotificationServiceClient(conn)
	request := notification.SendNotificationRequest{
		TemplateName: templateName,
		Recipient:    recipient,
		Placeholders: placeholder,
	}

	resp, err := client.SendNotification(ctx, &request)
	if err != nil {
		return err
	}

	if resp.Message != constants.SuccessMessage {
		return fmt.Errorf("get response error from notification: %v", resp.Message)
	}

	return nil
}
