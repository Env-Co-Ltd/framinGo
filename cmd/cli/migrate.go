package main

func doMigrate(arg2, arg3 string) error {
	// dsn := getDSN()
	checkForDB()
	tx, err := fra.PopConnect()
	if err != nil {
		exitGracefully(err)
	}
	defer tx.Close()

	// runn the migration command
	switch arg2 {
	case "up":
		// err := fra.MigrateUp(dsn)
		err := fra.RunPopMigrations(tx)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			// err := fra.MigrateDownAll(dsn)
			err := fra.PopMigrationsDown(tx, -1)
			if err != nil {
				return err
			}
		} else {
			// err := fra.Steps(-1, dsn)
			err := fra.PopMigrationsDown(tx, 1)
			if err != nil {
				return err
			}
		}
	case "reset":
		// err := fra.MigrateDownAll(dsn)
		err := fra.PopMigrationsReset(tx)
		if err != nil {
			return err
		}
		// err = fra.MigrateUp(dsn)
		err = fra.RunPopMigrations(tx)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
