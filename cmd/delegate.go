/*
 * Copyright (c) CERN 2016
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.cern.ch/flutter/fts/credentials/x509"
	"gitlab.cern.ch/flutter/go-proxy"
	"gitlab.cern.ch/flutter/http-jsonrpc"
	"net/rpc"
	"os"
	"time"
)

type PingReply struct {
	Version string
	Echo    string
}

var endpoint string
var proxyPath string
var lifetime time.Duration

var DelegateCmd = &cobra.Command{
	Use:   "delegate",
	Short: "Delegate a proxy",
	Run: func(cmd *cobra.Command, args []string) {
		codec, err := http_jsonrpc.NewClientCodec(endpoint)
		if err != nil {
			return
		}
		x509d := rpc.NewClientWithCodec(codec)
		defer x509d.Close()

		var reply PingReply
		if err = x509d.Call("X509.Ping", "echo", &reply); err != nil {
			log.Fatal(err)
		}
		log.Info("X509 ", reply.Version)

		proxy := proxy.X509Proxy{}
		if err = proxy.DecodeFromFile(proxyPath); err != nil {
			log.Fatal(err)
		}
		delegationID := proxy.DelegationID()
		log.Info("Delegation ID: ", delegationID)

		request := x509.ProxyRequest{}
		if err = x509d.Call("X509.GetRequest", delegationID, &request); err != nil {
			log.Fatal(err)
		}
		log.Info("Got request")

		newProxy, err := proxy.SignRequest(&request.X509ProxyRequest, lifetime)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Signed new proxy")

		if err = x509d.Call("X509.Put", &x509.Proxy{*newProxy, delegationID}, &delegationID); err != nil {
			log.Fatal(err)
		}
		log.Info("Done")
	},
}

func init() {
	RootCmd.AddCommand(DelegateCmd)
	DelegateCmd.PersistentFlags().StringVar(
		&endpoint, "endpoint", "http://localhost:42001/rpc", "X509 RPC endpoint")
	DelegateCmd.PersistentFlags().StringVar(
		&proxyPath, "proxy", fmt.Sprint("/tmp/x509up_u", os.Getuid()), "X509 proxy")
	DelegateCmd.PersistentFlags().DurationVar(
		&lifetime, "lifetime", 12*time.Hour, "Delegation lifetime")
}
