package ase

import (
	"bytes"
	"encoding/binary"
	"io"
	"unicode/utf16"
)

type Group struct {
	nameLen uint16
	Name    string
	Colors  []Color
}

func (group *Group) read(r io.Reader) (err error) {
	if err = group.readNameLen(r); err != nil {
		return
	}

	return group.readName(r)
}

func (group *Group) readNameLen(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, &group.nameLen)
}

func (group *Group) readName(r io.Reader) (err error) {
	//	make array for our color name based on block length
	name := make([]uint16, group.nameLen)
	if err = binary.Read(r, binary.BigEndian, &name); err != nil {
		return
	}

	//	decode our name. we trim off the last byte since it's zero terminated
	group.Name = string(utf16.Decode(name[:len(name)-1]))

	return
}

func (group *Group) write(w io.Writer) (err error) {

	// Write group start headers (block entry, block length,  nameLen, name)
	if err = group.writeBlockStart(w); err != nil {
		return
	}

	if err = group.writeBlockLength(w); err != nil {
		return
	}

	if err = group.writeNameLen(w); err != nil {
		return
	}
	if err = group.writeName(w); err != nil {
		return
	}

	// Write group's colors
	for _, color := range group.Colors {
		if err = color.write(w); err != nil {
			return err
		}
	}

	// Write the group's closing headers
	if err = group.writeBlockEnd(w); err != nil {
		return
	}

	return nil
}

func (group *Group) writeBlockStart(w io.Writer) (err error) {
	return binary.Write(w, binary.BigEndian, groupStart)
}

func (group *Group) writeBlockEnd(w io.Writer) (err error) {
	return binary.Write(w, binary.BigEndian, groupEnd)
}

// Encode the color's name length.
func (group *Group) writeNameLen(w io.Writer) (err error) {
	// Adding one to the name length accounts for the zero-terminated character.
	return binary.Write(w, binary.BigEndian, group.NameLen()+1)
}

// Encode the group's name.
func (group *Group) writeName(w io.Writer) (err error) {
	name := utf16.Encode([]rune(group.Name))
	name = append(name, uint16(0))
	return binary.Write(w, binary.BigEndian, name)
}

// Helper function that returns the length of a group's name.
func (group *Group) NameLen() uint16 {
	return uint16(len(group.Name))
}

// Write color's block length as a part of the ASE encoding.
func (group *Group) writeBlockLength(w io.Writer) (err error) {
	blockLength := group.calculateBlockLength()
	if err = binary.Write(w, binary.BigEndian, blockLength); err != nil {
		return err
	}
	return
}

// Calculates the block length to be written based on the color's attributes.
func (group *Group) calculateBlockLength() int32 {
	buf := new(bytes.Buffer)
	group.writeNameLen(buf)
	group.writeName(buf)
	return int32(buf.Len())
}
