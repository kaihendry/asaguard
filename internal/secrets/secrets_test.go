package secrets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaihendry/asaguard/internal/result"
)

func inDir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(orig) })
}

func TestCheckMissingHook(t *testing.T) {
	dir := t.TempDir()
	inDir(t, dir)

	findings := Check()
	if !result.HasFail(findings) {
		t.Error("expected FAIL when pre-commit hook is absent")
	}
}

func TestCheckHookNotExecutable(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0700)
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\ngitleaks protect\n"), 0600) // not executable
	inDir(t, dir)

	findings := Check()
	if !result.HasFail(findings) {
		t.Error("expected FAIL when hook is not executable")
	}
}

func TestCheckHookWithScanner(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0700)
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\ngitleaks protect --staged\n"), 0755)
	inDir(t, dir)

	findings := Check()
	if result.HasFail(findings) {
		t.Errorf("expected PASS with gitleaks in hook, got %v", findings)
	}
}

func TestCheckHookWithoutScanner(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0700)
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho hello\n"), 0755)
	inDir(t, dir)

	findings := Check()
	if !result.HasFail(findings) {
		t.Error("expected FAIL when hook contains no secret scanner")
	}
}

func TestCheckPreCommitConfig(t *testing.T) {
	dir := t.TempDir()
	// No .git/hooks/pre-commit, but .pre-commit-config.yaml with gitleaks
	os.WriteFile(filepath.Join(dir, ".pre-commit-config.yaml"), []byte(`
repos:
  - repo: https://github.com/gitleaks/gitleaks
    hooks:
      - id: gitleaks
`), 0600)
	inDir(t, dir)

	findings := Check()
	// Should pass via config detection even without the hook file
	hasPass := false
	for _, f := range findings {
		if f.Level == result.Pass {
			hasPass = true
		}
	}
	if !hasPass {
		t.Errorf("expected PASS from .pre-commit-config.yaml, got %v", findings)
	}
}
