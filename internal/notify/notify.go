/*
 * Copyright (C) 2023 Ahton
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package notify

import (
	"context"
	"fmt"
	"github.com/Ahton89/vacancies_scrapper/internal/configuration"
	"github.com/Ahton89/vacancies_scrapper/internal/worker/types"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"math/rand"
	"time"
)

func New(config configuration.Configuration) Notifier {
	return &notifier{
		config: config,
	}
}

func (n *notifier) Notify(ctx context.Context, vacancies []types.VacancyInfo) (err error) {
	for _, vacancy := range vacancies {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create blocks
			blocks := make([]slack.Block, 0)
			// Header
			header := slack.NewHeaderBlock(
				&slack.TextBlockObject{
					Type:  slack.PlainTextType,
					Text:  ":hidog: Новая вакансия детектед",
					Emoji: true,
				},
			)
			// Body
			body := slack.NewSectionBlock(
				&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: fmt.Sprintf(
						"%s `%s`\n%s %s\n",
						vacancy.TeamIcon,
						vacancy.Name,
						vacancy.RemoteIcon,
						vacancy.Team,
					),
				}, nil, nil)
			// Detailed description
			detailedDescription := slack.NewContextBlock(
				"",
				slack.NewTextBlockObject(
					slack.MarkdownType,
					"Подробное описание вакансии:",
					false,
					false,
				),
			)
			// Footer
			footer := slack.NewActionBlock(
				"",
				&slack.ButtonBlockElement{
					Type: slack.METButton,
					Text: &slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Посмотреть на сайте",
					},
					URL:   vacancy.Link,
					Style: "primary",
				},
			)

			// Add blocks to message
			blocks = append(blocks, header, body, detailedDescription, footer)

			// Create message
			message := slack.WebhookMessage{
				Attachments: []slack.Attachment{
					{
						Color: "#75FB4C",
						Blocks: slack.Blocks{
							BlockSet: blocks,
						},
					},
				},
			}

			// Try to send message
			for i := 1; i < n.config.SlackMaxRetry; i++ {
				err = slack.PostWebhookContext(ctx, n.config.SlackWebhook, &message)
				if err != nil {
					log.WithFields(log.Fields{
						"status": "failed and retrying",
						"error":  err,
						"try":    i,
						"max":    n.config.SlackMaxRetry,
					}).Error("Sending message to Slack...")

					// Randomly sleep from 1 to 5 seconds
					err = n.randomSleep(ctx)
					if err != nil {
						break
					}

					// Retry
					continue
				}
				break
			}
		}
	}
	return
}

func (n *notifier) randomSleep(ctx context.Context) error {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	sleepDuration := time.Duration(rand.Intn(5) + 1)

	select {
	case <-time.After(sleepDuration * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (n *notifier) WelcomeMessage(ctx context.Context, vacanciesCount int) error {
	// Create blocks
	blocks := make([]slack.Block, 0)
	// Header
	header := slack.NewHeaderBlock(
		&slack.TextBlockObject{
			Type:  slack.PlainTextType,
			Text:  ":hidog: Привет котлега!",
			Emoji: true,
		},
	)
	// Body
	body := slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("Я бот Vacancies Sniffer и я буду присылать тебе уведомления о новых вакансиях с сайта aviasales.ru :aviasales_times:\n\nСейчас на сайте есть %d вакансий, но я буду следить только за новыми :hug:\n\nЕсли ты хочешь посмотреть все вакансии что есть сейчас, жми кнопку :point_down:", vacanciesCount),
		},
		nil,
		nil,
	)
	// Button
	button := slack.NewActionBlock(
		"",
		&slack.ButtonBlockElement{
			Type: slack.METButton,
			Text: &slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Посмотреть все вакансии",
			},
			URL:   "https://aviasales.ru/about/vacancies",
			Style: "primary",
		},
	)

	// Add blocks to message
	blocks = append(blocks, header, body, button)

	// Create message
	message := slack.WebhookMessage{
		Attachments: []slack.Attachment{
			{
				Color: "#5C9DDB",
				Blocks: slack.Blocks{
					BlockSet: blocks,
				},
			},
		},
	}

	// Send message
	return slack.PostWebhookContext(ctx, n.config.SlackWebhook, &message)
}
