package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// These fixed observations are the independently authored oracle for the v2
// synthetic fixtures. Their expected values come from
// docs/fixtures/remediation/financial-table-experiment.v1.json,
// docs/fixtures/remediation/financial-table-experiment.v2.json, and the
// historical association record in
// docs/research/pdfcpu-financial-table-v2-repair-2026-10-13.md, rather than
// from tableObservations or a mapping/checker result. They characterize the
// named fixture pairs only; they do not measure inference accuracy or
// authorize a repair.
func TestMappingEvidenceV2FixtureObservations(t *testing.T) {
	tests := []struct {
		name         string
		positivePath string
		positiveSHA  string
		damagedPath  string
		damagedSHA   string
		readingOrder []string
		headers      []headerObservation
		positiveRows [][]cellObservation
		damagedRows  [][]cellObservation
	}{
		{
			name:         "statement",
			positivePath: "../../docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
			positiveSHA:  "4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec",
			damagedPath:  "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
			damagedSHA:   "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836",
			readingOrder: []string{"Date", "Description", "Amount", "2026-09-03", "North Harbor", "-$18.25", "2026-09-04", "City Transit", "$42.10"},
			headers: []headerObservation{
				{title: "Date", id: "dateHeader", scope: "Column", index: 0},
				{title: "Description", id: "descriptionHeader", scope: "Column", index: 1},
				{title: "Amount", id: "amountHeader", scope: "Column", index: 2},
			},
			positiveRows: [][]cellObservation{
				{{title: "2026-09-03", role: "TD", headerIDs: []string{"dateHeader"}, index: 0}, {title: "North Harbor", role: "TD", headerIDs: []string{"descriptionHeader"}, index: 1}, {title: "-$18.25", role: "TD", headerIDs: []string{"amountHeader"}, index: 2}},
				{{title: "2026-09-04", role: "TD", headerIDs: []string{"dateHeader"}, index: 0}, {title: "City Transit", role: "TD", headerIDs: []string{"descriptionHeader"}, index: 1}, {title: "$42.10", role: "TD", headerIDs: []string{"amountHeader"}, index: 2}},
			},
			damagedRows: [][]cellObservation{
				{{title: "2026-09-03", role: "TD", headerIDs: []string{"dateHeader"}, index: 0}, {title: "North Harbor", role: "TD", headerIDs: []string{"descriptionHeader"}, index: 1}, {title: "-$18.25", role: "TD", headerIDs: []string{"amountHeader"}, index: 2}},
				{{title: "2026-09-04", role: "TD", headerIDs: []string{"dateHeader"}, index: 0}, {title: "City Transit", role: "TD", headerIDs: []string{"descriptionHeader"}, index: 1}, {title: "$42.10", role: "TD", headerIDs: []string{"descriptionHeader"}, index: 2}},
			},
		},
		{
			name:         "invoice",
			positivePath: "../../docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
			positiveSHA:  "ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72",
			damagedPath:  "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
			damagedSHA:   "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad",
			readingOrder: []string{"Item", "Quantity", "Unit amount", "Paper", "2", "$12.50", "Envelopes", "1", "$8.00"},
			headers: []headerObservation{
				{title: "Item", id: "itemHeader", scope: "Column", index: 0},
				{title: "Quantity", id: "quantityHeader", scope: "Column", index: 1},
				{title: "Unit amount", id: "unitAmountHeader", scope: "Column", index: 2},
			},
			positiveRows: [][]cellObservation{
				{{title: "Paper", role: "TD", headerIDs: []string{"itemHeader"}, index: 0}, {title: "2", role: "TD", headerIDs: []string{"quantityHeader"}, index: 1}, {title: "$12.50", role: "TD", headerIDs: []string{"unitAmountHeader"}, index: 2}},
				{{title: "Envelopes", role: "TD", headerIDs: []string{"itemHeader"}, index: 0}, {title: "1", role: "TD", headerIDs: []string{"quantityHeader"}, index: 1}, {title: "$8.00", role: "TD", headerIDs: []string{"unitAmountHeader"}, index: 2}},
			},
			damagedRows: [][]cellObservation{
				{{title: "Paper", role: "TD", headerIDs: []string{"itemHeader"}, index: 0}, {title: "2", role: "TD", headerIDs: []string{"quantityHeader"}, index: 1}, {title: "$12.50", role: "TD", headerIDs: []string{"unitAmountHeader"}, index: 2}},
				{{title: "Envelopes", role: "TD", headerIDs: []string{"itemHeader"}, index: 0}, {title: "1", role: "TD", headerIDs: []string{"quantityHeader"}, index: 1}, {title: "$8.00", role: "TD", headerIDs: []string{"quantityHeader"}, index: 2}},
			},
		},
	}

	for _, fixture := range tests {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			for _, source := range []struct {
				name     string
				path     string
				digest   string
				wantRows [][]cellObservation
			}{
				{name: "positive", path: fixture.positivePath, digest: fixture.positiveSHA, wantRows: fixture.positiveRows},
				{name: "damaged", path: fixture.damagedPath, digest: fixture.damagedSHA, wantRows: fixture.damagedRows},
			} {
				source := source
				t.Run(source.name, func(t *testing.T) {
					ctx, _ := readV2PDFBounded(t, source.path, source.digest)
					assertReadingOrder(t, ctx, fixture.readingOrder)
					if err := assertCellContentBindings(context.Background(), ctx); err != nil {
						t.Fatalf("MCID/content/ParentTree/StructParents/Pg/P bindings: %v", err)
					}
					headers, rows := observeMappingEvidenceFixture(t, ctx)
					if !reflect.DeepEqual(headers, fixture.headers) {
						t.Fatalf("actual header identities/text/order = %#v, want %#v", headers, fixture.headers)
					}
					if !reflect.DeepEqual(rows, source.wantRows) {
						t.Fatalf("actual data cells and /Headers associations = %#v, want %#v", rows, source.wantRows)
					}
				})
			}
		})
	}
}

func TestMappingEvidenceRejectsCrossFamilySourceDigest(t *testing.T) {
	wrongFamilyPins := []struct {
		name     string
		path     string
		wrongSHA string
	}{
		{
			name:     "statement-positive-as-invoice",
			path:     "../../docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
			wrongSHA: "ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72",
		},
		{
			name:     "statement-damaged-as-invoice",
			path:     "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
			wrongSHA: "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad",
		},
		{
			name:     "invoice-positive-as-statement",
			path:     "../../docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
			wrongSHA: "4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec",
		},
		{
			name:     "invoice-damaged-as-statement",
			path:     "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
			wrongSHA: "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836",
		},
	}
	for _, tc := range wrongFamilyPins {
		t.Run(tc.name, func(t *testing.T) {
			data, err := readV2PDFBytesBoundedWithContext(t.Context(), tc.path, tc.wrongSHA)
			if err == nil || !strings.Contains(err.Error(), "SHA-256") {
				t.Fatalf("cross-family source digest error = %v, want explicit SHA-256 mismatch", err)
			}
			if data != nil {
				t.Fatalf("cross-family source returned %d bytes after digest mismatch, want no bytes", len(data))
			}
		})
	}
}

// These source-acquisition controls keep identity, size, and parse failures
// distinct from missing table associations. All mutations live in t.TempDir.
func TestMappingEvidenceSourceAcquisitionBoundaries(t *testing.T) {
	const statementPositive = "../../docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf"
	const statementSHA = "4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec"

	t.Run("changed-byte-keeps-original-digest", func(t *testing.T) {
		original, err := readV2PDFBytesBoundedWithContext(t.Context(), statementPositive, statementSHA)
		if err != nil {
			t.Fatal(err)
		}
		changed := bytes.Clone(original)
		changed[len(changed)-1] ^= 1
		if bytes.Equal(changed, original) {
			t.Fatal("changed-byte precondition did not change the bytes")
		}
		path := filepath.Join(t.TempDir(), "changed.pdf")
		if err := os.WriteFile(path, changed, 0o600); err != nil {
			t.Fatal(err)
		}
		data, err := readV2PDFBytesBoundedWithContext(t.Context(), path, statementSHA)
		if err == nil || !strings.Contains(err.Error(), "SHA-256") {
			t.Fatalf("changed-byte source error = %v, want SHA-256 mismatch", err)
		}
		if data != nil {
			t.Fatalf("changed-byte source returned %d bytes after digest mismatch", len(data))
		}
	})

	t.Run("unknown-digest", func(t *testing.T) {
		data, err := readV2PDFBytesBoundedWithContext(t.Context(), statementPositive, strings.Repeat("0", sha256.Size*2))
		if err == nil || !strings.Contains(err.Error(), "SHA-256") {
			t.Fatalf("unknown source digest error = %v, want SHA-256 mismatch", err)
		}
		if data != nil {
			t.Fatalf("unknown source digest returned %d bytes", len(data))
		}
	})

	t.Run("oversized-input", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "oversized.pdf")
		if err := os.WriteFile(path, bytes.Repeat([]byte("x"), v2PDFInputLimit+1), 0o600); err != nil {
			t.Fatal(err)
		}
		data, err := readV2PDFBytesBoundedWithContext(t.Context(), path, strings.Repeat("0", sha256.Size*2))
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("exceeds %d-byte limit", v2PDFInputLimit)) {
			t.Fatalf("oversized source error = %v, want input-limit rejection", err)
		}
		if data != nil {
			t.Fatalf("oversized source returned %d bytes", len(data))
		}
	})

	t.Run("malformed-pdf-is-not-missing-association", func(t *testing.T) {
		malformed := []byte("%PDF-2.0\nnot a valid PDF body")
		digest := sha256.Sum256(malformed)
		path := filepath.Join(t.TempDir(), "malformed.pdf")
		if err := os.WriteFile(path, malformed, 0o600); err != nil {
			t.Fatal(err)
		}
		data, err := readV2PDFBytesBoundedWithContext(t.Context(), path, hex.EncodeToString(digest[:]))
		if err != nil {
			t.Fatalf("bounded acquisition of correctly pinned malformed bytes: %v", err)
		}
		parsed, err := api.ReadContext(t.Context(), bytes.NewReader(data), boundedStatelessConfig(t))
		if err == nil {
			t.Fatalf("malformed PDF parsed successfully as %T, want parse error before association observation", parsed)
		}
		if strings.Contains(strings.ToLower(err.Error()), "header") || strings.Contains(strings.ToLower(err.Error()), "association") {
			t.Fatalf("malformed PDF parse error = %v, unexpectedly classified as a table/header association issue", err)
		}
	})
}

// TestMappingEvidenceProposalDiscrimination exercises a test-only comparison
// against fixed fixture-authored expectations. Proposal records are explicit
// test inputs; the validator does not generate proposals or infer mappings.
func TestMappingEvidenceProposalDiscrimination(t *testing.T) {
	tests := []struct {
		name              string
		positivePath      string
		positiveSHA       string
		damagedPath       string
		damagedSHA        string
		targetCell        string
		targetHeaderID    string
		targetHeaderTitle string
		wrongHeaderID     string
		wrongHeaderTitle  string
		positiveCurrentID string
		damagedCurrentID  string
		wrongCellTitle    string
	}{
		{
			name:              "statement",
			positivePath:      "../../docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
			positiveSHA:       "4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec",
			damagedPath:       "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
			damagedSHA:        "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836",
			targetCell:        "$42.10",
			targetHeaderID:    "amountHeader",
			targetHeaderTitle: "Amount",
			wrongHeaderID:     "descriptionHeader",
			wrongHeaderTitle:  "Description",
			positiveCurrentID: "amountHeader",
			damagedCurrentID:  "descriptionHeader",
			wrongCellTitle:    "-$18.25",
		},
		{
			name:              "invoice",
			positivePath:      "../../docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
			positiveSHA:       "ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72",
			damagedPath:       "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
			damagedSHA:        "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad",
			targetCell:        "$8.00",
			targetHeaderID:    "unitAmountHeader",
			targetHeaderTitle: "Unit amount",
			wrongHeaderID:     "quantityHeader",
			wrongHeaderTitle:  "Quantity",
			positiveCurrentID: "unitAmountHeader",
			damagedCurrentID:  "quantityHeader",
			wrongCellTitle:    "$12.50",
		},
	}

	for _, fixture := range tests {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			for _, source := range []struct {
				name      string
				path      string
				digest    string
				currentID string
			}{
				{name: "positive", path: fixture.positivePath, digest: fixture.positiveSHA, currentID: fixture.positiveCurrentID},
				{name: "damaged", path: fixture.damagedPath, digest: fixture.damagedSHA, currentID: fixture.damagedCurrentID},
			} {
				source := source
				t.Run(source.name+"/correct-proposal", func(t *testing.T) {
					ctx, _ := readV2PDFBounded(t, source.path, source.digest)
					headers, rows := observeMappingEvidenceFixture(t, ctx)
					oracle := mappingEvidenceOracle{
						family: fixture.name, sourceSHA256: source.digest,
						targetCell: fixture.targetCell, targetRow: 1, targetColumn: 2,
						targetHeaderID: fixture.targetHeaderID, targetHeaderTitle: fixture.targetHeaderTitle,
						observedTargetHeaderID: source.currentID,
					}
					proposal := mappingEvidenceProposal{
						family: fixture.name, sourceSHA256: source.digest,
						cellTitle: fixture.targetCell, row: 1, column: 2,
						headerID: fixture.targetHeaderID, headerTitle: fixture.targetHeaderTitle,
					}
					assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "valid", "")
					state, err := observeMappingEvidenceTarget(ctx)
					if err != nil {
						t.Fatalf("observe target before classification: %v", err)
					}
					classified := classifyMappingEvidence(ctx, headers, state, oracle, []mappingEvidenceProposal{proposal})
					if classified.status != "valid" || classified.reason != "" || !reflect.DeepEqual(classified.proposals, []mappingEvidenceProposal{proposal}) {
						t.Fatalf("correct proposal classification = %#v, want valid with the exact proposal", classified)
					}
					if !reflect.DeepEqual(classified.observedHeader, state.headerIDs) {
						t.Fatalf("successful classification observed IDs = %#v, want %#v", classified.observedHeader, state.headerIDs)
					}
					afterClassification, err := observeMappingEvidenceTarget(ctx)
					if err != nil || !reflect.DeepEqual(afterClassification, state) {
						t.Fatalf("successful classification changed target: before=%#v after=%#v err=%v", state, afterClassification, err)
					}
					assertMappingEvidenceNonTargetStateUnchanged(t, ctx, headers, rows)
				})
			}

			t.Run("damaged/wrong-existing-header", func(t *testing.T) {
				ctx, _ := readV2PDFBounded(t, fixture.damagedPath, fixture.damagedSHA)
				headers, rows := observeMappingEvidenceFixture(t, ctx)
				oracle := mappingEvidenceOracle{
					family: fixture.name, sourceSHA256: fixture.damagedSHA,
					targetCell: fixture.targetCell, targetRow: 1, targetColumn: 2,
					targetHeaderID: fixture.targetHeaderID, targetHeaderTitle: fixture.targetHeaderTitle,
					observedTargetHeaderID: fixture.damagedCurrentID,
				}
				proposal := mappingEvidenceProposal{
					family: fixture.name, sourceSHA256: fixture.damagedSHA,
					cellTitle: fixture.targetCell, row: 1, column: 2,
					headerID: fixture.wrongHeaderID, headerTitle: fixture.wrongHeaderTitle,
				}
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "rejected", "association_mismatch")
			})

			t.Run("damaged/wrong-target-cell", func(t *testing.T) {
				ctx, _ := readV2PDFBounded(t, fixture.damagedPath, fixture.damagedSHA)
				headers, rows := observeMappingEvidenceFixture(t, ctx)
				oracle := mappingEvidenceOracle{
					family: fixture.name, sourceSHA256: fixture.damagedSHA,
					targetCell: fixture.targetCell, targetRow: 1, targetColumn: 2,
					targetHeaderID: fixture.targetHeaderID, targetHeaderTitle: fixture.targetHeaderTitle,
					observedTargetHeaderID: fixture.damagedCurrentID,
				}
				proposal := mappingEvidenceProposal{
					family: fixture.name, sourceSHA256: fixture.damagedSHA,
					cellTitle: fixture.wrongCellTitle, row: 0, column: 2,
					headerID: fixture.targetHeaderID, headerTitle: fixture.targetHeaderTitle,
				}
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "rejected", "wrong_target_cell")
			})

			t.Run("damaged/cross-family-proposal", func(t *testing.T) {
				ctx, _ := readV2PDFBounded(t, fixture.damagedPath, fixture.damagedSHA)
				headers, rows := observeMappingEvidenceFixture(t, ctx)
				oracle := mappingEvidenceOracle{
					family: fixture.name, sourceSHA256: fixture.damagedSHA,
					targetCell: fixture.targetCell, targetRow: 1, targetColumn: 2,
					targetHeaderID: fixture.targetHeaderID, targetHeaderTitle: fixture.targetHeaderTitle,
					observedTargetHeaderID: fixture.damagedCurrentID,
				}
				wrongFamily := "statements"
				if fixture.name == "statement" {
					wrongFamily = "invoices"
				}
				proposal := mappingEvidenceProposal{
					family: wrongFamily, sourceSHA256: fixture.damagedSHA,
					cellTitle: fixture.targetCell, row: 1, column: 2,
					headerID: fixture.targetHeaderID, headerTitle: fixture.targetHeaderTitle,
				}
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "rejected", "family_mismatch")
			})
		})
	}
}

func TestMappingEvidenceProposalSourceAndHeaderReferences(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
			headers, rows := observeMappingEvidenceFixture(t, ctx)
			state, err := observeMappingEvidenceTarget(ctx)
			if err != nil {
				t.Fatalf("observe target association: %v", err)
			}
			oracle := fixture.oracle()
			correct := fixture.proposal()

			t.Run("correct-source-and-header-reference", func(t *testing.T) {
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, correct), "valid", "")
			})
			t.Run("wrong-source-digest", func(t *testing.T) {
				proposal := correct
				proposal.sourceSHA256 = strings.Repeat("0", sha256.Size*2)
				got := classifyMappingEvidence(ctx, headers, state, oracle, []mappingEvidenceProposal{proposal})
				if got.status != "rejected" || got.reason != "source_mismatch" {
					t.Fatalf("wrong-source proposal classification = %#v, want rejected/source_mismatch", got)
				}
				assertMappingEvidenceNonTargetStateUnchanged(t, ctx, headers, rows)
				after, err := observeMappingEvidenceTarget(ctx)
				if err != nil || !reflect.DeepEqual(after, state) {
					t.Fatalf("wrong-source classification changed target: before=%#v after=%#v err=%v", state, after, err)
				}
			})
			t.Run("unknown-proposed-header", func(t *testing.T) {
				proposal := correct
				proposal.headerID = "unlistedHeader"
				proposal.headerTitle = fixture.correctTitle
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "rejected", "proposal_header_unresolved")
			})
			t.Run("claimed-title-does-not-match-resolved-header", func(t *testing.T) {
				proposal := correct
				proposal.headerTitle = fixture.wrongTitle
				assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposal), "rejected", "proposal_header_reference_mismatch")
			})
		})
	}
}

// TestMappingEvidenceRejectsContentBindingMismatch changes only the decoded
// marked-content association in a fresh parsed copy. It verifies that PDF
// /T and /Headers observations stay fixed while the displayed-content binding
// is rejected. This is a test-only parsed-state mutation, not a new PDF file.
func TestMappingEvidenceRejectsContentBindingMismatch(t *testing.T) {
	tests := []struct {
		name       string
		family     string
		path       string
		digest     string
		cell       string
		headerID   string
		headerText string
		currentID  string
	}{
		{name: "statement-positive", family: "statement", path: "../../docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf", digest: "4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec", cell: "$42.10", headerID: "amountHeader", headerText: "Amount", currentID: "amountHeader"},
		{name: "statement-damaged", family: "statement", path: "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf", digest: "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836", cell: "$42.10", headerID: "amountHeader", headerText: "Amount", currentID: "descriptionHeader"},
		{name: "invoice-positive", family: "invoice", path: "../../docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf", digest: "ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72", cell: "$8.00", headerID: "unitAmountHeader", headerText: "Unit amount", currentID: "unitAmountHeader"},
		{name: "invoice-damaged", family: "invoice", path: "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf", digest: "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad", cell: "$8.00", headerID: "unitAmountHeader", headerText: "Unit amount", currentID: "quantityHeader"},
	}

	for _, fixture := range tests {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
			beforeHeaders, beforeRows := observeMappingEvidenceFixture(t, ctx)
			mutateMappingEvidenceMCIDContent(t, ctx)
			afterHeaders, afterRows := observeMappingEvidenceFixture(t, ctx)
			if !reflect.DeepEqual(beforeHeaders, afterHeaders) || !reflect.DeepEqual(beforeRows, afterRows) {
				t.Fatalf("parsed MCID content mutation changed /T or /Headers observations: before=%#v/%#v after=%#v/%#v", beforeHeaders, beforeRows, afterHeaders, afterRows)
			}
			bindingErr := assertCellContentBindings(context.Background(), ctx)
			if bindingErr == nil || !strings.Contains(bindingErr.Error(), "binds displayed content") {
				t.Fatalf("actual PDF content binding error = %v, want explicit displayed-content mismatch", bindingErr)
			}

			oracle := mappingEvidenceOracle{
				family: fixture.family, sourceSHA256: fixture.digest,
				targetCell: fixture.cell, targetRow: 1, targetColumn: 2,
				targetHeaderID: fixture.headerID, targetHeaderTitle: fixture.headerText,
				observedTargetHeaderID: fixture.currentID,
			}
			proposal := mappingEvidenceProposal{
				family: fixture.family, sourceSHA256: fixture.digest,
				cellTitle: fixture.cell, row: 1, column: 2,
				headerID: fixture.headerID, headerTitle: fixture.headerText,
			}
			assertMappingEvidenceProposal(t, validateMappingEvidenceProposalWithContext(ctx, afterHeaders, afterRows, oracle, proposal), "rejected", "content_binding_mismatch")
		})
	}
}

func mutateMappingEvidenceMCIDContent(t *testing.T, ctx *model.Context) {
	t.Helper()
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("resolve fixture page: %v", err)
	}
	contentsRef, ok := page["Contents"].(types.IndirectRef)
	if !ok {
		t.Fatalf("page /Contents = %T, want indirect reference", page["Contents"])
	}
	stream, _, err := ctx.DereferenceStreamDict(contentsRef)
	if err != nil {
		t.Fatalf("resolve fixture content stream: %v", err)
	}
	if err := stream.Decode(); err != nil {
		t.Fatalf("decode fixture content stream: %v", err)
	}
	mcidSeven, mcidEight := []byte("/MCID 7 >>"), []byte("/MCID 8 >>")
	if bytes.Count(stream.Content, mcidSeven) != 1 || bytes.Count(stream.Content, mcidEight) != 1 {
		t.Fatalf("fixture MCID controls occur %d/%d times, want exactly once each", bytes.Count(stream.Content, mcidSeven), bytes.Count(stream.Content, mcidEight))
	}
	before := append([]byte(nil), stream.Content...)
	stream.Content = bytes.ReplaceAll(stream.Content, mcidSeven, []byte("/MCID temporary >>"))
	stream.Content = bytes.ReplaceAll(stream.Content, mcidEight, mcidSeven)
	stream.Content = bytes.ReplaceAll(stream.Content, []byte("/MCID temporary >>"), mcidEight)
	if bytes.Equal(before, stream.Content) {
		t.Fatal("parsed MCID content mutation did not change the stream")
	}
	ctx.Table[contentsRef.ObjectNumber.Value()].Object = *stream
}

type mappingEvidenceOracle struct {
	family                 string
	sourceSHA256           string
	targetCell             string
	targetRow              int
	targetColumn           int
	targetHeaderID         string
	targetHeaderTitle      string
	observedTargetHeaderID string
}

type mappingEvidenceProposal struct {
	family       string
	sourceSHA256 string
	cellTitle    string
	row          int
	column       int
	headerID     string
	headerTitle  string
	provenance   string
}

type mappingEvidenceProposalResult struct {
	status string
	reason string
}

type mappingEvidenceClassification struct {
	status         string
	reason         string
	observedHeader []string
	proposals      []mappingEvidenceProposal
}

type mappingEvidenceDeltaResult struct {
	status       string
	reason       string
	changedPaths []string
}

type mappingEvidenceSnapshot struct {
	headers        []headerObservation
	rows           [][]cellObservation
	structureFacts []mappingEvidenceObjectFact
	resourceFacts  []mappingEvidenceObjectFact
	pageFacts      []mappingEvidenceObjectFact
}

type mappingEvidenceObjectFact struct {
	path       string
	reference  string
	indirect   bool
	objectType string
	content    string
}

func assertMappingEvidenceProposal(t *testing.T, got mappingEvidenceProposalResult, wantStatus, wantReason string) {
	t.Helper()
	if got.status != wantStatus || got.reason != wantReason {
		t.Fatalf("proposal validation = status %q, reason %q; want status %q, reason %q", got.status, got.reason, wantStatus, wantReason)
	}
}

func validateMappingEvidenceProposal(headers []headerObservation, rows [][]cellObservation, oracle mappingEvidenceOracle, proposal mappingEvidenceProposal) mappingEvidenceProposalResult {
	result := validateMappingEvidenceProposalRecord(headers, rows, oracle, proposal)
	if result.status != "valid" {
		return result
	}
	if proposal.headerID != oracle.targetHeaderID {
		return mappingEvidenceProposalResult{status: "rejected", reason: "association_mismatch"}
	}
	return mappingEvidenceProposalResult{status: "valid"}
}

// Validate proposal binding and references while leaving the authored
// association comparison to the single-proposal decision. This permits a
// conflict check to compare two well-bound, resolvable alternatives without
// treating either as an automatically selected repair.
func validateMappingEvidenceProposalRecord(headers []headerObservation, rows [][]cellObservation, oracle mappingEvidenceOracle, proposal mappingEvidenceProposal) mappingEvidenceProposalResult {
	reject := func(reason string) mappingEvidenceProposalResult {
		return mappingEvidenceProposalResult{status: "rejected", reason: reason}
	}
	if proposal.family != oracle.family {
		return reject("family_mismatch")
	}
	if oracle.sourceSHA256 == "" || proposal.sourceSHA256 != oracle.sourceSHA256 {
		return reject("source_mismatch")
	}
	if proposal.row != oracle.targetRow || proposal.column != oracle.targetColumn || proposal.cellTitle != oracle.targetCell {
		return reject("wrong_target_cell")
	}
	if oracle.targetRow < 0 || oracle.targetRow >= len(rows) || oracle.targetColumn < 0 || oracle.targetColumn >= len(rows[oracle.targetRow]) {
		return reject("target_observation_missing")
	}
	observedCell := rows[oracle.targetRow][oracle.targetColumn]
	if observedCell.title != oracle.targetCell || len(observedCell.headerIDs) != 1 || observedCell.headerIDs[0] != oracle.observedTargetHeaderID {
		return reject("target_observation_mismatch")
	}

	resolveHeader := func(headerID string) (string, bool) {
		title := ""
		found := false
		for _, header := range headers {
			if header.id != headerID {
				continue
			}
			if found {
				return "", false
			}
			title = header.title
			found = true
		}
		return title, found
	}
	wantTitle, ok := resolveHeader(oracle.targetHeaderID)
	if !ok || wantTitle != oracle.targetHeaderTitle {
		return reject("oracle_header_unresolved")
	}
	proposalTitle, ok := resolveHeader(proposal.headerID)
	if !ok {
		return reject("proposal_header_unresolved")
	}
	if proposalTitle != proposal.headerTitle {
		return reject("proposal_header_reference_mismatch")
	}
	return mappingEvidenceProposalResult{status: "valid"}
}

// Read the marked-content to structure binding from the actual parsed PDF.
// A caller-supplied flag cannot bypass or predeclare this result.
func validateMappingEvidenceProposalWithContext(ctx *model.Context, headers []headerObservation, rows [][]cellObservation, oracle mappingEvidenceOracle, proposal mappingEvidenceProposal) mappingEvidenceProposalResult {
	return validateMappingEvidenceProposalRecordWithContext(ctx, headers, rows, oracle, proposal, true)
}

func validateMappingEvidenceProposalRecordWithContext(ctx *model.Context, headers []headerObservation, rows [][]cellObservation, oracle mappingEvidenceOracle, proposal mappingEvidenceProposal, requireAuthoredAssociation bool) mappingEvidenceProposalResult {
	if ctx == nil || assertCellContentBindings(context.Background(), ctx) != nil {
		return mappingEvidenceProposalResult{status: "rejected", reason: "content_binding_mismatch"}
	}
	if requireAuthoredAssociation {
		return validateMappingEvidenceProposal(headers, rows, oracle, proposal)
	}
	return validateMappingEvidenceProposalRecord(headers, rows, oracle, proposal)
}

// observeMappingEvidenceFixture is the actual-structure read seam exercised
// by this evidence test. The shared bounded table reader resolves the fixture
// PDF's IDTree and reads each observed cell's /Headers values.
func observeMappingEvidenceFixture(t *testing.T, ctx *model.Context) ([]headerObservation, [][]cellObservation) {
	t.Helper()
	return tableObservations(t, ctx)
}

func snapshotMappingEvidenceFixture(t *testing.T, ctx *model.Context) mappingEvidenceSnapshot {
	t.Helper()
	headers, rows := observeMappingEvidenceFixture(t, ctx)
	if err := assertCellContentBindings(context.Background(), ctx); err != nil {
		t.Fatalf("snapshot marked-content bindings: %v", err)
	}
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatalf("snapshot StructTreeRoot: %v", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		t.Fatalf("snapshot Document structure: %v", err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		t.Fatalf("snapshot Table structure: %v", err)
	}
	structureFacts := make([]mappingEvidenceObjectFact, 0, 16)
	appendMappingEvidenceObjectFact(t, ctx, &structureFacts, "StructTreeRoot", ctx.RootDict["StructTreeRoot"], false)
	for _, key := range []string{"IDTree", "ParentTree"} {
		if object, ok := root[key]; ok {
			appendMappingEvidenceObjectFact(t, ctx, &structureFacts, "StructTreeRoot."+key, object, false)
		}
	}
	docObject := root["K"]
	appendMappingEvidenceObjectFact(t, ctx, &structureFacts, "StructTreeRoot.K", docObject, false)
	appendMappingEvidenceObjectFact(t, ctx, &structureFacts, "Document.K", doc["K"], false)
	appendMappingEvidenceObjectFact(t, ctx, &structureFacts, "Table.K", table["K"], false)
	rowRefs, ok := table["K"].(types.Array)
	if !ok || len(rowRefs) != 3 {
		t.Fatalf("snapshot Table /K shape = %T/%d, want three rows", table["K"], len(rowRefs))
	}
	for rowIndex, rowRef := range rowRefs {
		appendMappingEvidenceObjectFact(t, ctx, &structureFacts, fmt.Sprintf("Table.rows[%d]", rowIndex), rowRef, false)
		row, err := ctx.DereferenceDict(rowRef)
		if err != nil {
			t.Fatalf("snapshot row %d: %v", rowIndex, err)
		}
		cellRefs, ok := row["K"].(types.Array)
		if !ok || len(cellRefs) != 3 {
			t.Fatalf("snapshot row %d /K shape = %T/%d, want three cells", rowIndex, row["K"], len(cellRefs))
		}
		for cellIndex, cellRef := range cellRefs {
			excludeTargetHeaders := rowIndex == 2 && cellIndex == 2
			appendMappingEvidenceObjectFact(t, ctx, &structureFacts, fmt.Sprintf("Table.rows[%d].cells[%d]", rowIndex, cellIndex), cellRef, excludeTargetHeaders)
		}
	}
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("snapshot page: %v", err)
	}
	pageFacts := make([]mappingEvidenceObjectFact, 0, 11)
	if _, ok := page["StructParents"]; !ok {
		t.Fatal("fixed fixture page has no /StructParents value to snapshot")
	}
	appendMappingEvidenceObjectFact(t, ctx, &pageFacts, "Page./StructParents", page["StructParents"], false)
	contentsRef, ok := page["Contents"].(types.IndirectRef)
	if !ok {
		t.Fatalf("fixed fixture page /Contents = %T, want indirect stream", page["Contents"])
	}
	contentStream, _, err := ctx.DereferenceStreamDict(contentsRef)
	if err != nil {
		t.Fatalf("snapshot page content stream: %v", err)
	}
	if err := contentStream.Decode(); err != nil {
		t.Fatalf("decode fixed page content stream: %v", err)
	}
	pageFacts = append(pageFacts, mappingEvidenceObjectFact{
		path:       "Page.Contents.decoded",
		reference:  fmt.Sprintf("%#v", contentsRef),
		indirect:   true,
		objectType: "decoded-stream-bytes",
		content:    string(contentStream.Content),
	})
	pageFacts = append(pageFacts, mappingEvidenceMCIDTextFacts(t, contentsRef, contentStream.Content)...)
	resources, ok := page["Resources"].(types.Dict)
	if !ok {
		resources, err = ctx.DereferenceDict(page["Resources"])
		if err != nil {
			t.Fatalf("snapshot page Resources: %v", err)
		}
	}
	resourceNames := sortedDictKeys(resources)
	resourceFacts := make([]mappingEvidenceObjectFact, 0, len(resources)*2)
	appendMappingEvidenceObjectFact(t, ctx, &resourceFacts, "Page.Resources", page["Resources"], false)
	// Fixed-fixture-only resource traversal: snapshot each page category, named
	// resource, and one level of indirect dictionary children (such as the font
	// descriptor). This deliberately does not recurse into arbitrary PDF graphs.
	for _, name := range resourceNames {
		categoryPath := "Page.Resources." + name
		appendMappingEvidenceObjectFact(t, ctx, &resourceFacts, categoryPath, resources[name], false)
		category, err := ctx.DereferenceDict(resources[name])
		if err != nil {
			continue
		}
		for _, itemName := range sortedDictKeys(category) {
			itemPath := categoryPath + "." + itemName
			appendMappingEvidenceObjectFact(t, ctx, &resourceFacts, itemPath, category[itemName], false)
			itemObject, err := ctx.Dereference(category[itemName])
			if err != nil {
				t.Fatalf("snapshot resource object %s: %v", itemPath, err)
			}
			if itemDict, ok := itemObject.(types.Dict); ok {
				for _, childName := range sortedDictKeys(itemDict) {
					if _, indirect := itemDict[childName].(types.IndirectRef); indirect {
						appendMappingEvidenceObjectFact(t, ctx, &resourceFacts, itemPath+"."+childName, itemDict[childName], false)
					}
				}
			}
		}
	}
	if len(resourceFacts) == 0 {
		t.Fatal("fixture page has no resource facts to snapshot")
	}
	return mappingEvidenceSnapshot{
		headers:        append([]headerObservation(nil), headers...),
		rows:           cloneMappingEvidenceRows(rows),
		structureFacts: structureFacts,
		resourceFacts:  resourceFacts,
		pageFacts:      pageFacts,
	}
}

func mappingEvidenceMCIDTextFacts(t *testing.T, streamRef types.IndirectRef, content []byte) []mappingEvidenceObjectFact {
	t.Helper()
	sequence := regexp.MustCompile(`(?s)/Span\s*<<\s*/MCID\s+(\d+)\s*>>\s*BDC(.*?)EMC`)
	text := regexp.MustCompile(`\(([^()]*)\)\s*Tj`)
	facts := make([]mappingEvidenceObjectFact, 0, 9)
	seen := make(map[int]struct{}, 9)
	for _, match := range sequence.FindAllSubmatch(content, -1) {
		var mcid int
		if _, err := fmt.Sscan(string(match[1]), &mcid); err != nil {
			t.Fatalf("parse fixed MCID: %v", err)
		}
		textMatches := text.FindAllSubmatch(match[2], -1)
		if len(textMatches) != 1 || len(textMatches[0]) != 2 {
			t.Fatalf("fixed MCID %d has %d literal Tj values, want one", mcid, len(textMatches))
		}
		if _, duplicate := seen[mcid]; duplicate {
			t.Fatalf("fixed page content repeats MCID %d", mcid)
		}
		seen[mcid] = struct{}{}
		facts = append(facts, mappingEvidenceObjectFact{
			path:       fmt.Sprintf("Page.Contents.MCID[%d].text", mcid),
			reference:  fmt.Sprintf("%#v", streamRef),
			indirect:   true,
			objectType: "literal-text",
			content:    string(textMatches[0][1]),
		})
	}
	if len(facts) != 9 {
		t.Fatalf("fixed page content has %d MCID text facts, want 9", len(facts))
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].path < facts[j].path })
	return facts
}

func cloneMappingEvidenceRows(rows [][]cellObservation) [][]cellObservation {
	copyRows := make([][]cellObservation, len(rows))
	for i, row := range rows {
		copyRows[i] = make([]cellObservation, len(row))
		for j, cell := range row {
			copyRows[i][j] = cell
			copyRows[i][j].headerIDs = cloneMappingEvidenceIDs(cell.headerIDs)
		}
	}
	return copyRows
}

func cloneMappingEvidenceSnapshot(snapshot mappingEvidenceSnapshot) mappingEvidenceSnapshot {
	return mappingEvidenceSnapshot{
		headers:        append([]headerObservation(nil), snapshot.headers...),
		rows:           cloneMappingEvidenceRows(snapshot.rows),
		structureFacts: append([]mappingEvidenceObjectFact(nil), snapshot.structureFacts...),
		resourceFacts:  append([]mappingEvidenceObjectFact(nil), snapshot.resourceFacts...),
		pageFacts:      append([]mappingEvidenceObjectFact(nil), snapshot.pageFacts...),
	}
}

func sortedDictKeys(dict types.Dict) []string {
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func firstIndirectDictFact(facts []mappingEvidenceObjectFact) int {
	for i, fact := range facts {
		if fact.indirect && fact.objectType == "types.Dict" {
			return i
		}
	}
	return -1
}

func mappingEvidenceFactByPath(facts []mappingEvidenceObjectFact, path string) *mappingEvidenceObjectFact {
	for i := range facts {
		if facts[i].path == path {
			return &facts[i]
		}
	}
	return nil
}

func mutateMappingEvidenceFactContent(facts []mappingEvidenceObjectFact, path string) {
	if fact := mappingEvidenceFactByPath(facts, path); fact != nil {
		fact.content += "/requested-change"
	}
}

func mutateFirstIndirectDictFact(facts []mappingEvidenceObjectFact) {
	if i := firstIndirectDictFact(facts); i >= 0 {
		facts[i].content += "/same-reference-content-mutant"
	}
}

func appendMappingEvidenceObjectFact(t *testing.T, ctx *model.Context, facts *[]mappingEvidenceObjectFact, path string, object types.Object, excludeTargetHeaders bool) {
	t.Helper()
	resolved, err := ctx.Dereference(object)
	if err != nil {
		t.Fatalf("resolve snapshot object %s (%T): %v", path, object, err)
	}
	content := fmt.Sprintf("%#v", resolved)
	if dict, ok := resolved.(types.Dict); ok {
		copyDict := make(types.Dict, len(dict))
		for key, value := range dict {
			copyDict[key] = value
		}
		if excludeTargetHeaders {
			if attributes, ok := dict["A"].(types.Dict); ok {
				copyAttributes := make(types.Dict, len(attributes))
				for key, value := range attributes {
					if key != "Headers" {
						copyAttributes[key] = value
					}
				}
				copyDict["A"] = copyAttributes
			}
		}
		content = fmt.Sprintf("%#v", copyDict)
	}
	refIdentity, indirect := "direct", false
	if _, ok := object.(types.IndirectRef); ok {
		refIdentity, indirect = fmt.Sprintf("%#v", object), true
	}
	*facts = append(*facts, mappingEvidenceObjectFact{
		path:       path,
		reference:  refIdentity,
		indirect:   indirect,
		objectType: fmt.Sprintf("%T", resolved),
		content:    content,
	})
}

func TestMappingEvidenceClassifiesRequestedSnapshotDeltas(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			for _, tc := range []struct {
				name          string
				mutate        func(*mappingEvidenceSnapshot, mappingEvidenceSnapshot)
				wantStatus    string
				wantReason    string
				wantPaths     []string
				factPathGroup string
			}{
				{name: "target-header-only", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{fixture.correctID}
				}, wantStatus: "allowed", wantPaths: []string{"table.rows[1].cells[2]./Headers"}},
				{name: "target-header-wrong-valid-id", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{fixture.otherID}
				}, wantStatus: "rejected", wantReason: "invalid_requested_association", wantPaths: []string{"table.rows[1].cells[2]./Headers"}},
				{name: "target-header-unknown-id", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{"unlistedHeader"}
				}, wantStatus: "rejected", wantReason: "invalid_requested_association", wantPaths: []string{"table.rows[1].cells[2]./Headers"}},
				{name: "target-header-empty", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{}
				}, wantStatus: "rejected", wantReason: "invalid_requested_association", wantPaths: []string{"table.rows[1].cells[2]./Headers"}},
				{name: "target-header-multiple", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{fixture.correctID, fixture.wrongID}
				}, wantStatus: "rejected", wantReason: "invalid_requested_association", wantPaths: []string{"table.rows[1].cells[2]./Headers"}},
				{name: "target-role-with-header", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{fixture.correctID}
					snapshot.rows[1][2].role = "TH"
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"table.rows[1].cells[2]./Headers", "table.rows[1].cells[2].role"}},
				{name: "target-index-with-header", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].headerIDs = []string{fixture.correctID}
					snapshot.rows[1][2].index++
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"table.rows[1].cells[2]./Headers", "table.rows[1].cells[2].index"}},
				{name: "displayed-text", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[1][2].title = "requested different text"
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"table.rows[1].cells[2]./T"}},
				{name: "row-order", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[0], snapshot.rows[1] = snapshot.rows[1], snapshot.rows[0]
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"table.rows.order"}},
				{name: "another-association", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					snapshot.rows[0][0].headerIDs = []string{fixture.correctID}
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"table.rows[0].cells[0]./Headers"}},
				{name: "structure-object-content-same-reference", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					mutateFirstIndirectDictFact(snapshot.structureFacts)
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", factPathGroup: "structure"},
				{name: "resource-object-content-same-reference", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					mutateFirstIndirectDictFact(snapshot.resourceFacts)
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", factPathGroup: "resource"},
				{name: "resource-reference-swap-same-type", mutate: mutateResourceReferenceToSameTypeStructureObject, wantStatus: "rejected", wantReason: "forbidden_requested_change", factPathGroup: "resource"},
				{name: "struct-parents", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					mutateMappingEvidenceFactContent(snapshot.pageFacts, "Page./StructParents")
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"Page./StructParents"}},
				{name: "decoded-content-bytes", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					mutateMappingEvidenceFactContent(snapshot.pageFacts, "Page.Contents.decoded")
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"Page.Contents.decoded"}},
				{name: "decoded-mcid-text", mutate: func(snapshot *mappingEvidenceSnapshot, _ mappingEvidenceSnapshot) {
					mutateMappingEvidenceFactContent(snapshot.pageFacts, "Page.Contents.MCID[8].text")
				}, wantStatus: "rejected", wantReason: "forbidden_requested_change", wantPaths: []string{"Page.Contents.MCID[8].text"}},
			} {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
					before := snapshotMappingEvidenceFixture(t, ctx)
					assertMappingEvidenceSnapshotPinsRawFacts(t, before, fixture)
					requested := cloneMappingEvidenceSnapshot(before)
					tc.mutate(&requested, before)
					wantPaths := tc.wantPaths
					if tc.factPathGroup == "structure" {
						index := firstIndirectDictFact(before.structureFacts)
						if index < 0 {
							t.Fatal("fixed structure snapshot has no indirect dictionary object")
						}
						wantPaths = []string{before.structureFacts[index].path}
						if requested.structureFacts[index].reference != before.structureFacts[index].reference || requested.structureFacts[index].objectType != before.structureFacts[index].objectType || !requested.structureFacts[index].indirect {
							t.Fatal("structure content control changed reference identity or object type")
						}
					}
					if tc.factPathGroup == "resource" {
						index := firstIndirectDictFact(before.resourceFacts)
						if index < 0 {
							t.Fatal("fixed resource snapshot has no indirect dictionary object")
						}
						wantPaths = []string{before.resourceFacts[index].path}
						if requested.resourceFacts[index].objectType != before.resourceFacts[index].objectType || !requested.resourceFacts[index].indirect {
							t.Fatal("resource requested change altered object type or indirection")
						}
						if tc.name == "resource-object-content-same-reference" && requested.resourceFacts[index].reference != before.resourceFacts[index].reference {
							t.Fatal("resource content control changed reference identity")
						}
						if tc.name == "resource-reference-swap-same-type" && requested.resourceFacts[index].reference == before.resourceFacts[index].reference {
							t.Fatal("resource reference substitution did not change reference identity")
						}
					}
					beforeCopy := cloneMappingEvidenceSnapshot(before)
					requestedCopy := cloneMappingEvidenceSnapshot(requested)
					got := classifyMappingEvidenceSnapshotDelta(before, requested, fixture.oracle())
					if got.status != tc.wantStatus || got.reason != tc.wantReason {
						t.Fatalf("snapshot-delta classification = status %q reason %q paths %#v, want %q/%q paths %#v", got.status, got.reason, got.changedPaths, tc.wantStatus, tc.wantReason, wantPaths)
					}
					if !reflect.DeepEqual(got.changedPaths, wantPaths) {
						t.Fatalf("changed paths = %#v, want exact paths %#v", got.changedPaths, wantPaths)
					}
					if !reflect.DeepEqual(snapshotMappingEvidenceFixture(t, ctx), before) {
						t.Fatal("read-only requested-delta classification changed the parsed fixture state")
					}
					if !reflect.DeepEqual(before, beforeCopy) || !reflect.DeepEqual(requested, requestedCopy) {
						t.Fatal("snapshot-delta classification changed its before or requested-after snapshot")
					}
					if tc.wantStatus == "allowed" && len(got.changedPaths) != 1 {
						t.Fatalf("target-only bounded delta surfaced %d paths, want exactly one", len(got.changedPaths))
					}
				})
			}
		})
	}
}

// TestMappingEvidenceSnapshotDeltaBindsExpectedHeaderToOracle is the G1
// remediation: the snapshot-delta classifier must not authorize a requested
// association against a caller-supplied header ID the authored oracle never
// sanctioned. A classifier that "allows" a well-formed delta to the wrong
// header is fail-open.
func TestMappingEvidenceSnapshotDeltaBindsExpectedHeaderToOracle(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
			before := snapshotMappingEvidenceFixture(t, ctx)
			assertMappingEvidenceSnapshotPinsRawFacts(t, before, fixture)
			requested := cloneMappingEvidenceSnapshot(before)
			// Well-formed single-path delta, but to a non-authored header.
			requested.rows[1][2].headerIDs = []string{fixture.otherID}
			got := classifyMappingEvidenceSnapshotDelta(before, requested, fixture.oracle())
			if got.status != "rejected" || got.reason != "invalid_requested_association" {
				t.Fatalf("unbound expected-header classification = status %q reason %q, want rejected/invalid_requested_association", got.status, got.reason)
			}
			if !reflect.DeepEqual(got.changedPaths, []string{"table.rows[1].cells[2]./Headers"}) {
				t.Fatalf("changed paths = %#v, want the single target /Headers path", got.changedPaths)
			}
		})
	}
}

// TestMappingEvidenceSnapshotDeltaRejectsUnboundOracle pins the second half of
// the G1 fix: the allow decision must also refuse when the oracle's authored
// target header does not resolve to exactly one observed header with the
// oracle's title.
func TestMappingEvidenceSnapshotDeltaRejectsUnboundOracle(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
			before := snapshotMappingEvidenceFixture(t, ctx)
			assertMappingEvidenceSnapshotPinsRawFacts(t, before, fixture)
			requested := cloneMappingEvidenceSnapshot(before)
			requested.rows[1][2].headerIDs = []string{fixture.correctID}
			staleOracle := fixture.oracle()
			staleOracle.targetHeaderTitle = "Stale Title"
			got := classifyMappingEvidenceSnapshotDelta(before, requested, staleOracle)
			if got.status != "rejected" || got.reason != "invalid_requested_association" {
				t.Fatalf("stale-oracle classification = status %q reason %q, want rejected/invalid_requested_association", got.status, got.reason)
			}
		})
	}
}

// TestMappingEvidenceRejectsDuplicateHeaderIDs is the G3 remediation: the
// validator's duplicate-header-ID rejections (classifier
// "ambiguous_document_header", record validator "oracle_header_unresolved"
// and "proposal_header_unresolved") had no test coverage, so a regression
// dropping any of them would survive. These tests pin all three.
func TestMappingEvidenceRejectsDuplicateHeaderIDs(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
			headers, rows := observeMappingEvidenceFixture(t, ctx)
			state, err := observeMappingEvidenceTarget(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(state.headerIDs) != 1 {
				t.Fatalf("fresh target /Headers = %#v, want exactly one observed ID", state.headerIDs)
			}

			duplicateHeader := func(id string) []headerObservation {
				duped := append([]headerObservation(nil), headers...)
				for _, h := range headers {
					if h.id == id {
						return append(duped, h)
					}
				}
				t.Fatalf("no observed header with ID %q to duplicate", id)
				return nil
			}
			observedTitle := func(id string) string {
				for _, h := range headers {
					if h.id == id {
						return h.title
					}
				}
				t.Fatalf("no observed header with ID %q", id)
				return ""
			}

			t.Run("classifier/ambiguous-document-header", func(t *testing.T) {
				duped := duplicateHeader(state.headerIDs[0])
				got := classifyMappingEvidence(ctx, duped, state, fixture.oracle(), []mappingEvidenceProposal{fixture.proposal()})
				if got.status != "rejected" || got.reason != "ambiguous_document_header" {
					t.Fatalf("duplicate-header classification = status %q reason %q, want rejected/ambiguous_document_header", got.status, got.reason)
				}
				if len(got.proposals) != 0 {
					t.Fatalf("ambiguous-header classification surfaced proposals %#v", got.proposals)
				}
			})

			t.Run("record/oracle-header-unresolved", func(t *testing.T) {
				duped := duplicateHeader(fixture.oracle().targetHeaderID)
				got := validateMappingEvidenceProposalWithContext(ctx, duped, rows, fixture.oracle(), fixture.proposal())
				if got.status != "rejected" || got.reason != "oracle_header_unresolved" {
					t.Fatalf("duplicate oracle header validation = status %q reason %q, want rejected/oracle_header_unresolved", got.status, got.reason)
				}
			})

			t.Run("record/proposal-header-unresolved", func(t *testing.T) {
				// Duplicate a non-oracle header and propose it, isolating the
				// proposal-side rejection from the oracle-side one.
				duped := duplicateHeader(fixture.otherID)
				proposal := fixture.proposalWithHeader(fixture.otherID, observedTitle(fixture.otherID))
				got := validateMappingEvidenceProposalWithContext(ctx, duped, rows, fixture.oracle(), proposal)
				if got.status != "rejected" || got.reason != "proposal_header_unresolved" {
					t.Fatalf("duplicate proposal header validation = status %q reason %q, want rejected/proposal_header_unresolved", got.status, got.reason)
				}
			})
		})
	}
}

func assertMappingEvidenceSnapshotPinsRawFacts(t *testing.T, snapshot mappingEvidenceSnapshot, fixture mappingEvidenceFixtureCase) {
	t.Helper()
	var headerFact, fontFact *mappingEvidenceObjectFact
	for i := range snapshot.structureFacts {
		fact := &snapshot.structureFacts[i]
		if fact.path == "Table.rows[0].cells[2]" {
			headerFact = fact
		}
	}
	for i := range snapshot.resourceFacts {
		fact := &snapshot.resourceFacts[i]
		if strings.Contains(fact.content, "\"BaseFont\":\"Helvetica\"") {
			fontFact = fact
			break
		}
	}
	if headerFact == nil || headerFact.objectType != "types.Dict" || !strings.Contains(headerFact.content, fmt.Sprintf("\"ID\":\"%s\"", fixture.correctID)) || !strings.Contains(headerFact.content, fmt.Sprintf("\"T\":\"%s\"", fixture.correctTitle)) {
		t.Fatalf("raw header object fact does not contain fixed authored ID/title: %#v", headerFact)
	}
	if fontFact == nil || !fontFact.indirect || fontFact.objectType != "types.Dict" || !strings.Contains(fontFact.content, "\"Widths\":types.Array") || !strings.Contains(fontFact.content, "\"FontDescriptor\":types.IndirectRef") {
		t.Fatalf("raw same-typed font resource fact lacks actual dictionary fields: %#v", fontFact)
	}
	structParents := mappingEvidenceFactByPath(snapshot.pageFacts, "Page./StructParents")
	if structParents == nil || structParents.objectType != "types.Integer" || structParents.content != "0" {
		t.Fatalf("page /StructParents fact = %#v, want parsed integer 0", structParents)
	}
	decodedContents := mappingEvidenceFactByPath(snapshot.pageFacts, "Page.Contents.decoded")
	if decodedContents == nil || !strings.Contains(decodedContents.content, "/MCID 8 >>") || !strings.Contains(decodedContents.content, fixture.targetCell) {
		t.Fatalf("decoded page content fact lacks actual target MCID/text: %#v", decodedContents)
	}
	if targetText := mappingEvidenceFactByPath(snapshot.pageFacts, "Page.Contents.MCID[8].text"); targetText == nil || targetText.content != fixture.targetCell {
		t.Fatalf("target MCID text fact = %#v, want fixture target %q", targetText, fixture.targetCell)
	}
}

func mutateResourceReferenceToSameTypeStructureObject(requested *mappingEvidenceSnapshot, before mappingEvidenceSnapshot) {
	resourceIndex := firstIndirectDictFact(before.resourceFacts)
	structureIndex := firstIndirectDictFact(before.structureFacts)
	if resourceIndex < 0 || structureIndex < 0 {
		return
	}
	requested.resourceFacts[resourceIndex].reference = before.structureFacts[structureIndex].reference
	requested.resourceFacts[resourceIndex].content = before.structureFacts[structureIndex].content
}

func TestMappingEvidenceClassifiesParsedAssociationFailureStates(t *testing.T) {
	for _, fixture := range mappingEvidenceFixtureCases() {
		fixture := fixture
		t.Run(fixture.family, func(t *testing.T) {
			for _, mutation := range []struct {
				name       string
				headerIDs  []string
				remove     bool
				wantStatus string
				wantReason string
			}{
				{name: "missing-field", remove: true, wantStatus: "rejected", wantReason: "missing_document_association"},
				{name: "empty-array", headerIDs: []string{}, wantStatus: "rejected", wantReason: "missing_document_association"},
				{name: "multiple-current-then-correct", headerIDs: []string{fixture.currentID, fixture.correctID}, wantStatus: "held", wantReason: "multiple_document_associations"},
				{name: "multiple-correct-then-current", headerIDs: []string{fixture.correctID, fixture.currentID}, wantStatus: "held", wantReason: "multiple_document_associations"},
				{name: "duplicate-id", headerIDs: []string{fixture.correctID, fixture.correctID}, wantStatus: "held", wantReason: "multiple_document_associations"},
				{name: "unknown-header", headerIDs: []string{"unknownHeader"}, wantStatus: "rejected", wantReason: "unknown_document_header"},
			} {
				mutation := mutation
				t.Run("document/"+mutation.name, func(t *testing.T) {
					ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
					beforeHeaders, beforeRows := observeMappingEvidenceFixture(t, ctx)
					if got := beforeRows[1][2].headerIDs; !reflect.DeepEqual(got, []string{fixture.currentID}) {
						t.Fatalf("fresh target /Headers = %#v, want %#v", got, []string{fixture.currentID})
					}
					mutateMappingEvidenceTargetHeaders(t, ctx, mutation.headerIDs, mutation.remove)
					after, err := observeMappingEvidenceTarget(ctx)
					if err != nil {
						t.Fatalf("observe deliberately mutated target: %v", err)
					}
					if mutation.remove {
						if after.hasHeaders {
							t.Fatalf("target still has /Headers after removal: %#v", after.headerIDs)
						}
					} else if !reflect.DeepEqual(after.headerIDs, mutation.headerIDs) {
						t.Fatalf("parsed-state mutation /Headers = %#v, want %#v", after.headerIDs, mutation.headerIDs)
					}
					proposal := fixture.proposal()
					got := classifyMappingEvidence(ctx, beforeHeaders, after, fixture.oracle(), []mappingEvidenceProposal{proposal})
					if got.status != mutation.wantStatus || got.reason != mutation.wantReason {
						t.Fatalf("classification = status %q reason %q, want %q/%q", got.status, got.reason, mutation.wantStatus, mutation.wantReason)
					}
					if !reflect.DeepEqual(got.observedHeader, mutation.headerIDs) {
						t.Fatalf("surfaced observed IDs = %#v, want %#v", got.observedHeader, mutation.headerIDs)
					}
					if len(got.proposals) != 0 {
						t.Fatalf("failure classification selected proposals %#v", got.proposals)
					}
					afterClassification, err := observeMappingEvidenceTarget(ctx)
					if err != nil || !reflect.DeepEqual(afterClassification, after) {
						t.Fatalf("classifier changed parsed target: after=%#v current=%#v err=%v", after, afterClassification, err)
					}
					assertMappingEvidenceNonTargetStateUnchanged(t, ctx, beforeHeaders, beforeRows)
				})
			}

			t.Run("proposal/missing", func(t *testing.T) {
				ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
				headers, rows := observeMappingEvidenceFixture(t, ctx)
				state, err := observeMappingEvidenceTarget(ctx)
				if err != nil {
					t.Fatal(err)
				}
				got := classifyMappingEvidence(ctx, headers, state, fixture.oracle(), nil)
				if got.status != "rejected" || got.reason != "missing_proposal" || len(got.proposals) != 0 {
					t.Fatalf("missing proposal classification = %#v, want rejected/missing_proposal without selection", got)
				}
				assertMappingEvidenceNonTargetStateUnchanged(t, ctx, headers, rows)
				current, err := observeMappingEvidenceTarget(ctx)
				if err != nil || !reflect.DeepEqual(current, state) {
					t.Fatalf("missing-proposal classification changed parsed target: before=%#v after=%#v err=%v", state, current, err)
				}
			})

			wrongSource := fixture.proposal()
			wrongSource.sourceSHA256 = strings.Repeat("0", sha256.Size*2)
			wrongFamily := fixture.proposal()
			wrongFamily.family = "other-family"
			wrongCell := fixture.proposal()
			wrongCell.cellTitle = "other-cell"
			wrongHeader := fixture.proposal()
			wrongHeader.headerID = "unlistedHeader"
			wrongHeader.headerTitle = fixture.correctTitle
			wrongHeaderTitle := fixture.proposal()
			wrongHeaderTitle.headerID = fixture.otherID
			wrongHeaderTitle.headerTitle = fixture.correctTitle
			for _, tc := range []struct {
				name       string
				records    []mappingEvidenceProposal
				wantStatus string
				wantReason string
			}{
				{name: "multiple-valid-records-same-association", records: []mappingEvidenceProposal{fixture.proposalFrom("record-source-one"), fixture.proposalFrom("record-source-two")}, wantStatus: "held", wantReason: "multiple_proposal_alternatives"},
				{name: "conflicting-records-forward", records: []mappingEvidenceProposal{fixture.proposalFrom("authored-oracle"), fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, "independent-record")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "conflicting-records-reversed", records: []mappingEvidenceProposal{fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, "independent-record"), fixture.proposalFrom("authored-oracle")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "conflicting-records-same-provenance", records: []mappingEvidenceProposal{fixture.proposalFrom("same-record-source"), fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, "same-record-source")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "conflicting-records-same-provenance-reversed", records: []mappingEvidenceProposal{fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, "same-record-source"), fixture.proposalFrom("same-record-source")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "conflicting-records-empty-provenance", records: []mappingEvidenceProposal{fixture.proposalFrom(""), fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, "")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "conflicting-records-empty-provenance-reversed", records: []mappingEvidenceProposal{fixture.proposalWithHeaderFrom(fixture.wrongID, fixture.wrongTitle, ""), fixture.proposalFrom("")}, wantStatus: "held", wantReason: "conflicting_evidence"},
				{name: "mixed-wrong-source", records: []mappingEvidenceProposal{fixture.proposal(), wrongSource}, wantStatus: "rejected", wantReason: "source_mismatch"},
				{name: "mixed-wrong-family", records: []mappingEvidenceProposal{fixture.proposal(), wrongFamily}, wantStatus: "rejected", wantReason: "family_mismatch"},
				{name: "mixed-wrong-cell", records: []mappingEvidenceProposal{fixture.proposal(), wrongCell}, wantStatus: "rejected", wantReason: "wrong_target_cell"},
				{name: "mixed-unknown-header", records: []mappingEvidenceProposal{fixture.proposal(), wrongHeader}, wantStatus: "rejected", wantReason: "proposal_header_unresolved"},
				{name: "mixed-claimed-header-title-mismatch", records: []mappingEvidenceProposal{fixture.proposal(), wrongHeaderTitle}, wantStatus: "rejected", wantReason: "proposal_header_reference_mismatch"},
			} {
				tc := tc
				t.Run("proposal/"+tc.name, func(t *testing.T) {
					ctx, _ := readV2PDFBounded(t, fixture.path, fixture.digest)
					headers, rows := observeMappingEvidenceFixture(t, ctx)
					state, err := observeMappingEvidenceTarget(ctx)
					if err != nil {
						t.Fatal(err)
					}
					before := mappingEvidenceTargetState{cellTitle: state.cellTitle, hasHeaders: state.hasHeaders, headerIDs: append([]string(nil), state.headerIDs...)}
					got := classifyMappingEvidence(ctx, headers, state, fixture.oracle(), tc.records)
					if got.status != tc.wantStatus || got.reason != tc.wantReason {
						t.Fatalf("proposal classification = status %q reason %q, want %s/%q", got.status, got.reason, tc.wantStatus, tc.wantReason)
					}
					if tc.wantStatus == "held" && !reflect.DeepEqual(got.proposals, tc.records) {
						t.Fatalf("surfaced proposal records = %#v, want all exact records %#v", got.proposals, tc.records)
					}
					if tc.wantStatus == "rejected" && len(got.proposals) != 0 {
						t.Fatalf("invalid mixed records were surfaced as alternatives: %#v", got.proposals)
					}
					after, err := observeMappingEvidenceTarget(ctx)
					if err != nil || !reflect.DeepEqual(after, before) {
						t.Fatalf("read-only classification changed target state: before=%#v after=%#v err=%v", before, after, err)
					}
					assertMappingEvidenceNonTargetStateUnchanged(t, ctx, headers, rows)
				})
			}
		})
	}
}

type mappingEvidenceFixtureCase struct {
	family, path, digest, correctID, correctTitle, wrongID, wrongTitle, otherID, currentID string
	rowOneCells                                                                            []string
	rowOneIDs                                                                              []string
	targetCell                                                                             string
}

func mappingEvidenceFixtureCases() []mappingEvidenceFixtureCase {
	return []mappingEvidenceFixtureCase{
		{family: "statement", path: "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf", digest: "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836", correctID: "amountHeader", correctTitle: "Amount", wrongID: "descriptionHeader", wrongTitle: "Description", otherID: "dateHeader", currentID: "descriptionHeader", rowOneCells: []string{"2026-09-03", "North Harbor", "-$18.25"}, rowOneIDs: []string{"dateHeader", "descriptionHeader", "amountHeader"}, targetCell: "$42.10"},
		{family: "invoice", path: "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf", digest: "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad", correctID: "unitAmountHeader", correctTitle: "Unit amount", wrongID: "quantityHeader", wrongTitle: "Quantity", otherID: "itemHeader", currentID: "quantityHeader", rowOneCells: []string{"Paper", "2", "$12.50"}, rowOneIDs: []string{"itemHeader", "quantityHeader", "unitAmountHeader"}, targetCell: "$8.00"},
	}
}

func (f mappingEvidenceFixtureCase) oracle() mappingEvidenceOracle {
	return mappingEvidenceOracle{family: f.family, sourceSHA256: f.digest, targetCell: f.targetCell, targetRow: 1, targetColumn: 2, targetHeaderID: f.correctID, targetHeaderTitle: f.correctTitle, observedTargetHeaderID: f.currentID}
}

func (f mappingEvidenceFixtureCase) proposal() mappingEvidenceProposal {
	return f.proposalWithHeaderFrom(f.correctID, f.correctTitle, "single-explicit-proposal")
}

func (f mappingEvidenceFixtureCase) proposalFrom(provenance string) mappingEvidenceProposal {
	return f.proposalWithHeaderFrom(f.correctID, f.correctTitle, provenance)
}

func (f mappingEvidenceFixtureCase) proposalWithHeader(id, title string) mappingEvidenceProposal {
	return f.proposalWithHeaderFrom(id, title, "single-explicit-proposal")
}

func (f mappingEvidenceFixtureCase) proposalWithHeaderFrom(id, title, provenance string) mappingEvidenceProposal {
	return mappingEvidenceProposal{family: f.family, sourceSHA256: f.digest, cellTitle: f.targetCell, row: 1, column: 2, headerID: id, headerTitle: title, provenance: provenance}
}

type mappingEvidenceTargetState struct {
	cellTitle  string
	headerIDs  []string
	hasHeaders bool
}

func mutateMappingEvidenceTargetHeaders(t *testing.T, ctx *model.Context, ids []string, remove bool) {
	t.Helper()
	cell, err := mappingEvidenceTargetCellDict(ctx)
	if err != nil {
		t.Fatalf("locate fixture target cell: %v", err)
	}
	attributes, ok := cell["A"].(types.Dict)
	if !ok {
		t.Fatalf("target cell A = %T, want dictionary", cell["A"])
	}
	if remove {
		delete(attributes, "Headers")
		return
	}
	refs := make(types.Array, len(ids))
	for i, id := range ids {
		refs[i] = types.StringLiteral(id)
	}
	attributes["Headers"] = refs
}

func mappingEvidenceTargetCellDict(ctx *model.Context) (types.Dict, error) {
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return nil, err
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return nil, err
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return nil, err
	}
	rows, ok := table["K"].(types.Array)
	if !ok || len(rows) != 3 {
		return nil, fmt.Errorf("Table /K has shape %T/%d, want three rows", table["K"], len(rows))
	}
	row, err := ctx.DereferenceDict(rows[2])
	if err != nil {
		return nil, err
	}
	cells, ok := row["K"].(types.Array)
	if !ok || len(cells) != 3 {
		return nil, fmt.Errorf("target row /K has shape %T/%d, want three cells", row["K"], len(cells))
	}
	return ctx.DereferenceDict(cells[2])
}

func observeMappingEvidenceTarget(ctx *model.Context) (mappingEvidenceTargetState, error) {
	cell, err := mappingEvidenceTargetCellDict(ctx)
	if err != nil {
		return mappingEvidenceTargetState{}, err
	}
	title, ok := cell["T"].(types.StringLiteral)
	if !ok {
		return mappingEvidenceTargetState{}, fmt.Errorf("target /T is %T, want string", cell["T"])
	}
	attributes, ok := cell["A"].(types.Dict)
	if !ok || attributes["O"] != types.Name("Table") {
		return mappingEvidenceTargetState{}, fmt.Errorf("target /A = %T, want table attribute dictionary", cell["A"])
	}
	state := mappingEvidenceTargetState{cellTitle: string(title)}
	value, present := attributes["Headers"]
	if !present {
		return state, nil
	}
	state.hasHeaders = true
	refs, ok := value.(types.Array)
	if !ok {
		return mappingEvidenceTargetState{}, fmt.Errorf("target /Headers is %T, want array", value)
	}
	state.headerIDs = make([]string, 0, len(refs))
	for _, ref := range refs {
		id, ok := ref.(types.StringLiteral)
		if !ok {
			return mappingEvidenceTargetState{}, fmt.Errorf("target /Headers member is %T, want string ID", ref)
		}
		state.headerIDs = append(state.headerIDs, string(id))
	}
	return state, nil
}

// observeMappingEvidenceNonTargetState rereads the fixed table topology while
// deliberately omitting only the target cell's /Headers value. That lets the
// preservation assertion inspect every other cell after tests have mutated
// the target relation into a missing or otherwise invalid state.
func observeMappingEvidenceNonTargetState(ctx *model.Context) ([]headerObservation, [][]cellObservation, error) {
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return nil, nil, fmt.Errorf("resolve StructTreeRoot: %w", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return nil, nil, fmt.Errorf("resolve Document: %w", err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return nil, nil, fmt.Errorf("resolve Table: %w", err)
	}
	rowRefs, ok := table["K"].(types.Array)
	if !ok || len(rowRefs) != 3 {
		return nil, nil, fmt.Errorf("Table /K = %T/%d, want three rows", table["K"], len(rowRefs))
	}
	var headers []headerObservation
	dataRows := make([][]cellObservation, 0, 2)
	for rowIndex, rowRef := range rowRefs {
		row, err := ctx.DereferenceDict(rowRef)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve row %d: %w", rowIndex, err)
		}
		cells, ok := row["K"].(types.Array)
		if !ok || len(cells) != 3 {
			return nil, nil, fmt.Errorf("row %d /K = %T/%d, want three cells", rowIndex, row["K"], len(cells))
		}
		observations := make([]cellObservation, 0, len(cells))
		for column, cellRef := range cells {
			cell, err := ctx.DereferenceDict(cellRef)
			if err != nil {
				return nil, nil, fmt.Errorf("resolve row %d cell %d: %w", rowIndex, column, err)
			}
			title, ok := cell["T"].(types.StringLiteral)
			if !ok {
				return nil, nil, fmt.Errorf("row %d cell %d /T = %T, want string", rowIndex, column, cell["T"])
			}
			observation := cellObservation{title: string(title), index: column}
			if rowIndex == 0 {
				if cell["S"] != types.Name("TH") {
					return nil, nil, fmt.Errorf("header cell %d /S = %v, want TH", column, cell["S"])
				}
				id, ok := cell["ID"].(types.StringLiteral)
				if !ok {
					return nil, nil, fmt.Errorf("header cell %d /ID = %T, want string", column, cell["ID"])
				}
				attributes, ok := cell["A"].(types.Dict)
				if !ok || attributes["O"] != types.Name("Table") {
					return nil, nil, fmt.Errorf("header cell %d /A = %T, want table attributes", column, cell["A"])
				}
				scope, ok := attributes["Scope"].(types.Name)
				if !ok {
					return nil, nil, fmt.Errorf("header cell %d /Scope = %T, want name", column, attributes["Scope"])
				}
				headers = append(headers, headerObservation{title: string(title), id: string(id), scope: string(scope), index: column})
				continue
			}
			if cell["S"] != types.Name("TD") {
				return nil, nil, fmt.Errorf("data cell row %d col %d /S = %v, want TD", rowIndex, column, cell["S"])
			}
			observation.role = "TD"
			if rowIndex != 2 || column != 2 {
				attributes, ok := cell["A"].(types.Dict)
				if !ok || attributes["O"] != types.Name("Table") {
					return nil, nil, fmt.Errorf("data cell row %d col %d /A = %T, want table attributes", rowIndex, column, cell["A"])
				}
				refs, ok := attributes["Headers"].(types.Array)
				if !ok {
					return nil, nil, fmt.Errorf("data cell row %d col %d /Headers = %T, want array", rowIndex, column, attributes["Headers"])
				}
				for _, ref := range refs {
					id, ok := ref.(types.StringLiteral)
					if !ok {
						return nil, nil, fmt.Errorf("data cell row %d col %d header member = %T, want string", rowIndex, column, ref)
					}
					observation.headerIDs = append(observation.headerIDs, string(id))
				}
			}
			observations = append(observations, observation)
		}
		if rowIndex > 0 {
			dataRows = append(dataRows, observations)
		}
	}
	return headers, dataRows, nil
}

func assertMappingEvidenceNonTargetStateUnchanged(t *testing.T, ctx *model.Context, wantHeaders []headerObservation, wantRows [][]cellObservation) {
	t.Helper()
	headers, rows, err := observeMappingEvidenceNonTargetState(ctx)
	if err != nil {
		t.Fatalf("reobserve non-target structure after classification: %v", err)
	}
	if !reflect.DeepEqual(headers, wantHeaders) {
		t.Fatalf("classification changed observed headers: before=%#v after=%#v", wantHeaders, headers)
	}
	wantWithoutTarget := cloneMappingEvidenceRows(wantRows)
	gotWithoutTarget := cloneMappingEvidenceRows(rows)
	if len(wantWithoutTarget) < 2 || len(wantWithoutTarget[1]) < 3 || len(gotWithoutTarget) < 2 || len(gotWithoutTarget[1]) < 3 {
		t.Fatalf("non-target observer returned unexpected row shape: want=%#v got=%#v", wantWithoutTarget, gotWithoutTarget)
	}
	wantWithoutTarget[1][2].headerIDs = nil
	gotWithoutTarget[1][2].headerIDs = nil
	if !reflect.DeepEqual(gotWithoutTarget, wantWithoutTarget) {
		t.Fatalf("classification changed non-target cell observations: before=%#v after=%#v", wantWithoutTarget, gotWithoutTarget)
	}
	if err := assertCellContentBindings(context.Background(), ctx); err != nil {
		t.Fatalf("classification changed marked-content binding: %v", err)
	}
}

func classifyMappingEvidence(ctx *model.Context, headers []headerObservation, state mappingEvidenceTargetState, oracle mappingEvidenceOracle, proposals []mappingEvidenceProposal) mappingEvidenceClassification {
	result := mappingEvidenceClassification{observedHeader: cloneMappingEvidenceIDs(state.headerIDs)}
	if !state.hasHeaders || len(state.headerIDs) == 0 {
		result.status, result.reason = "rejected", "missing_document_association"
		return result
	}
	if len(state.headerIDs) > 1 {
		result.status, result.reason = "held", "multiple_document_associations"
		return result
	}
	resolvedTitle := ""
	resolved := false
	for _, header := range headers {
		if header.id != state.headerIDs[0] {
			continue
		}
		if resolved {
			result.status, result.reason = "rejected", "ambiguous_document_header"
			return result
		}
		resolvedTitle, resolved = header.title, true
	}
	if !resolved {
		result.status, result.reason = "rejected", "unknown_document_header"
		return result
	}
	if len(proposals) == 0 {
		result.status, result.reason = "rejected", "missing_proposal"
		return result
	}
	rows := make([][]cellObservation, oracle.targetRow+1)
	rows[oracle.targetRow] = make([]cellObservation, oracle.targetColumn+1)
	rows[oracle.targetRow][oracle.targetColumn] = cellObservation{title: state.cellTitle, role: "TD", headerIDs: append([]string(nil), state.headerIDs...), index: oracle.targetColumn}
	if len(proposals) > 1 {
		for _, proposal := range proposals {
			validated := validateMappingEvidenceProposalRecordWithContext(ctx, headers, rows, oracle, proposal, false)
			if validated.status != "valid" {
				result.status, result.reason = validated.status, validated.reason
				return result
			}
		}
		result.proposals = append([]mappingEvidenceProposal(nil), proposals...)
		conflicting := false
		for i := 0; i < len(proposals); i++ {
			for j := i + 1; j < len(proposals); j++ {
				if proposals[i].headerID != proposals[j].headerID {
					conflicting = true
				}
			}
		}
		result.status, result.reason = "held", "multiple_proposal_alternatives"
		if conflicting {
			result.reason = "conflicting_evidence"
		}
		return result
	}
	if resolvedTitle == "" {
		result.status, result.reason = "rejected", "unknown_document_header"
		return result
	}
	validated := validateMappingEvidenceProposalWithContext(ctx, headers, rows, oracle, proposals[0])
	if validated.status != "valid" {
		result.status, result.reason = validated.status, validated.reason
		return result
	}
	result.status = "valid"
	result.proposals = append([]mappingEvidenceProposal(nil), proposals...)
	return result
}

func classifyMappingEvidenceSnapshotDelta(before, requested mappingEvidenceSnapshot, oracle mappingEvidenceOracle) mappingEvidenceDeltaResult {
	paths := make([]string, 0)
	rowOrderChanged := mappingEvidenceRowOrderChanged(before.rows, requested.rows)
	if rowOrderChanged {
		paths = append(paths, "table.rows.order")
	}
	if len(before.headers) != len(requested.headers) {
		paths = append(paths, "table.headers.shape")
	} else if !reflect.DeepEqual(before.headers, requested.headers) {
		paths = append(paths, "table.headers")
	}
	if len(before.rows) != len(requested.rows) {
		paths = append(paths, "table.rows.shape")
	} else if !rowOrderChanged {
		for rowIndex := range before.rows {
			if len(before.rows[rowIndex]) != len(requested.rows[rowIndex]) {
				paths = append(paths, fmt.Sprintf("table.rows[%d].shape", rowIndex))
				continue
			}
			for cellIndex := range before.rows[rowIndex] {
				beforeCell, requestedCell := before.rows[rowIndex][cellIndex], requested.rows[rowIndex][cellIndex]
				if beforeCell.title != requestedCell.title {
					paths = append(paths, fmt.Sprintf("table.rows[%d].cells[%d]./T", rowIndex, cellIndex))
				}
				if beforeCell.role != requestedCell.role {
					paths = append(paths, fmt.Sprintf("table.rows[%d].cells[%d].role", rowIndex, cellIndex))
				}
				if beforeCell.index != requestedCell.index {
					paths = append(paths, fmt.Sprintf("table.rows[%d].cells[%d].index", rowIndex, cellIndex))
				}
				if !reflect.DeepEqual(beforeCell.headerIDs, requestedCell.headerIDs) {
					paths = append(paths, fmt.Sprintf("table.rows[%d].cells[%d]./Headers", rowIndex, cellIndex))
				}
			}
		}
	}
	paths = append(paths, mappingEvidenceObjectFactChanges(before.structureFacts, requested.structureFacts)...)
	paths = append(paths, mappingEvidenceObjectFactChanges(before.resourceFacts, requested.resourceFacts)...)
	paths = append(paths, mappingEvidenceObjectFactChanges(before.pageFacts, requested.pageFacts)...)
	sort.Strings(paths)
	result := mappingEvidenceDeltaResult{status: "rejected", changedPaths: paths}
	if len(paths) == 0 {
		result.reason = "no_requested_change"
		return result
	}
	if len(paths) == 1 && paths[0] == "table.rows[1].cells[2]./Headers" {
		if !mappingEvidenceOracleBindsTargetHeader(before.headers, oracle) {
			result.reason = "invalid_requested_association"
			return result
		}
		if len(requested.rows) <= 1 || len(requested.rows[1]) <= 2 || len(requested.rows[1][2].headerIDs) != 1 || requested.rows[1][2].headerIDs[0] != oracle.targetHeaderID {
			result.reason = "invalid_requested_association"
			return result
		}
		result.status = "allowed"
		return result
	}
	result.reason = "forbidden_requested_change"
	return result
}

// mappingEvidenceOracleBindsTargetHeader requires the authored oracle's target
// header to resolve to exactly one observed header carrying the oracle's
// title. The snapshot-delta allow decision must be anchored to authored
// evidence, never to a caller-supplied header ID.
func mappingEvidenceOracleBindsTargetHeader(headers []headerObservation, oracle mappingEvidenceOracle) bool {
	matched := false
	for _, header := range headers {
		if header.id != oracle.targetHeaderID {
			continue
		}
		if matched || header.title != oracle.targetHeaderTitle {
			return false
		}
		matched = true
	}
	return matched
}

func mappingEvidenceObjectFactChanges(before, requested []mappingEvidenceObjectFact) []string {
	beforeByPath := make(map[string]mappingEvidenceObjectFact, len(before))
	requestedByPath := make(map[string]mappingEvidenceObjectFact, len(requested))
	for _, fact := range before {
		beforeByPath[fact.path] = fact
	}
	for _, fact := range requested {
		requestedByPath[fact.path] = fact
	}
	paths := make(map[string]struct{}, len(beforeByPath)+len(requestedByPath))
	for path := range beforeByPath {
		paths[path] = struct{}{}
	}
	for path := range requestedByPath {
		paths[path] = struct{}{}
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		if beforeByPath[path] != requestedByPath[path] {
			ordered = append(ordered, path)
		}
	}
	sort.Strings(ordered)
	return ordered
}

func mappingEvidenceRowOrderChanged(before, requested [][]cellObservation) bool {
	if len(before) != len(requested) || len(before) < 2 {
		return false
	}
	beforeSignatures := make([]string, len(before))
	requestedSignatures := make([]string, len(requested))
	for i := range before {
		beforeSignatures[i] = mappingEvidenceRowSignature(before[i])
		requestedSignatures[i] = mappingEvidenceRowSignature(requested[i])
	}
	if reflect.DeepEqual(beforeSignatures, requestedSignatures) {
		return false
	}
	beforeSet := append([]string(nil), beforeSignatures...)
	requestedSet := append([]string(nil), requestedSignatures...)
	sort.Strings(beforeSet)
	sort.Strings(requestedSet)
	return reflect.DeepEqual(beforeSet, requestedSet)
}

func mappingEvidenceRowSignature(row []cellObservation) string {
	var values []string
	for _, cell := range row {
		values = append(values, cell.title)
	}
	return strings.Join(values, "\x00")
}

func cloneMappingEvidenceIDs(ids []string) []string {
	if ids == nil {
		return nil
	}
	copyIDs := make([]string, len(ids))
	copy(copyIDs, ids)
	return copyIDs
}
