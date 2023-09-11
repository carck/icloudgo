package command

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/chyroc/icloudgo"
	"github.com/urfave/cli/v2"
)

func NewListDBFlag() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "cookie-dir",
			Usage:    "cookie dir",
			Required: false,
			Aliases:  []string{"c"},
			EnvVars:  []string{"ICLOUD_COOKIE_DIR"},
		},
	}
}

func ListDB(c *cli.Context) error {
	cli, err := icloudgo.New(&icloudgo.ClientOption{
		Domain:    "cn",
		CookieDir: c.String("cookie-dir"),
	})
	if err != nil {
		return err
	}

	dbPath := cli.ConfigPath("badger.db")
	fmt.Println("db.path", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	r := &downloadCommand{
		db:   db,
		lock: new(sync.Mutex),
	}
	defer r.Close()

	if err := r.dalInit(); err != nil {
		return err
	}
	pos, err := r.dalGetUnDownloadAssets(nil)
	if err != nil {
		return err
	}

	for _, v := range pos {
		fmt.Printf("id: %s, name: %s, status: %d\n", v.ID, v.Name, v.Status)
	}

	return nil
}
