/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/nabancard/phone-tester/pkg/logging"
	"github.com/nabancard/phone-tester/pkg/version"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	// using standard library "flag" package
	flag.Bool("version", false, "display version")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Printf("Error getting command line arguments: %s", err.Error())
		os.Exit(1)
	}

	getVersion := viper.GetBool("version") // retrieve value from viper

	if getVersion {
		fmt.Printf("version: %s\n", version.Version)
		os.Exit(0)
	}

	viper.SetEnvPrefix("sms")
	viper.AutomaticEnv() // read value ENV variable
	viper.SetDefault("listen_address", "0.0.0.0")
	viper.SetDefault("listen_port", 8080)
	addr := viper.GetString("listen_address")
	port := viper.GetInt("listen_port")

	logger := logging.NewLogger("phone-tester", &zap.Options{})

	logger.Info("stating", "port", port)

	router := gin.Default()

	router.POST("/sms-reply", func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error(err, "failed to process incoming message", jsonData)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var prettyJSON bytes.Buffer
		if err = json.Indent(&prettyJSON, jsonData, " ", " "); err != nil {
			logger.Error(err, "failed to format json response")
			return
		}
		fmt.Printf("message...\n%s\n", prettyJSON.String())
		logger.Info("received sms reply", "message", prettyJSON.String())
		c.JSON(int(200), nil)
	})

	err := router.Run(fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		logger.Error(err, "failed to process incoming message")
	}
}
