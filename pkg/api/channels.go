package api

import (
	"github.com/gofiber/fiber"
	"ygp/pkg/youtube"
)

// Handler is default router handler for GET /channel endpoint
func ChannelsHandler(ctx *fiber.Ctx) {
	q := ctx.FormValue("q")

	response, err := youtube.Yt.ChannelsListFromCache(q)
	if err != nil {
		l.WithError(err).Errorf("can't get any channels")
		ctx.Next(err)
	}

	err = ctx.JSON(response)
	if err != nil {
		ctx.Next(err)
	}
}
