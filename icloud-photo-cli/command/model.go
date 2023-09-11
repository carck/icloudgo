package command

import (
	"fmt"
	"os"
	"strconv"

	"github.com/chyroc/icloudgo"
)

type PhotoAssetModel struct {
	ID     string `gorm:"column:id; index:uniq_id,unique"`
	Name   string `gorm:"column:name"`
	Data   string `gorm:"column:data"`
	Status int    `gorm:"column:status"`
}

func (r *downloadCommand) dalInit() error {
	_, err := r.db.Exec("CREATE TABLE if not exists assets(id text, data text, status integer, PRIMARY KEY (\"id\"))")
	return err
}

func (r *downloadCommand) dalAddAssets(assets []*icloudgo.PhotoAsset) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("replace into assets(id, data, status) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, v := range assets {
		_, err = stmt.Exec(v.ID(), string(v.Bytes()), 0)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *downloadCommand) dalDeleteAsset(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	_, err := r.db.Exec(fmt.Sprintf("delete from assets where id ='%s'", id))
	return err
}

func (r *downloadCommand) dalCountUnDownloadAssets() (cnt int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	row, err := r.db.Query("select count(*) as cnt from assets ")
	if err != nil {
		return 0
	}
	defer row.Close()
	err = row.Scan(&cnt)
	if err != nil {
		return 0
	}
	return cnt
}

func (r *downloadCommand) dalGetUnDownloadAssets(status *int) ([]*PhotoAssetModel, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pos := []*PhotoAssetModel{}

	rows, err := r.db.Query("select id, data, status from assets")
	if err != nil {
		return pos, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var data string
		var stat int
		err = rows.Scan(&id, &data, &stat)
		if err != nil {
			return pos, err
		}
		po := &PhotoAssetModel{
			ID:     id,
			Data:   data,
			Status: stat,
		}
		if status == nil {
			pos = append(pos, po)
		} else if po.Status == *status {
			pos = append(pos, po)
		}
	}
	err = rows.Err()

	return pos, err
}

func (r *downloadCommand) dalSetDownloaded(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.DelDownloaded {
		_, err := r.db.Exec(fmt.Sprintf("delete from assets where id ='%s'", id))
		return err
	}
	_, err := r.db.Exec(fmt.Sprintf("update assets set status=1 where id ='%s'", id))
	return err
}

func (r *downloadCommand) keyAssertPrefix() []byte {
	return []byte("assert_")
}

func (r *downloadCommand) keyAssert(id string) []byte {
	return []byte("assert_" + id)
}

func (r *downloadCommand) dalGetDownloadOffset(albumSize int) int {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.DelDownloaded {
		return 0
	}

	var result int
	offset, err := r.getDownloadOffset(false)
	if err != nil {
		fmt.Printf("[icloudgo] [offset] get db offset err: %s, reset to 0\n", err)
		return 0
	}
	fmt.Printf("[icloudgo] [offset] get db offset: %d\n", offset)
	if offset > albumSize {
		result = 0
		if err = r.saveDownloadOffset(0, false); err != nil {
			fmt.Printf("[icloudgo] [offset] db offset=%d, album_size=%d, reset to 0, and save_db failed: %s\n", offset, albumSize, err)
		} else {
			fmt.Printf("[icloudgo] [offset] db offset=%d, album_size=%d, reset to 0\n", offset, albumSize)
		}
	}
	result = offset
	return result
}

func (r *downloadCommand) getDownloadOffset(needLock bool) (int, error) {
	if needLock {
		r.lock.Lock()
		defer r.lock.Unlock()
	}
	data, err := os.ReadFile("offset")
	if err != nil {
		return 0, nil
	}
	offset, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (r *downloadCommand) saveDownloadOffset(offset int, needLock bool) error {
	if needLock {
		r.lock.Lock()
		defer r.lock.Unlock()
	}
	err := os.WriteFile("offset", []byte(fmt.Sprintf("%d", offset)), 0644)
	return err
}
