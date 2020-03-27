package example_test

import (
	"strings"
	"testing"

	"github.com/searis/subtest"
)

func TestHello(t *testing.T) {
	t.Run(`Given A="hello",B="world"`, func(t *testing.T) {
		a := "hello"
		b := "world"
		t.Run(`When calling C=strings.Join([A,B],"<space>")`, func(t *testing.T) {
			c := strings.Join([]string{a, b}, " ")
			t.Run(`Then C should be "hello<space>world"`,
				subtest.Value(c).DeepEqual("hello world"),
			)
		})
	})
}
