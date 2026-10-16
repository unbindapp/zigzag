package upload_service

import "encoding/binary"

func isWebP(header []byte) bool {
	return len(header) >= 16 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP"
}

// https://developers.google.com/speed/webp/docs/riff_container
func webPSize(header []byte) (*int, *int) {
	if !isWebP(header) || len(header) < 30 {
		return nil, nil
	}
	chunk := header[12:]
	var width, height int
	switch string(chunk[0:4]) {
	case "VP8X":
		width = 1 + int(uint32(chunk[12])|uint32(chunk[13])<<8|uint32(chunk[14])<<16)
		height = 1 + int(uint32(chunk[15])|uint32(chunk[16])<<8|uint32(chunk[17])<<16)
	case "VP8L":
		if chunk[8] != 0x2f {
			return nil, nil
		}
		bits := binary.LittleEndian.Uint32(chunk[9:13])
		width = 1 + int(bits&0x3fff)
		height = 1 + int((bits>>14)&0x3fff)
	case "VP8 ":
		if chunk[11] != 0x9d || chunk[12] != 0x01 || chunk[13] != 0x2a {
			return nil, nil
		}
		width = int(binary.LittleEndian.Uint16(chunk[14:16]) & 0x3fff)
		height = int(binary.LittleEndian.Uint16(chunk[16:18]) & 0x3fff)
	default:
		return nil, nil
	}
	return &width, &height
}
