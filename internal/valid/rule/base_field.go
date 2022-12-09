package rule

import (
	"fmt"
	v "github.com/ntt360/gin/valid"
	"reflect"
	"regexp"
	"strings"
)

type BaseFiled struct {
	CurParam      string            // filed form CurParam value
	CurKey        string            // current struct key name
	CurRule       Info              // current valid rule
	CurTag        reflect.StructTag // current valid field struct CurTag
	CurInheritTag reflect.StructTag // when dive valid parent has CurTag
	Track         []string          // current data index path
	CurDataIdx    string            // current data index string
	ValidTagName  string            // valid tag name: form、json、header
}

func (f *BaseFiled) Err() error {
	return f.WrapperErr(nil)
}

func (f *BaseFiled) WrapperErr(err error) error {
	msg, ok := f.CustomMsg()
	if !ok {
		msg = fmt.Sprintf(DefaultFieldMsg, f.CurParam)
	}

	return &v.Error{
		Code:     v.CodeParamsErr,
		Param:    f.CurParam,
		RuleName: f.CurRule.Name,
		Key:      f.CurKey,
		Type:     v.ErrorTypeField,
		Msg:      msg,
		CauseErr: err,
	}
}

func (f *BaseFiled) CustomMsg() (string, bool) {
	customMsg := f.CurTag.Get(KeyMsg)
	if customMsg == "" {
		customMsg = f.CurInheritTag.Get(KeyMsg)
	}

	if customMsg != "" {
		customMsgMap := make(map[string]string)
		detailMsg := strings.Contains(customMsg, "default=")
		for s := range ValidatorFunc {
			if strings.Contains(customMsg, s+"=") {
				detailMsg = true
				break
			}
		}

		if detailMsg {
			res := regexp.MustCompile(`(?U)([\w_.\->]+)='(.*)'`).FindAllStringSubmatch(customMsg, -1)
			for _, item := range res {
				customMsgMap[item[1]] = item[2]
			}
		} else {
			customMsgMap["default"] = customMsg
		}

		msgName := strings.Join(f.CurRule.Level, "") + f.CurRule.Name
		msg, ok := customMsgMap[msgName]
		if ok {
			return msg, true
		}

		defMsg, ok := customMsgMap["default"]
		if ok {
			return defMsg, true
		}
	}

	return "", false
}

func (f *BaseFiled) PopTrack() {
	if len(f.Track) > 0 {
		f.Track = f.Track[:len(f.Track)-1]
		f.CurDataIdx = strings.Join(f.Track, ".")
	}
}

func (f *BaseFiled) PushTrack(elem string) {
	f.Track = append(f.Track, elem)
	f.CurDataIdx = strings.Join(f.Track, ".")
}
