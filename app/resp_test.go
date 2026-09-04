package main

import (
	"reflect"
	"testing"
)

func TestRESPDecoder(t *testing.T) {
	// Test Case 1: Simple PING command
	input1 := []byte("*1\r\n$4\r\nPING\r\n")
	got1, err := RESPParser(input1)
	want1 := []string{"PING"}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)

	}
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("got %v, want %v", got1, want1)

	}

	// Test Case 2: ECHO with argument
	input2 := []byte("*2\r\n$4\r\nECHO\r\n$5\r\nkhang\r\n")
	got2, err := RESPParser(input2)
	want2 := []string{"ECHO", "khang"}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)

	}
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("got %v, want %v", got2, want2)

	}

}
