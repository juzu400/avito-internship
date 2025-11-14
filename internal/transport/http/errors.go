package http

import (
	"net/http"

	"github.com/juzu400/avito-internship/internal/service"
)

func mapErrorToHTTP(err error) (status int, code string) {
	if err == nil {
		return http.StatusOK, ""
	}

	code = service.ErrorCode(err)

	switch code {
	case service.ErrCodeValidation:
		return http.StatusBadRequest, code
	case service.ErrCodeNotFound:
		return http.StatusNotFound, code
	case service.ErrCodeTeamAlreadyExists:
		return http.StatusBadRequest, code
	case service.ErrCodePullRequestAlreadyExists,
		service.ErrCodePullRequestAlreadyMerged,
		service.ErrCodeReviewerNotAssigned,
		service.ErrCodeNoReviewerCandidates:
		return http.StatusConflict, code
	default:
		return http.StatusInternalServerError, service.ErrCodeInternal
	}
}
