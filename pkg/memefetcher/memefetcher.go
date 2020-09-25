package memefetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client represents a client that can fetch memes
type Client interface {
	Gimme() (*RedditMeme, error)
}

// client implements Client
type client struct {
	fetchURL string
}

// NewClient returns a new memefetcher client ready to go
func NewClient(fetchURL string) (Client, error) {
	_, err := url.Parse(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("fetchAddress was not a valid url: %s", err)
	}
	return &client{
		fetchURL: fetchURL,
	}, nil
}

// Gimme grabs a nice fresh meme from reddit along with its associated metadata
func (c *client) Gimme() (*RedditMeme, error) {
	md, err := c.fetchMemeMetadata()
	if err != nil {
		return nil, fmt.Errorf("error with meme metadata: %s", err)
	}

	meme, err := c.fetchMeme(md)
	if err != nil {
		return nil, fmt.Errorf("error fetching meme: %s", err)
	}
	
	return meme, nil
}

func (c *client) fetchMemeMetadata() (*MemeMetadata, error) {
	resp, err := http.Get(c.fetchURL)
	if err != nil {
		return nil, fmt.Errorf("error doing initial meme metadata fetch to %s: %s", c.fetchURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("received code %d from %s", resp.StatusCode, c.fetchURL)
	}

	mdBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading meme metadata body: %s", err)
	}

	md := &MemeMetadata{}
	if err := json.Unmarshal(mdBody, md); err != nil {
		return nil, fmt.Errorf("error parsing meme metadata body: %s", err)
	}

	return md, nil
}

func (c *client) fetchMeme(md *MemeMetadata) (*RedditMeme, error) {
	resp, err := http.Get(md.URL)
	if err != nil {
		return nil, fmt.Errorf("error doing initial meme fetch to %s: %s", md.URL, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("received code %d from %s", resp.StatusCode, md.URL)
	}

	mBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading meme body: %s", err)
	}

	return &RedditMeme{
		Img: mBody,
		MemeMetadata: md,
	}, nil
}
