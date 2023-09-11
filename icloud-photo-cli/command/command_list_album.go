package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/chyroc/icloudgo"
)

func NewListAlbumFlag() []cli.Flag {
	var res []cli.Flag
	res = append(res, commonFlag...)
	return res
}

func ListAlbum(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	cookieDir := c.String("cookie-dir")
	domain := c.String("domain")

	cli, err := icloudgo.New(&icloudgo.ClientOption{
		AppID:           username,
		CookieDir:       cookieDir,
		PasswordGetter:  getTextInput("apple id password", password, true),
		TwoFACodeGetter: getTextInput("2fa code", "", false),
		Domain:          domain,
	})
	if err != nil {
		return err
	}

	defer cli.Close()

	if err := cli.Authenticate(false, nil); err != nil {
		return err
	}

	photoCli, err := cli.PhotoCli()
	if err != nil {
		return err
	}

	if albums, err := photoCli.Albums(); err == nil {

		for key, _ := range albums {
			fmt.Println("Album:", key)
		}
	}

	return nil
}
