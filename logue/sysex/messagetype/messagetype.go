package messagetype

// Request / Dump types

// GlobalDataDumpRequest - get global data dump
const GlobalDataDumpRequest byte = 0x0E

// CurrentProgramDataDumpRequest - get current program dump
const CurrentProgramDataDumpRequest byte = 0x10

// LivesetDataDumpRequest - get liveset data dump
const LivesetDataDumpRequest byte = 0x16

// ProgramDataDumpRequest - get program dump
const ProgramDataDumpRequest byte = 0x1C

// CurrentProgramDataDump - current program dump
const CurrentProgramDataDump byte = 0x40

// LivesetDataDump - liveset dump
const LivesetDataDump byte = 0x46

// LivesetDataDump1Prog - liveset dump
const LivesetDataDump1Prog byte = 0x4C

// GlobalDataDump - global data dump
const GlobalDataDump byte = 0x51

// TuningScaleDataDumpRequest - get tuning scale data dump
const TuningScaleDataDumpRequest byte = 0x14

// TuningOctaveDataDumpRequest - get tuning octave data dump
const TuningOctaveDataDumpRequest byte = 0x15

// TuningScaleDataDump - tuning scale data dump
const TuningScaleDataDump byte = 0x44

// TuningOctaveDataDump - tuning octave data dump
const TuningOctaveDataDump byte = 0x45

// UserAPIVersionRequest - get API version dump
const UserAPIVersionRequest byte = 0x17

// UserModuleInfoRequest - get user module info
const UserModuleInfoRequest byte = 0x18

// UserSlotStatusRequest - get user slot status
const UserSlotStatusRequest byte = 0x19

// UserSlotDataRequest - get user slot data
const UserSlotDataRequest byte = 0x1A

// ClearUserSlot - clear user slot
const ClearUserSlot byte = 0x1B

// ClearUserModule - clear user module
const ClearUserModule byte = 0x1D

// SwapUserData - swap user data
const SwapUserData byte = 0x1E

// UserAPIVersion - user api version
const UserAPIVersion byte = 0x47

// UserModuleInfo - user module info
const UserModuleInfo byte = 0x48

// UserSlotStatus - user slot status
const UserSlotStatus byte = 0x49

// UserSlotData - user slot data
const UserSlotData byte = 0x4A

// Status/Error types

// DataLoadCompleted - data OK
const DataLoadCompleted byte = 0x23

// DataLoadError - load error
const DataLoadError byte = 0x24

// DataFormatError - format error
const DataFormatError byte = 0x26

// UserDataSizeError - user data size error
const UserDataSizeError byte = 0x27

// UserDataCRCError - user data CRC error
const UserDataCRCError byte = 0x28

// UserTargetError - user target error
const UserTargetError byte = 0x29

// UserAPIError - user API error
const UserAPIError byte = 0x2A

// UserLoadSizeError - user load size error
const UserLoadSizeError byte = 0x2B

// UserModuleError - user module error
const UserModuleError byte = 0x2C

// UserSlotError - user slot error
const UserSlotError byte = 0x2D

// UserFormatError - user format error
const UserFormatError byte = 0x2E

// UserInternalError - user internal error
const UserInternalError byte = 0x2F
