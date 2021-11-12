package imagehandler

import (
	"io"
	"os"

	"github.com/openshift/assisted-image-service/pkg/isoeditor"
)

type baseFile interface {
	Size() (int64, error)
	InsertIgnition(*isoeditor.IgnitionContent) (io.ReadSeeker, error)
}

type baseFileData struct {
	filename string
	size     int64
}

func (bf *baseFileData) Size() (int64, error) {
	if bf.size == 0 {
		fi, err := os.Stat(bf.filename)
		if err != nil {
			return 0, err
		}
		bf.size = fi.Size()
	}
	return bf.size, nil
}

type baseIso struct {
	baseFileData
}

func newBaseIso(filename string) *baseIso {
	return &baseIso{baseFileData{filename: filename}}
}

func (biso *baseIso) InsertIgnition(ignition *isoeditor.IgnitionContent) (io.ReadSeeker, error) {
	return isoeditor.NewRHCOSStreamReader(biso.filename, ignition, nil)
}
