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

package main

import (
	"context"
	"encoding/gob"
	"github.com/Ahton89/vacancies_scrapper/internal/configuration"
	"github.com/Ahton89/vacancies_scrapper/internal/worker"
	"github.com/Ahton89/vacancies_scrapper/internal/worker/types"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"os/signal"
	"sync"
	"syscall"
)

func init() {
	gob.Register(types.VacancyInfo{})
}

func main() {
	// Create a context that is cancelled when SIGINT or SIGTERM is received
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// Create a wait group to wait for all goroutines to finish
	wg := new(sync.WaitGroup)

	log.Info("Starting SALO vacancies scrapper 🦄...")

	// Initialize configuration
	scrapperConfig, err := configuration.New()
	if err != nil {
		log.WithFields(log.Fields{
			"status": "initializing failed",
			"error":  err,
		}).Fatal("Config")
	}

	// Set debug log level if debug mode is enabled
	if scrapperConfig.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Initialize cache
	scrapperCache := cache.New(cache.NoExpiration, cache.NoExpiration)

	if scrapperConfig.CacheStateFile.Exist() {
		err = scrapperCache.LoadFile(scrapperConfig.CacheStateFile.String())
		if err != nil {
			log.WithFields(log.Fields{
				"status": "failed",
				"error":  err,
			}).Fatalf("Loading cache from file %s...", scrapperConfig.CacheStateFile.String())
		}
		log.Infof("Cache from file %s loaded successfully", scrapperConfig.CacheStateFile.String())
	}

	// Set first run flag
	_, firstRunExist := scrapperCache.Get("flag__first_run")
	if !firstRunExist {
		scrapperCache.Set("flag__first_run", true, cache.NoExpiration)
	}

	// Initialize worker
	scrapperWorker := worker.New(scrapperConfig, scrapperCache, wg)
	wg.Add(1)
	go scrapperWorker.Start(ctx)

	// Wait for SIGINT or SIGTERM
	<-ctx.Done()

	cancel()

	wg.Wait()

	err = scrapperCache.SaveFile(scrapperConfig.CacheStateFile.String())
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Saving cache...")
	}
	log.Infof("Cache saved successfully to file %s", scrapperConfig.CacheStateFile.String())

	log.Info("Bye 👋")
}
