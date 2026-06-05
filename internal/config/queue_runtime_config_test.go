package config

import "testing"

func TestApplyRuntimeDefaultsControlsEmbeddedRunnerByServerMode(t *testing.T) {
	dev := &AppConfig{Server: ServerConfig{Mode: "debug"}}
	applyRuntimeDefaults(dev)
	if !dev.Queue.IsEmbeddedRunnerEnabled() {
		t.Fatalf("expected embedded runner enabled in non-release mode")
	}
	if !dev.Worker.IsEnabled() {
		t.Fatalf("expected standalone worker enabled by default")
	}

	prod := &AppConfig{Server: ServerConfig{Mode: "release"}}
	applyRuntimeDefaults(prod)
	if prod.Queue.IsEmbeddedRunnerEnabled() {
		t.Fatalf("expected embedded runner disabled in release mode")
	}
	if !prod.Worker.IsEnabled() {
		t.Fatalf("expected standalone worker enabled in release mode")
	}
}

func TestApplyRuntimeDefaultsKeepsExplicitRunnerFlags(t *testing.T) {
	embedded := true
	worker := false
	cfg := &AppConfig{
		Server: ServerConfig{Mode: "release"},
		Queue:  QueueConfig{EmbeddedRunnerEnabled: &embedded},
		Worker: WorkerConfig{Enabled: &worker},
	}

	applyRuntimeDefaults(cfg)
	if !cfg.Queue.IsEmbeddedRunnerEnabled() {
		t.Fatalf("expected explicit embedded runner flag to be kept")
	}
	if cfg.Worker.IsEnabled() {
		t.Fatalf("expected explicit standalone worker flag to be kept")
	}
}
