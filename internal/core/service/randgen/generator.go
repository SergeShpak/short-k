package randgen

import (
	"math/rand"

	"shortik/internal/core/service/randgen/model"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateRandomBytes(
	req model.GenerateRandomBytesRequest,
) (model.GenerateRandomBytesResponse, error) {
	resp := model.GenerateRandomBytesResponse{
		Bufs: make([][]byte, req.BufsCount),
	}
	for i := range req.BufsCount {
		resp.Bufs[i] = g.generateRandomBytes(req.Alphabet, req.Len)
	}
	return resp, nil
}

func (g *Generator) generateRandomBytes(alphabet []byte, n int) []byte {
	buf := make([]byte, n)
	for i := range n {
		b := alphabet[rand.Intn(len(alphabet))]
		buf[i] = b
	}
	return buf
}
