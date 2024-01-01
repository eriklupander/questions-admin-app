package views

import (
	"context"
	"os"
	"testing"
)

func TestHello(t *testing.T) {
	Index2().Render(context.Background(), os.Stdout)
}
