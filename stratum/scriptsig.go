package stratum

import "encoding/binary"

func ParseScriptSig(scriptSig []byte) map[string]any {
	out := make(map[string]any)
	height, ok := parseBIP34Height(scriptSig)
	if ok {
		out["block_height"] = height
	}
	return out
}

func parseBIP34Height(scriptSig []byte) (uint32, bool) {
	if len(scriptSig) == 0 {
		return 0, false
	}
	pushLen := int(scriptSig[0])
	if pushLen == 0 || pushLen > 5 || len(scriptSig) < 1+pushLen {
		return 0, false
	}
	b := scriptSig[1 : 1+pushLen]
	var height uint32
	switch len(b) {
	case 1:
		height = uint32(b[0])
	case 2:
		height = uint32(binary.LittleEndian.Uint16(b))
	case 3:
		height = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
	case 4:
		height = binary.LittleEndian.Uint32(b)
	case 5:
		height = binary.LittleEndian.Uint32(b[:4])
	default:
		return 0, false
	}
	return height, true
}

