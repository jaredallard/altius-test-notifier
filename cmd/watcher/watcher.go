// Copyright 2020 Jared Allard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jaredallard/altius-test-notifier/internal/altius"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() { //nolint:funlen,gocyclo
	ctx, cancel := context.WithCancel(context.Background())
	log := logrus.New()

	exitCode := 0
	defer os.Exit(exitCode)

	// this prevents the CLI from clobbering context cancellation
	cli.OsExiter = func(code int) {
		exitCode = code
	}

	app := cli.App{
		Version:              "1.0.0",
		EnableBashCompletion: true,
		Name:                 "watcher",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "retrieval-code",
				EnvVars: []string{"ALTIUS_RETRIEVAL_CODE"},
				Usage:   "Set the Altius Retrieval Code",
			},
			&cli.StringFlag{
				Name:    "date-of-birth",
				Aliases: []string{"d"},
				EnvVars: []string{"ALTIUS_DATE_OF_BIRTH"},
				Usage:   "Set the date of birth for the given code",
			},
			&cli.StringFlag{
				Name:    "telegram-token",
				EnvVars: []string{"TELEGRAM_TOKEN"},
				Usage:   "Telegram Token",
			},
			&cli.Int64Flag{
				Name:    "telegram-chat-id",
				EnvVars: []string{"TELEGRAM_CHAT_ID"},
				Usage:   "Telegram Group Chat ID",
			},
		},
		Commands: []*cli.Command{},
		Before: func(c *cli.Context) error {
			sigC := make(chan os.Signal)
			signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
			go func() {
				sig := <-sigC
				log.WithField("signal", sig.String()).Info("shutting down")
				cancel()
			}()

			return nil
		},
		Action: func(c *cli.Context) error {
			groupID := c.Int64("telegram-chat-id")
			var lastStatus altius.TestResult = ""

			b, err := tb.NewBot(tb.Settings{
				Token:  c.String("telegram-token"),
				Poller: &tb.LongPoller{Timeout: 10 * time.Second},
			})
			if err != nil {
				return err
			}

			t := time.NewTicker(30 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-t.C:
					log.Info("checking status of test")
					status, err := altius.GetTestResult(c.String("retrieval-code"), c.String("date-of-birth"))
					if err != nil {
						return err
					}
					log.WithFields(logrus.Fields{
						"lastStatus": lastStatus,
						"status":     status,
					}).Info("status returned")

					if status != lastStatus {
						log.WithField("group", groupID).Info("sending message to group")
						_, err := b.Send(&tb.Chat{
							ID:   groupID,
							Type: tb.ChatGroup,
						}, fmt.Sprintf("Test status has been updated: %s", status))
						if err != nil {
							return err
						}
					}

					// update our last found status
					lastStatus = status
				}
			}
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Errorf("failed to run: %v", err)
		exitCode = 1
	}
}
