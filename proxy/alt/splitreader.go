package alt

import (
	"errors"
	"io"
)

//TODO implement mutexes

type ReadCloserSplitter struct {
	ReadCloser   io.ReadCloser
	maxBytesRead int64
	splits       map[int]*SplitReadCloser
	nextId       int
}

func NewReadCloserSplitter(readCloser io.ReadCloser) *ReadCloserSplitter {
	return &ReadCloserSplitter{
		ReadCloser: readCloser,
		splits:     make(map[int]*SplitReadCloser),
	}
}

func (s *ReadCloserSplitter) NewReadCloser() io.ReadCloser {
	id := s.nextId
	split := &SplitReadCloser{
		Id:       id,
		splitter: s,
	}
	s.splits[id] = split
	s.nextId++
	return split
}

func (s *ReadCloserSplitter) upstreamRead(toLength int64) error {
	if toLength <= s.maxBytesRead {
		return nil
	}
	toRead := toLength - s.maxBytesRead
	buf := make([]byte, toRead)
	n, err := s.ReadCloser.Read(buf)
	if n > 0 {
		for _, split := range s.splits {
			split.addUnreadBytes(buf[:n])
		}
	}
	if err != nil {
		for _, split := range s.splits {
			split.setUpstreamError(err)
		}
		return err
	}
	return nil
}

func (s *ReadCloserSplitter) CloseSplit(id int) error {
	if _, ok := s.splits[id]; !ok {
		return errors.New("reader already closed")
	}
	delete(s.splits, id)
	return nil
}

type SplitReadCloser struct {
	Id            int
	splitter      *ReadCloserSplitter
	bytesRead     int64
	unreadBytes   []*[]byte
	upstreamError error
}

var _ io.ReadCloser = &SplitReadCloser{}

func (s *SplitReadCloser) Read(p []byte) (n int, err error) {
	toLength := s.bytesRead + int64(len(p))
	s.splitter.upstreamRead(toLength)

	n = s.getUnreadBytes(p)

	if s.upstreamError != nil {
		return n, s.upstreamError
	}
	return n, nil
}

func (s *SplitReadCloser) addUnreadBytes(b []byte) {
	s.unreadBytes = append(s.unreadBytes, &b)
}

func (s *SplitReadCloser) setUpstreamError(err error) {
	s.upstreamError = err
}

func (s *SplitReadCloser) getUnreadBytes(p []byte) int {
	if s.unreadBytes == nil || len(s.unreadBytes) == 0 {
		return 0
	}
	read := 0
	remaining := len(p)
	for remaining > 0 && len(s.unreadBytes) > 0 {
		firstSlice := s.unreadBytes[0]
		lenFirstSlice := len(*firstSlice)
		if remaining < lenFirstSlice {
			// take 'remaining' x bytes from s[0] and leave rest in slice
			read += copy(p[read:read+remaining], (*firstSlice)[:remaining])
			newFirstSlice := (*firstSlice)[remaining:]
			s.unreadBytes[0] = &newFirstSlice
			remaining = 0
		} else { // remaining >= len s[0]
			// take all slice and delete s[0]
			read += copy(p[read:read+lenFirstSlice], *firstSlice)
			s.unreadBytes = s.unreadBytes[1:] // TODO maybe reslice to save on wasted mem?
			remaining -= lenFirstSlice
		}
	}
	return read
}

func (s *SplitReadCloser) Close() error {
	return s.splitter.CloseSplit(s.Id)
}
