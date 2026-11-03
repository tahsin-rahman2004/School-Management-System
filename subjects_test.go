package main

import "testing"

func TestSubjectsByClassAndGroup(t *testing.T) {
	for _, c := range []struct {
		class, group string
		count        int
	}{{"1", "", 3}, {"2", "", 3}, {"3", "", 6}, {"5", "", 6}, {"6", "", 10}, {"8", "", 10}, {"9", "Science", 13}, {"10", "Humanities", 13}, {"9", "Business Studies", 12}, {"10", "", 9}, {"invalid", "", 0}} {
		if got := len(subjectsFor(User{Class: c.class, Group: c.group})); got != c.count {
			t.Errorf("class %s group %s: got %d subjects", c.class, c.group, got)
		}
	}
	for _, s := range subjectsFor(User{Class: "9", Group: "Business Studies"}) {
		if s.Name == "Physics" {
			t.Fatal("science subject leaked into business group")
		}
	}
}
