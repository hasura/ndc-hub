package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStagingStorageAuthentication(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "configured-in-workflow")
	v, err := storageOptions("staging", "/verified-bundle", stagingRegistryBucket, "")
	if err != nil || len(v) != 0 {
		t.Fatalf("expected ADC: %v", err)
	}
	for _, testcase := range []struct{ env, bucket, static string }{
		{"production", stagingRegistryBucket, ""},
		{"staging", "other-bucket", ""},
		{"staging", stagingRegistryBucket, "legacy credential"},
	} {
		if _, err := storageOptions(testcase.env, "/verified-bundle", testcase.bucket, testcase.static); err == nil {
			t.Fatal("accepted unexpected authentication mode")
		}
	}
	if v, err := storageOptions("production", "", "legacy-bucket", "legacy credential"); err != nil || len(v) != 1 {
		t.Fatalf("production no longer selects its configured JSON credential: %v", err)
	}
	if _, err := storageOptions("production", "", "legacy-bucket", ""); err == nil {
		t.Fatal("legacy production mode must not fall back to ADC")
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	if _, err := storageOptions("staging", "/verified-bundle", stagingRegistryBucket, ""); err == nil {
		t.Fatal("verified staging mode must require the workflow's ADC configuration")
	}
}

func TestStagingPackageReadsOnlyBundle(t *testing.T) {
	original := stagingInputDir
	t.Cleanup(func() { stagingInputDir = original })
	stagingInputDir = t.TempDir()
	path := "registry/hasura/example/releases/v1.2.3/connector-packaging.json"
	pkgPath := filepath.Join(stagingInputDir, "packages/hasura/example/releases/v1.2.3")
	if err := os.MkdirAll(pkgPath, 0700); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(pkgPath, "connector-metadata.yaml")
	archive := filepath.Join(pkgPath, "package.tgz")
	if err := os.WriteFile(metadata, []byte("version: v1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, []byte("already validated bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	m, name, err := readStagingPackage(path)
	if err != nil || name != archive || m["version"] != "v1" {
		t.Fatalf("expected offline package and metadata: %v", err)
	}
	if _, err := readFile("registry/hasura/example/../../../test.txt"); err == nil {
		t.Fatal("accepted non-canonical registry-relative path")
	}
	if err := os.Remove(metadata); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(archive, metadata); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readStagingPackage(path); err == nil {
		t.Fatal("accepted symlink as package metadata")
	}
}
