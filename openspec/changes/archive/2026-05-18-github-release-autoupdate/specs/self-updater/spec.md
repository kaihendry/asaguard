## ADDED Requirements

### Requirement: Skip update check for dev builds
The self-updater SHALL not perform any network request when the embedded version is `"dev"`.

#### Scenario: Dev build skips check
- **WHEN** `asaguard` runs with version string `"dev"`
- **THEN** no HTTP request is made to GitHub and the tool starts immediately

### Requirement: Check for a newer release on startup
Release binaries SHALL query the GitHub Releases API for `kaihendry/asaguard` at startup and compare the latest tag against the embedded version.

#### Scenario: No newer version available
- **WHEN** the running version equals the latest GitHub Release tag
- **THEN** the tool proceeds normally with no output about updates

#### Scenario: Network unreachable or API error
- **WHEN** the GitHub API is unreachable or returns an error within the 5-second timeout
- **THEN** the update check is silently skipped and the tool proceeds normally

### Requirement: Self-replace and re-exec when a newer version is found
When a newer release is available, the binary SHALL download the appropriate asset for the current OS/arch, replace itself atomically, and re-execute with the original arguments.

#### Scenario: Newer version downloaded and re-executed
- **WHEN** a newer release exists and the download succeeds
- **THEN** the running binary is replaced on disk and the new binary is executed with the same arguments, producing output as if the user had run the updated binary directly

#### Scenario: Download fails
- **WHEN** the asset download fails (network error, checksum mismatch)
- **THEN** the existing binary is left unchanged and the tool continues with the current version
