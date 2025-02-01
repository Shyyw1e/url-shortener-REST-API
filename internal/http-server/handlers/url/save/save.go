package save

import (
	"errors"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"log/slog"
	"net/http"

	resp "url-shorter-REST-API/internal/lib/api/response"
	"url-shorter-REST-API/internal/lib/logger/slpkg"
	random "url-shorter-REST-API/internal/lib/rndom"
	"url-shorter-REST-API/internal/storage"
)

// TODO: move to config
const aliasLength = 4

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
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

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength) //

		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("URL already exists"))

			return
		}

		if err != nil {
			log.Error("failed to add url", slpkg.Err(err))

			render.JSON(w, r, resp.Error("failed to add URL"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})

	}
}
