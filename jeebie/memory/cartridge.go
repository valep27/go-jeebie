package memory

import (
	"fmt"

	"github.com/valerio/go-jeebie/jeebie/bit"
)

const titleLength = 11

const (
	entryPointAddress       = 0x100
	logoAddress             = 0x104
	titleAddress            = 0x134
	manufacturerCodeAddress = 0x13F
	cgbFlagAddress          = 0x143
	newLicenseCodeAddress   = 0x144
	sgbFlagAddress          = 0x146
	cartridgeTypeAddress    = 0x147
	romSizeAddress          = 0x148
	ramSizeAddress          = 0x149
	destinationCodeAddress  = 0x14A
	oldLicenseCodeAddress   = 0x14B
	versionNumberAddress    = 0x14C
	headerChecksumAddress   = 0x14D
	globalChecksumAddress   = 0x14E
)

// Cartridge holds the data and metadata of a gameboy cartridge.
type Cartridge struct {
	data           []byte
	title          string
	headerChecksum uint16
	globalChecksum uint16
	version        uint8
	cartType       uint8
	romSize        uint8
	ramSize        uint8
}

// NewCartridge creates an empty cartridge, useful only for debugging purposes.
func NewCartridge() *Cartridge {
	return &Cartridge{
		data: make([]byte, 0x10000),
	}
}

// NewCartridgeWithData initializes a new Cartridge from a slice of bytes.
func NewCartridgeWithData(bytes []byte) *Cartridge {
	// TODO: process metadata into actual types instead of just reading it (cart type, rom/ram size, etc.)

	titleBytes := bytes[titleAddress : titleAddress+titleLength]

	cart := &Cartridge{
		data:           make([]byte, len(bytes)),
		title:          string(titleBytes),
		headerChecksum: bit.Combine(bytes[headerChecksumAddress], bytes[headerChecksumAddress+1]),
		globalChecksum: bit.Combine(bytes[globalChecksumAddress], bytes[globalChecksumAddress+1]),
		version:        bytes[versionNumberAddress],
		cartType:       bytes[cartridgeTypeAddress],
		romSize:        bytes[romSizeAddress],
		ramSize:        bytes[ramSizeAddress],
	}

	copy(cart.data, bytes)

	fmt.Printf("Bytes %v\n", bytes)

	fmt.Printf("title %v\n", cart.title)

	return cart
}

// ReadByte reads a byte at the specified address. Does not check bounds, so the caller must make sure the
// address is valid for the cartridge.
func (c Cartridge) Read(addr uint16) uint8 {
	return c.data[addr]
}

// WriteByte attempts a write to the specified address. Writing to a cartridge has sense if the cartridge
// has extra RAM or for some special operations, like switching ROM banks.
func (c Cartridge) Write(addr uint16, value uint8) uint8 {
	return c.data[addr]
}
