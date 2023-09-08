package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang/protobuf/jsonpb"
	log "github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/encoding"
	"zktoro/zktoro-core-go/ipfs"
	"zktoro/zktoro-core-go/protocol"
	"zktoro/zktoro-core-go/security"
)

func main() {
	ctx := context.Background()
	ic, err := ipfs.NewClient("https://ipfs.zktoro.network")
	if err != nil {
		log.WithError(err).Fatal("cannot create ipfs client")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("cid is a required field")
	}

	cid := args[0]
	b, err := ic.GetBytes(ctx, cid)
	if err != nil {
		log.WithError(err).Fatalf("cannot get %s from ipfs", cid)
	}

	jpb := jsonpb.Unmarshaler{AllowUnknownFields: true}

	var sr protocol.SignedPayload
	if err := jpb.Unmarshal(bytes.NewReader(b), &sr); err != nil {
		log.WithError(err).Fatal("cannot parse json")
	}

	var receipt protocol.BatchReceipt
	if err := encoding.DecodeGzippedProto(sr.Encoded, &receipt); err != nil {
		log.WithError(err).Fatal("cannot decode receipt")
	}

	var bs protocol.BatchSummary
	if err := encoding.DecodeGzippedProto(receipt.BatchSummary.Encoded, &bs); err != nil {
		log.WithError(err).Fatal("cannot decode receipt")
	}

	if err := security.VerifySignedPayload(&sr); err != nil {
		log.WithError(err).Error("receipt signature is not valid")
	}
	if err := security.VerifySignedPayload(receipt.BatchSummary); err != nil {
		log.WithError(err).Error("batch summary signature is not valid")
	}
	fmt.Println("receipt (signed payload)")
	fmt.Println(toJson(&sr))
	fmt.Println("")

	fmt.Println("receipt (decoded)")
	fmt.Println(toJson(&receipt))
	fmt.Println("")

	fmt.Println("batch summary (signed payload)")
	fmt.Println(toJson(receipt.BatchSummary))
	fmt.Println("")

	fmt.Println("batch summary (decoded)")
	fmt.Println(toJson(&bs))
	fmt.Println("")
}

func toJson(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}
