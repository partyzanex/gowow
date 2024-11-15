package proto

//nolint:gochecknoglobals
var masks = []byte{
	0x00,
	0x80,
	0xC0,
	0xE0,
	0xF0,
	0xF8,
	0xFC,
	0xFE,
	0xFF,
}

// HasLeadingZeroBits returns true if the hash has leading zero bits.
func HasLeadingZeroBits(hash []byte, bits uint8) bool {
	// Check if the hash has leading zero bits.
	if len(hash)*8 < int(bits) {
		return false // not enough bits.
	}

	fullBytes := bits / 8     //nolint:mnd
	remainingBits := bits % 8 //nolint:mnd

	for i := 0; i < int(fullBytes); i++ { //nolint:intrange
		if hash[i] != 0 {
			return false
		}
	}

	if remainingBits > 0 {
		return hash[fullBytes]&masks[remainingBits] == 0
	}

	return true
}
