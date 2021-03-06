package signer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/siler23/tezos-hsm-signer/signer/watermark"
)

type testSigner struct {
	SignedBytes []byte
}

func (signer *testSigner) Sign(_ context.Context, message []byte, key *Key) ([]byte, error) {
	return signer.SignedBytes, nil
}

func getTestServer(pkh string) *Server {
	return &Server{
		signer: &testSigner{},
		keys: []Key{Key{
			Name:          "test",
			PublicKeyHash: pkh,
			PublicKey:     "keyhash",
		}, Key{
			Name:          "test2",
			PublicKeyHash: pkh + "2",
			PublicKey:     "keyhash2",
		}},
		filter: OperationFilter{
			EnableTx: false,
		},
		watermark: watermark.GetSessionWatermark(),
	}
}

func TestGetEmptyKeys(t *testing.T) {
	// Test: GET /keys/<invalid key>
	// An empty signing server should return no key results
	server := Server{}

	r := httptest.NewRequest("GET", "/keys/123", strings.NewReader(""))
	w := httptest.NewRecorder()

	Middleware(server.RouteKeys)(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		log.Printf("TestGetEmptyKeys: Expected status code 404.  Received %v\n", resp.StatusCode)
		t.Fail()
	}
	if strings.Contains(string(body), "error") {
		log.Println("TestGetEmptyKeys: requesting in invalid key should return an error")
		t.Fail()
	}
}
func TestGetKeys(t *testing.T) {
	// Test: GET /keys/<valid key>
	// A valid pkh should return a pk
	server := getTestServer("tz123")

	path := "/keys/" + server.keys[0].PublicKeyHash
	r := httptest.NewRequest("GET", path, strings.NewReader(""))
	w := httptest.NewRecorder()

	Middleware(server.RouteKeys)(w, r)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Println("TestGetKeys: Status code should be 200")
		t.Fail()
	}
	if string(body) != "{\"public_key\":\""+server.keys[0].PublicKey+"\"}" {
		log.Println("TestGetKeys: GET /keys/key should return a Public Key")
		t.Fail()
	}
}

func TestRouteUnmatched(t *testing.T) {
	// Test: GET /<random endpoint>
	// Should return a 404
	path := "/random"
	r := httptest.NewRequest("GET", path, strings.NewReader(""))
	w := httptest.NewRecorder()

	Middleware(RouteUnmatched)(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		log.Println("TestRouteUnmatched: Status code should be 404")
		t.Fail()
	}
}

func TestAuthorizedKeys(t *testing.T) {
	// Test: GET /authorized_keys
	// An empy set of authorized keys should be returned
	r := httptest.NewRequest("GET", "/authorized_keys", strings.NewReader(""))
	w := httptest.NewRecorder()

	var server = Server{}
	Middleware(server.RouteAuthorizedKeys)(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("TestAuthorizedKeys: Expected status code 200.  Received %v\n", resp.StatusCode)
		t.Fail()
	}
	if string(body) != "{}" {
		log.Println("TestAuthorizedKeys: Body should always be an empty JSON object")
		log.Println("Received: ", string(body))
		t.Fail()
	}
}

func testPost(t *testing.T, server *Server, test testOperation) (*http.Response, string) {
	// Test: POST /keys/<valid key>
	// A correctly signed JSON payload should be returned

	signedBytes, _ := hex.DecodeString(test.HsmResponse)
	server.signer = &testSigner{
		SignedBytes: signedBytes,
	}
	server.keys[0].PublicKeyHash = test.PublicKeyHash

	// Mock the request
	postBytes := bytes.NewReader([]byte(test.Operation))
	postPath := fmt.Sprintf("/keys/%v", test.PublicKeyHash)
	r := httptest.NewRequest("POST", postPath, postBytes)
	w := httptest.NewRecorder()

	// Sign
	Middleware(server.RouteKeys)(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	return resp, string(body)
}

func compare(t *testing.T, testName string, statusReceived int, statusExpected int, bodyReceived string, bodyExpected string) {
	if statusReceived != statusExpected {
		log.Printf("%v: Expected status code %v.  Received %v\n", testName, statusExpected, statusReceived)
		t.Fail()
	}
	if statusExpected == http.StatusOK && bodyReceived != bodyExpected {
		log.Printf("%v: Body should match the expected signature\n", testName)
		log.Println("Expected: ", bodyExpected)
		log.Println("Received: ", bodyReceived)
		t.Fail()
	}
}

func TestPostTx(t *testing.T) {
	server := getTestServer("tz123")

	server.filter.EnableTx = true
	resp, body := testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 Tx Enabled", resp.StatusCode, http.StatusOK, body, testSecp256k1Tx.SignerResponse)
	resp, body = testPost(t, server, testP256Tx)
	compare(t, "p256 Tx Enabled", resp.StatusCode, http.StatusOK, body, testP256Tx.SignerResponse)

	server.filter.EnableTx = false
	resp, body = testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 Tx Disabled", resp.StatusCode, http.StatusForbidden, body, testSecp256k1Tx.SignerResponse)
	resp, body = testPost(t, server, testP256Tx)
	compare(t, "P256 Tx Disabled", resp.StatusCode, http.StatusForbidden, body, testP256Tx.SignerResponse)
}

func TestPostTxWhitelist(t *testing.T) {
	server := getTestServer("tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m")

	server.filter.EnableTx = true
	server.filter.TxWhitelistAddresses = []string{"tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m"}
	resp, body := testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 On Whitelist", resp.StatusCode, http.StatusOK, body, testSecp256k1Tx.SignerResponse)

	server.filter.EnableTx = false
	server.filter.TxWhitelistAddresses = []string{"tz3fNgiRyEZeXD5eh6rEocSp8PBzii2w38Ku"}
	resp, body = testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 Off Whitelist", resp.StatusCode, http.StatusForbidden, body, testSecp256k1Tx.SignerResponse)
}

func TestPostTxLimit(t *testing.T) {
	server := getTestServer("tz2G4TwEbsdFrJmApAxJ1vdQGmADnBp95n9m")

	server.filter.EnableTx = true
	server.filter.TxDailyMax = new(big.Int).SetInt64(1500000)
	resp, body := testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 Under Amount", resp.StatusCode, http.StatusOK, body, testSecp256k1Tx.SignerResponse)

	resp, body = testPost(t, server, testSecp256k1Tx)
	compare(t, "Secp256k1 Over  Amount", resp.StatusCode, http.StatusForbidden, body, testSecp256k1Tx.SignerResponse)
}

func TestPostEndorse(t *testing.T) {
	server := getTestServer("tz123")
	// Endorsing at the same level twice should fail
	resp, body := testPost(t, server, testEndorseLevel259938)
	compare(t, "Secp256k1 Endorse Same Level #1", resp.StatusCode, http.StatusOK, body, testEndorseLevel259938.SignerResponse)
	resp, body = testPost(t, server, testEndorseLevel259938)
	compare(t, "Secp256k1 Endorse Same Level #2", resp.StatusCode, http.StatusForbidden, body, testEndorseLevel259938.SignerResponse)

	server = getTestServer("tz123")
	// Endorsing at the same level twice should fail
	resp, body = testPost(t, server, testEndorseLevel259939)
	compare(t, "Secp256k1 Endorse Lower Level #1", resp.StatusCode, http.StatusOK, body, testEndorseLevel259939.SignerResponse)
	resp, body = testPost(t, server, testEndorseLevel259938)
	compare(t, "Secp256k1 Endorse Lower Level #2", resp.StatusCode, http.StatusForbidden, body, testEndorseLevel259938.SignerResponse)

	server = getTestServer("tz123")
	// Endorsing at increasing levels should succeed
	resp, body = testPost(t, server, testEndorseLevel259938)
	compare(t, "Secp256k1 Endorse Lower Level #1", resp.StatusCode, http.StatusOK, body, testEndorseLevel259938.SignerResponse)
	resp, body = testPost(t, server, testEndorseLevel259939)
	compare(t, "Secp256k1 Endorse Lower Level #2", resp.StatusCode, http.StatusOK, body, testEndorseLevel259939.SignerResponse)
}
