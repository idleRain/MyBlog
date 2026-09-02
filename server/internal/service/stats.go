// Package service 业务逻辑层
package service

import (
	"time"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// StatsServiceInterface 站点统计服务接口
type StatsServiceInterface interface {
	GetOverview() (*StatsOverview, error)
	GetArticleViewsTrend(days int) (*TrendResponse, error)
}

// StatsOverview 站点统计概览
type StatsOverview struct {
	ArticleCount   int64  `json:"articleCount"`
	PublishedCount int64  `json:"publishedCount"`
	TotalViews     uint64 `json:"totalViews"`
	TotalLikes     uint64 `json:"totalLikes"`
	CommentCount   int64  `json:"commentCount"`
	UserCount      int64  `json:"userCount"`
	CategoryCount  int64  `json:"categoryCount"`
	TagCount       int64  `json:"tagCount"`
}

// TrendResponse 时间序列统计响应
type TrendResponse struct {
	Dates  []string `json:"dates"`
	Values []uint64 `json:"values"`
}

// StatsService 站点统计服务实现
type StatsService struct {
	statsRepo repository.StatsRepositoryInterface
}

// NewStatsService 创建站点统计服务实例
func NewStatsService(statsRepo repository.StatsRepositoryInterface) StatsServiceInterface {
	return &StatsService{
		statsRepo: statsRepo,
	}
}

// GetOverview 获取站点统计概览。
func (s *StatsService) GetOverview() (*StatsOverview, error) {
	publishedCount, err := s.statsRepo.CountArticles(model.ArticleStatusPublished)
	if err != nil {
		return nil, err
	}
	articleCount, err := s.statsRepo.CountArticles("")
	if err != nil {
		return nil, err
	}
	totalViews, err := s.statsRepo.SumArticleViews()
	if err != nil {
		return nil, err
	}
	totalLikes, err := s.statsRepo.SumArticleLikes()
	if err != nil {
		return nil, err
	}
	commentCount, err := s.statsRepo.CountComments()
	if err != nil {
		return nil, err
	}
	userCount, err := s.statsRepo.CountUsers()
	if err != nil {
		return nil, err
	}
	categoryCount, err := s.statsRepo.CountCategories()
	if err != nil {
		return nil, err
	}
	tagCount, err := s.statsRepo.CountTags()
	if err != nil {
		return nil, err
	}

	return &StatsOverview{
		ArticleCount:   articleCount,
		PublishedCount: publishedCount,
		TotalViews:     totalViews,
		TotalLikes:     totalLikes,
		CommentCount:   commentCount,
		UserCount:      userCount,
		CategoryCount:  categoryCount,
		TagCount:       tagCount,
	}, nil
}

// GetArticleViewsTrend 获取指定天数内的文章浏览量趋势，缺失日期补零。
func (s *StatsService) GetArticleViewsTrend(days int) (*TrendResponse, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	startDate := time.Now().AddDate(0, 0, -(days - 1))
	stats, err := s.statsRepo.GetContentStats(model.ContentTypeArticle, model.StatTypeDailyViews, startDate)
	if err != nil {
		return nil, err
	}

	// 按日期建立索引，便于逐日填充。
	valueByDate := make(map[string]uint64, len(stats))
	for _, stat := range stats {
		valueByDate[stat.StatDate.Format("2006-01-02")] = uint64(stat.StatValue)
	}

	response := &TrendResponse{
		Dates:  make([]string, 0, days),
		Values: make([]uint64, 0, days),
	}
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateKey := date.Format("2006-01-02")
		response.Dates = append(response.Dates, dateKey)
		response.Values = append(response.Values, valueByDate[dateKey])
	}

	return response, nil
}
