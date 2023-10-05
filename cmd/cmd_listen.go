package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func handleZktoroListen(cmd *cobra.Command, args []string) error {

	http.HandleFunc("/putVC", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		// var vc VC

		_ = os.WriteFile(cfg.VcPath, body, 0644)
		fmt.Println("\033[32m", "Credential succesfully stored", "\033[0m")
	})

	http.HandleFunc("/retrieveVP", func(w http.ResponseWriter, r *http.Request) {
		vp, err := os.ReadFile(cfg.VpPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to read VP file"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(vp)
			fmt.Println("\033[32m", "VP Sent to Client", "\033[0m")
		}
	})

	http.HandleFunc("/signVP_temporary", func(w http.ResponseWriter, r *http.Request) {
		err := signVP()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Unable to sign vp: %s", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("VP signed"))

		}
	})
	fmt.Println("listen on 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
	return nil

}
