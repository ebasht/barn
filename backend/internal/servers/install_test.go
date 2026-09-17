package servers

import (
	"strings"
	"testing"
)

func TestBuildAgentUninstallScriptRemovesCurrentAndLegacyAgent(t *testing.T) {
	script := buildAgentUninstallScript()
	for _, want := range []string{
		"disable --now barn-agent.service",
		"disable --now dockpilot-agent.service",
		"/opt/barn-agent",
		"/etc/barn-agent",
		"/opt/dockpilot-agent",
		"/etc/dockpilot-agent",
		"systemctl daemon-reload",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("uninstall script does not contain %q", want)
		}
	}
}

func TestStrconvQuoteEscapes(t *testing.T) {
	q := strconvQuote("a'b\"c")
	if !strings.HasPrefix(q, `"`) || !strings.HasSuffix(q, `"`) {
		t.Fatalf("unexpected quote: %s", q)
	}
}
