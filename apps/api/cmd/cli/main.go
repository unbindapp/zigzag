package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	_ "go.uber.org/automaxprocs"
)

const usageText = `
Zigzag CLI - Management Tool

Usage:
  zigzag-cli [command] [options]

Commands:
  user:list                      List all users
  user:create --email=EMAIL --password=PASSWORD [--name=NAME] [--display-name=HANDLE] [--system-admin]
                                 Create an active user
  user:reset-password --email=EMAIL --password=PASSWORD
                                 Set a user's password and sign out their sessions
`

func main() {
	if err := godotenv.Overload(); err != nil && !os.IsNotExist(err) {
		log.Warn("Error loading .env file", "err", err)
	}
	if len(os.Args) < 2 {
		fmt.Print(usageText)
		os.Exit(1)
	}

	createFlags := flag.NewFlagSet("user:create", flag.ExitOnError)
	createEmail := createFlags.String("email", "", "Email of the new user")
	createPassword := createFlags.String("password", "", "Password of the new user, at least 8 characters")
	createName := createFlags.String("name", "", "Full name")
	createHandle := createFlags.String("display-name", "", "Mention handle, derived from the email when omitted")
	createAdmin := createFlags.Bool("system-admin", false, "Make the user a system admin")

	resetFlags := flag.NewFlagSet("user:reset-password", flag.ExitOnError)
	resetEmail := resetFlags.String("email", "", "Email of the user")
	resetPassword := resetFlags.String("password", "", "New password, at least 8 characters")

	switch os.Args[1] {
	case "user:list":
		newCLI(config.NewConfig()).listUsers()
	case "user:create":
		_ = createFlags.Parse(os.Args[2:])
		if *createEmail == "" || *createPassword == "" {
			fmt.Println("Error: email and password are required")
			createFlags.Usage()
			os.Exit(1)
		}
		newCLI(config.NewConfig()).createUser(*createEmail, *createPassword, *createName, *createHandle, *createAdmin)
	case "user:reset-password":
		_ = resetFlags.Parse(os.Args[2:])
		if *resetEmail == "" || *resetPassword == "" {
			fmt.Println("Error: email and password are required")
			resetFlags.Usage()
			os.Exit(1)
		}
		newCLI(config.NewConfig()).resetPassword(*resetEmail, *resetPassword)
	case "help", "-h", "--help":
		fmt.Print(usageText)
	default:
		fmt.Printf("Unknown command %q\n", os.Args[1])
		fmt.Print(usageText)
		os.Exit(1)
	}
}
