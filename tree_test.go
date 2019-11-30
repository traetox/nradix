// Copyright (C) 2015 Alex Sergeyev
// This project is licensed under the terms of the MIT license.
// Read LICENSE file for information for all notices and permissions.

package nradix

import (
	"net"
	"testing"
)

var (
	testValA = datum{"A"}
	testValB = datum{"A", "B"}
	testValC = datum{"A", "B", "C"}
	testValD = datum{"A", "B", "D"}
)

func TestTree(t *testing.T) {
	tr := NewTree(0)
	if tr == nil || tr.root == nil {
		t.Error("Did not create tree properly")
	}
	err := tr.AddCIDR("1.2.3.0/25", testValA)
	if err != nil {
		t.Error(err)
	}

	// Matching defined cidr
	inf, err := tr.FindCIDR("1.2.3.1/25")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	// Inside defined cidr
	inf, err = tr.FindCIDR("1.2.3.60/32")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}
	inf, err = tr.FindCIDR("1.2.3.60")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	// Outside defined cidr
	inf, err = tr.FindCIDR("1.2.3.160/32")
	if err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}
	inf, err = tr.FindCIDR("1.2.3.160")
	if err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

	inf, err = tr.FindCIDR("1.2.3.128/25")
	if err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

	// Covering not defined
	inf, err = tr.FindCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

	// Covering defined
	err = tr.AddCIDR("1.2.3.0/24", testValB)
	if err != nil {
		t.Error(err)
	}
	inf, err = tr.FindCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	} else if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}

	inf, err = tr.FindCIDR("1.2.3.160/32")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}

	// Hit both covering and internal, should choose most specific
	inf, err = tr.FindCIDR("1.2.3.0/32")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	// Delete internal
	err = tr.DeleteCIDR("1.2.3.0/25")
	if err != nil {
		t.Error(err)
	}

	// Hit covering with old IP
	inf, err = tr.FindCIDR("1.2.3.0/32")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}

	// Add internal back in
	err = tr.AddCIDR("1.2.3.0/25", testValA)
	if err != nil {
		t.Error(err)
	}

	// Delete covering
	err = tr.DeleteCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	}

	// Hit with old IP
	inf, err = tr.FindCIDR("1.2.3.0/32")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	// Find covering again
	inf, err = tr.FindCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	}
	if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

	// Add covering back in
	err = tr.AddCIDR("1.2.3.0/24", testValB)
	if err != nil {
		t.Error(err)
	}
	inf, err = tr.FindCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}

	// Delete the whole range
	err = tr.DeleteWholeRangeCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	}
	// should be no value for covering
	inf, err = tr.FindCIDR("1.2.3.0/24")
	if err != nil {
		t.Error(err)
	}
	if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}
	// should be no value for internal
	inf, err = tr.FindCIDR("1.2.3.0/32")
	if err != nil {
		t.Error(err)
	}
	if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}
}

func TestSet(t *testing.T) {
	tr := NewTree(0)
	if tr == nil || tr.root == nil {
		t.Error("Did not create tree properly")
	}

	tr.AddCIDR("1.1.1.0/24", testValA)
	inf, err := tr.FindCIDR("1.1.1.0")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	tr.AddCIDR("1.1.1.0/25", testValB)
	inf, err = tr.FindCIDR("1.1.1.0")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}
	inf, err = tr.FindCIDR("1.1.1.0/24")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValA) {
		t.Errorf("Wrong value, expected %v, got %v", testValA, inf)
	}

	// add covering should fail
	err = tr.AddCIDR("1.1.1.0/24", testValC)
	if err != ErrNodeBusy {
		t.Errorf("Should have gotten ErrNodeBusy, instead got err: %v", err)
	}

	// set covering
	err = tr.SetCIDR("1.1.1.0/24", testValC)
	if err != nil {
		t.Error(err)
	}
	inf, err = tr.FindCIDR("1.1.1.0")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValB) {
		t.Errorf("Wrong value, expected %v, got %v", testValB, inf)
	}
	inf, err = tr.FindCIDR("1.1.1.0/24")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValC) {
		t.Errorf("Wrong value, expected 3, got %v", inf)
	}

	// set internal
	err = tr.SetCIDR("1.1.1.0/25", testValD)
	if err != nil {
		t.Error(err)
	}
	inf, err = tr.FindCIDR("1.1.1.0")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValD) {
		t.Errorf("Wrong value, expected %v, got %v", testValD, inf)
	}
	inf, err = tr.FindCIDR("1.1.1.0/24")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValC) {
		t.Errorf("Wrong value, expected %v, got %v", testValC, inf)
	}
}

func TestRegression(t *testing.T) {
	tr := NewTree(0)
	if tr == nil || tr.root == nil {
		t.Error("Did not create tree properly")
	}

	tr.AddCIDR("1.1.1.0/24", testValA)

	tr.DeleteCIDR("1.1.1.0/24")
	tr.AddCIDR("1.1.1.0/25", testValB)

	// inside old range, outside new range
	inf, err := tr.FindCIDR("1.1.1.128")
	if err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}
	if inf, err = tr.FindIP(net.ParseIP("1.1.1.128")); err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}
	if inf, err = tr.FindIP(net.ParseIP("1.1.1.12")); err != nil {
		t.Error(err)
	} else if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

}

func TestTree6(t *testing.T) {
	tr := NewTree(0)
	if tr == nil || tr.root == nil {
		t.Error("Did not create tree properly")
	}
	err := tr.AddCIDR("dead::0/16", testValC)
	if err != nil {
		t.Error(err)
	}

	// Matching defined cidr
	inf, err := tr.FindCIDR("dead::beef")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValC) {
		t.Errorf("Wrong value, expected 3, got %v", inf)
	}

	// Outside
	inf, err = tr.FindCIDR("deed::beef/32")
	if err != nil {
		t.Error(err)
	}
	if inf != nil {
		t.Errorf("Wrong value, expected nil, got %v", inf)
	}

	// Subnet
	err = tr.AddCIDR("dead:beef::0/48", testValD)
	if err != nil {
		t.Error(err)
	}

	// Match defined subnet
	inf, err = tr.FindCIDR("dead:beef::0a5c:0/64")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValD) {
		t.Errorf("Wrong value, expected %v, got %v", testValD, inf)
	}

	// Match outside defined subnet
	inf, err = tr.FindCIDR("dead:0::beef:0a5c:0/64")
	if err != nil {
		t.Error(err)
	}
	if !inf.equal(testValC) {
		t.Errorf("Wrong value, expected %v, got %v", testValC, inf)
	}

}

func TestRegression6(t *testing.T) {
	tr := NewTree(0)
	if tr == nil || tr.root == nil {
		t.Error("Did not create tree properly")
	}
	// in one of the implementations /128 addresses were causing panic...
	tv := datum{"54321"}
	tr.AddCIDR("2620:10f::/32", tv)
	tr.AddCIDR("2620:10f:d000:100::5/128", tv)

	inf, err := tr.FindCIDR("2620:10f:d000:100::5/128")
	if err != nil {
		t.Errorf("Could not get /128 address from the tree, error: %s", err)
	} else if !inf.equal(tv) {
		t.Errorf("Wrong value from /128 test, got %v, expected %v", inf, tv)
	}
	inf, err = tr.FindIP(net.ParseIP("2620:10f:d000:100::5"))
	if err != nil {
		t.Errorf("Could not get ipv6 address from the tree, error: %s", err)
	} else if !inf.equal(tv) {
		t.Errorf("Wrong value from ipv6 test, got %v, expected %v", inf, tv)
	}
	inf, err = tr.FindIP(net.ParseIP("2620:10f:d000:100::2"))
	if err != nil {
		t.Errorf("Could not get ipv6 address from the tree, error: %s", err)
	} else if !inf.equal(tv) {
		t.Errorf("Wrong value from ipv6 test, got %v, expected %v", inf, tv)
	}

	inf, err = tr.FindIP(net.ParseIP("2620:10f:d000:100::13:44"))
	if err != nil {
		t.Errorf("Could not get ipv6 address from the tree, error: %s", err)
	} else if !inf.equal(tv) {
		t.Errorf("Wrong value from ipv6 test, got %v, expected %v", inf, tv)
	}
}
