package vago

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenFail(t *testing.T) {
	c := Config{
		Path: "/nonexistent",
	}
	_, err := Open(&c)
	if err == nil {
		t.Fatal("Expected non-nil")
	}
}

func TestOpenOK(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Log(err)
		t.Fatal("Expected nil")
	}
	v.Close()
}

func TestOpenFailedTimeout(t *testing.T) {
	c := Config{
		Timeout: 2000,
		Path:    "/nonexistent",
	}
	start := time.Now()
	_, err := Open(&c)
	end := time.Now()
	if err == nil {
		t.Fatal("Expected non-nil")
	}
	if end.Sub(start) < c.Timeout*time.Millisecond {
		t.Fatal("Expected timeout >= c.Timeout")
	}
}

func TestLog(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	defer v.Close()
	err = v.Log("", RAW, COPT_TAIL|COPT_BATCH, func(vxid uint32, tag, _type, data string) int {
		if vxid == 0 && tag == "CLI" && _type == "-" && strings.Contains(data, "PONG") {
			return -1
		}
		return 0
	})
	if err != nil {
		t.Fatal("Expected nil")
	}
}

func TestLogGoroutineClose(t *testing.T) {
	var wg sync.WaitGroup
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	wg.Add(1)
	go func(v *Varnish) {
		defer wg.Done()
		err := v.Log("", RAW, COPT_TAIL|COPT_BATCH, func(vxid uint32, tag, _type, data string) int {
			return -1
		})
		if err != nil {
			t.Fatal("Expected nil")
		}
	}(v)
	time.Sleep(10 * time.Millisecond)
	v.Stop()
	wg.Wait()
	v.Close()
}

func TestInvalidQuery(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	defer v.Close()
	err = v.Log("nonexistent", RAW, COPT_TAIL|COPT_BATCH, func(vxid uint32, tag, _type, data string) int {
		return -1
	})
	if _, ok := err.(ErrVSL); !ok {
		t.Fatal("Expected ErrVSL")
	}
}

func TestStats(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	defer v.Close()
	items := v.Stats()
	if len(items) == 0 {
		t.Fatal("Expected map with elements")
	}
}

func TestStatFail(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	defer v.Close()
	if _, ok := v.Stat("foo"); ok {
		t.Fatal("Expected false")
	}
}

func TestStatOK(t *testing.T) {
	c := Config{}
	v, err := Open(&c)
	if err != nil {
		t.Fatal("Expected nil")
	}
	defer v.Close()
	if _, ok := v.Stat("MAIN.uptime"); !ok {
		t.Fatal("Expected some value")
	}
}
