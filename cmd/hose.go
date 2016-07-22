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
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"gitlab.cern.ch/flutter/echelon/testutil"
	"gitlab.cern.ch/flutter/fts/config"
	"gitlab.cern.ch/flutter/fts/types/surl"
	"gitlab.cern.ch/flutter/fts/types/tasks"
	"gitlab.cern.ch/flutter/stomp"
	"time"
)

var sleep = time.Second
var nMsgs = 1
var sourceSes = []string{}
var destSes = []string{}
var states = []string{}
var delegationId = "123456789"
var vo = "dteam"
var activity = "default"
var persistent = false

var HoseCmd = &cobra.Command{
	Use:   "hose",
	Short: "Produce a set of messages",
	Run: func(cmd *cobra.Command, args []string) {
		destination := config.WorkerQueue
		if len(args) > 0 {
			destination = args[0]
		}

		producer, err := stomp.NewProducer(StompArgs)
		if err != nil {
			log.Panic(err)
		}

		sendParams := stomp.SendParams{
			Persistent:  persistent,
			ContentType: "application/json",
		}

		for i := 0; i < nMsgs; i++ {
			batch := GenerateRandomTransfer()
			data, err := json.Marshal(batch)
			if err != nil {
				log.Panic(err)
			}
			log.Info("Sending batch ", batch.GetID())
			producer.Send(destination, string(data), sendParams)

			time.Sleep(sleep)
		}
	},
}

func GenerateRandomTransfer() *tasks.Batch {
	sourceSe := testutil.RandomChoice(sourceSes)
	destSe := testutil.RandomChoice(destSes)
	file := testutil.RandomFile()

	sourceSurl, err := surl.Parse(sourceSe + file)
	if err != nil {
		log.Panic(err)
	}
	destSurl, err := surl.Parse(destSe + file)
	if err != nil {
		log.Panic(err)
	}

	batchState := tasks.BatchState(testutil.RandomChoice(states))
	var fileState tasks.TransferState
	switch batchState {
	case tasks.BatchSubmitted:
		fileState = tasks.TransferSubmitted
	case tasks.BatchReady:
		fileState = tasks.TransferActive
	case tasks.BatchRunning:
		fileState = tasks.TransferActive
	case tasks.BatchDone:
		fileState = tasks.TransferState(testutil.RandomChoice([]string{
			string(tasks.TransferFinished), string(tasks.TransferFailed)},
		))
	}

	return &tasks.Batch{
		Type:  tasks.BatchSimple,
		State: batchState,
		Transfers: []*tasks.Transfer{
			{
				State:       fileState,
				JobID:       tasks.JobID(uuid.NewV4().String()),
				TransferID:  tasks.TransferID(uuid.NewV4().String()),
				Retry:       0,
				Source:      *sourceSurl,
				Destination: *destSurl,
				Activity:    activity,
			},
		},
		DelegationID: delegationId,

		SourceSe: sourceSe,
		DestSe:   destSe,
		Vo:       vo,
		Activity: activity,
	}
}

func init() {
	RootCmd.AddCommand(HoseCmd)
	HoseCmd.PersistentFlags().DurationVar(&sleep, "sleep", time.Second, "Time to sleep between submissions")
	HoseCmd.PersistentFlags().IntVar(&nMsgs, "count", 1, "Number of messages")
	HoseCmd.PersistentFlags().StringSliceVar(
		&sourceSes, "source", []string{"mock://source.es"}, "Possible source storages")
	HoseCmd.PersistentFlags().StringSliceVar(
		&destSes, "dest", []string{"mock://dest.ch"}, "Possible destination storages")
	HoseCmd.PersistentFlags().StringSliceVar(
		&states, "states", []string{string(tasks.BatchReady)}, "Possible states")
	HoseCmd.PersistentFlags().StringVar(&delegationId, "delegation-id", "123456789", "Delegation id")
	HoseCmd.PersistentFlags().StringVar(&vo, "vo", "dteam", "VO")
	HoseCmd.PersistentFlags().BoolVar(&persistent, "persist", false, "Persist messages")
}
