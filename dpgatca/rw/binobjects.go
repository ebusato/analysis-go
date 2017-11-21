package rw

import (
	"fmt"
	"reflect"
)

const (
	numAMCFrameCounters      uint8  = 2
	numASMFrameCounters      uint8  = 4
	numCounters              uint8  = 4
	numTimeStampsASM         uint8  = 4
	numTimeStampsTrigThorASM uint8  = 4
	numPatterns              uint8  = 3
	numThorTrigTimeStamps    uint8  = 3
	numCptsTriggerThor       uint8  = 2
	numCptsTriggerASM        uint8  = 2
	ctrlStartOfFrame         uint16 = 0x1230
	ctrl0xfe                 uint16 = 0xfe
	ctrl0xfd                 uint16 = 0xfd
	ctrl0xCafe               uint16 = 0xCAFE
	ctrl0xDeca               uint16 = 0xDECA
	ctrl0xCRC                uint16 = 0x9876
	ctrl0xfb                 uint16 = 0xfb
)

func print(in interface{}) {
	v := reflect.ValueOf(in)
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("  -> %v = %x (%v)\n", v.Type().Field(i).Name, v.Field(i).Interface(), v.Field(i).Interface())
	}
}

type ErrorCode int

const (
	ErrorCode1 ErrorCode = iota + 1 // value of error code if block has 4 extra 16 bits words after each sample block
)

type FileHeader struct {
	ModeFile  uint32
	FEId      uint16 // not clear why I must put 16 bits while Daniel has 8 bits in frame.h
	NoSamples uint16
	Time      uint64 // clarify what this is
}

func (f *FileHeader) Print() {
	fmt.Println("File Header:")
	print(*f)
}

type FrameHeader struct {
	StartOfFrame            uint16
	NbFrameAmcMsb           uint16
	NbFrameAmcLsb           uint16
	FEIdK30                 uint16
	Mode                    uint16
	TriggerType             uint16
	NoFrameAsmMsb           uint16
	NoFrameAsmOsb           uint16
	NoFrameAsmUsb           uint16
	NoFrameAsmLsb           uint16
	Cafe                    uint16
	Deca                    uint16
	UndefinedMsb            uint16
	UndefinedOsb            uint16
	UndefinedUsb            uint16
	UndefinedLsb            uint16
	TimeStampAsmMsb         uint16
	TimeStampAsmOsb         uint16
	TimeStampAsmUsb         uint16
	TimeStampAsmLsb         uint16
	TimeStampTrigThorAsmMsb uint16
	TimeStampTrigThorAsmOsb uint16
	TimeStampTrigThorAsmUsb uint16
	TimeStampTrigThorAsmLsb uint16
	ThorTT                  uint16
	PatternMsb              uint16
	PatternOsb              uint16
	PatternLsb              uint16
	Bobo                    uint16
	ThorTrigTimeStampMsb    uint16
	ThorTrigTimeStampOsb    uint16
	ThorTrigTimeStampLsb    uint16
	CptTriggerThorMsb       uint16
	CptTriggerThorLsb       uint16
	CptTriggerAsmMsb        uint16
	CptTriggerAsmLsb        uint16
	NoSamples               uint16

	// derived quantities
	FEId uint16
}

func (f *FrameHeader) Print() {
	fmt.Println("Frame Header:")
	print(*f)
}

func (f *FrameHeader) Integrity() error {
	if f.StartOfFrame != ctrlStartOfFrame {
		return fmt.Errorf("asm: missing %x magic\n", ctrlStartOfFrame)
	}
	// 		if (f.ParityFEIdCtrl & 0xff) != ctrl0xfe {
	// 			return fmt.Errorf("asm: missing %x magic\n", ctrl0xfe)
	// 		}
	if f.Cafe != ctrl0xCafe {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCafe)
	}
	if f.Deca != ctrl0xDeca {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xDeca)
	}
	return nil
}

type FrameTrailer struct {
	Crc uint16
	EoF uint16
}

func (f *FrameTrailer) Integrity() error {
	if f.Crc != ctrl0xCRC {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCRC)
	}
	if (f.EoF & 0xff) != ctrl0xfb {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfb)
	}
	// 	if (f.ParityFEIdCtrl2 & 0xff) != ctrl0xfb {
	// 		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfb)
	// 	}
	// 	if (f.ParityFEIdCtrl2&0x7fff)>>8 != f.FrontEndId {
	// 		log.Fatalf("Front end ids in header and trailer don't match\n")
	// 	}
	return nil
}

func (f *FrameTrailer) Print() {
	fmt.Println("Frame Trailer:")
	print(*f)
}

type ChanData struct {
	FirstChanWord  uint16
	SecondChanWord uint16
	Amplitudes     []uint16

	// Derived quantities
	Channel uint16
}

func (c *ChanData) Print() {
	fmt.Printf("FirstChanWord = %x (channel = %v)\n", c.FirstChanWord, c.Channel)
	fmt.Printf("SecondChanWord = %x\n", c.SecondChanWord)
	fmt.Printf("Amplitudes = ")
	for i := range c.Amplitudes {
		if (i+1)%16 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%04x ", c.Amplitudes[i])
		if i >= 80 {
			break
		}
	}
	fmt.Printf("\n")
}

type HalfDRSData struct {
	Data [4]ChanData
}

func (h *HalfDRSData) Print() {
	for i := range h.Data {
		fmt.Printf("Printing ChanData %v\n", i)
		h.Data[i].Print()
	}
}

// Frame is a single data frame produced by AMC
// Each frame is associated to one half DRS
type Frame struct {
	// 	FirstBlockWord        uint16                           // 0
	// 	AMCFrameCounters      [numAMCFrameCounters]uint16      // 1 -> 2
	// 	ParityFEIdCtrl        uint16                           // 3
	// 	TriggerMode           uint16                           // 4
	// 	Trigger               uint16                           // 5
	// 	ASMFrameCounters      [numASMFrameCounters]uint16      // 6 -> 9
	// 	Cafe                  uint16                           // 10
	// 	Deca                  uint16                           // 11
	// 	Counters              [numCounters]uint16              // 12 -> 15
	// 	TimeStampsASM         [numTimeStampsASM]uint16         // 16 -> 19
	// 	TimeStampsTrigThorASM [numTimeStampsTrigThorASM]uint16 // 20 -> 23
	// 	ThorTT                uint16                           // 24
	// 	Patterns              [numPatterns]uint16              // 25 -> 27
	// 	Bobo                  uint16                           // 28
	// 	ThorTrigTimeStamps    [numThorTrigTimeStamps]uint16    // 29 -> 31
	// 	CptsTriggerThor       [numCptsTriggerThor]uint16       // 32 -> 33
	// 	CptsTriggerASM        [numCptsTriggerASM]uint16        // 34 -> 35
	// 	NoSamples             uint16                           // 36
	Header  FrameHeader
	Data    HalfDRSData
	Trailer FrameTrailer

	// Derived quantities
	AMCFrameCounter      uint32
	ASMFrameCounter      uint64
	TimeStampASM         uint64
	TimeStampTrigThorASM uint64
	Pattern              uint64
	ThorTrigTimeStamp    uint64
	CptTriggerThor       uint32
	CptTriggerASM        uint32
	QuartetAbsIdx60      uint8

	// Error handling
	Err ErrorCode

	// UDP Payload size in octects
	UDPPayloadSize int

	// for internal use
	noAttempts         int
	QuartetAbsIdx60old uint8
}

func (f *Frame) SetDataSliceLen(noSamples int) {
	for i := range f.Data.Data {
		chandata := &f.Data.Data[i]
		chandata.Amplitudes = make([]uint16, noSamples)
	}
}

/*func NewFrame(udppayloadsize int) *Frame {
	f := &Frame{}
	f.UDPPayloadSize = udppayloadsize
	return f
}*/

/*
func (f *Frame) FillHeader(buffer []byte) {
	// 	buffer = buffer[:42]
	// 	f.FirstBlockWord = binary.BigEndian.Uint16(buffer[0:2])
	// 	f.AMCFrameCounters[0] = binary.BigEndian.Uint16(buffer[2:4])
	// 	f.AMCFrameCounters[1] = binary.BigEndian.Uint16(buffer[4:6])
	// 	f.ParityFEIdCtrl = binary.BigEndian.Uint16(buffer[6:8])
	// 	f.TriggerMode = binary.BigEndian.Uint16(buffer[8:10])
	// 	f.Trigger = binary.BigEndian.Uint16(buffer[10:12])
	// 	f.ASMFrameCounters[0] = binary.BigEndian.Uint16(buffer[12:14])
	// 	f.ASMFrameCounters[1] = binary.BigEndian.Uint16(buffer[14:16])
	// 	f.ASMFrameCounters[2] = binary.BigEndian.Uint16(buffer[16:18])
	// 	f.ASMFrameCounters[3] = binary.BigEndian.Uint16(buffer[18:20])
	// 	f.Cafe = binary.BigEndian.Uint16(buffer[20:22])
	// 	f.Deca = binary.BigEndian.Uint16(buffer[22:24])
	// 	f.Counters[0] = binary.BigEndian.Uint16(buffer[24:26])
	// 	f.Counters[1] = binary.BigEndian.Uint16(buffer[26:28])
	// 	f.Counters[2] = binary.BigEndian.Uint16(buffer[28:30])
	// 	f.Counters[3] = binary.BigEndian.Uint16(buffer[30:32])
	// 	f.TimeStampsASM[0] = binary.BigEndian.Uint16(buffer[32:34])
	// 	f.TimeStampsASM[1] = binary.BigEndian.Uint16(buffer[34:36])
	// 	f.TimeStampsASM[2] = binary.BigEndian.Uint16(buffer[36:38])
	// 	f.TimeStampsASM[3] = binary.BigEndian.Uint16(buffer[38:40])
	// 	f.TimeStampsTrigThorASM[0] = binary.BigEndian.Uint16(buffer[40:42])
	// 	f.TimeStampsTrigThorASM[1] = binary.BigEndian.Uint16(buffer[42:44])
	// 	f.TimeStampsTrigThorASM[2] = binary.BigEndian.Uint16(buffer[44:46])
	// 	f.TimeStampsTrigThorASM[3] = binary.BigEndian.Uint16(buffer[46:48])
	// 	f.ThorTT = binary.BigEndian.Uint16(buffer[48:50])
	// 	f.Patterns[0] = binary.BigEndian.Uint16(buffer[50:52])
	// 	f.Patterns[1] = binary.BigEndian.Uint16(buffer[52:54])
	// 	f.Patterns[2] = binary.BigEndian.Uint16(buffer[54:56])
	// 	f.Bobo = binary.BigEndian.Uint16(buffer[56:58])
	// 	f.ThorTrigTimeStamps[0] = binary.BigEndian.Uint16(buffer[58:60])
	// 	f.ThorTrigTimeStamps[1] = binary.BigEndian.Uint16(buffer[60:62])
	// 	f.ThorTrigTimeStamps[2] = binary.BigEndian.Uint16(buffer[62:64])
	// 	f.CptsTriggerThor[0] = binary.BigEndian.Uint16(buffer[64:66])
	// 	f.CptsTriggerThor[1] = binary.BigEndian.Uint16(buffer[66:68])
	// 	f.CptsTriggerASM[0] = binary.BigEndian.Uint16(buffer[68:70])
	// 	f.CptsTriggerASM[1] = binary.BigEndian.Uint16(buffer[70:72])
	// 	f.NoSamples = binary.BigEndian.Uint16(buffer[72:74])
	// 	f.AMCFrameCounter = (uint32(f.AMCFrameCounters[0]) << 16) + uint32(f.AMCFrameCounters[1])
	// 	f.FrontEndId = (f.ParityFEIdCtrl & 0x7fff) >> 8
	// 	f.ASMFrameCounter = (uint64(f.ASMFrameCounters[0]) << 48) + (uint64(f.ASMFrameCounters[1]) << 32) + (uint64(f.ASMFrameCounters[2]) << 16) + uint64(f.ASMFrameCounters[3])
	// 	temp := (uint64(f.TimeStampsASM[0]) << 16) | uint64(f.TimeStampsASM[1])
	// 	temp = (temp << 32)
	// 	temp1 := (uint64(f.TimeStampsASM[2]) << 16) | uint64(f.TimeStampsASM[3])
	// 	// 	temp |= temp1
	// 	f.TimeStampASM = temp | temp1
	///////////////////////////////////////////////////////////////////////
	// This +11 is necessary but currently not really understood
	// 11 clock periods are generated by "machine d'etat" in ASM firmware
	// These additionnal 11 samples should currently be considered junk

	//f.Data.SetNoSamples(f.NoSamples + 11)
	///////////////////////////////////////////////////////////////////////

}
*/

// readParityChanIdCtrl is a temporary fix, until we understand where the additionnal 16 bits words come from
/*
func (f *Frame) fillParityChanIdCtrl(buffer []byte, i int) (bool, int) {
	data := &f.Data.Data[i]
	beg := 74 + i*2*1023 + 2*f.noAttempts
	end := beg + 2
	data.ParityChanIdCtrl = binary.BigEndian.Uint16(buffer[beg:end])

	// 	fmt.Printf("%v, %x (f.noAttempts=%v)\n", i, data.ParityChanIdCtrl, f.noAttempts)
	if (data.ParityChanIdCtrl & 0xff) != ctrl0xfd {
		//panic("(data.ParityChanIdCtrl & 0xff) != ctrl0xfd")
		return true, 0
	}
	data.Channel = (data.ParityChanIdCtrl & 0x7f00) >> 8
	if data.Channel != f.Data.Data[0].Channel+uint16(i) {
		//panic("reader.readParityChanIdCtrl: data.Channel != f.Data.Data[0].Channel+uint16(i)")
		return true, 0
	}
	f.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(f.FrontEndId, data.Channel, false)
	//fmt.Printf("   -> %v, %v, %v\n", data.Channel, f.QuartetAbsIdx60, f.QuartetAbsIdx60old)
	if i > 0 && f.QuartetAbsIdx60 != f.QuartetAbsIdx60old {
		//panic("i > 0 && f.QuartetAbsIdx60 != f.QuartetAbsIdx60old")
		return true, 0
	}
	f.QuartetAbsIdx60old = f.QuartetAbsIdx60
	return false, end
}
*/

/*
func (f *Frame) FillData(buffer []byte) {
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
		var pb bool
		var lastIdx int
		for pb, lastIdx = f.fillParityChanIdCtrl(buffer, i); pb == true; pb, lastIdx = f.fillParityChanIdCtrl(buffer, i) {
			//fmt.Println("i, noAttempts, lastIdx=", i, noAttempts, lastIdx)
			f.noAttempts++
			if f.noAttempts >= 4 {
				log.Fatalf("frame.FillData: f.noAttempts >= 4\n")
			}
		}
		if f.noAttempts == 1 {
			f.Err = ErrorCode1
		}
		f.noAttempts = 0
		//fmt.Printf("data.ParityChanIdCtrl = %x\n", data.ParityChanIdCtrl)
		for j := range data.Amplitudes {
			//data.Amplitudes[j] = binary.BigEndian.Uint16(buffer[44+2*j+i*2*1023+2*f.noAttempts : 46+2*j+i*2*1023+2*f.noAttempts])
			//fmt.Printf("j, lastIdx: %v, %v\n", j, lastIdx)
			data.Amplitudes[j] = binary.BigEndian.Uint16(buffer[lastIdx+2*j : lastIdx+2+2*j])
		}
		// 		for j := range data.Amplitudes {
		// 			fmt.Printf("data.Amplitudes[%v] = %x\n", j, data.Amplitudes[j])
		// 		}
	}
}
*/

/*
func (f *Frame) IntegrityData() error {
	for i := range f.Data.Data {
		if (f.Data.Data[i].ParityChanIdCtrl & 0xff) != ctrl0xfd {
			return fmt.Errorf("asm: missing %x magic\n", ctrl0xfd)
		}
	}
	return nil
}

func (f *Frame) FillTrailer(buffer []byte) {
	if f.Err == ErrorCode1 {
		f.CRC = binary.BigEndian.Uint16(buffer[len(buffer)-4 : len(buffer)-2])
		f.ParityFEIdCtrl2 = binary.BigEndian.Uint16(buffer[len(buffer)-2 : len(buffer)])
	} else {
		f.CRC = binary.BigEndian.Uint16(buffer[len(buffer)-12 : len(buffer)-10])
		f.ParityFEIdCtrl2 = binary.BigEndian.Uint16(buffer[len(buffer)-10 : len(buffer)-8])
	}
}

*/

/*
func (f *Frame) Buffer() []byte {
	var buffer []uint16
	// 	buffer = append(buffer, f.FirstBlockWord)
	// 	buffer = append(buffer, f.AMCFrameCounters[:]...)
	// 	buffer = append(buffer, f.ParityFEIdCtrl)
	// 	buffer = append(buffer, f.TriggerMode)
	// 	buffer = append(buffer, f.Trigger)
	// 	buffer = append(buffer, f.ASMFrameCounters[:]...)
	// 	buffer = append(buffer, f.Cafe)
	// 	buffer = append(buffer, f.Deca)
	// 	buffer = append(buffer, f.Counters[:]...)
	// 	buffer = append(buffer, f.TimeStampsASM[:]...)
	// 	buffer = append(buffer, f.TimeStampsTrigThorASM[:]...)
	// 	buffer = append(buffer, f.ThorTT)
	// 	buffer = append(buffer, f.Patterns[:]...)
	// 	buffer = append(buffer, f.Bobo)
	// 	buffer = append(buffer, f.ThorTrigTimeStamps[:]...)
	// 	buffer = append(buffer, f.CptsTriggerThor[:]...)
	// 	buffer = append(buffer, f.CptsTriggerASM[:]...)
	// 	buffer = append(buffer, f.NoSamples)
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
		buffer = append(buffer, data.ParityChanIdCtrl)
		buffer = append(buffer, data.Amplitudes[:]...)
		if f.Err == ErrorCode1 {
			//fmt.Println("ErrorCode1, add extra word")
			buffer = append(buffer, uint16(0))
		}
	}
	buffer = append(buffer, f.CRC)
	buffer = append(buffer, f.ParityFEIdCtrl2)

	var buffer8 []byte
	for i := range buffer {
		buffer8 = append(buffer8, uint8(buffer[i]>>8))
		buffer8 = append(buffer8, uint8(buffer[i]&0xFFFF))
		//fmt.Printf("buffer8 = %x %x\n", buffer8[len(buffer8)-2], buffer8[len(buffer8)-1])
	}
	return buffer8
}
*/
