package pagination

import (
	"net/http"
	"strconv"
)

// Params содержит параметры пагинации
type Params struct {
	Page    int
	PerPage int
	Limit   int
	Offset  int
}

// Meta содержит метаинформацию о результатах пагинации
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

const (
	// DefaultPerPage дефолтное количество элементов на странице
	DefaultPerPage = 20
	// MaxPerPage максимальное количество элементов на странице
	MaxPerPage = 100
)

// ParseParams парсит параметры пагинации из query параметров
// Поддерживает как page/per_page так и limit/offset стиль
// - page/per_page: страница (1-based) и количество элементов
// - limit/offset: количество элементов и смещение (0-based)
// Если оба стиля указаны, используется page/per_page
func ParseParams(r *http.Request) Params {
	p := Params{
		Page:    1,
		PerPage: DefaultPerPage,
	}

	// Проверяем page/per_page параметры
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			p.Page = page
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 {
			p.PerPage = perPage
		}
	}

	// Если page/per_page не установлены, проверяем limit/offset
	// (но page/per_page имеет приоритет если оба указаны)
	if p.Page == 1 && p.PerPage == DefaultPerPage {
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				p.Limit = limit
				p.PerPage = limit
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				p.Offset = offset
				// Если limit не был установлен, рассчитаем page на основе offset
				if p.Limit == 0 {
					p.Page = (offset / p.PerPage) + 1
				} else {
					p.Page = (offset / p.PerPage) + 1
				}
			}
		}
	}

	// Применяем ограничения
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPerPage
	}
	if p.Page < 1 {
		p.Page = 1
	}

	// Рассчитываем offset на основе page (если limit/offset не использовались)
	if p.Offset == 0 && p.Page > 1 {
		p.Offset = (p.Page - 1) * p.PerPage
	}

	return p
}

// NewMeta создает новый объект Meta с информацией о пагинации
func NewMeta(page, perPage, total int) Meta {
	totalPages := (total + perPage - 1) / perPage // Roundup division
	if total == 0 {
		totalPages = 1
	}
	return Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Response обертка для стандартизированного ответа со списком и пагинацией
type Response struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// NewResponse создает новый пагинированный ответ
func NewResponse(data interface{}, page, perPage, total int) Response {
	return Response{
		Data: data,
		Meta: NewMeta(page, perPage, total),
	}
}
