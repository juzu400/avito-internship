package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/juzu400/avito-internship/internal/repository"
	"github.com/juzu400/avito-internship/internal/repository/mocks"
	"github.com/juzu400/avito-internship/internal/service"
)

func TestRouter_AllRoutesRegistered(t *testing.T) {
	log := newTestLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	repos := &repository.Repositories{Users: userRepo, Teams: teamRepo, PullRequests: prRepo}
	services := service.NewServices(log, repos)

	r := NewRouter(log, services)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"POST", "/team/add"},
		{"GET", "/team/get"},
		{"POST", "/users/setIsActive"},
		{"GET", "/users/getReview"},
		{"POST", "/pullRequest/create"},
		{"POST", "/pullRequest/merge"},
		{"POST", "/pullRequest/reassign"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			r.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Fatalf("route %s %s is not registered (got 404)", tt.method, tt.path)
			}
		})
	}
}

func TestRouter_UnknownRouteReturns404(t *testing.T) {
	log := newTestLogger()
	services := service.NewServices(log, &repository.Repositories{})
	r := NewRouter(log, services)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", rr.Code)
	}
}
