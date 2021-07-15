package gen

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

func packageName(name string) string {
	spl := strings.Split(name, ".")
	return spl[0]
}

func structName(name string) string {
	spl := strings.Split(name, ".")
	return spl[len(spl)-1]
}

func alias(name string) string {
	pkg := packageName(name)
	hash := sha256.Sum256([]byte(pkg))
	str := hex.EncodeToString(hash[:])
	r := regexp.MustCompile(`[0-9]`)
	return r.ReplaceAllString(str, "")[:8]
}

func imports(names ...string) map[string]string {
	result := make(map[string]string)
	for _, name := range names {
		result[packageName(name)] = alias(name)
	}
	return result
}
