package main

import (
	"time"
	"fmt"
	"os"
	"strings"

	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
	"github.com/tsmethurst/memebot3000/pkg/mastodon"
	"github.com/tsmethurst/memebot3000/pkg/memefetcher"
)

func main() {
	log.Info("starting up...")

	var memeEndpoint, mastodonURL, mastodonAccessToken string
	flag.StringVar(&memeEndpoint, "meme_metadata_endpoint", "https://meme-api.herokuapp.com/gimme", "endpoint for fetching reddit meme metadata")
	flag.StringVar(&mastodonURL, "mastodon_url", "", "endpoint for the mastodon API")
	flag.StringVar(&mastodonAccessToken, "mastodon_access_token", "", "")
	flag.Parse()

	mc := mastodon.NewClient(mastodonURL, mastodonAccessToken)

	mfClient, err := memefetcher.NewClient(memeEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	for {
		if err := postMeme(mfClient, mc); err != nil {
			log.Fatal(err)
		}
		log.Info("status posted!")
		time.Sleep(2 * time.Hour)
	}
}

func writeToDisk(meme *memefetcher.RedditMeme) (string, error) {
	split := strings.Split(meme.MemeMetadata.URL, "/")
	filename := fmt.Sprintf("/tmp/%s", split[len(split) - 1])
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("error creating file %s: %s", filename, err)
	}
	l, err := f.Write(meme.Img)
	if err != nil {
		f.Close()
		return filename, fmt.Errorf("error writing to file %s: %s", filename, err)
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		return filename, fmt.Errorf("error closing file %s: %s", filename, err)
	}
	return filename, nil
}

func removeFile(filename string) error {
	if filename == "" {
		return nil
	}

	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("error removing file %s: %s", filename, err)
	}
	return nil
}

func postMeme(mfClient memefetcher.Client, mc mastodon.Client) error {
	meme, err := mfClient.Gimme()
	if err != nil {
		return fmt.Errorf("could not fetch meme: %s", err)
	}

	filename, err := writeToDisk(meme)
	defer removeFile(filename)
	if err != nil {
		return fmt.Errorf("error writing %s: %s", filename, err)
	}
	log.Infof("fetched meme and wrote it to %s", filename)
	
	mediaID, err := mc.PublishMedia(filename) 
	if err != nil {
		return fmt.Errorf("error publishing media: %s", err)
	}

	var spoilerText string
	sensitive := meme.MemeMetadata.Nsfw
	if sensitive {
		spoilerText = "nsfw"
	}
	status := fmt.Sprintf("#meme posted by %s at %s\n\n\"%s\"", meme.MemeMetadata.Author, meme.MemeMetadata.PostLink, meme.MemeMetadata.Title)
	
	if err := mc.PublishStatus(status, mediaID, sensitive, spoilerText); err != nil {
		return fmt.Errorf("error publishing status: %s", err)
	}
	return nil
}
