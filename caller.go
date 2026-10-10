package apilogger

import (
	"fmt"
	"runtime"
	"strings"
)

// CallerInfo holds script and function metadata of the caller.
type CallerInfo struct {
	Path   string
	Class  string
	Method string
	Line   string
}

// CaptureCaller inspects the runtime call stack and extracts:
// file path, receiver type/class, function/method name, and line number.
func CaptureCaller(skip int) CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return CallerInfo{}
	}

	info := CallerInfo{
		Path: file,
		Line: fmt.Sprintf("%d", line),
	}

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		fullName := fn.Name()

		// Strip package URL path if any (e.g., github.com/user/repo/pkg.(*Type).Method)
		lastSlash := strings.LastIndex(fullName, "/")
		pkgAndName := fullName
		if lastSlash >= 0 {
			pkgAndName = fullName[lastSlash+1:]
		}

		parts := strings.Split(pkgAndName, ".")
		switch {
		case len(parts) >= 3:
			classPart := parts[len(parts)-2]
			classPart = strings.TrimPrefix(classPart, "(")
			classPart = strings.TrimPrefix(classPart, "*")
			classPart = strings.TrimSuffix(classPart, ")")
			info.Class = classPart
			info.Method = parts[len(parts)-1]
		case len(parts) == 2:
			info.Method = parts[1]
		default:
			info.Method = fullName
		}
	}

	return info
}
