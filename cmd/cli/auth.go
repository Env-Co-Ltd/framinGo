package main

import (
	"github.com/fatih/color"
)

func doAuth() error {
	checkForDB()

	// migrations
	dbType := fra.DB.DataType

	tx, err := fra.PopConnect()
	if err != nil {
		exitGracefully(err)
	}
	defer tx.Close()

	upBytes, err := templateFS.ReadFile("templates/migrations/auth_tables." + dbType + ".sql")
	if err != nil {
		exitGracefully(err)
	}

	downBytes := []byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;")
	if err != nil {
		exitGracefully(err)
	}

	err = fra.CreatePopMigration(upBytes, downBytes, "auth", "sql")
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = fra.RunPopMigrations(tx)
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/user.go.txt", fra.RootPath + "/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/token.go.txt", fra.RootPath + "/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/remember_token.go.txt", fra.RootPath + "/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy over middleware
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", fra.RootPath + "/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", fra.RootPath + "/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/remember.go.txt", fra.RootPath + "/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/handlers/auth-handlers.go.txt", fra.RootPath + "/handlers/auth-handlers.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.html.tmpl", fra.RootPath + "/mail/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.plain.tmpl", fra.RootPath + "/mail/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/login.jet", fra.RootPath + "/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/forgot.jet", fra.RootPath + "/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/views/reset-password.jet", fra.RootPath + "/views/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}


	color.Yellow("  - users, tokens, and remember_tokens migrations created and executed")
	color.Yellow("  - user and token models created")
	color.Yellow("  - auth middleware created")
	color.Yellow("")
	color.Yellow("Don't forget to add user and token models in data/models.go, and to add appropriate middleware to your routes!")

	return nil
}