// Copyright (C) 2015 Alex Sergeyev
// This project is licensed under the terms of the MIT license.
// Read LICENSE file for information for all notices and permissions.

package nradix

import (
	"errors"
	"strings"
)

type node struct {
	left, right, parent *node
	value               interface{}
}

// Tree implements radix tree for working with IP/mask
type Tree struct {
	root   *node
	free   *node
	has128 bool
}

const startbit = uint32(0x80000000)

var (
	ErrNodeBusy = errors.New("Node Busy")
	ErrNotFound = errors.New("No Such Node")
	ErrBadIP    = errors.New("Bad IP address or mask")
)

// NewTree creates Tree and preallocates (if preallocate not zero) number of nodes that would be ready to fill with data.
func NewTree(preallocate int) *Tree {
	tree := new(Tree)
	tree.root = tree.newnode()
	if preallocate == 0 {
		return tree
	}

	// Simplification, static preallocate max 6 bits
	if preallocate > 6 || preallocate < 0 {
		preallocate = 6
	}

	var key, mask uint32

	for inc := startbit; preallocate > 0; inc, preallocate = inc>>1, preallocate-1 {
		key = 0
		mask >>= 1
		mask |= startbit

		for {
			tree.insert32(key, mask, nil)
			key += inc
			if key == 0 { // magic bits collide
				break
			}
		}
	}

	return tree
}

// AddCIDR adds value associated with IP/mask to the tree. Will return error for invalid CIDR or if value already exists.
// Note: Only IPV4 supported so far...
func (tree *Tree) AddCIDR(cidr string, val interface{}) error {
	ip, mask, err := parsecidr4(cidr)
	if err != nil {
		return err
	}
	return tree.insert32(ip, mask, val)
}

// DeleteCIDR removes value associated with IP/mask from the tree.
func (tree *Tree) DeleteCIDR(cidr string) error {
	ip, mask, err := parsecidr4(cidr)
	if err != nil {
		return err
	}
	return tree.delete32(ip, mask)
}

// Find CIDR traverses tree to proper Node and returns previously saved information in longest covered IP.
func (tree *Tree) FindCIDR(cidr string) (interface{}, error) {
	ip, mask, err := parsecidr4(cidr)
	if err != nil {
		return nil, err
	}
	return tree.find32(ip, mask), nil
}

func (tree *Tree) insert32(key, mask uint32, value interface{}) error {
	bit := startbit
	node := tree.root
	next := tree.root
	for bit&mask != 0 {
		if key&bit != 0 {
			next = node.right
		} else {
			next = node.left
		}
		if next == nil {
			break
		}
		bit >>= 1
		node = next
	}
	if next != nil {
		if node.value != nil {
			return ErrNodeBusy
		}
		node.value = value
		return nil
	}
	for bit&mask != 0 {
		next = tree.newnode()
		next.parent = node
		if key&bit != 0 {
			node.right = next
		} else {
			node.left = next
		}
		bit >>= 1
		node = next
	}
	node.value = value

	return nil
}

func (tree *Tree) delete32(key, mask uint32) error {
	bit := startbit
	node := tree.root
	for node != nil && bit&mask != 0 {
		if key&bit != 0 {
			node = node.right
		} else {
			node = node.left
		}
		bit >>= 1
	}
	if node == nil {
		return ErrNotFound
	}

	if node.right != nil && node.left != nil {
		// keep it just trim value
		if node.value != nil {
			node.value = nil
			return nil
		}
		return ErrNotFound
	}

	// need to trim leaf
	for {
		if node.parent.right == node {
			node.parent.right = nil
		} else {
			node.parent.left = nil
		}
		// reserve this node for future use
		node.right = tree.free
		tree.free = node
		// move to parent, check if it's free of value and children
		node = node.parent
		if node.right != nil || node.left != nil || node.value != nil {
			break
		}
		// do not delete root node
		if node.parent == nil {
			break
		}
	}

	return nil
}

func (tree *Tree) find32(key, mask uint32) (value interface{}) {
	bit := startbit
	node := tree.root
	for node != nil {
		if node.value != nil {
			value = node.value
		}
		if key&bit != 0 {
			node = node.right
		} else {
			node = node.left
		}
		if mask&bit == 0 {
			break
		}
		bit >>= 1

	}
	return value
}

func (tree *Tree) newnode() (p *node) {
	if tree.free != nil {
		p = tree.free
		tree.free = tree.free.right
		return p
	}

	// ideally should be aligned in array but for now just let Go decide:
	return new(node)
}

func loadip4(ipstr string) (uint32, error) {
	var (
		ip  uint32
		oct uint32
		b   byte
		num byte
	)

	for _, b = range []byte(ipstr) {
		switch {
		case b == '.':
			num++
			if 0xffffffff-ip < oct {
				return 0, ErrBadIP
			}
			ip = ip<<8 + oct
			oct = 0
		case b >= '0' && b <= '9':
			oct = oct*10 + uint32(b-'0')
			if oct > 255 {
				return 0, ErrBadIP
			}
		default:
			return 0, ErrBadIP
		}
	}
	if num != 3 {
		return 0, ErrBadIP
	}
	if 0xffffffff-ip < oct {
		return 0, ErrBadIP
	}
	return ip<<8 + oct, nil
}

func parsecidr4(cidr string) (uint32, uint32, error) {
	var mask uint32
	p := strings.IndexByte(cidr, '/')
	if p > 0 {
		for _, c := range cidr[p+1:] {
			if c < '0' || c > '9' {
				return 0, 0, ErrBadIP
			}
			mask = mask*10 + uint32(c-'0')
		}
		mask = 0xffffffff << (32 - mask)
		cidr = cidr[:p]
	} else {
		mask = 0xffffffff
	}
	ip, err := loadip4(cidr)
	if err != nil {
		return 0, 0, err
	}
	return ip, mask, nil
}
