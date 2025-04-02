package main

func doMigrate(arg2, arg3 string) error {
	dsn := getDSN()

	// runn the migration command
	switch arg2 {
	case "up":
		err := fra.MigrateUp(dsn)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			err := fra.MigrateDownAll(dsn)
			if err != nil {
				return err
			}
		} else {
			err := fra.Steps(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := fra.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		err = fra.MigrateUp(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
