package handler

import (
	"net/http"

	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// FeedHandler 处理RSS/Atom Feed订阅请求。
type FeedHandler struct {
	feedService    service.FeedService
	errorResponder ErrorResponder
}

// NewFeedHandler 创建Feed订阅处理器。
func NewFeedHandler(feedService service.FeedService, errorResponder ErrorResponder) *FeedHandler {
	return &FeedHandler{
		feedService:    feedService,
		errorResponder: ensureErrorResponder(errorResponder),
	}
}

// GetRSS 生成RSS Feed
// @Summary RSS订阅
// @Description 返回站点最新20条公开说说的RSS 2.0订阅源
// @Tags feed
// @Produce application/rss+xml
// @Produce json
// @Success 200 {string} string "RSS XML"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /feed/rss [get]
func (h *FeedHandler) GetRSS(c *gin.Context) {
	if !ensureServiceReady(c, h.feedService) {
		return
	}
	rss, err := h.feedService.GetRSS(c.Request.Context())
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}
	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(rss))
}

// GetAtom 生成Atom Feed
// @Summary Atom订阅
// @Description 返回站点最新20条公开说说的Atom订阅源
// @Tags feed
// @Produce application/atom+xml
// @Produce json
// @Success 200 {string} string "Atom XML"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /feed/atom [get]
func (h *FeedHandler) GetAtom(c *gin.Context) {
	if !ensureServiceReady(c, h.feedService) {
		return
	}
	atom, err := h.feedService.GetAtom(c.Request.Context())
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}
	c.Data(http.StatusOK, "application/atom+xml; charset=utf-8", []byte(atom))
}
