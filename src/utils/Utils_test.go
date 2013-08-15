package utils

import (
    "testing"
)

func TestRemote(t *testing.T) {
	b, err := ReadRemote("http://10.180.76.88/data/server_view/api/data_collector.php?testreadremote", "sh.ecc.com")
	if err != nil {
		t.Errorf("%v", err.Error())
	}
	println(string(b))
}

