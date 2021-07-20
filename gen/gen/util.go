package gen

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const homePkg = "adapter"

func packageName(name string) string {
	spl := strings.Split(name, ".")
	if len(spl) == 1 {
		return homePkg
	}
	return spl[0]
}

func structName(name string) string {
	spl := strings.Split(name, ".")
	return spl[len(spl)-1]
}

func alias(name string) string {
	pkg := packageName(name)
	if pkg == homePkg {
		return homePkg
	}
	hash := sha256.Sum256([]byte(pkg))
	str := hex.EncodeToString(hash[:])
	r := regexp.MustCompile(`[0-9]`)
	return r.ReplaceAllString(str, "")[:8]
}

func imports(names ...string) map[string]string {
	result := make(map[string]string)
	for _, name := range names {
		if packageName(name) == homePkg {
			continue
		}
		result[packageName(name)] = alias(name)
	}
	return result
}
