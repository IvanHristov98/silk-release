package lib

import (
	"database/sql"
	"fmt"

	"github.com/cloudfoundry-incubator/silk/daemon/config"
	"github.com/rubenv/sql-migrate"
)

type DatabaseHandler struct {
	migrations *migrate.MemoryMigrationSource
	db         *sql.DB
	dbType     string
}

func NewDatabaseHandler(databaseConfig config.DatabaseConfig) (*DatabaseHandler, error) {
	db, err := sql.Open(databaseConfig.Type, databaseConfig.ConnectionString)
	if err != nil {
		return &DatabaseHandler{}, fmt.Errorf("connecting to database: %s", err)
	}

	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			&migrate.Migration{
				Id:   "1",
				Up:   []string{createSubnetTable(databaseConfig.Type)},
				Down: []string{"DROP TABLE subnets"},
			},
		},
	}

	return &DatabaseHandler{
		migrations: migrations,
		db:         db,
		dbType:     databaseConfig.Type,
	}, nil
}

func (d *DatabaseHandler) Migrate() (int, error) {
	return migrate.Exec(d.db, d.dbType, d.migrations, migrate.Up)
}

func (d *DatabaseHandler) AddEntry(underlayIP, subnet string) error {
	_, err := d.db.Exec(fmt.Sprintf("INSERT INTO subnets (underlay_ip, subnet) VALUES ('%s', '%s')", underlayIP, subnet))
	return err
}

func (d *DatabaseHandler) SubnetExists(subnet string) (bool, error) {
	var exists int
	err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM subnets WHERE subnet = '%s'", subnet)).Scan(&exists)
	if exists == 1 {
		return true, err
	} else {
		return false, err
	}
}

func (d *DatabaseHandler) SubnetForUnderlayIP(underlayIP string) (string, error) {
	var subnet string
	result := d.db.QueryRow(fmt.Sprintf("SELECT subnet FROM subnets WHERE underlay_ip = '%s'", underlayIP))
	err := result.Scan(&subnet)
	if err != nil {
		return "", err
	}
	return subnet, nil
}

func createSubnetTable(dbType string) string {
	baseCreateTable := "CREATE TABLE IF NOT EXISTS subnets ( " +
		" %s, " +
		" underlay_ip varchar(15), " +
		" subnet varchar(18), " +
		" UNIQUE (underlay_ip), " +
		" UNIQUE (subnet) " +
		");"
	mysqlId := "id int NOT NULL AUTO_INCREMENT, PRIMARY KEY (id)"
	psqlId := "id SERIAL PRIMARY KEY"

	switch dbType {
	case "postgres":
		return fmt.Sprintf(baseCreateTable, psqlId)
	case "mysql":
		return fmt.Sprintf(baseCreateTable, mysqlId)
	}

	return ""
}
