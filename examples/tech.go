package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 4000 * time.Millisecond,
	}

	techs, err := gorecon.GetTech("https://github.com", client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(techs)
}
