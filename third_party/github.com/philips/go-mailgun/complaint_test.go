package mailgun

import (
	"testing"
)

func TestComplaint(t *testing.T) {
	n, res, err := c.Complaints(*domain, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total complaints: %d", n)
	for _, r := range res {
		t.Logf("%+v", r)
	}
}
