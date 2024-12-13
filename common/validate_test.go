package common

import "testing"

func TestValidateURL(t *testing.T) {
	failTests := []struct {
		input string
	}{
		{"//localhost:26657"},
		{"localhost:26657"},
		{"ws://localhost:26657/websocket"},
		{"wss://localhost:26657/websocket"},
	}

	for _, test := range failTests {
		err := ValidateURL(test.input)
		if err == nil {
			t.Errorf("For input '%s', expected error, but got nil", test.input)
		}
	}

	successTests := []struct {
		input string
	}{
		{"http://localhost:26657"},
		{"https://localhost:26657"},
		{"https://localhost:26657/abc"},
	}

	for _, test := range successTests {
		err := ValidateURL(test.input)
		if err != nil {
			t.Errorf("For input '%s', expected no error, but got '%v'", test.input, err)
		}
	}
}

func TestValidateWSURL(t *testing.T) {
	failTests := []struct {
		input string
	}{
		{"http://localhost:26657"},
		{"https://localhost:26657"},
	}

	for _, test := range failTests {
		err := ValidateWSURL(test.input)
		if err == nil {
			t.Errorf("For input '%s', expected error, but got nil", test.input)
		}
	}

	successTests := []struct {
		input string
	}{
		{"ws://localhost:26657/websocket"},
		{"wss://localhost:26657/websocket"},
	}

	for _, test := range successTests {
		err := ValidateWSURL(test.input)
		if err != nil {
			t.Errorf("For input '%s', expected no error, but got '%v'", test.input, err)
		}
	}
}
