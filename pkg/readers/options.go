package readers

import (
    //"github.com/helviojunior/adbcat/pkg/models"
)

// Options are global github.com/helviojunior/adbcatadbcat options
type Options struct {
    // Logging is logging options
    Logging Logging

    ExcludeFilterList []string

    IncludeFilterList []string

    LogFile string

    MinLevel string

    UseDevice bool
    UseEmulator bool
    DeviceSerial string

    PackageName string

    AdbBinPath string

    ClearOutput bool

    ShowTime bool
    ShowPid bool

    UseAnsiLog bool
}

// Logging is log related options
type Logging struct {
    // Debug display debug level logging
    Debug bool
    // Debug display debug level logging
    DebugDb bool
    // LogScanErrors log errors related to scanning
    LogScanErrors bool
    // Silence all logging
    Silence bool
}

// NewDefaultOptions returns Options with some default values
func NewDefaultOptions() *Options {
    return &Options{
        Logging: Logging{
            Debug:         true,
            LogScanErrors: true,
        },
        ExcludeFilterList: []string{},
        IncludeFilterList: []string{},
        LogFile: "",
        MinLevel: "V",
        UseDevice: false,
        UseEmulator: false,
        DeviceSerial: "",
        PackageName: "",
        AdbBinPath: "",
        ClearOutput: false,
        UseAnsiLog: false,
    }
}