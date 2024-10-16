package phones

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/paul-carlton/goutils/pkg/logging"
)

type sendMessage struct {
	messageBody string
}

type externalMessage struct {
	originationNumber string
	destinationNumber string
	messageKeyword    string //nolint: unused
	messageBody       string
	inboundMessageId  string //nolint: revive,stylecheck
}

type message struct {
	Message `json:"-"`
	// logger  logr.Logger `json:"-"`
	Sender string `json:"sender,omitempty"`
	Text   string `json:"text,omitempty"`
	Id     string `json:"id,omitempty"` //nolint: revive,stylecheck
}

type Message interface {
	Get() *message
	Reply(msg string) error
}

func (m *message) Get() *message {
	return m
}

func (m *message) Reply(msg string) error { //nolint: revive
	return nil
}

type phone struct {
	Phone    `json:"-"`
	logger   logr.Logger         `json:"-"`
	Number   string              `json:"Number,omitempty"`
	Messages map[string]*message `json:"Messages,omitempty"`
}

type Phone interface {
	ReceiveSMS(msg *message)
	SendSMS(destination, msg string) error
	GetMessages() []*message
	GetMessage(id string) (*message, error)
}

func (n *phone) ReceiveSMS(msg *message) {
	n.Messages[msg.Id] = msg
}

func (n *phone) SendSMS(destination, msg string) error { //nolint: revive
	return nil
}

func (n *phone) GetMessages() []*message {
	resp := []*message{}
	for _, msg := range n.Messages {
		resp = append(resp, msg)
	}
	return resp
}

func (n *phone) GetMessage(id string) (*message, error) {
	theMsg, ok := n.Messages[id]
	if !ok {
		return nil, fmt.Errorf("Message: %s, not found", id) //nolint: err113
	}
	return theMsg, nil
}

type phones struct {
	Phones
	logger logr.Logger
	router *gin.Engine
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

func InitPhones(log logr.Logger, router *gin.Engine) (Phones, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	phones := phones{
		logger: log,
		router: router,
	}

	if err := phones.InitHandlers(); err != nil {
		return nil, err
	}

	return &phones, nil
}

func (p *phones) InitHandlers() error {
	p.router.POST("/sms", p.ReceiveSMS)
	p.router.GET("/phones", p.GetPhones)
	p.router.GET("/phones/:number/messages", p.GetPhoneMessages)
	p.router.POST("/phones/:number/messages/:id", p.ReplyToPhoneMessage)
	return nil
}

func (p *phones) GetPhones(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	response := []string{}
	for _, phone := range p.phones {
		response = append(response, phone.Number)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func (p *phones) GetPhoneMessages(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	number := c.Param("number")
	thePhone, ok := p.phones[number]
	if !ok {
		c.JSON(int(404), fmt.Sprintf("Phone: %s, not found", number))
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
	}
	msg, err := thePhone.GetMessage(messageID)
	if err != nil {
		c.JSON(int(404), err.Error())
	}

	var msgData sendMessage
	if err := c.BindJSON(&msgData); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := msg.Reply(msgData.messageBody); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusAccepted, nil)
}

func (p *phones) ReceiveSMS(c *gin.Context) {
	logging.TraceCall(p.logger)
	defer logging.TraceExit(p.logger)

	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		p.logger.Error(err, "failed to process incoming message", jsonData)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var prettyJSON bytes.Buffer
	if err = json.Indent(&prettyJSON, jsonData, " ", " "); err != nil {
		p.logger.Error(err, "failed to format json response")
		return
	}
	
	if logging.TraceLevel > 3 {
		fmt.Printf("%smessage...\n%s\n", logging.CallerText(logging.MyCaller), prettyJSON.String())
	}

	var emsg externalMessage
	if err := c.BindJSON(&emsg); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	msg := &message{
		Sender: emsg.originationNumber,
		Text:   emsg.messageBody,
		Id:     emsg.inboundMessageId,
	}

	phone := p.GetPhone(emsg.destinationNumber)
	phone.ReceiveSMS(msg)

	c.JSON(int(200), nil)
}

func (p *phones) GetPhone(number string) Phone {
	thePhone, ok := p.phones[number]
	if !ok {
		thePhone = &phone{
			logger:   p.logger,
			Number:   number,
			Messages: map[string]*message{},
		}
	}
	return thePhone
}
