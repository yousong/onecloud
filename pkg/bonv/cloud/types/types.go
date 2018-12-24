package types

// TODO rename package name to cloud

import (
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/utils"
)

type Vpc struct {
	Id          string
	Name        string
	Description string

	Provider   string
	RegionId   string
	Status     string
	CidrBlock  string
	VRouterId  string   // TODO aliyun specific
	VSwitchIds []string // TODO remove this
}

func (v *Vpc) Validate() error {
	el := utils.ErrList{}
	if v.Id == "" {
		el = append(el, fmt.Errorf("Vpc Id must not be empty"))
	}
	if v.Provider == "" {
		// TODO check content
		el = append(el, fmt.Errorf("Vpc Provider must not be empty"))
	}
	if v.RegionId == "" {
		el = append(el, fmt.Errorf("Vpc RegionId must not be empty"))
	}
	if v.CidrBlock == "" {
		el = append(el, fmt.Errorf("Vpc CidrBlock must not be empty"))
	}
	return el.ToError()
}
