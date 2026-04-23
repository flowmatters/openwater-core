package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/flowmatters/openwater-core/sim"
)

const (
	VersionMajor  = 1
	VersionMinor  = 0
	VersionPatch  = 2
	VersionBranch = "" // Empty for main branch, e.g., "exp-nutrient" for experimental
)

var (
	// BuildSHA is injected at build time via -ldflags
	BuildSHA string = "dev"

	// BuildTime is injected at build time via -ldflags
	BuildTime string = "unknown"

	// SignatureHash is computed once at first access
	SignatureHash string

	sigHashOnce sync.Once
)

// ModelSignature represents a canonical model signature for hashing
type ModelSignature struct {
	Name       string               `json:"name"`
	Inputs     []string             `json:"inputs"`
	Outputs    []string             `json:"outputs"`
	States     []string             `json:"states"`
	Parameters []ParameterSignature `json:"parameters"`
	Dimensions []string             `json:"dimensions"`
}

// ParameterSignature represents canonical parameter info for signature
type ParameterSignature struct {
	Name       string   `json:"name"`
	Units      string   `json:"units"`
	Dimensions []string `json:"dimensions"`
}

// FullVersion returns the complete version string
func FullVersion() string {
	ensureSignatureHash()

	version := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)

	if BuildSHA != "" && BuildSHA != "dev" {
		version += "+" + BuildSHA
	}

	if VersionBranch != "" {
		version += "-" + VersionBranch
	}

	version += "." + SignatureHash

	return version
}

// ShortVersion returns version without signature hash
func ShortVersion() string {
	version := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)

	if BuildSHA != "" && BuildSHA != "dev" {
		version += "+" + BuildSHA
	}

	if VersionBranch != "" {
		version += "-" + VersionBranch
	}

	return version
}

// GetSignatureHash computes or retrieves the model signature hash
func GetSignatureHash() string {
	ensureSignatureHash()
	return SignatureHash
}

// IsCompatible checks if a given version string is compatible with current version
// based on signature hash comparison
func IsCompatible(otherVersion string) bool {
	ensureSignatureHash()
	otherSig := ExtractSignatureHash(otherVersion)
	return otherSig == SignatureHash
}

// ExtractSignatureHash extracts signature hash from version string
func ExtractSignatureHash(version string) string {
	// Version format: X.Y.Z+BUILD[-BRANCH].SIGHASH
	// Find last dot
	for i := len(version) - 1; i >= 0; i-- {
		if version[i] == '.' {
			if i+1 < len(version) {
				return version[i+1:]
			}
			break
		}
	}
	return ""
}

// ensureSignatureHash computes signature hash once
func ensureSignatureHash() {
	sigHashOnce.Do(func() {
		SignatureHash = computeSignatureHash()
	})
}

// computeSignatureHash generates hash from all model signatures in catalog
func computeSignatureHash() string {
	// Collect all model signatures
	var signatures []ModelSignature

	for modelName, modelFactory := range sim.Catalog {
		model := modelFactory()
		desc := model.Description()

		sig := ModelSignature{
			Name:       modelName,
			Inputs:     sim.VariableNames(desc.Inputs),
			Outputs:    sim.VariableNames(desc.Outputs),
			States:     copyStrings(desc.States),
			Dimensions: copyStrings(desc.Dimensions),
			Parameters: make([]ParameterSignature, len(desc.Parameters)),
		}

		// Extract parameter signatures (name, units, dimensions only)
		for i, param := range desc.Parameters {
			sig.Parameters[i] = ParameterSignature{
				Name:       param.Name,
				Units:      param.Units,
				Dimensions: copyStrings(param.Dimensions),
			}
		}

		signatures = append(signatures, sig)
	}

	// Sort by model name for deterministic ordering
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].Name < signatures[j].Name
	})

	// Generate canonical JSON representation
	jsonBytes, err := json.Marshal(signatures)
	if err != nil {
		// Should never happen with our simple structs
		return "error"
	}

	// Hash and take first 8 characters
	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash[:4]) // 4 bytes = 8 hex chars
}

// copyStrings creates a copy of string slice
func copyStrings(src []string) []string {
	if src == nil {
		return []string{}
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
