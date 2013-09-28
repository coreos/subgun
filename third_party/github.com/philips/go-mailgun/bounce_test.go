package mailgun

import (
	"testing"
)

func TestBounce(t *testing.T) {
	n, res, err := c.Bounces(*domain, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total bounces: %d", n)
	for _, r := range res {
		t.Logf("%+v", r)
	}
}
