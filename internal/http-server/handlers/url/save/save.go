package save

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"log/slog"
	"net/http"

	resp "url-shorter-REST-API/internal/lib/api/response"
	"url-shorter-REST-API/internal/lib/logger/slpkg"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	Save(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", slpkg.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", slpkg.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}
	}
}
