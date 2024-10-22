package tester

import (
	"context"
	"encoding/json"
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
	DataFile string `json:"datafile,omitempty" binding:"-"`
}

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
	logging.TraceCall()
	defer logging.TraceExit()

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
	logging.TraceCall()
	defer logging.TraceExit()

	t.router.POST("/test", t.SendReq)
	return nil
}

func (t *tester) SendReq(c *gin.Context) {
	var msgData testRequest
	if err := c.BindJSON(&msgData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("request...\n%s\n", logging.ToJSON(t.logger, msgData))

	method := &httpclient.Get
	var data []byte
	if len(msgData.DataFile) > 0 {
		var err error
		if data, err = os.ReadFile(fmt.Sprintf("testdata/%s.json", msgData.DataFile)); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		method = &httpclient.Post
	}

	if len(data) > 0 {
		fmt.Printf("message to send...\n%s\n", string(data))
	}

	if err := t.reqResp.HTTPreq(method, &url.URL{Scheme: msgData.Scheme, Host: msgData.Endpoint, Path: msgData.Path}, string(data), nil); err != nil {
		c.JSON(int(500), gin.H{"error": err.Error()})
		return
	}

	reply := *t.reqResp.RespBody()

	fmt.Printf("reply received, Response Code: %d", t.reqResp.RespCode())
	if len(reply) > 0 {
		fmt.Printf(", Payload...\n%s\n", reply)
		c.IndentedJSON(t.reqResp.RespCode(), json.RawMessage(reply))
		return
	}
	fmt.Printf("\n")
	c.Status(t.reqResp.RespCode())
}
