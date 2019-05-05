package feed

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"ytg/pkg/youtube"

	"github.com/sirupsen/logrus"
)

const (
	ytVideoURL = "https://youtube.com/watch?v="
)

type VideosResponse struct {
	RegionCode string        `json:"regionCode"`
	Items      []VideosItems `json:"items"`
}
type VideosItems struct {
	ID      ID            `json:"id"`
	Snippet VideosSnippet `json:"snippet"`
}
type ID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}
type VideosSnippet struct {
	PublishedAt          time.Time                `json:"publishedAt"`
	ChannelID            string                   `json:"channelId"`
	Title                string                   `json:"title"`
	Description          string                   `json:"description"`
	Thumbnails           ChannelDetailsThumbnails `json:"thumbnails"`
	ChannelTitle         string                   `json:"channelTitle"`
	LiveBroadcastContent string                   `json:"liveBroadcastContent"`
}

type VideoFileDetails struct {
	ContentType   string
	ContentLength string
}

type VideoDetails struct {
	Duration time.Duration
}

type VideosDetailsResponse struct {
	Items []VideosDetailsItems `json:"items"`
}
type VideosDetailsItems struct {
	Details VideosDetailsContent `json:"contentDetails"`
}
type VideosDetailsContent struct {
	Duration string `json:"duration"`
}

func (f *Feed) getVideos() (VideosResponse, error) {
	videos := VideosResponse{}
	req, err := http.NewRequest("GET", youtube.YouTubeURL+"search", nil)
	if err != nil {
		logrus.WithError(err).Fatal("[YT] Can't create new request")
		return VideosResponse{}, err
	}
	query := req.URL.Query()
	query.Add("part", "snippet")
	query.Add("order", "date")
	query.Add("channelId", f.ChannelID)
	query.Add("maxResults", "10")
	query.Add("fields", "items(id,snippet(channelId,channelTitle,description,publishedAt,thumbnails/high,title))")
	req.URL.RawQuery = query.Encode()

	err = youtube.Request(req, &videos)
	if err != nil {
		return VideosResponse{}, err
	}
	return videos, nil
}

func (f *Feed) setVideos(videos VideosResponse) error {
	stream := make(chan Item, len(videos.Items))

	for i, video := range videos.Items {
		s := video.Snippet
		go func(video VideosItems, i int) error {
			videoURL := os.Getenv("API_URL") + "video/" + video.ID.VideoID + ".mp3"
			fileDetails, _ := getVideoFileDetails(videoURL)
			videoDetails, err := getVideoDetails(video.ID.VideoID)
			if err != nil {
				logrus.Printf("Error %+v", err)
				stream <- Item{}
				return err
			}
			stream <- Item{
				GUID:        video.ID.VideoID,
				Title:       s.Title,
				Link:        videoURL,
				Description: s.Description,
				PubDate:     s.PublishedAt.String(),
				Enclosure: Enclosure{
					URL:    videoURL,
					Length: fileDetails.ContentLength,
					Type:   fileDetails.ContentType,
				},
				ITAuthor:   f.ITAuthor,
				ITSubtitle: s.Title,
				ITSummary: ITSummary{
					Text: s.Description,
				},
				ITImage: ITImage{
					Href: getImageURL(s.Thumbnails.High.URL),
				},
				ITExplicit: "no",
				ITDuration: videoDetails.Duration.String(),
				ITOrder:    strconv.Itoa(i),
			}
			return nil
		}(video, i)

	}
	counter := 0
	for {
		if counter >= len(videos.Items) {
			break
		}

		f.addItem(<-stream)
		counter++
	}
	return nil
}

func getVideoFileDetails(videoURL string) (VideoFileDetails, error) {
	resp, err := http.Head(videoURL)
	if err != nil {
		return VideoFileDetails{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return VideoFileDetails{}, errors.New("Can't get file details for " + videoURL)
	}
	return VideoFileDetails{
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.Header.Get("Content-Length"),
	}, nil
}

func getVideoDetails(videoID string) (VideoDetails, error) {
	videoDetails := VideosDetailsResponse{}
	req, err := http.NewRequest("GET", youtube.YouTubeURL+"videos", nil)
	if err != nil {
		logrus.WithError(err).Fatal("[YT] Can't create new request")
		return VideoDetails{}, err
	}
	query := req.URL.Query()
	query.Add("part", "contentDetails")
	query.Add("id", videoID)
	query.Add("maxResults", "1")
	query.Add("fields", "items/contentDetails/duration")
	req.URL.RawQuery = query.Encode()

	err = youtube.Request(req, &videoDetails)
	if err != nil {
		return VideoDetails{}, err
	}
	if len(videoDetails.Items) != 1 {
		return VideoDetails{}, errors.New("Can't get video details")
	}
	duration, err := parseDuration(videoDetails.Items[0].Details.Duration)
	if err != nil {
		return VideoDetails{}, err
	}

	return VideoDetails{
		Duration: duration,
	}, nil
}

func parseDuration(duration string) (time.Duration, error) {
	durationString := normalizeDurationString(duration)
	return time.ParseDuration(durationString)
}

func normalizeDurationString(duration string) string {
	return strings.ToLower(strings.Replace(duration, "PT", "", 1))
}

func getImageURL(src string) string {
	return src
}
