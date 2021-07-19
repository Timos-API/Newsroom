package service

import (
	"Timos-API/Newsroom/persistence"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

type TwitterService struct {
	c *anaconda.TwitterApi
}

func NewTwitterService() *TwitterService {

	anaconda.SetConsumerKey(os.Getenv("TWITTER_API_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_API_SECRET_KEY"))
	client := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	return &TwitterService{client}
}

func (s *TwitterService) SendTweet(news *persistence.News) (*anaconda.Tweet, error) {

	res, err := http.Get(news.Thumbnail)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("received non 200 response code")
	}

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(b)

	media, err := s.c.UploadMedia(encoded)
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("media_ids", strconv.Itoa(int(media.MediaID)))

	baseUrl, err := url.Parse("https://newsroom.timos.design")

	if err != nil {
		return nil, err
	}

	baseUrl.Path += "/news/" + news.Title + "." + news.NewsID.Hex()
	tweet, err := s.c.PostTweet("["+strings.ToUpper(news.Type)+"] "+news.Title+"\n"+baseUrl.String(), v)

	if err != nil {
		return nil, err
	}

	return &tweet, nil
}

func (s *TwitterService) DeleteTweet(id int64) (bool, error) {
	_, err := s.c.DeleteTweet(id, false)
	if err != nil {
		return false, err
	}
	return true, nil
}
