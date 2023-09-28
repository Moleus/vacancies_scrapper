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
	"context"
	"github.com/Ahton89/vacancies_scrapper/internal/configuration"
	"github.com/patrickmn/go-cache"
	"sync"
)

type worker struct {
	config configuration.Configuration
	wg     *sync.WaitGroup
	cache  *cache.Cache
}

type Worker interface {
	Start(ctx context.Context)
}
