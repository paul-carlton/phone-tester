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
	"cmp"
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/paul-carlton/goutils/pkg/logging"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nabancard/phone-tester/pkg/phones"
	"github.com/nabancard/phone-tester/pkg/sms"
	"github.com/nabancard/phone-tester/pkg/version"
)

func main() {
	flag.Bool("version", false, "display version")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Printf("Error getting command line arguments: %s", err.Error())
		os.Exit(1)
	}

	getVersion := viper.GetBool("version")

	if getVersion {
		fmt.Printf("version: %s\n", version.Version)
		os.Exit(0)
	}

	viper.SetEnvPrefix("sms")
	viper.AutomaticEnv() // read value ENV variable
	viper.SetDefault("listen_address", "0.0.0.0")
	viper.SetDefault("listen_port", 8080)
	viper.SetDefault("region", cmp.Or(os.Getenv("AWS_REGION"), "us-west-2"))
	addr := viper.GetString("listen_address")
	region := viper.GetString("region")
	port := viper.GetInt("listen_port")

	logger := logging.NewLogger("phone-tester", &zap.Options{})

	logger.Info("stating", "port", port)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error(err, "failed to initialize phones")
		os.Exit(1)
	}

	smsService, err := sms.NewSMSservice(logger, region)
	if err != nil {
		logger.Error(err, "failed to initialize sms service")
		os.Exit(1)
	}

	if _, err := phones.InitPhones(logger, router, smsService); err != nil {
		logger.Error(err, "failed to initialize phones")
		os.Exit(1)
	}

	err = router.Run(fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		logger.Error(err, "failed to process incoming message")
		os.Exit(1)
	}
}
