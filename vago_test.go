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
	defer v.Close()
	if err != nil {
		t.Fatal("Expected non nil")
	}
}

func TestLog(t *testing.T) {
	v, err := Open("")
	defer v.Close()
	if err != nil {
		t.Fatal(err)
	}
	v.Log("", RAW, func(vxid uint32, tag, _type, data string) int {
		if vxid == 0 && tag == "CLI" && _type == "-" && strings.Contains(data, "PONG") {
			return -1
		}
		return 0
	})
}

func TestStats(t *testing.T) {
	v, err := Open("")
	defer v.Close()
	if err != nil {
		t.Fatal(err)
	}
	items := v.Stats()
	if len(items) == 0 {
		t.Fatal("Expected map with elements")
	}
}

func TestStat(t *testing.T) {
	v, err := Open("")
	defer v.Close()
	if err != nil {
		t.Fatal(err)
	}
	uptime := v.Stat("MAIN.uptime")
	if uptime < 0 {
		t.Fatal("Expected value > 0")
	}
}
