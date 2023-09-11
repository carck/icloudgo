package command

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

func NewListDuplicateFlag() []cli.Flag {
	var res []cli.Flag
	res = append(res,
		&cli.StringFlag{
			Name:     "output",
			Usage:    "output dir",
			Required: false,
			Value:    "./iCloudPhotos",
			Aliases:  []string{"o"},
			EnvVars:  []string{"ICLOUD_OUTPUT"},
		},
		&cli.StringFlag{
			Name:     "hash-db",
			Usage:    "hash db to check for duplicate",
			Required: false,
			EnvVars:  []string{"ICLOUD_HASH_DB"},
		},
		&cli.StringFlag{
			Name:     "duplicate-action",
			Usage:    "action for files detected as duplicates",
			Required: false,
			Value:    "log",
			EnvVars:  []string{"ICLOUD_DUPLICATE_ACTION"},
		},
	)
	return res
}

func ListDuplicate(c *cli.Context) error {

	dbPath := c.String("hash-db")
	action := c.String("duplicate-action")
	outputDir := c.String("output")

	fmt.Printf("check duplicates, dbpath=%s, action=%s, dir=%s\n", dbPath, action, outputDir)

	hashDb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("error open db: %s\n", dbPath)
		return err
	}

	return filepath.Walk(outputDir, func(path string, file fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			return nil
		}
		hash, hErr := Hash(path)
		if hErr != nil {
			return hErr
		}

		if stmt, sErr := hashDb.Prepare("select file_name from files where file_hash=? and file_size= ?"); sErr != nil {
			return sErr
		} else {
			defer stmt.Close()
			var cnt string
			if qErr := stmt.QueryRow(hash, file.Size()).Scan(&cnt); qErr != nil {
				return err
			}
			if cnt != "" {
				fmt.Printf("dupliate %s %s\n", path, cnt)
				if action == "delete" {
					if fErr := os.Remove(path); fErr != nil {
						return fErr
					}

				}
			}
		}

		return nil
	})
}
