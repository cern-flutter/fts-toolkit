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
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"gitlab.cern.ch/flutter/stomp"
	"time"
)

var debug bool

var StompArgs = stomp.ConnectionParameters{
	ConnectionLost: func(b *stomp.Broker) {
		log.Warn("Lost connection with stomp, reconnect...")
		if err := b.Reconnect(); err != nil {
			l := log.WithField("broker", b.RemoteAddr())
			l.Warn("Lost connection with broker")
			if err := b.Reconnect(); err != nil {
				l.WithError(err).Errorf("Failed to reconnect, wait 1 second")
				time.Sleep(1 * time.Second)
			} else {
			}
		}
	},
	ClientID: "fts-toolkit-" + uuid.NewV4().String(),
}

var RootCmd = &cobra.Command{
	Use:   "fts-toolkit",
	Short: "Toolkit for FTS development",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug")
	RootCmd.PersistentFlags().StringVar(&StompArgs.Address, "stomp", "localhost:61613", "Stomp host and port")
	RootCmd.PersistentFlags().StringVar(&StompArgs.Login, "stomp-login", "fts", "Stomp loging")
	RootCmd.PersistentFlags().StringVar(&StompArgs.Passcode, "stomp-passcode", "fts", "Stomp passcode")
}
