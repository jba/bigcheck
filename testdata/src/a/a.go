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
	_, _ = Small{}, b // want `rhs #1 is 101 bytes`
	c := make(chan Big)
	c <- Big{}  // want `value being sent is 101 bytes`
	_ = <-c     // want `rhs #0 is 101 bytes` `received value is 101 bytes`
	_ = (<-c).a // want `received value is 101 bytes`
	select {
	case <-c: // want `received value is 101 bytes`
	case c <- Big{}: // want `value being sent is 101 bytes`
	}
	for _ = range []Big{} { // just indices, so no copy
	}
	for _, _ = range []Big{} { // want `ranged value is 101 bytes`
	}
	for _, _ = range [2]Big{} { // want `ranged value is 101 bytes`
	}

	for _, _ = range map[Big]int{} { // want `ranged key is 101 bytes`
	}
	for _, _ = range map[int]Big{} { // want `ranged value is 101 bytes`
	}
	for _ = range c { // want `ranged value is 101 bytes`
	}

	if _ = (Big{}); true { // want `rhs #0 is 101 bytes`
	}
	_ = sb[1].a  // OK
	return Big{} // want `return value #0 is 101 bytes`
}
