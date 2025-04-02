package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

func doMake(arg2, arg3 string) error {
	switch arg2 {
		//暗号化キーを生成
	case "key":
		rnd := fra.RandomString(32)
		color.Yellow("320character encryption key: %s", rnd)
		color.Yellow("Please save this key in a secure location.")
	case "migration":
		dbType := fra.DB.DataType
		if arg3 == "" {
			exitGracefully(errors.New("migration name required"))
		}

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg3)

		upFile := fra.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
		downFile := fra.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

		err := copyFileFromTemplate("templates/migrations/migration."+dbType+".up.sql", upFile)
		if err != nil {
			exitGracefully(err)
		}
		err = copyFileFromTemplate("templates/migrations/migration."+dbType+".down.sql", downFile)
		if err != nil {
			exitGracefully(err)
		}
	case "auth":
		err := doAuth()
		if err != nil {
			exitGracefully(err)
		}
	case "handler":
		if arg3 == "" {
			exitGracefully(errors.New("handler name required"))
		}

		fileName := fra.RootPath + "/handlers/" + strings.ToLower(arg3) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists"))
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))
		err = os.WriteFile(fileName, []byte(handler), 0644)
		if err != nil {
			exitGracefully(err)
		}
	case "model":
		if arg3 == "" {
			exitGracefully(errors.New("model name required"))
		}

		data, err := templateFS.ReadFile("templates/data/model.go.txt")
		if err != nil {
			exitGracefully(err)
		}
		model := string(data)
		plur := pluralize.NewClient()
		var modelName = arg3
		var tableName = arg3
		if plur.IsPlural(arg3) {
			modelName = plur.Singular(arg3)
			tableName = strings.ToLower(tableName)
		} else {
			tableName = strings.ToLower(plur.Plural(arg3))
		}

		fileName := fra.RootPath + "/data/" + strings.ToLower(modelName) + ".go"

		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists"))
		}

		model = strings.ReplaceAll(model, "$MODELNAME$", strcase.ToCamel(modelName))
		model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

		err = copyDataToFile([]byte(model), fileName)
		if err != nil {
			exitGracefully(err)
		}
	case "mail":
		if arg3 == "" {
			exitGracefully(errors.New("mail templatename required"))
		}

		htmlMail := fra.RootPath + "/mail/" + strings.ToLower(arg3) + ".html.tmpl"
		plainMail := fra.RootPath + "/mail/" + strings.ToLower(arg3) + ".plain.tmpl"

		if fileExists(htmlMail) {
			exitGracefully(errors.New(htmlMail + " already exists"))
		}
		if fileExists(plainMail) {
			exitGracefully(errors.New(plainMail + " already exists"))
		}

		err := copyFileFromTemplate("templates/mailer/mail.html.tmpl", htmlMail)
		if err != nil {
			exitGracefully(err)
		}
		err = copyFileFromTemplate("templates/mailer/mail.plain.tmpl", plainMail)
		if err != nil {
			exitGracefully(err)
		}
		
	case "session":
		err := doSessionTable()
		if err != nil {
			exitGracefully(err)
		}
	}

	return nil
}
