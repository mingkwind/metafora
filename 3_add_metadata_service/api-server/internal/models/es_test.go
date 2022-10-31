package models

import (
	"strconv"
	"testing"
	"time"
)

func TestPutMetadata(t *testing.T) {
	for i := 1; i <= 10; i++ {
		metadata := Metadata{
			Name: "test",
			Size: 100,
			Hash: "hash" + strconv.Itoa(i),
		}
		err := PutMetadata(&metadata)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second)
	}
}
func TestGetMetadata(t *testing.T) {
	metadata, err := GetMetadata("test", 0)
	if err != nil {
		t.Error(err)
	}
	t.Log(metadata)
	metadata, err = GetMetadata("test", 5)
	if err != nil {
		t.Error(err)
	}
	t.Log(metadata)
}

func TestGetLatestVersion(t *testing.T) {
	metadata, err := GetLatestVersion("test")
	if err != nil {
		t.Error(err)
	}
	t.Log(metadata)
}

func TestSearchAllVersions(t *testing.T) {
	ch, err := GetAllVersions("test")
	if err != nil {
		t.Error(err)
	}
	for metadata := range ch {
		t.Log(metadata)
	}
}

/*
func TestDelAllMetadata(t *testing.T) {
	err := DelAllMetadata("test")
	if err != nil {
		t.Error(err)
	}
}
*/
