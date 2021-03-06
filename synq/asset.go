package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
)

type AssetResponse struct {
	Asset *Asset `json:"data"`
}

type AssetList struct {
	Assets []Asset `json:"data"`
}

type Asset struct {
	AccountId        string                  `json:"account_id"`
	VideoId          string                  `json:"video_id"`
	Id               string                  `json:"id"`
	Location         string                  `json:"location"`
	Url              string                  `json:"url"`
	State            string                  `json:"state"`
	Type             string                  `json:"type"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
	Metadata         json.RawMessage         `json:"metadata"`
	VmafScore        float64                 `json:"vmaf_score"`
	UploadInfo       AssetUpload             `json:"upload_info,omitempty"`
	Api              ApiV2                   `json:"-"`
	Video            VideoV2                 `json:"-"`
	UploadParameters upload.UploadParameters `json:"-"`
}

type AssetUpload struct {
	Checksum     string     `json:"checksum,omitempty"`
	ChecksumSize int64      `json:"checksum_size,omitempty"`
	Size         int64      `json:"size,omitempty"`
	Started      *time.Time `json:"started,omitempty"`
	Finished     *time.Time `json:"finished,omitempty"`
	Filename     string     `json:"filename,omitempty"`
}

func (a *Asset) getApi() *ApiV2 {
	if a.Api.BaseApi != nil {
		return &a.Api
	}
	if a.Video.Api != nil && a.Video.Api.BaseApi != nil {
		return a.Video.Api
	}
	log.Panicln("asset has no valid apis to use")
	return nil
}

func (a *Asset) Update() error {
	url := a.getApi().getBaseUrl() + "/assets/" + a.Id
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(data)
	return a.handleAssetReq("PUT", url, body)
}

func (a *Asset) Delete() error {
	url := a.getApi().getBaseUrl() + "/assets/" + a.Id
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(data)
	return a.handleAssetReq("DELETE", url, body)
}

func (a *Asset) handleAssetReq(method, url string, body io.Reader) error {
	resp := AssetResponse{Asset: a}
	req, err := a.getApi().makeRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")

	err = handleReq(a.Api, req, &resp)
	if err != nil {
		return err
	}

	return nil
}

func (a *Asset) GetUrl() string {
	if a.Url != "" {
		return a.Url
	}
	return a.Location
}

func (a *Asset) UploadFile(fileName string) error {
	upUrl := a.Api.UploadUrl
	if upUrl == "" {
		return errors.New("invalid upload url, can not upload file")
	}
	if a.UploadParameters.Key == "" {
		// if the location exists, get the upload parameters again
		if a.Location != "" && a.Type != "" {
			ext := filepath.Ext(fileName)
			req := upload.UploadRequest{
				Type:        a.Type,
				ContentType: common.ExtToCtype(ext),
				AssetId:     a.Id,
			}
			up, err := a.Video.GetUploadParams(req)
			if err != nil {
				return err
			}
			a.UploadParameters = up
		} else {
			return errors.New("upload parameters is invalid")
		}
	}
	f, err := os.Open(fileName)
	defer f.Close()
	if os.IsNotExist(err) {
		return errors.New("file '" + fileName + "' does not exist")
	}

	params := a.UploadParameters
	if !strings.Contains(params.SignatureUrl, "http") {
		sigUrl := upUrl + params.SignatureUrl
		log.Printf("Updating sig url to include host '%s'\n", upUrl)
		params.SignatureUrl = sigUrl
	}
	aws, err := upload.CreatorFn(params)
	if err != nil {
		return err
	}
	_, err = aws.Upload(f)
	return err
}
