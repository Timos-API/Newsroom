package service

import (
	"Timos-API/Newsroom/helper"
	"Timos-API/Newsroom/persistence"
	"context"

	"github.com/Timos-API/transformer"
	"github.com/go-playground/validator"
)

type NewsService struct {
	p *persistence.NewsPersistor
	t *TwitterService
}

func NewNewsService(p *persistence.NewsPersistor, t *TwitterService) *NewsService {
	return &NewsService{p, t}
}

func (s *NewsService) GetAllNews(ctx context.Context, query *string, limit *int, skip *int) (*[]persistence.News, error) {
	return s.p.GetByQuery(ctx, query, limit, skip)
}

func (s *NewsService) GetFeaturedNews(ctx context.Context) (*[]persistence.News, error) {
	return s.p.GetFeatured(ctx)
}

func (s *NewsService) GetProjects(ctx context.Context) (*[]interface{}, error) {
	return s.p.GetProjects(ctx)
}

func (s *NewsService) GetProjectNews(ctx context.Context, projectId string, query *string, limit *int, skip *int) (*[]persistence.News, error) {
	return s.p.GetProjectNews(ctx, projectId, query, limit, skip)
}

func (s *NewsService) GetNews(ctx context.Context, newsId string) (*persistence.News, error) {
	return s.p.GetById(ctx, newsId)
}

func (s *NewsService) DeleteNews(ctx context.Context, newsId string) (bool, error) {
	success, tweetId, err := s.p.DeleteNews(ctx, newsId)
	if err != nil {
		return false, err
	}
	if success {
		s.t.DeleteTweet(tweetId)
	}
	return success, err
}

func (s *NewsService) PatchNews(ctx context.Context, newsId string, update persistence.News) (*persistence.News, error) {
	cleaned := transformer.Clean(update, "update")
	return s.p.PatchNews(ctx, newsId, cleaned)
}

func (s *NewsService) PostNews(ctx context.Context, content persistence.News) (*persistence.News, error) {
	validate := validator.New()
	err := validate.Struct(content)
	if err != nil {
		return nil, err
	}

	cleaned := transformer.Clean(content, "create")
	cleaned.(map[string]interface{})["timestamp"] = helper.CurrentTimeMillis()

	news, err := s.p.PostNews(ctx, cleaned)

	if err != nil {
		return nil, err
	}

	tweet, err := s.t.SendTweet(news)
	if err != nil || tweet == nil {
		s.DeleteNews(ctx, news.NewsID.Hex())
		return nil, err
	}

	tnews, err := s.p.SetTweetId(ctx, news.NewsID.Hex(), tweet.Id)
	if err != nil {
		return news, err
	}

	return tnews, nil
}
