package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

type PhotoAsset struct {
	service       *PhotoService
	_versions     map[PhotoVersion]*photoVersionDetail
	_masterRecord *photoRecord
	_assetRecord  *photoRecord
	lock          *sync.Mutex
}

var extMap = map[string]string{
	"public.heic":               ".HEIC",
	"public.jpeg":               ".jpg",
	"public.png":                ".png",
	"com.apple.quicktime-movie": ".MOV",
}

func (r *PhotoAsset) Bytes() []byte {
	bs, _ := json.Marshal(photoAssetData{
		MasterRecord: r._masterRecord,
		AssetRecord:  r._assetRecord,
	})
	return bs
}

func (r *PhotoService) NewPhotoAssetFromBytes(bs []byte) *PhotoAsset {
	data := &photoAssetData{}
	_ = json.Unmarshal(bs, data)

	return r.newPhotoAsset(data.MasterRecord, data.AssetRecord)
}

type photoAssetData struct {
	MasterRecord *photoRecord `json:"master_record"`
	AssetRecord  *photoRecord `json:"asset_record"`
}

func (r *PhotoService) newPhotoAsset(masterRecord, assetRecords *photoRecord) *PhotoAsset {
	return &PhotoAsset{
		service:       r,
		_masterRecord: masterRecord,
		_assetRecord:  assetRecords,
		_versions:     nil,
		lock:          new(sync.Mutex),
	}
}

func (r *PhotoAsset) Filename() string {
	if v := r._masterRecord.Fields.FilenameEnc.Value; v != "" {
		bs, _ := base64.StdEncoding.DecodeString(v)
		if len(bs) > 0 {
			return cleanFilename(string(bs))
		}
	}

	return cleanFilename(r.ID() + r.FileExt())
}

func (r *PhotoAsset) FileExt() string {
	if ext, ok := extMap[r._masterRecord.Fields.ItemType.Value]; ok {
		return ext
	} else {
		return ""
	}
}

func (r *PhotoAsset) LocalPath(outputDir string, size PhotoVersion, fileStructure string) string {
	version := r.getVersions()[size]
	ext := filepath.Ext(version.Filename)
	filename := ""
	switch fileStructure {
	case "name":
		filename = cleanFilename(version.Filename)
	default:
		filename = cleanFilename(r.ID()) + ext
	}

	if size == PhotoVersionOriginal || size == PhotoVersionOriginalVideo {
		return filepath.Join(outputDir, filename)
	}

	return filepath.Join(outputDir, filename+"_"+string(size)+ext)
}

func (r *PhotoAsset) ID() string {
	return r._masterRecord.RecordName
}

func (r *PhotoAsset) Size(version PhotoVersion) int {
	return r.getVersions()[version].Size
}

func (r *PhotoAsset) FormatSize(version PhotoVersion) string {
	return formatSize(r.Size(version))
}

func (r *PhotoAsset) Created() time.Time {
	return time.UnixMilli(r._masterRecord.Created.Timestamp)
}

func (r *PhotoAsset) OutputDir(output, folderStructure string) string {
	if folderStructure == "" || folderStructure == "/" {
		return output
	}

	createdFolderName := r.Created().Format(folderStructure)
	return filepath.Join(output, createdFolderName)
}

func formatSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.2fGB", float64(size)/1024/1024/1024)
	}
}
