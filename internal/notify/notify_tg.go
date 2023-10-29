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
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
      bot, err := tgbotapi.NewBotAPI(n.config.TelegramToken)
      if err != nil {
        // log that we failed to init telegram bot
        log.Fatal(err)
      }
      bot.Debug = true
      log.Infof("Authorized on account %s", bot.Self.UserName)
      msgContent := fmt.Sprintf("%s `%s`\n%s %s\nОписание вакансии: %s", vacancy.TeamIcon, vacancy.Name, vacancy.RemoteIcon, vacancy.Team, vacancy.Link)
      msg := tgbotapi.NewMessage(n.config.TelegramChatId, msgContent)
      msg.ParseMode = "markdown"
      _, err = bot.Send(msg)
      if err != nil {
        log.WithFields(log.Fields{
          "vacancy": vacancy,
          "error": err,
          "chat_id": n.config.TelegramChatId,
        }).Error("failed to send message to telegram...")
        continue
      }
    }
  }
  return nil
}

func (n *tg_notifier) WelcomeMessage(ctx context.Context, vacanciesCount int) error {
  msgContent := fmt.Sprintf("Я бот Vacancies Sniffer и я буду присылать тебе уведомления о новых вакансиях с сайта aviasales.ru\n\nСейчас на сайте есть %d вакансий, но я буду следить только за новыми\n\nЕсли ты хочешь посмотреть все вакансии что есть сейчас, жми кнопку", vacanciesCount)
  bot, err := tgbotapi.NewBotAPI(n.config.TelegramToken)
  if err != nil {
    // log that we failed to init telegram bot
    log.Fatal(err)
  }
  bot.Debug = true
  log.Infof("Authorized on account %s", bot.Self.UserName)
  msg := tgbotapi.NewMessage(n.config.TelegramChatId, msgContent)
  msg.ParseMode = "markdown"

  // add button
  keyboard := tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
      tgbotapi.NewInlineKeyboardButtonURL("Посмотреть вакансии", "https://aviasales.ru/about/vacancies"),
    ),
  )
  msg.ReplyMarkup = keyboard

  _, err = bot.Send(msg)
  return err
}
