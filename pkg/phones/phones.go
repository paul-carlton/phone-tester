package phones

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/paul-carlton/goutils/pkg/logging"

	"github.com/paul-carlton/phone-tester/pkg/sms"
)

var (
	ErrorMessageNotFound = errors.New("message not found")
	ErrorFailedToSendSMS = errors.New("failed to send SMS message")
)

func messageNotFoundError(msg string) error {
	return fmt.Errorf("%w: %s", ErrorMessageNotFound, msg)
}

func failedToSendSMSmessage(msg string) error {
	return fmt.Errorf("%w: %s", ErrorFailedToSendSMS, msg)
}

type sendMessage struct {
	MessageBody string `json:"messageBody" binding:"required"`
}

type externalMessage struct {
	OriginationNumber string `json:"originationNumber" binding:"required"`
	DestinationNumber string `json:"destinationNumber" binding:"required"`
	MessageKeyword    string `json:"messageKeyword" binding:"required"`
	MessageBody       string `json:"messageBody" binding:"required"`
	SentMessageId     string `json:"previousPublishedMessageId,omitempty" binding:"-"` //nolint:revive,stylecheck
	InboundMessageID  string `json:"inboundMessageId" binding:"required"`
}

type message struct {
	Message `json:"-"`
	logger  logr.Logger `json:"-"`
	Sender  string      `json:"sender,omitempty"`
	Text    string      `json:"text,omitempty"`
	Id      string      `json:"id,omitempty"` //nolint: revive,stylecheck
}

type Message interface {
	Get() *message
}

func (m *message) Get() *message {
	logging.TraceCall(m.logger)
	defer logging.TraceExit(m.logger)

	return m
}

type phone struct {
	Phone      `json:"-"`
	logger     logr.Logger         `json:"-"`
	smsService sms.SMSservice      `json:"-"`
	Number     string              `json:"Number,omitempty"`
	Messages   map[string]*message `json:"Messages,omitempty"`
}

type Phone interface {
	ReceiveSMS(msg *message)
	SendSMS(destination, msg string) error
	GetMessages() []*message
	GetMessage(id string) (*message, error)
}

func (n *phone) ReceiveSMS(msg *message) {
	logging.TraceCall(n.logger)
	defer logging.TraceExit(n.logger)

	n.Messages[msg.Id] = msg
}

func (n *phone) SendSMS(destination, msg string) error {
	logging.TraceCall(n.logger)
	defer logging.TraceExit(n.logger)

	_, err := n.smsService.SendSMS(destination, msg, n.Number)
	if err != nil {
		return failedToSendSMSmessage(err.Error())
	}

	return nil
}

func (n *phone) GetMessages() []*message {
	logging.TraceCall(n.logger)
	defer logging.TraceExit(n.logger)

	resp := []*message{}
	for _, msg := range n.Messages {
		resp = append(resp, msg)
	}
	return resp
}

func (n *phone) GetMessage(id string) (*message, error) {
	logging.TraceCall(n.logger)
	defer logging.TraceExit(n.logger)

	theMsg, ok := n.Messages[id]
	if !ok {
		return nil, messageNotFoundError(fmt.Sprintf("Message: %s, not found", id))
	}
	return theMsg, nil
}

type phones struct {
	Phones
	logger logr.Logger
	router *gin.Engine
	sms    sms.SMSservice
	phones map[string]*phone
}

type Phones interface {
	InitHandlers() error

	ReceiveSMS(c *gin.Context)
	ReplyToPhoneMessage(c *gin.Context)
	GetPhones(c *gin.Context)
	GetPhoneMessages(c *gin.Context)

	GetPhone(number string) Phone
}

func InitPhones(log logr.Logger, router *gin.Engine, sms sms.SMSservice) (Phones, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	phones := phones{
		logger: log,
		router: router,
		sms:    sms,
		phones: make(map[string]*phone),
	}

	if err := phones.InitHandlers(); err != nil {
		return nil, err
	}

	return &phones, nil
}

func (p *phones) InitHandlers() error {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	p.router.POST("/sms", p.ReceiveSMS)
	p.router.GET("/phones", p.GetPhones)
	p.router.GET("/phones/:number/messages", p.GetPhoneMessages)
	p.router.POST("/phones/:number/messages/:id", p.ReplyToPhoneMessage)
	return nil
}

func (p *phones) GetPhones(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	if logging.TraceLevel == 0 {
		fmt.Printf("%sphones...\n%+v\n", logging.CallerText(logging.MyCaller), p.phones)
	}

	response := []string{}
	for _, phone := range p.phones {
		response = append(response, phone.Number)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func (p *phones) GetPhoneMessages(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	if logging.TraceLevel == 0 {
		fmt.Printf("%sphones...\n%+v\n", logging.CallerText(logging.MyCaller), p.phones)
	}

	number := c.Param("number")

	if logging.TraceLevel == 0 {
		fmt.Printf("%snumber:%+v\n", logging.CallerText(logging.MyCaller), number)
	}

	thePhone, ok := p.phones[number]
	if !ok {
		c.JSON(int(404), fmt.Sprintf("Phone: %s, not found", number))
		return
	}

	if logging.TraceLevel == 0 {
		fmt.Printf("%sphone...\n%+v\n", logging.CallerText(logging.MyCaller), thePhone)
	}

	c.IndentedJSON(http.StatusOK, thePhone.GetMessages())
}

func (p *phones) ReplyToPhoneMessage(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	number := c.Param("number")
	messageID := c.Param("id")
	thePhone, ok := p.phones[number]
	if !ok {
		c.JSON(int(404), fmt.Sprintf("Phone: %s, not found", number))
		return
	}
	msg, err := thePhone.GetMessage(messageID)
	if err != nil {
		c.JSON(int(404), err.Error())
	}

	var msgData sendMessage
	if err := c.BindJSON(&msgData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if logging.TraceLevel == 0 {
		fmt.Printf("%smsg: %+v\n", logging.CallerText(logging.MyCaller), msgData)
	}

	if err := thePhone.SendSMS(msg.Sender, msgData.MessageBody); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusAccepted, nil)
}

func (p *phones) ReceiveSMS(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	var emsg externalMessage
	if err := c.BindJSON(&emsg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if logging.TraceLevel == 0 {
		fmt.Printf("%smessage...\n%+v\n", logging.CallerText(logging.MyCaller), emsg)
	}

	msg := &message{
		logger: p.logger,
		Sender: emsg.OriginationNumber,
		Text:   emsg.MessageBody,
		Id:     emsg.InboundMessageID,
	}

	phone := p.GetPhone(emsg.DestinationNumber)
	phone.ReceiveSMS(msg)

	c.JSON(int(200), nil)
}

func (p *phones) GetPhone(number string) Phone {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	thePhone, ok := p.phones[number]
	if !ok {
		thePhone = &phone{
			logger:     p.logger,
			smsService: p.sms,
			Number:     number,
			Messages:   map[string]*message{},
		}
		p.phones[number] = thePhone
	}
	return thePhone
}
