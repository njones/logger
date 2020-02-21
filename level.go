package logger

// logLevel is defined so the we can supress different levels as needed
type logLevel uint16

// print is the Print(f,ln) log level (which is the same as Info), but not exported
const print logLevel = 1 << iota //`short:"INF" long:"Info" color:"green"`

// A bitwise representation of the different log levels
// NOTE: the comment tags are used when using go generate to
// generate parts of the logging code
const (
	Info  logLevel = 1 << iota //`short:"INF" color:"green"`
	Warn                       //`short:"WRN" color:"yellow"`
	Debug                      //`short:"DBG" color:"cyan"`
	Error                      //`short:"ERR" color:"magenta"`
	Trace                      //`short:"TRC" color:"blue"`
	Fatal                      //`short:"FAT" color:"red"`
	Panic                      //`short:"PAN" color:"red"`
	HTTP                       //`short:"-" long:"-" color:"-" fn:"ln"`
)
