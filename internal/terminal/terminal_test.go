package terminal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/muesli/reflow/wordwrap"
)

func TestTerminal(t *testing.T) {
	s := wordwrap.String(fmt.Sprintf("What is your problem with this kind of chess play I honestly do not understand."), 70)
	a := strings.Split(s, "\n")
	//fmt.Println(a)
	fmt.Println(a[0])
	fmt.Println()
	fmt.Println(a[1])
	/*for _, c := range s {
		fmt.Println(c)
		fmt.Printf("%U\n", c)
	}
	lines := bufio.NewScanner(strings.NewReader(s))
	var c int
	for lines.Scan() {
		c++
	}
	fmt.Println(c)*/
}
