package alt

import (
	"bytes"
	"io"
	"testing"
)

func TestNewReadCloserSplitter(t *testing.T) {
	byteReadCloser := io.NopCloser(bytes.NewReader([]byte("moo quack plop")))

	splitter := NewReadCloserSplitter(byteReadCloser)

	readCloser1 := splitter.NewReadCloser()
	//_ = splitter.NewReadCloser()

	read1 := make([]byte, 3)
	n, err := readCloser1.Read(read1)
	if err != nil {
		t.Error("Error reading first reader from first readCloser")
	}
	if n != 3 {
		t.Error("Expected 3 bytes read from first readCloser")
	}
	if string(read1) != "moo" {
		t.Error("First reader did not read expected string")
	}

}

func TestSplitReadCloser_getUnreadBytes(t *testing.T) {
	tests := []struct {
		name        string
		unreadBytes []*[]byte
		chunks      [][]byte
	}{
		{"nothing", nil, makeChunks("")},
		{"abc", makeUnreadBytes("abc"), makeChunks("abc")},
		{"a,bc", makeUnreadBytes("a", "bc"), makeChunks("abc")},
		{"abc-ab", makeUnreadBytes("abc"), makeChunks("ab")},
		{"abc-ab,c", makeUnreadBytes("abc"), makeChunks("ab", "c")},
		{"moo quack plop-moo, qua,ck p", makeUnreadBytes("moo quack plop"), makeChunks("moo", " qua", "ck p")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SplitReadCloser{
				unreadBytes: tt.unreadBytes,
			}
			for _, chunk := range tt.chunks {
				sizeOfChunk := len(chunk)
				bytes := make([]byte, sizeOfChunk)
				if n := s.getUnreadBytes(bytes); n != sizeOfChunk {
					t.Errorf("getUnreadBytes() = %v, want %v", n, sizeOfChunk)
				}
				if string(bytes) != string(chunk) {
					t.Errorf("getUnreadBytes() = %v, want %v", string(bytes), string(chunk))
				}
			}
		})
	}

	t.Run("data safety", func(t *testing.T) {
		data1 := []byte("something")
		data2 := []byte("stupid")

		s1 := &SplitReadCloser{
			unreadBytes: []*[]byte{&data1, &data2},
		}
		s2 := &SplitReadCloser{
			unreadBytes: []*[]byte{&data1, &data2},
		}
		read1 := make([]byte, 10)
		if n := s1.getUnreadBytes(read1); n != 10 {
			t.Errorf("getUnreadBytes() = %v, want %v", n, 2)
		}
		if string(read1) != "somethings" {
			t.Errorf("getUnreadBytes() = %v, want %v", string(read1), "so")
		}
		read2 := make([]byte, 2)
		if n := s2.getUnreadBytes(read2); n != 2 {
			t.Errorf("getUnreadBytes() = %v, want %v", n, 2)
		}
		if string(read2) != "so" {
			t.Errorf("getUnreadBytes() = %v, want %v", string(read2), "so")
		}
	})
}

func makeUnreadBytes(slices ...string) []*[]byte {
	pointerSlice := make([]*[]byte, len(slices))
	for i := range slices {
		bytes := []byte(slices[i])
		pointerSlice[i] = &bytes
	}
	return pointerSlice
}

func makeChunks(chunks ...string) [][]byte {
	chunksSlice := make([][]byte, len(chunks))
	for i := range chunks {
		chunksSlice[i] = []byte(chunks[i])
	}
	return chunksSlice
}
