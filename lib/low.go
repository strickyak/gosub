package low

func Open(filename string, flags uint, mode uint) (fd int, errno int)
func Creat(filename string, mode uint) (fd int, errno int)
func Close(fd int) (errno int)

func Read(fd int, buf uintptr, size int) (count int, errno int)
func Write(fd int, buf uintptr, size int) (count int, errno int)

func O_RDONLY() uint
func O_WRONLY() uint
func O_RDWR() uint

func Exit(status byte)

// These use an intern static buffer, max 254 bytes.
func FormatToBuffer(format string, args ...interface{}) (count int)
func BufferToString() string
func WriteBuffer(fd int) (count int, errno int)
