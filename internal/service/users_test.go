package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/repository/mocks"
)

func newTestUsersService(ctrl *gomock.Controller) (*UsersService, *mocks.MockUserRepository, *mocks.MockPullRequestRepository) {
	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	svc := &UsersService{
		log:   newTestLogger(),
		users: userRepo,
		prs:   prRepo,
	}

	return svc, userRepo, prRepo
}

func TestUsersService_SetIsActive_ValidationEmptyUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, _ := newTestUsersService(ctrl)

	err := svc.SetIsActive(context.Background(), "", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestUsersService_SetIsActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, userRepo, _ := newTestUsersService(ctrl)

	userID := domain.UserID("u1")

	userRepo.EXPECT().
		SetIsActive(gomock.Any(), userID, true).
		Return(nil)

	err := svc.SetIsActive(context.Background(), userID, true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUsersService_SetIsActive_RepoErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, userRepo, _ := newTestUsersService(ctrl)

	userID := domain.UserID("u1")
	repoErr := errors.New("db error")

	userRepo.EXPECT().
		SetIsActive(gomock.Any(), userID, false).
		Return(repoErr)

	err := svc.SetIsActive(context.Background(), userID, false)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}

func TestUsersService_GetReviews_ValidationEmptyUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, _ := newTestUsersService(ctrl)

	prs, err := svc.GetReviews(context.Background(), "")
	if prs != nil {
		t.Fatalf("expected nil prs, got %#v", prs)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestUsersService_GetReviews_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, userRepo, prRepo := newTestUsersService(ctrl)

	userID := domain.UserID("u1")

	pr1 := &domain.PullRequest{}
	pr2 := &domain.PullRequest{}
	expected := []*domain.PullRequest{pr1, pr2}

	gomock.InOrder(
		userRepo.EXPECT().
			GetByID(gomock.Any(), userID).
			Return(&domain.User{
				ID:       userID,
				Username: "user-1",
				IsActive: true,
			}, nil),

		prRepo.EXPECT().
			ListByReviewer(gomock.Any(), userID).
			Return(expected, nil),
	)

	got, err := svc.GetReviews(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d prs, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("expected prs[%d] == %p, got %p", i, expected[i], got[i])
		}
	}
}

func TestUsersService_GetReviews_RepoErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, userRepo, prRepo := newTestUsersService(ctrl)

	userID := domain.UserID("u1")
	repoErr := errors.New("db error")

	gomock.InOrder(
		userRepo.EXPECT().
			GetByID(gomock.Any(), userID).
			Return(&domain.User{
				ID:       userID,
				Username: "user-1",
				IsActive: true,
			}, nil),

		prRepo.EXPECT().
			ListByReviewer(gomock.Any(), userID).
			Return(nil, repoErr),
	)

	got, err := svc.GetReviews(context.Background(), userID)
	if got != nil {
		t.Fatalf("expected nil prs, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}

func TestUsersService_GetReviews_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, userRepo, prRepo := newTestUsersService(ctrl)

	userID := domain.UserID("u-missing")

	userRepo.EXPECT().
		GetByID(gomock.Any(), userID).
		Return(nil, domain.ErrNotFound)

	prRepo.EXPECT().
		ListByReviewer(gomock.Any(), gomock.Any()).
		Times(0)

	prs, err := svc.GetReviews(context.Background(), userID)
	if prs != nil {
		t.Fatalf("expected nil prs, got %#v", prs)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
