// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.
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

import (
	"bytes"

	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	rules1 = []Rule{
		{Match: MatchCriteria{"-m foobar --foobar baz"}, Action: JumpAction{Target: "biff"}},
	}
	rules2 = []Rule{
		{Match: MatchCriteria{"-m foobar --foobar baz"}, Action: JumpAction{Target: "boff"}},
	}
	rules3 = []Rule{
		{Match: MatchCriteria{"-m foobar --foobar baz"}, Action: JumpAction{Target: "biff"}},
		{Match: MatchCriteria{"-m foobar --foobar baz"}, Action: JumpAction{Target: "boff"}},
	}
)

var _ = Describe("Rule hashing tests", func() {
	It("should generate different hashes for different rules", func() {
		hashes1 := calculateHashes("chain", rules1)
		hashes2 := calculateHashes("chain", rules2)
		Expect(hashes1).NotTo(Equal(hashes2))
	})
	It("should generate same hash for prefix to same chain", func() {
		hashes1 := calculateHashes("chain", rules1)
		hashes2 := calculateHashes("chain", rules3)
		Expect(hashes1[0]).To(Equal(hashes2[0]))
	})
	It("should generate different hashes for different chains with same rules", func() {
		hashes1 := calculateHashes("chain", rules1)
		hashes2 := calculateHashes("chain2", rules1)
		Expect(hashes1).NotTo(Equal(hashes2))
	})
	It("should generate different hashes for same rule at different position", func() {
		hashes2 := calculateHashes("chain", rules2)
		hashes3 := calculateHashes("chain", rules3)
		Expect(hashes2[0]).NotTo(Equal(hashes3[1]))
	})
	It("should generate a slice of same length as input", func() {
		Expect(len(calculateHashes("foo", rules1))).To(Equal(len(rules1)))
		Expect(len(calculateHashes("foo", rules2))).To(Equal(len(rules2)))
		Expect(len(calculateHashes("foo", rules3))).To(Equal(len(rules3)))
	})
})

var _ = Describe("Hash extraction tests", func() {
	var table *Table

	BeforeEach(func() {
		table = NewTable(
			"filter",
			4,
			"cali:",
			&sync.Mutex{},
			TableOptions{
				HistoricChainPrefixes:    []string{"felix-", "cali"},
				ExtraCleanupRegexPattern: "an-old-rule",
				BackendMode:              "legacy",
				LookPathOverride: func(file string) (s string, e error) {
					return s, nil
			},
		)
	})

	It("should extract an old felix rule by prefix", func() {
		hashes, err := table.readHashesFrom(newClosableBuf("-A FORWARD -j felix-FORWARD\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(hashes).To(Equal(map[string][]string{
			"FORWARD": []string{"OLD INSERT RULE"},
		}))
	})
	It("should extract an old felix rule by special case", func() {
		hashes, err := table.readHashesFrom(newClosableBuf(
			"-A FORWARD -j an-old-rule\n" +
				"-A FORWARD -j ignore-me\n",
		))
		Expect(err).NotTo(HaveOccurred())
		Expect(hashes).To(Equal(map[string][]string{
			"FORWARD": []string{
				"OLD INSERT RULE",
				"",
			},
		}))
	})
	It("should extract a hash", func() {
		hashes, err := table.readHashesFrom(newClosableBuf(
			"-A FORWARD -m comment --comment \"cali:wUHhoiAYhphO9Mso\" -j cali-FORWARD\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(hashes).To(Equal(map[string][]string{
			"FORWARD": []string{"wUHhoiAYhphO9Mso"},
		}))
	})
	It("should extract a hash or a gap from each rule", func() {
		hashes, err := table.readHashesFrom(newClosableBuf(
			"-A FORWARD -m comment --comment \"cali:wUHhoiAYhphO9Mso\" -j cali-FORWARD\n" +
				"-A FORWARD -m comment --comment \"cali:abcdefghij1234-_\" -j cali-FORWARD\n" +
				"-A FORWARD --src '1.2.3.4'\n" +
				"-A FORWARD -m comment --comment \"cali:1234567890093213\" -j cali-FORWARD\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(hashes).To(Equal(map[string][]string{
			"FORWARD": []string{
				"wUHhoiAYhphO9Mso",
				"abcdefghij1234-_",
				"",
				"1234567890093213",
			},
		}))
	})
	It("should handle multiple chains", func() {
		hashes, err := table.readHashesFrom(newClosableBuf(
			"-A cali-abcd -m comment --comment \"cali:wUHhoiAYhphO9Mso\" -j cali-FORWARD\n" +
				"-A cali-abcd -m comment --comment \"cali:abcdefghij1234-_\" -j cali-FORWARD\n" +
				"-A FORWARD --src '1.2.3.4'\n" +
				"-A FORWARD -m comment --comment \"cali:1234567890093213\" -j cali-FORWARD\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(hashes).To(Equal(map[string][]string{
			"cali-abcd": []string{
				"wUHhoiAYhphO9Mso",
				"abcdefghij1234-_",
			},
			"FORWARD": []string{
				"",
				"1234567890093213",
			},
		}))
	})
})

func newClosableBuf(s string) *withDummyClose {
	return (*withDummyClose)(bytes.NewBufferString(s))
}

type withDummyClose bytes.Buffer

func (b *withDummyClose) Read(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Read(p)
}

func (b *withDummyClose) Close() error {
	return nil
}

func calculateHashes(chainName string, rules []Rule) []string {
	chain := &Chain{
		Name:  chainName,
		Rules: rules,
	}
	return chain.RuleHashes()
}
