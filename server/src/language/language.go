package language

import (
	"common"
	"csvcfg"
	"fmt"
	"path"
	// "rpc"
	// "strconv"
	"strings"
)

type LanguageCfg struct {
	TID string
	CH  string
	EN  string
}

var languageCfg map[string]*[]LanguageCfg

type LocationLanguageCfg struct {
	TID string
	STR string
}

var locationLanguageCfg map[string]*[]LocationLanguageCfg

func init() {
	designerDir := common.GetDesignerDir()
	fullPath := path.Join(designerDir, "textsutf8.csv")
	csvcfg.LoadCSVConfig(fullPath, &languageCfg)
}

//支持变参Format
func GetLanguage(id string, args ...interface{}) string {
	id = strings.TrimSpace(strings.ToLower(id))

	if lan, ok := languageCfg[id]; ok {
		return fmt.Sprintf((*lan)[0].CH, args...)
	}

	return ""
}
