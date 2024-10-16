package sms

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoice"
)

type PinpointSMSVoice struct {
    *client.Client
}

func init() {
	mySession := session.Must(session.NewSession())

	// Create a PinpointSMSVoice client with additional configuration
	svc := pinpointsmsvoice.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
}