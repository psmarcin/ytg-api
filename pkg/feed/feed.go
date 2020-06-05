package feed

import (
	"github.com/eduncan911/podcast"
	"github.com/sirupsen/logrus"
	"sort"
)

var l = logrus.WithField("source", "feed")

type Feed struct {
	ChannelID string
	Content   podcast.Podcast
}

func (f *Feed) Serialize() ([]byte, error) {
	f.Content.Items = sortByOrder(f.Content.Items)
	return f.Content.Bytes(), nil
}

func (f *Feed) SortVideos() {
	f.Content.Items = sortByOrder(f.Content.Items)
}

func New(channelID string) Feed {
	feed := Feed{
		ChannelID: channelID,
	}
	return feed
}

func sortByOrder(items []*podcast.Item) []*podcast.Item {
	sort.Slice(items, func(i, j int) bool {
		return items[i].PubDate.After(*items[j].PubDate)
	})
	return items
}
