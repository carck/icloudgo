package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type PhotoVersion string

const (
	PhotoVersionOriginal      PhotoVersion = "original"
	PhotoVersionOriginalVideo              = "originalVideo"
	PhotoVersionMedium        PhotoVersion = "medium"
	PhotoVersionThumb         PhotoVersion = "thumb"
)

var extRegexp = regexp.MustCompile("\\.[^.]+$")

func (r *PhotoAsset) IsLivePhoto() bool {
	_, ok := r.getVersions()[PhotoVersionOriginalVideo]
	return ok
}

func (r *PhotoAsset) DownloadTo(version PhotoVersion, target string) error {
	body, err := r.Download(version)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0o644)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}

	_, err = io.Copy(f, body)
	if err != nil {
		return fmt.Errorf("copy file error: %v", err)
	}

	// 1676381385791 to time.time
	created := r.Created()
	if err := os.Chtimes(target, created, created); err != nil {
		return fmt.Errorf("change file time error: %v", err)
	}

	return nil
}

func (r *PhotoAsset) Download(version PhotoVersion) (io.ReadCloser, error) {
	versionDetail, ok := r.getVersions()[version]
	if !ok {
		var keys []string
		for k := range r.getVersions() {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("version %s not found, valid: %s", version, strings.Join(keys, ","))
	}

	timeout := time.Minute * 10 // 10分钟
	if versionDetail.Size > 0 {
		slowSecond := time.Duration(versionDetail.Size/1024/100) * time.Second // 100 KB/s 秒
		if slowSecond > timeout {
			timeout = slowSecond
		}
	}

	body, err := r.service.icloud.requestStream(&rawReq{
		Method:       http.MethodGet,
		URL:          versionDetail.URL,
		Headers:      r.service.icloud.getCommonHeaders(map[string]string{}),
		ExpectStatus: newSet[int](http.StatusOK),
		Timeout:      timeout,
	})
	if err != nil {
		return body, fmt.Errorf("download %s(timeout: %s) failed: %w", r.Filename(), timeout, err)
	}
	return body, nil
}

func (r *PhotoAsset) getVersions() map[PhotoVersion]*photoVersionDetail {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r._versions) == 0 {
		r._versions = r.packVersion()
	}

	return r._versions
}

func (r *PhotoAsset) packVersion() map[PhotoVersion]*photoVersionDetail {
	fields := r._masterRecord.Fields

	versions := map[PhotoVersion]*photoVersionDetail{
		PhotoVersionOriginal: {
			Filename: r.Filename(),
			Width:    fields.ResOriginalWidth.Value,
			Height:   fields.ResOriginalHeight.Value,
			Size:     fields.ResOriginalRes.Value.Size,
			URL:      fields.ResOriginalRes.Value.DownloadURL,
			Type:     fields.ResOriginalFileType.Value,
		},
	}
	if fields.ResOriginalVidComplRes.Value.Size != 0 {
		versions[PhotoVersionOriginalVideo] = &photoVersionDetail{
			Filename: extRegexp.ReplaceAllString(r.Filename(), ".MOV"),
			Width:    fields.ResOriginalVidComplWidth.Value,
			Height:   fields.ResOriginalVidComplHeight.Value,
			Size:     fields.ResOriginalVidComplRes.Value.Size,
			URL:      fields.ResOriginalVidComplRes.Value.DownloadURL,
			Type:     fields.ResOriginalVidComplFileType.Value,
		}
	}
	return versions
}

type photoVersionDetail struct {
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
	URL      string `json:"url"`
	Type     string `json:"type"`
}
