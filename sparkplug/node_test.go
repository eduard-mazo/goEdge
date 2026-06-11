package sparkplug

import "testing"

// TestNextSeqWraps pins the Sparkplug B sequence-number contract: data
// messages count 1..255 after the NBIRTH's 0, then wrap 255 → 0 (NOT → 1).
// The old `(seq % 255) + 1` skipped 0 on wrap, which a spec-compliant
// consumer flags as a gap once per cycle — a guaranteed periodic rebirth.
func TestNextSeqWraps(t *testing.T) {
	n := NewNode("G", "N")
	for i := 1; i <= 255; i++ {
		if got := n.nextSeq(); got != uint64(i) {
			t.Fatalf("seq #%d = %d; want %d", i, got, i)
		}
	}
	if got := n.nextSeq(); got != 0 {
		t.Fatalf("seq after 255 = %d; want 0 (spec wrap)", got)
	}
	if got := n.nextSeq(); got != 1 {
		t.Fatalf("seq after wrap = %d; want 1", got)
	}
}

// TestBdSeqStablePerSession pins the bdSeq session contract (Sparkplug B
// §6.4.5): the bdSeq registered in the NDEATH will at client construction is
// repeated by every NBIRTH of that session — an in-session rebirth must not
// change it, or the will can never be correlated with the last birth. A new
// client session advances it.
func TestBdSeqStablePerSession(t *testing.T) {
	n := NewNode("G", "N")

	n.NewClientOptions("tcp://localhost:1883", "test", "", "")
	if got := n.Bdseq(); got != 0 {
		t.Fatalf("session 1 bdSeq = %d; want 0", got)
	}

	// publishBinary is a no-op with a nil client; births still run the
	// bookkeeping paths.
	if err := n.PublishNBirth(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := n.PublishNBirth(nil, nil); err != nil { // in-session rebirth
		t.Fatal(err)
	}
	if got := n.Bdseq(); got != 0 {
		t.Fatalf("bdSeq after rebirth = %d; want 0 (stable within session)", got)
	}
	if got := n.lastBirthBdSeq; got != 0 {
		t.Fatalf("lastBirthBdSeq = %d; want 0", got)
	}

	n.NewClientOptions("tcp://localhost:1883", "test", "", "")
	if got := n.Bdseq(); got != 1 {
		t.Fatalf("session 2 bdSeq = %d; want 1", got)
	}
}

// TestNBirthResetsSeq verifies a rebirth resets the node seq so the next data
// message is 1.
func TestNBirthResetsSeq(t *testing.T) {
	n := NewNode("G", "N")
	for i := 0; i < 40; i++ {
		n.nextSeq()
	}
	if err := n.PublishNBirth(nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := n.nextSeq(); got != 1 {
		t.Fatalf("first seq after NBIRTH = %d; want 1", got)
	}
}
