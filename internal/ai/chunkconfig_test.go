package ai

import (
	"testing"

	"github.com/abhinavxd/libredesk/internal/stringutil"
)

func TestChunkConfigForLimit(t *testing.T) {
	def := stringutil.DefaultChunkConfig()

	// bge-small style: 480-token input. Chunks must fit, and the config must stay valid.
	got := chunkConfigForLimit(def, 480)
	if got.MaxTokens != 384 || got.MinTokens != 96 || got.OverlapTokens != 48 {
		t.Errorf("480 -> max %d min %d overlap %d, want 384/96/48", got.MaxTokens, got.MinTokens, got.OverlapTokens)
	}
	if _, err := stringutil.ChunkHTMLContent("t", "<p>hello</p>", got); err != nil {
		t.Errorf("shrunk config rejected by the chunker: %v", err)
	}

	// A tiny limit must never produce a config the chunker rejects.
	for _, limit := range []int{1, 2, 4, 9, 40} {
		if _, err := stringutil.ChunkHTMLContent("t", "<p>hello</p>", chunkConfigForLimit(def, limit)); err != nil {
			t.Errorf("limit %d: chunker rejected the config: %v", limit, err)
		}
	}

	// A large-context model keeps the defaults.
	if got := chunkConfigForLimit(def, 8192); got.MaxTokens != def.MaxTokens || got.MinTokens != def.MinTokens {
		t.Errorf("8192 changed the defaults: %+v", got)
	}
}
