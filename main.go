package main

import (
    "github.com/kreshikhin/twitter-media-uploader/twitter"

    "github.com/mrjones/oauth"

    "fmt"
    "io/ioutil"
    "log"
    "encoding/json"
)

func ReadAccessToken(filePath string) *oauth.AccessToken {
    accessData, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Println(err)
        return nil
    }

    var access oauth.AccessToken
    err = json.Unmarshal(accessData, &access)
    if err != nil {
        fmt.Println("error:", err)
        return nil
    }

    return &access
}

func main() {
    consumerKey := "YOUR_COSUMER_KEY"
    consumerSecret := "YOUR_COSUMER_SECRET"

    c := oauth.NewConsumer(
        consumerKey,
        consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
        })

    c.Debug(false)

    accessToken := ReadAccessToken("twitter.json")

    if accessToken == nil {
        return;

        requestToken, u, err := c.GetRequestTokenAndUrl("oob")
        if err != nil {
            log.Fatal(err)
        }

        fmt.Println("(1) Go to: " + u)
        fmt.Println("(2) Grant access, you should get back a verification code.")
        fmt.Println("(3) Enter that verification code here: ")

        verificationCode := ""
        fmt.Scanln(&verificationCode)

        accessToken, err = c.AuthorizeToken(requestToken, verificationCode)
        if err != nil {
            log.Fatal(err)
        }
    }

    client, err := c.MakeHttpClient(accessToken)
    if err != nil {
        panic(err)
    }

    tw := twitter.NewTwitter(client)

    media, err := ioutil.ReadFile("test-video.mp4")
    if err != nil {
        panic(err)
    }

    tw.MakeTwitWithMedia("Hello World", media)
}
