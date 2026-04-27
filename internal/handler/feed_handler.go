package handler

import (
	"net/http"

	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	feedService    service.FeedService
	errorResponder ErrorResponder
}

func NewFeedHandler(feedService service.FeedService, errorResponder ErrorResponder) *FeedHandler {
	return &FeedHandler{
		feedService:    feedService,
		errorResponder: ensureErrorResponder(errorResponder),
	}
}

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
