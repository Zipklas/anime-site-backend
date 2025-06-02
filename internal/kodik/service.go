package kodik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

type Episode struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type Season struct {
	Number   int       `json:"number"`
	Link     string    `json:"link"`
	Episodes []Episode `json:"episodes"`
}

type Video struct {
	Title       string   `json:"title"`
	Quality     string   `json:"quality"`
	VoiceActing []string `json:"translation_title"`
	URL         string   `json:"link"`
	Thumbnail   string   `json:"thumbnail"`
	Duration    int      `json:"duration"`
	Seasons     []Season `json:"seasons,omitempty"`
}

type KodikResponse struct {
	Time    string `json:"time"`
	Total   int    `json:"total"`
	Results []struct {
		Title       string `json:"title"`
		Translation struct {
			Title string `json:"title"`
			ID    int    `json:"id"`
		} `json:"translation"`
		Link      string `json:"link"`
		Quality   string `json:"quality"`
		Duration  int    `json:"duration"`
		Thumbnail string `json:"thumbnail"`
		Seasons   map[string]struct {
			Link     string            `json:"link"`
			Episodes map[string]string `json:"episodes"`
		} `json:"seasons"`
	} `json:"results"`
}

type Service struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewService(apiKey string) *Service {
	return &Service{
		apiKey:     "7f873cf964150722ed66efa5f82c018a",
		baseURL:    "https://kodikapi.com",
		httpClient: &http.Client{},
	}
}

func (s *Service) SearchAnime(ctx context.Context, title string) ([]Video, error) {
	query := url.Values{}
	query.Add("token", s.apiKey)
	query.Add("title", title)
	query.Add("types", "anime-serial,anime")
	query.Add("with_episodes", "true")

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/search?%s", s.baseURL, query.Encode()))
	if err != nil {
		return nil, fmt.Errorf("kodik search failed: %w", err)
	}
	defer resp.Body.Close()

	var searchResp KodikResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode kodik response: %w", err)
	}

	var videos []Video
	for _, result := range searchResp.Results {
		videos = append(videos, Video{
			Title:       fmt.Sprintf("%s (%s)", result.Title, result.Translation.Title),
			VoiceActing: []string{result.Translation.Title},
			URL:         result.Link,
		})
	}

	return videos, nil
}

func (s *Service) GetVideoByShikimoriID(
	ctx context.Context,
	baseParams url.Values,
) ([]Video, error) {

	query := make(url.Values)
	for k, v := range baseParams {
		query[k] = v
	}

	// Добавляем обязательные параметры
	query.Set("token", s.apiKey)
	query.Set("with_material_data", "true")

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/search?%s", s.baseURL, query.Encode()),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kodik search failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse KodikResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode kodik response: %w", err)
	}

	var videos []Video
	for _, result := range apiResponse.Results {
		video := Video{
			Title:       result.Title,
			Quality:     result.Quality,
			VoiceActing: []string{result.Translation.Title},
			URL:         result.Link,
			Thumbnail:   result.Thumbnail,
			Duration:    result.Duration,
		}

		if result.Seasons != nil {
			for seasonNum, seasonData := range result.Seasons {
				seasonNumber, _ := strconv.Atoi(seasonNum)
				season := Season{
					Number: seasonNumber,
					Link:   seasonData.Link,
				}

				for epNum, epURL := range seasonData.Episodes {
					epNumber, _ := strconv.Atoi(epNum)
					season.Episodes = append(season.Episodes, Episode{
						Number: epNumber,
						URL:    epURL,
					})
				}

				sort.Slice(season.Episodes, func(i, j int) bool {
					return season.Episodes[i].Number < season.Episodes[j].Number
				})

				video.Seasons = append(video.Seasons, season)
			}

			// Сортируем сезоны по номеру
			sort.Slice(video.Seasons, func(i, j int) bool {
				return video.Seasons[i].Number < video.Seasons[j].Number
			})
		}

		videos = append(videos, video)
	}

	return videos, nil
}
