package fetcher

import (
	"context"
	"go.tomakado.io/containers/set"
	"log"
	"main/internal/model"
	"main/internal/source"
	"strings"
	"sync"
	"time"
)

type ArticleStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourcesProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Sources interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourcesProvider

	fetcherInterval time.Duration
	filterKeywords  []string
}

func New(
	articleStorage ArticleStorage,
	sourcesProvider SourcesProvider,
	fetcherInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:        articleStorage,
		sources:         sourcesProvider,
		fetcherInterval: fetcherInterval,
		filterKeywords:  filterKeywords,
	}
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetcherInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRSSSourceFromModel(src)

		go func(source Sources) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("failed to fetch items for %s: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("failed to process items for %s: %v", source.Name(), err)
				return
			}
		}(&rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Sources, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShowBeSkipped(item) {
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
			ID:          source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fetcher) itemShowBeSkipped(item model.Item) bool {
	categoriesSet := set.New(item.Categories...)

	for _, keyword := range f.filterKeywords {
		titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)
		if categoriesSet.Contains(keyword) || titleContainsKeyword {
			return true
		}
	}
	return false
}
