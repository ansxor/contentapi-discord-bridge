package bot

import "testing"

func TestFilterMentions(t *testing.T) {
	if FilterMentions("@test") != "@test" {
		t.Error("filterMentions failed")
	}

	if FilterMentions("@everyone") != "@\u200Beveryone" {
		t.Error("filterMentions failed")
	}

	if FilterMentions("@here") != "@\u200Bhere" {
		t.Error("filterMentions failed")
	}

	if FilterMentions("This is my message containing @everyone.") != "This is my message containing @\u200Beveryone." {
		t.Error("filterMentions failed")
	}
}
