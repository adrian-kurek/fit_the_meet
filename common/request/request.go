// Package request hold whole logic associated with request helpers functions
package request

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/response"
	"github.com/adrian-kurek/fit_the_meet/common/validation"
)

type HTTPFunc func(w http.ResponseWriter, r *http.Request) error

func Make(f HTTPFunc, logger commoninterfaces.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if err := f(w, r); err != nil {
			apiErr := &commonerrors.API{}
			if errors.As(err, &apiErr) {
				response.Send(w, apiErr.StatusCode, map[string]string{"message": apiErr.Message})
			} else {
				response.Send(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
		}
		durationOfTheRoute := time.Since(start) / time.Millisecond
		formattedDurationOfTheRoute := strconv.FormatInt(int64(durationOfTheRoute), 10) + "ms"

		logger.Info("request closed", "stats", map[string]string{
			"method":         r.Method,
			"path":           r.URL.Path,
			"remote_address": r.RemoteAddr,
			"duration":       formattedDurationOfTheRoute,
		})
	}
}

func ReadUserIDFromToken(r *http.Request) (int, error) {
	userID, ok := r.Context().Value("id").(int)
	if !ok || userID == 0 {
		err := errors.New("failed to read user from context")
		return 0, err
	}

	return userID, nil
}

func ReadBody[T any](r *http.Request) (*T, error) {
	if r.Body == nil {
		return nil, errors.New("no request body provided")
	}
	var body T

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func ReadQueryParam(r *http.Request, queryName string) string {
	name := r.URL.Query().Get(queryName)
	return name
}

func SetContext(r *http.Request, key, data any) *http.Request {
	ctx := context.WithValue(r.Context(), key, data)
	return r.WithContext(ctx)
}

func ReadAndValidateBody[T any](r *http.Request) (*T, error) {
	reqData, err := ReadBody[T](r)
	if err != nil {
		return nil, commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return nil, err
	}
	return reqData, nil
}
