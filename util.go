package pciinfo

import (
	"io/ioutil"
	"strings"
)

// Read one-liner text files, strip newline.
func fileString(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

//// Write one-liner text files, add newline, ignore errors (best effort).
//func spewFile(path string, data string, perm os.FileMode) {
//	_ = ioutil.WriteFile(path, []byte(data+"\n"), perm)
//}
