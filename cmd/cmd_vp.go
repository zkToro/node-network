package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hyperledger/aries-framework-go/pkg/doc/util/signature"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	jsonld "github.com/piprate/json-gold/ld"
	"github.com/spf13/cobra"
)

type VC struct {
	Issuer_Public_Key string `json:"issuer_public_key"`
	Holder_DID        string `json:"holder_did"`
	VC                string `json:"vc"`
}

func signVP() error {
	vcData, err := os.ReadFile(cfg.VcPath)

	// claims := jwt.MapClaims{}
	// token, err := jwt.ParseWithClaims(string(dat), claims, func(token *jwt.Token) (interface{}, error) {
	// 	return []byte("<YOUR VERIFICATION KEY>"), nil
	// })
	// // ... error handling
	// if token == nil {
	// 	fmt.Println(err)
	// }
	var vcJson VC
	err = json.Unmarshal(vcData, &vcJson)
	if err != nil {
		fmt.Println(err)
		return err
	}
	issuerPublicKey := vcJson.Issuer_Public_Key
	// builder := verifiable.NewCredentialSchemaLoaderBuilder()

	// factory := gojsonschema.DefaultJSONLoaderFactory{}

	// builder.SetJSONLoader(factory.New("proofOfName.json"))

	documentLoader := jsonld.NewDefaultDocumentLoader(http.DefaultClient)
	fmt.Println("VC:")
	fmt.Println(vcJson.VC)
	fmt.Println("issuer pk: ", vcJson.Issuer_Public_Key)
	fmt.Println("holder did: ", vcJson.Holder_DID)
	vc, err3 := verifiable.ParseCredential([]byte(vcJson.VC),

		verifiable.WithPublicKeyFetcher(verifiable.SingleKey([]byte(issuerPublicKey), kms.ED25519)),
		verifiable.WithJSONLDDocumentLoader(documentLoader),
		verifiable.WithDisabledProofCheck(),
		// verifiable.WithCredentialSchemaLoader(builder.Build()),
	)

	if err3 != nil {
		return err3
	}
	vp, err := verifiable.NewPresentation(verifiable.WithCredentials(vc))

	if err != nil {
		panic(fmt.Errorf("failed to build VP from VC: %w", err))
	}

	vp.Holder = vcJson.Holder_DID
	jwtClaims, err := vp.JWTClaims(nil, true)
	if err != nil {
		panic(fmt.Errorf("failed to create JWT claims of VP: %w", err))
	}

	prvKeyBytes, err := os.ReadFile(cfg.DIDKeyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read node private key: %w", err))
	}
	privateKey := prvKeyBytes
	// fmt.Println("Private Key: ")
	// fmt.Println(hex.EncodeToString(prvKeyBytes))
	// fmt.Println("Public Key: ")
	// fmt.Println(hex.EncodeToString(prvKeyBytes)[64:])
	publicKey, _ := hex.DecodeString(hex.EncodeToString(prvKeyBytes)[64:])
	signer := signature.GetEd25519Signer(privateKey, publicKey)

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, signer, "")
	if err != nil {
		panic(fmt.Errorf("failed to sign VP inside JWT: %w", err))
	}

	fmt.Println(jws)
	fmt.Println("\033[32m", "VP signed", "\033[0m")
	_ = os.WriteFile(cfg.VpPath, []byte(jws), 0644)
	return nil
}

func handleZktoroSignVp(cmd *cobra.Command, args []string) error {
	return signVP()
	// subject := claims["sub"]
}
