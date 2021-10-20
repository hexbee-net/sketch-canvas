package keygen

import (
	"math/rand"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/xerrors"
)

//go:generate mockery --name KeyGen
type KeyGen interface {
	Generate() string
}

type SnowflakeKeyGen struct {
	node *snowflake.Node
}

func New() (KeyGen, error) {
	const NodeMax = 1023

	node, err := snowflake.NewNode(rand.Int63n(NodeMax)) //nolint:gosec
	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate SnowFlake node: %w", err)
	}

	return &SnowflakeKeyGen{node: node}, nil
}

func (g *SnowflakeKeyGen) Generate() string {
	return g.node.Generate().String()
}
