package tester

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/paul-carlton/goutils/pkg/httpclient"
	"github.com/paul-carlton/goutils/pkg/logging"
)

type testRequest struct {
	Scheme   string `json:"scheme" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	Path     string `json:"path" binding:"required"`
	DataFile string `json:"datafile" binding:"required"`
}

// type sms struct {
// 	OriginationNumber          string `json:"originationNumber"`
// 	DestinationNumber          string `json:"destinationNumber"`
// 	MessageKeyword             string `json:"messageKeyword"`
// 	MessageBody                string `json:"messageBody"`
// 	PreviousPublishedMessageID string `json:"previousPublishedMessageId"`
// 	InboundMessageID           string `json:"inboundMessageId"`
// }

type tester struct {
	Tester
	logger  *slog.Logger
	router  *gin.Engine
	reqResp httpclient.ReqResp
}

type Tester interface {
	InitHandlers() error

	SendReq(c *gin.Context)
}

func InitTester(log *slog.Logger, router *gin.Engine) (Tester, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	reqResp, err := httpclient.NewReqResp(context.TODO(), log, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	tester := tester{
		logger:  log,
		router:  router,
		reqResp: reqResp,
	}

	if err := tester.InitHandlers(); err != nil {
		return nil, err
	}

	return &tester, nil
}

func (t *tester) InitHandlers() error {
	logging.TraceCall(t.logger)
	defer logging.TraceExit(t.logger)

	t.router.POST("/test", t.SendReq)
	return nil
}

func (t *tester) SendReq(c *gin.Context) {
	var msgData testRequest
	if err := c.BindJSON(&msgData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", msgData.DataFile))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	t.logger.Log(context.TODO(), logging.LevelTrace, "message to send", "message data", string(data))

	// sms1 := sms{
	// 	OriginationNumber: "1",
	// 	DestinationNumber: "2",
	// 	MessageKeyword:    "3",
	// 	MessageBody:       "STOP",
	// 	InboundMessageID:  "444",
	// }

	// data, err := json.Marshal(&sms1)
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
	// fmt.Printf("message...\n%s\n", data)

	if err = t.reqResp.HTTPreq(&httpclient.Post, &url.URL{Scheme: msgData.Scheme, Host: msgData.Endpoint, Path: msgData.Path}, string(data), nil); err != nil {
		c.JSON(int(500), gin.H{"error": err.Error()})
		return
	}

	reply := *t.reqResp.RespBody()

	t.logger.Log(context.TODO(), logging.LevelTrace, "reply", "response code", t.reqResp.RespCode(), "response body", reply)
	if len(reply) > 0 {
		c.IndentedJSON(t.reqResp.RespCode(), reply)
		return
	}
	c.Status(t.reqResp.RespCode())
}
