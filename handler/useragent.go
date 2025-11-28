package handler

import (
	"regexp"
	"strconv"
	"strings"
)

type UAProfile struct {
	Head          string
	VersionMatch  string
	VersionTarget string
	Target        string
	ClashNewName  *bool // nil means undefined/don't change
	SurgeVer      int   // -1 means undefined
}

var (
	trueVal  = true
	falseVal = false
)

var UAMatchList = []UAProfile{
	{Head: "ClashForAndroid", VersionMatch: `\/([0-9.]+)`, VersionTarget: "2.0", Target: "clash", ClashNewName: &trueVal},
	{Head: "ClashForAndroid", VersionMatch: `\/([0-9.]+)R`, Target: "clashr", ClashNewName: &falseVal},
	{Head: "ClashForAndroid", Target: "clash", ClashNewName: &falseVal},
	{Head: "ClashforWindows", VersionMatch: `\/([0-9.]+)`, VersionTarget: "0.11", Target: "clash", ClashNewName: &trueVal},
	{Head: "ClashforWindows", Target: "clash", ClashNewName: &falseVal},
	{Head: "clash-verge", Target: "clash", ClashNewName: &trueVal},
	{Head: "ClashX Pro", Target: "clash", ClashNewName: &trueVal},
	{Head: "ClashX", VersionMatch: `\/([0-9.]+)`, VersionTarget: "0.13", Target: "clash", ClashNewName: &trueVal},
	{Head: "Clash", Target: "clash", ClashNewName: &trueVal},
	{Head: "Kitsunebi", Target: "v2ray"},
	{Head: "Loon", Target: "loon"},
	{Head: "Pharos", Target: "mixed"},
	{Head: "Potatso", Target: "mixed"},
	{Head: "Quantumult%20X", Target: "quanx"},
	{Head: "Quantumult", Target: "quan"},
	{Head: "Qv2ray", Target: "v2ray"},
	{Head: "Shadowrocket", Target: "mixed"},
	{Head: "Surfboard", Target: "surfboard"},
	{Head: "Surge", VersionMatch: `\/([0-9.]+).*x86`, VersionTarget: "906", Target: "surge", ClashNewName: &falseVal, SurgeVer: 4},
	{Head: "Surge", VersionMatch: `\/([0-9.]+).*x86`, VersionTarget: "368", Target: "surge", ClashNewName: &falseVal, SurgeVer: 3},
	{Head: "Surge", VersionMatch: `\/([0-9.]+)`, VersionTarget: "1419", Target: "surge", ClashNewName: &falseVal, SurgeVer: 4},
	{Head: "Surge", VersionMatch: `\/([0-9.]+)`, VersionTarget: "900", Target: "surge", ClashNewName: &falseVal, SurgeVer: 3},
	{Head: "Surge", Target: "surge", ClashNewName: &falseVal, SurgeVer: 2},
	{Head: "Trojan-Qt5", Target: "trojan"},
	{Head: "V2rayU", Target: "v2ray"},
	{Head: "V2RayX", Target: "v2ray"},
}

func verGreaterEqual(srcVer, targetVer string) bool {
	srcParts := strings.Split(srcVer, ".")
	targetParts := strings.Split(targetVer, ".")

	for i := 0; i < len(srcParts) && i < len(targetParts); i++ {
		s, err1 := strconv.Atoi(srcParts[i])
		t, err2 := strconv.Atoi(targetParts[i])
		if err1 != nil || err2 != nil {
			continue
		}
		if s < t {
			return false
		} else if s > t {
			return true
		}
	}

	if len(srcParts) >= len(targetParts) {
		return true
	}
	return false
}

func matchUserAgent(userAgent string) (target string, clashNewName *bool, surgeVer int) {
	surgeVer = -1
	if userAgent == "" {
		return
	}
	for _, x := range UAMatchList {
		if strings.HasPrefix(userAgent, x.Head) {
			if x.VersionMatch != "" {
				// Compile regex on the fly. For performance critical paths, these should be pre-compiled.
				// We extract the version from the first capturing group.
				re, err := regexp.Compile(x.VersionMatch)
				if err != nil {
					continue
				}
				matches := re.FindStringSubmatch(userAgent)
				if len(matches) < 2 {
					continue
				}
				version := matches[1]

				if x.VersionTarget != "" && !verGreaterEqual(version, x.VersionTarget) {
					continue
				}
			}
			target = x.Target
			clashNewName = x.ClashNewName
			if x.SurgeVer != 0 {
				surgeVer = x.SurgeVer
			}
			return
		}
	}
	return
}
