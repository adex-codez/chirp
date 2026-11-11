package profilepicture

import (
	"bytes"
	"errors"
	"image/jpeg"
	"io"
)

const MaxUploadBytes = 2 << 20

const (
	maxDimension = 1024
	maxPixels    = 1024 * 1024
)

type JPEG struct {
	data []byte
}

var (
	errInvalidJPEG = errors.New("choose a valid JPEG picture")
	errTooLarge    = errors.New("choose a picture no larger than 2 MiB")
	errDimensions  = errors.New("choose a JPEG picture no larger than 1024 by 1024 pixels")
	errUnreadable  = errors.New("could not read the picture, choose it again")
)

func ReadJPEG(body io.Reader) (JPEG, error) {
	data, err := io.ReadAll(io.LimitReader(body, MaxUploadBytes+1))
	if err != nil {
		return JPEG{}, errUnreadable
	}
	if len(data) > MaxUploadBytes {
		return JPEG{}, errTooLarge
	}
	if len(data) == 0 {
		return JPEG{}, errInvalidJPEG
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return JPEG{}, errInvalidJPEG
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > maxDimension || config.Height > maxDimension || config.Width*config.Height > maxPixels {
		return JPEG{}, errDimensions
	}
	image, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return JPEG{}, errInvalidJPEG
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, image, &jpeg.Options{Quality: 85}); err != nil {
		return JPEG{}, errInvalidJPEG
	}
	if encoded.Len() > MaxUploadBytes {
		return JPEG{}, errTooLarge
	}
	return JPEG{data: encoded.Bytes()}, nil
}
