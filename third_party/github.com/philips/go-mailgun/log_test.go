package mailgun

import (
	"testing"
)

func TestLog(t *testing.T) {
	n, res, err := c.Logs(*domain, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total logs: %d", n)
	for _, r := range res {
		t.Logf("%+v", r)
	}
}
