package main

import (
	"log"
	"os"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"

	"github.com/at-wat/ebml-go/webm"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

// FIXME 保存的文件好像元数据有问题（x）大概是关键帧问题
// FIXME screen流一直没有videoKeyframe以致于文件无法保存
// FIXME 进度条无法拖动，可能因为是元数据的问题

type webmSaver struct {
	audioWriter, videoWriter       webm.BlockWriteCloser
	audioBuilder, videoBuilder     *samplebuilder.SampleBuilder
	audioTimestamp, videoTimestamp time.Duration
	file_name                      string
	// closed                         bool
}

func newWebmSaver(fname string) *webmSaver {
	return &webmSaver{
		audioBuilder: samplebuilder.New(10, &codecs.OpusPacket{}, 48000),
		videoBuilder: samplebuilder.New(10, &codecs.VP8Packet{}, 90000),
		file_name:    fname,
		// closed:       false,
	}
}

// func (s *webmSaver) Close() {
// 	if s.closed {
// 		return
// 	}
// 	s.closed = true

// 	if s.audioWriter != nil {
// 		if err := s.audioWriter.Close(); err != nil {
// 			log.Print(err)
// 		}
// 	}
// 	if s.videoWriter != nil {
// 		if err := s.videoWriter.Close(); err != nil {
// 			log.Print(err)
// 		}
// 	}
// }
func (s *webmSaver) PushOpus(rtpPacket *rtp.Packet) {
	// 不加这个会panic
	if s.audioBuilder == nil {
		return
	}
	s.audioBuilder.Push(rtpPacket)
	for {
		sample := s.audioBuilder.Pop()
		if sample == nil {
			return
		}
		if s.audioWriter != nil {
			s.audioTimestamp += sample.Duration
			if _, err := s.audioWriter.Write(true, int64(s.audioTimestamp/time.Millisecond), sample.Data); err != nil {
				log.Println("webmsaver: ", err)
				return
			}
		}
	}
}
func (s *webmSaver) PushVP8(rtpPacket *rtp.Packet) {
	// 不加这个会panic
	if s.videoBuilder == nil {
		return
	}
	s.videoBuilder.Push(rtpPacket)
	for {
		sample := s.videoBuilder.Pop()
		if sample == nil {
			return
		}
		// Read VP8 header.
		videoKeyframe := (sample.Data[0]&0x1 == 0)
		// FIXME 这里screen流一直没有videoKeyframe不知道咋回事
		if videoKeyframe {
			// Keyframe has frame information.
			raw := uint(sample.Data[6]) | uint(sample.Data[7])<<8 | uint(sample.Data[8])<<16 | uint(sample.Data[9])<<24
			width := int(raw & 0x3FFF)
			height := int((raw >> 16) & 0x3FFF)

			if s.videoWriter == nil || s.audioWriter == nil {
				// Initialize WebM saver using received frame size.
				s.InitWriter(width, height)
			}
		}
		if s.videoWriter != nil {
			s.videoTimestamp += sample.Duration
			if _, err := s.videoWriter.Write(videoKeyframe, int64(s.audioTimestamp/time.Millisecond), sample.Data); err != nil {
				log.Println("webmsaver: ", err)
				return
			}
		}
	}
}
func (s *webmSaver) InitWriter(width, height int) {
	w, err := os.OpenFile(s.file_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Println("webmsaver: ", err)
		return
	}

	ws, err := webm.NewSimpleBlockWriter(w,
		[]webm.TrackEntry{
			{
				Name:            "Audio",
				TrackNumber:     1,
				TrackUID:        12345,
				CodecID:         "A_OPUS",
				TrackType:       2,
				DefaultDuration: 20000000,
				Audio: &webm.Audio{
					SamplingFrequency: 48000.0,
					Channels:          2,
				},
			}, {
				Name:            "Video",
				TrackNumber:     2,
				TrackUID:        67890,
				CodecID:         "V_VP8",
				TrackType:       1,
				DefaultDuration: 33333333,
				Video: &webm.Video{
					PixelWidth:  uint64(width),
					PixelHeight: uint64(height),
				},
			},
		})
	if err != nil {
		log.Println("webmsaver: ", err)
		return
	}
	s.audioWriter = ws[0]
	s.videoWriter = ws[1]
}
