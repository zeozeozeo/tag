package tag

import (
	"fmt"
	"io"
)

func setWavOffset(r io.ReadSeeker) error {
	// verify RIFF format
	str, err := readString(r, 4)
	if err != nil {
		return err
	}
	if str != "RIFF" {
		return fmt.Errorf("format %v does not match expected 'RIFF'", str)
	}

	// verify WAVE format
	_, err = r.Seek(4, io.SeekCurrent)
	if err != nil {
		return err
	}
	str, err = readString(r, 4)
	if err != nil {
		return err
	}
	if str != "WAVE" {
		return fmt.Errorf("filetype %v does not match exptected 'WAVE'", str)
	}

	// identify chunk length
	_, err = r.Seek(24, io.SeekCurrent) // 24-byte data format chunk is unneeded
	if err != nil {
		return err
	}
	str, err = readString(r, 4)
	if err != nil {
		return err
	}
	if str != "data" {
		return fmt.Errorf("identifier %v does not match expected 'data'", err)
	}
	dataSize, err := readUint32LittleEndian(r)
	if err != nil {
		return err
	}

	_, err = r.Seek(int64(dataSize), io.SeekCurrent)
	if err != nil {
		return err
	}

	// skip past 8-byte Serato-specific identifier & size
	// keep in mind, no documentation for this
	str, err = readString(r, 3)
	if err != nil {
		return err
	}
	r.Seek(-3, io.SeekCurrent)
	if str == "id3" {
		_, err = r.Seek(8, io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	return nil
}
