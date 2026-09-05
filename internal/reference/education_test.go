package reference

import "testing"

func TestCanonicalSchoolDaLat(t *testing.T) {
	want := "Trường Đại học Đà Lạt"
	for _, input := range []string{"DLU", "đh đà lạt", "Đại học Đà Lạt", "Trường Đại học Đà Lạt"} {
		got, ok := CanonicalSchool(input)
		if !ok || got != want {
			t.Fatalf("CanonicalSchool(%q) = %q, %v; want %q, true", input, got, ok, want)
		}
	}
}

func TestSameSchoolAlias(t *testing.T) {
	if !SameSchool("DLU", "Trường Đại học Đà Lạt") {
		t.Fatal("DLU and Trường Đại học Đà Lạt must be the same school")
	}
}

func TestUnknownSchoolRejected(t *testing.T) {
	if got, ok := CanonicalSchool("Trường Không Tồn Tại 123"); ok || got != "" {
		t.Fatalf("unknown school = %q, %v; want empty, false", got, ok)
	}
}

func TestCanonicalMajor(t *testing.T) {
	if got := CanonicalMajor("công nghệ thông tin"); got != "Công nghệ thông tin" {
		t.Fatalf("major normalized to %q", got)
	}
	if got := CanonicalMajor("Ngành thử nghiệm mới"); got != "Ngành thử nghiệm mới" {
		t.Fatalf("unknown major should remain editable, got %q", got)
	}
}

func TestSchoolCatalogueHasUniqueNamesAndCodes(t *testing.T) {
	names := map[string]bool{}
	codes := map[string]string{}
	for _, school := range VietnamSchools {
		nameKey := normalizeText(school.Name)
		if nameKey == "" {
			t.Fatal("catalogue contains an empty school name")
		}
		if names[nameKey] {
			t.Fatalf("duplicate school name: %q", school.Name)
		}
		names[nameKey] = true
		codeKey := normalizeText(school.Code)
		if codeKey != "" {
			if prev, exists := codes[codeKey]; exists {
				t.Fatalf("duplicate school code %q: %q and %q", school.Code, prev, school.Name)
			}
			codes[codeKey] = school.Name
		}
	}
	if len(VietnamSchools) < 200 {
		t.Fatalf("catalogue unexpectedly small: %d", len(VietnamSchools))
	}
}
