package scheduler

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

type Job struct {
	Name  string
	Every time.Duration
	Run   func(ctx context.Context) error
}

type CycleProcessor interface {
	ProcessCycles(ctx context.Context) error
}

type OAuthCleaner interface {
	Cleanup(ctx context.Context) error
}

type InvitationCleaner interface {
	CleanupExpiredInvitations(ctx context.Context) error
}

type PasswordResetCleaner interface {
	CleanupPasswordResetTokens(ctx context.Context) error
}

type TrashPurger interface {
	PurgeTrashed(ctx context.Context) error
}

type IssueAutomation interface {
	AutoCloseStaleIssues(ctx context.Context) error
	AutoArchiveDoneIssues(ctx context.Context) error
}

type NotificationMailer interface {
	SendEmailBatch(ctx context.Context) error
	NotifyDueSoon(ctx context.Context) error
}

type Dependencies struct {
	Cycles         CycleProcessor
	OAuth          OAuthCleaner
	Invitations    InvitationCleaner
	PasswordResets PasswordResetCleaner
	Trash          TrashPurger
	Automation     IssueAutomation
	Notifications  NotificationMailer
}

func Jobs(deps Dependencies) []Job {
	return []Job{
		{Name: "cycle rollover", Every: 10 * time.Minute, Run: deps.Cycles.ProcessCycles},
		{Name: "oauth cleanup", Every: time.Hour, Run: deps.OAuth.Cleanup},
		{Name: "invitation expiry", Every: time.Hour, Run: deps.Invitations.CleanupExpiredInvitations},
		{Name: "password reset expiry", Every: time.Hour, Run: deps.PasswordResets.CleanupPasswordResetTokens},
		{Name: "trash purge", Every: time.Hour, Run: deps.Trash.PurgeTrashed},
		{Name: "issue auto close", Every: time.Hour, Run: deps.Automation.AutoCloseStaleIssues},
		{Name: "issue auto archive", Every: time.Hour, Run: deps.Automation.AutoArchiveDoneIssues},
		{Name: "notification email", Every: 5 * time.Minute, Run: deps.Notifications.SendEmailBatch},
		{Name: "issue due soon", Every: 24 * time.Hour, Run: deps.Notifications.NotifyDueSoon},
	}
}

type Scheduler struct {
	cron gocron.Scheduler
}

func Start(ctx context.Context, jobs []Job) (*Scheduler, error) {
	cron, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, err
	}
	systemCtx := permissions_repo.WithSystemCaller(ctx)
	for _, job := range jobs {
		_, err := cron.NewJob(
			gocron.DurationJob(job.Every),
			gocron.NewTask(run, systemCtx, job),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
		if err != nil {
			return nil, err
		}
	}
	cron.Start()
	return &Scheduler{cron: cron}, nil
}

func (self *Scheduler) Shutdown() error {
	return self.cron.Shutdown()
}

func run(ctx context.Context, job Job) {
	if err := job.Run(ctx); err != nil {
		log.Error("Scheduled job failed", "job", job.Name, "err", err)
	}
}
