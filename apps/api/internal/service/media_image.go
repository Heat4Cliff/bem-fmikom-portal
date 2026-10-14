package service

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"mime"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type processedImage struct {
	width        int
	height       int
	format       string
	storedMIME   string
	storedExt    string
	storedBytes  []byte
	wasConverted bool
}

var acceptedImageTypes = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".jfif": "image/jpeg",
	".png": "image/png", ".webp": "image/webp", ".gif": "image/gif",
	".tif": "image/tiff", ".tiff": "image/tiff", ".bmp": "image/bmp",
}

func processMediaImage(data []byte, extension, declaredMIME string, maxBytes int64) (processedImage, error) {
	if len(data) == 0 {
		return processedImage{}, ErrMediaCorrupt
	}
	if int64(len(data)) > maxBytes {
		return processedImage{}, ErrMediaTooLarge
	}
	expectedMIME, supported := acceptedImageTypes[strings.ToLower(extension)]
	if !supported || !mediaMIMECompatible(declaredMIME, expectedMIME) {
		return processedImage{}, ErrMediaUnsupported
	}

	config, decoded, actualFormat, err := decodeMediaImage(data, strings.ToLower(extension))
	if errors.Is(err, errAnimatedGIF) {
		return processedImage{}, ErrMediaAnimationUnsupported
	}
	if errors.Is(err, ErrMediaDimensions) {
		return processedImage{}, ErrMediaDimensions
	}
	if err != nil {
		return processedImage{}, ErrMediaCorrupt
	}
	if !validMediaDimensions(config.Width, config.Height) || decoded.Bounds().Dx() != config.Width || decoded.Bounds().Dy() != config.Height {
		return processedImage{}, ErrMediaDimensions
	}
	if !formatMatchesMIME(actualFormat, expectedMIME) {
		return processedImage{}, ErrMediaUnsupported
	}

	storedMIME := expectedMIME
	storedExtension := strings.ToLower(extension)
	storedBytes := data
	converted := actualFormat == "tiff" || actualFormat == "bmp"
	if converted {
		var normalized cappedImageBuffer
		normalized.limit = maxBytes
		if err := png.Encode(&normalized, decoded); err != nil {
			if errors.Is(err, ErrMediaTooLarge) {
				return processedImage{}, ErrMediaTooLarge
			}
			return processedImage{}, ErrMediaCorrupt
		}
		storedBytes = normalized.Bytes()
		storedMIME = "image/png"
		storedExtension = ".png"
	}

	return processedImage{
		width: config.Width, height: config.Height,
		format: actualFormat, storedMIME: storedMIME, storedExt: storedExtension,
		storedBytes: storedBytes, wasConverted: converted,
	}, nil
}

func mediaMIMECompatible(declared, expected string) bool {
	if strings.TrimSpace(declared) == "" || strings.EqualFold(strings.TrimSpace(declared), "application/octet-stream") {
		return true
	}
	parsed, _, err := mime.ParseMediaType(declared)
	if err != nil {
		return false
	}
	parsed = strings.ToLower(parsed)
	switch parsed {
	case "image/x-tiff":
		parsed = "image/tiff"
	case "image/x-ms-bmp", "image/x-bmp":
		parsed = "image/bmp"
	}
	return parsed == expected
}

func formatMatchesMIME(format, expected string) bool {
	switch format {
	case "jpeg":
		return expected == "image/jpeg"
	case "png":
		return expected == "image/png"
	case "gif":
		return expected == "image/gif"
	case "webp":
		return expected == "image/webp"
	case "tiff":
		return expected == "image/tiff"
	case "bmp":
		return expected == "image/bmp"
	default:
		return false
	}
}

func decodeMediaImage(data []byte, extension string) (image.Config, image.Image, string, error) {
	reader := bytes.NewReader(data)
	var config image.Config
	var format string
	var err error
	switch extension {
	case ".jpg", ".jpeg", ".jfif", ".png":
		config, format, err = image.DecodeConfig(reader)
	case ".gif":
		var frames int
		frames, err = gifFrameCount(data)
		if err == nil && frames > 1 {
			return image.Config{}, nil, "", errAnimatedGIF
		}
		if err == nil {
			config, err = gif.DecodeConfig(bytes.NewReader(data))
		}
		format = "gif"
	case ".tif", ".tiff":
		config, err = tiff.DecodeConfig(bytes.NewReader(data))
		format = "tiff"
	case ".bmp":
		config, err = bmp.DecodeConfig(bytes.NewReader(data))
		format = "bmp"
	case ".webp":
		config, err = webp.DecodeConfig(bytes.NewReader(data))
		format = "webp"
	default:
		return image.Config{}, nil, "", ErrMediaUnsupported
	}
	if err != nil {
		return config, nil, format, err
	}
	if !validMediaDimensions(config.Width, config.Height) {
		return config, nil, format, ErrMediaDimensions
	}
	var decoded image.Image
	switch extension {
	case ".jpg", ".jpeg", ".jfif", ".png":
		decoded, format, err = image.Decode(bytes.NewReader(data))
	case ".gif":
		decoded, err = gif.Decode(bytes.NewReader(data))
	case ".tif", ".tiff":
		decoded, err = tiff.Decode(bytes.NewReader(data))
	case ".bmp":
		decoded, err = bmp.Decode(bytes.NewReader(data))
	case ".webp":
		decoded, err = webp.Decode(bytes.NewReader(data))
	}
	return config, decoded, format, err
}

var errAnimatedGIF = errors.New("animated GIF is not supported")

// gifFrameCount validates GIF block boundaries without decompressing frame pixels.
// Full pixel decoding is only performed after the file is known to contain one frame.
func gifFrameCount(data []byte) (int, error) {
	if len(data) < 14 || (string(data[:6]) != "GIF87a" && string(data[:6]) != "GIF89a") {
		return 0, ErrMediaCorrupt
	}
	position := 13
	if data[10]&0x80 != 0 {
		position += 3 * (1 << ((data[10] & 0x07) + 1))
	}
	if position > len(data) {
		return 0, ErrMediaCorrupt
	}
	frames := 0
	for position < len(data) {
		introducer := data[position]
		position++
		switch introducer {
		case 0x3B:
			if position != len(data) {
				return 0, ErrMediaCorrupt
			}
			return frames, nil
		case 0x21:
			if position >= len(data) {
				return 0, ErrMediaCorrupt
			}
			position++ // extension label
			var err error
			position, err = skipGIFSubBlocks(data, position)
			if err != nil {
				return 0, err
			}
		case 0x2C:
			if position+9 > len(data) {
				return 0, ErrMediaCorrupt
			}
			frameWidth := int(binary.LittleEndian.Uint16(data[position+4 : position+6]))
			frameHeight := int(binary.LittleEndian.Uint16(data[position+6 : position+8]))
			if !validMediaDimensions(frameWidth, frameHeight) {
				return 0, ErrMediaDimensions
			}
			packed := data[position+8]
			frames++
			if frames > 1 {
				return frames, nil
			}
			position += 9
			if packed&0x80 != 0 {
				position += 3 * (1 << ((packed & 0x07) + 1))
			}
			if position >= len(data) {
				return 0, ErrMediaCorrupt
			}
			position++ // LZW minimum code size
			var err error
			position, err = skipGIFSubBlocks(data, position)
			if err != nil {
				return 0, err
			}
		default:
			return 0, ErrMediaCorrupt
		}
	}
	return 0, ErrMediaCorrupt
}

func skipGIFSubBlocks(data []byte, position int) (int, error) {
	for position < len(data) {
		length := int(data[position])
		position++
		if length == 0 {
			return position, nil
		}
		if position+length > len(data) {
			return 0, ErrMediaCorrupt
		}
		position += length
	}
	return 0, ErrMediaCorrupt
}

type cappedImageBuffer struct {
	bytes.Buffer
	limit int64
}

func (b *cappedImageBuffer) Write(value []byte) (int, error) {
	if int64(b.Len()+len(value)) > b.limit {
		return 0, ErrMediaTooLarge
	}
	return b.Buffer.Write(value)
}
