package helper

import (
	"encoding/json"
	"log"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
)

// fow now, the query will be the account id
func LoadVideosByAccount(accountId, name string, c common.Cacheable, api synq.ApiV2) (videos []synq.VideoV2, err error) {
	ok := common.LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("get all videos in account %s\n", accountId)
	videos, err = api.GetVideos(accountId)
	if err != nil {
		return videos, err
	}
	common.SaveToCache(name, c, videos)
	return videos, err
}

// fow now, the query will be the account id
func LoadRawVideosByAccount(accountId, name string, c common.Cacheable, api synq.ApiV2) (videos []json.RawMessage, err error) {
	ok := common.LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("get all raw videos (filter by account '%s')\n", accountId)
	videos, err = api.GetRawVideos(accountId)
	if err != nil {
		return videos, err
	}
	common.SaveToCache(name, c, videos)
	return videos, err
}

func LoadVideoV2(id string, c common.Cacheable, api synq.ApiV2) (video synq.VideoV2, err error) {
	id = common.ConvertToUUIDFormat(id)
	ok := common.LoadFromCache(id, c, &video)
	if ok {
		api.SetApi(&video)
		return video, nil
	}
	log.Printf("Getting video v2 %s\n", id)
	video, err = api.GetVideo(id)
	if err != nil {
		return video, err
	}
	common.SaveToCache(id, c, &video)
	return video, nil
}

func LoadUploadParameters(id string, req upload.UploadRequest, c common.Cacheable, api synq.ApiV2) (up upload.UploadParameters, err error) {
	lookId := id
	if req.AssetId != "" {
		lookId = req.AssetId
	}
	ok := common.LoadFromCache(lookId+"_up", c, &up)
	if ok {
		return up, nil
	}
	log.Printf("Getting upload parameters for %s\n", id)
	up, err = api.GetUploadParams(id, req)
	if err != nil {
		return up, err
	}
	common.SaveToCache(lookId+"_up", c, &up)
	return up, nil
}

func LoadAsset(id string, c common.Cacheable, api synq.ApiV2) (asset synq.Asset, err error) {
	ok := common.LoadFromCache(id, c, &asset)
	if !ok {
		log.Printf("Getting asset %s\n", id)
		asset, err = api.GetAsset(id)
		if err != nil {
			return asset, err
		}
	} else {
		asset.Api = api
	}

	if asset.Video.Id == "" {
		video, err := LoadVideoV2(asset.VideoId, c, api)
		if err != nil {
			return asset, err
		}
		asset.Video = video
	} else {
		// cache the video for re-use later
		common.SaveToCache(asset.Video.Id, c, &asset.Video)
	}
	common.SaveToCache(id, c, &asset)
	return asset, nil
}
