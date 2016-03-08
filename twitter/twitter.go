package twitter

import (
    "io/ioutil"
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "mime/multipart"
    "net/url"
    "strings"
)

const StatusUpdate string = "https://api.twitter.com/1.1/statuses/update.json"
const MediaUpload string = "https://upload.twitter.com/1.1/media/upload.json"

type Twitter struct {
    client *http.Client
}

type MediaInitResponse struct {
    MediaId uint64 `json:"media_id"`
    MediaIdString string `json:"media_id_string"`
    ExpiresAfterSecs uint64 `json:"expires_after_secs"`
}

func NewTwitter(client *http.Client) *Twitter {
    self := &Twitter{}
    self.client = client
    return self
}

func (self *Twitter) MakeTwitWithMedia(text string, media []byte){
    fmt.Println("bytes", len(media))

    mediaInitResponse, err := self.MediaInit(media)
    if err != nil {
        fmt.Println("Can't init media", err)
    }

    fmt.Println(mediaInitResponse)

    mediaId := mediaInitResponse.MediaId

    if self.MediaAppend(mediaId, media) != nil {
        fmt.Println("Cant't append media")
    }

    if self.MediaFinilize(mediaId) != nil {
        fmt.Println("Cant't fin media")
    }

    if self.UpdateStatusWithMedia(text, mediaId) != nil {
        fmt.Println("Can't update status")
    }
}

func (self *Twitter) MediaInit(media []byte) (*MediaInitResponse, error) {
    form := url.Values{}
    form.Add("command", "INIT")
    form.Add("media_type", "video/mp4")
    form.Add("total_bytes", fmt.Sprint(len(media)))

    fmt.Println(form.Encode())

    req, err := http.NewRequest("POST", MediaUpload, strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    res, err := self.client.Do(req)

    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    fmt.Println("response", string(body))

    var mediaInitResponse MediaInitResponse
    err = json.Unmarshal(body, &mediaInitResponse)

    if err != nil {
        return nil, err
    }

    fmt.Println("Initialized media: ", mediaInitResponse);

    return &mediaInitResponse, nil
}

func (self *Twitter) MediaAppend(mediaId uint64, media []byte) error {
    step := 500 * 1024
    for s := 0; s * step < len(media); s++ {
        var body bytes.Buffer
        rangeBegining := s * step
        rangeEnd := (s + 1) * step
        if rangeEnd > len(media) {
            rangeEnd = len(media)
        }

        fmt.Println("try to append ", rangeBegining, "-", rangeEnd)

        w := multipart.NewWriter(&body)

        w.WriteField("command", "APPEND")
        w.WriteField("media_id", fmt.Sprint(mediaId))
        w.WriteField("segment_index", fmt.Sprint(s))

        fw, err := w.CreateFormFile("media", "example.mp4")

        fmt.Println(body.String())

        n, err := fw.Write(media[rangeBegining:rangeEnd])

        fmt.Println("len ", n)

        w.Close()

        req, err := http.NewRequest("POST", MediaUpload, &body)

        req.Header.Add("Content-Type", w.FormDataContentType())

        res, err := self.client.Do(req)
        if err != nil {
            return err
        }

        resBody, err := ioutil.ReadAll(res.Body)
        fmt.Println("append response ", string(resBody))
    }

    return nil
}

func (self *Twitter) MediaFinilize(mediaId uint64) error {
    form := url.Values{}
    form.Add("command", "FINALIZE")
    form.Add("media_id", fmt.Sprint(mediaId))

    req, err := http.NewRequest("POST", MediaUpload, strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    res, err := self.client.Do(req)
    if err != nil {
        return err
    }

    body, err := ioutil.ReadAll(res.Body)
    fmt.Println("final response ", string(body))

    return nil
}

func (self *Twitter) UpdateStatusWithMedia(text string, mediaId uint64) error {
    form := url.Values{}
    form.Add("status", text)
    form.Add("media_ids", fmt.Sprint(mediaId))

    req, err := http.NewRequest("POST", StatusUpdate, strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    res, err := self.client.Do(req)
    if err != nil {
        return err
    }

    body, err := ioutil.ReadAll(res.Body)
    fmt.Println("status response ", string(body))

    return nil
}
