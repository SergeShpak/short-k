package model

type (
	URL      string
	ShortURL string
)

type ShortenURLRequest struct {
	URL URL
}

type ShortenURLResponse struct {
	URL      URL
	ShortURL ShortURL
}

type GetFullURLRequest struct {
	ShortURL ShortURL
}

type GetFullURLResponse struct {
	URL string
}
