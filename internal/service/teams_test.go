package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/repository/mocks"
)

func newTestTeamsService(ctrl *gomock.Controller) (*TeamsService, *mocks.MockTeamRepository) {
	teamRepo := mocks.NewMockTeamRepository(ctrl)

	svc := &TeamsService{
		log:   newTestLogger(),
		teams: teamRepo,
	}

	return svc, teamRepo
}

func TestTeamsService_UpsertTeam_ValidationNilTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	err := svc.UpsertTeam(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_UpsertTeam_ValidationEmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name:    "",
		Members: []domain.User{},
	}

	err := svc.UpsertTeam(context.Background(), team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_UpsertTeam_ValidationEmptyMemberID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{
				ID:       "",
				Username: "NoID",
				IsActive: true,
			},
		},
	}

	err := svc.UpsertTeam(context.Background(), team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_UpsertTeam_ValidationDuplicateMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{
				ID:       "u1",
				Username: "Alice",
				IsActive: true,
			},
			{
				ID:       "u1",
				Username: "Alice2",
				IsActive: true,
			},
		},
	}

	err := svc.UpsertTeam(context.Background(), team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_UpsertTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{
				ID:       "u1",
				Username: "Alice",
				IsActive: true,
			},
			{
				ID:       "u2",
				Username: "Bob",
				IsActive: true,
			},
		},
	}

	userID1 := domain.UserID("u1")
	userID2 := domain.UserID("u2")

	gomock.InOrder(
		teamRepo.EXPECT().
			GetTeamsByMemberIDs(gomock.Any(), []domain.UserID{userID1, userID2}).
			Return(map[domain.UserID]*domain.Team{}, nil),

		teamRepo.EXPECT().
			UpsertTeam(gomock.Any(), team).
			Return(nil),
	)

	err := svc.UpsertTeam(context.Background(), team)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestTeamsService_UpsertTeam_RepoErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{
				ID:       "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	repoErr := errors.New("db error")
	userID := domain.UserID("u1")

	gomock.InOrder(
		teamRepo.EXPECT().
			GetTeamsByMemberIDs(gomock.Any(), []domain.UserID{userID}).
			Return(map[domain.UserID]*domain.Team{}, nil),

		teamRepo.EXPECT().
			UpsertTeam(gomock.Any(), team).
			Return(repoErr),
	)

	err := svc.UpsertTeam(context.Background(), team)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}

func TestTeamsService_GetByName_ValidationEmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	team, err := svc.GetByName(context.Background(), "")
	if team != nil {
		t.Fatalf("expected nil team, got %#v", team)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_GetByName_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	expected := &domain.Team{
		Name: "backend",
	}

	teamRepo.EXPECT().
		GetByName(gomock.Any(), "backend").
		Return(expected, nil)

	got, err := svc.GetByName(context.Background(), "backend")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != expected {
		t.Fatalf("expected team pointer %p, got %p", expected, got)
	}
}

func TestTeamsService_GetByName_RepoErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	repoErr := errors.New("not found")

	teamRepo.EXPECT().
		GetByName(gomock.Any(), "backend").
		Return(nil, repoErr)

	got, err := svc.GetByName(context.Background(), "backend")
	if got != nil {
		t.Fatalf("expected nil team, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}

func TestTeamsService_GetByMemberID_ValidationEmptyUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := newTestTeamsService(ctrl)

	team, err := svc.GetByMemberID(context.Background(), "")
	if team != nil {
		t.Fatalf("expected nil team, got %#v", team)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTeamsService_GetByMemberID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	userID := domain.UserID("u1")
	expected := &domain.Team{Name: "backend"}

	teamRepo.EXPECT().
		GetByMemberID(gomock.Any(), userID).
		Return(expected, nil)

	got, err := svc.GetByMemberID(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != expected {
		t.Fatalf("expected team pointer %p, got %p", expected, got)
	}
}

func TestTeamsService_GetByMemberID_RepoErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	userID := domain.UserID("u1")
	repoErr := errors.New("db error")

	teamRepo.EXPECT().
		GetByMemberID(gomock.Any(), userID).
		Return(nil, repoErr)

	got, err := svc.GetByMemberID(context.Background(), userID)
	if got != nil {
		t.Fatalf("expected nil team, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}

func TestTeamsService_UpsertTeam_UserAlreadyInAnotherTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, teamRepo := newTestTeamsService(ctrl)

	team := &domain.Team{
		Name: "frontend",
		Members: []domain.User{
			{ID: "u1", Username: "Alice", IsActive: true},
		},
	}

	userID := domain.UserID("u1")

	teamRepo.EXPECT().
		GetTeamsByMemberIDs(gomock.Any(), []domain.UserID{userID}).
		Return(map[domain.UserID]*domain.Team{
			userID: {Name: "backend"},
		}, nil)

	err := svc.UpsertTeam(context.Background(), team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
