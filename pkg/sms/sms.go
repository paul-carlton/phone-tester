package sms

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/paul-carlton/goutils/pkg/logging"
)

type smsService struct {
	SMSservice
	logger *slog.Logger
	svc    *pinpointsmsvoicev2.Client
}

type SMSservice interface {
	Init() error

	SendSMS(destination, msg, sender string) (*string, error)
	SendMMS(destination, msg, sender string, urls []string) (*string, error)
}

func NewSMSservice(log *slog.Logger, region string) (SMSservice, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("%s - failed to load AWS SDK config, %w", logging.CallerStr(logging.Me), err)
	}

	// Create a PinpointSMSVoice client with additional configuration
	svc := pinpointsmsvoicev2.NewFromConfig(cfg)

	logging.TraceCall(log)
	defer logging.TraceExit(log)

	smsService := smsService{
		logger: log,
		svc:    svc,
	}

	return &smsService, nil
}

func (s *smsService) SendMMS(destination, msg, sender string, urls []string) (*string, error) {
	logging.TraceCall(s.logger)
	defer logging.TraceExit(s.logger)

	input := &pinpointsmsvoicev2.SendMediaMessageInput{
		//ConfigurationSetName: nil // *string `min:"1" type:"string"`
		// You can specify custom data in this field. If you do, that data is logged
		// to the event destination.
		//Context: nil // map[string]*string `type:"map"`
		DestinationPhoneNumber: &destination,
		MessageBody:            &msg,
		MediaUrls:              urls,
		OriginationIdentity:    &sender,
	}
	output, err := s.svc.SendMediaMessage(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("%s - failed to send message to: %s, %w", logging.CallerStr(logging.Me), destination, err)
	}
	return output.MessageId, nil
}

func (s *smsService) SendSMS(destination, msg, sender string) (*string, error) {
	logging.TraceCall(s.logger)
	defer logging.TraceExit(s.logger)

	s.logger.Log(context.TODO(), logging.LevelTrace, "message", "destination", destination)

	input := &pinpointsmsvoicev2.SendTextMessageInput{
		//ConfigurationSetName: nil // *string `min:"1" type:"string"`
		// You can specify custom data in this field. If you do, that data is logged
		// to the event destination.
		//Context: nil // map[string]*string `type:"map"`
		DestinationPhoneNumber: &destination,
		MessageBody:            &msg,
		OriginationIdentity:    &sender,
	}

	s.logger.Log(context.TODO(), logging.LevelTrace, "input", "data", *input.MessageBody)

	output, err := s.svc.SendTextMessage(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("%s - failed to send message to: %s, %w", logging.CallerStr(logging.Me), destination, err)
	}

	s.logger.Log(context.TODO(), logging.LevelTrace, "output", "messageID", *output.MessageId)
	return output.MessageId, nil
}
