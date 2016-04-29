package vago

import (
	"strings"
	"testing"
)

func TestOpenFail(t *testing.T) {
	_, err := Open("/nonexistent")
	if err == nil {
		t.Fatal("Expected nil")
	}
}

func TestOpenOK(t *testing.T) {
	v, err := Open("")
	if err != nil {
		t.Fatal("Expected non nil")
	}
	v.Close()
}

func TestLog(t *testing.T) {
	v, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	v.Log("", RAW, func(vxid uint32, tag, _type, data string) int {
		if vxid == 0 && tag == "CLI" && _type == "-" && strings.Contains(data, "PONG") {
			t.Log("Got PONG")
			return -1
		}
		return 0
	})
	v.Close()
}

func TestStats(t *testing.T) {
	v, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	items := v.Stats()
	if len(items) == 0 {
		t.Fatal("Expected map with elements")
	}
}

func TestStatFail(t *testing.T) {
	v, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, ok := v.Stat("foo"); ok {
		t.Fatal("Expected false")
	}
}

func TestStatOK(t *testing.T) {
	v, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, ok := v.Stat("MAIN.uptime"); !ok {
		t.Fatal("Expected some value")
	}
}
