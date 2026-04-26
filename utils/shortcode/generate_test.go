package shortcode

import (
	"os"
	"testing"
)

func TestSuccessGenerateShortURL(t *testing.T) {
	os.Setenv("CODE_LENGTH", "6")
	got := GenerateShortURL()
	expected := 6

	if len(got) != expected {
		t.Errorf("have got: %d and expected: %d", len(got), expected)
	}
}

func Test_Empty_GenerateShortURL(t *testing.T) {
	os.Setenv("CODE_LENGTH", "")
	got := GenerateShortURL()
	expected := ""

	if got != expected {
		t.Errorf("have got: %s and expected: %s", got, expected)
	}
}

func BenchmarkGenerateShortURL(b *testing.B) {
	os.Setenv("CODE_LENGTH", "6")
	for b.Loop() {
		_ = GenerateShortURL()
	}
}
