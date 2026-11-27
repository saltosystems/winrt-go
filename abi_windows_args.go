//go:build windows

package winrt

// CallArgs flattens the arguments of a WinRT method call into the list
// syscall.SyscallN takes. Each part is a uintptr or a []uintptr (from
// Int64Arg, Uint64Arg or GUIDArg).
func CallArgs(parts ...interface{}) []uintptr {
	args := make([]uintptr, 0, len(parts)+4)
	for _, p := range parts {
		switch v := p.(type) {
		case uintptr:
			args = append(args, v)
		case []uintptr:
			args = append(args, v...)
		default:
			panic("winrt: CallArgs takes uintptr or []uintptr")
		}
	}
	return args
}

// Int64Arg lays out an int64 passed by value.
func Int64Arg(v int64) []uintptr { return Uint64Arg(uint64(v)) }
