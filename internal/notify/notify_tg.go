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
	telego "github.com/mymmrac/telego"
  tu "github.com/mymmrac/telego/telegoutil"
	log "github.com/sirupsen/logrus"
)

func NewTg(config configuration.Configuration) Notifier {
  return &tg_notifier{
    config: config,
  }
}

func (n *tg_notifier) Notify(ctx context.Context, vacancies []types.VacancyInfo) (err error) {
  for _, vacancy := range vacancies {
    select {
    case <-ctx.Done():
      return ctx.Err()
    default:
      // notify telegram
      bot, err := telego.NewBot(n.config.TelegramToken, telego.WithDefaultDebugLogger())
      if err != nil {
        // log that we failed to init telegram bot
        log.Fatal(err)
      }
      botName, err := bot.GetMyName(&telego.GetMyNameParams{})
      if err != nil {
        log.Fatal(err)
      }
      log.Infof("Authorized on account %s", botName.Name)
      msgContent := fmt.Sprintf("%s `%s`\n%s %s\nОписание вакансии: %s", vacancy.TeamIcon, vacancy.Name, vacancy.RemoteIcon, vacancy.Team, vacancy.Link)
      msg, err := bot.SendMessage(&telego.SendMessageParams{
        ChatID: telego.ChatID{ID: n.config.TelegramChatId},
        Text:   msgContent,
        ParseMode: "MarkdownV2",
      })
      if err != nil {
        log.WithFields(log.Fields{
          "vacancy": vacancy,
          "error": err,
          "chat_id": n.config.TelegramChatId,
        }).Error("failed to send message to telegram...")
        continue
      }
      log.Debug("Message sent to telegram: ", msg)
    }
  }
  return nil
}

func (n *tg_notifier) WelcomeMessage(ctx context.Context, vacanciesCount int) error {
  msgContent := fmt.Sprintf("Я бот Vacancies Sniffer и я буду присылать тебе уведомления о новых вакансиях с сайта aviasales.ru\n\nСейчас на сайте есть %d вакансий, но я буду следить только за новыми\n\nЕсли ты хочешь посмотреть все вакансии что есть сейчас, жми кнопку", vacanciesCount)
  bot, err := telego.NewBot(n.config.TelegramToken, telego.WithDefaultDebugLogger())
  if err != nil {
    // log that we failed to init telegram bot
    log.Fatal(err)
  }
  botName, err := bot.GetMyName(&telego.GetMyNameParams{})
  if err != nil {
    log.Fatal(err)
  }
  log.Infof("Authorized on account %s", botName.Name)

  // add button
  msg, err := bot.SendMessage(
    tu.Message(telego.ChatID{ID: n.config.TelegramChatId}, msgContent).WithReplyMarkup(
      tu.InlineKeyboard(
        tu.InlineKeyboardRow(
          tu.InlineKeyboardButton("Посмотреть вакансии").WithURL("https://aviasales.ru/about/vacancies"),
          ),
        ),
      ),
    )
  if err != nil {
    log.WithFields(log.Fields{
      "error": err,
      "chat_id": n.config.TelegramChatId,
    }).Error("failed to send message to telegram...")
    return err
  }
  log.Debug("Message sent to telegram: ", msg)
  return nil
}
