package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

func doAuth() error {
	//migration
	dbType := fra.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := fra.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := fra.RootPath + "/migrations/" + fileName + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".sql", upFile)
	if err != nil {
		exitGracefully(err)
	}
	err = copyDataToFile([]byte("drop table if exists users cascade;drop table if exists tokens cascade;drop table if exists remember_tokens;"), downFile)
	if err != nil {
		exitGracefully(err)
	}
	//run migration
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	//copy files over
	err = copyFileFromTemplate("templates/data/user.go.txt", fra.RootPath + "/data/user.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/data/token.go.txt", fra.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/data/remember_token.go.txt", fra.RootPath+"/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	//copy ocer middleware
		err = copyFileFromTemplate("templates/middleware/auth.go.txt", fra.RootPath+"/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", fra.RootPath+"/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/middleware/remember.go.txt", fra.RootPath+"/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/handlers/auth-handlers.go.txt", fra.RootPath+"/handlers/auth-handlers.go")
	if err != nil {
		exitGracefully(err)
	}
	//mail
	err = copyFileFromTemplate("templates/mail/password-reset.html.tmpl", fra.RootPath+"/mail/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/mail/password-reset.plain.tmpl", fra.RootPath+"/mail/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}
	//templates
	err = copyFileFromTemplate("templates/views/login.jet", fra.RootPath+"/templates/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/views/forgot.jet", fra.RootPath+"/templates/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/pages/reset-password.jet", fra.RootPath+"/templates/pages/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}
	//routes
	err = copyFileFromTemplate("templates/routes/auth-routes.go.txt", fra.RootPath+"/routes/auth-routes.go")
	if err != nil {
		exitGracefully(err)
	}
	
	color.Yellow("  - users, tokens, remember_tokens created and excuted")
	color.Yellow("  - users and token models created")
	color.Yellow("  - auth middleware created")
	color.Yellow("")
	color.Yellow("Dont forget to add user and token models to data/models.go, and to add appropriate middleware to your routes")
	return nil
}
