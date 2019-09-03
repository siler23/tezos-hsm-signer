package signer

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"
)

type testGenericOperation struct {
	Name         string
	Operation    string
	Kind         uint8
	Source       string
	Fee          *big.Int
	Counter      *big.Int
	GasLimit     *big.Int
	StorageLimit *big.Int
	Amount       *big.Int
	Destination  string
}

func TestParseKind(t *testing.T) {
	var op *Operation
	op, _ = ParseOperation([]byte(testP256Tx.Operation))
	generic := GetGenericOperation(op)
	if generic.Kind() != opKindTransaction {
		log.Println("Tx was not parsed as a generic transaction")
		t.Fail()
	}
	op, _ = ParseOperation([]byte(testSecp256k1Tx.Operation))
	if generic.Kind() != opKindTransaction {
		log.Println("Tx was not parsed as a generic transaction")
		t.Fail()
	}
}

func testParseGenericOperation(t *testing.T, test *testGenericOperation) {
	op, _ := ParseOperation([]byte(test.Operation))
	generic := GetGenericOperation(op)
	// Verify Each Field
	if generic.Kind() != test.Kind {
		log.Printf("[Generic Test - %v] Kind mismatch. Expected %v but %v parsed from bytes\n", test.Name, test.Kind, generic.Kind())
		t.Fail()
	}

	if generic.TransactionSource() != PubkeyHashToByteString(test.Source) {
		log.Printf("[Generic Test - %v] Source mismatch. Expected %v but parsed from bytes %v\n", test.Name, PubkeyHashToByteString(test.Source), generic.TransactionSource())
		t.Fail()
	}
	if generic.TransactionFee().Cmp(test.Fee) != 0 {
		log.Printf("[Generic Test - %v] Fee mismatch. Expected %v but received %v\n", test.Name, test.Fee, generic.TransactionFee())
		t.Fail()
	}
	if generic.TransactionCounter().Cmp(test.Counter) != 0 {
		log.Printf("[Generic Test - %v] Counter mismatch. Expected %v but received %v\n", test.Name, test.Counter, generic.TransactionCounter())
		t.Fail()
	}
	if generic.TransactionGasLimit().Cmp(test.GasLimit) != 0 {
		log.Printf("[Generic Test - %v] GasLimit mismatch. Expected %v but received %v\n", test.Name, test.GasLimit, generic.TransactionGasLimit())
		t.Fail()
	}
	if generic.TransactionStorageLimit().Cmp(test.StorageLimit) != 0 {
		log.Printf("[Generic Test - %v] StorageLimit mismatch. Expected %v but received %v\n", test.Name, test.StorageLimit, generic.TransactionStorageLimit())
		t.Fail()
	}
	if generic.TransactionAmount().Cmp(test.Amount) != 0 {
		log.Printf("[Generic Test - %v] Amount mismatch. Expected %v but received %v\n", test.Name, test.Amount, generic.TransactionAmount())
		t.Fail()
	}
	if generic.TransactionDestination() != PubkeyHashToByteString(test.Destination) {
		log.Printf("[Generic Test - %v] Destination mismatch. Expected %v but %v parsed from bytes\n", test.Name, PubkeyHashToByteString(test.Destination), generic.TransactionDestination())
		t.Fail()
	}
}

// Test cases generated like:
// ./tezos-client transfer 0.000001 from secp256k1 to p256 \
//   --burn-cap 0.257 \
//   --fee .001272 \
//   --counter 13514 \
//   --gas-limit 10200 \
//   --storage-limit 0 \
//   --verbose-signing --dry-run --force-low-fee
func TestParseTransactions(t *testing.T) {

	testParseGenericOperation(t, &testGenericOperation{
		Name:         "Small Values",
		Kind:         opKindTransaction,
		Operation:    "\"030c4886e771509274c81d97195d0c6c13a9d96287e7d2ed3b086e0e509a1ade0f6c0154f5d8f71ce18f9f05bb885a4120e64c667bc1b4010203040500008c947bf65254cf1a813eb8c6d3f980a89751e2af00\"",
		Source:       "tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m",
		Fee:          new(big.Int).SetInt64(1),
		Counter:      new(big.Int).SetInt64(2),
		GasLimit:     new(big.Int).SetInt64(3),
		StorageLimit: new(big.Int).SetInt64(4),
		Amount:       new(big.Int).SetInt64(5),
		Destination:  "tz1YTMAqhU9icfuDG6FQDdsgWQB4izbSfNSf",
	})
	// 0154f5d8f71ce18f9f05bb885a4120e64c667bc1b47f	]
	testParseGenericOperation(t, &testGenericOperation{
		Name:         "Large Values",
		Kind:         opKindTransaction,
		Operation:    "\"0337761ccb2efac1301653f5f9dd70f29f41145142bd0c7f5a94621cb6b556ef2f6c0154f5d8f71ce18f9f05bb885a4120e64c667bc1b47f80018101ffff038080040001b42958e42271f454f914da474650d580dc9a63ae00\"",
		Source:       "tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m",
		Fee:          new(big.Int).SetInt64(127),
		Counter:      new(big.Int).SetInt64(128),
		GasLimit:     new(big.Int).SetInt64(129),
		StorageLimit: new(big.Int).SetInt64(65535),
		Amount:       new(big.Int).SetInt64(65536),
		Destination:  "tz2QjqpipTjio1q6qsy9wvQcrah33Mx8PWEv",
	})
	testParseGenericOperation(t, &testGenericOperation{
		Name:         "Zero Values",
		Kind:         opKindTransaction,
		Operation:    "\"0368225c9f1857643c6da85eb32ddba298c71a977a05a9b96c2d380097089ab26a6c0154f5d8f71ce18f9f05bb885a4120e64c667bc1b464040200000002a88430950b81e860bc6d7cec866864e46a66781900\"",
		Source:       "tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m",
		Fee:          new(big.Int).SetInt64(100),
		Counter:      new(big.Int).SetInt64(4),
		GasLimit:     new(big.Int).SetInt64(2),
		StorageLimit: new(big.Int).SetInt64(0),
		Amount:       new(big.Int).SetInt64(0),
		Destination:  "tz3bh5VbXnLMyHGUMfhRKYzVXQE1axzTm9FN",
	})
	testParseGenericOperation(t, &testGenericOperation{
		Name:         "tz3 Address",
		Kind:         opKindTransaction,
		Operation:    "\"0329f9e567a875b52e1b03751d38b19b6bf182c1ec95efe5ed7598f9c16b2cbf386c008c947bf65254cf1a813eb8c6d3f980a89751e2af830ace69bc509502c0843d0002a88430950b81e860bc6d7cec866864e46a66781900\"",
		Source:       "tz1YTMAqhU9icfuDG6FQDdsgWQB4izbSfNSf",
		Fee:          new(big.Int).SetInt64(1283),
		Counter:      new(big.Int).SetInt64(13518),
		GasLimit:     new(big.Int).SetInt64(10300),
		StorageLimit: new(big.Int).SetInt64(277),
		Amount:       new(big.Int).SetInt64(1000000),
		Destination:  "tz3bh5VbXnLMyHGUMfhRKYzVXQE1axzTm9FN",
	})

	testParseGenericOperation(t, &testGenericOperation{
		Name:         "KT Address",
		Kind:         opKindTransaction,
		Operation:    "\"0331b45e6df3bb6931e65ab542cc5c5a953f959156fddffc1d554d4b60159cc05b6c0154f5d8f71ce18f9f05bb885a4120e64c667bc1b4f809ca69d84f00c0843d016e7c23cc06c7b0743256f65e34d5b0f7c91e4eb20000\"",
		Source:       "tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m",
		Fee:          new(big.Int).SetInt64(1272),
		Counter:      new(big.Int).SetInt64(13514),
		GasLimit:     new(big.Int).SetInt64(10200),
		StorageLimit: new(big.Int).SetInt64(0),
		Amount:       new(big.Int).SetInt64(1000000),
		Destination:  "KT1JexcFezMnUAaWmvUGY99jwTA4jcKiUgFp",
	})

}

func TestParseProposal(t *testing.T) {
	op, _ := ParseOperation([]byte("\"03ce69c5713dac3537254e7be59759cf59c15abd530d10501ccf9028a5786314cf05008fb5cea62d147c696afd9a93dbce962f4c8a9c910000000a00000020ab22e46e7872aa13e366e455bb4f5dbede856ab0864e1da7e122554579ee71f8\""))
	generic := GetGenericOperation(op)
	// Verify Each Field
	if generic.Kind() != opKindProposal {
		log.Printf("[Proposal Test] Kind mismatch. Expected %v but received %v\n", opKindProposal, generic.Kind())
		t.Fail()
	}
}
func TestParseBallot(t *testing.T) {
	op, _ := ParseOperation([]byte("\"03ce69c5713dac3537254e7be59759cf59c15abd530d10501ccf9028a5786314cf0600531ab5764a29f77c5d40b80a5da45c84468f08a10000000bab22e46e7872aa13e366e455bb4f5dbede856ab0864e1da7e122554579ee71f800\""))
	generic := GetGenericOperation(op)
	// Verify Each Field
	if generic.Kind() != opKindBallot {
		log.Printf("[Proposal Test] Kind mismatch. Expected %v but received %v\n", opKindProposal, generic.Kind())
		t.Fail()
	}
}

func testParseBytes(t *testing.T, bytes string, expect int64) {
	var op GenericOperation
	hex, _ := hex.DecodeString(bytes)
	op = GenericOperation{hex: hex}

	num, _ := op.parseSerializedNumber(0)
	if num.Int64() != expect {
		log.Printf("Expecting %v, received %v\n", expect, num.String())
		t.Fail()
	}
}
func TestParseBytes(t *testing.T) {
	testParseBytes(t, "8001", 128)

	testParseBytes(t, "ff7f", 16383)
	testParseBytes(t, "808001", 16384)
	testParseBytes(t, "818001", 16385)

	testParseBytes(t, "ffff01", 32767)
	testParseBytes(t, "808002", 32768)
	testParseBytes(t, "818002", 32769)

	testParseBytes(t, "ff8002", 32895)
	testParseBytes(t, "808102", 32896)

	testParseBytes(t, "ffff03", 65535)
	testParseBytes(t, "808004", 65536)
}
