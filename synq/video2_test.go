package synq

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/stretchr/testify/require"
)

func TestMarshalVideo(t *testing.T) {
	assert := require.New(t)
	video := setupTestVideoV2()
	video.Metadata = []byte(`{"test":true}`)
	bytes, err := json.Marshal(video)
	assert.Nil(err)
	v := VideoV2{}
	json.Unmarshal(bytes, &v)
	assert.Equal(video.Metadata, v.Metadata)
}

func TestVideoUpdate(t *testing.T) {
	assert := require.New(t)
	video := setupTestVideoV2()
	video.Metadata = json.RawMessage(`{"meta":"new"}`)
	video.Userdata = json.RawMessage(`{"user":"new"}`)
	video.CompletenessScore = 95.4
	// this is just fake, the updated value will come from a hard coded json
	err := video.Update()
	assert.Nil(err)
	assert.Equal(`{"type":"show"}`, string(video.Metadata))
	assert.Contains(string(video.Userdata), "test2")
	reqs, vals := test_server.GetReqs()
	assert.Len(reqs, 1)
	req := reqs[0]
	assert.Equal("application/json", req.Header.Get("Content-Type"))
	assert.Len(vals, 1)
	assert.Equal(`{"metadata":{"meta":"new"},"user_data":{"user":"new"},"completeness_score":95.4}`, vals[0].Get("body"))
}

func TestCreateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset, err := video.CreateAsset(ASSET_CREATED, ASSET_TYPE, ASSET_LOCATION)
	assert.Nil(err)
	assert.NotNil(asset.Id)
	assert.Equal(testAssetId, asset.Id)
}

func TestCreateOrUpdateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset := Asset{State: ASSET_CREATED, Type: ASSET_TYPE, Location: ASSET_LOCATION}
	err := video.CreateOrUpdateAsset(&asset)
	assert.Nil(err)
	assert.Equal(testAssetId, asset.Id)
	asset.State = ASSET_UPLOADED
	err = video.CreateOrUpdateAsset(&asset)
	reqs, vals := test_server.GetReqs()
	assert.Nil(err)
	assert.Len(reqs, 2)
	assert.Len(vals, 2)
	req := reqs[1]
	val := vals[1]
	assert.Equal("PUT", req.Method)
	a := Asset{}
	body := val.Get("body")
	json.Unmarshal([]byte(body), &a)
	assert.Equal(asset.State, a.State)
	assert.Equal(asset.Location, a.Location)
}

func TestCreateAssetForUpload(t *testing.T) {
	log.Println("Testing GettingAssetForUpload")
	assert := require.New(t)
	video := setupTestVideoV2()
	req := upload.UploadRequest{
		ContentType: "video/mp4",
	}
	asset, err := video.CreateAssetForUpload(req)
	assert.Nil(err)
	assert.Len(video.Assets, 1)
	assert.Equal(asset.Id, video.Assets[0].Id)
	assert.Equal("uploads/9e/9d/9e9dc8c8f70541db88dab3034894deb9/01823629bcf24c34b714ae21e1a4647f.mp4", asset.UploadParameters.Key)
	assert.Equal("https://synq-bruce.s3.amazonaws.com", asset.UploadParameters.Action)
}

func TestAddAccount(t *testing.T) {
	assert := require.New(t)
	video := setupTestVideoV2()
	err := video.AddAccount(test_server.ACCOUNT_ID)
	assert.Nil(err)
	reqs, vals := test_server.GetReqs()
	assert.Len(reqs, 1)
	val := vals[0]
	body := val.Get("body")
	obj := struct {
		Accounts []VideoAccount `json:"video_accounts"`
	}{}
	json.Unmarshal([]byte(body), &obj)
	assert.Len(obj.Accounts, 1)
	assert.Equal(test_server.ACCOUNT_ID, obj.Accounts[0].Id)
}

func TestFindAsset(t *testing.T) {
	assert := require.New(t)
	loc := "myloc"
	id := "myid"
	orgUrl := "http://orgurl"
	video := VideoV2{}
	asset := Asset{Location: loc}
	asset2 := Asset{Location: "diffloc", Id: id}
	bytes := []byte(`{"org_url":"` + orgUrl + `"}`)
	asset3 := Asset{Location: "http://cdn", Id: "123", Metadata: bytes, Type: "thumbnail"}
	_, found := video.FindAsset(loc)
	assert.False(found)
	video.Assets = append(video.Assets, asset)
	// can't find location if Id is blank
	_, found = video.FindAsset(loc)
	assert.False(found)
	asset.Id = "blah"
	video.Assets = []Asset{asset}
	_, found = video.FindAsset(loc)
	assert.True(found)
	video.Assets = append(video.Assets, asset2)
	_, found = video.FindAsset(id)
	assert.True(found)
	video.Assets = append(video.Assets, asset3)
	f, found := video.FindAsset(orgUrl)
	assert.True(found)
	assert.Equal("123", f.Id)
}

func TestFindAssetByType(t *testing.T) {
	assert := require.New(t)
	video := VideoV2{}
	types := [3]string{"test_type", "test_type_1", "test_type_2"}
	asset := Asset{Type: types[0]}
	asset1 := Asset{Type: types[1], Id: "test_id_1"}
	asset2 := Asset{Type: types[2], Id: "test_id_2"}
	asset3 := Asset{Type: types[2], Id: "test_id_3"}
	video.Assets = append(video.Assets, asset)
	_, found := video.FindAssetByType(types[0])
	assert.False(found)
	video.Assets = append(video.Assets, asset1)
	type1Assets, found := video.FindAssetByType(types[1])
	assert.True(found)
	assert.Equal(type1Assets[0].Id, asset1.Id)
	video.Assets = append(video.Assets, asset2)
	video.Assets = append(video.Assets, asset3)
	type2Assets, found := video.FindAssetByType(types[2])
	assert.True(found)
	assert.Equal(len(type2Assets), 2)
}

// Test misc video functions
func TestMisc(t *testing.T) {
	assert := require.New(t)
	video := VideoV2{}
	api := setupTestApiV2("fake")
	assert.Equal("Empty Video\n", video.Display())
	video.Id = "abc123"
	assert.Equal("Video abc123\n\tAssets : 0\n", video.Display())
	req := upload.UploadRequest{}
	_, err := video.GetUploadParams(req)
	assert.NotNil(err)
	assert.Equal("api is blank", err.Error())
	_, err = video.CreateAssetForUpload(req)
	assert.NotNil(err)
	assert.Equal("api is blank", err.Error())
	video.Api = &api
	err = video.AddAccount(test_server.ACCOUNT_ID)
	assert.NotNil(err)
	assert.Equal("invalid auth", err.Error())
	r, e := video.Value()
	assert.Nil(e)
	val := r.([]byte)
	bytes, _ := json.Marshal(video)
	assert.Equal(bytes, val)
	video2 := VideoV2{}
	err = video2.Scan("a")
	assert.NotNil(err)
	assert.Equal("Type assertion .([]byte) failed.", err.Error())
	err = video2.Scan(val)
	assert.Nil(err)
	assert.Equal(video.Id, video2.Id)
}
