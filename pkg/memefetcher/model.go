package memefetcher

// MemeMetadata represents metadata about a meme found on reddit
type MemeMetadata struct {
	PostLink  string
	Subreddit string
	Title     string
	URL       string
	Nsfw      bool
	Spoiler   bool
	Author    string
	Ups       int
}

// RedditMeme contains the img bytes of a reddit meme as well as metadata about that meme
type RedditMeme struct {
	Img []byte
	MemeMetadata *MemeMetadata
}
