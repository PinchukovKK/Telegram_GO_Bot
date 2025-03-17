package source

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"
	"main/internal/model"
)

type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

func NewRSSSourceFromModel(m model.Source) RSSSource {
	return RSSSource{
		URL:        m.FeedUrl,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

func (s *RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := rss.Fetch(s.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		}
	}), nil

}

func (s *RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var (
		feedChan = make(chan *rss.Feed)
		errChan  = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errChan <- err
			return
		}
		feedChan <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case feed := <-feedChan:
		return feed, nil
	}
}

func (s RSSSource) ID() int64 {
	return s.SourceID
}

func (s RSSSource) Name() string {
	return s.SourceName
}
