package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/repository/mocks"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPullRequestService_Create_AssignsReviewersFromTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)

	authorID := domain.UserID("author")

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: "u1", Username: "U1", IsActive: true},
			{ID: "u2", Username: "U2", IsActive: false},
		},
	}

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), authorID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.PullRequest{})).
		DoAndReturn(func(_ context.Context, pr *domain.PullRequest) error {
			if pr.Status != domain.PRStatusOpen {
				t.Errorf("expected status %q, got %q", domain.PRStatusOpen, pr.Status)
			}
			if pr.AuthorID != authorID {
				t.Errorf("expected author %q, got %q", authorID, pr.AuthorID)
			}
			if len(pr.AssignedReviewers) == 0 || len(pr.AssignedReviewers) > 2 {
				t.Errorf("expected 1-2 reviewers, got %d", len(pr.AssignedReviewers))
			}
			for _, id := range pr.AssignedReviewers {
				if id == authorID {
					t.Errorf("author must not be in reviewers")
				}
				if id == "u2" {
					t.Errorf("inactive member must not be in reviewers")
				}
			}
			return nil
		})

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: userRepo,
		teams: teamRepo,
		prs:   prRepo,
	}

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")

	pr, err := svc.Create(ctx, prID, "Test PR", authorID)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if pr.ID != prID {
		t.Errorf("expected id %q, got %q", prID, pr.ID)
	}
	if pr.Name != "Test PR" {
		t.Errorf("expected name %q, got %q", "Test PR", pr.Name)
	}
}

func TestPullRequestService_Merge_Idempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)

	now := time.Now().UTC()
	prID := domain.PullRequestID("pr-1")

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{"u1"},
		CreatedAt:         now,
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	prRepo.
		EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(&domain.PullRequest{})).
		DoAndReturn(func(_ context.Context, updated *domain.PullRequest) error {
			if updated.Status != domain.PRStatusMerged {
				t.Errorf("expected status %q, got %q", domain.PRStatusMerged, updated.Status)
			}
			if updated.MergedAt == nil {
				t.Errorf("MergedAt must be set")
			}
			return nil
		})

	mergedPR := *pr
	mergedPR.Status = domain.PRStatusMerged
	mergedPR.MergedAt = func() *time.Time { x := time.Now().UTC(); return &x }()

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(&mergedPR, nil)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: userRepo,
		teams: teamRepo,
		prs:   prRepo,
	}

	ctx := context.Background()

	if _, err := svc.Merge(ctx, prID); err != nil {
		t.Fatalf("first Merge returned error: %v", err)
	}

	if _, err := svc.Merge(ctx, prID); err != nil {
		t.Fatalf("second Merge returned error: %v", err)
	}
}

func TestPullRequestService_Create_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   mocks.NewMockPullRequestRepository(ctrl),
	}

	ctx := context.Background()

	_, err := svc.Create(ctx, "", "name", "author")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty id, got %v", err)
	}

	_, err = svc.Create(ctx, "pr-1", "", "author")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty name, got %v", err)
	}

	_, err = svc.Create(ctx, "pr-1", "name", "")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty author, got %v", err)
	}
}

func TestPullRequestService_Create_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)

	authorID := domain.UserID("author")

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), authorID).
		Return(nil, domain.ErrNotFound)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: userRepo,
		teams: teamRepo,
		prs:   prRepo,
	}

	_, err := svc.Create(context.Background(), "pr-1", "Test PR", authorID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPullRequestService_Create_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)

	authorID := domain.UserID("author")
	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: authorID, Username: "Author", IsActive: true},
		},
	}

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), authorID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: userRepo,
		teams: teamRepo,
		prs:   prRepo,
	}

	_, err := svc.Create(context.Background(), "pr-1", "Test PR", authorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPullRequestService_Merge_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   mocks.NewMockPullRequestRepository(ctrl),
	}

	_, err := svc.Merge(context.Background(), "")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestPullRequestService_Merge_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	prID := domain.PullRequestID("pr-1")

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(nil, domain.ErrNotFound)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   prRepo,
	}

	_, err := svc.Merge(context.Background(), prID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPullRequestService_Merge_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{"u1"},
		CreatedAt:         time.Now().UTC(),
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	prRepo.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   prRepo,
	}

	_, err := svc.Merge(context.Background(), prID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPullRequestService_ReassignReviewer_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   mocks.NewMockPullRequestRepository(ctrl),
	}

	ctx := context.Background()

	_, _, err := svc.ReassignReviewer(ctx, "", "u1")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty prID, got %v", err)
	}

	_, _, err = svc.ReassignReviewer(ctx, "pr-1", "")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty oldReviewerID, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("u1")

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(nil, domain.ErrNotFound)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   prRepo,
	}

	_, _, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_AlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("u1")
	now := time.Now().UTC()

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []domain.UserID{"u1"},
		CreatedAt:         now,
		MergedAt:          &now,
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   prRepo,
	}

	_, _, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if !errors.Is(err, domain.ErrPullRequestAlreadyMerged) {
		t.Fatalf("expected ErrPullRequestAlreadyMerged, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_OldReviewerNotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("uX")

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{"u1", "u2"},
		CreatedAt:         time.Now().UTC(),
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: mocks.NewMockTeamRepository(ctrl),
		prs:   prRepo,
	}

	_, _, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if !errors.Is(err, domain.ErrReviewerNotAssigned) {
		t.Fatalf("expected ErrReviewerNotAssigned, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("u1")

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{"u1"},
		CreatedAt:         time.Now().UTC(),
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), oldID).
		Return(nil, domain.ErrNotFound)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: teamRepo,
		prs:   prRepo,
	}

	_, _, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_NoCandidates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("u1")

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{"u1", "u2"},
		CreatedAt:         time.Now().UTC(),
	}

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Username: "U1", IsActive: true},
			{ID: "u2", Username: "U2", IsActive: true},
			{ID: "u3", Username: "U3", IsActive: false},
		},
	}

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), oldID).
		Return(team, nil)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: teamRepo,
		prs:   prRepo,
	}

	_, _, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if !errors.Is(err, domain.ErrNoReviewerCandidates) {
		t.Fatalf("expected ErrNoReviewerCandidates, got %v", err)
	}
}

func TestPullRequestService_Create_PrAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)

	authorID := domain.UserID("author")
	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: authorID, Username: "Author", IsActive: true},
		},
	}

	teamRepo.
		EXPECT().
		GetByMemberID(gomock.Any(), authorID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(domain.ErrPullRequestAlreadyExists)

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: userRepo,
		teams: teamRepo,
		prs:   prRepo,
	}

	_, err := svc.Create(context.Background(), "pr-1", "Test PR", authorID)
	if !errors.Is(err, domain.ErrPullRequestAlreadyExists) {
		t.Fatalf("expected ErrPullRequestAlreadyExists, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)

	prID := domain.PullRequestID("pr-1")
	oldID := domain.UserID("u1")
	authorID := domain.UserID("author")
	newReviewerID := domain.UserID("u2")

	now := time.Now().UTC()

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              "Test PR",
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []domain.UserID{oldID},
		CreatedAt:         now,
	}

	team := &domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: oldID, Username: "Old", IsActive: true},
			{ID: newReviewerID, Username: "New", IsActive: true},
		},
	}
	prRepo.EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.EXPECT().
		GetByMemberID(gomock.Any(), oldID).
		Return(team, nil)

	prRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(&domain.PullRequest{})).
		DoAndReturn(func(_ context.Context, updated *domain.PullRequest) error {
			if len(updated.AssignedReviewers) != 1 {
				t.Fatalf("expected 1 reviewer, got %d", len(updated.AssignedReviewers))
			}
			if updated.AssignedReviewers[0] != newReviewerID {
				t.Fatalf("expected reviewer %q, got %q", newReviewerID, updated.AssignedReviewers[0])
			}
			if updated.AuthorID != authorID {
				t.Fatalf("expected author %q, got %q", authorID, updated.AuthorID)
			}
			return nil
		})

	svc := &PullRequestService{
		log:   newTestLogger(),
		users: mocks.NewMockUserRepository(ctrl),
		teams: teamRepo,
		prs:   prRepo,
	}

	gotPR, gotUser, err := svc.ReassignReviewer(context.Background(), prID, oldID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotUser == nil {
		t.Fatal("expected non-nil new reviewer, got nil")
	}
	if gotUser.ID != newReviewerID {
		t.Fatalf("expected new reviewer %q, got %q", newReviewerID, gotUser.ID)
	}

	if gotPR != pr {
		t.Fatalf("expected returned PR pointer %p, got %p", pr, gotPR)
	}
	if len(gotPR.AssignedReviewers) != 1 || gotPR.AssignedReviewers[0] != newReviewerID {
		t.Fatalf("expected reviewers [%q], got %v", newReviewerID, gotPR.AssignedReviewers)
	}
}
