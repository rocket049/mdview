package main

import (
	"testing"
)

func TestGetSrcs(t *testing.T) {
	body := `<body>
	<img src="a.jpg" />
	<img SRC='ab.jpg' />
	<img Src="好人.jpg">
	<img src="data:;abcd">
	</body>`
	srcs := getSrcs(body)
	if len(srcs) != 3 {
		t.Fail()
	}

	for _, v := range srcs {
		body = replaceSrc(body, v, "data:xxx")
	}
	t.Log(body)
}
