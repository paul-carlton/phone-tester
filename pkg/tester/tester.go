package tester

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/paul-carlton/goutils/pkg/httpclient"
	"github.com/paul-carlton/goutils/pkg/logging"
)

var (
	ErrorMessageNotFound = errors.New("message not found")
)

func messageNotFoundError(msg string) error {
	return fmt.Errorf("%w: %s", ErrorMessageNotFound, msg)
}

type testRequest struct {
	Scheme   string `json:"scheme" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Payload  string `json:"payload" binding:"required"`
}

type tester struct {
	Tester
	logger logr.Logger
	router *gin.Engine
}

type Tester interface {
	InitHandlers() error

	SendReq(c *gin.Context)
}

func InitTester(log logr.Logger, router *gin.Engine, scheme, serviceEnpoint string) (Tester, error) {
	logging.TraceCall(log)
	defer logging.TraceExit(log)

	tester := tester{
		logger: log,
		router: router,
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

	fmt.Printf("Processing message: %+v\n", msgData)

	hdr := make(httpclient.Header)
	hdr["Content-Type"] = "application/json"
	// hdr["content-length"]=len(message)
	reqResp, err := httpclient.NewReqResp(context.TODO(),
		&url.URL{Scheme: msgData.Scheme, Host: msgData.Endpoint, Path: msgData.Path},
		&httpclient.Post, msgData.Payload, hdr, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if err = reqResp.HTTPreq(); err != nil {
		c.JSON(int(500), gin.H{"error": err.Error()})
	}

	c.IndentedJSON(reqResp.ResponseCode(), *reqResp.RespBody())
}
