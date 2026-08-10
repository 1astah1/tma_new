package public

import (
	"encoding/json"
	"net/http"
	"sort"

	"tma-backend/internal/handler"
)

type homeBannerSetting struct {
	ID        string `json:"id"`
	ImageURL  string `json:"image_url"`
	LinkURL   string `json:"link_url,omitempty"`
	Title     string `json:"title,omitempty"`
	SortOrder int    `json:"sort_order"`
	Active    bool   `json:"active"`
}

type homeBannerItem struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
	LinkURL  string `json:"link_url,omitempty"`
	Title    string `json:"title,omitempty"`
}

func (h *ContentHandler) loadHomeBanners(r *http.Request) ([]homeBannerSetting, error) {
	setting, err := h.settingsRepo.Get(r.Context(), "home_banners")
	if err != nil {
		return []homeBannerSetting{}, nil
	}

	value, _ := setting["value"].(string)
	var raw []homeBannerSetting
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return []homeBannerSetting{}, nil
	}

	sort.Slice(raw, func(i, j int) bool {
		if raw[i].SortOrder == raw[j].SortOrder {
			return raw[i].Title < raw[j].Title
		}
		return raw[i].SortOrder < raw[j].SortOrder
	})
	return raw, nil
}

func (h *ContentHandler) HomeBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.loadHomeBanners(r)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	items := make([]homeBannerItem, 0)
	for _, banner := range banners {
		if !banner.Active || banner.ImageURL == "" {
			continue
		}
		items = append(items, homeBannerItem{
			ID:       banner.ID,
			ImageURL: banner.ImageURL,
			LinkURL:  banner.LinkURL,
			Title:    banner.Title,
		})
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}
