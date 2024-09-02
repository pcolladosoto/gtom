package main

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type (
	response struct {
		StatusCode int    `json:"statusCode"`
		Data       string `json:"data"`
		Error      string `json:"err"`
	}

	// Note from and to are specified as ISO 8601 dates as seen
	// on https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to
	findReq struct {
		From       string `json:"from"`
		To         string `json:"to"`
		Collection string `json:"collection"`
		Filter     string `json:"filter"`
	}

	// Check https://echo.labstack.com/docs/request#validate-data
	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func NewServer(database *db) *echo.Echo {
	e := echo.New()

	// Hide Echo's banners and overall publicity xD
	e.HideBanner = true
	e.HidePort = true

	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/find", func(c echo.Context) error {
		var fr findReq
		if err := c.Bind(&fr); err != nil {
			return c.JSON(http.StatusBadRequest, response{
				StatusCode: http.StatusBadRequest,
				Data:       "bad request",
				Error:      err.Error(),
			})
		}

		if err := c.Validate(fr); err != nil {
			return c.JSON(http.StatusBadRequest, response{
				StatusCode: http.StatusBadRequest,
				Data:       "bad request",
				Error:      err.Error(),
			})
		}

		slog.Debug("find request", "req", fr)

		resp, err := database.find(fr.Collection, fr.From, fr.To, fr.Filter)
		if err != nil {
			return c.JSON(http.StatusBadRequest, response{
				StatusCode: http.StatusBadRequest,
				Data:       "bad request",
				Error:      err.Error(),
			})
		}

		return c.JSON(http.StatusOK, response{
			StatusCode: http.StatusOK,
			Data:       string(resp),
		})
	})

	return e
}
