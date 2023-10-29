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

package configuration

import (
	"os"
	"time"
)

type Configuration struct {
	CacheStateFile       cacheStateFile `env:"CACHE_STATE_FILE" envDefault:"/Users/ahton/Downloads/vacancies-scrapper.cache"`
	Debug                bool           `env:"DEBUG" envDefault:"false"`
	ScrapeInterval       time.Duration  `env:"SCRAPE_INTERVAL" envDefault:"1m"`
	ScrapeRequestTimeout time.Duration  `env:"SCRAPE_REQUEST_TIMEOUT" envDefault:"5s"`
	ScrapeDomain         string         `env:"SCRAPE_DOMAIN,required"`
	ScrapeUrl            string         `env:"SCRAPE_URL,required"`
  TelegramToken        string         `env:"TELEGRAM_TOKEN,required"`
  TelegramChatId       int64          `env:"TELEGRAM_CHAT_ID,required"`
}

type cacheStateFile string

func (c *cacheStateFile) String() string {
	return string(*c)
}

func (c *cacheStateFile) Exist() bool {
	_, err := os.Stat(c.String())
	if err != nil {
		return false
	}
	return true
}
