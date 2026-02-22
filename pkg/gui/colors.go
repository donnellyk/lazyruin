package gui

const (
	AnsiReset       = "\x1b[0m"
	AnsiDim         = "\x1b[2m"
	AnsiGreen       = "\x1b[32m"
	AnsiYellow      = "\x1b[33m"
	AnsiCyan        = "\x1b[36m"
	AnsiBlueBgWhite = "\x1b[44;37m"
	AnsiDimBg       = "\x1b[48;5;238m" // subtle dark gray background
	AnsiBgReset     = "\x1b[49m"      // reset background only
	AnsiBoldWhite   = "\x1b[1;37m"
	AnsiGreen1      = "\x1b[38;5;22m" // dark green (1 note)
	AnsiGreen2      = "\x1b[38;5;28m" // medium green (2 notes)
	AnsiGreen3      = "\x1b[38;5;34m" // bright green (3+ notes)
)
