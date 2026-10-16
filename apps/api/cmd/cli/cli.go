package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/database"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	upload_service "github.com/zigzagapp-org/zigzag-api/internal/services/upload"
)

const minPasswordLength = 8

type cli struct {
	cfg  *config.Config
	repo *repositories.Repositories
	ctx  context.Context
}

func newCLI(cfg *config.Config) *cli {
	conn, err := database.GetSqlDbConn(cfg)
	if err != nil {
		log.Fatalf("Failed to get database connection info: %v", err)
	}
	db, _, err := database.NewEntClient(conn)
	if err != nil {
		log.Fatalf("Failed to create ent client: %v", err)
	}
	return &cli{cfg: cfg, repo: repositories.NewRepositories(db), ctx: permissions_repo.WithSystemCaller(context.Background())}
}

func (self *cli) listUsers() {
	users, err := self.repo.User().List(self.ctx, &user_repo.ListUsersInput{})
	if err != nil {
		fail("listing users: %v", err)
	}
	for _, user := range users {
		flags := []string{}
		if user.SystemAdmin {
			flags = append(flags, "system admin")
		}
		if !user.Active {
			flags = append(flags, "inactive")
		}
		fmt.Printf("%s\t%s\t@%s\t%s\n", user.ID, user.Email, user.DisplayName, strings.Join(flags, ", "))
	}
}

func (self *cli) createUser(rawEmail, password, name, handle string, systemAdmin bool) {
	email, err := names.Email(rawEmail)
	if err != nil {
		fail("%v", err)
	}
	if len(password) < minPasswordLength {
		fail("password must be at least %d characters", minPasswordLength)
	}
	if handle == "" {
		handle = names.SuggestDisplayName(email)
	}
	handle, err = names.DisplayName(handle)
	if err != nil {
		fail("%v", err)
	}
	if name == "" {
		name = handle
	}
	user, err := self.repo.User().Create(self.ctx, nil, &user_repo.CreateUserInput{
		Email:       email,
		Password:    password,
		Name:        name,
		DisplayName: handle,
		SystemAdmin: systemAdmin,
		Active:      true,
	})
	if ent.IsConstraintError(err) {
		fail("a user with this email or display name already exists")
	}
	if err != nil {
		fail("creating user: %v", err)
	}
	fmt.Printf("Created user %s (%s, @%s)\n", user.ID, user.Email, user.DisplayName)
}

func (self *cli) resetPassword(email, password string) {
	if len(password) < minPasswordLength {
		fail("password must be at least %d characters", minPasswordLength)
	}
	user, err := self.repo.User().GetByEmail(self.ctx, email)
	if ent.IsNotFound(err) {
		fail("no user with email %s", email)
	}
	if err != nil {
		fail("loading user: %v", err)
	}
	if err := self.repo.User().UpdatePassword(self.ctx, nil, user.ID, password); err != nil {
		fail("updating password: %v", err)
	}
	if err := self.repo.Oauth().RevokeAllForUser(self.ctx, user.ID); err != nil {
		fail("revoking sessions: %v", err)
	}
	fmt.Printf("Password updated for %s; existing sessions were signed out\n", user.Email)
}

func (self *cli) importUploads(dir string) {
	result, err := upload_service.NewUploadService(self.cfg, self.repo).ImportDir(self.ctx, dir)
	if err != nil {
		fail("importing uploads: %v", err)
	}
	fmt.Printf("Imported %d uploads, %d files missing, %d failed\n", result.Imported, result.Missing, result.Failed)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
