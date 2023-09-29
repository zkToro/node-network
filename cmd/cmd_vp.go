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
	"github.com/xeipuuv/gojsonschema"
)

type Keys struct {
	HOLDER_EDDSA_PRIVATE_KEY  string `json:"HOLDER_EDDSA_PRIVATE_KEY"`
	HOLDER_ES256K_PRIVATE_KEY string `json:"HOLDER_ES256K_PRIVATE_KEY"`
	HOLDER_EDDSA_PUBLIC_KEY   string `json:"HOLDER_EDDSA_PUBLIC_KEY"`
	DID_ISSUER                string `json:"DID_ISSUER"`
	ISSUER_PUBLIC_KEY         string `json:"ISSUER_PUBLIC_KEY"`
}

func handleZktoroSignVp(cmd *cobra.Command, args []string) error {

	dat, err := os.ReadFile("proofOfName.jwt")

	// claims := jwt.MapClaims{}
	// token, err := jwt.ParseWithClaims(string(dat), claims, func(token *jwt.Token) (interface{}, error) {
	// 	return []byte("<YOUR VERIFICATION KEY>"), nil
	// })
	// // ... error handling
	// if token == nil {
	// 	fmt.Println(err)
	// }

	keyData, err2 := os.ReadFile("key.json")
	if err2 != nil {
		fmt.Println(err2)
		return err2
	}
	keys := Keys{}
	err = json.Unmarshal([]byte(keyData), &keys)
	if err != nil {
		fmt.Println(err)
		return err
	}
	issuerPublicKey := keys.ISSUER_PUBLIC_KEY
	builder := verifiable.NewCredentialSchemaLoaderBuilder()

	factory := gojsonschema.DefaultJSONLoaderFactory{}

	builder.SetJSONLoader(factory.New("proofOfName.json"))

	documentLoader := jsonld.NewDefaultDocumentLoader(http.DefaultClient)

	vc, err3 := verifiable.ParseCredential(dat,

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
	vp.Holder = keys.DID_ISSUER

	jwtClaims, err := vp.JWTClaims(nil, true)
	if err != nil {
		panic(fmt.Errorf("failed to create JWT claims of VP: %w", err))
	}
	privateKey, _ := hex.DecodeString(keys.HOLDER_EDDSA_PRIVATE_KEY)
	publicKey, _ := hex.DecodeString(keys.HOLDER_EDDSA_PUBLIC_KEY)
	signer := signature.GetEd25519Signer(privateKey, publicKey)

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, signer, "")
	if err != nil {
		panic(fmt.Errorf("failed to sign VP inside JWT: %w", err))
	}

	fmt.Println(jws)
	// subject := claims["sub"]
	return nil
}
