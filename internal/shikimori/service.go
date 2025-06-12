package shikimori

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/machinebox/graphql"
)

type Service struct {
	graphqlClient *graphql.Client
}

func NewService() *Service {
	// Инициализация клиента для запросов к API Shikimori
	graphqlClient := graphql.NewClient("https://shikimori.one/api/graphql")

	return &Service{
		graphqlClient: graphqlClient,
	}
}

func (s *Service) SearchAnime(ctx context.Context, search string, limit int) ([]Anime, error) {
	// Формируем запрос GraphQL
	req := graphql.NewRequest(`
        query($search: String!, $limit: Int!) {
          animes(search: $search, limit: $limit) {
            id
            malId
            name
            russian
			rating
            score
            description
			poster {
                id
                originalUrl
                mainUrl
            }
			genres {
                id
                name
                russian
                kind
            }
          }
        }
    `)

	// Параметры запроса
	req.Var("search", search)
	req.Var("limit", limit)

	// Обязательные заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://shikimori.one")
	req.Header.Set("User-Agent", "shiki_api_test")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SHIKIMORI_TOKEN"))

	log.Printf("%s", req)
	// Структура для ответа
	var resp AnimeSearchResponseData

	// Логируем перед запросом
	log.Printf("Параметры запроса: search=%s, limit=%d", search, limit)

	// Выполняем запрос
	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return nil, err
	}
	log.Printf("Полученные данные: %+v", resp)
	log.Printf("Полученные данные: %+v", resp.Animes)

	// Возвращаем найденные аниме
	return resp.Animes, nil
}
func (s *Service) GetTopAnime(ctx context.Context, limit int, page int, genre string) ([]Anime, error) {
	req := graphql.NewRequest(`
		query($limit: PositiveInt = 30, $page: PositiveInt, $genre: String) {
			animes(limit: $limit, page: $page, order: ranked, genre: $genre) {
				id
				malId
				name
				russian
				score
				description
				poster {
                id
                originalUrl
                mainUrl
            }
			genres {
                id
                name
                russian
                kind
            }
			}
		}
	`)

	req.Var("limit", limit)
	req.Var("page", page)
	if genre != "" {
		req.Var("genre", genre)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://shikimori.one")
	req.Header.Set("User-Agent", "shiki_api_test")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SHIKIMORI_TOKEN"))

	var resp AnimeSearchResponseData
	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		log.Printf("Ошибка запроса топовых аниме: %v", err)
		return nil, err
	}

	log.Printf("→ Загружено %d топ-аниме на странице %d", len(resp.Animes), page)
	return resp.Animes, nil
}

func (s *Service) GetAnimeByID(ctx context.Context, id string) (*Anime, error) {
	req := graphql.NewRequest(`
        query($ids: String) {
        animes(ids: $ids) {
        id
        name
        russian
        episodes
        score
		description
		poster {
                id
                originalUrl
                mainUrl
            }
		genres {
            	id
                name
                russian
                kind
        }	
        }
       }
    `)

	req.Var("ids", id)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://shikimori.one")
	req.Header.Set("User-Agent", "shiki_api_test")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SHIKIMORI_TOKEN"))

	var resp AnimeSearchResponseData

	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		log.Printf("Ошибка запроса аниме по ID: %v", err)
		return nil, err
	}

	return &resp.Animes[0], nil
}

func (s *Service) GetAnimesByIDs(ctx context.Context, ids []string) ([]Anime, error) {
	req := graphql.NewRequest(`
        query($ids: [String!]!) {
            animes(ids: $ids) {
                id
                name
                russian
				description
                score
                status
				poster {
                id
                originalUrl
                mainUrl
            }
			genres {
                id
                name
                russian
                kind
            }
            }
        }
    `)

	req.Var("ids", ids)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://shikimori.one")
	req.Header.Set("User-Agent", "shiki_api_test")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SHIKIMORI_TOKEN"))

	var resp AnimeSearchResponseData

	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Animes, nil
}

func (s *Service) GetNewReleases(ctx context.Context, limit int) ([]Anime, error) {
	req := graphql.NewRequest(`
	query($limit: Int!, $season: SeasonString!, $status: AnimeStatusString!) {
		animes(
			limit: $limit,
			order: popularity,
			season: $season,
			status: $status
		) {
			id
			name
			russian
			score
			poster {
				originalUrl
				mainUrl
			}
			airedOn {
				year
				month
				day
				date
			}
			genres {
                id
                name
                russian
                kind
        }
		}
	}
`)
	season := ""
	currentTime := time.Now()
	month := int(currentTime.Month())
	year := int(currentTime.Year())
	var seasonPart string
	switch {
	case month >= 1 && month <= 3:
		seasonPart = "winter"
	case month >= 4 && month <= 6:
		seasonPart = "spring"
	case month >= 7 && month <= 9:
		seasonPart = "summer"
	case month >= 10 && month <= 12:
		seasonPart = "fall"
	}
	season = fmt.Sprintf("%s_%d", seasonPart, year)
	req.Var("limit", limit)
	req.Var("season", season)
	req.Var("status", "ongoing")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SHIKIMORI_TOKEN"))

	var resp AnimeSearchResponseData
	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Animes, nil
}
