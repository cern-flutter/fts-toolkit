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
	"gitlab.cern.ch/flutter/fts/config"
	"gitlab.cern.ch/flutter/stomp"
)

var DrainCmd = &cobra.Command{
	Use:   "drain",
	Short: "Drain a Stomp queue",
	Run: func(cmd *cobra.Command, args []string) {
		drainDestinations := args
		if len(drainDestinations) == 0 {
			drainDestinations = []string{config.TransferTopic}
		}

		consumer, err := stomp.NewConsumer(StompArgs)
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Close()
		log.Info("Connected")

		for _, destination := range drainDestinations {
			go func(destination string) {
				id := "drain-" + uuid.NewV4().String()
				msgs, errors, err := consumer.Subscribe(destination, id, stomp.AckAuto)
				if err != nil {
					log.Fatal(err)
				}

				log.Info("Subscribed to ", destination)

				for {
					select {
					case m := <-msgs:
						log.Debug(m.Headers)
						log.Info(destination, ": ", string(m.Body))
					case e := <-errors:
						log.Error(destination, ": ", e)
					}
				}
			}(destination)
		}

		_ = <-make(chan bool)
	},
}

func init() {
	RootCmd.AddCommand(DrainCmd)
}
