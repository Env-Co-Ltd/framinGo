package main

import (
	"fmt"
	"time"
)

func doSessionTable() error {
	dbType := fra.DB.DataType
	if dbType == "mariadb" {
		dbType = "mysql"
	}
	if dbType == "postgres" {
		dbType = "postgres"
	}

	fileName := fmt.Sprintf("%d_create_session_table", time.Now().UnixMicro())
	upFile := fra.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
	downFile := fra.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/"+dbType+"_session.sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table session"), downFile)

	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}
	return nil
}
