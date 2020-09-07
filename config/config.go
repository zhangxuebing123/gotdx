package config

import (
	"bufio"
	"fmt"
	"github.com/axgle/mahonia"
	"os"
	"strconv"
	"strings"
)

// 通达信网关信息

const (
	HQHOST   = "HQHOST" // 证券行情
	EXHQHOST = "DSHOST" // 拓展行情
)

type TDXConfig struct {
	sections map[string]map[string]string
}

func (c *TDXConfig) Load() error {
	file, err := os.Open("connect.cfg")
	if err != nil {
		return err
	}
	defer file.Close()

	c.sections = make(map[string]map[string]string)
	reader := bufio.NewReader(file)
	var curSectionName string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ':' {
			continue
		}
		if line[0] == '[' && line[len(line)-1] == ']' {
			curSectionName = line[1 : len(line)-1]
			c.sections[curSectionName] = make(map[string]string)
		} else {
			pair := strings.Split(line, "=")
			if len(pair) == 2 {
				c.sections[curSectionName][strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
			}
		}
	}
	return nil
}

func (c *TDXConfig) Remoter(tag string) (m map[string]string) {
	m = make(map[string]string)
	if (tag == HQHOST) || (tag == EXHQHOST) {
		section, ok := c.sections[tag]
		if !ok {
			return
		}
		hostName, _ := strconv.Atoi(c.sections[tag]["HostNum"])
		enc := mahonia.NewDecoder("gbk")
		for index := 0; index < hostName; index++ {
			hostname := fmt.Sprintf("HostName%02d", index+1)
			ipaddress := fmt.Sprintf("IPAddress%02d", index+1)
			port := fmt.Sprintf("Port%02d", index+1)
			m[enc.ConvertString(section[hostname])] = section[ipaddress] + ":" + section[port]
		}
	}
	return
}
