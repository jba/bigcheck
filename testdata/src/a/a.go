// Assumes the size threshold is 100.

package a

import "fmt"

type Small struct {
	a int
	b string
}

type Big struct {
	a int8
	b [100]byte
}

func foo() Big {
	fmt.Println([99]byte{}, [100]byte{}) // want `arg #1 is 100 bytes`
	fmt.Println(Big{}, Small{}, &Big{})  // want `arg #0 is 101 bytes`
	var b Big
	var sb []Big
	x1, x2 := Small{}, b // want `rhs #1 is 101 bytes`
	c := make(chan Big)
	c <- Big{}    // want `value being sent is 101 bytes`
	_ = <-c       // want `received value is 101 bytes`
	x3 := (<-c).a // want `received value is 101 bytes`
	_, _, _ = x1, x2, x3
	select {
	case <-c: // want `received value is 101 bytes`
	case c <- Big{}: // want `value being sent is 101 bytes`
	}
	for _ = range []Big{} { // just indices, so no copy
	}
	for _, _ = range []Big{} { // never a copy with _
	}
	for _, x := range []Big{} { // want `ranged value is 101 bytes`
		_ = x
	}
	for _, x := range [2]Big{} { // want `ranged value is 101 bytes`
		_ = x
	}
	for k, v := range map[Big]int{} { // want `ranged key is 101 bytes`
		_, _ = k, v
	}
	for k, v := range map[int]Big{} { // want `ranged value is 101 bytes`
		_, _ = k, v
	}
	for x := range c { // want `ranged value is 101 bytes`
		_ = x
	}

	if x := (Big{}); true { // want `rhs #0 is 101 bytes`
		_ = x
	}
	if _ = (Big{}); true {
	}
	_ = sb[1].a  // OK
	return Big{} // want `return value #0 is 101 bytes`
}
