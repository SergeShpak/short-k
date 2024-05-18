package model

type GenerateRandomBytesRequest struct {
	Alphabet  []byte
	BufsCount int
	Len       int
}

type GenerateRandomBytesResponse struct {
	Bufs [][]byte
}
