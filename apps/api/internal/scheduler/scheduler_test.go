package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

type recorder struct {
	calls map[string]bool
}

func (self *recorder) mark(name string) error {
	self.calls[name] = true
	return nil
}

func (self *recorder) ProcessCycles(context.Context) error { return self.mark("cycles") }
func (self *recorder) Cleanup(context.Context) error       { return self.mark("oauth") }
func (self *recorder) PurgeTrashed(context.Context) error  { return self.mark("trash") }
func (self *recorder) AutoCloseStaleIssues(context.Context) error {
	return self.mark("auto close")
}
func (self *recorder) AutoArchiveDoneIssues(context.Context) error {
	return self.mark("auto archive")
}
func (self *recorder) SendEmailBatch(context.Context) error { return self.mark("notifications") }
func (self *recorder) NotifyDueSoon(context.Context) error  { return self.mark("due soon") }
func (self *recorder) CleanupExpiredInvitations(context.Context) error {
	return self.mark("invitations")
}
func (self *recorder) CleanupPasswordResetTokens(context.Context) error {
	return self.mark("resets")
}

func TestJobsCoverTheDocumentedSchedule(t *testing.T) {
	r := &recorder{calls: map[string]bool{}}
	jobs := Jobs(Dependencies{Cycles: r, OAuth: r, Invitations: r, PasswordResets: r, Trash: r, Automation: r, Notifications: r})

	every := map[string]time.Duration{}
	for _, job := range jobs {
		every[job.Name] = job.Every
		if err := job.Run(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if every["cycle rollover"] != 10*time.Minute || every["notification email"] != 5*time.Minute || every["oauth cleanup"] != time.Hour || every["issue due soon"] != 24*time.Hour {
		t.Fatalf("intervals = %v", every)
	}
	for _, name := range []string{"cycles", "oauth", "invitations", "resets", "trash", "auto close", "auto archive", "notifications", "due soon"} {
		if !r.calls[name] {
			t.Errorf("%s is never scheduled", name)
		}
	}
}

func TestRunUsesTheSystemCallerAndSurvivesFailures(t *testing.T) {
	var sawSystem bool
	job := Job{Name: "x", Run: func(ctx context.Context) error {
		sawSystem = permissions_repo.IsSystemCaller(ctx)
		return errors.New("boom")
	}}
	run(permissions_repo.WithSystemCaller(context.Background()), job)
	if !sawSystem {
		t.Fatal("jobs must run as the system caller")
	}
}
