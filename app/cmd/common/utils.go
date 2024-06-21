package common

import (
	"fmt"
	"net/url"
	"path"

	"strconv"
	"strings"
)

func GetPath(nodeID uint64) string {
	return fmt.Sprintf("%s%d", NodeExplorePath, nodeID)
}

func ParseNodeId(path string) (uint64, error) {
	str := strings.Split(path, "/")[3]
	nodeID, err := strconv.Atoi(str)
	return uint64(nodeID), err
}

func JoinURL(base string, paths ...string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	joinedPath := path.Join(paths...)
	baseURL.Path = path.Join(baseURL.Path, joinedPath)

	return baseURL.String(), nil
}
