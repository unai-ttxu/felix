// Copyright (c) 2017-2018 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iptables

import "fmt"

type Action interface {
	ToFragment(features *Features) string
}

type GotoAction struct {
	Target   string
	TypeGoto struct{}
}

func (g GotoAction) ToFragment(features *Features) string {
	return "--goto " + g.Target
}

func (g GotoAction) String() string {
	return "Goto->" + g.Target
}

type JumpAction struct {
	Target   string
	TypeJump struct{}
}

func (g JumpAction) ToFragment(features *Features) string {
	return "--jump " + g.Target
}

func (g JumpAction) String() string {
	return "Jump->" + g.Target
}

type ReturnAction struct {
	TypeReturn struct{}
}

func (r ReturnAction) ToFragment(features *Features) string {
	return "--jump RETURN"
}

func (r ReturnAction) String() string {
	return "Return"
}

type DropAction struct {
	TypeDrop struct{}
}

func (g DropAction) ToFragment(features *Features) string {
	return "--jump DROP"
}

func (g DropAction) String() string {
	return "Drop"
}

type LogAction struct {
	Prefix  string
	TypeLog struct{}
}

func (g LogAction) ToFragment(features *Features) string {
	return fmt.Sprintf(`--jump LOG --log-prefix "%s: " --log-level 5`, g.Prefix)
}

func (g LogAction) String() string {
	return "Log"
}

type AcceptAction struct {
	TypeAccept struct{}
}

func (g AcceptAction) ToFragment(features *Features) string {
	return "--jump ACCEPT"
}

func (g AcceptAction) String() string {
	return "Accept"
}

type DNATAction struct {
	DestAddr string
	DestPort uint16
	TypeDNAT struct{}
}

func (g DNATAction) ToFragment(features *Features) string {
	if g.DestPort == 0 {
		return fmt.Sprintf("--jump DNAT --to-destination %s", g.DestAddr)
	} else {
		return fmt.Sprintf("--jump DNAT --to-destination %s:%d", g.DestAddr, g.DestPort)
	}
}

func (g DNATAction) String() string {
	return fmt.Sprintf("DNAT->%s:%d", g.DestAddr, g.DestPort)
}

type SNATAction struct {
	ToAddr   string
	TypeSNAT struct{}
}

func (g SNATAction) ToFragment(features *Features) string {
	fullyRand := ""
	if features.SNATFullyRandom {
		fullyRand = " --random-fully"
	}
	return fmt.Sprintf("--jump SNAT --to-source %s%s", g.ToAddr, fullyRand)
}

func (g SNATAction) String() string {
	return fmt.Sprintf("SNAT->%s", g.ToAddr)
}

type MasqAction struct {
	ToPorts  string
	TypeMasq struct{}
}

func (g MasqAction) ToFragment(features *Features) string {
	fullyRand := ""
	if features.MASQFullyRandom {
		fullyRand = " --random-fully"
	}
	if g.ToPorts != "" {
		return fmt.Sprintf("--jump MASQUERADE --to-ports %s"+fullyRand, g.ToPorts)
	}
	return "--jump MASQUERADE" + fullyRand
}

func (g MasqAction) String() string {
	return "Masq"
}

type ClearMarkAction struct {
	Mark          uint32
	TypeClearMark struct{}
}

func (c ClearMarkAction) ToFragment(features *Features) string {
	return fmt.Sprintf("--jump MARK --set-mark 0/%#x", c.Mark)
}

func (c ClearMarkAction) String() string {
	return fmt.Sprintf("Clear:%#x", c.Mark)
}

type SetMarkAction struct {
	Mark        uint32
	TypeSetMark struct{}
}

func (c SetMarkAction) ToFragment(features *Features) string {
	return fmt.Sprintf("--jump MARK --set-mark %#x/%#x", c.Mark, c.Mark)
}

func (c SetMarkAction) String() string {
	return fmt.Sprintf("Set:%#x", c.Mark)
}

type SetMaskedMarkAction struct {
	Mark              uint32
	Mask              uint32
	TypeSetMaskedMark struct{}
}

func (c SetMaskedMarkAction) ToFragment(features *Features) string {
	return fmt.Sprintf("--jump MARK --set-mark %#x/%#x", c.Mark, c.Mask)
}

func (c SetMaskedMarkAction) String() string {
	return fmt.Sprintf("Set:%#x", c.Mark)
}

type NoTrackAction struct {
	TypeNoTrack struct{}
}

func (g NoTrackAction) ToFragment(features *Features) string {
	return "--jump NOTRACK"
}

func (g NoTrackAction) String() string {
	return "NOTRACK"
}