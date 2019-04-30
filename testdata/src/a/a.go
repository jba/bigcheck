// Assumes the size threshold is 100.

package a

import "fmt"

type Small struct {
	a int
	b string
}

type Big struct {
	a int
	b [98]byte
}

func foo() Big {
	fmt.Println([99]byte{}, [100]byte{}) // want `arg #1 is large`
	fmt.Println(Big{}, Small{}, &Big{})  // want `arg #0 is large`
	var b Big
	var sb []Big
	_, _ = Small{}, b // want `rhs #1 is large`
	c := make(chan Big)
	c <- Big{}  // want `value being sent is large`
	_ = <-c     // want `rhs is large` `received value is large`
	_ = (<-c).a // want `received value is large`
	select {
	case <-c: // want `received value is large`
	case c <- Big{}: // want `value being sent is large`
	}
	for _ = range []Big{} { // just indices, so no copy
	}
	for _, _ = range []Big{} { // want `ranged value is large`
	}
	for _, _ = range [2]Big{} { // want `ranged value is large`
	}

	for _, _ = range map[Big]int{} { // want `ranged key is large`
	}
	for _, _ = range map[int]Big{} { // want `ranged value is large`
	}
	for _ = range c { // want `ranged value is large`
	}

	if _ = (Big{}); true { // want `rhs is large`
	}
	_ = sb[1].a  // OK
	return Big{} // want `return value #0 is large`
}
