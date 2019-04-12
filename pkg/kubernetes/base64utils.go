package kubernetes

import (
	"bufio"
	"bytes"
	"encoding/base64"
)

func base64EncodeString(src string) []byte {
	var buffer bytes.Buffer
	bufferedWriter := bufio.NewWriter(&buffer)

	encoder := base64.NewEncoder(base64.StdEncoding, bufferedWriter)
	encoder.Write([]byte(src))

	encoder.Close()
	bufferedWriter.Flush()
	return buffer.Bytes()
}

func base64EncodeMapOfStrings(src map[string]string) map[string][]byte {
	result := make(map[string][]byte)
	for key, value := range src {
		result[key] = base64EncodeString(value)
	}
	return result
}
