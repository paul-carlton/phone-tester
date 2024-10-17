package sms

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/go-logr/logr"
	"github.com/paul-carlton/goutils/pkg/logging"
)

type smsService struct {
	SMSservice
	logger logr.Logger
	svc    *pinpointsmsvoicev2.PinpointSMSVoiceV2
}

type SMSservice interface {
	Init() error

	SendSMS(destination, msg, sender string) (*string, error)
	SendMMS(destination, msg, sender string, urls []*string) (*string, error)
}

func NewSMSservice(log logr.Logger, region string) (SMSservice, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	awsSession := session.Must(session.NewSession())

	// Create a PinpointSMSVoice client with additional configuration
	svc := pinpointsmsvoicev2.New(awsSession, aws.NewConfig().WithRegion(region))

	logging.TraceCall(log)
	defer logging.TraceExit(log)

	smsService := smsService{
		logger: log,
		svc:    svc,
	}

	return &smsService, nil
}

func (s *smsService) SendMMS(destination, msg, sender string, urls []*string) (*string, error) {
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
	output, err := s.svc.SendMediaMessage(input)
	if err != nil {
		return nil, fmt.Errorf("%s - failed to send message to: %s, %w", logging.CallerStr(logging.Me), destination, err)
	}
	return output.MessageId, nil
}

func (s *smsService) SendSMS(destination, msg, sender string) (*string, error) {
	logging.TraceCall(s.logger)
	defer logging.TraceExit(s.logger)

	input := &pinpointsmsvoicev2.SendTextMessageInput{
		//ConfigurationSetName: nil // *string `min:"1" type:"string"`
		// You can specify custom data in this field. If you do, that data is logged
		// to the event destination.
		//Context: nil // map[string]*string `type:"map"`
		DestinationPhoneNumber: &destination,
		MessageBody:            &msg,
		OriginationIdentity:    &sender,
	}
	output, err := s.svc.SendTextMessage(input)
	if err != nil {
		return nil, fmt.Errorf("%s - failed to send message to: %s, %w", logging.CallerStr(logging.Me), destination, err)
	}
	return output.MessageId, nil
}
