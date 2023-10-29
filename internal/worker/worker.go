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

package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Ahton89/vacancies_scrapper/internal/configuration"
	"github.com/Ahton89/vacancies_scrapper/internal/notify"
	"github.com/Ahton89/vacancies_scrapper/internal/worker/types"
	"github.com/antchfx/htmlquery"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

func New(config configuration.Configuration, scrapperCache *cache.Cache, wg *sync.WaitGroup) Worker {

	return &worker{
		config: config,
		wg:     wg,
		cache:  scrapperCache,
	}
}

func (w *worker) Start(ctx context.Context) {
	defer log.WithFields(log.Fields{
		"name":  "scrapper",
		"state": "stopped",
	}).Info("Worker")

	defer w.wg.Done()

	log.WithFields(log.Fields{
		"name":  "scrapper",
		"state": "started",
	}).Info("Worker")

	ticker := time.NewTicker(w.config.ScrapeInterval)

	for {
		// First run
		firstRun, _ := w.cache.Get("flag__first_run")

		// Scrape vacancies
		vacancies, err := w.scrape(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Scrape")
			continue
		}

		newVacancies := make([]types.VacancyInfo, 0)

		// Add new vacancies to cache
		for _, vacancy := range vacancies {
			select {
			case <-ctx.Done():
				return
			default:
				vacancyKey := fmt.Sprintf("vacancy__%s", vacancy.Id)

				_, exist := w.cache.Get(vacancyKey)
				if !exist {
					w.cache.Set(vacancyKey, vacancy, cache.NoExpiration)
					log.WithFields(log.Fields{
						"name":   vacancy.Name,
						"team":   vacancy.Team,
						"link":   vacancy.Link,
						"id":     vacancy.Id,
						"added":  time.Unix(vacancy.Added, 0),
						"remote": vacancy.Remote,
					}).Info("Vacancy added to cache")

					if !firstRun.(bool) {
						newVacancies = append(newVacancies, vacancy)
					}
				}
			}
		}

		// Delete vacancies that were removed from the site
		w.removeDeletedVacancies(ctx, vacancies)

		// Send new vacancies to Slack
		notifier := notify.NewTg(w.config)

		if firstRun.(bool) {
			// Send welcome message
			err = notifier.WelcomeMessage(ctx, len(vacancies))
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("WelcomeMessage")
			} else {
				log.WithFields(log.Fields{
					"count": len(vacancies),
				}).Info("Welcome message sent to Slack")
			}
			// Set first run flag
			w.cache.Set("flag__first_run", false, cache.NoExpiration)
		} else {
			// Send new vacancies
			if len(newVacancies) > 0 {
				err = notifier.Notify(ctx, newVacancies)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Notify")
				} else {
					log.WithFields(log.Fields{
						"count": len(newVacancies),
					}).Info("Vacancies sent to Slack")
				}

				// Save cache to file
				err = w.cache.SaveFile(w.config.CacheStateFile.String())
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Saving cache...")
				} else {
					log.WithFields(log.Fields{
						"file": w.config.CacheStateFile.String(),
					}).Info("Cache saved successfully")
				}
			}
		}

		// Done
		log.WithFields(log.Fields{"timestamp": time.Now().Format("02.01.2006 15:04:05")}).Debug("Scrapping done")

		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
		}
	}
}

func (w *worker) scrape(ctx context.Context) ([]types.VacancyInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, w.config.ScrapeRequestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/%s", w.config.ScrapeDomain, w.config.ScrapeUrl)
	vacancies := make([]types.VacancyInfo, 0)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	vacanciesPage, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	vacanciesList, err := htmlquery.QueryAll(vacanciesPage, "//a[@class='vacancies_vacancy']")

	for _, vacancy := range vacanciesList {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			vacancyElem := htmlquery.FindOne(vacancy, "//a")

			vacancyLink := htmlquery.SelectAttr(vacancyElem, "href")
			vacancyName := htmlquery.InnerText(htmlquery.FindOne(vacancyElem, "//p[@class='vacancies_vacancy__name']"))
			vacancyTeam := htmlquery.InnerText(htmlquery.FindOne(vacancyElem, "//div[@class='team']"))
			vacancyId := path.Base(vacancyLink)
			vacanciesClasses := strings.Split(htmlquery.SelectAttr(htmlquery.FindOne(vacancyElem, "//div"), "class"), " ")

			teamIcon := ":aviasales_blue:"
			if len(vacanciesClasses) > 1 {
				teamIcon = ":tp_true2:"
			}

			team := strings.Replace(strings.Trim(vacancyTeam, " "), "  /  ", " / ", -1)

			remote := strings.HasPrefix(team, "Remote")

			remoteIcon := ":globe_with_meridians:"
			if !remote {
				remoteIcon = ":house:"
			}

			vacancies = append(vacancies, types.VacancyInfo{
				Name:       vacancyName,
				Team:       team,
				TeamIcon:   teamIcon,
				Link:       fmt.Sprintf("%s%s", w.config.ScrapeDomain, vacancyLink),
				Id:         vacancyId,
				Added:      time.Now().Unix(),
				Remote:     remote,
				RemoteIcon: remoteIcon,
			})
		}
	}

	return vacancies, nil
}

func (w *worker) removeDeletedVacancies(ctx context.Context, currentVacancies []types.VacancyInfo) {
	// Get all vacancies from cache
	vacanciesListCache := w.cache.Items()
	for vacancyId, VacancyValue := range vacanciesListCache {
		if strings.HasPrefix(vacancyId, "vacancy__") {
			select {
			case <-ctx.Done():
				return
			default:
				vacancyExists := false

				for _, currentVacancy := range currentVacancies {
					currentVacancyId := fmt.Sprintf("vacancy__%s", currentVacancy.Id)

					if vacancyId == currentVacancyId {
						vacancyExists = true
						break
					}
				}

				if !vacancyExists {
					w.cache.Delete(vacancyId)
					log.WithFields(log.Fields{
						"name":  VacancyValue.Object.(types.VacancyInfo).Name,
						"team":  VacancyValue.Object.(types.VacancyInfo).Team,
						"link":  VacancyValue.Object.(types.VacancyInfo).Link,
						"added": time.Unix(VacancyValue.Object.(types.VacancyInfo).Added, 0),
					}).Info("Vacancy deleted from cache")
				}
			}
		}
	}
}
