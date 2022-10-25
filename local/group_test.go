package local

import (
	"testing"
)

func TestCalculateShard(t *testing.T) {
	tests := []struct {
		size   int
		shard  int
		shards int
		start  int
		count  int
	}{
		{22, 1, 10, 0, 3},
		{22, 2, 10, 3, 3},
		{22, 3, 10, 6, 2},
		{29, 10, 10, 27, 2},
		{2, 1, 6, 0, 1}, // more shards than elements
		{2, 2, 6, 1, 1}, // more shards than elements
		{2, 3, 6, 2, 0}, // more shards than elements
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if gotStart, gotCount := calculateShard(tt.size, tt.shard, tt.shards); gotStart != tt.start || gotCount != tt.count {
				t.Errorf("calculateShard() = %v, %v, want %v, %v", gotStart, gotCount, tt.start, tt.count)
			}
		})
	}

}
